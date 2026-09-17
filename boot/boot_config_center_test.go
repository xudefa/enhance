package boot

import (
	"context"
	"fmt"
	"testing"

	"github.com/xudefa/enhance/config"
)

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
