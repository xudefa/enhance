package metrics

import (
	"testing"
)

func TestCounterBuilder_Build(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()

	counter := NewMetricBuilder(registry, "requests").
		Tag("method", "GET").
		Tag("status", "200").
		BuildCounter()

	counter.Inc()
	if counter.Value() != 1 {
		t.Errorf("expected 1, got %f", counter.Value())
	}

	// 验证标签
	metrics := registry.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Tags["method"] != "GET" {
		t.Errorf("expected method=GET, got %s", metrics[0].Tags["method"])
	}
	if metrics[0].Tags["status"] != "200" {
		t.Errorf("expected status=200, got %s", metrics[0].Tags["status"])
	}
}

func TestGaugeBuilder_Build(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()

	gauge := NewMetricBuilder(registry, "memory").
		Tag("type", "heap").
		BuildGauge()

	gauge.Set(1024.5)
	if gauge.Value() != 1024.5 {
		t.Errorf("expected 1024.5, got %f", gauge.Value())
	}

	// 验证标签
	metrics := registry.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Tags["type"] != "heap" {
		t.Errorf("expected type=heap, got %s", metrics[0].Tags["type"])
	}
}

func TestHistogramBuilder_Build(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()

	hist := NewMetricBuilder(registry, "duration").
		Tag("endpoint", "/api").
		BuildHistogram()

	hist.Record(100.5)
	if hist.Count() != 1 {
		t.Errorf("expected count 1, got %d", hist.Count())
	}
	if hist.Sum() != 100.5 {
		t.Errorf("expected sum 100.5, got %f", hist.Sum())
	}

	// 验证标签
	metrics := registry.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Tags["endpoint"] != "/api" {
		t.Errorf("expected endpoint=/api, got %s", metrics[0].Tags["endpoint"])
	}
}

func TestCounterBuilder_Tags(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()

	counter := NewMetricBuilder(registry, "requests").
		Tags(map[string]string{
			"method": "POST",
			"status": "201",
		}).
		BuildCounter()

	counter.Inc()
	metrics := registry.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Tags["method"] != "POST" {
		t.Errorf("expected method=POST, got %s", metrics[0].Tags["method"])
	}
	if metrics[0].Tags["status"] != "201" {
		t.Errorf("expected status=201, got %s", metrics[0].Tags["status"])
	}
}

func TestCounterBuilder_NoTags(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()

	counter := NewMetricBuilder(registry, "simple_counter").
		BuildCounter()

	counter.Inc()
	if counter.Value() != 1 {
		t.Errorf("expected 1, got %f", counter.Value())
	}

	metrics := registry.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Name != "simple_counter" {
		t.Errorf("expected name=simple_counter, got %s", metrics[0].Name)
	}
}

func TestCounterBuilder_ChainMultipleTags(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()

	counter := NewMetricBuilder(registry, "http_requests").
		Tag("service", "api").
		Tag("method", "GET").
		Tag("status", "200").
		BuildCounter()

	counter.Inc()
	counter.Inc()
	counter.Inc()

	if counter.Value() != 3 {
		t.Errorf("expected 3, got %f", counter.Value())
	}

	metrics := registry.Collect()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Tags["service"] != "api" {
		t.Errorf("expected service=api, got %s", metrics[0].Tags["service"])
	}
}

func TestCounterBuilder_SameNameDifferentTags(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()

	counter1 := NewMetricBuilder(registry, "requests").
		Tag("method", "GET").
		BuildCounter()

	counter2 := NewMetricBuilder(registry, "requests").
		Tag("method", "POST").
		BuildCounter()

	counter1.Inc()
	counter2.Inc()
	counter2.Inc()

	metrics := registry.Collect()
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}

	// 验证两个计数器独立工作
	for _, m := range metrics {
		if m.Tags["method"] == "GET" && m.Value != 1 {
			t.Errorf("expected GET counter value 1, got %f", m.Value)
		}
		if m.Tags["method"] == "POST" && m.Value != 2 {
			t.Errorf("expected POST counter value 2, got %f", m.Value)
		}
	}
}
