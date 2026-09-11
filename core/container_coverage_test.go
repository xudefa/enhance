package core

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core/registry"
)

// TestRegisterInstance_Coverage 测试 RegisterInstance 方法（覆盖率补充）
func TestRegisterInstance(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	instance := &TestService{Name: "direct"}
	err := container.RegisterInstance(instance, reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	svc, err := GetByName[*TestService](container, "")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if svc.Name != "direct" {
		t.Errorf("Expected name 'direct', got '%s'", svc.Name)
	}
}

// TestRegisterInstanceAfterInitialized_Coverage 测试初始化后注册实例应失败
func TestRegisterInstanceAfterInitialized(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)
	container.Initialize()

	instance := &TestService{Name: "late"}
	err := container.RegisterInstance(instance, reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error when registering instance after initialization")
	}
}

// TestCreateBean_Coverage 测试手动创建 Bean
func TestCreateBean(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container,
		WithName[*TestService]("manual"),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "created"}, nil
		}),
	)

	instance, err := container.CreateBean("github.com/xudefa/enhance/core.TestService#manual")
	if err != nil {
		t.Fatalf("CreateBean failed: %v", err)
	}

	svc, ok := instance.(*TestService)
	if !ok {
		t.Fatalf("Expected *TestService, got %T", instance)
	}

	if svc.Name != "created" {
		t.Errorf("Expected name 'created', got '%s'", svc.Name)
	}
}

// TestCreateBeanNotFound_Coverage 测试创建不存在的 Bean
func TestCreateBeanNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_, err := container.CreateBean("nonexistent")
	if err == nil {
		t.Error("Expected error when creating nonexistent bean")
	}
}

// TestCreateBeanAfterDestroyed_Coverage 测试销毁后创建 Bean 应失败
func TestCreateBeanAfterDestroyed(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container, WithName[*TestService]("svc"))
	container.Initialize()
	container.Destroy()

	_, err := container.CreateBean("github.com/xudefa/enhance/core.TestService#svc")
	if err == nil {
		t.Error("Expected error when creating bean after destroy")
	}
}

// TestResolveBean_Coverage 测试 resolveBean 方法
func TestResolveBean(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container,
		WithName[*TestService]("resolve"),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "resolved"}, nil
		}),
	)

	container.Initialize()

	// 从缓存获取
	instance, err := container.resolveBean("github.com/xudefa/enhance/core.TestService#resolve", reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("resolveBean failed: %v", err)
	}

	svc, ok := instance.(*TestService)
	if !ok {
		t.Fatalf("Expected *TestService, got %T", instance)
	}

	if svc.Name != "resolved" {
		t.Errorf("Expected name 'resolved', got '%s'", svc.Name)
	}
}

// TestResolveBeanNotFound_Coverage 测试解析不存在的 Bean
func TestResolveBeanNotFound_Coverage(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_, err := container.resolveBean("nonexistent", reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error when resolving nonexistent bean")
	}
}

// TestResolveBeanAfterDestroy_Coverage 测试销毁后解析 Bean 应失败
func TestResolveBeanAfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container, WithName[*TestService]("svc"))
	container.Initialize()
	container.Destroy()

	_, err := container.resolveBean("github.com/xudefa/enhance/core.TestService#svc", reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error when resolving bean after destroy")
	}
}

// TestRegisterBeanNilType_Coverage 测试 nil 类型注册
func TestRegisterBeanNilType(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	def := registry.BeanDef{
		Type:    nil,
		Factory: func(c ...any) (any, error) { return nil, nil },
	}

	err := container.RegisterBean(def)
	if err == nil {
		t.Error("Expected error when registering bean with nil type")
	}
}

// TestRegisterBeanNilFactory_Coverage 测试 nil 工厂注册
func TestRegisterBeanNilFactory(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	def := registry.BeanDef{
		Type:    reflect.TypeOf((*TestService)(nil)),
		Factory: nil,
	}

	err := container.RegisterBean(def)
	if err == nil {
		t.Error("Expected error when registering bean with nil factory")
	}
}

// TestCurrentGoroutineID_Coverage 测试获取 goroutine ID
func TestCurrentGoroutineID(t *testing.T) {
	t.Parallel()

	gid := currentGoroutineID()
	if gid == "" {
		t.Error("Expected non-empty goroutine ID")
	}
}
