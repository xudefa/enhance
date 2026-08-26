package resilience

import (
	"testing"
)

func TestNewLeastConnections(t *testing.T) {
	t.Parallel()
	lc := NewLeastConnections()
	if lc == nil {
		t.Fatal("expected non-nil LeastConnections")
	}
}

func TestLeastConnections_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	lc := NewLeastConnections()
	_, err := lc.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestLeastConnections_Next_SingleBackend(t *testing.T) {
	t.Parallel()
	lc := NewLeastConnections()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1", Active: 5},
	}

	result, err := lc.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestLeastConnections_Next_SelectLeastActive(t *testing.T) {
	t.Parallel()
	lc := NewLeastConnections()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1", Active: 10},
		{URL: "http://backend2", ID: "2", Active: 5},
		{URL: "http://backend3", ID: "3", Active: 8},
	}

	result, err := lc.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend2" {
		t.Errorf("expected http://backend2 (least active), got %s", result.URL)
	}
}

func TestLeastConnections_Next_WithNilBackends(t *testing.T) {
	t.Parallel()
	lc := NewLeastConnections()
	backends := []*ServiceInstance{
		nil,
		{URL: "http://backend1", ID: "1", Active: 5},
		nil,
	}

	result, err := lc.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}
