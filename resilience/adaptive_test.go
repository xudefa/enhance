package resilience

import (
	"testing"
)

func TestNewAdaptiveWeight(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	if aw == nil {
		t.Fatal("expected non-nil AdaptiveWeight")
	}
}

func TestAdaptiveWeight_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	_, err := aw.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestAdaptiveWeight_Next_SingleBackend(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}
	result, err := aw.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestAdaptiveWeight_Next_MultipleBackends(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
		{URL: "http://backend3", ID: "3"},
	}

	selected := make(map[string]int)
	for i := 0; i < 100; i++ {
		result, err := aw.Next(backends)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		selected[result.URL]++
	}

	if len(selected) == 0 {
		t.Error("expected at least one backend to be selected")
	}
}

func TestAdaptiveWeight_RecordRequest(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	aw.RecordRequest("http://backend1", 100.0, false)
	aw.RecordRequest("http://backend1", 200.0, true)

	stats, ok := aw.GetStats("http://backend1")
	if !ok {
		t.Fatal("expected stats to exist")
	}
	if stats.TotalRequests.Load() != 2 {
		t.Errorf("expected 2 total requests, got %d", stats.TotalRequests.Load())
	}
	if stats.FailedRequests.Load() != 1 {
		t.Errorf("expected 1 failed request, got %d", stats.FailedRequests.Load())
	}
}

func TestAdaptiveWeight_RecordConnection(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	aw.RecordConnection("http://backend1", 5)

	stats, ok := aw.GetStats("http://backend1")
	if !ok {
		t.Fatal("expected stats to exist")
	}
	if stats.ActiveConnections.Load() != 5 {
		t.Errorf("expected 5 active connections, got %d", stats.ActiveConnections.Load())
	}

	aw.RecordConnection("http://backend1", -3)
	if stats.ActiveConnections.Load() != 2 {
		t.Errorf("expected 2 active connections, got %d", stats.ActiveConnections.Load())
	}
}

func TestAdaptiveWeight_RecordConnection_NegativeClamp(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	aw.RecordConnection("http://backend1", -10)

	stats, ok := aw.GetStats("http://backend1")
	if !ok {
		t.Fatal("expected stats to exist")
	}
	if stats.ActiveConnections.Load() != 0 {
		t.Errorf("expected 0 active connections (clamped), got %d", stats.ActiveConnections.Load())
	}
}

func TestAdaptiveWeight_GetStats_NotFound(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	_, ok := aw.GetStats("http://nonexistent")
	if ok {
		t.Error("expected stats to not exist")
	}
}

func TestAdaptiveWeight_GetAllStats(t *testing.T) {
	t.Parallel()
	aw := NewAdaptiveWeight()
	aw.RecordRequest("http://backend1", 100.0, false)
	aw.RecordRequest("http://backend2", 200.0, false)

	allStats := aw.GetAllStats()
	if len(allStats) != 2 {
		t.Errorf("expected 2 stats entries, got %d", len(allStats))
	}
}

func TestBackendStats_UpdateAvgResponseTime(t *testing.T) {
	t.Parallel()
	stats := &BackendStats{}
	stats.UpdateAvgResponseTime(100.0)
	if stats.GetAvgResponseTime() != 100.0 {
		t.Errorf("expected 100.0, got %f", stats.GetAvgResponseTime())
	}

	stats.UpdateAvgResponseTime(200.0)
	expected := 0.9*100.0 + 0.1*200.0
	if stats.GetAvgResponseTime() != expected {
		t.Errorf("expected %f, got %f", expected, stats.GetAvgResponseTime())
	}
}

func TestSortByResponseTime(t *testing.T) {
	t.Parallel()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
		{URL: "http://backend3", ID: "3"},
	}

	stats := map[string]*BackendStats{
		"http://backend1": {avgResponseTime: 300.0},
		"http://backend2": {avgResponseTime: 100.0},
		"http://backend3": {avgResponseTime: 200.0},
	}

	sorted := SortByResponseTime(backends, stats)
	if sorted[0].URL != "http://backend2" {
		t.Errorf("expected first backend to be http://backend2, got %s", sorted[0].URL)
	}
	if sorted[1].URL != "http://backend3" {
		t.Errorf("expected second backend to be http://backend3, got %s", sorted[1].URL)
	}
	if sorted[2].URL != "http://backend1" {
		t.Errorf("expected third backend to be http://backend1, got %s", sorted[2].URL)
	}
}

func TestSortByResponseTime_WithNilBackends(t *testing.T) {
	t.Parallel()
	backends := []*ServiceInstance{
		nil,
		{URL: "http://backend1", ID: "1"},
		nil,
	}

	stats := map[string]*BackendStats{}
	sorted := SortByResponseTime(backends, stats)

	if len(sorted) != 3 {
		t.Errorf("expected 3 backends, got %d", len(sorted))
	}
}

func TestSortByResponseTime_NoStats(t *testing.T) {
	t.Parallel()
	backends := []*ServiceInstance{
		{URL: "http://backend2", ID: "2"},
		{URL: "http://backend1", ID: "1"},
	}

	stats := map[string]*BackendStats{}
	sorted := SortByResponseTime(backends, stats)

	if sorted[0].URL != "http://backend1" {
		t.Errorf("expected first backend to be http://backend1 (sorted by URL), got %s", sorted[0].URL)
	}
}
