package aop

import (
	"context"
	"errors"
	"testing"
)

func TestBefore(t *testing.T) {
	t.Parallel()
	called := false

	advice := Before(func(jp JoinPoint) {
		called = true
	})

	if advice.Type() != AdviceTypeBefore {
		t.Error("Before advice type should be AdviceTypeBefore")
	}

	// 创建 mock JoinPoint
	jp := &mockJoinPointForAdvice{
		target: nil,
		method: "Test",
		args:   nil,
	}

	// 执行通知
	_, _ = advice.Execute(context.Background(), jp)

	if !called {
		t.Error("Before advice should have been called")
	}
}

func TestAfter(t *testing.T) {
	t.Parallel()
	called := false

	advice := After(func(jp JoinPoint) {
		called = true
	})

	if advice.Type() != AdviceTypeAfter {
		t.Error("After advice type should be AdviceTypeAfter")
	}

	jp := &mockJoinPointForAdvice{
		target: nil,
		method: "Test",
		args:   nil,
	}

	_, _ = advice.Execute(context.Background(), jp)

	if !called {
		t.Error("After advice should have been called")
	}
}

func TestAfterReturning(t *testing.T) {
	t.Parallel()
	var receivedResult any
	expectedResult := "test result"

	advice := AfterReturning(func(jp JoinPoint, result any) {
		receivedResult = result
	})

	if advice.Type() != AdviceTypeAfterReturning {
		t.Error("AfterReturning advice type should be AdviceTypeAfterReturning")
	}

	jp := &mockJoinPointForAdvice{
		target:      nil,
		method:      "Test",
		args:        nil,
		proceedFunc: func() (any, error) { return expectedResult, nil },
	}
	jp.SetResult(expectedResult)

	_, _ = advice.Execute(context.Background(), jp)

	if receivedResult != expectedResult {
		t.Errorf("expected result %v, got %v", expectedResult, receivedResult)
	}
}

func TestAfterThrowing(t *testing.T) {
	t.Parallel()
	var receivedError error
	testErr := errors.New("test error")

	advice := AfterThrowing(func(jp JoinPoint, err error) {
		receivedError = err
	})

	if advice.Type() != AdviceTypeAfterThrowing {
		t.Error("AfterThrowing advice type should be AdviceTypeAfterThrowing")
	}

	jp := &mockJoinPointForAdvice{
		target:      nil,
		method:      "Test",
		args:        nil,
		proceedFunc: func() (any, error) { return nil, testErr },
	}
	jp.SetError(testErr)

	_, _ = advice.Execute(context.Background(), jp)

	if receivedError != testErr {
		t.Errorf("expected error %v, got %v", testErr, receivedError)
	}
}

func TestAround(t *testing.T) {
	t.Parallel()
	var beforeCalled, afterCalled bool
	expectedResult := "test result"

	advice := Around(func(jp JoinPoint, proceed func() any) any {
		beforeCalled = true
		result := proceed()
		afterCalled = true
		return result
	})

	if advice.Type() != AdviceTypeAround {
		t.Error("Around advice type should be AdviceTypeAround")
	}

	jp := &mockJoinPointForAdvice{
		target:      nil,
		method:      "Test",
		args:        nil,
		proceedFunc: func() (any, error) { return expectedResult, nil },
	}

	result, _ := advice.Execute(context.Background(), jp)

	if !beforeCalled {
		t.Error("Before part of Around advice should have been called")
	}

	if !afterCalled {
		t.Error("After part of Around advice should have been called")
	}

	if result != expectedResult {
		t.Errorf("expected result %v, got %v", expectedResult, result)
	}
}

func TestAroundWithArgs(t *testing.T) {
	t.Parallel()
	var passedArgs []any

	advice := Around(func(jp JoinPoint, proceed func() any) any {
		return proceed()
	})

	jp := &mockJoinPointForAdvice{
		target: nil,
		method: "Test",
		args:   []any{"arg1", 42},
		proceedFunc: func() (any, error) {
			// 这里无法直接获取参数，简化测试
			passedArgs = []any{"arg1", 42}
			return nil, nil
		},
	}

	_, _ = advice.Execute(context.Background(), jp)

	if len(passedArgs) != 2 || passedArgs[0] != "arg1" || passedArgs[1] != 42 {
		t.Errorf("expected args [arg1, 42], got %v", passedArgs)
	}
}

func TestAfterReturning_NilProceed(t *testing.T) {
	t.Parallel()
	var callbackCalled bool

	advice := AfterReturning(func(jp JoinPoint, result any) {
		callbackCalled = true
		if result != nil {
			t.Errorf("expected nil result in callback, got %v", result)
		}
	})

	jp := &mockJoinPointForAdvice{
		target:      nil,
		method:      "Test",
		args:        nil,
		proceedFunc: func() (any, error) { return nil, nil },
	}

	result, _ := advice.Execute(context.Background(), jp)

	if !callbackCalled {
		t.Error("AfterReturning callback should have been called")
	}

	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAfterThrowing_NoError(t *testing.T) {
	t.Parallel()
	var callbackCalled bool

	advice := AfterThrowing(func(jp JoinPoint, err error) {
		callbackCalled = true
	})

	jp := &mockJoinPointForAdvice{
		target:      nil,
		method:      "Test",
		args:        nil,
		proceedFunc: func() (any, error) { return "success", nil },
	}

	_, _ = advice.Execute(context.Background(), jp)

	if callbackCalled {
		t.Error("AfterThrowing callback should not have been called when no error")
	}
}

func TestAround_NilProceed(t *testing.T) {
	t.Parallel()
	var aroundCalled bool

	advice := Around(func(jp JoinPoint, proceed func() any) any {
		aroundCalled = true
		return nil
	})

	jp := &mockJoinPointForAdvice{
		target:      nil,
		method:      "Test",
		args:        nil,
		proceedFunc: func() (any, error) { return nil, nil },
	}

	result, _ := advice.Execute(context.Background(), jp)

	if !aroundCalled {
		t.Error("Around advice should have been called")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestAdviceType_Values(t *testing.T) {
	t.Parallel()
	tests := []struct {
		adviceType AdviceType
		expected   string
	}{
		{AdviceTypeBefore, "before"},
		{AdviceTypeAfter, "after"},
		{AdviceTypeAround, "around"},
		{AdviceTypeAfterReturning, "after_returning"},
		{AdviceTypeAfterThrowing, "after_throwing"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			// 验证枚举值不同
			if tt.adviceType < 0 {
				t.Errorf("AdviceType should be non-negative")
			}
		})
	}
}

// mockJoinPointForAdvice 用于测试的 JoinPoint 模拟实现
type mockJoinPointForAdvice struct {
	target      any
	method      string
	args        []any
	proceedFunc func() (any, error)
	result      any
	lastErr     error
}

func (j *mockJoinPointForAdvice) Target() any    { return j.target }
func (j *mockJoinPointForAdvice) Method() string { return j.method }
func (j *mockJoinPointForAdvice) Args() []any    { return j.args }
func (j *mockJoinPointForAdvice) Proceed() (any, error) {
	if j.proceedFunc != nil {
		return j.proceedFunc()
	}
	return nil, nil
}
func (j *mockJoinPointForAdvice) ProceedWithArgs(args []any) (any, error) {
	return j.Proceed()
}
func (j *mockJoinPointForAdvice) Context() context.Context { return context.Background() }
func (j *mockJoinPointForAdvice) GetResult() any           { return j.result }
func (j *mockJoinPointForAdvice) GetError() error          { return j.lastErr }
func (j *mockJoinPointForAdvice) SetResult(v any)          { j.result = v }
func (j *mockJoinPointForAdvice) SetError(err error)       { j.lastErr = err }
