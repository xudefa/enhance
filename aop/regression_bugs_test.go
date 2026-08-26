// Package aop 提供面向切面编程（AOP）支持。
// 本文件包含已确认 Bug 的回归测试，遵循 TDD：先写失败测试，再修复实现。
package aop

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
)

// ==================== Bug 1: InterfaceProxyWrapper 调用接口方法时 panic ====================
//
// 现象：InvokeContext 通过 w.iface.MethodByName 获取接口方法后，
// 使用 method.Func.Call 反射调用。对于接口类型，reflect.Method.Func 是零值
// reflect.Value，调用时 panic("reflect: call of reflect.Value.Call on zero Value")。

func TestInterfaceProxyWrapper_InvokeContext_InterfaceMethods(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("InvokeContext panicked: %v", r)
		}
	}()

	var svc TestServiceInterface = &TestServiceImpl{}
	wrapper := NewInterfaceProxyWrapper(svc, nil, reflect.TypeOf((*TestServiceInterface)(nil)).Elem())

	result, err := wrapper.InvokeContext(context.Background(), "DoSomething")
	if err != nil {
		t.Fatalf("InvokeContext failed: %v", err)
	}
	if result != nil {
		t.Fatalf("unexpected result: %v", result)
	}
}

// ==================== Bug 2: 方法唯一返回值为 error 时未提取为错误 ====================
//
// 现象：executor 仅从 []any（len>=2）中提取 error，对于 func() error 这类
// 唯一返回值即 error 的方法，错误被当作普通结果返回，AfterThrowing 不会触发，
// CallContext 返回 (errValue, nil) 而非 (nil, err)。

type soleErrorSvc struct{}

func (s *soleErrorSvc) Close() error { return errors.New("close boom") }

func (s *soleErrorSvc) Open() error { return nil }

func TestCallContext_SoleErrorReturn_TriggersAfterThrowing(t *testing.T) {
	t.Parallel()

	svc := &soleErrorSvc{}
	var threw atomic.Int32
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Close"),
			Advice:   AfterThrowing(func(jp JoinPoint, err error) { threw.Add(1) }),
			Order:    1,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	result, err := proxy.CallContext(context.Background(), "Close")
	if err == nil {
		t.Fatal("expected error for Close() error method")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
	if threw.Load() == 0 {
		t.Fatal("AfterThrowing advice did not run for sole-error-return method")
	}
}

func TestCallContext_NilSoleError_ReturnsNilResult(t *testing.T) {
	t.Parallel()

	svc := &soleErrorSvc{}
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Open"),
			Advice:   AfterReturning(func(jp JoinPoint, result any) {}),
			Order:    1,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	result, err := proxy.CallContext(context.Background(), "Open")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

// ==================== Bug 3: Around 通知错误传播缺失 ====================
//
// 现象：
//   - (a) executor.go 的 proceed() 硬编码返回 (result, nil)，目标方法返回的错误
//     无法通过 proceed 的 error 返回值被 Around 通知感知；
//   - (b) Around 通知 Execute 返回的 error 只写入 invocation，未写入 JoinPoint，
//     且链式执行只返回结果，导致错误丢失、AfterThrowing 不触发；
//   - (c) aop.Around 便捷包装器丢弃 proceed() 的错误。

func TestAroundProceed_PropagatesTargetError(t *testing.T) {
	t.Parallel()

	svc := &soleErrorSvc{}
	var sawErr error
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Close"),
			Advice: NewAroundAdvice(func(ctx context.Context, jp JoinPoint, proceed func() (any, error)) (any, error) {
				_, err := proceed()
				sawErr = err
				return nil, err
			}, 1),
			Order: 1,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	result, err := proxy.CallContext(context.Background(), "Close")
	if err == nil {
		t.Fatal("expected error from target method through around chain")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
	if sawErr == nil {
		t.Fatal("around advice proceed() should return target method error")
	}
}

func TestAroundAdvice_ReturnsError_TriggersAfterThrowing(t *testing.T) {
	t.Parallel()

	svc := &echoSvc{}
	var threw atomic.Int32
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Echo"),
			Advice: NewAroundAdvice(func(ctx context.Context, jp JoinPoint, proceed func() (any, error)) (any, error) {
				return nil, errors.New("around boom")
			}, 1),
			Order: 1,
		},
		{
			PointCut: MatchByName("Echo"),
			Advice:   AfterThrowing(func(jp JoinPoint, err error) { threw.Add(1) }),
			Order:    2,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	result, err := proxy.CallContext(context.Background(), "Echo", "x", 1)
	if err == nil {
		t.Fatal("expected error from around advice")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
	if threw.Load() == 0 {
		t.Fatal("AfterThrowing advice did not run when around advice returned error")
	}
}

func TestLegacyAround_ForwardedError_TriggersAfterThrowing(t *testing.T) {
	t.Parallel()

	svc := &soleErrorSvc{}
	var threw atomic.Int32
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Close"),
			Advice: Around(func(jp JoinPoint, proceed ProceedFunc) any {
				return proceed()
			}),
			Order: 1,
		},
		{
			PointCut: MatchByName("Close"),
			Advice:   AfterThrowing(func(jp JoinPoint, err error) { threw.Add(1) }),
			Order:    2,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	result, err := proxy.CallContext(context.Background(), "Close")
	if err == nil {
		t.Fatal("expected error propagated through legacy Around advice")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
	if threw.Load() == 0 {
		t.Fatal("AfterThrowing advice did not run after legacy Around propagated error")
	}
}

func TestAround_RecoveredError_DoesNotTriggerAfterThrowing(t *testing.T) {
	t.Parallel()

	svc := &soleErrorSvc{}
	var threw atomic.Int32
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Close"),
			Advice: NewAroundAdvice(func(ctx context.Context, jp JoinPoint, proceed func() (any, error)) (any, error) {
				_, _ = proceed()
				return "recovered", nil
			}, 1),
			Order: 1,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	result, err := proxy.CallContext(context.Background(), "Close")
	if err != nil {
		t.Fatalf("recovered error should not propagate: %v", err)
	}
	if result != "recovered" {
		t.Fatalf("expected recovered result, got %v", result)
	}
	if threw.Load() != 0 {
		t.Fatal("AfterThrowing advice should not run when around advice recovered the error")
	}
}

// ==================== Bug 4: SortAspectsByOrder 不稳定排序 ====================
//
// 现象：slices.SortFunc 是不稳定排序，当多个切面的 Order 相同时，
// 其相对执行顺序不受保证（可能被随机打乱），导致切面执行顺序不确定。
// 应改用 slices.SortStableFunc 保证相同 Order 的切面按注册顺序执行。

func TestSortAspectsByOrder_StableForEqualOrder(t *testing.T) {
	t.Parallel()

	aspects := make([]*AspectMeta, 0, 13)
	for i := 0; i < 11; i++ {
		aspects = append(aspects, &AspectMeta{Order: 1, Instance: "high-" + strconv.Itoa(i)})
	}
	aspects = append(aspects, &AspectMeta{Order: 0, Instance: "low"})
	aspects = append(aspects, &AspectMeta{Order: 1, Instance: "high-11"})

	SortAspectsByOrder(aspects)

	if aspects[0].Instance != "low" {
		t.Fatalf("first aspect should be low, got %v", aspects[0].Instance)
	}
	highs := make([]string, 0, 12)
	for _, a := range aspects[1:] {
		name, _ := a.Instance.(string)
		highs = append(highs, name)
	}
	want := []string{"high-0", "high-1", "high-2", "high-3", "high-4",
		"high-5", "high-6", "high-7", "high-8", "high-9", "high-10", "high-11"}
	if len(highs) != len(want) {
		t.Fatalf("expected %d high aspects, got %d: %v", len(want), len(highs), highs)
	}
	for i := range want {
		if highs[i] != want[i] {
			t.Fatalf("equal-order aspects reordered: got %v, want %v", highs, want)
		}
	}
}

// ==================== Bug 5: MatchByName 的 glob 转换破坏正则表达式 ====================
//
// 现象：MatchByName 先判断是否包含 glob 通配符（*?），再进行 glob->正则转换。
// 对于含 ? 或 * 的正则表达式（如 "foo(bar)?baz"、"^Do.*"），会被误当作 glob 处理，
// 导致转换后的正则与原意不符，匹配结果错误。

func TestMatchByName_RegexWithMetacharNotMangled(t *testing.T) {
	t.Parallel()

	pc := MatchByName("foo(bar)?baz")

	if !pc.Matches(nil, "foobaz") {
		t.Error("foo(bar)?baz should match foobaz (bar is optional)")
	}
	if !pc.Matches(nil, "foobarbaz") {
		t.Error("foo(bar)?baz should match foobarbaz")
	}
	if pc.Matches(nil, "fooxbaz") {
		t.Error("foo(bar)?baz should not match fooxbaz")
	}
}

func TestMatchByName_RegexWithDotStar(t *testing.T) {
	t.Parallel()

	pc := MatchByName("^Do.*")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("^Do.* should match DoSomething")
	}
	if pc.Matches(nil, "GetValue") {
		t.Error("^Do.* should not match GetValue")
	}
}
