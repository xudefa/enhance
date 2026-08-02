// Package aop 提供面向切面编程（AOP）支持。
// InterfaceProxyWrapper 结构体定义在 doc.go 中，此处为实现。
package aop

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// InterfaceProxyWrapper 接口代理包装器（实现 doc.go 中定义的结构体）。
//
// 通过反射转发所有接口方法调用，支持 AOP 切面织入。
// 由于 Go 运行时无法动态替换接口方法，此包装器提供显式的 Invoke/InvokeContext 方法。
//
// 使用方式:
//
//	wrapper := aop.NewInterfaceProxyWrapper(target, advisors, iface)
//	result, err := wrapper.InvokeContext(ctx, "MethodName", arg1, arg2)
//
// 设计模式: Proxy
type InterfaceProxyWrapper struct {
	target      any
	advisors    []*AspectMeta
	iface       reflect.Type
	methodCache map[string]reflect.Method
	cacheMu     sync.RWMutex
	executor    ChainExecutor
}

// NewInterfaceProxyWrapper 创建接口代理包装器。
func NewInterfaceProxyWrapper(target any, advisors []*AspectMeta, iface reflect.Type) *InterfaceProxyWrapper {
	copied := make([]*AspectMeta, len(advisors))
	copy(copied, advisors)
	return &InterfaceProxyWrapper{
		target:      target,
		advisors:    copied,
		iface:       iface,
		methodCache: make(map[string]reflect.Method),
	}
}

// Invoke 调用接口方法
func (w *InterfaceProxyWrapper) Invoke(methodName string, args ...any) (any, error) {
	return w.InvokeContext(context.Background(), methodName, args...)
}

// InvokeContext 带上下文的方法调用
func (w *InterfaceProxyWrapper) InvokeContext(ctx context.Context, methodName string, args ...any) (any, error) {
	if _, err := w.getMethod(methodName); err != nil {
		return nil, err
	}

	targetFunc := func(callArgs ...any) any {
		results := w.invokeTarget(methodName, callArgs)
		switch len(results) {
		case 0:
			return nil
		case 1:
			return results[0].Interface()
		default:
			ret := make([]any, len(results))
			for i, r := range results {
				ret[i] = r.Interface()
			}
			return ret
		}
	}

	proceedFn := func() (any, error) {
		return extractResult(targetFunc(args...))
	}
	joinPoint := NewJoinPointWithContext(ctx, w.target, methodName, args, proceedFn, nil)
	inv := NewInvocation(joinPoint, proceedFn)

	result := w.getExecutor().Execute(inv, w.advisors, targetFunc)
	if jp := inv.JoinPoint(); jp != nil {
		if err := jp.GetError(); err != nil {
			return result, err
		}
	}
	return result, nil
}

// GetTarget 获取原始目标对象
func (w *InterfaceProxyWrapper) GetTarget() any {
	return w.target
}

// GetAdvisors 获取切面列表
func (w *InterfaceProxyWrapper) GetAdvisors() []*AspectMeta {
	return w.advisors
}

// SetExecutor 设置通知链执行器
func (w *InterfaceProxyWrapper) SetExecutor(executor ChainExecutor) {
	w.executor = executor
}

// getExecutor 获取执行器，优先使用自定义执行器
func (w *InterfaceProxyWrapper) getExecutor() ChainExecutor {
	if w.executor != nil {
		return w.executor
	}
	return getDefaultExecutor()
}

// invokeTarget 通过反射调用目标对象的方法。
//
// 注意：不能使用接口类型上 MethodByName 返回的 reflect.Method.Func，
// 因为对于接口类型该方法值为零值 reflect.Value，调用会 panic。
// 这里使用目标对象值上的 MethodByName 获取绑定方法，并处理指针/值接收者场景。
func (w *InterfaceProxyWrapper) invokeTarget(methodName string, callArgs []any) []reflect.Value {
	targetVal := reflect.ValueOf(w.target)
	methodVal := targetVal.MethodByName(methodName)
	if !methodVal.IsValid() && targetVal.Kind() == reflect.Pointer {
		methodVal = targetVal.Elem().MethodByName(methodName)
	}
	if !methodVal.IsValid() {
		panic(fmt.Sprintf("method %s not found on target %T", methodName, w.target))
	}

	in := make([]reflect.Value, 0, len(callArgs))
	for _, a := range callArgs {
		in = append(in, reflect.ValueOf(a))
	}
	return methodVal.Call(in)
}

// getMethod 获取接口方法（带缓存）
func (w *InterfaceProxyWrapper) getMethod(methodName string) (reflect.Method, error) {
	w.cacheMu.RLock()
	if method, ok := w.methodCache[methodName]; ok {
		w.cacheMu.RUnlock()
		return method, nil
	}
	w.cacheMu.RUnlock()

	method, ok := w.iface.MethodByName(methodName)
	if !ok {
		return reflect.Method{}, fmt.Errorf("method %s not found on interface %s", methodName, w.iface.Name())
	}

	w.cacheMu.Lock()
	w.methodCache[methodName] = method
	w.cacheMu.Unlock()

	return method, nil
}

// extractResult 从目标函数返回值中提取结果与错误。
//
// 支持以下返回形态：
//   - 单个返回值：直接作为结果返回
//   - 单个 error 返回值：作为错误返回
//   - 多返回值：末尾为 error 时拆分结果与错误
func extractResult(result any) (any, error) {
	switch v := result.(type) {
	case nil:
		return nil, nil
	case []any:
		if len(v) == 0 {
			return nil, nil
		}
		last := v[len(v)-1]
		if err, ok := last.(error); ok {
			switch len(v) {
			case 1:
				return nil, err
			case 2:
				return v[0], err
			default:
				rest := make([]any, len(v)-1)
				copy(rest, v[:len(v)-1])
				return rest, err
			}
		}
		if len(v) == 1 {
			return v[0], nil
		}
		return v, nil
	default:
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	}
}
