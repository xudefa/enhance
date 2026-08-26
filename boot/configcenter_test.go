package boot

import (
	"context"
	"testing"
	"time"

	"github.com/xudefa/enhance/config"
)

// mockConfigCenter 实现config.ConfigCenter接口
type mockConfigCenter struct {
	loadData config.ConfigData
	closeErr error
}

func (m *mockConfigCenter) Load() (config.ConfigData, error) {
	return m.loadData, nil
}

func (m *mockConfigCenter) Watch(key string, callback func(config.ConfigData)) error {
	return nil
}

func (m *mockConfigCenter) Close() error {
	return m.closeErr
}

func TestRegisterConfigCenterFactory(t *testing.T) {
	t.Parallel()

	// 保存原始工厂映射
	origFactories := make(map[string]ConfigCenterFactory)
	factoryMutex.RLock()
	for k, v := range configCenterFactories {
		origFactories[k] = v
	}
	factoryMutex.RUnlock()
	t.Cleanup(func() {
		factoryMutex.Lock()
		for k := range configCenterFactories {
			delete(configCenterFactories, k)
		}
		for k, v := range origFactories {
			configCenterFactories[k] = v
		}
		factoryMutex.Unlock()
	})

	// 注册一个mock工厂
	RegisterConfigCenterFactory("mock", func(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
		return &mockConfigCenter{loadData: map[string]any{"test.key": "test-value"}}, nil
	})

	// 验证工厂已注册
	factoryMutex.RLock()
	_, ok := configCenterFactories["mock"]
	factoryMutex.RUnlock()
	if !ok {
		t.Error("expected mock factory to be registered")
	}
}

func TestBoot_BuildConfigCenterConfig_Nacos(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("nacos", []string{"http://localhost:8848"},
			WithConfigCenterDataID("custom-data-id"),
			WithConfigCenterGroup("custom-group"),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := boot.buildConfigCenterConfig()
	if cfg.DataID != "custom-data-id" {
		t.Errorf("expected DataID 'custom-data-id', got '%s'", cfg.DataID)
	}
	if cfg.Group != "custom-group" {
		t.Errorf("expected Group 'custom-group', got '%s'", cfg.Group)
	}
}

func TestBoot_BuildConfigCenterConfig_Nacos_Defaults(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("nacos", []string{"http://localhost:8848"}),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := boot.buildConfigCenterConfig()
	if cfg.DataID != "app-config" {
		t.Errorf("expected default DataID 'app-config', got '%s'", cfg.DataID)
	}
	if cfg.Group != "DEFAULT_GROUP" {
		t.Errorf("expected default Group 'DEFAULT_GROUP', got '%s'", cfg.Group)
	}
}

func TestBoot_BuildConfigCenterConfig_Etcd(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("etcd", []string{"http://localhost:2379"},
			WithConfigCenterPrefix("/my/config"),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := boot.buildConfigCenterConfig()
	if cfg.Prefix != "/my/config" {
		t.Errorf("expected Prefix '/my/config', got '%s'", cfg.Prefix)
	}
}

func TestBoot_BuildConfigCenterConfig_Etcd_DefaultPrefix(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("etcd", []string{"http://localhost:2379"}),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := boot.buildConfigCenterConfig()
	if cfg.Prefix != "/config" {
		t.Errorf("expected default Prefix '/config', got '%s'", cfg.Prefix)
	}
}

func TestBoot_BuildConfigCenterConfig_Consul(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("consul", []string{"http://localhost:8500"},
			WithConfigCenterPrefix("consul-config"),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := boot.buildConfigCenterConfig()
	if cfg.Prefix != "consul-config" {
		t.Errorf("expected Prefix 'consul-config', got '%s'", cfg.Prefix)
	}
}

func TestBoot_BuildConfigCenterConfig_Consul_DefaultPrefix(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("consul", []string{"http://localhost:8500"}),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := boot.buildConfigCenterConfig()
	if cfg.Prefix != "config" {
		t.Errorf("expected default Prefix 'config', got '%s'", cfg.Prefix)
	}
}

func TestBoot_BuildConfigCenterConfig_UnknownType(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("unknown", []string{"http://localhost:8888"}),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	cfg := boot.buildConfigCenterConfig()
	// 未知类型应该只包含基础配置
	if len(cfg.Endpoints) == 0 {
		t.Error("expected Endpoints to be set")
	}
}

func TestBoot_BuildConfigCenterConfig_Timeout(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(
		WithAppName("test-app"),
		WithConfigCenter("nacos", []string{"http://localhost:8848"},
			WithConfigCenterTimeout(10*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	if boot.config.ConfigCenterTimeout != 10*time.Second {
		t.Errorf("expected ConfigCenterTimeout 10s, got %v", boot.config.ConfigCenterTimeout)
	}
}
