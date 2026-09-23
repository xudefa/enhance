package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAdminServer_ListApplications(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/applications", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var apps []*Application
	if err := json.NewDecoder(rr.Body).Decode(&apps); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(apps) != 1 {
		t.Errorf("expected 1 application, got %d", len(apps))
	}
}

func TestAdminServer_ListInstances(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var instances []*ApplicationInstance
	if err := json.NewDecoder(rr.Body).Decode(&instances); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(instances) != 1 {
		t.Errorf("expected 1 instance, got %d", len(instances))
	}
}

func TestAdminServer_Register(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	instance := map[string]any{
		"id":             "instance-1",
		"application_id": "my-app",
		"name":           "my-app",
		"url":            "http://localhost:8080",
	}

	body, _ := json.Marshal(instance)
	req := httptest.NewRequest("POST", "/admin/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	if registry.GetInstanceCount() != 1 {
		t.Errorf("expected 1 instance, got %d", registry.GetInstanceCount())
	}
}

func TestAdminServer_Deregister(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	server := NewAdminServer(registry)

	reqBody := map[string]string{
		"instance_id": instance.ID,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/admin/deregister", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if registry.GetInstanceCount() != 0 {
		t.Errorf("expected 0 instance after deregister, got %d", registry.GetInstanceCount())
	}
}

func TestAdminServer_OverallHealth(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	// 注册两个应用
	instance1 := NewApplicationInstance("app1", "http://localhost:8081")
	instance2 := NewApplicationInstance("app2", "http://localhost:8082")

	registry.Register(instance1)
	registry.Register(instance2)

	// 更新健康状态
	_ = registry.UpdateHealth(instance1.ID, NewHealthInfo(StatusUp))
	_ = registry.UpdateHealth(instance2.ID, NewHealthInfo(StatusDown))

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/health", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var health map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	status := health["status"].(map[string]any)
	if int(status["total"].(float64)) != 2 {
		t.Errorf("expected total 2, got %d", int(status["total"].(float64)))
	}

	if int(status["up"].(float64)) != 1 {
		t.Errorf("expected up 1, got %d", int(status["up"].(float64)))
	}

	if int(status["down"].(float64)) != 1 {
		t.Errorf("expected down 1, got %d", int(status["down"].(float64)))
	}
}

func TestAdminServer_GetHealth(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	health := NewHealthInfo(StatusUp)
	_ = registry.UpdateHealth(instance.ID, health)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/"+instance.ID+"/health?id="+instance.ID, nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var healthResp HealthInfo
	if err := json.NewDecoder(rr.Body).Decode(&healthResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if healthResp.Status != StatusUp {
		t.Errorf("expected status UP, got %s", healthResp.Status)
	}
}

func TestAdminServer_GetMetrics(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	registry.Register(instance)

	metrics := map[string]float64{
		"cpu_usage": 75.5,
	}
	_ = registry.UpdateMetrics(instance.ID, metrics)

	server := NewAdminServer(registry)

	req := httptest.NewRequest("GET", "/admin/instances/"+instance.ID+"/metrics?id="+instance.ID, nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var metricsResp map[string]float64
	if err := json.NewDecoder(rr.Body).Decode(&metricsResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if metricsResp["cpu_usage"] != 75.5 {
		t.Errorf("expected cpu_usage 75.5, got %f", metricsResp["cpu_usage"])
	}
}

func TestApplicationInstance_AddMetric(t *testing.T) {
	t.Parallel()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")

	instance.AddMetric("cpu_usage", 75.5)
	instance.AddMetric("memory_usage", 80.2)

	value, exists := instance.GetMetric("cpu_usage")
	if !exists {
		t.Error("expected cpu_usage metric to exist")
	}

	if value != 75.5 {
		t.Errorf("expected cpu_usage 75.5, got %f", value)
	}

	value, exists = instance.GetMetric("memory_usage")
	if !exists {
		t.Error("expected memory_usage metric to exist")
	}

	if value != 80.2 {
		t.Errorf("expected memory_usage 80.2, got %f", value)
	}
}

func TestApplicationInstance_IsHealthy(t *testing.T) {
	t.Parallel()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")

	if instance.IsHealthy() {
		t.Error("expected instance to be unhealthy initially")
	}

	// 更新健康状态
	instance.Status = StatusUp

	if !instance.IsHealthy() {
		t.Error("expected instance to be healthy after status update")
	}
}

func TestHealthInfo_AddComponent(t *testing.T) {
	t.Parallel()
	health := NewHealthInfo(StatusUp)

	health.AddComponent("database", StatusUp, map[string]any{
		"connection": "ok",
	})

	health.AddComponent("cache", StatusDown, map[string]any{
		"error": "connection refused",
	})

	if len(health.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(health.Components))
	}

	if health.Components["database"].Status != StatusUp {
		t.Errorf("expected database status UP, got %s", health.Components["database"].Status)
	}

	if health.Components["cache"].Status != StatusDown {
		t.Errorf("expected cache status DOWN, got %s", health.Components["cache"].Status)
	}
}

func TestHealthInfo_AddDetail(t *testing.T) {
	t.Parallel()
	health := NewHealthInfo(StatusUp)

	health.AddDetail("version", "1.0.0")
	health.AddDetail("uptime", 3600)

	if health.Details["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", health.Details["version"])
	}

	if health.Details["uptime"] != 3600 {
		t.Errorf("expected uptime 3600, got %v", health.Details["uptime"])
	}
}

func TestApplicationRegistry_GetUpInstanceCount(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance1 := NewApplicationInstance("app1", "http://localhost:8081")
	instance2 := NewApplicationInstance("app2", "http://localhost:8082")

	registry.Register(instance1)
	registry.Register(instance2)

	// 更新一个实例为 UP
	_ = registry.UpdateHealth(instance1.ID, NewHealthInfo(StatusUp))

	if registry.GetUpInstanceCount() != 1 {
		t.Errorf("expected 1 up instance, got %d", registry.GetUpInstanceCount())
	}
}

func TestApplicationRegistry_GetDownInstanceCount(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	instance1 := NewApplicationInstance("app1", "http://localhost:8081")
	instance2 := NewApplicationInstance("app2", "http://localhost:8082")

	registry.Register(instance1)
	registry.Register(instance2)

	// 更新一个实例为 DOWN
	_ = registry.UpdateHealth(instance1.ID, NewHealthInfo(StatusDown))

	if registry.GetDownInstanceCount() != 1 {
		t.Errorf("expected 1 down instance, got %d", registry.GetDownInstanceCount())
	}
}

func TestAdminServer_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	// 测试 POST 到 GET 端点
	req := httptest.NewRequest("POST", "/admin/applications", nil)
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestAdminServer_RegisterMissingFields(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	// 缺少必需字段
	instance := map[string]any{
		"id": "instance-1",
	}

	body, _ := json.Marshal(instance)
	req := httptest.NewRequest("POST", "/admin/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestAdminServer_DeregisterMissingInstanceID(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()
	server := NewAdminServer(registry)

	// 缺少 instance_id
	reqBody := map[string]string{}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/admin/deregister", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestApplicationInstance_Timestamps(t *testing.T) {
	t.Parallel()
	before := time.Now()
	instance := NewApplicationInstance("my-app", "http://localhost:8080")
	after := time.Now()

	if instance.RegisteredAt.Before(before) || instance.RegisteredAt.After(after) {
		t.Error("expected RegisteredAt to be set to current time")
	}

	if instance.LastSeen.Before(before) || instance.LastSeen.After(after) {
		t.Error("expected LastSeen to be set to current time")
	}
}

func TestApplicationRegistry_DeregisterNonExistent(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	// 注销不存在的实例
	registry.Deregister("nonexistent")

	// 不应该报错
	if registry.GetInstanceCount() != 0 {
		t.Errorf("expected 0 instances, got %d", registry.GetInstanceCount())
	}
}

func TestApplicationRegistry_GetNonExistentApplication(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	_, err := registry.GetApplication("nonexistent")
	if err == nil {
		t.Error("expected error when getting non-existent application")
	}
}

func TestApplicationRegistry_GetNonExistentInstance(t *testing.T) {
	t.Parallel()
	registry := NewApplicationRegistry()

	_, err := registry.GetInstance("nonexistent")
	if err == nil {
		t.Error("expected error when getting non-existent instance")
	}
}
