package redis

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestRedisConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-redis", environment.PriorityNormal, map[string]any{
		"redis.enabled": "true",
		"redis.host":    "localhost",
		"redis.port":    6379,
		"redis.db":      2,
		"redis.prefix":  "test:",
	}))

	cfg := &RedisConfig{
		Host:     DefaultRedisHost,
		Port:     DefaultRedisPort,
		DB:       DefaultRedisDB,
		Prefix:   DefaultRedisPrefix,
		PoolSize: DefaultRedisPoolSize,
	}

	err := env.BindPrefix("redis", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected redis.enabled to be true")
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 6379 {
		t.Errorf("expected port 6379, got %d", cfg.Port)
	}
	if cfg.DB != 2 {
		t.Errorf("expected db 2, got %d", cfg.DB)
	}
	if cfg.Prefix != "test:" {
		t.Errorf("expected prefix 'test:', got '%s'", cfg.Prefix)
	}
}

func TestRedisConfig_DefaultValues(t *testing.T) {
	cfg := &RedisConfig{
		Host:     DefaultRedisHost,
		Port:     DefaultRedisPort,
		DB:       DefaultRedisDB,
		Prefix:   DefaultRedisPrefix,
		PoolSize: DefaultRedisPoolSize,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 6379 {
		t.Errorf("expected default port 6379, got %d", cfg.Port)
	}
	if cfg.Prefix != "enhance:" {
		t.Errorf("expected default prefix 'enhance:', got '%s'", cfg.Prefix)
	}
}
