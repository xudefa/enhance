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

// TestBoot_BindConfig_Coverage 测试 BindConfig 方法
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

	// 测试 BindConfig
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

// TestBoot_BindConfigPrefix_Coverage 测试 BindConfigPrefix 方法
func TestBoot_BindConfigPrefix_Coverage(t *testing.T) {
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

	// 测试 BindConfigPrefix
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

// TestBoot_GetProperty_Coverage 测试 GetProperty 方法
func TestBoot_GetProperty_Coverage(t *testing.T) {
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

	// 测试 GetProperty
	name, ok := app.Environment().GetProperty("app.name")
	if !ok {
		t.Error("Expected GetProperty to return true")
	}
	if name != "property-test" {
		t.Errorf("Expected 'property-test', got %v", name)
	}

	// 测试不存在的属性
	_, ok = app.Environment().GetProperty("non.existent")
	if ok {
		t.Error("Expected GetProperty to return false for non-existent key")
	}
}

// TestBoot_GetRequiredProperty_Coverage 测试 GetRequiredProperty 方法
func TestBoot_GetRequiredProperty_Coverage(t *testing.T) {
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

	// 测试 GetRequiredProperty
	name, err := app.Environment().GetRequiredProperty("app.name")
	if err != nil {
		t.Fatalf("GetRequiredProperty failed: %v", err)
	}
	if name != "required-test" {
		t.Errorf("Expected 'required-test', got %v", name)
	}

	// 测试不存在的属性（应返回错误）
	_, err = app.Environment().GetRequiredProperty("non.existent")
	if err == nil {
		t.Error("Expected error for non-existent property")
	}
}

// TestBoot_WithConfigFile_Coverage 测试 WithConfigLocation 选项
func TestBoot_WithConfigFile(t *testing.T) {
	t.Parallel()

	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "application.json")
	configContent := `{"app": {"name": "config-file-test", "port": 7070}}`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 使用绝对路径
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

	// 验证配置已加载（JSON 嵌套结构）
	name, ok := app.Environment().GetProperty("app.name")
	if !ok || name != "config-file-test" {
		t.Errorf("Expected app.name='config-file-test', got %v", name)
	}
}

// TestBoot_WithPropertySource_Coverage 测试 WithPropertySource
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

	// 验证自定义属性源已添加
	value, ok := app.Environment().GetProperty("custom.key")
	if !ok || value != "custom-value" {
		t.Errorf("Expected custom.key='custom-value', got %v", value)
	}
}

// TestBoot_WithProfiles_Coverage 测试 WithProfiles 选项
func TestBoot_WithProfiles(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProfiles("custom"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.Profiles) == 0 {
		t.Fatal("Expected profiles to be set")
	}
}

// TestBoot_WithConfigCenter_Coverage 测试配置中心相关选项
func TestBoot_WithConfigCenter(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.ConfigCenterEnabled {
		t.Error("Expected ConfigCenterEnabled to be false by default")
	}
}

// TestBoot_WithAutoConfigExclude_Coverage 测试自动配置排除
func TestBoot_WithAutoConfigExclude(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithExclude("TestAutoConfig"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.ExcludedAutoConfigs) == 0 {
		t.Error("Expected ExcludedAutoConfigs to be set")
	}
}

// TestBoot_WithProperty_Coverage 测试 WithProperty 选项
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

	// 验证属性已添加
	value, ok := app.Environment().GetProperty("test.key")
	if !ok || value != "test-value" {
		t.Errorf("Expected test.key='test-value', got %v", value)
	}
}

// TestBoot_WithProperties_Coverage 测试 WithProperties 选项
func TestBoot_WithProperties(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProperties(
			"key1", "value1",
			"key2", "value2",
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证属性已添加
	value1, ok := app.Environment().GetProperty("key1")
	if !ok || value1 != "value1" {
		t.Errorf("Expected key1='value1', got %v", value1)
	}

	value2, ok := app.Environment().GetProperty("key2")
	if !ok || value2 != "value2" {
		t.Errorf("Expected key2='value2', got %v", value2)
	}
}

// TestBoot_Config_Coverage 测试 Config 方法
func TestBoot_Config(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("my-app"),
		WithVersion("1.2.3"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
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

// TestBoot_CollectAutoConfigReport_Coverage 测试 collectAutoConfigReport 方法
func TestBoot_CollectAutoConfigReport(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 这个方法在 Start 时会被调用，这里只验证不会 panic
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()
}

// TestBoot_BindConfigWithOptions_Coverage 测试 BindConfig 带选项
func TestBoot_BindConfigWithOptions(t *testing.T) {
	t.Parallel()

	type DatabaseConfig struct {
		Host string
		Port int
	}

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
			"database.host": "localhost",
			"database.port": "5432",
		})),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 测试带前缀的配置绑定
	cfg, err := BindConfig[DatabaseConfig](app, WithConfigPrefix("database"))
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected Port 5432, got %d", cfg.Port)
	}
}

// TestBoot_WithConfigType_Coverage 测试 WithConfigType 选项
func TestBoot_WithConfigType_Coverage(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigType("json"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.ConfigType != "json" {
		t.Errorf("Expected ConfigType 'json', got %s", app.config.ConfigType)
	}
}

// TestBoot_WithExcludeMultiple_Coverage 测试排除多个自动配置
func TestBoot_WithExcludeMultiple(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithExclude("Config1", "Config2", "Config3"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.ExcludedAutoConfigs) != 3 {
		t.Errorf("Expected 3 excluded configs, got %d", len(app.config.ExcludedAutoConfigs))
	}
}

// TestBoot_WithPropertiesOddArgs_Panic_Coverage 测试 WithProperties 奇数参数 panic
func TestBoot_WithPropertiesOddArgs_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with odd number of arguments")
		}
	}()

	_ = WithProperties("key1", "value1", "key2")
}

// TestBoot_WithConfigCenterOptions_Coverage 测试配置中心选项
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

// TestBoot_WithConfigCenterTimeout_Coverage 测试配置中心超时设置
func TestBoot_WithConfigCenterTimeout(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("etcd", []string{"localhost:2379"},
			WithConfigCenterTimeout(10000000000), // 10s
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.ConfigCenterTimeout != 10000000000 {
		t.Errorf("Expected timeout 10s, got %v", app.config.ConfigCenterTimeout)
	}
}

// TestBoot_LoadConfigCenterConfig_Coverage 测试 loadConfigCenterConfig 方法
func TestBoot_LoadConfigCenterConfig(t *testing.T) {
	t.Parallel()

	// 测试没有配置中心地址时的错误
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848"}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 手动调用 loadConfigCenterConfig 应该返回错误（因为没有注册工厂）
	err = app.loadConfigCenterConfig()
	if err == nil {
		t.Error("Expected error when no factory registered")
	}
}

// TestBoot_LoadConfigCenterConfig_UnsupportedType_Coverage 测试不支持的配置中心类型
func TestBoot_LoadConfigCenterConfig_UnsupportedType(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("unsupported-type", []string{"localhost:8888"}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// 手动调用 loadConfigCenterConfig 应该返回错误
	err = app.loadConfigCenterConfig()
	if err == nil {
		t.Error("Expected error for unsupported config center type")
	}
}

// TestBoot_WithConfigCenterEnabled_Coverage 测试 WithConfigCenterEnabled 选项
func TestBoot_WithConfigCenterEnabled(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848"}),
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
	if len(app.config.ConfigCenterAddr) != 1 || app.config.ConfigCenterAddr[0] != "localhost:8848" {
		t.Errorf("Expected ConfigCenterAddr ['localhost:8848'], got %v", app.config.ConfigCenterAddr)
	}
}

// TestBoot_WithConfigCenterDataID_Coverage 测试 WithConfigCenterDataID 选项
func TestBoot_WithConfigCenterDataID(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848"},
			WithConfigCenterDataID("my-data-id"),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.ConfigCenterDataID != "my-data-id" {
		t.Errorf("Expected ConfigCenterDataID 'my-data-id', got %s", app.config.ConfigCenterDataID)
	}
}

// TestBoot_WithConfigCenterGroup_Coverage 测试 WithConfigCenterGroup 选项
func TestBoot_WithConfigCenterGroup(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848"},
			WithConfigCenterGroup("my-group"),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.ConfigCenterGroup != "my-group" {
		t.Errorf("Expected ConfigCenterGroup 'my-group', got %s", app.config.ConfigCenterGroup)
	}
}

// TestBoot_WithConfigCenterPrefix_Coverage 测试 WithConfigCenterPrefix 选项
func TestBoot_WithConfigCenterPrefix(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848"},
			WithConfigCenterPrefix("my-prefix"),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.ConfigCenterPrefix != "my-prefix" {
		t.Errorf("Expected ConfigCenterPrefix 'my-prefix', got %s", app.config.ConfigCenterPrefix)
	}
}

// TestBoot_CollectAutoConfigReport_WithAutoConfig_Coverage 测试 collectAutoConfigReport 在有自动配置时被调用
func TestBoot_CollectAutoConfigReport_WithAutoConfig(t *testing.T) {
	t.Parallel()

	// 启用自动配置报告
	EnableAutoConfigReport()
	defer DisableAutoConfigReport()

	// 创建一个带自动配置的应用
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// Start 时会调用 collectAutoConfigReport
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证应用正常启动
	if !app.IsRunning() {
		t.Error("Expected app to be running")
	}
}

// TestBoot_CollectAutoConfigReport_WithDebug_Coverage 测试 debug 模式下 collectAutoConfigReport 被调用
func TestBoot_CollectAutoConfigReport_WithDebug(t *testing.T) {
	t.Parallel()

	// 确保报告开关关闭，使用 debug 模式
	DisableAutoConfigReport()

	// 创建一个带 debug 模式的应用
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithProperty("enhance.debug", true),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// Start 时会调用 collectAutoConfigReport（因为 debug=true）
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证应用正常启动
	if !app.IsRunning() {
		t.Error("Expected app to be running")
	}
}

// TestBoot_WithConfigCenterTimeout_Coverage2 测试 WithConfigCenterTimeout 选项
func TestBoot_WithConfigCenterTimeout_Extended(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{"localhost:8848"},
			WithConfigCenterTimeout(5000000000), // 5s
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if app.config.ConfigCenterTimeout != 5000000000 {
		t.Errorf("Expected ConfigCenterTimeout 5s, got %v", app.config.ConfigCenterTimeout)
	}
}

// TestBoot_WithConfigCenterMultipleAddrs_Coverage 测试多个配置中心地址
func TestBoot_WithConfigCenterMultipleAddrs(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("nacos", []string{
			"localhost:8848",
			"localhost:8849",
			"localhost:8850",
		}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.ConfigCenterAddr) != 3 {
		t.Errorf("Expected 3 addresses, got %d", len(app.config.ConfigCenterAddr))
	}
}

// TestBoot_WithPropertiesMultiple_Coverage 测试 WithProperties 选项（多个属性）
func TestBoot_WithPropertiesMultiple(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProperties(
			"key1", "value1",
			"key2", "value2",
			"key3", "value3",
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证属性已设置
	val, ok := app.Environment().GetProperty("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected key1='value1', got %v", val)
	}

	val, ok = app.Environment().GetProperty("key2")
	if !ok || val != "value2" {
		t.Errorf("Expected key2='value2', got %v", val)
	}
}

// TestBoot_WithProfilesMultiple_Coverage 测试多个 Profile
func TestBoot_WithProfilesMultiple(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithProfiles("dev", "local", "test"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.Profiles) != 3 {
		t.Errorf("Expected 3 profiles, got %d", len(app.config.Profiles))
	}
}

// TestBoot_LoadConfigCenterConfig_Success_Coverage 测试 loadConfigCenterConfig 成功场景
func TestBoot_LoadConfigCenterConfig_Success(t *testing.T) {
	t.Parallel()

	// 注册 mock 工厂（使用已有的 mockConfigCenter）
	RegisterConfigCenterFactory("mock-success", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return &mockConfigCenter{
			loadData: config.ConfigData{
				"mock.key": "mock-value",
			},
		}, nil
	})

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("mock-success", []string{"localhost:8848"}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()

	// 验证配置已加载
	value, ok := app.Environment().GetProperty("mock.key")
	if !ok || value != "mock-value" {
		t.Errorf("Expected mock.key='mock-value', got %v", value)
	}
}

// TestBoot_LoadConfigCenterConfig_FactoryError_Coverage 测试 loadConfigCenterConfig 工厂错误
func TestBoot_LoadConfigCenterConfig_FactoryError(t *testing.T) {
	t.Parallel()

	// 注册会返回错误的工厂
	RegisterConfigCenterFactory("mock-factory-error", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return nil, fmt.Errorf("factory error")
	})

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("mock-factory-error", []string{"localhost:8848"}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// Start 应该失败，因为工厂返回错误
	err = app.Start()
	if err == nil {
		t.Fatal("Expected error when factory fails")
	}
}

// TestBoot_LoadConfigCenterConfig_EmptyData_Coverage 测试 loadConfigCenterConfig 空数据
func TestBoot_LoadConfigCenterConfig_EmptyData(t *testing.T) {
	t.Parallel()

	// 注册返回空数据的工厂
	RegisterConfigCenterFactory("mock-empty-data", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return &mockConfigCenter{
			loadData: config.ConfigData{},
		}, nil
	})

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithConfigCenter("mock-empty-data", []string{"localhost:8848"}),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	// Start 应该成功，即使配置中心返回空数据
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	defer app.Stop()
}
func TestBoot_WithExcludeMultiple_Extended(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithExclude("Config1", "Config2", "Config3", "Config4"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.ExcludedAutoConfigs) != 4 {
		t.Errorf("Expected 4 excluded configs, got %d", len(app.config.ExcludedAutoConfigs))
	}
}
