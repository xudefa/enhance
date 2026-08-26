package boot

import (
	"context"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
)

// 测试用数据库类型
type testDatabase struct {
	URL string
}

// 测试用仓库类型
type testRepository struct {
	DB *testDatabase
}

func TestModule_Install(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()

	// 创建模块
	dbModule := NewModule("database",
		Provide(func(c core.Container) (*testDatabase, error) {
			return &testDatabase{URL: "localhost:5432"}, nil
		}),
		Provide(func(c core.Container) (*testRepository, error) {
			db, err := core.GetByName[*testDatabase](c, "")
			if err != nil {
				return nil, err
			}
			return &testRepository{DB: db}, nil
		}),
	)

	// 安装模块
	err := dbModule.Install(container)
	if err != nil {
		t.Fatalf("failed to install module: %v", err)
	}

	// 列出所有 Bean
	beans := container.ListBeans()
	t.Logf("registered beans: %v", beans)

	// 获取 Bean
	repo, err := container.Get(reflect.TypeFor[*testRepository]())
	if err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}

	repository := repo[0].(*testRepository)
	if repository.DB.URL != "localhost:5432" {
		t.Errorf("expected URL 'localhost:5432', got '%s'", repository.DB.URL)
	}
}

func TestModule_Invoke(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()

	type Database struct {
		URL      string
		Migrated bool
	}

	invoked := false

	// 创建模块
	dbModule := NewModule("database",
		Provide(func(c core.Container) (*Database, error) {
			return &Database{URL: "localhost:5432"}, nil
		}),
		Invoke(func(db *Database) error {
			db.Migrated = true
			invoked = true
			return nil
		}),
	)

	// 安装模块
	err := dbModule.Install(container)
	if err != nil {
		t.Fatalf("failed to install module: %v", err)
	}

	if !invoked {
		t.Error("expected invoke function to be called")
	}

	// 验证数据库已迁移
	db, err := container.Get(reflect.TypeFor[*Database]())
	if err != nil {
		t.Fatalf("failed to get database: %v", err)
	}

	database := db[0].(*Database)
	if !database.Migrated {
		t.Error("expected database to be migrated")
	}
}

func TestModule_WithApplication(t *testing.T) {
	t.Parallel()
	// 定义模块
	dbModule := NewModule("database",
		Provide(func(c core.Container) (string, error) {
			return "test-db", nil
		}),
	)

	// 创建应用并安装模块
	app, err := NewApplication(
		WithAppName("test-app"),
		WithModulesOption(dbModule),
	)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	// 启动应用（会安装模块）
	err = app.Start()
	if err != nil {
		t.Fatalf("failed to start application: %v", err)
	}

	// 验证模块已安装
	dbs, err := app.Container().Get(reflect.TypeFor[string]())
	if err != nil {
		t.Fatalf("failed to get database bean: %v", err)
	}

	if dbs[0].(string) != "test-db" {
		t.Errorf("expected 'test-db', got '%s'", dbs[0])
	}

	// 停止应用
	err = app.Stop()
	if err != nil {
		t.Fatalf("failed to stop application: %v", err)
	}
}

func TestProvide_Generic(t *testing.T) {
	t.Parallel()
	c := core.NewContainer()

	type repo struct{}
	type svc struct {
		r *repo
	}

	err := Provide(func(c core.Container) (*repo, error) {
		return &repo{}, nil
	})(c)
	if err != nil {
		t.Fatalf("Provide(repo) error = %v", err)
	}

	err = Provide(func(c core.Container) (*svc, error) {
		r, err := core.GetByName[*repo](c, "")
		if err != nil {
			return nil, err
		}
		return &svc{r: r}, nil
	})(c)
	if err != nil {
		t.Fatalf("Provide(svc) error = %v", err)
	}

	got, err := core.GetByName[*svc](c, "")
	if err != nil {
		t.Fatalf("GetBean(svc) error = %v", err)
	}
	if got.r == nil {
		t.Fatal("expected injected repo")
	}
}

func TestProvideReflect_NilConstructor(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	provider := ProvideReflect(nil)

	err := provider(container)
	if err == nil {
		t.Error("expected error for nil constructor")
	}
}

func TestProvideReflect_NonFuncConstructor(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	provider := ProvideReflect("not a function")

	err := provider(container)
	if err == nil {
		t.Error("expected error for non-function constructor")
	}
}

func TestProvideReflect_NoOutput(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	provider := ProvideReflect(func() {})

	err := provider(container)
	if err == nil {
		t.Error("expected error for function with no output")
	}
}

func TestInvoke_NilFunction(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	provider := Invoke(nil)

	err := provider(container)
	if err == nil {
		t.Error("expected error for nil function")
	}
}

func TestInvoke_NonFunction(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	provider := Invoke("not a function")

	err := provider(container)
	if err == nil {
		t.Error("expected error for non-function")
	}
}

func TestInvoke_WithMissingDependency(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	type MissingDep struct{}
	provider := Invoke(func(dep *MissingDep) error {
		return nil
	})

	err := provider(container)
	if err == nil {
		t.Error("expected error for missing dependency")
	}
}

func TestInvoke_WithErrorReturn(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	provider := Invoke(func() error {
		return &bootError{message: "test error"}
	})

	err := provider(container)
	if err == nil {
		t.Error("expected error from invoke function")
	}
}

func TestInvoke_Success(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	type Config struct {
		Value string
	}

	Provide(func(c core.Container) (*Config, error) {
		return &Config{Value: "test"}, nil
	})(container)

	invoked := false
	provider := Invoke(func(cfg *Config) error {
		invoked = true
		if cfg.Value != "test" {
			t.Errorf("expected config value 'test', got '%s'", cfg.Value)
		}
		return nil
	})

	err := provider(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !invoked {
		t.Error("expected function to be invoked")
	}
}

func TestProvideBean(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	type TestBean struct {
		Value string
	}

	provider := ProvideBean(&TestBean{Value: "test"})

	err := provider(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	beans := container.ListBeans()
	if len(beans) == 0 {
		t.Error("expected at least one bean")
	}
}

func TestProvideFactory(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	type TestBean struct {
		Value string
	}

	factory := func(c core.Container) (*TestBean, error) {
		return &TestBean{Value: "factory"}, nil
	}

	provider := ProvideFactory(factory)

	err := provider(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProvideNamed(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	type TestBean struct {
		Value string
	}

	provider := ProvideNamed("myBean", &TestBean{Value: "named"})

	err := provider(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProvidePrimary(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	type TestBean struct {
		Value string
	}

	provider := ProvidePrimary(&TestBean{Value: "primary"})

	err := provider(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMergeModules(t *testing.T) {
	t.Parallel()

	type TestBean1 struct {
		Value string
	}
	type TestBean2 struct {
		Value string
	}

	module1 := NewModule(
		Provide(func(c core.Container) (*TestBean1, error) {
			return &TestBean1{Value: "bean1"}, nil
		}),
	).Build()

	module2 := NewModule(
		Provide(func(c core.Container) (*TestBean2, error) {
			return &TestBean2{Value: "bean2"}, nil
		}),
	).Build()

	merged := MergeModules(module1, module2)
	container := core.NewContainer()
	err := merged.Install(container)
	if err != nil {
		t.Fatalf("unexpected error installing merged module: %v", err)
	}
}

func TestModule_ModuleName(t *testing.T) {
	t.Parallel()

	module := NewModule("test-module").Build()
	name := module.ModuleName()
	if name != "test-module" {
		t.Errorf("expected module name 'test-module', got '%s'", name)
	}
}

func TestModule_ModuleConditions(t *testing.T) {
	t.Parallel()

	module := NewModule("test-module").Build()
	conditions := module.ModuleConditions()
	_ = conditions
}

func TestWithModule(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Value string
	}

	module := NewModule(
		Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Value: "test"}, nil
		}),
	).Build()

	opt := WithModule(module)
	if opt == nil {
		t.Error("expected non-nil BootOption")
	}
}

func TestModuleBuilder_Name(t *testing.T) {
	t.Parallel()

	builder := NewModule().Name("test")
	module := builder.Build()
	if module.ModuleName() != "test" {
		t.Errorf("expected name 'test', got '%s'", module.ModuleName())
	}
}

func TestModuleBuilder_Bean(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Value string
	}

	builder := NewModule().
		Bean(ProvideBean(&TestBean{Value: "bean"}))

	container := core.NewContainer()
	module := builder.Build()
	err := module.Install(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModuleBuilder_Invoke(t *testing.T) {
	t.Parallel()

	invoked := false
	builder := NewModule().
		Invoke(func() error {
			invoked = true
			return nil
		})

	container := core.NewContainer()
	module := builder.Build()
	err := module.Install(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !invoked {
		t.Error("expected invoke to be called")
	}
}

func TestModuleBuilder_Hook(t *testing.T) {
	t.Parallel()

	hookCalled := false
	builder := NewModule("test-hook").
		Hook(&testHookForModule{fn: func(ctx context.Context) error {
			hookCalled = true
			return nil
		}})

	container := core.NewContainer()
	module := builder.Build()
	err := module.Install(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = hookCalled
}

type testHookForModule struct {
	fn func(context.Context) error
}

func (h *testHookForModule) OnInit(ctx context.Context) error {
	return h.fn(ctx)
}

func (h *testHookForModule) OnStart(ctx context.Context) error {
	return nil
}

func (h *testHookForModule) OnStop(ctx context.Context) error {
	return nil
}

func TestModuleBuilder_Hooks(t *testing.T) {
	t.Parallel()

	builder := NewModule()
	hooks := builder.Hooks()
	if hooks == nil {
		t.Error("expected non-nil hooks")
	}
}

func TestModuleBuilder_Starters(t *testing.T) {
	t.Parallel()

	builder := NewModule("test-starters")
	module := builder.Build()
	if module.ModuleName() != "test-starters" {
		t.Errorf("expected module name 'test-starters', got '%s'", module.ModuleName())
	}
}

func TestModuleBuilder_Condition(t *testing.T) {
	t.Parallel()

	builder := NewModule()
	conditions := builder.Conditions()
	if conditions == nil {
		t.Error("expected non-nil conditions")
	}
}

func TestModuleBuilder_Conditions(t *testing.T) {
	t.Parallel()

	builder := NewModule()
	conditions := builder.Conditions()
	if conditions == nil {
		t.Error("expected non-nil conditions")
	}
}

func TestModuleBuilder_Module(t *testing.T) {
	t.Parallel()

	builder := NewModule("test-module")
	module := builder.Build()
	if module.ModuleName() != "test-module" {
		t.Errorf("expected module name 'test-module', got '%s'", module.ModuleName())
	}
}

func TestConditionalModule(t *testing.T) {
	t.Parallel()

	cond := condition.OnProperty("test.enabled", "true")
	mod := NewModule("test-module")

	result := ConditionalModule([]condition.Condition{cond}, mod.Build())
	if len(result.conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result.conditions))
	}
}

func TestNamedModule(t *testing.T) {
	t.Parallel()

	mod := NewModule("original")

	result := NamedModule("new-name", mod.Build())
	if result.moduleName != "new-name" {
		t.Errorf("expected module name 'new-name', got '%s'", result.moduleName)
	}
}
