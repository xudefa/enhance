package aop

import (
	"context"
	"testing"
)

// TestAdvice_Interface 测试 Advice 接口
func TestAdvice_Interface(t *testing.T) {
	t.Parallel()

	executed := false
	advice := NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
		executed = true
		return nil, nil
	}, 0)

	if advice.Type() != AdviceTypeBefore {
		t.Errorf("Type() = %v, want %v", advice.Type(), AdviceTypeBefore)
	}

	joinPoint := &mockJoinPointForTest{
		target: "target",
		method: "DoSomething",
		args:   []any{},
	}

	_, _ = advice.Execute(context.Background(), joinPoint)
	if !executed {
		t.Error("advice function should be executed")
	}
}

// TestPointCut_Interface 测试 PointCut 接口
func TestPointCut_Interface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pointCut   PointCut
		methodName string
		expected   bool
	}{
		{"exact match", MatchByName("DoSomething"), "DoSomething", true},
		{"no match", MatchByName("DoSomething"), "DoSomethingElse", false},
		{"prefix match", MatchByName("Do*"), "DoSomething", true},
		{"prefix no match", MatchByName("Do*"), "GetSomething", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// 测试 Matches 方法
			result := tt.pointCut.Matches(nil, tt.methodName)
			if result != tt.expected {
				t.Errorf("Matches(%q) = %v, want %v", tt.methodName, result, tt.expected)
			}
			// 测试 Expression 方法
			if tt.pointCut.Expression() == "" {
				t.Error("Expression() should not be empty")
			}
		})
	}
}

// TestAdvisor_Interface 测试 Advisor 接口
func TestAdvisor_Interface(t *testing.T) {
	t.Parallel()

	advice := NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
		return nil, nil
	}, 1)
	pointCut := MatchByName("DoSomething")

	advisor := NewAdvisor(advice, pointCut, 1)

	if advisor.Advice() != advice {
		t.Error("Advice() should return the advice")
	}

	if advisor.PointCut() != pointCut {
		t.Error("PointCut() should return the pointcut")
	}

	if advisor.Order() != 1 {
		t.Errorf("Order() = %v, want 1", advisor.Order())
	}
}

// TestJoinPoint_Interface 测试 JoinPoint 接口
func TestJoinPoint_Interface(t *testing.T) {
	t.Parallel()

	joinPoint := &mockJoinPointForTest{
		target: "target",
		method: "DoSomething",
		args:   []any{"arg1", "arg2"},
	}

	if joinPoint.Target() != "target" {
		t.Error("Target() should return target")
	}

	if joinPoint.Method() != "DoSomething" {
		t.Errorf("Method() = %v, want DoSomething", joinPoint.Method())
	}

	if len(joinPoint.Args()) != 2 {
		t.Errorf("Args() length = %v, want 2", len(joinPoint.Args()))
	}

	// 测试 Proceed 方法
	result, err := joinPoint.Proceed()
	if result != nil || err != nil {
		t.Errorf("Proceed() = (%v, %v), want (nil, nil)", result, err)
	}

	// 测试 ProceedWithArgs 方法
	result, err = joinPoint.ProceedWithArgs([]any{"newArg"})
	if result != nil || err != nil {
		t.Errorf("ProceedWithArgs() = (%v, %v), want (nil, nil)", result, err)
	}
}

// TestInvocation_Interface 测试 Invocation 接口
func TestInvocation_Interface(t *testing.T) {
	t.Parallel()

	inv := &mockInvocationForTest{
		joinPoint: &mockJoinPointForTest{
			target: "target",
			method: "DoSomething",
			args:   []any{},
		},
	}

	if inv.JoinPoint() == nil {
		t.Error("JoinPoint() should not return nil")
	}

	result, err := inv.Proceed()
	if result != nil || err != nil {
		t.Errorf("Proceed() should return (nil, nil) for mock, got (%v, %v)", result, err)
	}

	if len(inv.Arguments()) != 0 {
		t.Errorf("Arguments() should return empty slice for mock, got %v", inv.Arguments())
	}
}

// mockJoinPointForTest 测试用的 JoinPoint 模拟实现
type mockJoinPointForTest struct {
	target any
	method string
	args   []any
	ctx    context.Context
}

func (j *mockJoinPointForTest) Target() any                             { return j.target }
func (j *mockJoinPointForTest) Method() string                          { return j.method }
func (j *mockJoinPointForTest) Args() []any                             { return j.args }
func (j *mockJoinPointForTest) Proceed() (any, error)                   { return nil, nil }
func (j *mockJoinPointForTest) ProceedWithArgs(args []any) (any, error) { return nil, nil }
func (j *mockJoinPointForTest) Context() context.Context {
	if j.ctx != nil {
		return j.ctx
	}
	return context.Background()
}
func (j *mockJoinPointForTest) GetResult() any     { return nil }
func (j *mockJoinPointForTest) GetError() error    { return nil }
func (j *mockJoinPointForTest) SetResult(v any)    {}
func (j *mockJoinPointForTest) SetError(err error) {}

// mockInvocationForTest 测试用的 Invocation 模拟实现
type mockInvocationForTest struct {
	joinPoint JoinPoint
}

func (i *mockInvocationForTest) JoinPoint() JoinPoint  { return i.joinPoint }
func (i *mockInvocationForTest) Arguments() []any      { return i.joinPoint.Args() }
func (i *mockInvocationForTest) Proceed() (any, error) { return nil, nil }
