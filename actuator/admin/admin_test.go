package admin

import (
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
