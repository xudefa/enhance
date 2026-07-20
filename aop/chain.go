// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"context"
	"reflect"
)

// Method 获取方法。
func (m *MethodInvocation) Method() any {
	return m.Func
}

// Args 获取参数
func (m *MethodInvocation) Args() []any {
	return m.Params
}

// This 获取代理对象
func (m *MethodInvocation) This() any {
	return m.Proxy
}

// Target 获取目标对象
func (m *MethodInvocation) Target() any {
	return m.Object
}

// Signature 获取方法签名
func (m *MethodInvocation) Signature() MethodSignature {
	if m.Func == nil {
		return nil
	}
	fnValue := reflect.ValueOf(m.Func)
	fnType := fnValue.Type()
	return NewMethodSignature(m.MethodName, fnType)
}

// Context 获取上下文
//
// 如果已设置上下文,则返回该上下文;否则返回 context.Background()。
func (m *MethodInvocation) Context() context.Context {
	if m.Ctx != nil {
		return m.Ctx
	}
	return context.Background()
}

// SetContext 设置上下文
//
// 设置后,后续通知链中的 JoinPoint.Context() 将返回新的上下文。
func (m *MethodInvocation) SetContext(ctx context.Context) {
	m.Ctx = ctx
}

// Proceed 继续执行
//
// 如果已通过 SetProceed 设置了执行函数，则调用它；
// 否则通过反射调用目标方法。
func (m *MethodInvocation) Proceed(args ...any) any {
	if m.proceed != nil {
		return m.proceed(args...)
	}
	return m.callMethod(args...)
}

// SetProceed 设置继续执行函数
//
// 在Around通知中,用于设置继续执行目标方法或下一个通知的函数。
func (m *MethodInvocation) SetProceed(p ProceedFunc) {
	m.proceed = p
}

// callMethod 通过反射调用目标方法
//
// 性能优化：
//   - 预分配 reflect.Value 切片，避免循环中多次分配
//   - 使用切片容量预计算，减少扩容开销
//
// 如果未传入参数,则使用 MethodInvocation.Params 作为调用参数。
// 返回值根据结果数量处理:无返回值返回 nil,单个返回值直接返回,多个返回值返回 []any。
func (m *MethodInvocation) callMethod(args ...any) any {
	if m.Func == nil {
		return nil
	}
	methodValue := reflect.ValueOf(m.Func)
	if methodValue.Kind() != reflect.Func {
		return nil
	}

	callArgs := args
	if len(callArgs) == 0 {
		callArgs = m.Params
	}

	// 预分配切片容量，避免循环中多次扩容
	numArgs := len(callArgs)
	argValues := make([]reflect.Value, numArgs)
	for i, arg := range callArgs {
		argValues[i] = reflect.ValueOf(arg)
	}

	results := methodValue.Call(argValues)
	numResults := len(results)
	switch numResults {
	case 0:
		return nil
	case 1:
		return results[0].Interface()
	default:
		// 预分配结果切片
		resultSlice := make([]any, numResults)
		for i, r := range results {
			resultSlice[i] = r.Interface()
		}
		return resultSlice
	}
}

// ExecuteChain 执行通知链
//
// 为代码生成的代理类提供通知链执行功能。
// 按照切面的 Order 排序，通过默认 ChainExecutor 执行通知链。
//
// 参数:
//   - jp: 方法调用信息
//   - aspects: 切面元数据列表（指针类型）
//
// 返回值:
//   - any: 方法执行结果
func ExecuteChain(jp *MethodInvocation, aspects []*AspectMeta) any {
	if jp == nil || jp.Func == nil {
		return nil
	}

	SortAspectsByOrder(aspects)

	targetFunc := func(args ...any) any {
		return jp.callMethod(args...)
	}

	return getDefaultExecutor().Execute(jp, aspects, targetFunc)
}
