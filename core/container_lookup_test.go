package core

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
)

// ==================== 类型安全测试 ====================

func TestHasWithTypeCheck(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("testService"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 应该存在
	if !Has[*TestService](container, "testService") {
		t.Error("Expected Has to return true for registered bean")
	}

	// 类型不匹配应该返回 false
	if Has[*TestRepository](container, "testService") {
		t.Error("Expected Has to return false for mismatched type")
	}

	// 名称不存在应该返回 false
	if Has[*TestService](container, "nonexistent") {
		t.Error("Expected Has to return false for nonexistent bean")
	}
}

func TestGetByType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册多个同类型的 Bean
	err := Register[*TestService](container, WithName[*TestService]("svc1"))
	if err != nil {
		t.Fatalf("Register svc1 failed: %v", err)
	}

	err = Register[*TestService](container, WithName[*TestService]("svc2"))
	if err != nil {
		t.Fatalf("Register svc2 failed: %v", err)
	}

	// 使用指针类型查询（与注册时一致）
	services, err := container.Get(reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("Get by type failed: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}
}

func TestGetByTypeNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_, err := container.Get(reflect.TypeOf((*TestService)(nil)).Elem())
	if err == nil {
		t.Error("Expected error when getting nonexistent type")
	}
}

func TestGetByNameNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_, err := GetByName[*TestService](container, "nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent bean")
	}
}

func TestHasTypeBasic(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 未注册时应该返回 false
	if container.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return false for unregistered type")
	}

	// 注册后应该返回 true
	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !container.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return true for registered type")
	}
}

func TestGetParent(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	// 默认没有父容器
	if child.(ContainerExt).GetParent() != nil {
		t.Error("Expected parent to be nil by default")
	}

	// 设置父容器
	child.(ContainerExt).SetParent(parent)

	// 应该能获取到父容器
	if child.(ContainerExt).GetParent() != parent {
		t.Error("Expected to get parent container")
	}
}

func TestMustGet(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "must-get"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 成功获取
	svc := MustGet[*TestService](container, "test")
	if svc.Name != "must-get" {
		t.Errorf("Expected name 'must-get', got '%s'", svc.Name)
	}

	// 失败时应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nonexistent bean")
		}
	}()
	MustGet[*TestService](container, "nonexistent")
}

func TestGetByNameByType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("svc1"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 不指定名称，按类型获取
	svc, err := GetByName[*TestService](container, "")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestGetAllWithErrors(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册一个会失败的 Bean
	err := Register[*TestService](container, WithName[*TestService]("fail"), WithFactory[*TestService](func(c ...any) (any, error) {
		return nil, fmt.Errorf("factory error")
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// GetAll 应该跳过错误的 Bean
	all := container.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected 0 beans (factory error), got %d", len(all))
	}
}

func TestResolveBeanNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 尝试获取不存在的 Bean
	_, err := container.GetByTypeAndName("nonexistent", reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error for nonexistent bean")
	}
}

func TestCreateAndInitializeCached(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := int32(0)
	err := Register[*TestService](container, WithName[*TestService]("cached"), WithFactory[*TestService](func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return &TestService{Name: "cached"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 第一次获取
	svc1, err := GetByName[*TestService](container, "cached")
	if err != nil {
		t.Fatalf("First GetByName failed: %v", err)
	}

	// 第二次获取（应该从缓存）
	svc2, err := GetByName[*TestService](container, "cached")
	if err != nil {
		t.Fatalf("Second GetByName failed: %v", err)
	}

	// 应该是同一个实例
	if svc1 != svc2 {
		t.Error("Expected same cached instance")
	}

	// 工厂只应该被调用一次
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected factory to be called once, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestGetDestroyedContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Get 应该报错
	_, err = container.Get(reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error when getting from destroyed container")
	}

	// 销毁后 GetByTypeAndName 应该报错
	_, err = container.GetByTypeAndName("test", reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error when getting by name from destroyed container")
	}

	// 销毁后 GetAll 应该返回空
	all := container.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected empty list from destroyed container, got %d", len(all))
	}
}

func TestInitializeDestroyedContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后初始化应该报错
	err = container.Initialize()
	if err == nil {
		t.Error("Expected error when initializing destroyed container")
	}
}

func TestRegisterBeanWithCustomType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 使用 WithType 覆盖类型
	err := Register[*TestService](container,
		WithName[*TestService]("customType"),
		WithType[*TestService](reflect.TypeOf((*TestRepository)(nil))),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "custom"}, nil
		}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 应该能通过覆盖后的类型获取
	services, err := container.Get(reflect.TypeOf((*TestRepository)(nil)))
	if err != nil {
		t.Fatalf("Get by custom type failed: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
}

func TestRegisterPrimaryBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册非首选 Bean
	err := Register[*TestService](container, WithName[*TestService]("secondary"))
	if err != nil {
		t.Fatalf("Register secondary failed: %v", err)
	}

	// 注册首选 Bean
	err = Register[*TestService](container, WithName[*TestService]("primary"), WithPrimary[*TestService](true))
	if err != nil {
		t.Fatalf("Register primary failed: %v", err)
	}

	// 检查 Bean 数量
	ext := container.(ContainerExt)
	if ext.BeanCountType(reflect.TypeOf((*TestService)(nil))) != 2 {
		t.Errorf("Expected 2 beans, got %d", ext.BeanCountType(reflect.TypeOf((*TestService)(nil))))
	}
}

func TestGetWithParentContainer(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	err := Register[*TestService](parent, WithName[*TestService]("parentBean"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "from-parent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register to parent failed: %v", err)
	}

	child.(ContainerExt).SetParent(parent)

	// 子容器应该能获取父容器的 Bean
	svc, err := GetByName[*TestService](child, "parentBean")
	if err != nil {
		t.Fatalf("Get from parent failed: %v", err)
	}

	if svc.Name != "from-parent" {
		t.Errorf("Expected name 'from-parent', got '%s'", svc.Name)
	}
}

func TestTypes(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc"))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	types := container.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(types))
	}
}

func TestTypesEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	types := container.Types()
	if len(types) != 0 {
		t.Errorf("Expected 0 types in empty container, got %d", len(types))
	}
}

func TestTypesWithMultipleSameTypeBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))
	_ = Register[*TestService](container, WithName[*TestService]("svc2"))
	_ = Register[*TestService](container, WithName[*TestService]("svc3"))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	types := container.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(types))
	}
}

func TestTypesAfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc"))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Bean 定义仍然存在
	types := container.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 types after destroy, got %d", len(types))
	}
}

func TestTypesWithInterfaceBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[fmt.Stringer](container, WithName[fmt.Stringer]("stringer"), WithFactory[fmt.Stringer](func(c ...any) (any, error) {
		return &TestService{Name: "test"}, nil
	}))

	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}

	// 验证类型是接口类型
	expectedType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	if types[0] != expectedType {
		t.Errorf("Expected type %v, got %v", expectedType, types[0])
	}
}

func TestTypesWithFactoryBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("factory"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "factory-created"}, nil
	}))

	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}
}

func TestTypesWithLazyBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("lazy"), WithLazy[*TestService](true))

	// 懒加载 Bean 注册后应该出现在 Types 中
	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}
}

func TestTypesWithPrototypeBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("prototype"), WithScope[*TestService]("prototype"))

	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}
}
