package core

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/xudefa/enhance/core/registry"
)

// ==================== BeanID 生成测试 ====================

func TestGenerateBeanID(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf((*TestService)(nil)).Elem()

	// 测试无自定义名称
	beanID := container.Generate(typ)
	expected := "github.com/xudefa/enhance/core.TestService"
	if beanID != expected {
		t.Errorf("Expected %s, got %s", expected, beanID)
	}

	// 测试有自定义名称
	beanID = container.Generate(typ, "myBean")
	expected = "github.com/xudefa/enhance/core.TestService#myBean"
	if beanID != expected {
		t.Errorf("Expected %s, got %s", expected, beanID)
	}

	// 测试已有标准格式
	beanID = container.Generate(typ, "github.com/xudefa/enhance/core.TestService#custom")
	expected = "github.com/xudefa/enhance/core.TestService#custom"
	if beanID != expected {
		t.Errorf("Expected %s, got %s", expected, beanID)
	}
}

func TestParseBeanID(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 测试标准格式
	pkg, typ, custom := container.Parse("github.com/example.MyType#myBean")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "MyType" {
		t.Errorf("Expected type 'MyType', got '%s'", typ)
	}
	if custom != "myBean" {
		t.Errorf("Expected custom 'myBean', got '%s'", custom)
	}

	// 测试无自定义名称
	pkg, typ, custom = container.Parse("github.com/example.MyType")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "MyType" {
		t.Errorf("Expected type 'MyType', got '%s'", typ)
	}
	if custom != "" {
		t.Errorf("Expected empty custom, got '%s'", custom)
	}
}

// ==================== 注册边界条件测试 ====================

func TestRegisterNilType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	def := registry.BeanDef{
		Type: nil,
	}
	err := container.RegisterBean(def)
	if err == nil {
		t.Error("Expected error for nil type")
	}
}

func TestRegisterDuplicateBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("duplicate"))
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	// 重复注册应该不报错（幂等）
	err = Register[*TestService](container, WithName[*TestService]("duplicate"))
	if err != nil {
		t.Errorf("Expected no error for duplicate registration, got: %v", err)
	}
}

func TestRegisterAfterInitialize(t *testing.T) {
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

	// 初始化后再注册应该报错
	err = Register[*TestService](container, WithName[*TestService]("test2"))
	if err == nil {
		t.Error("Expected error when registering after initialization")
	}
}

func TestRegisterWithEmptyName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 不提供名称，应该使用默认生成
	err := Register[*TestService](container)
	if err != nil {
		t.Fatalf("Register with empty name failed: %v", err)
	}

	// 应该能通过类型获取（使用指针类型）
	services, err := container.Get(reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("Get by type failed: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
}

// ==================== 作用域测试 ====================

func TestSingletonScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := int32(0)
	err := Register[*TestService](container, WithName[*TestService]("singleton"), WithFactory[*TestService](func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return &TestService{Name: "singleton"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 并发获取多次
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc, err := GetByName[*TestService](container, "singleton")
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if svc.Name != "singleton" {
				t.Errorf("Expected name 'singleton', got '%s'", svc.Name)
			}
		}()
	}
	wg.Wait()

	// 工厂只应该被调用一次
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected factory to be called once, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestPrototypeScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := int32(0)
	err := Register[*TestService](container, WithName[*TestService]("prototype"), WithFactory[*TestService](func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return &TestService{Name: fmt.Sprintf("instance-%d", atomic.LoadInt32(&callCount))}, nil
	}), WithScope[*TestService]("prototype"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 获取多次
	svc1, _ := GetByName[*TestService](container, "prototype")
	svc2, _ := GetByName[*TestService](container, "prototype")
	svc3, _ := GetByName[*TestService](container, "prototype")

	// 每次应该不同
	if svc1 == svc2 || svc2 == svc3 || svc1 == svc3 {
		t.Error("Expected different instances for prototype scope")
	}

	// 工厂应该被调用 3 次
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("Expected factory to be called 3 times, got %d", atomic.LoadInt32(&callCount))
	}
}

// ==================== 并发安全测试 ====================

func TestConcurrentRegister(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent%d", i)
			err := Register[*TestService](container, WithName[*TestService](name), WithFactory[*TestService](func(c ...any) (any, error) {
				return &TestService{Name: name}, nil
			}))
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent register failed: %v", err)
	}

	// 验证所有 Bean 都注册成功
	ext := container.(ContainerExt)
	count := ext.BeanCount()
	if count != 10 {
		t.Errorf("Expected 10 beans, got %d", count)
	}
}

func TestConcurrentGetAndInitialize(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("concurrent"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "concurrent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 20)

	// 并发初始化和获取
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := container.Initialize(); err != nil && err != ErrContainerAlreadyInitialized {
				errChan <- err
			}
		}()
		go func() {
			defer wg.Done()
			_, err := GetByName[*TestService](container, "concurrent")
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent get/initialize failed: %v", err)
	}
}

func TestFactoryCircularDependency(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*FactoryCircularA](container, WithFactory[*FactoryCircularA](func(c ...any) (any, error) {
		if _, err := container.Get(reflect.TypeOf((*FactoryCircularB)(nil))); err != nil {
			return nil, err
		}
		return &FactoryCircularA{}, nil
	}))
	_ = Register[*FactoryCircularB](container, WithFactory[*FactoryCircularB](func(c ...any) (any, error) {
		if _, err := container.Get(reflect.TypeOf((*FactoryCircularA)(nil))); err != nil {
			return nil, err
		}
		return &FactoryCircularB{}, nil
	}))

	// 工厂型循环依赖应返回 ErrCircularDependency 而非死锁
	_, err := container.Get(reflect.TypeOf((*FactoryCircularA)(nil)))
	if !errors.Is(err, ErrCircularDependency) {
		t.Fatalf("Expected ErrCircularDependency, got: %v", err)
	}
}

// ==================== BeanCount 测试 ====================

func TestBeanCount(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	if ext.BeanCount() != 0 {
		t.Errorf("Expected 0 beans, got %d", ext.BeanCount())
	}

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))
	_ = Register[*TestService](container, WithName[*TestService]("svc2"))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	if ext.BeanCount() != 3 {
		t.Errorf("Expected 3 beans, got %d", ext.BeanCount())
	}

	// 使用指针类型查询（与注册时一致）
	typ := reflect.TypeOf((*TestService)(nil))
	if ext.BeanCountType(typ) != 2 {
		t.Errorf("Expected 2 TestService beans, got %d", ext.BeanCountType(typ))
	}
}

// ==================== GetAll 测试 ====================

func TestGetAll(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc1"}, nil
	}))
	_ = Register[*TestService](container, WithName[*TestService]("svc2"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc2"}, nil
	}))

	// 初始化容器以实例化所有 Bean
	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	instances := container.GetAll()
	if len(instances) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(instances))
	}
}

// ==================== 边界条件测试 ====================

func TestEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	if ext.BeanCount() != 0 {
		t.Errorf("Expected 0 beans in empty container, got %d", ext.BeanCount())
	}

	types := ext.Types()
	if len(types) != 0 {
		t.Errorf("Expected 0 types in empty container, got %d", len(types))
	}

	instances := container.GetAll()
	if len(instances) != 0 {
		t.Errorf("Expected 0 instances in empty container, got %d", len(instances))
	}
}

func TestRegisterWithNilFactory(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 不提供 Factory，应该使用默认工厂
	err := Register[*TestService](container, WithName[*TestService]("nilFactory"))
	if err != nil {
		t.Fatalf("Register with nil factory failed: %v", err)
	}

	svc, err := GetByName[*TestService](container, "nilFactory")
	if err != nil {
		t.Fatalf("Get with nil factory failed: %v", err)
	}

	if svc == nil {
		t.Error("Expected non-nil service")
	}
}
