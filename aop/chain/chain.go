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
func (a *Advisor) Matches(methodName string) bool {
	return a.Pointcut.Matches(nil, methodName)
}

// MethodInvocation 方法调用实现
type MethodInvocation struct {
	TargetObj any
	MethodObj reflect.Method
	ArgsList  []reflect.Value
	Chain     *AdviceChain
	Index     int
	proceed   func() (any, error)
}

// Target 实现 aop.JoinPoint.Target
func (m *MethodInvocation) Target() any {
	return m.TargetObj
}

// Method 实现 aop.JoinPoint.Method
func (m *MethodInvocation) Method() string {
	return m.MethodObj.Name
}

// Args 实现 aop.JoinPoint.Args
func (m *MethodInvocation) Args() []any {
	args := make([]any, len(m.ArgsList))
	for i := range m.ArgsList {
		args[i] = m.ArgsList[i].Interface()
	}
	return args
}

// Proceed 实现 aop.JoinPoint.Proceed
func (m *MethodInvocation) Proceed() (any, error) {
	if m.proceed != nil {
		return m.proceed()
	}

	callArgs := make([]reflect.Value, 0, len(m.ArgsList)+1)
	callArgs = append(callArgs, reflect.ValueOf(m.TargetObj))
	callArgs = append(callArgs, m.ArgsList...)
	results := m.MethodObj.Func.Call(callArgs)
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0].Interface(), nil
	}

	// 检查最后一个返回值是否为 error
	lastResult := results[len(results)-1]
	if lastResult.Type().Implements(reflect.TypeFor[error]()) {
		if !lastResult.IsNil() {
			return nil, lastResult.Interface().(error)
		}
	}
	return results[0].Interface(), nil
}

// ProceedWithArgs 实现 aop.JoinPoint.ProceedWithArgs
func (m *MethodInvocation) ProceedWithArgs(args []any) (any, error) {
	callArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		callArgs[i] = reflect.ValueOf(arg)
	}

	results := m.MethodObj.Func.Call(callArgs)
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0].Interface(), nil
	}

	lastResult := results[len(results)-1]
	if lastResult.Type().Implements(reflect.TypeFor[error]()) {
		if !lastResult.IsNil() {
			return nil, lastResult.Interface().(error)
		}
	}
	return results[0].Interface(), nil
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

// Execute 执行通知链
func (c *AdviceChain) Execute(target any, method reflect.Method, args []reflect.Value) (any, error) {
	inv := &MethodInvocation{
		TargetObj: target,
		MethodObj: method,
		ArgsList:  args,
		Chain:     c,
		Index:     0,
	}

	return c.executeNext(inv)
}

// executeNext 递归执行下一个通知
func (c *AdviceChain) executeNext(inv *MethodInvocation) (any, error) {
	if inv.Index >= len(c.advisors) {
		return inv.Proceed()
	}

	advisor := c.advisors[inv.Index]
	inv.Index++

	ctx := context.Background()
	return advisor.Advice.Execute(ctx, inv)
}
