package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplicationRegistry_Register(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	if registry.GetApplicationCount() != 1 {
		t.Errorf("expected 1 application, got %d", registry.GetApplicationCount())
	}

	if registry.GetInstanceCount() != 1 {
		t.Errorf("expected 1 instance, got %d", registry.GetInstanceCount())
	}

	// 获取应用
	app, err := registry.GetApplication("my-app")
	if err != nil {
		t.Fatalf("Failed to get application: %v", err)
	}

	if app.Name != "my-app" {
		t.Errorf("expected application name 'my-app', got %s", app.Name)
	}

	if len(app.Instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(app.Instances))
	}

	// 获取实例
	retrieved, err := registry.GetInstance(instance.ID)
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if retrieved.URL != "http://localhost:8080" {
		t.Errorf("expected instance URL 'http://localhost:8080', got %s", retrieved.URL)
	}
}

func TestApplicationRegistry_Deregister(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	if registry.GetInstanceCount() != 1 {
		t.Fatalf("expected 1 instance before deregister")
	}

	// 注销实例
	registry.Deregister(instance.ID)

	if registry.GetInstanceCount() != 0 {
		t.Errorf("expected 0 instance after deregister, got %d", registry.GetInstanceCount())
	}

	if registry.GetApplicationCount() != 0 {
		t.Errorf("expected 0 application after deregister, got %d", registry.GetApplicationCount())
	}
}

func TestApplicationRegistry_UpdateHealth(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	// 更新健康信息
	health := NewHealthInfo(StatusUp)
	health.AddComponent("database", StatusUp, map[string]any{
		"connection": "ok",
	})
	health.AddDetail("version", "1.0.0")

	err := registry.UpdateHealth(instance.ID, health)
	if err != nil {
		t.Fatalf("Failed to update health: %v", err)
	}

	// 验证健康信息
	retrieved, err := registry.GetInstance(instance.ID)
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if retrieved.Status != StatusUp {
		t.Errorf("expected status UP, got %s", retrieved.Status)
	}

	if retrieved.Health == nil {
		t.Fatal("expected health information to be set")
	}

	if len(retrieved.Health.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(retrieved.Health.Components))
	}
}

func TestApplicationRegistry_UpdateMetrics(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	// 更新指标
	metrics := map[string]float64{
		"cpu_usage":    75.5,
		"memory_usage": 80.2,
	}

	err := registry.UpdateMetrics(instance.ID, metrics)
	if err != nil {
		t.Fatalf("Failed to update metrics: %v", err)
	}

	// 验证指标信息
	retrieved, err := registry.GetInstance(instance.ID)
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	if retrieved.Metrics["cpu_usage"] != 75.5 {
		t.Errorf("expected cpu_usage 75.5, got %f", retrieved.Metrics["cpu_usage"])
	}

	if retrieved.Metrics["memory_usage"] != 80.2 {
		t.Errorf("expected memory_usage 80.2, got %f", retrieved.Metrics["memory_usage"])
	}
}

func TestApplicationRegistry_ListApplications(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	// 注册多个应用
	instance1 := NewApplicationInstance("app1", "http://localhost:8081")
	instance2 := NewApplicationInstance("app2", "http://localhost:8082")

	registry.Register(instance1)
	registry.Register(instance2)

	apps := registry.ListApplications()
	if len(apps) != 2 {
		t.Errorf("expected 2 applications, got %d", len(apps))
	}
}

func TestApplicationRegistry_ListInstances(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	// 注册多个实例
	instance1 := NewApplicationInstance("app1", "http://localhost:8081")
	instance2 := NewApplicationInstance("app1", "http://localhost:8082")

	registry.Register(instance1)
	registry.Register(instance2)

	instances := registry.ListInstances()
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}
}

func TestApplicationRegistry_UpdateApplicationStatus(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	// 注册两个实例
	instance1 := NewApplicationInstance("my-app", "http://localhost:8081")
	instance2 := NewApplicationInstance("my-app", "http://localhost:8082")

	registry.Register(instance1)
	registry.Register(instance2)

	// 更新两个实例都为 UP
	health1 := NewHealthInfo(StatusUp)
	health2 := NewHealthInfo(StatusUp)

	_ = registry.UpdateHealth(instance1.ID, health1)
	_ = registry.UpdateHealth(instance2.ID, health2)

	// 验证应用状态为 UP
	app, _ := registry.GetApplication("my-app")
	if app.Status != StatusUp {
		t.Errorf("expected application status UP, got %s", app.Status)
	}

	// 更新一个实例为 DOWN
	health2Down := NewHealthInfo(StatusDown)
	_ = registry.UpdateHealth(instance2.ID, health2Down)

	// 验证应用状态为 OUT_OF_SERVICE
	app, _ = registry.GetApplication("my-app")
	if app.Status != StatusOutOfService {
		t.Errorf("expected application status OUT_OF_SERVICE, got %s", app.Status)
	}

	// 更新两个实例都为 DOWN
	health1Down := NewHealthInfo(StatusDown)
	_ = registry.UpdateHealth(instance1.ID, health1Down)

	// 验证应用状态为 DOWN
	app, _ = registry.GetApplication("my-app")
	if app.Status != StatusDown {
		t.Errorf("expected application status DOWN, got %s", app.Status)
	}
}

func TestAdminServer_GetApplication_ByID_Coverage(t *testing.T) {
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

func TestAdminServer_GetApplication_ByQueryParam_Coverage(t *testing.T) {
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

func TestAdminServer_GetApplication_MissingID_Coverage(t *testing.T) {
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

func TestAdminServer_GetApplication_NotFound_Coverage(t *testing.T) {
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

func TestAdminServer_GetInstance_ByID_Coverage(t *testing.T) {
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

func TestAdminServer_GetInstance_ByQueryParam_Coverage(t *testing.T) {
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

func TestAdminServer_GetInstance_MissingID_Coverage(t *testing.T) {
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

func TestAdminServer_GetInstance_NotFound_Coverage(t *testing.T) {
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

func TestAdminServer_GetHealth_NoHealth_Coverage(t *testing.T) {
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

func TestAdminServer_GetHealth_NotFound_Coverage(t *testing.T) {
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

func TestAdminServer_GetHealth_MissingID_Coverage(t *testing.T) {
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

func TestAdminServer_GetMetrics_NotFound_Coverage(t *testing.T) {
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

func TestAdminServer_GetMetrics_MissingID_Coverage(t *testing.T) {
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

func TestAdminServer_GetMetrics_ByQueryParam_Coverage(t *testing.T) {
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

func TestAdminServer_Register_InvalidBody_Coverage(t *testing.T) {
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

func TestAdminServer_Register_MethodNotAllowed_Coverage(t *testing.T) {
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

func TestAdminServer_Deregister_InvalidBody_Coverage(t *testing.T) {
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

func TestAdminServer_Deregister_MethodNotAllowed_Coverage(t *testing.T) {
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

func TestAdminServer_Register_WithStatusAndMetrics_Coverage(t *testing.T) {
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

func TestApplicationRegistry_UpdateHealth_NotFound_Coverage(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	err := registry.UpdateHealth("nonexistent", NewHealthInfo(StatusUp))
	if err == nil {
		t.Error("expected error for nonexistent instance")
	}
}

func TestApplicationRegistry_UpdateMetrics_NotFound_Coverage(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	err := registry.UpdateMetrics("nonexistent", map[string]float64{"cpu": 50.0})
	if err == nil {
		t.Error("expected error for nonexistent instance")
	}
}

func TestApplicationInstance_AddMetric_NilMetrics_Coverage(t *testing.T) {
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

func TestApplicationInstance_GetMetric_NotFound_Coverage(t *testing.T) {
	t.Parallel()
	instance := NewApplicationInstance("app", "http://localhost:8080")

	_, exists := instance.GetMetric("nonexistent")
	if exists {
		t.Error("expected false for nonexistent metric")
	}
}

func TestApplicationRegistry_UpdateApplicationStatus_AllUnknown_Coverage(t *testing.T) {
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

func TestAdminServer_OverallHealth_Empty_Coverage(t *testing.T) {
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

func TestApplicationRegistry_Deregister_OneOfMany_Coverage(t *testing.T) {
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

func TestApplicationRegistry_Deregister_AppDeletedWhenEmpty_Coverage(t *testing.T) {
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
