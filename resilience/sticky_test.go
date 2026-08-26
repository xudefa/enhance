package resilience

import (
	"testing"
)

func TestNewStickySession(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	if ss == nil {
		t.Fatal("expected non-nil StickySession")
	}
	if ss.sessionCookieName != "JSESSIONID" {
		t.Errorf("expected default cookie name JSESSIONID, got %s", ss.sessionCookieName)
	}
}

func TestNewStickySession_CustomCookieName(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("MYSESSION")
	if ss.sessionCookieName != "MYSESSION" {
		t.Errorf("expected cookie name MYSESSION, got %s", ss.sessionCookieName)
	}
}

func TestStickySession_Next_EmptyBackends(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	_, err := ss.Next(nil)
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestStickySession_Next(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	result, err := ss.Next(backends)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestStickySession_NextWithSession_EmptySessionID(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
	}

	result, err := ss.NextWithSession(backends, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend1" {
		t.Errorf("expected http://backend1, got %s", result.URL)
	}
}

func TestStickySession_NextWithSession_NewSession(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
	}

	result, err := ss.NextWithSession(backends, "session-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	backend, exists := ss.GetSessionBackend("session-123")
	if !exists {
		t.Error("expected session to be bound to a backend")
	}
	if backend.URL != result.URL {
		t.Errorf("expected backend URL %s, got %s", result.URL, backend.URL)
	}
}

func TestStickySession_NextWithSession_ExistingSession(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
		{URL: "http://backend2", ID: "2"},
	}

	result1, err := ss.NextWithSession(backends, "session-456")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	result2, err := ss.NextWithSession(backends, "session-456")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if result1.URL != result2.URL {
		t.Errorf("expected same backend for same session, got %s and %s", result1.URL, result2.URL)
	}
}

func TestStickySession_NextWithSession_BackendRemoved(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	backends1 := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	_, _ = ss.NextWithSession(backends1, "session-789")

	backends2 := []*ServiceInstance{
		{URL: "http://backend2", ID: "2"},
	}

	result, err := ss.NextWithSession(backends2, "session-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "http://backend2" {
		t.Errorf("expected http://backend2, got %s", result.URL)
	}
}

func TestStickySession_GetSessionBackend_NotFound(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	_, exists := ss.GetSessionBackend("nonexistent")
	if exists {
		t.Error("expected session to not exist")
	}
}

func TestStickySession_RemoveSession(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	_, _ = ss.NextWithSession(backends, "session-to-remove")

	_, exists := ss.GetSessionBackend("session-to-remove")
	if !exists {
		t.Fatal("expected session to exist before removal")
	}

	ss.RemoveSession("session-to-remove")

	_, exists = ss.GetSessionBackend("session-to-remove")
	if exists {
		t.Error("expected session to be removed")
	}
}

func TestStickySession_GetSessionCount(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	backends := []*ServiceInstance{
		{URL: "http://backend1", ID: "1"},
	}

	if ss.GetSessionCount() != 0 {
		t.Errorf("expected 0 sessions, got %d", ss.GetSessionCount())
	}

	_, _ = ss.NextWithSession(backends, "session-1")
	_, _ = ss.NextWithSession(backends, "session-2")

	if ss.GetSessionCount() != 2 {
		t.Errorf("expected 2 sessions, got %d", ss.GetSessionCount())
	}
}

func TestStickySession_NextWithSession_EmptyBackends(t *testing.T) {
	t.Parallel()
	ss := NewStickySession("")
	_, err := ss.NextWithSession(nil, "session-123")
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}
