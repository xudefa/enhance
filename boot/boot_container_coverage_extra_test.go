package boot

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
)

// TestBoot_ProvideReflect_Coverage 测试 ProvideReflect 方法
func TestBoot_ProvideReflect(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	// 使用 ProvideReflect 创建 Bean
	provider := ProvideReflect(func() (*TestBean, error) {
		return &TestBean{Name: "reflect-bean"}, nil
	})

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvideReflect failed: %v", err)
	}

	// 验证 Bean 已注册
	if !core.Has[*TestBean](container, "") {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_ProvideReflect_WithDependency_Coverage 测试 ProvideReflect 带依赖注入
func TestBoot_ProvideReflect_WithDependency(t *testing.T) {
	t.Parallel()

	type Config struct {
		Value string
	}

	type Service struct {
		Config *Config
	}

	container := core.NewContainer()

	// 先注册 Config
	if err := container.RegisterInstance(&Config{Value: "test-config"}, reflect.TypeOf(&Config{})); err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	// 使用 ProvideReflect 创建带依赖的 Service
	provider := ProvideReflect(func(c *Config) (*Service, error) {
		return &Service{Config: c}, nil
	})

	if err := provider(container); err != nil {
		t.Fatalf("ProvideReflect with dependency failed: %v", err)
	}

	// 验证 Service 已注册
	if !core.Has[*Service](container, "") {
		t.Error("Expected Service to exist")
	}
}

// TestBoot_ProvideReflect_ErrorReturn_Coverage 测试 ProvideReflect 返回错误的情况
func TestBoot_ProvideReflect_ErrorReturn(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	// 使用 ProvideReflect 创建会返回错误的 Bean
	provider := ProvideReflect(func() (*TestBean, error) {
		return nil, fmt.Errorf("construction failed")
	})

	container := core.NewContainer()
	// 注册 Bean
	if err := provider(container); err != nil {
		t.Fatalf("Register should succeed, got error: %v", err)
	}

	// 获取 Bean 时才会执行工厂函数，此时会返回错误
	_, err := core.GetByName[*TestBean](container, "")
	if err == nil {
		t.Fatal("Expected error when construction fails")
	}
}

// TestBoot_ProvideReflect_MissingDependency_Coverage 测试 ProvideReflect 依赖找不到的情况
func TestBoot_ProvideReflect_MissingDependency(t *testing.T) {
	t.Parallel()

	type Config struct {
		Value string
	}

	type Service struct {
		Config *Config
	}

	container := core.NewContainer()

	// 不注册 Config，直接创建需要 Config 的 Service
	provider := ProvideReflect(func(c *Config) (*Service, error) {
		return &Service{Config: c}, nil
	})

	// 注册 Bean 应该成功
	if err := provider(container); err != nil {
		t.Fatalf("Register should succeed, got error: %v", err)
	}

	// 获取 Bean 时才会执行依赖注入，此时会返回依赖找不到的错误
	_, err := core.GetByName[*Service](container, "")
	if err == nil {
		t.Fatal("Expected error when dependency is missing")
	}
}

// TestBoot_ProvideBean_Coverage 测试 ProvideBean 方法
func TestBoot_ProvideBean(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	instance := &TestBean{Name: "test-instance"}
	provider := ProvideBean(instance)

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvideBean failed: %v", err)
	}

	// 验证 Bean 已注册
	if !core.Has[*TestBean](container, "") {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_ProvideBean_WithOpts_Coverage 测试 ProvideBean 带选项
func TestBoot_ProvideBean_WithOpts(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	instance := &TestBean{Name: "test-instance-with-opts"}
	provider := ProvideBean(instance, core.WithPrimary[TestBean](true))

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvideBean with opts failed: %v", err)
	}

	// 验证 Bean 已注册
	beans := container.ListBeans()
	if len(beans) == 0 {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_ProvideFactory_WithOpts_Coverage 测试 ProvideFactory 带选项
func TestBoot_ProvideFactory_WithOpts(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	provider := ProvideFactory(func(c core.Container) (*TestBean, error) {
		return &TestBean{Name: "factory-bean-with-opts"}, nil
	}, core.WithPrimary[*TestBean](true))

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvideFactory with opts failed: %v", err)
	}

	// 验证 Bean 已注册
	beans := container.ListBeans()
	if len(beans) == 0 {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_ProvideNamed_WithOpts_Coverage 测试 ProvideNamed 带选项
func TestBoot_ProvideNamed_WithOpts(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	provider := ProvideNamed("my-bean-opts", &TestBean{Name: "named-bean-with-opts"}, core.WithPrimary[*TestBean](true))

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvideNamed with opts failed: %v", err)
	}

	// 验证 Bean 已注册
	beans := container.ListBeans()
	if len(beans) == 0 {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_ProvidePrimary_WithOpts_Coverage 测试 ProvidePrimary 带选项
func TestBoot_ProvidePrimary_WithOpts(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	provider := ProvidePrimary(&TestBean{Name: "primary-bean-with-opts"}, core.WithLazy[*TestBean](true))

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvidePrimary with opts failed: %v", err)
	}

	// 验证 Bean 已注册
	beans := container.ListBeans()
	if len(beans) == 0 {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_ProvideFactory_Coverage 测试 ProvideFactory 方法
func TestBoot_ProvideFactory(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	provider := ProvideFactory(func(c core.Container) (*TestBean, error) {
		return &TestBean{Name: "factory-bean"}, nil
	})

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvideFactory failed: %v", err)
	}

	// 验证 Bean 已注册
	if !core.Has[*TestBean](container, "") {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_ProvideNamed_Coverage 测试 ProvideNamed 方法
func TestBoot_ProvideNamed(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	provider := ProvideNamed("my-bean", &TestBean{Name: "named-bean"})

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvideNamed failed: %v", err)
	}

	// 验证 Bean 已注册（通过名称检查）
	beans := container.ListBeans()
	found := false
	for beanName := range beans {
		if beanName == "my-bean" || beanName == "*TestBean#my-bean" || beanName == "github.com/xudefa/enhance/boot.TestBean#my-bean" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected TestBean to exist with name 'my-bean', got beans: %v", beans)
	}
}

// TestBoot_ProvidePrimary_Coverage 测试 ProvidePrimary 方法
func TestBoot_ProvidePrimary(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	provider := ProvidePrimary(&TestBean{Name: "primary-bean"})

	container := core.NewContainer()
	if err := provider(container); err != nil {
		t.Fatalf("ProvidePrimary failed: %v", err)
	}

	// 验证 Bean 已注册
	beans := container.ListBeans()
	if len(beans) == 0 {
		t.Error("Expected TestBean to exist")
	}
}

// TestBoot_Invoke_Coverage 测试 Invoke 方法
func TestBoot_Invoke(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	invoked := false
	module := NewModule().
		Name("invoke-module").
		Bean(Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Name: "invoke-bean"}, nil
		})).
		Invoke(func(bean *TestBean) error {
			invoked = true
			if bean.Name != "invoke-bean" {
				t.Errorf("Expected bean name 'invoke-bean', got %s", bean.Name)
			}
			return nil
		}).
		Build()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(module),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	if !invoked {
		t.Error("Expected invoke to be called")
	}
}

// TestBoot_IsNilReflectValue_Coverage 测试 isNilReflectValue 函数
func TestBoot_IsNilReflectValue(t *testing.T) {
	t.Parallel()

	testBootIsNilReflectValueInvalid(t)
	testBootIsNilReflectValuePointers(t)
	testBootIsNilReflectValueCollections(t)
}

func testBootIsNilReflectValueInvalid(t *testing.T) {
	t.Helper()
	if !isNilReflectValue(reflect.Value{}) {
		t.Error("Expected invalid Value to be nil")
	}
}

func testBootIsNilReflectValuePointers(t *testing.T) {
	t.Helper()
	var ptr *int
	if !isNilReflectValue(reflect.ValueOf(ptr)) {
		t.Error("Expected nil pointer to be nil")
	}

	nonNilVal := 42
	if isNilReflectValue(reflect.ValueOf(&nonNilVal)) {
		t.Error("Expected non-nil pointer to not be nil")
	}

	var iface interface{}
	if !isNilReflectValue(reflect.ValueOf(iface)) {
		t.Error("Expected nil interface to be nil")
	}

	iface = 42
	if isNilReflectValue(reflect.ValueOf(iface)) {
		t.Error("Expected non-nil interface to not be nil")
	}
}

func testBootIsNilReflectValueCollections(t *testing.T) {
	t.Helper()
	var m map[string]int
	if !isNilReflectValue(reflect.ValueOf(m)) {
		t.Error("Expected nil map to be nil")
	}

	var s []int
	if !isNilReflectValue(reflect.ValueOf(s)) {
		t.Error("Expected nil slice to be nil")
	}

	var ch chan int
	if !isNilReflectValue(reflect.ValueOf(ch)) {
		t.Error("Expected nil chan to be nil")
	}

	var fn func()
	if !isNilReflectValue(reflect.ValueOf(fn)) {
		t.Error("Expected nil func to be nil")
	}
}

// TestBoot_IsNilReflectValue_InterfaceWithTypedNil_Coverage 测试接口包裹 typed-nil 的情况
func TestBoot_IsNilReflectValue_InterfaceWithTypedNil(t *testing.T) {
	t.Parallel()

	// 使用 error 接口测试 typed-nil
	var err error = nil
	reflVal := reflect.ValueOf(err)
	if !isNilReflectValue(reflVal) {
		t.Error("Expected nil error interface to be nil")
	}

	// 测试非 nil 的 error 接口
	err = fmt.Errorf("test error")
	reflVal = reflect.ValueOf(err)
	if isNilReflectValue(reflVal) {
		t.Error("Expected non-nil error interface to not be nil")
	}
}
