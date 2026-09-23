package core

import (
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

func TestGenerateBeanIDWithEmptyCustomName(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf((*TestService)(nil))

	// 测试空字符串作为自定义名称
	beanID := container.Generate(typ, "")
	expected := "github.com/xudefa/enhance/core.TestService"
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

func TestParseBeanIDWithNoDot(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 测试没有点号的 Bean ID
	pkg, typ, custom := container.Parse("SimpleType")
	if pkg != "" {
		t.Errorf("Expected empty pkg, got '%s'", pkg)
	}
	if typ != "SimpleType" {
		t.Errorf("Expected type 'SimpleType', got '%s'", typ)
	}
	if custom != "" {
		t.Errorf("Expected empty custom, got '%s'", custom)
	}
}

func TestParseBeanIDWithCustomName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 测试带自定义名称的 Bean ID
	pkg, typ, custom := container.Parse("github.com/example.MyType#custom")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "MyType" {
		t.Errorf("Expected type 'MyType', got '%s'", typ)
	}
	if custom != "custom" {
		t.Errorf("Expected custom 'custom', got '%s'", custom)
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

func TestCreateBeanNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_, err := container.CreateBean("nonexistent")
	if err == nil {
		t.Error("Expected error when creating nonexistent bean")
	}
}

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

// ==================== 并发测试 ====================

func TestConcurrentGetBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("concurrent"))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := GetByName[*TestService](container, "concurrent")
			if err != nil {
				t.Errorf("Concurrent GetByName failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestConcurrentInitialize(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc"))

	var wg sync.WaitGroup
	var initCount atomic.Int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := container.Initialize()
			if err == nil {
				initCount.Add(1)
			}
		}()
	}
	wg.Wait()

	// 只有第一次 Initialize 应该成功
	if initCount.Load() != 1 {
		t.Errorf("Expected exactly 1 successful Initialize, got %d", initCount.Load())
	}
}

// ==================== 懒加载测试 ====================

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
		t.Error("Expected lazy bean to not be initialized during Initialize")
	}

	// 获取 Bean 时才应该初始化
	_, err = GetByName[*TestService](container, "lazy")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if !initCalled {
		t.Error("Expected lazy bean to be initialized when accessed")
	}
}

// ==================== 辅助函数测试 ====================

func TestToPtrType_FromValueType(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(TestService{})
	ptrType := toPtrType(typ)
	if ptrType.Kind() != reflect.Ptr {
		t.Errorf("expected pointer kind, got %v", ptrType.Kind())
	}
}

func TestToPtrType_AlreadyPointer(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf((*TestService)(nil))
	ptrType := toPtrType(typ)
	if ptrType != typ {
		t.Errorf("expected same type for already-pointer, got %v != %v", ptrType, typ)
	}
}

func TestContainer_GetAll(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))
	_ = Register[*TestService](container, WithName[*TestService]("svc2"))

	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	instances := container.GetAll()
	if len(instances) < 2 {
		t.Errorf("expected at least 2 instances, got %d", len(instances))
	}
}

func TestContainer_GetAll_AfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	container.Destroy()

	instances := container.GetAll()
	if len(instances) != 0 {
		t.Errorf("expected empty slice after destroy, got %d", len(instances))
	}
}

func TestContainer_HasType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container)

	typ := reflect.TypeOf((*TestService)(nil))
	if !container.HasType(typ) {
		t.Error("expected HasType to return true for registered type")
	}

	otherTyp := reflect.TypeOf((*testing.T)(nil))
	if container.HasType(otherTyp) {
		t.Error("expected HasType to return false for unregistered type")
	}
}

func TestContainer_ListBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("listTest"))

	beans := container.ListBeans()
	if len(beans) == 0 {
		t.Error("expected non-empty beans map")
	}
}

func TestContainer_SetParent_GetParent(t *testing.T) {
	t.Parallel()
	parent := NewContainer().(*defaultContainer)
	child := NewContainer().(*defaultContainer)

	child.SetParent(parent)
	if child.GetParent() != parent {
		t.Error("expected GetParent to return the set parent container")
	}

	if parent.GetParent() != nil {
		t.Error("expected nil parent for root container")
	}
}

func TestContainer_Types(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container)

	types := container.Types()
	if len(types) == 0 {
		t.Error("expected non-empty types list")
	}
}

func TestContainer_BeanCount(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container, WithName[*TestService]("count1"))
	_ = Register[*TestService](container, WithName[*TestService]("count2"))

	count := container.BeanCount()
	if count < 2 {
		t.Errorf("expected at least 2 beans, got %d", count)
	}
}

func TestContainer_BeanCountType(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container, WithName[*TestService]("type1"))
	_ = Register[*TestService](container, WithName[*TestService]("type2"))

	typ := reflect.TypeOf((*TestService)(nil))
	count := container.BeanCountType(typ)
	if count < 2 {
		t.Errorf("expected at least 2 beans of type TestService, got %d", count)
	}
}

func TestContainer_Validate_OK(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*TestService](container)

	err := container.Validate()
	if err != nil {
		t.Errorf("expected no validation error, got %v", err)
	}
}

func TestContainer_Validate_AfterInit(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)
	container.Initialize()

	err := container.Validate()
	if err == nil {
		t.Error("expected error when validating after initialization")
	}
}

func TestContainer_Validate_CircularDependency(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	_ = Register[*A](container)
	_ = Register[*B](container)

	err := container.Validate()
	if err == nil {
		t.Error("expected circular dependency error")
	}
}

func TestContainer_Validate_MissingDependency(t *testing.T) {
	t.Parallel()
	container := NewContainer().(*defaultContainer)

	type Dep struct{}
	type Owner struct {
		D *Dep `inject:""`
	}
	_ = Register[*Owner](container)

	err := container.Validate()
	if err == nil {
		t.Error("expected missing dependency error")
	}
}
