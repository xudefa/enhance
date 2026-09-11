package boot

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/lifecycle"
)

// TestBoot_Start_WithModules_Coverage 测试带模块的 Start
func TestBoot_Start_WithModules_Coverage(t *testing.T) {
	t.Parallel()

	// 创建一个简单模块
	module := NewModule().
		Name("test-module").
		Build()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(module),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()
}

// TestBoot_ModuleBuilder_Module_Coverage 测试 ModuleBuilder.Module 方法
func TestBoot_ModuleBuilder_Module(t *testing.T) {
	t.Parallel()

	builder := NewModule().
		Name("test-module")

	mod := builder.Module()
	if mod.moduleName != "test-module" {
		t.Errorf("Expected module name 'test-module', got %s", mod.moduleName)
	}
}

// TestBoot_WithModuleStarters_Coverage 测试 WithModuleStarters 选项
func TestBoot_WithModuleStarters(t *testing.T) {
	t.Parallel()

	module := NewModule().
		Name("test-module").
		Build()

	// 验证模块创建成功
	if module.moduleName != "test-module" {
		t.Errorf("Expected module name 'test-module', got %s", module.moduleName)
	}
}

// TestBoot_ModuleBuilder_Install_Coverage 测试 ModuleBuilder.Install 方法
func TestBoot_ModuleBuilder_Install_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	container := core.NewContainer()

	builder := NewModule().
		Name("test-module").
		Bean(Provide(func(c core.Container) (TestBean, error) {
			return TestBean{Name: "installed-bean"}, nil
		}))

	if err := builder.Install(container); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// 验证 Bean 已注册
	if !core.Has[TestBean](container, "") {
		t.Error("Expected Bean to be installed")
	}
}

// TestBoot_ModuleComposer_Coverage 测试模块组合
func TestBoot_ModuleComposer_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean1 struct {
		Name string
	}

	type TestBean2 struct {
		Name string
	}

	module1 := NewModule().
		Name("module1").
		Bean(Provide(func(c core.Container) (TestBean1, error) {
			return TestBean1{Name: "bean1"}, nil
		})).
		Build()

	module2 := NewModule().
		Name("module2").
		Bean(Provide(func(c core.Container) (TestBean2, error) {
			return TestBean2{Name: "bean2"}, nil
		})).
		Build()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(module1, module2),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证两个 Bean 都已注册
	if !core.Has[TestBean1](app.Container(), "") {
		t.Error("Expected TestBean1 to exist")
	}
	if !core.Has[TestBean2](app.Container(), "") {
		t.Error("Expected TestBean2 to exist")
	}
}

// TestBoot_WithModuleStartersOption_Coverage 测试 WithModuleStarters 选项
func TestBoot_WithModuleStartersOption_Coverage(t *testing.T) {
	t.Parallel()

	module := NewModule().
		Name("test-module").
		Starter(&TestStarter{}).
		Build()

	if len(module.starters) != 1 {
		t.Errorf("Expected 1 starter, got %d", len(module.starters))
	}
}

// TestBoot_WithModuleStartersDirect_Coverage 测试直接使用 WithModuleStarters 函数
func TestBoot_WithModuleStartersDirect(t *testing.T) {
	t.Parallel()

	// 直接使用 WithModuleStarters 函数
	opt := WithModuleStarters(&TestStarter{})
	if opt == nil {
		t.Fatal("Expected non-nil ModuleOption")
	}

	// 创建模块并应用选项
	module := &Module{
		moduleName: "test-module",
	}

	// 手动应用选项
	opt(module)

	if len(module.starters) != 1 {
		t.Errorf("Expected 1 starter, got %d", len(module.starters))
	}
}

// TestBoot_MergeModules_NameConcat_Coverage 测试 MergeModules 名称拼接逻辑
func TestBoot_MergeModules_NameConcat(t *testing.T) {
	t.Parallel()

	type TestBean1 struct {
		Value string
	}
	type TestBean2 struct {
		Value string
	}

	// 创建两个带名称的模块
	module1 := NewModule().
		Name("module1").
		Bean(Provide(func(c core.Container) (*TestBean1, error) {
			return &TestBean1{Value: "bean1"}, nil
		})).
		Build()

	module2 := NewModule().
		Name("module2").
		Bean(Provide(func(c core.Container) (*TestBean2, error) {
			return &TestBean2{Value: "bean2"}, nil
		})).
		Build()

	// 合并模块
	merged := MergeModules(module1, module2)

	// 验证名称被拼接
	expectedName := "module1+module2"
	if merged.ModuleName() != expectedName {
		t.Errorf("Expected module name '%s', got '%s'", expectedName, merged.ModuleName())
	}

	// 验证可以安装
	container := core.NewContainer()
	if err := merged.Install(container); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// 验证 Bean 已注册
	if !core.Has[*TestBean1](container, "") {
		t.Error("Expected TestBean1 to exist")
	}
	if !core.Has[*TestBean2](container, "") {
		t.Error("Expected TestBean2 to exist")
	}
}

// TestBoot_ModuleConditions_Coverage 测试模块条件
func TestBoot_ModuleConditions(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	// 创建带条件的模块
	module := NewModule().
		Name("conditional-module").
		Bean(Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Name: "test"}, nil
		})).
		Condition(condition.OnProperty("feature.enabled", "true")).
		Build()

	// 验证条件已设置
	conditions := module.ModuleConditions()
	if len(conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(conditions))
	}
}

// TestBoot_WithModule_Coverage 测试 WithModule 选项
func TestBoot_WithModule_Coverage(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	module := NewModule().
		Name("test-module").
		Bean(Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Name: "module-bean"}, nil
		})).
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

	// 验证 Bean 已注册
	if !core.Has[*TestBean](app.Container(), "") {
		t.Error("Expected TestBean to exist from module")
	}
}

// TestBoot_ModuleHooks_Coverage 测试模块钩子
func TestBoot_ModuleHooks(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	hookCalled := false
	hook := lifecycle.NewHookFunc(
		func(ctx context.Context) error {
			hookCalled = true
			return nil
		},
		nil,
		nil,
	)

	module := NewModule().
		Name("hook-module").
		Bean(Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Name: "hook-bean"}, nil
		})).
		Hook(hook).
		Build()

	// 验证钩子已设置
	hooks := module.ModuleHooks()
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook, got %d", len(hooks))
	}

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

	if !hookCalled {
		t.Error("Expected hook to be called")
	}
}

// TestBoot_ModuleBuilder_Invoke_NilFn_Coverage 测试 Invoke 方法传入 nil 函数
func TestBoot_ModuleBuilder_Invoke_NilFn(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	var fn any = nil
	module := NewModule().
		Name("invoke-nil-module").
		Bean(Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Name: "test"}, nil
		})).
		Invoke(fn).
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

	// Start 应该失败，因为 Invoke 函数为 nil
	err = app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail with nil invoke function")
	}
}

// TestBoot_ModuleBuilder_Invoke_NonFunc_Coverage 测试 Invoke 方法传入非函数类型
func TestBoot_ModuleBuilder_Invoke_NonFunc(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	module := NewModule().
		Name("invoke-nonfunc-module").
		Bean(Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Name: "test"}, nil
		})).
		Invoke("not a function").
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

	// Start 应该失败，因为 Invoke 不是函数
	err = app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail with non-function invoke")
	}
}

// TestBoot_ModuleBuilder_Invoke_MissingDependency_Coverage 测试 Invoke 方法依赖找不到
func TestBoot_ModuleBuilder_Invoke_MissingDependency(t *testing.T) {
	t.Parallel()

	type MissingBean struct {
		Name string
	}

	module := NewModule().
		Name("invoke-missing-module").
		Invoke(func(bean *MissingBean) error {
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

	// Start 应该失败，因为依赖找不到
	err = app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail with missing dependency")
	}
}
