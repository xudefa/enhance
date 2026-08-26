package aop

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestNewJoinPointWithContext_AllFields(t *testing.T) {
	t.Parallel()

	target := "myTarget"
	method := "MyMethod"
	args := []any{"a", 1}
	proceedFn := func() (any, error) { return "proceed-result", nil }
	proceedWithArgsFn := func(a []any) (any, error) { return "proceedWithArgs-result", nil }
	ctx := context.WithValue(context.Background(), struct{}{}, "val")

	jp := NewJoinPointWithContext(ctx, target, method, args, proceedFn, proceedWithArgsFn)

	if jp.Target() != target {
		t.Errorf("Target() = %v, want %v", jp.Target(), target)
	}
	if jp.Method() != method {
		t.Errorf("Method() = %v, want %v", jp.Method(), method)
	}
	if len(jp.Args()) != 2 || jp.Args()[0] != "a" {
		t.Errorf("Args() = %v, want [a, 1]", jp.Args())
	}
	if jp.Context() != ctx {
		t.Error("Context() should match provided context")
	}

	result, err := jp.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "proceed-result" {
		t.Errorf("Proceed() = %v, want 'proceed-result'", result)
	}

	result, err = jp.ProceedWithArgs([]any{"new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "proceedWithArgs-result" {
		t.Errorf("ProceedWithArgs() = %v, want 'proceedWithArgs-result'", result)
	}
}

func TestJoinPointImpl_NilProceedFunctions(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", nil, nil, nil)

	result, err := jp.Proceed()
	if result != nil || err != nil {
		t.Errorf("nil Proceed: got (%v, %v), want (nil, nil)", result, err)
	}

	result, err = jp.ProceedWithArgs(nil)
	if result != nil || err != nil {
		t.Errorf("nil ProceedWithArgs: got (%v, %v), want (nil, nil)", result, err)
	}
}

func TestJoinPointImpl_SetResultAndGetResult(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", nil, nil, nil)

	jp.SetResult("hello")
	if jp.GetResult() != "hello" {
		t.Errorf("GetResult() = %v, want 'hello'", jp.GetResult())
	}

	jp.SetResult(42)
	if jp.GetResult() != 42 {
		t.Errorf("GetResult() = %v, want 42", jp.GetResult())
	}
}

func TestJoinPointImpl_SetErrorAndGetError(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", nil, nil, nil)

	if jp.GetError() != nil {
		t.Error("GetError() should return nil initially")
	}

	sentinel := errors.New("test error")
	jp.SetError(sentinel)
	if !errors.Is(jp.GetError(), sentinel) {
		t.Errorf("GetError() = %v, want %v", jp.GetError(), sentinel)
	}
}

func TestNewInvocation_AllMethods(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(
		context.Background(),
		"target",
		"M",
		[]any{"arg1", "arg2"},
		func() (any, error) {
			return "result", nil
		},
		nil,
	)

	inv := NewInvocation(jp, func() (any, error) { return "inv-result", nil })

	if inv.JoinPoint() != jp {
		t.Error("JoinPoint() should return the provided join point")
	}

	args := inv.Arguments()
	if len(args) != 2 || args[0] != "arg1" || args[1] != "arg2" {
		t.Errorf("Arguments() = %v, want [arg1, arg2]", args)
	}

	// Verify Arguments returns a copy
	args[0] = "modified"
	origArgs := inv.Arguments()
	if origArgs[0] != "arg1" {
		t.Error("Arguments() should return a copy")
	}

	result, err := inv.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "inv-result" {
		t.Errorf("Proceed() = %v, want 'inv-result'", result)
	}
}

func TestInvocationImpl_SetArgs(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", []any{"old"}, nil, nil)
	inv := NewInvocation(jp, nil)

	invImpl, ok := inv.(*invocationImpl)
	if !ok {
		t.Fatal("expected *invocationImpl")
	}

	invImpl.SetArgs([]any{"new1", "new2"})
	args := invImpl.Arguments()
	if len(args) != 2 || args[0] != "new1" {
		t.Errorf("Arguments() after SetArgs = %v, want [new1, new2]", args)
	}
}

func TestInvocationImpl_SetErrorAndGetError(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", nil, nil, nil)
	inv := NewInvocation(jp, nil)

	invImpl, ok := inv.(*invocationImpl)
	if !ok {
		t.Fatal("expected *invocationImpl")
	}

	if invImpl.Error() != nil {
		t.Error("Error() should return nil initially")
	}

	sentinel := errors.New("inv error")
	invImpl.SetError(sentinel)
	if !errors.Is(invImpl.Error(), sentinel) {
		t.Errorf("Error() = %v, want %v", invImpl.Error(), sentinel)
	}
}

func TestInvocationImpl_NilProceedFunc(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", nil, nil, nil)
	inv := NewInvocation(jp, nil)

	result, err := inv.Proceed()
	if result != nil || err != nil {
		t.Errorf("nil Proceed: got (%v, %v), want (nil, nil)", result, err)
	}
}

func TestNewJoinPointWithContext_NilContext_Test(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(nil, "target", "Method", nil, nil, nil)
	if jp.Context() == nil {
		t.Fatal("nil context should fallback to context.Background()")
	}
	if jp.Context() != context.Background() {
		t.Error("nil context should fallback to context.Background()")
	}
}

func TestNewJoinPointWithContext_NilProceedFuncs(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), "target", "M", nil, nil, nil)

	result, err := jp.Proceed()
	if result != nil || err != nil {
		t.Errorf("Proceed with nil: got (%v, %v), want (nil, nil)", result, err)
	}

	result, err = jp.ProceedWithArgs([]any{"arg"})
	if result != nil || err != nil {
		t.Errorf("ProceedWithArgs with nil: got (%v, %v), want (nil, nil)", result, err)
	}
}

func TestNewInvocation_NilProceed(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", nil, nil, nil)
	inv := NewInvocation(jp, nil)

	invImpl, ok := inv.(*invocationImpl)
	if !ok {
		t.Fatal("expected *invocationImpl")
	}

	if invImpl.JoinPoint() != jp {
		t.Error("JoinPoint() should return the jp")
	}

	if len(invImpl.Arguments()) != 0 {
		t.Error("Arguments() should be empty")
	}

	result, err := invImpl.Proceed()
	if result != nil || err != nil {
		t.Errorf("nil Proceed: got (%v, %v), want (nil, nil)", result, err)
	}
}

func TestInvocationImpl_ArgumentsCopy(t *testing.T) {
	t.Parallel()

	jp := NewJoinPointWithContext(context.Background(), nil, "M", []any{"a", "b"}, nil, nil)
	inv := NewInvocation(jp, nil)

	args := inv.Arguments()
	args[0] = "modified"

	origArgs := inv.Arguments()
	if origArgs[0] != "a" {
		t.Errorf("Arguments() should return copy, got %v", origArgs[0])
	}
}

func TestNewMethodSignature_WithReflectType(t *testing.T) {
	t.Parallel()

	type Dummy struct{}
	sig := NewMethodSignature("DoStuff", reflect.TypeOf(Dummy{}))
	if sig.Name() != "DoStuff" {
		t.Errorf("Name() = %q, want %q", sig.Name(), "DoStuff")
	}
	if sig.DeclaringType() != reflect.TypeOf(Dummy{}) {
		t.Errorf("DeclaringType() = %v, want %v", sig.DeclaringType(), reflect.TypeOf(Dummy{}))
	}
}

func TestNewMethodSignature_NilReflectType(t *testing.T) {
	t.Parallel()

	sig := NewMethodSignature("Method", nil)
	if sig.Name() != "Method" {
		t.Errorf("Name() = %q, want %q", sig.Name(), "Method")
	}
	if sig.DeclaringType() != nil {
		t.Error("DeclaringType() should be nil")
	}
}
