package nacos

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestNacosConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-nacos", environment.PriorityNormal, map[string]any{
		"nacos.enabled":      "true",
		"nacos.server_addr":  "192.168.1.100",
		"nacos.port":         8849,
		"nacos.namespace_id": "dev",
	}))

	cfg := &NacosConfig{
		ServerAddr:  DefaultNacosAddr,
		Port:        DefaultNacosPort,
		NamespaceID: DefaultNamespaceID,
		TimeoutMs:   DefaultTimeoutMs,
		LogLevel:    DefaultLogLevel,
	}

	err := env.BindPrefix("nacos", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected nacos.enabled to be true")
	}
	if cfg.ServerAddr != "192.168.1.100" {
		t.Errorf("expected server_addr '192.168.1.100', got '%s'", cfg.ServerAddr)
	}
	if cfg.Port != 8849 {
		t.Errorf("expected port 8849, got %d", cfg.Port)
	}
	if cfg.NamespaceID != "dev" {
		t.Errorf("expected namespace_id 'dev', got '%s'", cfg.NamespaceID)
	}
}

func TestNacosConfig_DefaultValues(t *testing.T) {
	cfg := &NacosConfig{
		ServerAddr:  DefaultNacosAddr,
		Port:        DefaultNacosPort,
		NamespaceID: DefaultNamespaceID,
		TimeoutMs:   DefaultTimeoutMs,
		LogLevel:    DefaultLogLevel,
	}

	if cfg.ServerAddr != "127.0.0.1" {
		t.Errorf("expected default server_addr '127.0.0.1', got '%s'", cfg.ServerAddr)
	}
	if cfg.Port != 8848 {
		t.Errorf("expected default port 8848, got %d", cfg.Port)
	}
	if cfg.NamespaceID != "public" {
		t.Errorf("expected default namespace_id 'public', got '%s'", cfg.NamespaceID)
	}
}
