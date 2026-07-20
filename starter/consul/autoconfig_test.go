package consul

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestConsulConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-consul", environment.PriorityNormal, map[string]any{
		ConsulEnabled: "true",
		ConsulHost:    "192.168.1.100",
		ConsulPort:    "8501",
		ConsulToken:   "test-token",
	}))

	cfg := &ConsulConfig{
		Host: DefaultConsulHost,
		Port: DefaultConsulPort,
	}

	err := env.BindPrefix("consul", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected consul.enabled to be true")
	}
	if cfg.Host != "192.168.1.100" {
		t.Errorf("expected host '192.168.1.100', got '%s'", cfg.Host)
	}
	if cfg.Port != 8501 {
		t.Errorf("expected port 8501, got %d", cfg.Port)
	}
	if cfg.Token != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", cfg.Token)
	}
}

func TestConsulConfig_DefaultValues(t *testing.T) {
	cfg := &ConsulConfig{
		Host: DefaultConsulHost,
		Port: DefaultConsulPort,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 8500 {
		t.Errorf("expected default port 8500, got %d", cfg.Port)
	}
}
