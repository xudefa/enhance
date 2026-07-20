package micro

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestMicroConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-micro", environment.PriorityNormal, map[string]any{
		"micro.enabled":       "true",
		"micro.service_name":  "test-service",
		"micro.version":       "2.0.0",
		"micro.registry_addr": "consul://localhost:8500",
	}))

	cfg := &MicroConfig{
		ServiceName: DefaultServiceName,
		Version:     DefaultVersion,
	}

	err := env.BindPrefix("micro", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected micro.enabled to be true")
	}
	if cfg.ServiceName != "test-service" {
		t.Errorf("expected service_name 'test-service', got '%s'", cfg.ServiceName)
	}
	if cfg.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got '%s'", cfg.Version)
	}
	if cfg.RegistryAddr != "consul://localhost:8500" {
		t.Errorf("expected registry_addr 'consul://localhost:8500', got '%s'", cfg.RegistryAddr)
	}
}

func TestMicroConfig_DefaultValues(t *testing.T) {
	cfg := &MicroConfig{
		ServiceName: DefaultServiceName,
		Version:     DefaultVersion,
	}

	if cfg.ServiceName != "enhance-service" {
		t.Errorf("expected default service_name 'enhance-service', got '%s'", cfg.ServiceName)
	}
	if cfg.Version != "latest" {
		t.Errorf("expected default version 'latest', got '%s'", cfg.Version)
	}
}
