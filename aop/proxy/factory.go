package proxy

import (
	"fmt"
	"reflect"

	"github.com/xudefa/enhance/aop"
)

// defaultProxyFactory 默认代理工厂实现。
type defaultProxyFactory struct {
	defaultIface reflect.Type
}

// NewProxyFactory 创建默认代理工厂。
func NewProxyFactory() ProxyFactory {
	return &defaultProxyFactory{}
}

// NewProxyFactoryWithIface 创建指定接口的代理工厂。
func NewProxyFactoryWithIface(iface reflect.Type) ProxyFactory {
	return &defaultProxyFactory{
		defaultIface: iface,
	}
}

// CreateProxy 创建代理对象。
func (f *defaultProxyFactory) CreateProxy(target any, handler InvocationHandler) (Proxy, error) {
	return f.CreateProxyWithAdvisors(target, handler, nil)
}

// CreateProxyWithAdvisors 创建带通知器的代理对象。
func (f *defaultProxyFactory) CreateProxyWithAdvisors(target any, handler InvocationHandler, advisors []aop.Advisor) (Proxy, error) {
	if target == nil {
		return nil, fmt.Errorf("proxy: target cannot be nil")
	}
	if handler == nil {
		return nil, fmt.Errorf("proxy: handler cannot be nil")
	}

	targetType := reflect.TypeOf(target)
	if targetType == nil {
		return nil, fmt.Errorf("proxy: target type cannot be determined")
	}

	// 如果是指针类型，获取元素类型
	originalType := targetType
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	// 检查是否为接口类型或指定了默认接口
	if targetType.Kind() == reflect.Interface || f.defaultIface != nil {
		iface := f.defaultIface
		if iface == nil {
			iface = targetType
		}
		proxy := NewJdkDynamicProxy(target, iface, handler, advisors...)
		return proxy, nil
	}

	// 使用 CGLIB 代理（结构体）
	if targetType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("proxy: target must be a struct or interface, got %s", targetType.Kind())
	}

	// 确保原始目标是指针类型
	if originalType.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("proxy: struct target must be a pointer")
	}

	proxy := NewCglibProxy(target, handler, advisors...)
	return proxy, nil
}
