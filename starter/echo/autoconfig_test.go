package echo

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestEchoConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-echo", environment.PriorityNormal, map[string]any{
		"echo.enabled":     "true",
		"echo.host":        "127.0.0.1",
		"echo.port":        9090,
		"echo.hide_banner": "true",
	}))

	cfg := &EchoConfig{
		Host:          "0.0.0.0",
		Port:          8080,
		HideBanner:    false,
		HidePort:      false,
		EnableRecover: true,
		EnableLogger:  true,
		EnableCORS:    false,
	}

	err := env.BindPrefix("echo", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected echo.enabled to be true")
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if !cfg.HideBanner {
		t.Error("expected hide_banner to be true")
	}
}

func TestEchoConfig_DefaultValues(t *testing.T) {
	cfg := &EchoConfig{
		Host:          "0.0.0.0",
		Port:          8080,
		HideBanner:    false,
		HidePort:      false,
		EnableRecover: true,
		EnableLogger:  true,
		EnableCORS:    false,
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
