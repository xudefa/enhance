package core

import (
	"fmt"
	"reflect"
	"testing"
)

func TestHasWithNonExistentBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 不存在的 Bean 应该返回 false
	if container.Has("nonexistent", reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected Has to return false for non-existent bean")
	}
}

func TestHasWithWrongType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 名称存在但类型不匹配应该返回 false
	if container.Has("test", reflect.TypeOf((*TestRepository)(nil))) {
		t.Error("Expected Has to return false for mismatched type")
	}
}

func TestGetByNameMultipleBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("svc1"))
	if err != nil {
		t.Fatalf("Register svc1 failed: %v", err)
	}

	err = Register[*TestService](container, WithName[*TestService]("svc2"))
	if err != nil {
		t.Fatalf("Register svc2 failed: %v", err)
	}

	// 不指定名称时应该返回第一个
	svc, err := GetByName[*TestService](container, "")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestInitializeWithLazyBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestService](container,
		WithName[*TestService]("lazy"),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "lazy"}, nil
		}),
		WithInit[*TestService](func(bean any) error {
			initCalled = true
			return nil
		}),
		WithLazy[*TestService](true))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Initialize 不应该触发懒加载 Bean 的初始化
	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if initCalled {
		t.Error("Expected lazy bean not to be initialized during Initialize")
	}
}

func TestInitializeWithPrototypeBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestService](container,
		WithName[*TestService]("prototype"),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "prototype"}, nil
		}),
		WithInit[*TestService](func(bean any) error {
			initCalled = true
			return nil
		}),
		WithScope[*TestService]("prototype"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Initialize 不应该初始化 prototype Bean
	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if initCalled {
		t.Error("Expected prototype bean not to be initialized during Initialize")
	}
}

func TestHasTypeWithMultipleTypes(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	if !container.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return true for TestService")
	}

	if !container.HasType(reflect.TypeOf((*TestRepository)(nil))) {
		t.Error("Expected HasType to return true for TestRepository")
	}
}

func TestBeanCountTypeWithPrototypeScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	_ = Register[*TestService](container, WithName[*TestService]("proto1"), WithScope[*TestService]("prototype"))
	_ = Register[*TestService](container, WithName[*TestService]("proto2"), WithScope[*TestService]("prototype"))

	count := ext.BeanCountType(reflect.TypeOf((*TestService)(nil)))
	if count != 2 {
		t.Errorf("Expected 2 prototype beans, got %d", count)
	}
}

func TestHasTypeWithInterfaceType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册接口类型的 Bean
	_ = Register[fmt.Stringer](container, WithName[fmt.Stringer]("stringer"), WithFactory[fmt.Stringer](func(c ...any) (any, error) {
		return &TestService{Name: "test"}, nil
	}))

	if !container.HasType(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
		t.Error("Expected HasType to return true for interface type")
	}
}

func TestBeanCountTypeAfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))
	_ = Register[*TestService](container, WithName[*TestService]("svc2"))

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Bean 定义仍然存在，但实例被清除
	count := ext.BeanCountType(reflect.TypeOf((*TestService)(nil)))
	if count != 2 {
		t.Errorf("Expected 2 bean defs after destroy, got %d", count)
	}
}

func TestHasTypeWithParentContainer(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	_ = Register[*TestService](parent, WithName[*TestService]("parentSvc"))

	child := NewContainer()
	child.(ContainerExt).SetParent(parent)

	// HasType 只检查当前容器，不检查父容器
	if child.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return false for child container (does not check parent)")
	}

	// 但可以通过父容器获取
	if !parent.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return true for parent container")
	}
}

func TestBeanCountType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))
	_ = Register[*TestService](container, WithName[*TestService]("svc2"))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	// 测试类型计数
	svcCount := ext.BeanCountType(reflect.TypeOf((*TestService)(nil)))
	if svcCount != 2 {
		t.Errorf("Expected 2 TestService beans, got %d", svcCount)
	}

	repoCount := ext.BeanCountType(reflect.TypeOf((*TestRepository)(nil)))
	if repoCount != 1 {
		t.Errorf("Expected 1 TestRepository bean, got %d", repoCount)
	}

	// 测试不存在的类型
	nonExistentCount := ext.BeanCountType(reflect.TypeOf((*string)(nil)))
	if nonExistentCount != 0 {
		t.Errorf("Expected 0 for non-existent type, got %d", nonExistentCount)
	}
}
