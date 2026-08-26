package aop

import (
	"context"
	"sync/atomic"
	"testing"
)

// failSvc 方法返回错误的测试服务。
type failSvc struct{}

// Do 返回 (string, error)。
func (f *failSvc) Do() (string, error) { return "", &failErr{} }

// failErr 实现 error。
type failErr struct{}

func (f *failErr) Error() string { return "boom" }

// TestAfterThrowing_NonAroundPath 验证无 Around 通知时，方法返回错误也会触发 AfterThrowing。
func TestAfterThrowing_NonAroundPath(t *testing.T) {
	t.Parallel()

	svc := &failSvc{}
	var threw atomic.Int32
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Do"),
			Advice:   AfterThrowing(func(jp JoinPoint, err error) { threw.Add(1) }),
			Order:    1,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	if _, err := proxy.Call("Do"); err == nil {
		t.Fatal("expected error")
	}
	if threw.Load() == 0 {
		t.Fatal("afterThrowing advice did not run for method returning error (non-around path)")
	}
}

// echoSvc 记录收到的参数。
type echoSvc struct{ got []any }

// Echo 返回收到的参数。
func (e *echoSvc) Echo(name string, age int) (string, int) {
	e.got = []any{name, age}
	return name, age
}

// TestProceedWithArgs_EffectiveThroughChain 验证 Around 通知中 ProceedWithArgs 修改的参数会传递到目标方法。
func TestProceedWithArgs_EffectiveThroughChain(t *testing.T) {
	t.Parallel()

	svc := &echoSvc{}
	factory := NewProxyFactory(svc)
	factory.SetAspects([]*AspectMeta{
		{
			PointCut: MatchByName("Echo"),
			Advice: Around(func(jp JoinPoint, proceed ProceedFunc) any {
				jp.ProceedWithArgs([]any{"changed", 99})
				return nil
			}),
			Order: 1,
		},
	})

	proxy := factory.GetProxy().(*ReflectiveAopProxy)
	if _, err := proxy.CallContext(context.Background(), "Echo", "orig", 1); err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if len(svc.got) != 2 || svc.got[0] != "changed" || svc.got[1] != 99 {
		t.Fatalf("target received stale args: %v", svc.got)
	}
}

// TestProxyFactory_NoMatch_ReturnsSameInstance 验证无匹配切面时 GetProxy 返回原对象而非拷贝。
func TestProxyFactory_NoMatch_ReturnsSameInstance(t *testing.T) {
	t.Parallel()

	target := &TestUserService{}
	factory := NewProxyFactory(target)
	proxy := factory.GetProxy()

	if proxy != target {
		t.Fatal("GetProxy should return the same instance when no aspects match")
	}
}
