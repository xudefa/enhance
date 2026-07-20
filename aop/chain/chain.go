package chain

import (
	"context"
	"reflect"

	"github.com/xudefa/enhance/aop"
)

// Advisor 顾问
type Advisor struct {
	Pointcut aop.PointCut
	Advice   aop.Advice
}

// Matches 检查方法是否匹配
func (a *Advisor) Matches(method reflect.Method) bool {
	return a.Pointcut.MatchMethod(method)
}

// MethodInvocation 方法调用实现
type MethodInvocation struct {
	TargetObj any
	MethodObj reflect.Method
	ArgsList  []reflect.Value
	Chain     *AdviceChain
	Index     int
}

// Method 实现 aop.JoinPoint.Method
func (m *MethodInvocation) Method() any {
	return m.MethodObj.Func
}

// Args 实现 aop.JoinPoint.Args
func (m *MethodInvocation) Args() []any {
	args := make([]any, len(m.ArgsList))
	for i := range m.ArgsList {
		args[i] = m.ArgsList[i].Interface()
	}
	return args
}

// Signature 实现 aop.JoinPoint.Signature
func (m *MethodInvocation) Signature() aop.MethodSignature {
	return aop.NewMethodSignature(m.MethodObj.Name, m.MethodObj.Type)
}

// This 实现 aop.JoinPoint.This
func (m *MethodInvocation) This() any {
	return nil
}

// Target 实现 aop.JoinPoint.Target
func (m *MethodInvocation) Target() any {
	return m.TargetObj
}

// Context 实现 aop.JoinPoint.Context
func (m *MethodInvocation) Context() context.Context {
	return context.Background()
}

// Proceed 实现 aop.Invocation.Proceed
func (m *MethodInvocation) Proceed(args ...any) any {
	if m.Index >= len(m.Chain.advisors) {
		callArgs := m.ArgsList
		if len(args) > 0 {
			callArgs = make([]reflect.Value, len(args))
			for i, arg := range args {
				callArgs[i] = reflect.ValueOf(arg)
			}
		}
		results := m.MethodObj.Func.Call(callArgs)
		if len(results) == 0 {
			return nil
		}
		if len(results) == 1 {
			return results[0].Interface()
		}
		lastResult := results[len(results)-1]
		if lastResult.Type().Implements(reflect.TypeFor[error]()) {
			if !lastResult.IsNil() {
				return lastResult.Interface().(error)
			}
		}
		return results[0].Interface()
	}

	advisor := m.Chain.advisors[m.Index]
	m.Index++

	return advisor.Advice.Apply(m, m.Proceed)
}

// SetContext 实现 aop.Invocation.SetContext
func (m *MethodInvocation) SetContext(ctx context.Context) {
	// no-op for now
}

// AdviceChain 通知链
type AdviceChain struct {
	advisors []Advisor
}

// NewAdviceChain 创建通知链
func NewAdviceChain(advisors []Advisor) *AdviceChain {
	return &AdviceChain{
		advisors: advisors,
	}
}

// CreateInvocation 创建方法调用
func (c *AdviceChain) CreateInvocation(target any, method reflect.Method, args []reflect.Value) *MethodInvocation {
	return &MethodInvocation{
		TargetObj: target,
		MethodObj: method,
		ArgsList:  args,
		Chain:     c,
		Index:     0,
	}
}
