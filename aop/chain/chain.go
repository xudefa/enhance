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
	ctx       context.Context
	proceed   func() (any, error)
	result    any
	lastErr   error
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
	return extractResults(results)
}

// ProceedWithArgs 实现 aop.JoinPoint.ProceedWithArgs
func (m *MethodInvocation) ProceedWithArgs(args []any) (any, error) {
	if m.proceed != nil {
		// 通过 proceed 函数执行时，更新 ArgsList 以传递新参数
		m.ArgsList = make([]reflect.Value, len(args))
		for i, arg := range args {
			m.ArgsList[i] = reflect.ValueOf(arg)
		}
		return m.proceed()
	}

	callArgs := make([]reflect.Value, 0, len(args)+1)
	callArgs = append(callArgs, reflect.ValueOf(m.TargetObj))
	for _, arg := range args {
		callArgs = append(callArgs, reflect.ValueOf(arg))
	}

	results := m.MethodObj.Func.Call(callArgs)
	return extractResults(results)
}

func (m *MethodInvocation) GetResult() any {
	return m.result
}

func (m *MethodInvocation) GetError() error {
	return m.lastErr
}

func (m *MethodInvocation) SetResult(v any) {
	m.result = v
}

func (m *MethodInvocation) SetError(err error) {
	m.lastErr = err
}

// Context 实现 aop.JoinPoint.Context
func (m *MethodInvocation) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

// extractResults 从反射返回值中提取结果和错误
func extractResults(results []reflect.Value) (any, error) {
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0].Interface(), nil
	}

	// 检查最后一个返回值是否为 error
	lastResult := results[len(results)-1]
	if lastResult.Type().Implements(reflect.TypeFor[error]()) {
		if err, _ := lastResult.Interface().(error); err != nil {
			return nil, err
		}
		// error 为 nil，收集前面的非 error 返回值
		if len(results) == 2 {
			return results[0].Interface(), nil
		}
		nonErrResults := make([]any, 0, len(results)-1)
		for _, r := range results[:len(results)-1] {
			nonErrResults = append(nonErrResults, r.Interface())
		}
		return nonErrResults, nil
	}

	// 没有 error 返回值，收集所有结果
	allResults := make([]any, 0, len(results))
	for _, r := range results {
		allResults = append(allResults, r.Interface())
	}
	return allResults, nil
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
//
// 逐个跳过与当前方法不匹配的 Advisor，只执行切点匹配的通知。
func (c *AdviceChain) executeNext(inv *MethodInvocation) (any, error) {
	for inv.Index < len(c.advisors) {
		advisor := c.advisors[inv.Index]
		inv.Index++

		if !advisor.Matches(inv.Method()) {
			continue
		}

		ctx := inv.Context()
		return advisor.Advice.Execute(ctx, inv)
	}
	return inv.Proceed()
}
