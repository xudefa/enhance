package gin

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestGinConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-gin", environment.PriorityNormal, map[string]any{
		"gin.enabled":        "true",
		"gin.host":           "127.0.0.1",
		"gin.port":           "9090",
		"gin.mode":           "release",
		"gin.enable_recover": "false",
		"gin.enable_logger":  "false",
	}))

	autoConfig := &GinAutoConfiguration{}
	cfg, err := autoConfig.loadConfig(env)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected gin.enabled to be true")
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Mode != "release" {
		t.Errorf("expected mode 'release', got '%s'", cfg.Mode)
	}
	if cfg.EnableRecover {
		t.Error("expected enable_recover to be false")
	}
	if cfg.EnableLogger {
		t.Error("expected enable_logger to be false")
	}
}

func TestGinConfig_DefaultValues(t *testing.T) {
	env := environment.NewEnvironment()

	autoConfig := &GinAutoConfiguration{}
	cfg, err := autoConfig.loadConfig(env)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default host '0.0.0.0', got '%s'", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.Mode != "debug" {
		t.Errorf("expected default mode 'debug', got '%s'", cfg.Mode)
	}
	if !cfg.EnableRecover {
		t.Error("expected default enable_recover to be true")
	}
	if !cfg.EnableLogger {
		t.Error("expected default enable_logger to be true")
	}
}
