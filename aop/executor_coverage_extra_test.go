package aop

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

var covSentinelErr = errors.New("cov sentinel")

func TestChainJoinPoint_ArgsAndContextDelegation(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), struct{}{}, "v")
	inner := &joinPointImpl{target: "t", method: "M", args: []any{1, 2}, ctx: ctx}
	cjp := &chainJoinPoint{
		inner:   inner,
		inv:     &invocationImpl{joinPoint: inner},
		proceed: func() (any, error) { return nil, nil },
	}

	if got := cjp.Args(); len(got) != 2 || got[0] != 1 {
		t.Errorf("Args delegation failed: %v", got)
	}
	if cjp.Context() != ctx {
		t.Error("Context delegation failed")
	}
}

func TestExecutor_NilInputs(t *testing.T) {
	t.Parallel()

	exec := NewChainExecutor()

	if got := exec.Execute(nil, nil, func(...any) any { return 1 }); got != nil {
		t.Errorf("expected nil for nil invocation, got %v", got)
	}

	inv := NewInvocation(&joinPointImpl{}, func() (any, error) { return nil, nil })
	if got := exec.Execute(inv, nil, nil); got != nil {
		t.Errorf("expected nil for nil targetFunc, got %v", got)
	}
}

func TestExecutor_InvocationWithoutJoinPoint(t *testing.T) {
	t.Parallel()

	exec := NewChainExecutor()
	inv := &invocationImpl{args: []any{7}}

	got := exec.Execute(inv, nil, func(args ...any) any {
		return args[0].(int) + 1
	})

	if got != 8 {
		t.Errorf("expected targetFunc invoked with args, got %v", got)
	}
}

func TestExecutor_ProceedErrorSetsErrorOnJoinPointAndInvocation(t *testing.T) {
	t.Parallel()

	threw := false
	jp := &joinPointImpl{
		proceed: func() (any, error) { return nil, covSentinelErr },
	}
	inv := NewInvocation(jp, jp.Proceed)

	aspects := []*AspectMeta{{
		PointCut: MatchAll(),
		Advice:   AfterThrowing(func(j JoinPoint, err error) { threw = true }),
	}}

	result := NewChainExecutor().Execute(inv, aspects, func(...any) any { return "ignored" })

	if result != nil {
		t.Errorf("expected nil result on proceed error, got %v", result)
	}
	if !errors.Is(jp.GetError(), covSentinelErr) {
		t.Errorf("expected join point error, got %v", jp.GetError())
	}
	if impl, ok := inv.(*invocationImpl); ok {
		if !errors.Is(impl.Error(), covSentinelErr) {
			t.Errorf("expected invocation error, got %v", impl.Error())
		}
	} else {
		t.Fatal("expected invocationImpl")
	}
	if !threw {
		t.Error("expected AfterThrowing advice to run on proceed error")
	}
}

func TestExecutor_NoRecoveryOption(t *testing.T) {
	t.Parallel()

	afterRan := false
	exec := &defaultChainExecutor{config: chainExecutorConfig{recoverPanic: false}}
	jp := &joinPointImpl{proceed: func() (any, error) { panic("no-recovery boom") }}
	inv := NewInvocation(jp, jp.Proceed)

	aspects := []*AspectMeta{{
		PointCut: MatchAll(),
		Advice:   After(func(_ JoinPoint) { afterRan = true }),
	}}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic to propagate without recovery")
		}
		info, ok := r.(*PanicInfo)
		if !ok {
			t.Fatalf("expected *PanicInfo, got %T", r)
		}
		if info.Value != "no-recovery boom" {
			t.Errorf("unexpected panic value: %v", info.Value)
		}
		if afterRan {
			t.Error("After advice must not run when panic escapes without recovery")
		}
	}()

	exec.Execute(inv, aspects, func(...any) any { return "unused" })
	t.Error("expected Execute to panic")
}

func TestExecutor_PanicTriggersAfterThrowingThenPanicInfo(t *testing.T) {
	t.Parallel()

	threw := false
	afterRan := false
	aspects := []*AspectMeta{
		{
			PointCut: MatchAll(),
			Advice:   After(func(_ JoinPoint) { afterRan = true }),
		},
		{
			PointCut: MatchAll(),
			Advice:   AfterThrowing(func(_ JoinPoint, err error) { threw = true }),
		},
	}
	jp := &joinPointImpl{proceed: func() (any, error) { panic("target exploded") }}
	inv := NewInvocation(jp, jp.Proceed)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected re-panicked PanicInfo")
		}
		info, ok := r.(*PanicInfo)
		if !ok {
			t.Fatalf("expected *PanicInfo, got %T", r)
		}
		if info.Value != "target exploded" {
			t.Errorf("unexpected panic value: %v", info.Value)
		}
		if len(info.Stack) == 0 {
			t.Error("expected stack in PanicInfo")
		}
		if !afterRan {
			t.Error("expected After to run before re-panic")
		}
		if !threw {
			t.Error("expected AfterThrowing to run before re-panic")
		}
	}()

	NewChainExecutor().Execute(inv, aspects, func(...any) any { return "unused" })
}

func TestExecutor_InterceptorWrapsExecution(t *testing.T) {
	t.Parallel()

	events := make([]string, 0, 2)
	var mu sync.Mutex
	interceptor := func(inv Invocation, next func(Invocation) any) any {
		mu.Lock()
		events = append(events, "before")
		mu.Unlock()
		result := next(inv)
		mu.Lock()
		events = append(events, "after")
		mu.Unlock()
		return result
	}

	exec := NewChainExecutor(WithInterceptor(interceptor))
	jp := &joinPointImpl{proceed: func() (any, error) { return "core", nil }}
	inv := NewInvocation(jp, jp.Proceed)

	got := exec.Execute(inv, nil, func(...any) any { return "unused" })
	if got != "core" {
		t.Errorf("interceptor must pass through core result, got %v", got)
	}
	if len(events) != 2 || events[0] != "before" || events[1] != "after" {
		t.Errorf("unexpected interceptor events: %v", events)
	}
}

func TestMethodInvocation_CallMethod_Branches(t *testing.T) {
	t.Parallel()

	voidFn := func(a string) {}
	singleFn := func(a int) int { return a * 3 }
	multiFn := func(a, b int) (int, int) { return a, b }

	tests := []struct {
		name   string
		mi     *MethodInvocation
		callAr []any
		want   any
	}{
		{"nil func returns nil", &MethodInvocation{Func: nil}, nil, nil},
		{"non-func returns nil", &MethodInvocation{Func: "not-a-func"}, nil, nil},
		{"void function returns nil", &MethodInvocation{Func: voidFn}, []any{"x"}, nil},
		{"single result", &MethodInvocation{Func: singleFn}, []any{4}, 12},
		{"multi result as slice", &MethodInvocation{Func: multiFn}, []any{1, 2}, []any{1, 2}},
		{
			"no call args falls back to Params",
			&MethodInvocation{Func: singleFn, Params: []any{5}},
			nil,
			15,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.mi.callMethod(tt.callAr...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("callMethod() = %v (%T), want %v", got, got, tt.want)
			}
		})
	}
}

func TestMethodInvocation_ProceedWithArgs_WithoutProceed(t *testing.T) {
	t.Parallel()

	mi := &MethodInvocation{
		Func:   func(n int) int { return n + 100 },
		Params: []any{0},
	}

	got, err := mi.ProceedWithArgs([]any{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 101 {
		t.Errorf("ProceedWithArgs via callMethod = %v, want 101", got)
	}
}

func TestMethodInvocation_Proceed_FallsBackToCallMethod(t *testing.T) {
	t.Parallel()

	mi := &MethodInvocation{
		Func:   func(n int) int { return n - 1 },
		Params: []any{0},
	}

	got, err := mi.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != -1 {
		t.Errorf("Proceed via callMethod with Params = %v, want -1", got)
	}
}

func TestExecuteChain_TargetExecutesWithoutAround(t *testing.T) {
	t.Parallel()

	executed := false
	beforeRan := false

	jp := &MethodInvocation{
		MethodName: "Compute",
		Func:       func(n int) int { executed = true; return n * 10 },
		Params:     []any{3},
		Object:     &covPcfSvc{},
	}

	aspects := []*AspectMeta{{
		PointCut: MatchByName("Compute"),
		Advice:   Before(func(JoinPoint) { beforeRan = true }),
		Order:    1,
	}}

	got := ExecuteChain(jp, aspects)
	if got != 30 {
		t.Errorf("ExecuteChain result = %v, want 30", got)
	}
	if !executed || !beforeRan {
		t.Errorf("expected target and before advice to run (executed=%v beforeRan=%v)", executed, beforeRan)
	}
}

func TestNewJoinPointWithContext_NilContext(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(nil, "target", "Method", nil, nil, nil)
	if jp.Context() == nil {
		t.Fatal("expected fallback context.Background(), got nil")
	}
}

func TestAdviceType_String_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		advice AdviceType
		want   string
	}{
		{"before", AdviceTypeBefore, "before"},
		{"after", AdviceTypeAfter, "after"},
		{"around", AdviceTypeAround, "around"},
		{"after returning", AdviceTypeAfterReturning, "after_returning"},
		{"after throwing", AdviceTypeAfterThrowing, "after_throwing"},
		{"unknown", AdviceType(99), "unknown"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.advice.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSortAspectsByOrder_NilAfterNonNil(t *testing.T) {
	t.Parallel()

	a := &AspectMeta{Order: 2}
	b := (*AspectMeta)(nil)
	list := []*AspectMeta{a, b}

	SortAspectsByOrder(list)

	if list[0] != a || list[1] != b {
		t.Errorf("nil element should sort after non-nil, got order change")
	}
}

func TestAroundConvenience_PropagatesProceedError(t *testing.T) {
	t.Parallel()

	jp := &joinPointImpl{
		proceedWithArgs: func([]any) (any, error) { return nil, covSentinelErr },
	}

	adv := Around(func(j JoinPoint, p ProceedFunc) any {
		return p("arg")
	})

	_, err := adv.Execute(context.Background(), jp)
	if !errors.Is(err, covSentinelErr) {
		t.Fatalf("expected sentinel error propagated by Around, got %v", err)
	}
}
