package asynq

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestAsynqConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-asynq", environment.PriorityNormal, map[string]any{
		"asynq.enabled":          "true",
		"asynq.host":             "192.168.1.100",
		"asynq.port":             "6380",
		"asynq.password":         "secret",
		"asynq.db":               "1",
		"asynq.pool-size":        "20",
		"asynq.enable-scheduler": "true",
	}))

	cfg := &AsynqConfig{
		Host:            DefaultAsynqHost,
		Port:            DefaultAsynqPort,
		DB:              DefaultAsynqDB,
		PoolSize:        DefaultAsynqPoolSize,
		EnableScheduler: DefaultEnableScheduler,
	}

	err := env.BindPrefix("asynq", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected asynq.enabled to be true")
	}
	if cfg.Host != "192.168.1.100" {
		t.Errorf("expected host '192.168.1.100', got '%s'", cfg.Host)
	}
	if cfg.Port != 6380 {
		t.Errorf("expected port 6380, got %d", cfg.Port)
	}
	if !cfg.EnableScheduler {
		t.Error("expected enable-scheduler to be true")
	}
}

func TestAsynqConfig_DefaultValues(t *testing.T) {
	cfg := &AsynqConfig{
		Host:            DefaultAsynqHost,
		Port:            DefaultAsynqPort,
		DB:              DefaultAsynqDB,
		PoolSize:        DefaultAsynqPoolSize,
		EnableScheduler: DefaultEnableScheduler,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 6379 {
		t.Errorf("expected default port 6379, got %d", cfg.Port)
	}
	if cfg.DB != 0 {
		t.Errorf("expected default db 0, got %d", cfg.DB)
	}
	if cfg.PoolSize != 10 {
		t.Errorf("expected default pool-size 10, got %d", cfg.PoolSize)
	}
}
