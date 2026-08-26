package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminServer_GetApplication_ByID(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/applications/my-app", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var app Application
	if err := json.NewDecoder(rr.Body).Decode(&app); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if app.ID != "my-app" {
		t.Errorf("expected app ID 'my-app', got %s", app.ID)
	}
}

func TestAdminServer_GetApplication_ByQueryParam(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/applications?id=my-app", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestAdminServer_GetApplication_MissingID(t *testing.T) {
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

func TestAdminServer_GetApplication_NotFound(t *testing.T) {
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

func TestAdminServer_GetInstance_ByID(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/"+instance.ID, nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var inst ApplicationInstance
	if err := json.NewDecoder(rr.Body).Decode(&inst); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if inst.ID != instance.ID {
		t.Errorf("expected instance ID %s, got %s", instance.ID, inst.ID)
	}
}

func TestAdminServer_GetInstance_ByQueryParam(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances?id="+instance.ID, nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestAdminServer_GetInstance_MissingID(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/nonexistent", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestAdminServer_GetInstance_NotFound(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/nonexistent", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestAdminServer_GetHealth_NoHealth(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/"+instance.ID+"/health?id="+instance.ID, nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for nil health, got %d", rr.Code)
	}
}

func TestAdminServer_GetHealth_NotFound(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/nonexistent/health", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestAdminServer_GetHealth_MissingID(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/nonexistent/health", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestAdminServer_GetMetrics_NotFound(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/nonexistent/metrics", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestAdminServer_GetMetrics_MissingID(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/nonexistent/metrics", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestAdminServer_GetMetrics_ByQueryParam(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	_ = registry.UpdateMetrics(instance.ID, map[string]float64{"cpu": 75.0})

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/"+instance.ID+"/metrics?id="+instance.ID, nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestAdminServer_Register_InvalidBody(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("POST", "/admin/register", bytes.NewBufferString("not json"))
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestAdminServer_Register_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/register", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestAdminServer_Deregister_InvalidBody(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("POST", "/admin/deregister", bytes.NewBufferString("not json"))
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestAdminServer_Deregister_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/deregister", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestAdminServer_Register_WithStatusAndMetrics(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	instance := map[string]any{
		"id":             "inst-1",
		"application_id": "my-app",
		"url":            "http://localhost:8080",
		"status":         "UP",
		"metrics": map[string]float64{
			"cpu": 50.0,
		},
	}

	body, _ := json.Marshal(instance)
	req := httptest.NewRequest("POST", "/admin/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	inst, err := registry.GetInstance("inst-1")
	if err != nil {
		t.Fatalf("GetInstance error: %v", err)
	}

	if inst.Status != StatusUp {
		t.Errorf("expected status UP, got %s", inst.Status)
	}

	if inst.Metrics["cpu"] != 50.0 {
		t.Errorf("expected cpu metric 50.0, got %f", inst.Metrics["cpu"])
	}
}

func TestApplicationRegistry_UpdateHealth_NotFound(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	err := registry.UpdateHealth("nonexistent", NewHealthInfo(StatusUp))
	if err == nil {
		t.Error("expected error for nonexistent instance")
	}
}

func TestApplicationRegistry_UpdateMetrics_NotFound(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	err := registry.UpdateMetrics("nonexistent", map[string]float64{"cpu": 50.0})
	if err == nil {
		t.Error("expected error for nonexistent instance")
	}
}

func TestApplicationInstance_AddMetric_NilMetrics(t *testing.T) {
	t.Parallel()
	instance := &ApplicationInstance{
		ID:            "test",
		ApplicationID: "app",
		Metrics:       nil,
	}

	instance.AddMetric("cpu", 75.0)

	if instance.Metrics == nil {
		t.Fatal("expected Metrics to be initialized")
	}

	value, exists := instance.GetMetric("cpu")
	if !exists {
		t.Error("expected metric to exist")
	}

	if value != 75.0 {
		t.Errorf("expected 75.0, got %f", value)
	}
}

func TestApplicationInstance_GetMetric_NotFound(t *testing.T) {
	t.Parallel()
	instance := NewApplicationInstance("app", "http://localhost:8080")

	_, exists := instance.GetMetric("nonexistent")
	if exists {
		t.Error("expected false for nonexistent metric")
	}
}

func TestApplicationRegistry_UpdateApplicationStatus_AllUnknown(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	inst1 := NewApplicationInstance("app", "http://localhost:8081")
	inst2 := NewApplicationInstance("app", "http://localhost:8082")
	registry.Register(inst1)
	registry.Register(inst2)

	app, _ := registry.GetApplication("app")
	if app.Status != StatusUnknown {
		t.Errorf("expected status UNKNOWN, got %s", app.Status)
	}
}

func TestAdminServer_OverallHealth_Empty(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/health", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestApplicationRegistry_Deregister_OneOfMany(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	inst1 := NewApplicationInstance("app", "http://localhost:8081")
	inst2 := NewApplicationInstance("app", "http://localhost:8082")
	registry.Register(inst1)
	registry.Register(inst2)

	registry.Deregister(inst1.ID)

	if registry.GetInstanceCount() != 1 {
		t.Errorf("expected 1 instance, got %d", registry.GetInstanceCount())
	}

	app, err := registry.GetApplication("app")
	if err != nil {
		t.Fatalf("app should still exist: %v", err)
	}

	if len(app.Instances) != 1 {
		t.Errorf("expected 1 instance in app, got %d", len(app.Instances))
	}
}

func TestApplicationRegistry_Deregister_AppDeletedWhenEmpty(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	inst1 := NewApplicationInstance("app1", "http://localhost:8081")
	registry.Register(inst1)

	registry.Deregister(inst1.ID)

	_, err := registry.GetApplication("app1")
	if err == nil {
		t.Error("expected app to be deleted when last instance deregistered")
	}
}
