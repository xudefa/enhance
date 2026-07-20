package cobra

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestCobraConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-cobra", environment.PriorityNormal, map[string]any{
		"cobra.enabled": "true",
		"cobra.use":     "mycli",
		"cobra.short":   "My CLI App",
		"cobra.long":    "A powerful CLI application",
		"cobra.version": "2.0.0",
	}))

	cfg := &CobraConfig{
		Use:     DefaultCobraUse,
		Short:   DefaultCobraShort,
		Long:    DefaultCobraLong,
		Version: DefaultCobraVersion,
	}

	err := env.BindPrefix("cobra", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected cobra.enabled to be true")
	}
	if cfg.Use != "mycli" {
		t.Errorf("expected use 'mycli', got '%s'", cfg.Use)
	}
	if cfg.Short != "My CLI App" {
		t.Errorf("expected short 'My CLI App', got '%s'", cfg.Short)
	}
	if cfg.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got '%s'", cfg.Version)
	}
}

func TestCobraConfig_DefaultValues(t *testing.T) {
	cfg := &CobraConfig{
		Use:     DefaultCobraUse,
		Short:   DefaultCobraShort,
		Long:    DefaultCobraLong,
		Version: DefaultCobraVersion,
	}

	if cfg.Use != "app" {
		t.Errorf("expected default use 'app', got '%s'", cfg.Use)
	}
	if cfg.Short != "A CLI application" {
		t.Errorf("expected default short 'A CLI application', got '%s'", cfg.Short)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected default version '1.0.0', got '%s'", cfg.Version)
	}
}
