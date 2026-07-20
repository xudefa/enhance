package otel

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestOtelConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-otel", environment.PriorityNormal, map[string]any{
		"otel.enabled":       "true",
		"otel.endpoint":      "localhost:4318",
		"otel.service_name":  "test-service",
		"otel.sampling_rate": "0.5",
	}))

	cfg := &OtelConfig{
		Endpoint:       DefaultOtelEndpoint,
		ServiceName:    DefaultServiceName,
		ServiceVersion: DefaultServiceVersion,
		SamplingRate:   DefaultSamplingRate,
	}

	err := env.BindPrefix("otel", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected otel.enabled to be true")
	}
	if cfg.Endpoint != "localhost:4318" {
		t.Errorf("expected endpoint 'localhost:4318', got '%s'", cfg.Endpoint)
	}
	if cfg.ServiceName != "test-service" {
		t.Errorf("expected service_name 'test-service', got '%s'", cfg.ServiceName)
	}
}

func TestOtelConfig_DefaultValues(t *testing.T) {
	cfg := &OtelConfig{
		Endpoint:       DefaultOtelEndpoint,
		ServiceName:    DefaultServiceName,
		ServiceVersion: DefaultServiceVersion,
		SamplingRate:   DefaultSamplingRate,
	}

	if cfg.Endpoint != "localhost:4317" {
		t.Errorf("expected default endpoint 'localhost:4317', got '%s'", cfg.Endpoint)
	}
	if cfg.SamplingRate != 1.0 {
		t.Errorf("expected default sampling_rate 1.0, got %f", cfg.SamplingRate)
	}
}
