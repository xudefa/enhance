package chain

import (
	"context"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/aop"
)

type mockAdvice struct {
	called bool
}

func (m *mockAdvice) Type() aop.AdviceType {
	return aop.AdviceTypeAround
}

func (m *mockAdvice) Order() int {
	return 0
}

func (m *mockAdvice) Execute(ctx context.Context, jp aop.JoinPoint) (any, error) {
	m.called = true
	return jp.Proceed()
}

type mockTarget struct{}

func (m *mockTarget) DoWork() string {
	return "work"
}

func TestAdviceChain_Matches(t *testing.T) {
	t.Parallel()
	pc := aop.MatchByName("Do*")

	if !pc.Matches(&mockTarget{}, "DoWork") {
		t.Error("expected pointcut to match DoWork")
	}

	if pc.Matches(&mockTarget{}, "OtherMethod") {
		t.Error("expected pointcut not to match OtherMethod")
	}
}

func TestAdvisor_Matches(t *testing.T) {
	t.Parallel()
	advice := &mockAdvice{}

	advisor := Advisor{
		Pointcut: aop.MatchByName("Do*"),
		Advice:   advice,
	}

	if !advisor.Matches("DoWork") {
		t.Error("expected advisor to match DoWork")
	}

	if advisor.Matches("OtherMethod") {
		t.Error("expected advisor not to match OtherMethod")
	}
}

func TestMethodInvocation(t *testing.T) {
	t.Parallel()
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	invocation := &MethodInvocation{
		TargetObj: target,
		MethodObj: method,
		ArgsList:  []reflect.Value{},
	}

	// 测试 Target
	if invocation.Target() != target {
		t.Error("expected Target to return target object")
	}

	// 测试 Method
	if invocation.Method() != "DoWork" {
		t.Errorf("expected Method to return 'DoWork', got %v", invocation.Method())
	}

	// 测试 Args
	if len(invocation.Args()) != 0 {
		t.Error("expected Args to return empty slice")
	}

	// 测试 Proceed
	result, err := invocation.Proceed()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
}

func TestNewAdviceChain(t *testing.T) {
	t.Parallel()
	advice := &mockAdvice{}

	advisors := []Advisor{
		{Pointcut: aop.MatchByName("Do*"), Advice: advice},
	}

	chain := NewAdviceChain(advisors)
	if chain == nil {
		t.Error("expected chain to be created")
	}
}
