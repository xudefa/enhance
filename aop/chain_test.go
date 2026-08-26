package aop

import (
	"context"
	"errors"
	"testing"
)

func TestMethodInvocation_Target(t *testing.T) {
	t.Parallel()

	obj := &testTarget{}
	invocation := &MethodInvocation{
		Object: obj,
	}

	result := invocation.Target()
	if result != obj {
		t.Errorf("expected target object, got %v", result)
	}
}

func TestMethodInvocation_Method(t *testing.T) {
	t.Parallel()

	invocation := &MethodInvocation{
		MethodName: "TestMethod",
	}

	result := invocation.Method()
	if result != "TestMethod" {
		t.Errorf("expected method name 'TestMethod', got %s", result)
	}
}

func TestMethodInvocation_Args(t *testing.T) {
	t.Parallel()

	args := []any{"arg1", 42, true}
	invocation := &MethodInvocation{
		Params: args,
	}

	result := invocation.Args()
	if len(result) != 3 {
		t.Errorf("expected 3 args, got %d", len(result))
	}
	if result[0] != "arg1" {
		t.Errorf("expected first arg 'arg1', got %v", result[0])
	}
}

func TestMethodInvocation_Proceed(t *testing.T) {
	t.Parallel()

	// Test with proceed function
	callCount := 0
	invocation := &MethodInvocation{
		Params: []any{"test"},
		proceed: func(args ...any) any {
			callCount++
			return "result"
		},
	}

	result, err := invocation.Proceed()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "result" {
		t.Errorf("expected 'result', got %v", result)
	}
	if callCount != 1 {
		t.Error("expected proceed to be called once")
	}
}

func TestMethodInvocation_Proceed_WithoutProceed(t *testing.T) {
	t.Parallel()

	// Test without proceed function (uses callMethod)
	invocation := &MethodInvocation{
		Params: []any{"hello"},
		Func: func(s string) string {
			return "processed: " + s
		},
	}

	result, err := invocation.Proceed()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "processed: hello" {
		t.Errorf("expected 'processed: hello', got %v", result)
	}
}

func TestMethodInvocation_ProceedWithArgs(t *testing.T) {
	t.Parallel()

	invocation := &MethodInvocation{
		Params: []any{"original"},
		proceed: func(args ...any) any {
			return args[0].(string) + "_modified"
		},
	}

	result, err := invocation.ProceedWithArgs([]any{"new"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "new_modified" {
		t.Errorf("expected 'new_modified', got %v", result)
	}
}

func TestMethodInvocation_GetResult(t *testing.T) {
	t.Parallel()

	invocation := &MethodInvocation{
		result: "test result",
	}

	result := invocation.GetResult()
	if result != "test result" {
		t.Errorf("expected 'test result', got %v", result)
	}
}

func TestMethodInvocation_GetError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("test error")
	invocation := &MethodInvocation{
		lastErr: expectedErr,
	}

	result := invocation.GetError()
	if result != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, result)
	}
}

func TestMethodInvocation_SetResult(t *testing.T) {
	t.Parallel()

	invocation := &MethodInvocation{}
	invocation.SetResult("new result")

	if invocation.result != "new result" {
		t.Errorf("expected result 'new result', got %v", invocation.result)
	}
}

func TestMethodInvocation_SetError(t *testing.T) {
	t.Parallel()

	invocation := &MethodInvocation{}
	expectedErr := errors.New("new error")
	invocation.SetError(expectedErr)

	if invocation.lastErr != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, invocation.lastErr)
	}
}

func TestMethodInvocation_Context(t *testing.T) {
	t.Parallel()

	// Test with nil context
	invocation := &MethodInvocation{
		Ctx: nil,
	}

	ctx := invocation.Context()
	if ctx == nil {
		t.Error("expected non-nil context")
	}
	if ctx != context.Background() {
		t.Error("expected background context")
	}
}

func TestMethodInvocation_Context_WithCtx(t *testing.T) {
	t.Parallel()

	// Test with custom context
	customCtx := context.WithValue(context.Background(), "key", "value")
	invocation := &MethodInvocation{
		Ctx: customCtx,
	}

	ctx := invocation.Context()
	if ctx != customCtx {
		t.Error("expected custom context")
	}
}

func TestMethodInvocation_Arguments(t *testing.T) {
	t.Parallel()

	args := []any{"arg1", "arg2"}
	invocation := &MethodInvocation{
		Params: args,
	}

	result := invocation.Arguments()
	if len(result) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(result))
	}
}

func TestMethodInvocation_SetArgs(t *testing.T) {
	t.Parallel()

	invocation := &MethodInvocation{
		Params: []any{"old"},
	}

	newArgs := []any{"new1", "new2"}
	invocation.SetArgs(newArgs)

	if len(invocation.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(invocation.Params))
	}
	if invocation.Params[0] != "new1" {
		t.Errorf("expected first param 'new1', got %v", invocation.Params[0])
	}
}

func TestMethodInvocation_JoinPoint(t *testing.T) {
	t.Parallel()

	invocation := &MethodInvocation{}
	jp := invocation.JoinPoint()

	if jp == nil {
		t.Error("expected non-nil JoinPoint")
	}
}

func TestExecuteChain_NilInputs(t *testing.T) {
	t.Parallel()

	// Test with nil invocation
	result := ExecuteChain(nil, nil)
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}

	// Test with nil Func
	invocation := &MethodInvocation{
		Func: nil,
	}
	result = ExecuteChain(invocation, nil)
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestExecuteChain_WithAspects(t *testing.T) {
	t.Parallel()

	callCount := 0
	invocation := &MethodInvocation{
		MethodName: "TestMethod",
		Params:     []any{"test"},
		Func: func(s string) string {
			callCount++
			return "result: " + s
		},
	}

	aspects := []*AspectMeta{
		{
			Order: 1,
		},
	}

	result := ExecuteChain(invocation, aspects)
	if result == nil {
		t.Error("expected non-nil result")
	}
	if callCount != 1 {
		t.Error("expected target function to be called once")
	}
}

type testTarget struct{}
