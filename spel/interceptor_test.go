package spel

import (
	"errors"
	"testing"
)

func TestNewInterceptorChain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		interceptors []MethodInterceptor
	}{
		{"empty", nil},
		{"single", []MethodInterceptor{NewLoggingInterceptor()}},
		{"multiple", []MethodInterceptor{NewLoggingInterceptor(), NewLoggingInterceptor()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			chain := NewInterceptorChain(tt.interceptors)
			if chain == nil {
				t.Error("expected non-nil chain")
			}
		})
	}
}

func TestInterceptorChain_AddInterceptor(t *testing.T) {
	t.Parallel()
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.AddInterceptor(NewLoggingInterceptor())
	chain.AddInterceptor(NewLoggingInterceptor())

	if len(chain.interceptors) != 2 {
		t.Errorf("expected 2 interceptors, got %d", len(chain.interceptors))
	}
}

func TestInterceptorChain_Proceed_NoInterceptors(t *testing.T) {
	t.Parallel()
	handler := func() (any, error) {
		return "done", nil
	}
	inv := NewSimpleMethodInvocation("Do", nil, "target", handler)

	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.SetInvocation(inv)
	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "done" {
		t.Errorf("expected 'done', got %v", result)
	}
}

func TestInterceptorChain_Proceed_NilHandler(t *testing.T) {
	t.Parallel()
	inv := NewSimpleMethodInvocation("Do", nil, "target", nil)
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.SetInvocation(inv)
	_, err := chain.Proceed()
	if err == nil {
		t.Error("expected error for nil handler")
	}
}

func TestInterceptorChain_Proceed_WithInterceptors(t *testing.T) {
	t.Parallel()
	var callOrder []string

	interceptor1 := &orderTrackingInterceptor{order: &callOrder, name: "i1"}
	interceptor2 := &orderTrackingInterceptor{order: &callOrder, name: "i2"}

	handler := func() (any, error) {
		callOrder = append(callOrder, "handler")
		return "result", nil
	}
	inv := NewSimpleMethodInvocation("Do", []any{1, 2}, "target", handler)
	chain := NewInterceptorChain([]MethodInterceptor{interceptor1, interceptor2}).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "result" {
		t.Errorf("expected 'result', got %v", result)
	}
	expected := []string{"i1", "i2", "handler"}
	if len(callOrder) != len(expected) {
		t.Fatalf("expected call order %v, got %v", expected, callOrder)
	}
	for i, v := range expected {
		if callOrder[i] != v {
			t.Errorf("expected callOrder[%d]=%s, got %s", i, v, callOrder[i])
		}
	}
}

func TestInterceptorChain_Invoke(t *testing.T) {
	t.Parallel()
	var callOrder []string

	interceptor := &orderTrackingInterceptor{order: &callOrder, name: "i1"}
	handler := func() (any, error) {
		callOrder = append(callOrder, "handler")
		return 42, nil
	}
	inv := NewSimpleMethodInvocation("Do", nil, "target", handler)
	chain := NewInterceptorChain([]MethodInterceptor{interceptor}).(*interceptorChainImpl)

	result, err := chain.Invoke(inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}
	expected := []string{"i1", "handler"}
	for i, v := range expected {
		if callOrder[i] != v {
			t.Errorf("expected callOrder[%d]=%s, got %s", i, v, callOrder[i])
		}
	}
}

func TestInterceptorChain_Invoke_ResetsIndex(t *testing.T) {
	t.Parallel()
	handler := func() (any, error) {
		return "ok", nil
	}
	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)

	_, err := chain.Invoke(inv)
	if err != nil {
		t.Fatalf("first invoke: %v", err)
	}

	_, err = chain.Invoke(inv)
	if err != nil {
		t.Fatalf("second invoke: %v", err)
	}
}

func TestInterceptorChain_GetMethod(t *testing.T) {
	t.Parallel()
	inv := NewSimpleMethodInvocation("MyMethod", nil, nil, nil)
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	if got := chain.GetMethod(); got != "MyMethod" {
		t.Errorf("expected 'MyMethod', got %v", got)
	}
}

func TestInterceptorChain_GetArguments(t *testing.T) {
	t.Parallel()
	args := []any{1, "two", 3.0}
	inv := NewSimpleMethodInvocation("Do", args, nil, nil)
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	got := chain.GetArguments()
	if len(got) != len(args) {
		t.Fatalf("expected %d args, got %d", len(args), len(got))
	}
	for i, v := range args {
		if got[i] != v {
			t.Errorf("expected args[%d]=%v, got %v", i, v, got[i])
		}
	}
}

func TestInterceptorChain_GetTarget(t *testing.T) {
	t.Parallel()
	target := "myTarget"
	inv := NewSimpleMethodInvocation("Do", nil, target, nil)
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	if got := chain.GetTarget(); got != target {
		t.Errorf("expected %v, got %v", target, got)
	}
}

func TestSimpleMethodInvocation_GetMethod(t *testing.T) {
	t.Parallel()
	inv := NewSimpleMethodInvocation("Foo", nil, nil, nil)
	if got := inv.GetMethod(); got != "Foo" {
		t.Errorf("expected 'Foo', got %v", got)
	}
}

func TestSimpleMethodInvocation_GetArguments(t *testing.T) {
	t.Parallel()
	args := []any{"a", 42}
	inv := NewSimpleMethodInvocation("Foo", args, nil, nil)
	got := inv.GetArguments()
	if len(got) != 2 || got[0] != "a" || got[1] != 42 {
		t.Errorf("expected [a, 42], got %v", got)
	}
}

func TestSimpleMethodInvocation_GetTarget(t *testing.T) {
	t.Parallel()
	target := struct{ Name string }{Name: "x"}
	inv := NewSimpleMethodInvocation("Foo", nil, target, nil)
	got := inv.GetTarget()
	if got != target {
		t.Errorf("expected same target pointer, got %v", got)
	}
}

func TestSimpleMethodInvocation_Proceed(t *testing.T) {
	t.Parallel()
	handler := func() (any, error) {
		return "handled", nil
	}
	inv := NewSimpleMethodInvocation("Foo", nil, nil, handler)
	result, err := inv.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "handled" {
		t.Errorf("expected 'handled', got %v", result)
	}
}

func TestSimpleMethodInvocation_Proceed_NilHandler(t *testing.T) {
	t.Parallel()
	inv := NewSimpleMethodInvocation("Foo", nil, nil, nil)
	_, err := inv.Proceed()
	if err == nil {
		t.Error("expected error for nil handler")
	}
}

func TestSimpleMethodInvocation_Proceed_Error(t *testing.T) {
	t.Parallel()
	handler := func() (any, error) {
		return nil, errors.New("boom")
	}
	inv := NewSimpleMethodInvocation("Foo", nil, nil, handler)
	_, err := inv.Proceed()
	if err == nil || err.Error() != "boom" {
		t.Errorf("expected 'boom' error, got %v", err)
	}
}

func TestNewLoggingInterceptor(t *testing.T) {
	t.Parallel()
	l := NewLoggingInterceptor()
	if l == nil {
		t.Error("expected non-nil interceptor")
	}
}

func TestLoggingInterceptor_Invoke_Success(t *testing.T) {
	t.Parallel()
	handler := func() (any, error) {
		return "log-result", nil
	}
	inv := NewSimpleMethodInvocation("LogMethod", []any{1, 2}, "tgt", handler)
	l := NewLoggingInterceptor()

	result, err := l.Invoke(inv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "log-result" {
		t.Errorf("expected 'log-result', got %v", result)
	}
}

func TestLoggingInterceptor_Invoke_Error(t *testing.T) {
	t.Parallel()
	handler := func() (any, error) {
		return nil, errors.New("log-error")
	}
	inv := NewSimpleMethodInvocation("FailMethod", nil, nil, handler)
	l := NewLoggingInterceptor()

	_, err := l.Invoke(inv)
	if err == nil || err.Error() != "log-error" {
		t.Errorf("expected 'log-error', got %v", err)
	}
}

func TestInterceptorChain_EmptyChain_ProceedToHandler(t *testing.T) {
	t.Parallel()
	called := false
	handler := func() (any, error) {
		called = true
		return nil, nil
	}
	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	_, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestInterceptorChain_SetInvocation(t *testing.T) {
	t.Parallel()
	inv1 := NewSimpleMethodInvocation("M1", nil, nil, nil)
	inv2 := NewSimpleMethodInvocation("M2", nil, nil, nil)
	chain := NewInterceptorChain(nil).(*interceptorChainImpl)

	chain.SetInvocation(inv1)
	if chain.GetMethod() != "M1" {
		t.Errorf("expected M1, got %v", chain.GetMethod())
	}

	chain.SetInvocation(inv2)
	if chain.GetMethod() != "M2" {
		t.Errorf("expected M2, got %v", chain.GetMethod())
	}
}

func TestInterceptorChain_InterceptorCanModifyResult(t *testing.T) {
	t.Parallel()
	modifier := &resultModifyingInterceptor{modified: "modified-value"}

	handler := func() (any, error) {
		return "original", nil
	}
	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)
	chain := NewInterceptorChain([]MethodInterceptor{modifier}).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "modified-value" {
		t.Errorf("expected 'modified-value', got %v", result)
	}
}

func TestInterceptorChain_InterceptorCanShortCircuit(t *testing.T) {
	t.Parallel()
	handlerCalled := false
	shortCircuit := &shortCircuitInterceptor{}

	handler := func() (any, error) {
		handlerCalled = true
		return "original", nil
	}
	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)
	chain := NewInterceptorChain([]MethodInterceptor{shortCircuit}).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "short-circuited" {
		t.Errorf("expected 'short-circuited', got %v", result)
	}
	if handlerCalled {
		t.Error("handler should not be called when short-circuited")
	}
}

type orderTrackingInterceptor struct {
	order *[]string
	name  string
}

func (o *orderTrackingInterceptor) Invoke(invocation MethodInvocation) (any, error) {
	*o.order = append(*o.order, o.name)
	return invocation.Proceed()
}

type resultModifyingInterceptor struct {
	modified any
}

func (r *resultModifyingInterceptor) Invoke(invocation MethodInvocation) (any, error) {
	_, err := invocation.Proceed()
	return r.modified, err
}

type shortCircuitInterceptor struct{}

func (s *shortCircuitInterceptor) Invoke(invocation MethodInvocation) (any, error) {
	return "short-circuited", nil
}
