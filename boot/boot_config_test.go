package boot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xudefa/enhance/config"
	"github.com/xudefa/enhance/config/environment"
)

func TestBoot_BindConfig(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Host string `config:"server.host"`
	}
	boot, err := NewApplication(WithAppName("test-app"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}
	boot.ctx.Environment().AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"server.host": "localhost",
	}))
	cfg, err := BindConfig[TestConfig](boot)
	if err != nil {
		t.Fatalf("BindConfig() error = %v", err)
	}
	if cfg.Host != "localhost" {
		t.Errorf("BindConfig().Host = %q, want %q", cfg.Host, "localhost")
	}
}

func TestBoot_BindConfig_Coverage(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Name string `config:"app.name"`
		Port int    `config:"app.port"`
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
			"app.name": "bind-test",
			"app.port": "9090",
		})),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	cfg, err := BindConfig[TestConfig](app)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}
	if cfg.Name != "bind-test" {
		t.Errorf("Expected Name 'bind-test', got %s", cfg.Name)
	}
	if cfg.Port != 9090 {
		t.Errorf("Expected Port 9090, got %d", cfg.Port)
	}
}

func TestBoot_BindConfigPrefix(t *testing.T) {
	t.Parallel()
	type ServerConfig struct {
		Host string
		Port int
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
			"server.host": "localhost",
			"server.port": "8080",
		})),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	cfg, err := BindConfig[ServerConfig](app, WithConfigPrefix("server"))
	if err != nil {
		t.Fatalf("BindConfigPrefix failed: %v", err)
	}
	if cfg.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port 8080, got %d", cfg.Port)
	}
}

func TestBoot_GetProperty(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
			"app.name": "property-test",
		})),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
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

func TestBoot_GetRequiredProperty(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
			"app.name": "required-test",
		})),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
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

func TestBoot_WithConfigFile(t *testing.T) {
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
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigLocation(absConfigFile),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	name, ok := app.Environment().GetProperty("app.name")
	if !ok || name != "config-file-test" {
		t.Errorf("Expected app.name='config-file-test', got %v", name)
	}
}

func TestBoot_WithPropertySource(t *testing.T) {
	t.Parallel()
	source := environment.NewMapPropertySource("custom", environment.PriorityNormal, map[string]any{
		"custom.key": "custom-value",
	})
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPropertySource(source),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	value, ok := app.Environment().GetProperty("custom.key")
	if !ok || value != "custom-value" {
		t.Errorf("Expected custom.key='custom-value', got %v", value)
	}
}

func TestBoot_WithProperty(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProperty("test.key", "test-value"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	value, ok := app.Environment().GetProperty("test.key")
	if !ok || value != "test-value" {
		t.Errorf("Expected test.key='test-value', got %v", value)
	}
}

func TestBoot_WithProperties(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProperties("key1", "value1", "key2", "value2"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	value1, ok := app.Environment().GetProperty("key1")
	if !ok || value1 != "value1" {
		t.Errorf("Expected key1='value1', got %v", value1)
	}
	value2, ok := app.Environment().GetProperty("key2")
	if !ok || value2 != "value2" {
		t.Errorf("Expected key2='value2', got %v", value2)
	}
}

func TestBoot_WithPropertiesOddArgs_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with odd number of arguments")
		}
	}()
	_ = WithProperties("key1", "value1", "key2")
}

func TestBoot_Config(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("my-app"), WithVersion("1.2.3"), WithoutAutoConfig(), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
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

func TestBoot_WithConfigType(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigType("json"))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if app.config.ConfigType != "json" {
		t.Errorf("Expected ConfigType 'json', got %s", app.config.ConfigType)
	}
}

func TestBoot_WithProfiles(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithProfiles("custom"))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if len(app.config.Profiles) == 0 {
		t.Fatal("Expected profiles to be set")
	}
}

func TestBoot_WithProfilesMultiple(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithProfiles("dev", "local", "test"))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if len(app.config.Profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(app.config.Profiles))
	}
}

func TestBoot_WithAutoConfigExclude(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutStarters(), WithExclude("TestAutoConfig"))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if len(app.config.ExcludedAutoConfigs) == 0 {
		t.Error("Expected ExcludedAutoConfigs to be set")
	}
}

func TestBoot_WithExcludeMultiple(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutStarters(), WithExclude("Config1", "Config2", "Config3"))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if len(app.config.ExcludedAutoConfigs) != 3 {
		t.Errorf("Expected 3 excluded configs, got %d", len(app.config.ExcludedAutoConfigs))
	}
}

func TestBoot_CollectAutoConfigReport(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
}

func TestBoot_CollectAutoConfigReport_WithAutoConfig(t *testing.T) {
	t.Parallel()
	EnableAutoConfigReport()
	defer DisableAutoConfigReport()
	app, err := NewApplication(WithAppName("test-app"), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !app.IsRunning() {
		t.Error("Expected app to be running")
	}
}

func TestBoot_CollectAutoConfigReport_WithDebug(t *testing.T) {
	t.Parallel()
	DisableAutoConfigReport()
	app, err := NewApplication(WithAppName("test-app"), WithoutStarters(), WithProperty("enhance.debug", true))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !app.IsRunning() {
		t.Error("Expected app to be running")
	}
}

func TestBoot_WithConfigCenter(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters())
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if app.config.ConfigCenterEnabled {
		t.Error("Expected ConfigCenterEnabled to be false by default")
	}
}

func TestBoot_WithConfigCenterOptions(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848"},
			WithConfigCenterDataID("test-data"),
			WithConfigCenterGroup("test-group"),
			WithConfigCenterPrefix("test"),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if !app.config.ConfigCenterEnabled {
		t.Error("Expected ConfigCenterEnabled to be true")
	}
	if app.config.ConfigCenterType != "nacos" {
		t.Errorf("Expected ConfigCenterType 'nacos', got %s", app.config.ConfigCenterType)
	}
	if app.config.ConfigCenterDataID != "test-data" {
		t.Errorf("Expected ConfigCenterDataID 'test-data', got %s", app.config.ConfigCenterDataID)
	}
}

func TestBoot_WithConfigCenterTimeout(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("etcd", []string{"localhost:2379"}, WithConfigCenterTimeout(10000000000)),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if app.config.ConfigCenterTimeout != 10000000000 {
		t.Errorf("Expected timeout 10s, got %v", app.config.ConfigCenterTimeout)
	}
}

func TestBoot_LoadConfigCenterConfig(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("nacos", []string{"localhost:8848"}))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	err = app.loadConfigCenterConfig()
	if err == nil {
		t.Error("Expected error when no factory registered")
	}
}

func TestBoot_LoadConfigCenterConfig_UnsupportedType(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("unsupported-type", []string{"localhost:8888"}))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	err = app.loadConfigCenterConfig()
	if err == nil {
		t.Error("Expected error for unsupported config center type")
	}
}

func TestBoot_WithConfigCenterEnabled(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("nacos", []string{"localhost:8848"}))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if !app.config.ConfigCenterEnabled {
		t.Error("Expected ConfigCenterEnabled to be true")
	}
	if app.config.ConfigCenterType != "nacos" {
		t.Errorf("Expected ConfigCenterType 'nacos', got %s", app.config.ConfigCenterType)
	}
	if len(app.config.ConfigCenterAddr) != 1 || app.config.ConfigCenterAddr[0] != "localhost:8848" {
		t.Errorf("Expected ConfigCenterAddr ['localhost:8848'], got %v", app.config.ConfigCenterAddr)
	}
}

func TestBoot_WithConfigCenterDataID(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("nacos", []string{"localhost:8848"}, WithConfigCenterDataID("my-data-id")))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if app.config.ConfigCenterDataID != "my-data-id" {
		t.Errorf("Expected ConfigCenterDataID 'my-data-id', got %s", app.config.ConfigCenterDataID)
	}
}

func TestBoot_WithConfigCenterGroup(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("nacos", []string{"localhost:8848"}, WithConfigCenterGroup("my-group")))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if app.config.ConfigCenterGroup != "my-group" {
		t.Errorf("Expected ConfigCenterGroup 'my-group', got %s", app.config.ConfigCenterGroup)
	}
}

func TestBoot_WithConfigCenterPrefix(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("nacos", []string{"localhost:8848"}, WithConfigCenterPrefix("my-prefix")))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if app.config.ConfigCenterPrefix != "my-prefix" {
		t.Errorf("Expected ConfigCenterPrefix 'my-prefix', got %s", app.config.ConfigCenterPrefix)
	}
}

func TestBoot_WithConfigCenterMultipleAddrs(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848", "localhost:8849", "localhost:8850"}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if len(app.config.ConfigCenterAddr) != 3 {
		t.Errorf("Expected 3 addresses, got %d", len(app.config.ConfigCenterAddr))
	}
}

func TestBoot_LoadConfigCenterConfig_Success(t *testing.T) {
	t.Parallel()
	RegisterConfigCenterFactory("mock-success", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return &mockConfigCenter{loadData: config.ConfigData{"mock.key": "mock-value"}}, nil
	})
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("mock-success", []string{"localhost:8848"}))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	value, ok := app.Environment().GetProperty("mock.key")
	if !ok || value != "mock-value" {
		t.Errorf("Expected mock.key='mock-value', got %v", value)
	}
}

func TestBoot_LoadConfigCenterConfig_FactoryError(t *testing.T) {
	t.Parallel()
	RegisterConfigCenterFactory("mock-factory-error", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return nil, fmt.Errorf("factory error")
	})
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("mock-factory-error", []string{"localhost:8848"}))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	err = app.Start()
	if err == nil {
		t.Fatal("Expected error when factory fails")
	}
}

func TestBoot_LoadConfigCenterConfig_EmptyData(t *testing.T) {
	t.Parallel()
	RegisterConfigCenterFactory("mock-empty-data", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return &mockConfigCenter{loadData: config.ConfigData{}}, nil
	})
	app, err := NewApplication(WithAppName("test-app"), WithoutAutoConfig(), WithoutStarters(), WithConfigCenter("mock-empty-data", []string{"localhost:8848"}))
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
}
