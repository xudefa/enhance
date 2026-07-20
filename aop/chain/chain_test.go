package chain

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/aop"
)

type mockAdvice struct {
	called bool
}

func (m *mockAdvice) Type() aop.AdviceType {
	return aop.AdviceAround
}

func (m *mockAdvice) Apply(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
	m.called = true
	return proceed()
}

type mockTarget struct{}

func (m *mockTarget) DoWork() string {
	return "work"
}

func TestAdviceChain_Proceed(t *testing.T) {
	t.Parallel()
	advice1 := &mockAdvice{}
	advice2 := &mockAdvice{}

	advisors := []Advisor{
		{Pointcut: aop.MatchByName("Do.*"), Advice: advice1},
		{Pointcut: aop.MatchByName("Do.*"), Advice: advice2},
	}

	chain := NewAdviceChain(advisors)

	target := &mockTarget{}
	method := reflect.ValueOf(target).MethodByName("DoWork")
	methodVal := method.Interface().(func() string)

	// 使用反射调用，DoWork 无参数
	methodType := reflect.TypeOf(methodVal)
	methodReflect := reflect.ValueOf(methodVal)

	invocation := chain.CreateInvocation(target, reflect.Method{
		Name: "DoWork",
		Type: methodType,
		Func: methodReflect,
	}, []reflect.Value{})

	result := invocation.Proceed()

	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
	if !advice1.called {
		t.Error("expected first advice to be called")
	}
	if !advice2.called {
		t.Error("expected second advice to be called")
	}
}

func TestAdviceChain_Matches(t *testing.T) {
	t.Parallel()
	pc := aop.MatchByRegex("Do.*")
	method, _ := reflect.TypeOf(&mockTarget{}).MethodByName("DoWork")

	if !pc.MatchMethod(method) {
		t.Error("expected pointcut to match DoWork")
	}
}
