package boot

import (
	"fmt"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

// TestPluginManagerTests tests PluginManager functionality.
func TestPluginManagerTests(t *testing.T) {
	t.Parallel()

	t.Run("create and register", func(t *testing.T) {
		t.Parallel()
		testPluginManagerCreateRegister(t)
	})

	t.Run("init all", func(t *testing.T) {
		t.Parallel()
		testPluginManagerInitAll(t)
	})

	t.Run("start and stop all", func(t *testing.T) {
		t.Parallel()
		testPluginManagerStartStopAll(t)
	})
}

func testPluginManagerCreateRegister(t *testing.T) {
	t.Helper()
	manager := NewPluginManager()
	if manager == nil {
		t.Fatal("Expected non-nil PluginManager")
	}

	plugin := &mockPlugin{name: "test-plugin", version: "1.0.0"}
	manager.Register(plugin)

	if p, ok := manager.Get("test-plugin"); !ok || p == nil {
		t.Error("Expected to find test-plugin")
	}
	if plugins := manager.List(); len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
}

func testPluginManagerInitAll(t *testing.T) {
	t.Helper()
	manager := NewPluginManager()
	plugin := &mockPlugin{name: "plugin-a", version: "1.0.0"}
	manager.Register(plugin)
	manager.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})

	if err := manager.InitAll(); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}
}

func testPluginManagerStartStopAll(t *testing.T) {
	t.Helper()
	manager := NewPluginManager()
	plugin1 := &mockPlugin{name: "plugin-a", version: "1.0.0"}
	plugin2 := &mockPlugin{name: "plugin-b", version: "1.0.0"}
	manager.Register(plugin1)
	manager.Register(plugin2)
	manager.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})

	if err := manager.InitAll(); err != nil {
		t.Fatalf("InitAll failed: %v", err)
	}
	if err := manager.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}
	if !plugin1.startCalled || !plugin2.startCalled {
		t.Error("Expected all plugins Start to be called")
	}
	if err := manager.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}
	if !plugin1.stopCalled || !plugin2.stopCalled {
		t.Error("Expected all plugins Stop to be called")
	}
}

// TestPluginManagerErrors tests PluginManager error scenarios.
func TestPluginManagerErrors(t *testing.T) {
	t.Parallel()

	for _, tt := range testPluginManagerErrorCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testPluginManagerErrorCase(t, tt.plugin, tt.action)
		})
	}
}

func testPluginManagerErrorCases() []struct {
	name   string
	plugin *mockPlugin
	action func(manager *PluginManager) error
} {
	return []struct {
		name   string
		plugin *mockPlugin
		action func(manager *PluginManager) error
	}{
		{
			name:   "init error",
			plugin: &mockPlugin{name: "failing-init", version: "1.0.0", initErr: fmt.Errorf("intentional init failure")},
			action: func(m *PluginManager) error { return m.InitAll() },
		},
		{
			name:   "start error",
			plugin: &mockPlugin{name: "failing-start", version: "1.0.0", startErr: fmt.Errorf("intentional start failure")},
			action: func(m *PluginManager) error {
				if err := m.InitAll(); err != nil {
					return err
				}
				return m.StartAll()
			},
		},
		{
			name:   "stop error",
			plugin: &mockPlugin{name: "failing-stop", version: "1.0.0", stopErr: fmt.Errorf("intentional stop failure")},
			action: func(m *PluginManager) error {
				if err := m.InitAll(); err != nil {
					return err
				}
				if err := m.StartAll(); err != nil {
					return err
				}
				return m.StopAll()
			},
		},
	}
}

func testPluginManagerErrorCase(t *testing.T, plugin *mockPlugin, action func(manager *PluginManager) error) {
	t.Helper()
	manager := NewPluginManager()
	manager.Register(plugin)
	manager.SetContext(&mockPluginContext{
		container:   core.NewContainer(),
		environment: environment.NewEnvironment(),
	})
	if err := action(manager); err == nil {
		t.Fatal("Expected error, got nil")
	}
}
