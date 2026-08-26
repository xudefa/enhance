package resilience

import (
	"testing"
)

func TestNewHealthAware(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha, err := NewHealthAware(inner)
	if err != nil {
		t.Fatalf("NewHealthAware failed: %v", err)
	}
	if ha == nil {
		t.Fatal("expected non-nil HealthAware")
	}
}

func TestNewHealthAware_NilInner(t *testing.T) {
	t.Parallel()
	_, err := NewHealthAware(nil)
	if err == nil {
		t.Error("expected error for nil inner balancer")
	}
}

func TestMustNewHealthAware(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha := MustNewHealthAware(inner)
	if ha == nil {
		t.Fatal("expected non-nil HealthAware")
	}
}

func TestMustNewHealthAware_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil inner balancer")
		}
	}()
	MustNewHealthAware(nil)
}

func TestHealthAware_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha := MustNewHealthAware(inner)

	_, err := ha.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestHealthAware_Next_HealthyBackends(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha := MustNewHealthAware(inner)

	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1", Health: HealthUp},
		{URL: "http://backend2", ID: "2", Health: HealthDown},
	}

	result, err := ha.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestHealthAware_Next_AllDown(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha := MustNewHealthAware(inner)

	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1", Health: HealthDown},
		{URL: "http://backend2", ID: "2", Health: HealthDown},
	}

	result, err := ha.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestHealthAware_RecordFailure(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha := MustNewHealthAware(inner)

	ha.RecordFailure("http://backend1")
	ha.RecordFailure("http://backend1")

	count := ha.GetFailureCount("http://backend1")
	if count != 2 {
		t.Errorf("expected 2 failures, got %d", count)
	}
}

func TestHealthAware_RecordSuccess(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha := MustNewHealthAware(inner)

	ha.RecordFailure("http://backend1")
	ha.RecordSuccess("http://backend1")

	count := ha.GetFailureCount("http://backend1")
	if count != 0 {
		t.Errorf("expected 0 failures after success, got %d", count)
	}
}

func TestHealthAware_GetFailureCount_NotFound(t *testing.T) {
	t.Parallel()
	inner := NewRoundRobin()
	ha := MustNewHealthAware(inner)

	count := ha.GetFailureCount("http://nonexistent")
	if count != 0 {
		t.Errorf("expected 0 failures, got %d", count)
	}
}
