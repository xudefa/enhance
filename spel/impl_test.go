package spel

import (
	"errors"
	"testing"
)

type testSettable struct {
	Name   string
	Count  int64
	Items  []string
	Value  interface{}
}

func TestNewSpelParser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{"simple property", "name", false},
		{"empty expression", "", true},
		{"literal true", "true", false},
		{"complex expr", "age > 18", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := NewSpelParser()
			expr, err := parser.ParseExpression(tt.expression)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr == nil {
				t.Error("expected non-nil expression")
			}
		})
	}
}

func TestNewInterceptorChain_WithMultipleInterceptors(t *testing.T) {
	t.Parallel()

	var order []string

	i1 := &testOrderInterceptor{name: "i1", order: &order}
	i2 := &testOrderInterceptor{name: "i2", order: &order}
	i3 := &testOrderInterceptor{name: "i3", order: &order}

	handler := func() (any, error) {
		order = append(order, "handler")
		return "result", nil
	}

	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)
	chain := NewInterceptorChain([]MethodInterceptor{i1, i2, i3}).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "result" {
		t.Errorf("expected 'result', got %v", result)
	}
	expected := []string{"i1", "i2", "i3", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d] = %s, want %s", i, order[i], v)
		}
	}
}

func TestNewInterceptorChain_EmptyProceedWithNilInvocation(t *testing.T) {
	t.Parallel()

	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestNewSimpleMethodInvocation_AllFields(t *testing.T) {
	t.Parallel()

	handler := func() (any, error) {
		return "handled", nil
	}
	inv := NewSimpleMethodInvocation("MyMethod", []any{1, "two"}, "myTarget", handler)

	if inv.GetMethod() != "MyMethod" {
		t.Errorf("expected 'MyMethod', got %v", inv.GetMethod())
	}
	if len(inv.GetArguments()) != 2 {
		t.Errorf("expected 2 args, got %d", len(inv.GetArguments()))
	}
	if inv.GetTarget() != "myTarget" {
		t.Errorf("expected 'myTarget', got %v", inv.GetTarget())
	}

	result, err := inv.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "handled" {
		t.Errorf("expected 'handled', got %v", result)
	}
}

func TestNewStandardEvaluationContext_Variables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		varName  string
		varValue any
	}{
		{"string variable", "name", "Alice"},
		{"int variable", "age", 30},
		{"bool variable", "active", true},
		{"nil variable", "empty", nil},
		{"slice variable", "items", []int{1, 2, 3}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := NewStandardEvaluationContext(nil)
			ctx.SetVariable(tt.varName, tt.varValue)

			got, ok := ctx.GetVariable(tt.varName)
			if !ok {
				t.Fatal("variable not found")
			}

			switch expected := tt.varValue.(type) {
			case []int:
				gotSlice, ok := got.([]int)
				if !ok {
					t.Fatalf("expected []int, got %T", got)
				}
				if len(gotSlice) != len(expected) {
					t.Errorf("expected %d items, got %d", len(expected), len(gotSlice))
				}
			default:
				if got != expected {
					t.Errorf("got %v, want %v", got, expected)
				}
			}
		})
	}
}

func TestNewStandardEvaluationContext_GetVariable_NotFound(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext(nil)
	_, ok := ctx.GetVariable("nonexistent")
	if ok {
		t.Error("expected variable not found")
	}
}

func TestNewStandardEvaluationContext_GetRootObject_Nil(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext(nil)
	if ctx.GetRootObject() != nil {
		t.Error("expected nil root object")
	}
}

func TestNewStandardEvaluationContext_GetPropertyAccessor(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext(nil)
	pa := ctx.GetPropertyAccessor()
	if pa == nil {
		t.Error("expected non-nil PropertyAccessor")
	}
}

func TestNewStandardEvaluationContext_SetRootObject(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext("initial")
	if ctx.GetRootObject() != "initial" {
		t.Errorf("expected 'initial', got %v", ctx.GetRootObject())
	}

	ctx.SetRootObject("updated")
	if ctx.GetRootObject() != "updated" {
		t.Errorf("expected 'updated', got %v", ctx.GetRootObject())
	}
}

func TestNewReflectPropertyAccessor_GetProperty_NonStruct(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	_, err := accessor.GetProperty("not a struct", "Name")
	if err == nil {
		t.Error("expected error for non-struct target")
	}
}

func TestNewReflectPropertyAccessor_GetProperty_Unexported(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	target := struct {
		name string //nolint:unused
	}{name: "hidden"}

	_, err := accessor.GetProperty(target, "name")
	if err == nil {
		t.Error("expected error for unexported field")
	}
}

func TestNewReflectPropertyAccessor_SetProperty_NonStruct(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	err := accessor.SetProperty("not a struct", "Name", "value")
	if err == nil {
		t.Error("expected error for non-struct target")
	}
}

func TestNewReflectPropertyAccessor_SetProperty_Unexported(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	target := struct {
		name string //nolint:unused
	}{name: "old"}

	err := accessor.SetProperty(target, "name", "new")
	if err == nil {
		t.Error("expected error for unexported field")
	}
}

func TestNewReflectPropertyAccessor_SetProperty_NonNilableType(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	target := struct {
		Value int
	}{Value: 10}

	err := accessor.SetProperty(target, "Value", nil)
	if err == nil {
		t.Error("expected error for setting nil to non-nilable type")
	}
}

func TestNewReflectPropertyAccessor_SetProperty_TypeConversion(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	target := &testSettable{Count: 0}

	err := accessor.SetProperty(target, "Count", int(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Count != 42 {
		t.Errorf("expected 42, got %d", target.Count)
	}
}

func TestNewReflectPropertyAccessor_SetProperty_TypeNotConvertible(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	target := struct {
		Name string
	}{Name: ""}

	err := accessor.SetProperty(target, "Name", 12345)
	if err == nil {
		t.Error("expected error for non-convertible type")
	}
}

func TestNewReflectPropertyAccessor_GetProperty_NilTarget(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	_, err := accessor.GetProperty(nil, "Name")
	if err == nil {
		t.Error("expected error for nil target")
	}
}

func TestNewReflectPropertyAccessor_SetProperty_NilTarget(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	err := accessor.SetProperty(nil, "Name", "value")
	if err == nil {
		t.Error("expected error for nil target")
	}
}

func TestNewReflectPropertyAccessor_SetProperty_NilToNilableType(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	target := &testSettable{Items: []string{"a"}}

	err := accessor.SetProperty(target, "Items", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Items != nil {
		t.Errorf("expected nil Items, got %v", target.Items)
	}
}

func TestNewLoggingInterceptor_Invoke(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		handler    func() (any, error)
		wantResult any
		wantErr    bool
	}{
		{
			name:       "success",
			handler:    func() (any, error) { return "result", nil },
			wantResult: "result",
			wantErr:    false,
		},
		{
			name:    "error",
			handler: func() (any, error) { return nil, errors.New("fail") },
			wantErr: true,
		},
		{
			name:       "nil result",
			handler:    func() (any, error) { return nil, nil },
			wantResult: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			inv := NewSimpleMethodInvocation("Test", []any{1, 2}, "tgt", tt.handler)
			l := NewLoggingInterceptor()

			result, err := l.Invoke(inv)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestInterceptorChain_WithInterceptorError(t *testing.T) {
	t.Parallel()

	failInterceptor := &errorInterceptor{err: errors.New("interceptor error")}
	handler := func() (any, error) {
		return "never", nil
	}

	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)
	chain := NewInterceptorChain([]MethodInterceptor{failInterceptor}).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	_, err := chain.Proceed()
	if err == nil || err.Error() != "interceptor error" {
		t.Errorf("expected 'interceptor error', got %v", err)
	}
}

type errorInterceptor struct {
	err error
}

func (e *errorInterceptor) Invoke(invocation MethodInvocation) (any, error) {
	return nil, e.err
}

type testOrderInterceptor struct {
	name  string
	order *[]string
}

func (o *testOrderInterceptor) Invoke(invocation MethodInvocation) (any, error) {
	*o.order = append(*o.order, o.name)
	return invocation.Proceed()
}

func TestNewSimpleMethodInvocation_ProceedHandlerReturnsError(t *testing.T) {
	t.Parallel()

	handler := func() (any, error) {
		return nil, errors.New("handler failed")
	}
	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)

	_, err := inv.Proceed()
	if err == nil || err.Error() != "handler failed" {
		t.Errorf("expected 'handler failed', got %v", err)
	}
}

func TestNewSimpleMethodInvocation_EmptyArgs(t *testing.T) {
	t.Parallel()

	inv := NewSimpleMethodInvocation("Do", nil, nil, nil)
	if len(inv.GetArguments()) != 0 {
		t.Errorf("expected 0 args, got %d", len(inv.GetArguments()))
	}
}

func TestInterceptorChain_Proceed_AfterAllInterceptors(t *testing.T) {
	t.Parallel()

	i1 := &testOrderInterceptor{name: "i1", order: &[]string{}}
	i2 := &testOrderInterceptor{name: "i2", order: &[]string{}}

	handlerCalled := false
	handler := func() (any, error) {
		handlerCalled = true
		return "done", nil
	}

	inv := NewSimpleMethodInvocation("Do", nil, nil, handler)
	chain := NewInterceptorChain([]MethodInterceptor{i1, i2}).(*interceptorChainImpl)
	chain.SetInvocation(inv)

	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "done" {
		t.Errorf("expected 'done', got %v", result)
	}
	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestInterceptorChain_Invoke_NilInvocation(t *testing.T) {
	t.Parallel()

	chain := NewInterceptorChain(nil).(*interceptorChainImpl)
	chain.SetInvocation(nil)

	result, err := chain.Proceed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}
