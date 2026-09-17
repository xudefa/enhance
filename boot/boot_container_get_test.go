package boot

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
)

// TestBoot_GetBean_Coverage 测试通过容器获取 Bean
func TestBoot_GetBean_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(
			NewModule().
				Name("test-module").
				Bean(Provide(func(c core.Container) (TestBean, error) {
					return TestBean{Name: "test-bean"}, nil
				})).
				Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 测试通过容器获取 Bean
	bean, err := core.GetByName[TestBean](app.Container(), "")
	if err != nil {
		t.Fatalf("GetBean failed: %v", err)
	}

	if bean.Name != "test-bean" {
		t.Errorf("Expected Name 'test-bean', got %s", bean.Name)
	}
}

// TestBoot_HasBean_Coverage 测试检查 Bean 是否存在
func TestBoot_HasBean_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(
			NewModule().
				Name("test-module").
				Bean(Provide(func(c core.Container) (TestBean, error) {
					return TestBean{Name: "test-bean"}, nil
				})).
				Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 测试 Bean 已注册
	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected Bean to exist")
	}
}

// TestBoot_Register_Coverage 测试 Register 方法
func TestBoot_Register_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "test"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证 Bean 已注册
	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected TestBean to exist after Register")
	}
}

// TestBoot_GetByType_Coverage 测试 GetByType 方法
func TestBoot_GetByType(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "test-bean"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 通过 GetByType 获取 Bean
	ctx := app.Context()
	bean, err := ctx.GetByType(reflect.TypeOf(TestBean{}))
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}

	testBean, ok := bean.(TestBean)
	if !ok {
		t.Fatal("Expected TestBean type")
	}
	if testBean.Name != "test-bean" {
		t.Errorf("Expected Name 'test-bean', got %s", testBean.Name)
	}
}

// TestBoot_Register_Adapter_Coverage 测试 appCtxAdapter.Register 方法
func TestBoot_Register_Adapter(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 在启动前通过 adapter 注册 Bean
	ctx := app.Context()
	err = ctx.Register(reflect.TypeOf(TestBean{}), core.WithFactory[TestBean](func(c ...any) (any, error) {
		return TestBean{Name: "registered-via-register"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证 Bean 已注册
	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected TestBean to exist after Register")
	}

	// 验证 Bean 可以获取
	bean, err := core.GetByName[TestBean](app.Container(), "")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if bean.Name != "registered-via-register" {
		t.Errorf("Expected Name 'registered-via-register', got %s", bean.Name)
	}
}

// TestBoot_GetByType_Adapter_Coverage 测试 appCtxAdapter.GetByType 方法
func TestBoot_GetByType_Adapter(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "test-bean"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 获取 adapter
	ctx := app.Context()

	// 测试 GetByType 方法
	bean, err := ctx.GetByType(reflect.TypeOf(TestBean{}))
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}

	testBean, ok := bean.(TestBean)
	if !ok {
		t.Fatal("Expected TestBean type")
	}
	if testBean.Name != "test-bean" {
		t.Errorf("Expected Name 'test-bean', got %s", testBean.Name)
	}
}

// TestBoot_GetByType_NotFound_Coverage 测试 GetByType 找不到 Bean 的情况
func TestBoot_GetByType_NotFound(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 获取 adapter
	ctx := app.Context()

	// 测试 GetByType 方法 - 应该找不到
	_, err = ctx.GetByType(reflect.TypeOf(TestBean{}))
	if err == nil {
		t.Error("Expected error when bean not found")
	}
}
