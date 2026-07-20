package elasticsearch

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestElasticsearchConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-elasticsearch", environment.PriorityNormal, map[string]any{
		"elasticsearch.enabled":        "true",
		"elasticsearch.addresses":      "localhost:9200",
		"elasticsearch.username":       "elastic",
		"elasticsearch.password":       "changeme",
		"elasticsearch.timeout":        "30",
		"elasticsearch.max-idle-conns": "20",
	}))

	cfg := &ElasticsearchConfig{
		Timeout:      DefaultTimeout,
		MaxIdleConns: DefaultMaxIdleConns,
	}

	err := env.BindPrefix("elasticsearch", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected elasticsearch.enabled to be true")
	}
	if len(cfg.Addresses) == 0 {
		t.Error("expected addresses to be set")
	}
	if cfg.Username != "elastic" {
		t.Errorf("expected username 'elastic', got '%s'", cfg.Username)
	}
	if cfg.Timeout != 30 {
		t.Errorf("expected timeout 30, got %d", cfg.Timeout)
	}
}

func TestElasticsearchConfig_DefaultValues(t *testing.T) {
	cfg := &ElasticsearchConfig{
		Timeout:      DefaultTimeout,
		MaxIdleConns: DefaultMaxIdleConns,
	}

	if cfg.Timeout != 10 {
		t.Errorf("expected default timeout 10, got %d", cfg.Timeout)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("expected default max-idle-conns 10, got %d", cfg.MaxIdleConns)
	}
}
