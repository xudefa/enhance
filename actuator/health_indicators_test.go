package actuator

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/actuator/health"
)

func TestDiskSpaceHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("Name", func(t *testing.T) {
		t.Parallel()
		testDiskSpaceHealthIndicatorName(t)
	})

	t.Run("Health_BelowThreshold", func(t *testing.T) {
		t.Parallel()
		testDiskSpaceHealthIndicatorBelowThreshold(t)
	})

	t.Run("Health_AboveThreshold", func(t *testing.T) {
		t.Parallel()
		testDiskSpaceHealthIndicatorAboveThreshold(t)
	})
}

func testDiskSpaceHealthIndicatorName(t *testing.T) {
	t.Helper()
	indicator := NewDiskSpaceHealthIndicator("/tmp", 0.9)
	expected := "disk_space_/tmp"
	if indicator.Name() != expected {
		t.Errorf("expected name %s, got %s", expected, indicator.Name())
	}
}

func testDiskSpaceHealthIndicatorBelowThreshold(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	indicator := NewDiskSpaceHealthIndicator(dir, 1.0)
	healthResult := indicator.Health(context.Background())

	if healthResult.Status != health.StatusUp {
		t.Errorf("expected status UP, got %s", healthResult.Status)
	}

	if healthResult.Details["path"] != dir {
		t.Errorf("expected path %s, got %v", dir, healthResult.Details["path"])
	}

	if _, ok := healthResult.Details["total_bytes"]; !ok {
		t.Error("expected total_bytes in details")
	}
	if _, ok := healthResult.Details["used_bytes"]; !ok {
		t.Error("expected used_bytes in details")
	}
	if _, ok := healthResult.Details["free_bytes"]; !ok {
		t.Error("expected free_bytes in details")
	}
	if _, ok := healthResult.Details["usage_percent"]; !ok {
		t.Error("expected usage_percent in details")
	}
}

func testDiskSpaceHealthIndicatorAboveThreshold(t *testing.T) {
	t.Helper()
	indicator := NewDiskSpaceHealthIndicator(t.TempDir(), 0.0)
	healthResult := indicator.Health(context.Background())

	if healthResult.Status != health.StatusDegraded {
		t.Errorf("expected status Degraded, got %s", healthResult.Status)
	}

	if _, ok := healthResult.Details["message"]; !ok {
		t.Error("expected message in details when degraded")
	}
}

func TestMemoryHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("Name", func(t *testing.T) {
		t.Parallel()
		indicator := NewMemoryHealthIndicator(0.9)
		if indicator.Name() != "memory_usage" {
			t.Errorf("expected name 'memory_usage', got %s", indicator.Name())
		}
	})

	t.Run("Health", func(t *testing.T) {
		t.Parallel()
		indicator := NewMemoryHealthIndicator(0.99)
		healthResult := indicator.Health(context.Background())

		if healthResult.Timestamp.IsZero() {
			t.Error("expected non-zero timestamp")
		}

		if _, ok := healthResult.Details["alloc_bytes"]; !ok {
			t.Error("expected alloc_bytes in details")
		}
		if _, ok := healthResult.Details["sys_bytes"]; !ok {
			t.Error("expected sys_bytes in details")
		}
		if _, ok := healthResult.Details["heap_alloc"]; !ok {
			t.Error("expected heap_alloc in details")
		}
		if _, ok := healthResult.Details["heap_sys"]; !ok {
			t.Error("expected heap_sys in details")
		}
		if _, ok := healthResult.Details["heap_objects"]; !ok {
			t.Error("expected heap_objects in details")
		}
		if _, ok := healthResult.Details["heap_percent"]; !ok {
			t.Error("expected heap_percent in details")
		}
	})

	t.Run("Health_Degraded", func(t *testing.T) {
		t.Parallel()
		indicator := NewMemoryHealthIndicator(0.0000001)
		healthResult := indicator.Health(context.Background())
		if healthResult.Status != health.StatusDegraded {
			t.Errorf("expected status Degraded with very low threshold, got %s", healthResult.Status)
		}
	})
}

func TestProcessHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("Name", func(t *testing.T) {
		t.Parallel()
		indicator := NewProcessHealthIndicator(1000)
		if indicator.Name() != "process_status" {
			t.Errorf("expected name 'process_status', got %s", indicator.Name())
		}
	})

	t.Run("Health_Normal", func(t *testing.T) {
		t.Parallel()
		indicator := NewProcessHealthIndicator(10000)
		healthResult := indicator.Health(context.Background())

		if healthResult.Status != health.StatusUp {
			t.Errorf("expected status UP, got %s", healthResult.Status)
		}

		if _, ok := healthResult.Details["goroutines"]; !ok {
			t.Error("expected goroutines in details")
		}
		if _, ok := healthResult.Details["cpu_num"]; !ok {
			t.Error("expected cpu_num in details")
		}
	})

	t.Run("Health_Degraded", func(t *testing.T) {
		t.Parallel()
		indicator := NewProcessHealthIndicator(1)
		healthResult := indicator.Health(context.Background())

		if healthResult.Status != health.StatusDegraded {
			t.Errorf("expected status Degraded, got %s", healthResult.Status)
		}

		if _, ok := healthResult.Details["message"]; !ok {
			t.Error("expected message in details when degraded")
		}
	})
}
