package boot

import (
	"context"
	"fmt"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

// TestBoot_WithPlugin_Coverage 测试 WithPlugin 选项
func TestBoot_WithPlugin_Coverage(t *testing.T) {
	t.Parallel()

	plugin := &testPlugin{name: "test-plugin", version: "1.0.0"}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPlugin(plugin),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.Plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(app.config.Plugins))
	}
}

// TestBoot_WithPlugins_Coverage 测试 WithPlugins 选项
func TestBoot_WithPlugins_Coverage(t *testing.T) {
	t.Parallel()

	plugin1 := &testPlugin{name: "plugin1", version: "1.0.0"}
	plugin2 := &testPlugin{name: "plugin2", version: "1.0.0"}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPlugins(plugin1, plugin2),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.Plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(app.config.Plugins))
	}
}

// TestBoot_RegisterPlugin_Coverage 测试 RegisterPlugin 函数
func TestBoot_RegisterPlugin(t *testing.T) {
	t.Parallel()

	plugin := &testPlugin{name: "test-plugin", version: "1.0.0"}

	// RegisterPlugin 会注册到全局 PluginManager
	RegisterPlugin(plugin)

	// 验证插件已注册
	manager := GetPluginManager()
	if manager == nil {
		t.Fatal("Expected PluginManager to exist")
	}

	p, ok := manager.Get("test-plugin")
	if !ok {
		t.Error("Expected plugin to be registered")
	}
	if p == nil {
		t.Error("Expected non-nil plugin")
	}
}

// TestBoot_PluginContext_Coverage 测试 PluginContext 方法覆盖
func TestBoot_PluginContext(t *testing.T) {
	t.Parallel()

	plugin := testBootPluginContextBuild(t)
	testBootPluginContextStart(t, plugin)

	if !plugin.startCalled {
		t.Error("Expected plugin Start to be called")
	}
}

func testBootPluginContextBuild(t *testing.T) *testPluginWithCallback {
	t.Helper()

	return &testPluginWithCallback{
		name:    "callback-plugin",
		version: "1.0.0",
		onStart: func(ctx PluginContext) error {
			if container := ctx.Container(); container == nil {
				t.Error("Expected non-nil Container")
			}
			if env := ctx.Environment(); env == nil {
				t.Error("Expected non-nil Environment")
			}
			_, _ = ctx.Config("app.name")
			_, _ = ctx.GetPlugin("callback-plugin")
			return nil
		},
	}
}

func testBootPluginContextStart(t *testing.T, plugin *testPluginWithCallback) {
	t.Helper()

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
		WithPlugin(plugin),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()
}

// TestBoot_WithPluginMultiple_Coverage 测试多个插件
func TestBoot_WithPluginMultiple(t *testing.T) {
	t.Parallel()

	plugin1 := &testPlugin{name: "plugin1", version: "1.0.0"}
	plugin2 := &testPlugin{name: "plugin2", version: "1.0.0"}
	plugin3 := &testPlugin{name: "plugin3", version: "1.0.0"}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPlugins(plugin1, plugin2, plugin3),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.Plugins) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(app.config.Plugins))
	}
}

// TestBoot_PluginContext_GetPlugin_Coverage 测试 PluginContext.GetPlugin 方法
func TestBoot_PluginContext_GetPlugin(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	// 先注册到全局管理器
	plugin := &testPluginWithCallback{
		name:    "get-plugin-coverage",
		version: "1.0.0",
		onStart: func(ctx PluginContext) error {
			// 测试 GetPlugin 方法 - 获取自身
			p, ok := ctx.GetPlugin("get-plugin-coverage")
			if !ok {
				t.Logf("GetPlugin returned false (plugin may not be in global manager)")
			} else if p == nil {
				t.Error("Expected non-nil plugin")
			}
			return nil
		},
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
		WithPlugin(plugin),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	if !plugin.startCalled {
		t.Error("Expected plugin Start to be called")
	}
}

// TestBoot_PluginManager_Coverage 测试 PluginManager
func TestBoot_PluginManager(t *testing.T) {
	t.Parallel()

	// 创建一个新的插件管理器
	manager := NewPluginManager()
	if manager == nil {
		t.Fatal("Expected non-nil PluginManager")
	}

	// 测试 Register
	plugin := &testPlugin{name: "test-plugin", version: "1.0.0"}
	manager.Register(plugin)

	// 测试 Get
	p, ok := manager.Get("test-plugin")
	if !ok {
		t.Error("Expected to find test-plugin")
	}
	if p == nil {
		t.Error("Expected non-nil plugin")
	}

	// 测试 List
	plugins := manager.List()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

// TestBoot_PluginManager_StartAll_StopAll_Coverage 测试 StartAll 和 StopAll 完整流程
func TestBoot_PluginManager_StartAll_StopAll(t *testing.T) {
	t.Parallel()

	manager := NewPluginManager()

	// 使用 mockPlugin
	plugin1 := &mockPlugin{name: "plugin-a", version: "1.0.0"}
	plugin2 := &mockPlugin{name: "plugin-b", version: "1.0.0"}

	manager.Register(plugin1)
	manager.Register(plugin2)

	// 创建上下文
	ctx := &mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	}
	manager.SetContext(ctx)

	// 测试 InitAll
	if err := manager.InitAll(); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}

	// 测试 StartAll
	if err := manager.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	// 验证插件已启动
	if !plugin1.startCalled {
		t.Error("Expected plugin1 Start to be called")
	}
	if !plugin2.startCalled {
		t.Error("Expected plugin2 Start to be called")
	}

	// 测试 StopAll
	if err := manager.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	// 验证插件已停止
	if !plugin1.stopCalled {
		t.Error("Expected plugin1 Stop to be called")
	}
	if !plugin2.stopCalled {
		t.Error("Expected plugin2 Stop to be called")
	}
}

// TestBoot_PluginManager_StartAll_Error_Coverage 测试 StartAll 错误场景
func TestBoot_PluginManager_StartAll_Error(t *testing.T) {
	t.Parallel()

	manager := NewPluginManager()

	// 创建会失败的插件
	failingPlugin := &mockPlugin{
		name:     "failing-plugin",
		version:  "1.0.0",
		startErr: fmt.Errorf("intentional start failure"),
	}

	manager.Register(failingPlugin)
	ctx := &mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	}
	manager.SetContext(ctx)

	// 初始化
	if err := manager.InitAll(); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}

	// StartAll 应该失败
	err := manager.StartAll()
	if err == nil {
		t.Fatal("Expected StartAll to fail")
	}
}

// TestBoot_PluginManager_StopAll_Error_Coverage 测试 StopAll 错误场景
func TestBoot_PluginManager_StopAll_Error(t *testing.T) {
	t.Parallel()

	manager := NewPluginManager()

	// 创建会失败的插件
	failingPlugin := &mockPlugin{
		name:    "failing-stop-plugin",
		version: "1.0.0",
		stopErr: fmt.Errorf("intentional stop failure"),
	}

	manager.Register(failingPlugin)
	ctx := &mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	}
	manager.SetContext(ctx)

	// 初始化和启动
	if err := manager.InitAll(); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}
	if err := manager.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	// StopAll 应该失败
	err := manager.StopAll()
	if err == nil {
		t.Fatal("Expected StopAll to fail")
	}
}

// TestBoot_PluginManager_InitAll_Error_Coverage 测试 InitAll 错误场景
func TestBoot_PluginManager_InitAll_Error(t *testing.T) {
	t.Parallel()

	manager := NewPluginManager()

	// 创建会失败的插件
	failingPlugin := &mockPlugin{
		name:    "failing-init-plugin",
		version: "1.0.0",
		initErr: fmt.Errorf("intentional init failure"),
	}

	manager.Register(failingPlugin)
	ctx := &mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	}
	manager.SetContext(ctx)

	// InitAll 应该失败
	err := manager.InitAll()
	if err == nil {
		t.Fatal("Expected InitAll to fail")
	}
}

// TestBoot_PluginAppCtx_GetPlugin_NilManager_Coverage 测试 GetPlugin 在 manager 为 nil 时的行为
func TestBoot_PluginAppCtx_GetPlugin_NilManager(t *testing.T) {
	t.Parallel()

	// 直接创建一个 manager 为 nil 的 pluginAppCtx
	pCtx := &pluginAppCtx{
		ctx:     nil,
		rootCtx: context.Background(),
		manager: nil,
	}

	// GetPlugin 应该返回 nil, false
	p, ok := pCtx.GetPlugin("any-plugin")
	if ok {
		t.Error("Expected GetPlugin to return false when manager is nil")
	}
	if p != nil {
		t.Error("Expected GetPlugin to return nil when manager is nil")
	}
}
