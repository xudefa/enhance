package cron

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestCronConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-cron", environment.PriorityNormal, map[string]any{
		"cron.enabled":     "true",
		"cron.with-logger": "true",
	}))

	cfg := &CronConfig{
		WithLogger: DefaultWithLogger,
	}

	err := env.BindPrefix("cron", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected cron.enabled to be true")
	}
	if !cfg.WithLogger {
		t.Error("expected with-logger to be true")
	}
}

func TestCronConfig_DefaultValues(t *testing.T) {
	cfg := &CronConfig{
		WithLogger: DefaultWithLogger,
	}

	if cfg.WithLogger {
		t.Error("expected default with-logger to be false")
	}
}
