package resilience

import (
	"testing"
)

func TestNewRoundRobin(t *testing.T) {
	t.Parallel()
	rr := NewRoundRobin()
	if rr == nil {
		t.Fatal("expected non-nil RoundRobin")
	}
}

func TestRoundRobin_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	rr := NewRoundRobin()
	_, err := rr.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestRoundRobin_Next_SingleBackend(t *testing.T) {
	t.Parallel()
	rr := NewRoundRobin()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	result, err := rr.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestRoundRobin_Next_RoundRobin(t *testing.T) {
	t.Parallel()
	rr := NewRoundRobin()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
		{URL: "http://backend3", ID: "3"},
	}

	expected := []string{"http://backend1", "http://backend2", "http://backend3", "http://backend1"}
	for i, exp := range expected {
		result, err := rr.Next(backends)
		if err != nil {
			t.Fatalf("call %d failed: %v", i+1, err)
		}
		if result.URL != exp {
			t.Errorf("call %d: expected %s, got %s", i+1, exp, result.URL)
		}
	}
}

func TestRoundRobin_Next_WithNilBackends(t *testing.T) {
	t.Parallel()
	rr := NewRoundRobin()
	backends := []*ServiceInstance{
		nil,
		{URL: "http://backend1", ID: "1"},
		nil,
	}

	result, err := rr.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestNewWeightedRoundRobin(t *testing.T) {
	t.Parallel()
	wrr := NewWeightedRoundRobin()
	if wrr == nil {
		t.Fatal("expected non-nil WeightedRoundRobin")
	}
}

func TestWeightedRoundRobin_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	wrr := NewWeightedRoundRobin()
	_, err := wrr.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestWeightedRoundRobin_Next_Weighted(t *testing.T) {
	t.Parallel()
	wrr := NewWeightedRoundRobin()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1", Weight: 3},
		{URL: "http://backend2", ID: "2", Weight: 1},
	}

	counts := make(map[string]int)
	for i := 0; i < 4; i++ {
		result, err := wrr.Next(backends)
		if err != nil {
			t.Fatalf("call %d failed: %v", i+1, err)
		}
		counts[result.URL]++
	}

	if counts["http://backend1"] != 3 {
		t.Errorf("expected backend1 to be selected 3 times, got %d", counts["http://backend1"])
	}
	if counts["http://backend2"] != 1 {
		t.Errorf("expected backend2 to be selected 1 time, got %d", counts["http://backend2"])
	}
}

func TestWeightedRoundRobin_Next_ZeroWeight(t *testing.T) {
	t.Parallel()
	wrr := NewWeightedRoundRobin()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1", Weight: 0},
		{URL: "http://backend2", ID: "2", Weight: 0},
	}

	result, err := wrr.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestNonNilBackends(t *testing.T) {
	t.Parallel()
	backends := []*ServiceInstance{
		nil,
		{URL: "http://backend1", ID: "1"},
		nil,
		{URL: "http://backend2", ID: "2"},
		nil,
	}

	result := nonNilBackends(backends)
	if len(result) != 2 {
		t.Errorf("expected 2 non-nil backends, got %d", len(result))
	}
	if result[0].URL != "http://backend1" {
		t.Errorf("expected first backend to be http://backend1, got %s", result[0].URL)
	}
	if result[1].URL != "http://backend2" {
		t.Errorf("expected second backend to be http://backend2, got %s", result[1].URL)
	}
}

func TestNonNilBackends_AllNil(t *testing.T) {
	t.Parallel()
	backends := []*ServiceInstance{nil, nil, nil}
	result := nonNilBackends(backends)
	if len(result) != 0 {
		t.Errorf("expected 0 non-nil backends, got %d", len(result))
	}
}

func TestNonNilBackends_Empty(t *testing.T) {
	t.Parallel()
	backends := []*ServiceInstance{}
	result := nonNilBackends(backends)
	if len(result) != 0 {
		t.Errorf("expected 0 non-nil backends, got %d", len(result))
	}
}
