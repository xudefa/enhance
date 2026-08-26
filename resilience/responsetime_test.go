package resilience

import (
	"testing"
)

func TestNewResponseTimeWeighted(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted()
	if rtw == nil {
		t.Fatal("expected non-nil ResponseTimeWeighted")
	}
	if rtw.decay != 0.9 {
		t.Errorf("expected decay 0.9, got %f", rtw.decay)
	}
}

func TestNewResponseTimeWeighted_CustomDecay(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted(0.8)
	if rtw.decay != 0.8 {
		t.Errorf("expected decay 0.8, got %f", rtw.decay)
	}
}

func TestResponseTimeWeighted_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted()
	_, err := rtw.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestResponseTimeWeighted_Next_SingleBackend(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	result, err := rtw.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestResponseTimeWeighted_Next_MultipleBackends(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
	}

	result, err := rtw.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestResponseTimeWeighted_RecordResponseTime(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted()
	rtw.RecordResponseTime("http://backend1", 100.0)
	rtw.RecordResponseTime("http://backend1", 200.0)

	avgTime, ok := rtw.GetAvgResponseTime("http://backend1")
	if !ok {
		t.Fatal("expected avg response time to exist")
	}
	if avgTime <= 0 {
		t.Errorf("expected positive avg response time, got %f", avgTime)
	}
}

func TestResponseTimeWeighted_GetAvgResponseTime_NotFound(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted()
	_, ok := rtw.GetAvgResponseTime("http://nonexistent")
	if ok {
		t.Error("expected avg response time to not exist")
	}
}

func TestResponseTimeWeighted_Next_WithRecordedTimes(t *testing.T) {
	t.Parallel()
	rtw := NewResponseTimeWeighted()
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
	}

	rtw.RecordResponseTime("http://backend1", 100.0)
	rtw.RecordResponseTime("http://backend2", 50.0)

	selected := make(map[string]int)
	for i := 0; i < 100; i++ {
		result, err := rtw.Next(backends)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		selected[result.URL]++
	}

	if len(selected) == 0 {
		t.Error("expected at least one backend to be selected")
	}
}
