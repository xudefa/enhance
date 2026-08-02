package proxy

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/xudefa/enhance/aop"
)

// JdkDynamicProxy JDK 动态代理。
//
// 基于反射实现的动态代理，支持接口代理。
// 通过 InvocationHandler 拦截方法调用，实现 AOP 增强。
//
// 使用方式:
//
//	handler := &MyHandler{}
//	proxy := proxy.NewJdkDynamicProxy(target, handler)
//	result, err := proxy.Invoke(target, "DoSomething", []any{arg1, arg2})
//
// 线程安全：是
type JdkDynamicProxy struct {
	target      any
	handler     InvocationHandler
	iface       reflect.Type
	advisors    []aop.Advisor
	methodCache map[string]reflect.Method
	cacheMu     sync.RWMutex
}

// NewJdkDynamicProxy 创建 JDK 动态代理。
//
// 参数:
//   - target: 目标对象（必须实现 iface 指定的接口）
//   - iface: 接口类型（通过 reflect.Type 传递）
//   - handler: 调用处理器
//
// 返回值:
//   - *JdkDynamicProxy: 代理实例
//
// 使用示例:
//
//	var svc TestService = &TestServiceImpl{}
//	proxy := proxy.NewJdkDynamicProxy(svc, reflect.TypeOf((*TestService)(nil)).Elem(), handler)
//
// panic: 如果 target 或 iface 为 nil
func NewJdkDynamicProxy(target any, iface reflect.Type, handler InvocationHandler, advisors ...aop.Advisor) *JdkDynamicProxy {
	if target == nil {
		panic("proxy: target cannot be nil")
	}
	if iface == nil {
		panic("proxy: iface cannot be nil")
	}
	if handler == nil {
		panic("proxy: handler cannot be nil")
	}

	if iface.Kind() != reflect.Interface {
		panic(fmt.Sprintf("proxy: iface must be an interface type, got %s", iface.Kind()))
	}

	return &JdkDynamicProxy{
		target:      target,
		handler:     handler,
		iface:       iface,
		advisors:    advisors,
		methodCache: make(map[string]reflect.Method),
	}
}

// Invoke 调用代理方法。
//
// 通过反射查找接口方法，并委托给 InvocationHandler 处理。
//
// 参数:
//   - target: 目标对象
//   - method: 方法名
//   - args: 方法参数列表
//
// 返回值:
//   - any: 方法返回值
//   - error: 调用错误
func (p *JdkDynamicProxy) Invoke(target any, method string, args []any) (any, error) {
	if p.handler == nil {
		return nil, fmt.Errorf("proxy: invocation handler is nil")
	}

	return p.handler.Invoke(target, method, args)
}

// GetTarget 获取原始目标对象。
func (p *JdkDynamicProxy) GetTarget() any {
	return p.target
}

// GetHandler 获取调用处理器。
func (p *JdkDynamicProxy) GetHandler() InvocationHandler {
	return p.handler
}

// GetIface 获取接口类型。
func (p *JdkDynamicProxy) GetIface() reflect.Type {
	return p.iface
}

// GetMethod 获取接口方法（带缓存）。
//
// 参数:
//   - name: 方法名
//
// 返回值:
//   - reflect.Method: 方法信息
//   - error: 方法不存在时返回错误
func (p *JdkDynamicProxy) GetMethod(name string) (reflect.Method, error) {
	p.cacheMu.RLock()
	if method, ok := p.methodCache[name]; ok {
		p.cacheMu.RUnlock()
		return method, nil
	}
	p.cacheMu.RUnlock()

	method, ok := p.iface.MethodByName(name)
	if !ok {
		return reflect.Method{}, fmt.Errorf("proxy: method %s not found on %s", name, p.iface.Name())
	}

	p.cacheMu.Lock()
	p.methodCache[name] = method
	p.cacheMu.Unlock()

	return method, nil
}

// InvokeMethod 直接通过反射调用接口方法（不经过 handler）。
//
// 用于在 InvocationHandler 中调用原始方法。
//
// 参数:
//   - name: 方法名
//   - args: 方法参数列表
//
// 返回值:
//   - any: 方法返回值
//   - error: 调用错误
func (p *JdkDynamicProxy) InvokeMethod(name string, args []any) (any, error) {
	// 验证方法存在于接口上
	if _, err := p.GetMethod(name); err != nil {
		return nil, err
	}

	// 通过 target 的 reflect.Value 调用方法
	targetVal := reflect.ValueOf(p.target)
	method := targetVal.MethodByName(name)
	if !method.IsValid() {
		return nil, fmt.Errorf("proxy: method %s not found on target", name)
	}

	in := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		in = append(in, reflect.ValueOf(arg))
	}

	results := method.Call(in)

	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		// 单个 error 返回值必须作为错误返回，而非普通结果
		if results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if results[0].IsValid() && (!isNilable(results[0]) || !results[0].IsNil()) {
				if err, ok := results[0].Interface().(error); ok {
					return nil, err
				}
			}
			return nil, nil
		}
		if isNilable(results[0]) && results[0].IsNil() {
			return nil, nil
		}
		return results[0].Interface(), nil
	default:
		if len(results) == 2 && results[1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			var errVal error
			if results[1].IsValid() && (!isNilable(results[1]) || !results[1].IsNil()) {
				errVal, _ = results[1].Interface().(error)
			}
			if isNilable(results[0]) && results[0].IsNil() {
				return nil, errVal
			}
			return results[0].Interface(), errVal
		}
		// 多个返回值且末尾为 error 时，拆分结果与错误
		if results[len(results)-1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			var errVal error
			if results[len(results)-1].IsValid() && (!isNilable(results[len(results)-1]) || !results[len(results)-1].IsNil()) {
				errVal, _ = results[len(results)-1].Interface().(error)
			}
			ret := make([]any, len(results)-1)
			for i, r := range results[:len(results)-1] {
				if r.IsValid() && (!isNilable(r) || !r.IsNil()) {
					ret[i] = r.Interface()
				}
			}
			return ret, errVal
		}
		ret := make([]any, len(results))
		for i, r := range results {
			if r.IsValid() && (!isNilable(r) || !r.IsNil()) {
				ret[i] = r.Interface()
			}
		}
		return ret, nil
	}
}

// EnsureJdkDynamicProxy 确保对象是 JdkDynamicProxy。
//
// 返回值:
//   - *JdkDynamicProxy: 代理实例
//   - bool: 是否为 JdkDynamicProxy
func EnsureJdkDynamicProxy(obj any) (*JdkDynamicProxy, bool) {
	proxy, ok := obj.(*JdkDynamicProxy)
	return proxy, ok
}

// IsJdkDynamicProxy 检查对象是否为 JdkDynamicProxy。
func IsJdkDynamicProxy(obj any) bool {
	_, ok := obj.(*JdkDynamicProxy)
	return ok
}

// Target 获取原始目标对象。
// 实现 Proxy 接口。
func (p *JdkDynamicProxy) Target() any {
	return p.target
}

// AdvisedAdvisors 获取已应用的通知器列表。
// 实现 Proxy 接口。
func (p *JdkDynamicProxy) AdvisedAdvisors() []aop.Advisor {
	return p.advisors
}
