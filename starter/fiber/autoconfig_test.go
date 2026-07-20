package fiber

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestFiberConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-fiber", environment.PriorityNormal, map[string]any{
		"fiber.enabled":    "true",
		"fiber.host":       "127.0.0.1",
		"fiber.port":       "8080",
		"fiber.prefork":    "true",
		"fiber.body-limit": "8388608",
	}))

	cfg := &FiberConfig{
		Host:        "0.0.0.0",
		Port:        3000,
		BodyLimit:   4194304,
		Concurrency: 262144,
	}

	err := env.BindPrefix("fiber", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected fiber.enabled to be true")
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.BodyLimit != 8388608 {
		t.Errorf("expected body-limit 8388608, got %d", cfg.BodyLimit)
	}
}

func TestFiberConfig_DefaultValues(t *testing.T) {
	cfg := &FiberConfig{
		Host:        "0.0.0.0",
		Port:        3000,
		BodyLimit:   4194304,
		Concurrency: 262144,
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default host '0.0.0.0', got '%s'", cfg.Host)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected default port 3000, got %d", cfg.Port)
	}
	if cfg.BodyLimit != 4194304 {
		t.Errorf("expected default body-limit 4194304, got %d", cfg.BodyLimit)
	}
}
