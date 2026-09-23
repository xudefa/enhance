package metrics

import (
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	t.Parallel()
	c := NewCounter("test_counter")

	c.Inc()
	c.Add(2)

	if c.Value() != 3 {
		t.Errorf("expected 3, got %f", c.Value())
	}

	c.Reset()
	if c.Value() != 0 {
		t.Errorf("expected 0 after reset, got %f", c.Value())
	}
}

func TestGauge(t *testing.T) {
	t.Parallel()
	gauge := NewGauge("test_gauge")

	gauge.Set(42.5)
	if gauge.Value() != 42.5 {
		t.Errorf("expected 42.5, got %f", gauge.Value())
	}

	gauge.Set(10.0)
	if gauge.Value() != 10.0 {
		t.Errorf("expected 10.0, got %f", gauge.Value())
	}
}

func TestHistogram(t *testing.T) {
	t.Parallel()
	histogram := NewHistogram("test_histogram")

	histogram.Observe(10.0)
	histogram.Observe(20.0)
	histogram.Observe(30.0)

	if histogram.Count() != 3 {
		t.Errorf("expected count 3, got %d", histogram.Count())
	}
	if histogram.Sum() != 60.0 {
		t.Errorf("expected sum 60.0, got %f", histogram.Sum())
	}
	if histogram.Min() != 10.0 {
		t.Errorf("expected min 10.0, got %f", histogram.Min())
	}
	if histogram.Max() != 30.0 {
		t.Errorf("expected max 30.0, got %f", histogram.Max())
	}
	if histogram.Value() != 20.0 {
		t.Errorf("expected avg 20.0, got %f", histogram.Value())
	}
}

func TestTimer(t *testing.T) {
	t.Parallel()
	timer := NewTimer("test_timer")
	timer.Start()
	time.Sleep(10 * time.Millisecond)
	timer.Stop()

	if timer.Duration() == 0 {
		t.Error("expected non-zero duration")
	}
	if timer.Value() == 0 {
		t.Error("expected non-zero value")
	}
}

func TestMetricsRegistry(t *testing.T) {
	t.Parallel()
	registry := NewMetricsRegistry()

	counter := NewCounter("reg_counter")
	registry.Register(counter)

	metric, exists := registry.Get("reg_counter")
	if !exists {
		t.Fatal("expected metric to exist")
	}
	if metric.Name() != "reg_counter" {
		t.Errorf("expected name 'reg_counter', got %s", metric.Name())
	}

	metrics := registry.List()
	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}
}
