package boot

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

// newTestApp creates a new application with standard test options.
func newTestApp(t *testing.T, opts ...BootOption) *Boot {
	t.Helper()
	defaultOpts := []BootOption{
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	}
	app, err := NewApplication(append(defaultOpts, opts...)...)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	return app
}

// startTestApp starts the application.
func startTestApp(t *testing.T, app *Boot) {
	t.Helper()
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

// startAndStopTestApp starts the application and registers a cleanup to stop it.
func startAndStopTestApp(t *testing.T, app *Boot) {
	t.Helper()
	startTestApp(t, app)
	t.Cleanup(func() { app.Stop() })
}

// createTestPropertySource creates a simple property source for testing.
func createTestPropertySource(name, key, value string) *environment.MapPropertySource {
	return environment.NewMapPropertySource(name, environment.PriorityNormal, map[string]any{
		key: value,
	})
}

// mockPlugin is a mock Plugin for testing.
type mockPlugin struct {
	name         string
	version      string
	dependencies []string
	initCalled   bool
	startCalled  bool
	stopCalled   bool
	initErr      error
	startErr     error
	stopErr      error
}

// Name 返回 mock 插件的名称。
func (m *mockPlugin) Name() string { return m.name }

// Version 返回 mock 插件的版本号。
func (m *mockPlugin) Version() string { return m.version }

// Dependencies 返回 mock 插件的依赖列表。
func (m *mockPlugin) Dependencies() []string { return m.dependencies }

// Init 标记 mock 插件已初始化并返回预设的初始化错误。
func (m *mockPlugin) Init(ctx PluginContext) error { m.initCalled = true; return m.initErr }

// Start 标记 mock 插件已启动并返回预设的启动错误。
func (m *mockPlugin) Start(ctx PluginContext) error { m.startCalled = true; return m.startErr }

// Stop 标记 mock 插件已停止并返回预设的停止错误。
func (m *mockPlugin) Stop(ctx PluginContext) error { m.stopCalled = true; return m.stopErr }

// mockPluginContext is a mock PluginContext for testing.
type mockPluginContext struct {
	container   core.Container
	environment *environment.Environment
	plugins     map[string]Plugin
	config      map[string]any
}

// Container 返回 mock 插件上下文的容器。
func (m *mockPluginContext) Container() core.Container { return m.container }

// Environment 返回 mock 插件上下文的环境。
func (m *mockPluginContext) Environment() *environment.Environment { return m.environment }

// GetPlugin 按名称查找 mock 插件并返回其是否存在。
func (m *mockPluginContext) GetPlugin(name string) (Plugin, bool) {
	p, ok := m.plugins[name]
	return p, ok
}

// Config 按键读取 mock 配置并返回其是否存在。
func (m *mockPluginContext) Config(key string) (any, bool) {
	v, ok := m.config[key]
	return v, ok
}

// testPlugin is a basic mock Plugin for testing.
type testPlugin struct {
	name    string
	version string
}

// Name 返回测试插件的名称。
func (p *testPlugin) Name() string { return p.name }

// Version 返回测试插件的版本号。
func (p *testPlugin) Version() string { return p.version }

// Dependencies 返回测试插件的依赖列表。
func (p *testPlugin) Dependencies() []string { return nil }

// Init 初始化测试插件。
func (p *testPlugin) Init(ctx PluginContext) error { return nil }

// Start 启动测试插件。
func (p *testPlugin) Start(ctx PluginContext) error { return nil }

// Stop 停止测试插件。
func (p *testPlugin) Stop(ctx PluginContext) error { return nil }

// testPluginWithCallback is a mock Plugin with customizable callbacks.
type testPluginWithCallback struct {
	name        string
	version     string
	onInit      func(PluginContext) error
	onStart     func(PluginContext) error
	initCalled  bool
	startCalled bool
}

// Name 返回带回调测试插件的名称。
func (p *testPluginWithCallback) Name() string { return p.name }

// Version 返回带回调测试插件的版本号。
func (p *testPluginWithCallback) Version() string { return p.version }

// Dependencies 返回带回调测试插件的依赖列表。
func (p *testPluginWithCallback) Dependencies() []string { return nil }

// Init 初始化带回调测试插件，可选调用预设的初始化回调。
func (p *testPluginWithCallback) Init(ctx PluginContext) error {
	p.initCalled = true
	if p.onInit != nil {
		return p.onInit(ctx)
	}
	return nil
}

// Start 启动带回调测试插件，可选调用预设的启动回调。
func (p *testPluginWithCallback) Start(ctx PluginContext) error {
	p.startCalled = true
	if p.onStart != nil {
		return p.onStart(ctx)
	}
	return nil
}

// Stop 停止带回调测试插件。
func (p *testPluginWithCallback) Stop(ctx PluginContext) error { return nil }

// mockConfigCenter implements config.ConfigCenter interface for testing.
type mockConfigCenter struct {
	loadData config.ConfigData
	closeErr error
}

// Load 返回 mock 配置中心的预设配置数据。
func (m *mockConfigCenter) Load() (config.ConfigData, error) { return m.loadData, nil }

// Watch 订阅 mock 配置中心的配置变更。
func (m *mockConfigCenter) Watch(key string, callback func(config.ConfigData)) error { return nil }

// Close 关闭 mock 配置中心并返回预设的关闭错误。
func (m *mockConfigCenter) Close() error { return m.closeErr }

// mockContainerForHas is a mock Container for testing.
type mockContainerForHas struct {
	listBeansResult map[string]*registry.BeanDef
	getAllResult    []any
	generateResult  string
}

// ListBeans 返回 mock 容器的 Bean 定义列表。
func (m *mockContainerForHas) ListBeans() map[string]*registry.BeanDef { return m.listBeansResult }

// GetAll 返回 mock 容器中的全部 Bean 实例。
func (m *mockContainerForHas) GetAll() []any { return m.getAllResult }

// Generate 生成 mock 容器的 Bean 名称。
func (m *mockContainerForHas) Generate(typ reflect.Type, customName ...string) string {
	return m.generateResult
}

// Get 按类型返回 mock 容器中的 Bean 实例。
func (m *mockContainerForHas) Get(typ reflect.Type) ([]any, error) { return nil, nil }

// GetByTypeAndName 按类型和名称返回 mock 容器中的 Bean 实例。
func (m *mockContainerForHas) GetByTypeAndName(name string, typ reflect.Type) (any, error) {
	return nil, nil
}

// Has 判断 mock 容器中是否存在指定名称与类型的 Bean。
func (m *mockContainerForHas) Has(name string, typ reflect.Type) bool { return false }

// HasType 判断 mock 容器中是否存在指定类型的 Bean。
func (m *mockContainerForHas) HasType(typ reflect.Type) bool { return false }

// Types 返回 mock 容器中全部 Bean 的类型。
func (m *mockContainerForHas) Types() []reflect.Type { return nil }

// RegisterBean 向 mock 容器注册 Bean 定义。
func (m *mockContainerForHas) RegisterBean(def registry.BeanDef) error { return nil }

// RegisterInstance 向 mock 容器注册 Bean 实例。
func (m *mockContainerForHas) RegisterInstance(instance any, typ reflect.Type) error {
	return nil
}

// CreateBean 在 mock 容器中创建指定 ID 的 Bean。
func (m *mockContainerForHas) CreateBean(beanID string) (any, error) { return nil, nil }

// Initialize 初始化 mock 容器。
func (m *mockContainerForHas) Initialize() error { return nil }

// Destroy 销毁 mock 容器。
func (m *mockContainerForHas) Destroy() error { return nil }

// Parse 解析 mock 容器中 Bean 的包路径、类型名与自定义名。
func (m *mockContainerForHas) Parse(beanID string) (pkgPath, typeName, customName string) {
	return "", "", ""
}

// TestStarter is a test Starter.
type TestStarter struct {
	name         string
	dependencies []string
	onConfigure  func()
	onStart      func()
	onStop       func()
}

// Name 返回测试 Starter 的名称，未设置时使用默认名。
func (s *TestStarter) Name() string {
	if s.name != "" {
		return s.name
	}
	return "test-starter"
}

// Dependencies 返回测试 Starter 的依赖列表。
func (s *TestStarter) Dependencies() []string { return s.dependencies }

// Configure 执行测试 Starter 的配置阶段，可选调用预设回调。
func (s *TestStarter) Configure(ctx ApplicationContext) error {
	if s.onConfigure != nil {
		s.onConfigure()
	}
	return nil
}

// Start 启动测试 Starter，可选调用预设的启动回调。
func (s *TestStarter) Start(ctx ApplicationContext) error {
	if s.onStart != nil {
		s.onStart()
	}
	return nil
}

// Stop 停止测试 Starter，可选调用预设的停止回调。
func (s *TestStarter) Stop(ctx ApplicationContext) error {
	if s.onStop != nil {
		s.onStop()
	}
	return nil
}

// GetCondition 返回测试 Starter 的启用条件。
func (s *TestStarter) GetCondition() condition.Condition {
	return condition.OnPropertyOrDefault("test.enabled", "true", "true")
}

// TestBean is a simple test bean.
type TestBean struct{ Name string }

// TestBean2 is another test bean.
type TestBean2 struct{ Name string }
