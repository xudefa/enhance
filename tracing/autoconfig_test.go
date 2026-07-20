package tracing

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestTracingConfig_LoadConfig(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-tracing", environment.PriorityNormal, map[string]any{
		"tracing.enabled":       "true",
		"tracing.service_name":  "test-service",
		"tracing.sampling_rate": "0.5",
		"tracing.max_spans":     "5000",
	}))

	cfg := &TracingConfig{
		ServiceName:  DefaultServiceName,
		SamplingRate: DefaultSamplingRate,
		MaxSpans:     DefaultMaxSpans,
	}

	err := env.BindPrefix("tracing", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if cfg.ServiceName != "test-service" {
		t.Errorf("期望 ServiceName 'test-service'，实际 '%s'", cfg.ServiceName)
	}
	if cfg.SamplingRate != 0.5 {
		t.Errorf("期望 SamplingRate 0.5，实际 %f", cfg.SamplingRate)
	}
	if cfg.MaxSpans != 5000 {
		t.Errorf("期望 MaxSpans 5000，实际 %d", cfg.MaxSpans)
	}
}

func TestTracingConfig_DefaultValues(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()

	cfg := &TracingConfig{
		ServiceName:  DefaultServiceName,
		SamplingRate: DefaultSamplingRate,
		MaxSpans:     DefaultMaxSpans,
	}

	err := env.BindPrefix("tracing", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if cfg.ServiceName != DefaultServiceName {
		t.Errorf("期望 ServiceName '%s'，实际 '%s'", DefaultServiceName, cfg.ServiceName)
	}
	if cfg.SamplingRate != DefaultSamplingRate {
		t.Errorf("期望 SamplingRate %f，实际 %f", DefaultSamplingRate, cfg.SamplingRate)
	}
	if cfg.MaxSpans != DefaultMaxSpans {
		t.Errorf("期望 MaxSpans %d，实际 %d", DefaultMaxSpans, cfg.MaxSpans)
	}
}
