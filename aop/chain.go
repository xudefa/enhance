// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"context"
	"reflect"
	"slices"
)

// Target 获取目标对象
func (m *MethodInvocation) Target() any {
	return m.Object
}

// Method 获取方法名
func (m *MethodInvocation) Method() string {
	return m.MethodName
}

// Args 获取参数
func (m *MethodInvocation) Args() []any {
	return m.Params
}

// Proceed 执行原方法
func (m *MethodInvocation) Proceed() (any, error) {
	if m.proceed != nil {
		result := m.proceed(m.Params...)
		return result, nil
	}
	return m.callMethod(m.Params...), nil
}

// ProceedWithArgs 带参数执行原方法
func (m *MethodInvocation) ProceedWithArgs(args []any) (any, error) {
	if m.proceed != nil {
		result := m.proceed(args...)
		return result, nil
	}
	return m.callMethod(args...), nil
}

func (m *MethodInvocation) GetResult() any     { return m.result }
func (m *MethodInvocation) GetError() error    { return m.lastErr }
func (m *MethodInvocation) SetResult(v any)    { m.result = v }
func (m *MethodInvocation) SetError(err error) { m.lastErr = err }

// Context 获取上下文
func (m *MethodInvocation) Context() context.Context {
	if m.Ctx != nil {
		return m.Ctx
	}
	return context.Background()
}

// Arguments 实现 Invocation.Arguments
func (m *MethodInvocation) Arguments() []any {
	return m.Params
}

// SetArgs 设置参数，供通知链 ProceedWithArgs 使用。
func (m *MethodInvocation) SetArgs(args []any) {
	m.Params = args
}

// JoinPoint 实现 Invocation.JoinPoint
func (m *MethodInvocation) JoinPoint() JoinPoint {
	return m
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

	// 复制切面列表后再排序，避免修改调用方共享切片导致的数据竞争与顺序污染
	sorted := slices.Clone(aspects)
	SortAspectsByOrder(sorted)

	targetFunc := func(args ...any) any {
		return jp.callMethod(args...)
	}

	return getDefaultExecutor().Execute(jp, sorted, targetFunc)
}
