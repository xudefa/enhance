package viper

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestViperConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-viper", environment.PriorityNormal, map[string]any{
		"viper.enabled":       "true",
		"viper.config-name":   "custom-config",
		"viper.config-type":   "json",
		"viper.config-path":   "/etc/myapp",
		"viper.watch-changes": "true",
	}))

	cfg := &ViperConfig{
		ConfigName:   DefaultConfigName,
		ConfigType:   DefaultConfigType,
		ConfigPath:   DefaultConfigPath,
		WatchChanges: DefaultWatchChanges,
	}

	err := env.BindPrefix("viper", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected viper.enabled to be true")
	}
	if cfg.ConfigName != "custom-config" {
		t.Errorf("expected config-name 'custom-config', got '%s'", cfg.ConfigName)
	}
	if cfg.ConfigType != "json" {
		t.Errorf("expected config-type 'json', got '%s'", cfg.ConfigType)
	}
	if cfg.ConfigPath != "/etc/myapp" {
		t.Errorf("expected config-path '/etc/myapp', got '%s'", cfg.ConfigPath)
	}
	if !cfg.WatchChanges {
		t.Error("expected watch-changes to be true")
	}
}

func TestViperConfig_DefaultValues(t *testing.T) {
	cfg := &ViperConfig{
		ConfigName:   DefaultConfigName,
		ConfigType:   DefaultConfigType,
		ConfigPath:   DefaultConfigPath,
		WatchChanges: DefaultWatchChanges,
	}

	if cfg.ConfigName != "application" {
		t.Errorf("expected default config-name 'application', got '%s'", cfg.ConfigName)
	}
	if cfg.ConfigType != "yaml" {
		t.Errorf("expected default config-type 'yaml', got '%s'", cfg.ConfigType)
	}
	if cfg.ConfigPath != "." {
		t.Errorf("expected default config-path '.', got '%s'", cfg.ConfigPath)
	}
	if cfg.WatchChanges {
		t.Error("expected default watch-changes to be false")
	}
}
