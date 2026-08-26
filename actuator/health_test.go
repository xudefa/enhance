package actuator

import (
	"context"
	"runtime"
	"testing"

	"github.com/xudefa/enhance/actuator/health"
)

func TestNewDiskSpaceHealthIndicator(t *testing.T) {
	t.Parallel()
	ind := NewDiskSpaceHealthIndicator("/tmp", 0.9)
	if ind == nil {
		t.Fatal("NewDiskSpaceHealthIndicator returned nil")
	}
}

func TestDiskSpaceHealthIndicator_Name(t *testing.T) {
	t.Parallel()
	ind := NewDiskSpaceHealthIndicator("/tmp", 0.9)
	got := ind.Name()
	want := "disk_space_/tmp"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestDiskSpaceHealthIndicator_Health(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		path      string
		threshold float64
		wantUp    bool
	}{
		{"valid path low threshold", "/tmp", 0.01, false},
		{"valid path high threshold", "/tmp", 0.99, true},
		{"root path high threshold", "/", 0.99, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ind := NewDiskSpaceHealthIndicator(tt.path, tt.threshold)
			h := ind.Health(context.Background())

			if h.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
			if h.Details == nil {
				t.Error("Details should not be nil")
			}
			if _, ok := h.Details["path"]; !ok {
				t.Error("Details should contain path")
			}

			isUp := h.Status.String() == "UP"
			if isUp != tt.wantUp {
				t.Errorf("status = %s, wantUp = %v", h.Status, tt.wantUp)
			}
		})
	}
}

func TestDiskSpaceHealthIndicator_HealthInvalidPath(t *testing.T) {
	t.Parallel()
	ind := NewDiskSpaceHealthIndicator("/nonexistent/path/xyz", 0.9)
	h := ind.Health(context.Background())

	if h.Status != health.StatusUnknown {
		t.Errorf("expected UNKNOWN status for invalid path, got %s", h.Status)
	}
}

func TestNewMemoryHealthIndicator(t *testing.T) {
	t.Parallel()
	ind := NewMemoryHealthIndicator(0.9)
	if ind == nil {
		t.Fatal("NewMemoryHealthIndicator returned nil")
	}
}

func TestMemoryHealthIndicator_Name(t *testing.T) {
	t.Parallel()
	ind := NewMemoryHealthIndicator(0.9)
	if ind.Name() != "memory_usage" {
		t.Errorf("Name() = %s, want memory_usage", ind.Name())
	}
}

func TestMemoryHealthIndicator_Health(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		threshold float64
	}{
		{"very low threshold", 0.001},
		{"very high threshold", 0.99},
		{"mid threshold", 0.5},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ind := NewMemoryHealthIndicator(tt.threshold)
			h := ind.Health(context.Background())

			if h.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
			if h.Details == nil {
				t.Error("Details should not be nil")
			}
			if _, ok := h.Details["alloc_bytes"]; !ok {
				t.Error("should contain alloc_bytes")
			}
			if _, ok := h.Details["sys_bytes"]; !ok {
				t.Error("should contain sys_bytes")
			}
			if _, ok := h.Details["heap_percent"]; !ok {
				t.Error("should contain heap_percent")
			}
		})
	}
}

func TestNewProcessHealthIndicator(t *testing.T) {
	t.Parallel()
	ind := NewProcessHealthIndicator(1000)
	if ind == nil {
		t.Fatal("NewProcessHealthIndicator returned nil")
	}
}

func TestProcessHealthIndicator_Name(t *testing.T) {
	t.Parallel()
	ind := NewProcessHealthIndicator(1000)
	if ind.Name() != "process_status" {
		t.Errorf("Name() = %s, want process_status", ind.Name())
	}
}

func TestProcessHealthIndicator_Health(t *testing.T) {
	t.Parallel()
	current := runtime.NumGoroutine()

	tests := []struct {
		name      string
		threshold int
		wantUp    bool
	}{
		{"below threshold", current + 1000, true},
		{"above threshold", 1, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ind := NewProcessHealthIndicator(tt.threshold)
			h := ind.Health(context.Background())

			if h.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
			if _, ok := h.Details["goroutines"]; !ok {
				t.Error("should contain goroutines")
			}
			if _, ok := h.Details["cpu_num"]; !ok {
				t.Error("should contain cpu_num")
			}

			isUp := h.Status.String() == "UP"
			if isUp != tt.wantUp {
				t.Errorf("status = %s, wantUp = %v", h.Status, tt.wantUp)
			}
		})
	}
}
