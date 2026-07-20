package zap

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestZapConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-zap", environment.PriorityNormal, map[string]any{
		"log.zap.enabled":     "true",
		"log.zap.level":       "debug",
		"log.zap.format":      "console",
		"log.zap.output-path": "logs/app.log",
	}))

	cfg := &ZapConfig{
		Level:      DefaultZapLevel,
		Format:     DefaultZapFormat,
		OutputPath: DefaultZapOutputPath,
	}

	err := env.BindPrefix("log.zap", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected log.zap.enabled to be true")
	}
	if cfg.Level != "debug" {
		t.Errorf("expected level 'debug', got '%s'", cfg.Level)
	}
	if cfg.Format != "console" {
		t.Errorf("expected format 'console', got '%s'", cfg.Format)
	}
	if cfg.OutputPath != "logs/app.log" {
		t.Errorf("expected output-path 'logs/app.log', got '%s'", cfg.OutputPath)
	}
}

func TestZapConfig_DefaultValues(t *testing.T) {
	cfg := &ZapConfig{
		Level:      DefaultZapLevel,
		Format:     DefaultZapFormat,
		OutputPath: DefaultZapOutputPath,
	}

	if cfg.Level != "info" {
		t.Errorf("expected default level 'info', got '%s'", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("expected default format 'json', got '%s'", cfg.Format)
	}
	if cfg.OutputPath != "stdout" {
		t.Errorf("expected default output-path 'stdout', got '%s'", cfg.OutputPath)
	}
}
