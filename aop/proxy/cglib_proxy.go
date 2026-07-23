package proxy

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/xudefa/enhance/aop"
)

// CglibProxy CGLIB 代理。
//
// 基于结构体嵌入实现的代理，支持非接口方法。
// 通过嵌入目标结构体并重写方法，实现 AOP 增强。
//
// 使用方式:
//
//	handler := &MyHandler{}
//	proxy := proxy.NewCglibProxy(&MyService{}, handler)
//	result, err := proxy.Invoke(target, "DoSomething", []any{arg1, arg2})
//
// 线程安全：是
type CglibProxy struct {
	target      any
	handler     InvocationHandler
	targetType  reflect.Type
	methodCache map[string]reflect.Method
	cacheMu     sync.RWMutex
}

// NewCglibProxy 创建 CGLIB 代理。
//
// 参数:
//   - target: 目标对象（必须是结构体指针）
//   - handler: 调用处理器
//
// 返回值:
//   - *CglibProxy: 代理实例
//
// panic: 如果 target 不是结构体指针
func NewCglibProxy(target any, handler InvocationHandler) *CglibProxy {
	if target == nil {
		panic("proxy: target cannot be nil")
	}
	if handler == nil {
		panic("proxy: handler cannot be nil")
	}

	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	if targetType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("proxy: target must be a struct pointer, got %s", targetType.Kind()))
	}

	return &CglibProxy{
		target:      target,
		handler:     handler,
		targetType:  targetType,
		methodCache: make(map[string]reflect.Method),
	}
}

// Invoke 调用代理方法。
//
// 通过反射查找结构体方法，并委托给 InvocationHandler 处理。
//
// 参数:
//   - target: 目标对象
//   - method: 方法名
//   - args: 方法参数列表
//
// 返回值:
//   - any: 方法返回值
//   - error: 调用错误
func (p *CglibProxy) Invoke(target any, method string, args []any) (any, error) {
	if p.handler == nil {
		return nil, fmt.Errorf("proxy: invocation handler is nil")
	}

	return p.handler.Invoke(target, method, args)
}

// GetTarget 获取原始目标对象。
func (p *CglibProxy) GetTarget() any {
	return p.target
}

// GetHandler 获取调用处理器。
func (p *CglibProxy) GetHandler() InvocationHandler {
	return p.handler
}

// GetTargetType 获取目标类型。
func (p *CglibProxy) GetTargetType() reflect.Type {
	return p.targetType
}

// GetMethod 获取结构体方法（带缓存）。
//
// 查找指针接收者方法。
//
// 参数:
//   - name: 方法名
//
// 返回值:
//   - reflect.Method: 方法信息
//   - error: 方法不存在时返回错误
func (p *CglibProxy) GetMethod(name string) (reflect.Method, error) {
	p.cacheMu.RLock()
	if method, ok := p.methodCache[name]; ok {
		p.cacheMu.RUnlock()
		return method, nil
	}
	p.cacheMu.RUnlock()

	ptrType := reflect.PointerTo(p.targetType)
	method, ok := ptrType.MethodByName(name)
	if !ok {
		return reflect.Method{}, fmt.Errorf("proxy: method %s not found on %s", name, p.targetType.Name())
	}

	p.cacheMu.Lock()
	p.methodCache[name] = method
	p.cacheMu.Unlock()

	return method, nil
}

// InvokeMethod 直接通过反射调用结构体方法（不经过 handler）。
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
func (p *CglibProxy) InvokeMethod(name string, args []any) (any, error) {
	method, err := p.GetMethod(name)
	if err != nil {
		return nil, err
	}

	in := make([]reflect.Value, 0, len(args)+1)
	in = append(in, reflect.ValueOf(p.target))
	for _, arg := range args {
		in = append(in, reflect.ValueOf(arg))
	}

	results := method.Func.Call(in)

	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		if isNilable(results[0]) && results[0].IsNil() {
			return nil, nil
		}
		return results[0].Interface(), nil
	default:
		if len(results) == 2 && results[1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			var errVal error
			if results[1].IsValid() && (!isNilable(results[1]) || !results[1].IsNil()) {
				errVal = results[1].Interface().(error)
			}
			if isNilable(results[0]) && results[0].IsNil() {
				return nil, errVal
			}
			return results[0].Interface(), errVal
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

// isNilable reports whether reflect.Value.IsNil can be called without panicking.
func isNilable(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	}
	return false
}

// EnsureCglibProxy 确保对象是 CglibProxy。
//
// 返回值:
//   - *CglibProxy: 代理实例
//   - bool: 是否为 CglibProxy
func EnsureCglibProxy(obj any) (*CglibProxy, bool) {
	proxy, ok := obj.(*CglibProxy)
	return proxy, ok
}

// IsCglibProxy 检查对象是否为 CglibProxy。
func IsCglibProxy(obj any) bool {
	_, ok := obj.(*CglibProxy)
	return ok
}

// Target 获取原始目标对象。
// 实现 Proxy 接口。
func (p *CglibProxy) Target() any {
	return p.target
}

// AdvisedAdvisors 获取已应用的通知器列表。
// 实现 Proxy 接口。CglibProxy 不直接管理 advisors，返回空切片。
func (p *CglibProxy) AdvisedAdvisors() []aop.Advisor {
	return nil
}
