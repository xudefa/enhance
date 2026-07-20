package chi

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestChiConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-chi", environment.PriorityNormal, map[string]any{
		"chi.enabled":        "true",
		"chi.host":           "127.0.0.1",
		"chi.port":           9090,
		"chi.enable_real_ip": "true",
	}))

	cfg := &ChiConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		EnableRecover:   true,
		EnableLogger:    true,
		EnableRequestID: true,
		EnableRealIP:    false,
	}

	err := env.BindPrefix("chi", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected chi.enabled to be true")
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if !cfg.EnableRealIP {
		t.Error("expected enable_real_ip to be true")
	}
}

func TestChiConfig_DefaultValues(t *testing.T) {
	cfg := &ChiConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		EnableRecover:   true,
		EnableLogger:    true,
		EnableRequestID: true,
		EnableRealIP:    false,
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default host '0.0.0.0', got '%s'", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if !cfg.EnableRecover {
		t.Error("expected default enable_recover to be true")
	}
}
