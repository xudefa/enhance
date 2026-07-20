package prometheus

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestPrometheusConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-prometheus", environment.PriorityNormal, map[string]any{
		"prometheus.enabled":             "true",
		"prometheus.host":                "127.0.0.1",
		"prometheus.port":                "9091",
		"prometheus.metrics_path":        "/metrics",
		"prometheus.enable_open_metrics": "true",
	}))

	cfg := &PrometheusConfig{
		Host:              DefaultPrometheusHost,
		Port:              DefaultPrometheusPort,
		MetricsPath:       DefaultMetricsPath,
		EnableOpenMetrics: DefaultEnableOpenMetrics,
	}

	err := env.BindPrefix("prometheus", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected prometheus.enabled to be true")
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.Port != 9091 {
		t.Errorf("expected port 9091, got %d", cfg.Port)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("expected metrics_path '/metrics', got '%s'", cfg.MetricsPath)
	}
	if !cfg.EnableOpenMetrics {
		t.Error("expected enable_open_metrics to be true")
	}
}

func TestPrometheusConfig_DefaultValues(t *testing.T) {
	cfg := &PrometheusConfig{
		Host:              DefaultPrometheusHost,
		Port:              DefaultPrometheusPort,
		MetricsPath:       DefaultMetricsPath,
		EnableOpenMetrics: DefaultEnableOpenMetrics,
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default host '0.0.0.0', got '%s'", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected default port 9090, got %d", cfg.Port)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("expected default metrics_path '/metrics', got '%s'", cfg.MetricsPath)
	}
	if cfg.EnableOpenMetrics {
		t.Error("expected default enable_open_metrics to be false")
	}
}
