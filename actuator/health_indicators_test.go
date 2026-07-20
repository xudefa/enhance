package actuator

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/actuator/health"
)

func TestDiskSpaceHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("Name", func(t *testing.T) {
		indicator := NewDiskSpaceHealthIndicator("/tmp", 0.9)
		expected := "disk_space_/tmp"
		if indicator.Name() != expected {
			t.Errorf("expected name %s, got %s", expected, indicator.Name())
		}
	})

	t.Run("Health_BelowThreshold", func(t *testing.T) {
		indicator := NewDiskSpaceHealthIndicator("/tmp", 0.9)
		h := indicator.Health(context.Background())

		if h.Status != health.StatusUp {
			t.Errorf("expected status UP, got %s", h.Status)
		}

		if h.Details["path"] != "/tmp" {
			t.Errorf("expected path /tmp, got %v", h.Details["path"])
		}

		if _, ok := h.Details["total_bytes"]; !ok {
			t.Error("expected total_bytes in details")
		}
		if _, ok := h.Details["used_bytes"]; !ok {
			t.Error("expected used_bytes in details")
		}
		if _, ok := h.Details["free_bytes"]; !ok {
			t.Error("expected free_bytes in details")
		}
		if _, ok := h.Details["usage_percent"]; !ok {
			t.Error("expected usage_percent in details")
		}
	})

	t.Run("Health_AboveThreshold", func(t *testing.T) {
		indicator := NewDiskSpaceHealthIndicator("/tmp", 0.3)
		h := indicator.Health(context.Background())

		if h.Status != health.StatusDegraded {
			t.Errorf("expected status Degraded, got %s", h.Status)
		}

		if _, ok := h.Details["message"]; !ok {
			t.Error("expected message in details when degraded")
		}
	})
}

func TestMemoryHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("Name", func(t *testing.T) {
		indicator := NewMemoryHealthIndicator(0.9)
		if indicator.Name() != "memory_usage" {
			t.Errorf("expected name 'memory_usage', got %s", indicator.Name())
		}
	})

	t.Run("Health", func(t *testing.T) {
		indicator := NewMemoryHealthIndicator(0.99)
		h := indicator.Health(context.Background())

		if h.Timestamp.IsZero() {
			t.Error("expected non-zero timestamp")
		}

		if _, ok := h.Details["alloc_bytes"]; !ok {
			t.Error("expected alloc_bytes in details")
		}
		if _, ok := h.Details["sys_bytes"]; !ok {
			t.Error("expected sys_bytes in details")
		}
		if _, ok := h.Details["heap_alloc"]; !ok {
			t.Error("expected heap_alloc in details")
		}
		if _, ok := h.Details["heap_sys"]; !ok {
			t.Error("expected heap_sys in details")
		}
		if _, ok := h.Details["heap_objects"]; !ok {
			t.Error("expected heap_objects in details")
		}
		if _, ok := h.Details["heap_percent"]; !ok {
			t.Error("expected heap_percent in details")
		}
	})

	t.Run("Health_Degraded", func(t *testing.T) {
		indicator := NewMemoryHealthIndicator(0.0000001)
		h := indicator.Health(context.Background())

		if h.Status != health.StatusDegraded {
			t.Errorf("expected status Degraded with very low threshold, got %s", h.Status)
		}
	})
}

func TestProcessHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("Name", func(t *testing.T) {
		indicator := NewProcessHealthIndicator(1000)
		if indicator.Name() != "process_status" {
			t.Errorf("expected name 'process_status', got %s", indicator.Name())
		}
	})

	t.Run("Health_Normal", func(t *testing.T) {
		indicator := NewProcessHealthIndicator(10000)
		h := indicator.Health(context.Background())

		if h.Status != health.StatusUp {
			t.Errorf("expected status UP, got %s", h.Status)
		}

		if _, ok := h.Details["goroutines"]; !ok {
			t.Error("expected goroutines in details")
		}
		if _, ok := h.Details["cpu_num"]; !ok {
			t.Error("expected cpu_num in details")
		}
	})

	t.Run("Health_Degraded", func(t *testing.T) {
		indicator := NewProcessHealthIndicator(1)
		h := indicator.Health(context.Background())

		if h.Status != health.StatusDegraded {
			t.Errorf("expected status Degraded, got %s", h.Status)
		}

		if _, ok := h.Details["message"]; !ok {
			t.Error("expected message in details when degraded")
		}
	})
}
