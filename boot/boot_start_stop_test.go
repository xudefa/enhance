package boot

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xudefa/enhance/core"
)

// TestBootStartStop tests Start and Stop functionality using table-driven tests.
func TestBootStartStop(t *testing.T) {
	t.Parallel()

	for _, tt := range testBootStartStopCases() {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := newTestApp(t, tt.opts...)
			startAndStopTestApp(t, app)
			tt.verify(t, app)
		})
	}
}

func testBootStartStopCases() []struct {
	name   string
	opts   []BootOption
	verify func(t *testing.T, app *Boot)
} {
	return []struct {
		name   string
		opts   []BootOption
		verify func(t *testing.T, app *Boot)
	}{
		{
			name: "basic start and stop",
			opts: []BootOption{WithVersion("1.0.0")},
			verify: func(t *testing.T, app *Boot) {
				if !app.started.Load() {
					t.Error("Expected started to be true after Start")
				}
			},
		},
		{
			name: "start without auto config",
			opts: []BootOption{WithoutAutoConfig()},
			verify: func(t *testing.T, app *Boot) {
				if !app.IsRunning() {
					t.Error("Expected app to be running")
				}
			},
		},
		{
			name: "duplicate start returns nil",
			opts: []BootOption{WithoutAutoConfig()},
			verify: func(t *testing.T, app *Boot) {
				if err := app.Start(); err != nil {
					t.Errorf("Second Start should return nil, got %v", err)
				}
			},
		},
		{
			name: "duplicate stop returns nil",
			opts: []BootOption{WithoutAutoConfig()},
			verify: func(t *testing.T, app *Boot) {
				if err := app.Stop(); err != nil {
					t.Errorf("Second Stop should return nil, got %v", err)
				}
			},
		},
	}
}

// TestBootStartWithModules tests Start with modules.
func TestBootStartWithModules(t *testing.T) {
	t.Parallel()

	module := NewModule().Name("test-module").Build()
	app := newTestApp(t, WithModules(module))
	startAndStopTestApp(t, app)
}

// TestBootStartWithProfiles tests Start with profiles.
func TestBootStartWithProfiles(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, WithProfiles("dev", "local", "test"))
	startAndStopTestApp(t, app)

	if len(app.config.Profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(app.config.Profiles))
	}
}

// TestBootStartWithExcludeConfigs tests Start with excluded configs.
func TestBootStartWithExcludeConfigs(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, WithoutStarters(), WithExclude("Config1", "Config2", "Config3", "Config4"))
	startAndStopTestApp(t, app)

	if len(app.config.ExcludedAutoConfigs) != 4 {
		t.Errorf("Expected 4 excluded configs, got %d", len(app.config.ExcludedAutoConfigs))
	}
}

// TestBootGetProperty tests GetProperty method.
func TestBootGetProperty(t *testing.T) {
	t.Parallel()

	app := newTestApp(t,
		WithPropertySource(createTestPropertySource("test", "app.name", "property-test")),
	)
	startAndStopTestApp(t, app)

	name, ok := app.Environment().GetProperty("app.name")
	if !ok {
		t.Error("Expected GetProperty to return true")
	}
	if name != "property-test" {
		t.Errorf("Expected 'property-test', got %v", name)
	}

	_, ok = app.Environment().GetProperty("non.existent")
	if ok {
		t.Error("Expected GetProperty to return false for non-existent key")
	}
}

// TestBootConfig tests Config method.
func TestBootConfig(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, WithAppName("my-app"), WithVersion("1.2.3"))
	cfg := app.Config()
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}
	if cfg.AppName != "my-app" {
		t.Errorf("Expected AppName 'my-app', got %s", cfg.AppName)
	}
	if cfg.Version != "1.2.3" {
		t.Errorf("Expected Version '1.2.3', got %s", cfg.Version)
	}
}

// TestBootWithoutStartupReport tests WithoutStartupReport option.
func TestBootWithoutStartupReport(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, WithoutStartupReport())
	startAndStopTestApp(t, app)
}

// TestBootWithModuleStarters tests WithModuleStarters option.
func TestBootWithModuleStarters(t *testing.T) {
	t.Parallel()

	module := NewModule().Name("test-module").Build()
	if module.moduleName != "test-module" {
		t.Errorf("Expected module name 'test-module', got %s", module.moduleName)
	}
}

// TestBootWithModuleStartersDirect tests direct use of WithModuleStarters function.
func TestBootWithModuleStartersDirect(t *testing.T) {
	t.Parallel()

	opt := WithModuleStarters(&TestStarter{})
	if opt == nil {
		t.Fatal("Expected non-nil ModuleOption")
	}
	module := &Module{moduleName: "test-module"}
	opt(module)
	if len(module.starters) != 1 {
		t.Errorf("Expected 1 starter, got %d", len(module.starters))
	}
}

// TestBootCollectAutoConfigReport tests collectAutoConfigReport method.
func TestBootCollectAutoConfigReport(t *testing.T) {
	t.Parallel()

	DisableAutoConfigReport()

	app := newTestApp(t, WithProperty("enhance.debug", true))
	startAndStopTestApp(t, app)

	if !app.IsRunning() {
		t.Error("Expected app to be running")
	}
}

// TestBootCollectAutoConfigReportWithoutDebug tests auto config report without debug mode.
func TestBootCollectAutoConfigReportWithoutDebug(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, WithoutStarters())
	startAndStopTestApp(t, app)
}

// TestBootWaitForSignal tests WaitForSignal method.
//
// 注意：不得使用 t.Parallel()。WaitForSignal 依赖进程级 signal.Notify，
// 且 syscall.Kill 直接向本进程发送 SIGTERM，多个此类测试并行会互抢
// 进程信号，导致偶发 "signal: terminated"（进程被默认处理器终止）或
// 等待者永久阻塞。因此信号类测试必须全局串行（与 AGENTS §2.3 的
// "禁止并发：操作共享资源/全局状态" 条款一致）。
func TestBootWaitForSignal(t *testing.T) {
	app := newTestApp(t)
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	app.WaitForSignal()
}

// TestBootModuleBuilderInvokeMissingDependency tests Invoke with missing dependency.
func TestBootModuleBuilderInvokeMissingDependency(t *testing.T) {
	t.Parallel()

	type MissingBean struct{ Name string }
	module := NewModule().
		Name("invoke-missing-module").
		Invoke(func(bean *MissingBean) error { return nil }).
		Build()

	app := newTestApp(t, WithModule(module))
	err := app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail with missing dependency")
	}
}

// TestBootWithConfigFile tests WithConfigLocation option.
func TestBootWithConfigFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "application.json")
	configContent := `{"app": {"name": "config-file-test", "port": 7070}}`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	absConfigFile, err := filepath.Abs(configFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	app := newTestApp(t, WithConfigLocation(absConfigFile))
	startAndStopTestApp(t, app)

	name, ok := app.Environment().GetProperty("app.name")
	if !ok || name != "config-file-test" {
		t.Errorf("Expected app.name='config-file-test', got %v", name)
	}
}

// TestBootGetRequiredProperty tests GetRequiredProperty method.
func TestBootGetRequiredProperty(t *testing.T) {
	t.Parallel()

	app := newTestApp(t,
		WithPropertySource(createTestPropertySource("test", "app.name", "required-test")),
	)
	startAndStopTestApp(t, app)

	name, err := app.Environment().GetRequiredProperty("app.name")
	if err != nil {
		t.Fatalf("GetRequiredProperty failed: %v", err)
	}
	if name != "required-test" {
		t.Errorf("Expected 'required-test', got %v", name)
	}

	_, err = app.Environment().GetRequiredProperty("non.existent")
	if err == nil {
		t.Error("Expected error for non-existent property")
	}
}

// TestBootContext tests Context, Container, and Environment accessors.
func TestBootContext(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	startAndStopTestApp(t, app)

	ctx := app.Context()
	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}
	if app.Container() == nil {
		t.Fatal("Expected non-nil container")
	}
	if app.Environment() == nil {
		t.Fatal("Expected non-nil environment")
	}
}

// TestBootGetBean tests bean retrieval from the container.
func TestBootGetBean(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app := newTestApp(t,
		WithModules(
			NewModule().
				Name("test-module").
				Bean(Provide(func(c core.Container) (TestBean, error) {
					return TestBean{Name: "test-bean"}, nil
				})).
				Build(),
		),
	)
	startAndStopTestApp(t, app)

	bean, err := core.GetByName[TestBean](app.Container(), "")
	if err != nil {
		t.Fatalf("GetBean failed: %v", err)
	}
	if bean.Name != "test-bean" {
		t.Errorf("Expected Name 'test-bean', got %s", bean.Name)
	}
}

// TestBootHasBean tests bean existence check.
func TestBootHasBean(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	app := newTestApp(t,
		WithModules(
			NewModule().
				Name("test-module").
				Bean(Provide(func(c core.Container) (TestBean, error) {
					return TestBean{Name: "test-bean"}, nil
				})).
				Build(),
		),
	)
	startAndStopTestApp(t, app)

	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected TestBean to exist")
	}
}

// TestBootModuleComposer tests multiple modules.
func TestBootModuleComposer(t *testing.T) {
	t.Parallel()

	type TestBean1 struct{ Name string }
	type TestBean2 struct{ Name string }

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

	app := newTestApp(t, WithModules(module1, module2))
	startAndStopTestApp(t, app)

	if !core.Has[TestBean1](app.Container(), "") {
		t.Error("Expected TestBean1 to exist")
	}
	if !core.Has[TestBean2](app.Container(), "") {
		t.Error("Expected TestBean2 to exist")
	}
}

// TestBootModuleBuilderInstall tests ModuleBuilder.Install method.
func TestBootModuleBuilderInstall(t *testing.T) {
	t.Parallel()

	type TestBean struct{ Name string }

	container := core.NewContainer()
	builder := NewModule().
		Name("test-module").
		Bean(Provide(func(c core.Container) (TestBean, error) {
			return TestBean{Name: "installed-bean"}, nil
		}))

	if err := builder.Install(container); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if !core.Has[TestBean](container, "") {
		t.Error("Expected TestBean to be installed")
	}
}

// TestBootWithModuleStarterOption tests WithModuleStarters option on module.
func TestBootWithModuleStarterOption(t *testing.T) {
	t.Parallel()

	module := NewModule().
		Name("test-module").
		Starter(&TestStarter{}).
		Build()

	if len(module.starters) != 1 {
		t.Errorf("Expected 1 starter, got %d", len(module.starters))
	}
}

// TestBootWithModule tests WithModule option.
func TestBootWithModule(t *testing.T) {
	t.Parallel()

	type TestBean struct{ Name string }

	module := NewModule().
		Name("test-module").
		Bean(Provide(func(c core.Container) (*TestBean, error) {
			return &TestBean{Name: "module-bean"}, nil
		})).
		Build()

	app := newTestApp(t, WithModule(module))
	startAndStopTestApp(t, app)

	if !core.Has[*TestBean](app.Container(), "") {
		t.Error("Expected TestBean to exist from module")
	}
}
