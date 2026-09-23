package boot

import (
	"context"
	"strings"
	"testing"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/lifecycle"
)

func TestBoot_WithModule(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	mod := NewModule().
		Name("test-module").
		Bean(Provide(func(c core.Container) (TestBean, error) {
			return TestBean{Name: "module-bean"}, nil
		})).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected TestBean to exist from module")
	}
}

func TestBoot_WithModuleMultiple(t *testing.T) {
	t.Parallel()
	type Bean1 struct{ Name string }
	type Bean2 struct{ Name string }
	mod1 := NewModule().Name("module-1").Bean(Provide(func(c core.Container) (Bean1, error) {
		return Bean1{Name: "bean1"}, nil
	})).Build()
	mod2 := NewModule().Name("module-2").Bean(Provide(func(c core.Container) (Bean2, error) {
		return Bean2{Name: "bean2"}, nil
	})).Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod1),
		WithModule(mod2),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !core.Has[Bean1](app.Container(), "") || !core.Has[Bean2](app.Container(), "") {
		t.Error("Expected both beans to exist")
	}
}

func TestBoot_ModuleInstallError(t *testing.T) {
	t.Parallel()
	mod := NewModule().
		Name("failing-module").
		Invoke(func() error {
			return &bootError{code: ErrCodeModuleInstall, message: "install failed", phase: "安装"}
		}).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	err = app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail")
	}
}

func TestBoot_ModuleWithHook(t *testing.T) {
	t.Parallel()
	initCalled := false
	stopCalled := false
	hook := lifecycle.NewHookFunc(
		func(ctx context.Context) error { initCalled = true; return nil },
		func(ctx context.Context) error { stopCalled = true; return nil },
		nil,
	)
	mod := NewModule().
		Name("hook-module").
		Hook(hook).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := app.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if !initCalled {
		t.Error("Expected hook OnInit to be called")
	}
	if !stopCalled {
		t.Error("Expected hook OnStop to be called")
	}
}

func TestBoot_ModuleWithStarter(t *testing.T) {
	t.Parallel()
	starter := newMockStarter("module-starter")
	mod := NewModule().
		Name("starter-module").
		Starter(starter).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !starter.started.Load() {
		t.Error("Expected module starter to be started")
	}
}

func TestBoot_ModuleWithCondition(t *testing.T) {
	t.Parallel()
	type TestBean struct{ Name string }
	mod := NewModule().
		Name("conditional-module").
		Bean(Provide(func(c core.Container) (TestBean, error) {
			return TestBean{Name: "conditional"}, nil
		})).
		Condition(condition.OnProperty("module.enabled", "true")).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
		WithProperty("module.enabled", "true"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected conditional bean to exist")
	}
}

func TestBoot_ModuleWithConditionNotMatch(t *testing.T) {
	t.Parallel()
	type TestBean struct{ Name string }
	mod := NewModule().
		Name("conditional-module").
		Bean(Provide(func(c core.Container) (TestBean, error) {
			return TestBean{Name: "conditional"}, nil
		})).
		Condition(condition.OnProperty("module.enabled", "true")).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if core.Has[TestBean](app.Container(), "") {
		t.Error("Expected conditional bean NOT to exist when condition not met")
	}
}

func TestBoot_ModuleInvoke(t *testing.T) {
	t.Parallel()
	invoked := false
	mod := NewModule().
		Name("invoke-module").
		Invoke(func() error { invoked = true; return nil }).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !invoked {
		t.Error("Expected invoke function to be called")
	}
}

func TestBoot_ModuleInvokeWithDependency(t *testing.T) {
	t.Parallel()
	type TestBean struct{ Name string }
	invokedValue := ""
	mod := NewModule().
		Name("invoke-dep-module").
		Bean(Provide(func(c core.Container) (TestBean, error) {
			return TestBean{Name: "invoke-test"}, nil
		})).
		Invoke(func(bean TestBean) error { invokedValue = bean.Name; return nil }).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if invokedValue != "invoke-test" {
		t.Errorf("Expected invokedValue 'invoke-test', got %s", invokedValue)
	}
}

func TestBoot_ModuleInvokeNilFn(t *testing.T) {
	t.Parallel()
	var fn any = nil
	mod := NewModule().
		Name("invoke-nil-module").
		Invoke(fn).
		Build()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModule(mod),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	err = app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail with nil invoke function")
	}
}

func TestBoot_PluginLifecycle(t *testing.T) {
	t.Parallel()
	plugin := &mockPlugin{name: "test-plugin", version: "1.0.0"}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPlugin(plugin),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !plugin.initCalled {
		t.Error("Expected plugin to be initialized")
	}
	if !plugin.startCalled {
		t.Error("Expected plugin to be started")
	}
}

func TestBoot_PluginMultiple(t *testing.T) {
	t.Parallel()
	plugin1 := &mockPlugin{name: "plugin-1", version: "1.0.0"}
	plugin2 := &mockPlugin{name: "plugin-2", version: "1.0.0"}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPlugin(plugin1),
		WithPlugin(plugin2),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !plugin1.startCalled || !plugin2.startCalled {
		t.Error("Expected both plugins to be started")
	}
}

func TestBoot_PluginInitError(t *testing.T) {
	t.Parallel()
	plugin := &mockPlugin{
		name:    "failing-plugin",
		version: "1.0.0",
		initErr: &bootError{code: ErrCodeUnknown, message: "init failed", phase: "插件初始化"},
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPlugin(plugin),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	err = app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail")
	}
	if !strings.Contains(err.Error(), "init failed") {
		t.Errorf("Expected error to contain 'init failed', got: %v", err)
	}
}

func TestBoot_PluginStartError(t *testing.T) {
	t.Parallel()
	plugin := &mockPlugin{
		name:     "failing-plugin",
		version:  "1.0.0",
		startErr: &bootError{code: ErrCodeUnknown, message: "start failed", phase: "插件启动"},
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPlugin(plugin),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	err = app.Start()
	if err == nil {
		t.Fatal("Expected Start to fail")
	}
	if !strings.Contains(err.Error(), "start failed") {
		t.Errorf("Expected error to contain 'start failed', got: %v", err)
	}
}

func TestBoot_PluginStopError(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{
		name:    "failing-plugin",
		version: "1.0.0",
		stopErr: &bootError{code: ErrCodeUnknown, message: "stop failed", phase: "插件停止"},
	}
	pm.Register(plugin)
	pm.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	pm.InitAll()
	pm.StartAll()
	err := pm.StopAll()
	if err == nil {
		t.Fatal("Expected StopAll to fail")
	}
	if !strings.Contains(err.Error(), "stop failed") {
		t.Errorf("Expected error to contain 'stop failed', got: %v", err)
	}
}

func TestBoot_PluginManager(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{name: "test-plugin", version: "1.0.0"}
	if err := pm.Register(plugin); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	pm.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	if err := pm.InitAll(); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}
	if !plugin.initCalled {
		t.Error("Expected plugin to be initialized")
	}
	if err := pm.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}
	if !plugin.startCalled {
		t.Error("Expected plugin to be started")
	}
}

func TestBoot_PluginManagerStop(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{name: "test-plugin", version: "1.0.0"}
	if err := pm.Register(plugin); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	pm.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	if err := pm.InitAll(); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}
	if err := pm.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}
	if err := pm.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}
	if !plugin.stopCalled {
		t.Error("Expected plugin to be stopped")
	}
}

func TestBoot_PluginManagerDuplicate(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{name: "test-plugin", version: "1.0.0"}
	if err := pm.Register(plugin); err != nil {
		t.Fatalf("First Register failed: %v", err)
	}
	err := pm.Register(plugin)
	if err == nil {
		t.Fatal("Expected error for duplicate registration")
	}
}

func TestBoot_PluginManagerStopOrder(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin1 := &mockPlugin{name: "plugin-1", version: "1.0.0"}
	plugin2 := &mockPlugin{name: "plugin-2", version: "1.0.0"}
	pm.Register(plugin1)
	pm.Register(plugin2)
	pm.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	pm.InitAll()
	pm.StartAll()
	pm.StopAll()
	if !plugin1.stopCalled || !plugin2.stopCalled {
		t.Error("Expected both plugins to be stopped")
	}
}

func TestBoot_PluginManagerInitError(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{
		name:    "failing-init",
		version: "1.0.0",
		initErr: &bootError{code: ErrCodeUnknown, message: "init failed", phase: "插件初始化"},
	}
	pm.Register(plugin)
	pm.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	err := pm.InitAll()
	if err == nil {
		t.Fatal("Expected InitAll to fail")
	}
	if !strings.Contains(err.Error(), "init failed") {
		t.Errorf("Expected error to contain 'init failed', got: %v", err)
	}
}

func TestBoot_PluginManagerStartError(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{
		name:     "failing-start",
		version:  "1.0.0",
		startErr: &bootError{code: ErrCodeUnknown, message: "start failed", phase: "插件启动"},
	}
	pm.Register(plugin)
	pm.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	pm.InitAll()
	err := pm.StartAll()
	if err == nil {
		t.Fatal("Expected StartAll to fail")
	}
	if !strings.Contains(err.Error(), "start failed") {
		t.Errorf("Expected error to contain 'start failed', got: %v", err)
	}
}

func TestBoot_PluginManagerStopAllError(t *testing.T) {
	t.Parallel()
	pm := NewPluginManager()
	plugin := &mockPlugin{
		name:    "failing-stop",
		version: "1.0.0",
		stopErr: &bootError{code: ErrCodeUnknown, message: "stop failed", phase: "插件停止"},
	}
	pm.Register(plugin)
	pm.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	pm.InitAll()
	pm.StartAll()
	err := pm.StopAll()
	if err == nil {
		t.Fatal("Expected StopAll to fail")
	}
	if !strings.Contains(err.Error(), "stop failed") {
		t.Errorf("Expected error to contain 'stop failed', got: %v", err)
	}
}
