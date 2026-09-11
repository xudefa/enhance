package core

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core/registry"
)

// ==================== BeanID 补充测试 ====================

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

// ==================== Validation 测试 ====================

func TestValidateSuccess(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container)
	_ = Register[*TestRepository](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestValidateMissingDependency(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type TestBean struct {
		Service *TestService `inject:""`
	}

	_ = Register[*TestBean](container)
	// 注意：没有注册 TestService

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("Expected validation error for missing dependency")
	}
}

func TestValidateCircularDependency(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*A](container)
	_ = Register[*B](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("Expected validation error for circular dependency")
	}
}

func TestValidateAlreadyInitialized(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	_ = Register[*TestService](container)

	// 先初始化
	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ext := container.(ContainerExt)
	err = ext.Validate()
	if err == nil {
		t.Error("Expected error when validating initialized container")
	}
}

func TestValidateWithParent(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	_ = Register[*TestService](parent)

	child := NewContainer()
	ext := child.(ContainerExt)
	ext.SetParent(parent)

	type TestBean struct {
		Service *TestService `inject:""`
	}

	_ = Register[*TestBean](child)

	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed with parent: %v", err)
	}
}

func TestValidateWithEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed for empty container: %v", err)
	}
}

func TestValidateWithNonStructType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	_ = Register[*TestService](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed for non-struct type: %v", err)
	}
}

func TestValidateWithMultipleDependencies(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type Config struct {
		Timeout int `value:"timeout"`
	}

	type ServiceA struct {
		Config *Config `inject:""`
	}

	type ServiceB struct {
		Config *Config `inject:""`
	}

	_ = Register[*Config](container)
	_ = Register[*ServiceA](container)
	_ = Register[*ServiceB](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestValidateWithPartialDependencies(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type DependencyA struct {
		Name string
	}

	type DependencyB struct {
		Name string
	}

	type Service struct {
		DepA *DependencyA `inject:""`
		DepB *DependencyB `inject:""`
	}

	_ = Register[*DependencyA](container)
	_ = Register[*Service](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("Expected validation error for partial dependencies")
	}
}

// ==================== Parent Container 测试 ====================

func TestParentContainer(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	err := Register[*TestService](parent, WithName[*TestService]("parentBean"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "parent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register to parent failed: %v", err)
	}

	// 通过类型断言访问 SetParent
	childExt := child.(ContainerExt)
	childExt.SetParent(parent)

	// 子容器应该能获取父容器的 Bean
	svc, err := GetByName[*TestService](child, "parentBean")
	if err != nil {
		t.Fatalf("Get from parent failed: %v", err)
	}

	if svc.Name != "parent" {
		t.Errorf("Expected name 'parent', got '%s'", svc.Name)
	}
}

func TestParentContainerGetByType(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	err := Register[*TestService](parent, WithName[*TestService]("parentBean"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "parent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register to parent failed: %v", err)
	}

	// 通过类型断言访问 SetParent
	childExt := child.(ContainerExt)
	childExt.SetParent(parent)

	// 子容器应该能通过类型获取父容器的 Bean（使用指针类型）
	services, err := child.Get(reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("Get by type from parent failed: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service from parent, got %d", len(services))
	}
}

// ==================== Container State 测试 ====================

func TestContainerDestroyState(t *testing.T) {
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

	// 销毁后获取 Bean 应该报错
	_, err = GetByName[*TestService](container, "test")
	if err == nil {
		t.Error("Expected error when getting bean from destroyed container")
	}

	// 销毁后再次销毁应该报错
	err = container.Destroy()
	if err == nil {
		t.Error("Expected error when destroying already destroyed container")
	}

	// 销毁后初始化应该报错
	err = container.Initialize()
	if err == nil {
		t.Error("Expected error when initializing destroyed container")
	}
}

func TestContainerAlreadyInitialized(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("First Initialize failed: %v", err)
	}

	// 重复初始化应该报错
	err = container.Initialize()
	if err == nil {
		t.Error("Expected error when initializing already initialized container")
	}
}

// ==================== ListBeans & GetBeanDef 测试 ====================

func TestListBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc1"}, nil
	}))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	beanLister := container.(BeanLister)
	beanDefs := beanLister.ListBeans()
	if len(beanDefs) != 2 {
		t.Errorf("Expected 2 bean defs, got %d", len(beanDefs))
	}

	// 验证 Bean 信息
	foundSvc := false
	foundRepo := false
	for id, def := range beanDefs {
		if def.Type == nil {
			continue
		}
		typeStr := def.Type.String()
		if typeStr == "*core.TestService" {
			foundSvc = true
		}
		if typeStr == "*core.TestRepository" {
			foundRepo = true
		}
		_ = id
	}
	if !foundSvc || !foundRepo {
		t.Error("Expected to find both TestService and TestRepository beans")
	}
}

func TestListBeansEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	beanLister := container.(BeanLister)
	beanDefs := beanLister.ListBeans()
	if len(beanDefs) != 0 {
		t.Errorf("Expected 0 beans in empty container, got %d", len(beanDefs))
	}
}

func TestGetBeanDef(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("myService"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 获取存在的 BeanDef
	beanLister := container.(BeanLister)
	beanDefs := beanLister.ListBeans()
	expectedID := "github.com/xudefa/enhance/core.TestService#myService"
	def, ok := beanDefs[expectedID]
	if !ok {
		t.Error("Expected to find BeanDef")
	}
	if def == nil {
		t.Error("Expected non-nil BeanDef")
	}

	// 获取不存在的 BeanDef
	_, ok = beanDefs["nonexistent"]
	if ok {
		t.Error("Expected not to find BeanDef for nonexistent bean")
	}
}

func TestListBeansWithInitializedContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc1"}, nil
	}))

	beanLister := container.(BeanLister)

	// 初始化前
	beansBefore := beanLister.ListBeans()
	if len(beansBefore) != 1 {
		t.Errorf("Expected 1 bean before initialization, got %d", len(beansBefore))
	}

	// 初始化后
	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	beansAfter := beanLister.ListBeans()
	if len(beansAfter) != 1 {
		t.Errorf("Expected 1 bean after initialization, got %d", len(beansAfter))
	}
}

func TestListBeansWithNilType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册 nil 类型的 BeanDef 应该报错
	def := registry.BeanDef{
		Type: nil,
	}
	err := container.RegisterBean(def)
	if err == nil {
		t.Error("Expected error when registering bean with nil type")
	}
}

func TestListBeansAfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Bean 定义仍然存在
	beanLister := container.(BeanLister)
	beanDefs := beanLister.ListBeans()
	if len(beanDefs) != 1 {
		t.Errorf("Expected 1 bean def after destroy, got %d", len(beanDefs))
	}
}

func TestGetBeanDefWithMultipleBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*TestService](container, WithName[*TestService]("svc1"))
	_ = Register[*TestService](container, WithName[*TestService]("svc2"))
	_ = Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	// 获取所有 BeanDef
	beanLister := container.(BeanLister)
	defs := beanLister.ListBeans()

	if len(defs) != 3 {
		t.Errorf("Expected 3 bean defs, got %d", len(defs))
	}
}
