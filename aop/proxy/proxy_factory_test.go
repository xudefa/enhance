package proxy

import (
	"testing"

	"github.com/xudefa/enhance/aop"
)

// TestProxyFactory_CreateProxy_Interface 测试创建接口代理
func TestProxyFactory_CreateProxy_Interface(t *testing.T) {
	t.Parallel()

	factory := NewProxyFactory()
	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}

	proxy, err := factory.CreateProxy(svc, handler)
	if err != nil {
		t.Fatalf("CreateProxy() error = %v", err)
	}

	if proxy == nil {
		t.Fatal("CreateProxy() should return non-nil proxy")
	}

	if proxy.Target() != svc {
		t.Error("Target() should return original target")
	}

	advisors := proxy.AdvisedAdvisors()
	if len(advisors) != 0 {
		t.Errorf("AdvisedAdvisors() should return empty slice, got %d", len(advisors))
	}
}

// TestProxyFactory_CreateProxy_Struct 测试创建结构体代理
func TestProxyFactory_CreateProxy_Struct(t *testing.T) {
	t.Parallel()

	factory := NewProxyFactory()
	target := &TestStruct{Value: 10}
	handler := &MockHandler{}

	proxy, err := factory.CreateProxy(target, handler)
	if err != nil {
		t.Fatalf("CreateProxy() error = %v", err)
	}

	if proxy == nil {
		t.Fatal("CreateProxy() should return non-nil proxy")
	}

	if proxy.Target() != target {
		t.Error("Target() should return original target")
	}
}

// TestProxyFactory_CreateProxy_NilTarget 测试 nil 目标
func TestProxyFactory_CreateProxy_NilTarget(t *testing.T) {
	t.Parallel()

	factory := NewProxyFactory()
	handler := &MockHandler{}

	_, err := factory.CreateProxy(nil, handler)
	if err == nil {
		t.Error("CreateProxy() should return error for nil target")
	}
}

// TestProxyFactory_CreateProxy_NilHandler 测试 nil 处理器
func TestProxyFactory_CreateProxy_NilHandler(t *testing.T) {
	t.Parallel()

	factory := NewProxyFactory()
	target := &TestStruct{Value: 10}

	_, err := factory.CreateProxy(target, nil)
	if err == nil {
		t.Error("CreateProxy() should return error for nil handler")
	}
}

// TestProxyFactory_CreateProxyWithAdvisors 测试带通知器的代理创建
func TestProxyFactory_CreateProxyWithAdvisors(t *testing.T) {
	t.Parallel()

	factory := NewProxyFactory()
	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}

	advisors := []aop.Advisor{
		aop.NewAdvisor(aop.Before(func(jp aop.JoinPoint) {}), aop.MatchByName("DoSomething"), 0),
	}

	proxy, err := factory.CreateProxyWithAdvisors(svc, handler, advisors)
	if err != nil {
		t.Fatalf("CreateProxyWithAdvisors() error = %v", err)
	}

	if proxy == nil {
		t.Fatal("CreateProxyWithAdvisors() should return non-nil proxy")
	}
}

// TestProxyFactory_CreateProxy_WithIface 测试指定接口的代理创建
func TestProxyFactory_CreateProxy_WithIface(t *testing.T) {
	t.Parallel()

	iface := testIface()
	factory := NewProxyFactoryWithIface(iface)
	target := &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}

	proxy, err := factory.CreateProxy(target, handler)
	if err != nil {
		t.Fatalf("CreateProxy() error = %v", err)
	}

	if proxy == nil {
		t.Fatal("CreateProxy() should return non-nil proxy")
	}
}

// TestProxyType_Enum 测试代理类型枚举
func TestProxyType_Enum(t *testing.T) {
	t.Parallel()

	if ProxyTypeJDK != 0 {
		t.Errorf("ProxyTypeJDK should be 0, got %d", ProxyTypeJDK)
	}

	if ProxyTypeCGLIB != 1 {
		t.Errorf("ProxyTypeCGLIB should be 1, got %d", ProxyTypeCGLIB)
	}
}
