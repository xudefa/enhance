package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_NewApplicationInstance(t *testing.T) {
	t.Parallel()
	inst := NewApplicationInstance("my-app", "http://localhost:8080")
	if inst == nil {
		t.Fatal("NewApplicationInstance returned nil")
	}
	if inst.ApplicationID != "my-app" {
		t.Errorf("ApplicationID = %s, want my-app", inst.ApplicationID)
	}
	if inst.URL != "http://localhost:8080" {
		t.Errorf("URL = %s, want http://localhost:8080", inst.URL)
	}
	if inst.Status != StatusUnknown {
		t.Errorf("Status = %s, want %s", inst.Status, StatusUnknown)
	}
	if inst.Metrics == nil {
		t.Error("Metrics should be initialized")
	}
	if inst.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
	if inst.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestHandlers_NewHealthInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status ApplicationStatus
	}{
		{"up", StatusUp},
		{"down", StatusDown},
		{"unknown", StatusUnknown},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := NewHealthInfo(tt.status)
			if h == nil {
				t.Fatal("NewHealthInfo returned nil")
			}
			if h.Status != tt.status {
				t.Errorf("Status = %s, want %s", h.Status, tt.status)
			}
			if h.Components == nil {
				t.Error("Components should be initialized")
			}
			if h.Details == nil {
				t.Error("Details should be initialized")
			}
		})
	}
}

func TestHandlers_AdminServer_HandlerReturnsMux(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)
	handler := server.Handler()
	if handler == nil {
		t.Error("Handler() returned nil")
	}
}

func TestHandlers_GetApplication_NotFound(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/applications/nonexistent", nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestHandlers_GetHealth_NoHealth(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	inst := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(inst)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/"+inst.ID+"/health?id="+inst.ID, nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestHandlers_GetMetrics_NoMetrics(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	inst := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(inst)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/"+inst.ID+"/metrics?id="+inst.ID, nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestHandlers_CountsEmpty(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	if registry.GetApplicationCount() != 0 {
		t.Errorf("app count = %d, want 0", registry.GetApplicationCount())
	}
	if registry.GetInstanceCount() != 0 {
		t.Errorf("instance count = %d, want 0", registry.GetInstanceCount())
	}
	if registry.GetUpInstanceCount() != 0 {
		t.Errorf("up count = %d, want 0", registry.GetUpInstanceCount())
	}
	if registry.GetDownInstanceCount() != 0 {
		t.Errorf("down count = %d, want 0", registry.GetDownInstanceCount())
	}
}
