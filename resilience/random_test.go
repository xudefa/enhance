package resilience

import (
	"testing"
)

func TestNewRandom(t *testing.T) {
	t.Parallel()
	r := NewRandom()
	if r == nil {
		t.Fatal("expected non-nil Random")
	}
}

func TestRandom_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	r := NewRandom()
	_, err := r.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestRandom_Next_SingleBackend(t *testing.T) {
	t.Parallel()
	r := NewRandom()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	result, err := r.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestRandom_Next_MultipleBackends(t *testing.T) {
	t.Parallel()
	r := NewRandom()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
		{URL: "http://backend3", ID: "3"},
	}

	selected := make(map[string]int)
	for i := 0; i < 100; i++ {
		result, err := r.Next(backends)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		selected[result.URL]++
	}

	if len(selected) == 0 {
		t.Error("expected at least one backend to be selected")
	}
}

func TestRandom_Next_WithNilBackends(t *testing.T) {
	t.Parallel()
	r := NewRandom()
	backends := []*ServiceInstance{
		nil,
		{URL: "http://backend1", ID: "1"},
		nil,
	}

	result, err := r.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}
