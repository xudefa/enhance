package boot

import (
	"os"
	"path/filepath"
	"testing"

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
	value, ok := app.Environment().GetProperty("key1")
	if !ok || value != "value1" {
		t.Errorf("Expected key1='value1', got %v", value)
	}

	value, ok = app.Environment().GetProperty("key2")
	if !ok || value != "value2" {
		t.Errorf("Expected key2='value2', got %v", value)
	}
}
