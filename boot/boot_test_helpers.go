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

func (m *mockPlugin) Name() string                  { return m.name }
func (m *mockPlugin) Version() string               { return m.version }
func (m *mockPlugin) Dependencies() []string        { return m.dependencies }
func (m *mockPlugin) Init(ctx PluginContext) error  { m.initCalled = true; return m.initErr }
func (m *mockPlugin) Start(ctx PluginContext) error { m.startCalled = true; return m.startErr }
func (m *mockPlugin) Stop(ctx PluginContext) error  { m.stopCalled = true; return m.stopErr }

// mockPluginContext is a mock PluginContext for testing.
type mockPluginContext struct {
	container   core.Container
	environment *environment.Environment
	plugins     map[string]Plugin
	config      map[string]any
}

func (m *mockPluginContext) Container() core.Container             { return m.container }
func (m *mockPluginContext) Environment() *environment.Environment { return m.environment }
func (m *mockPluginContext) GetPlugin(name string) (Plugin, bool) {
	p, ok := m.plugins[name]
	return p, ok
}
func (m *mockPluginContext) Config(key string) (any, bool) {
	v, ok := m.config[key]
	return v, ok
}

// testPlugin is a basic mock Plugin for testing.
type testPlugin struct {
	name    string
	version string
}

func (p *testPlugin) Name() string                  { return p.name }
func (p *testPlugin) Version() string               { return p.version }
func (p *testPlugin) Dependencies() []string        { return nil }
func (p *testPlugin) Init(ctx PluginContext) error  { return nil }
func (p *testPlugin) Start(ctx PluginContext) error { return nil }
func (p *testPlugin) Stop(ctx PluginContext) error  { return nil }

// testPluginWithCallback is a mock Plugin with customizable callbacks.
type testPluginWithCallback struct {
	name        string
	version     string
	onInit      func(PluginContext) error
	onStart     func(PluginContext) error
	initCalled  bool
	startCalled bool
}

func (p *testPluginWithCallback) Name() string           { return p.name }
func (p *testPluginWithCallback) Version() string        { return p.version }
func (p *testPluginWithCallback) Dependencies() []string { return nil }
func (p *testPluginWithCallback) Init(ctx PluginContext) error {
	p.initCalled = true
	if p.onInit != nil {
		return p.onInit(ctx)
	}
	return nil
}
func (p *testPluginWithCallback) Start(ctx PluginContext) error {
	p.startCalled = true
	if p.onStart != nil {
		return p.onStart(ctx)
	}
	return nil
}
func (p *testPluginWithCallback) Stop(ctx PluginContext) error { return nil }

// mockConfigCenter implements config.ConfigCenter interface for testing.
type mockConfigCenter struct {
	loadData config.ConfigData
	closeErr error
}

func (m *mockConfigCenter) Load() (config.ConfigData, error)                         { return m.loadData, nil }
func (m *mockConfigCenter) Watch(key string, callback func(config.ConfigData)) error { return nil }
func (m *mockConfigCenter) Close() error                                             { return m.closeErr }

// mockContainerForHas is a mock Container for testing.
type mockContainerForHas struct {
	listBeansResult map[string]*registry.BeanDef
	getAllResult    []any
	generateResult  string
}

func (m *mockContainerForHas) ListBeans() map[string]*registry.BeanDef { return m.listBeansResult }
func (m *mockContainerForHas) GetAll() []any                           { return m.getAllResult }
func (m *mockContainerForHas) Generate(typ reflect.Type, customName ...string) string {
	return m.generateResult
}
func (m *mockContainerForHas) Get(typ reflect.Type) ([]any, error) { return nil, nil }
func (m *mockContainerForHas) GetByTypeAndName(name string, typ reflect.Type) (any, error) {
	return nil, nil
}
func (m *mockContainerForHas) Has(name string, typ reflect.Type) bool  { return false }
func (m *mockContainerForHas) HasType(typ reflect.Type) bool           { return false }
func (m *mockContainerForHas) Types() []reflect.Type                   { return nil }
func (m *mockContainerForHas) RegisterBean(def registry.BeanDef) error { return nil }
func (m *mockContainerForHas) RegisterInstance(instance any, typ reflect.Type) error {
	return nil
}
func (m *mockContainerForHas) CreateBean(beanID string) (any, error) { return nil, nil }
func (m *mockContainerForHas) Initialize() error                     { return nil }
func (m *mockContainerForHas) Destroy() error                        { return nil }
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

func (s *TestStarter) Name() string {
	if s.name != "" {
		return s.name
	}
	return "test-starter"
}
func (s *TestStarter) Dependencies() []string { return s.dependencies }
func (s *TestStarter) Configure(ctx ApplicationContext) error {
	if s.onConfigure != nil {
		s.onConfigure()
	}
	return nil
}
func (s *TestStarter) Start(ctx ApplicationContext) error {
	if s.onStart != nil {
		s.onStart()
	}
	return nil
}
func (s *TestStarter) Stop(ctx ApplicationContext) error {
	if s.onStop != nil {
		s.onStop()
	}
	return nil
}

func (s *TestStarter) GetCondition() condition.Condition {
	return condition.OnPropertyOrDefault("test.enabled", "true", "true")
}

// TestBean is a simple test bean.
type TestBean struct{ Name string }

// TestBean2 is another test bean.
type TestBean2 struct{ Name string }
