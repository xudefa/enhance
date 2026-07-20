package health

import (
	"context"
	"testing"
	"time"
)

// TestSimpleIndicator 验证简单健康指标功能:
//  1. 创建和名称
//  2. 执行健康检查
//  3. nil 函数处理
func TestSimpleIndicator(t *testing.T) {
	t.Parallel()
	// 测试正常指标
	indicator := NewSimpleIndicator("test", func(ctx context.Context) Health {
		return Health{
			Status:    StatusUp,
			Timestamp: time.Now(),
		}
	})

	if indicator.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", indicator.Name())
	}

	health := indicator.Health(context.Background())
	if health.Status != StatusUp {
		t.Errorf("expected status UP, got %s", health.Status)
	}

	// 测试 nil 函数
	nilIndicator := NewSimpleIndicator("nil", nil)
	nilHealth := nilIndicator.Health(context.Background())
	if nilHealth.Status != StatusUnknown {
		t.Errorf("expected status UNKNOWN for nil function, got %s", nilHealth.Status)
	}
}

// TestHealthBuilder 验证 HealthBuilder 功能:
//  1. Up/Down/Degraded/Outage 状态
//  2. WithDetail 添加详细信息
//  3. WithError 设置错误
//  4. Build 构建
func TestHealthBuilder(t *testing.T) {
	t.Parallel()
	// 测试 Up 状态
	health1 := Up().WithDetail("version", "1.0").Build()
	if health1.Status != StatusUp {
		t.Errorf("expected status UP, got %s", health1.Status)
	}
	if health1.Details["version"] != "1.0" {
		t.Errorf("expected detail version='1.0', got '%v'", health1.Details["version"])
	}

	// 测试 Down 状态
	health2 := Down().WithDetail("error", "connection failed").Build()
	if health2.Status != StatusDown {
		t.Errorf("expected status DOWN, got %s", health2.Status)
	}

	// 测试 Degraded 状态
	health3 := Degraded().WithDetail("latency", "500ms").Build()
	if health3.Status != StatusDegraded {
		t.Errorf("expected status DEGRADED, got %s", health3.Status)
	}

	// 测试 Outage 状态
	health4 := Outage().Build()
	if health4.Status != StatusOutage {
		t.Errorf("expected status OUTAGE, got %s", health4.Status)
	}

	// 测试 WithError
	err := &testError{"test error"}
	health5 := Down().WithError(err).Build()
	if health5.Error != err {
		t.Error("expected error to be set")
	}

	// 测试 WithDetails
	details := map[string]any{"key1": "value1", "key2": 123}
	health6 := Up().WithDetails(details).Build()
	if health6.Details["key1"] != "value1" {
		t.Errorf("expected detail key1='value1', got '%v'", health6.Details["key1"])
	}
	if health6.Details["key2"] != 123 {
		t.Errorf("expected detail key2=123, got '%v'", health6.Details["key2"])
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestIndicatorRegistry 验证指标注册表功能:
//  1. 注册和查询
//  2. 获取所有指标
//  3. 移除指标
//  4. RegisterFunc 函数式注册
func TestIndicatorRegistry(t *testing.T) {
	t.Parallel()
	registry := NewIndicatorRegistry()

	// 测试注册和查询
	indicator1 := NewSimpleIndicator("db", func(ctx context.Context) Health {
		return Up().Build()
	})
	registry.Register(indicator1)

	got, ok := registry.Get("db")
	if !ok {
		t.Error("expected to find 'db' indicator")
	}
	if got.Name() != "db" {
		t.Errorf("expected name 'db', got '%s'", got.Name())
	}

	// 测试 RegisterFunc
	registry.RegisterFunc("redis", func(ctx context.Context) Health {
		return Up().WithDetail("version", "6.2").Build()
	})

	got2, ok := registry.Get("redis")
	if !ok {
		t.Error("expected to find 'redis' indicator")
	}
	if got2.Name() != "redis" {
		t.Errorf("expected name 'redis', got '%s'", got2.Name())
	}

	// 测试 GetAll
	all := registry.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 indicators, got %d", len(all))
	}

	// 测试 Remove
	registry.Remove("db")
	_, ok = registry.Get("db")
	if ok {
		t.Error("expected 'db' indicator to be removed")
	}
}

// TestRuntimeHealthIndicator 验证运行时健康指标:
//  1. 指标名称
//  2. 健康状态
//  3. 详细信息包含运行时指标
func TestRuntimeHealthIndicator(t *testing.T) {
	t.Parallel()
	indicator := NewRuntimeHealthIndicator()

	if indicator.Name() != "runtime" {
		t.Errorf("expected name 'runtime', got '%s'", indicator.Name())
	}

	health := indicator.Health(context.Background())
	if health.Status != StatusUp {
		t.Errorf("expected status UP, got %s", health.Status)
	}

	// 验证包含关键指标
	requiredKeys := []string{"goroutines", "heap_alloc", "gc_pause_ns", "go_version"}
	for _, key := range requiredKeys {
		if _, ok := health.Details[key]; !ok {
			t.Errorf("expected detail '%s' to be present", key)
		}
	}

	// 验证 goroutine 数量大于 0
	if goroutines, ok := health.Details["goroutines"].(int); ok {
		if goroutines <= 0 {
			t.Error("expected goroutines > 0")
		}
	}
}

// TestSystemHealthIndicator 验证系统健康指标:
//  1. 指标名称
//  2. 健康状态
//  3. 详细信息包含系统信息
func TestSystemHealthIndicator(t *testing.T) {
	t.Parallel()
	indicator := NewSystemHealthIndicator()

	if indicator.Name() != "system" {
		t.Errorf("expected name 'system', got '%s'", indicator.Name())
	}

	health := indicator.Health(context.Background())
	if health.Status != StatusUp {
		t.Errorf("expected status UP, got %s", health.Status)
	}

	// 验证包含系统信息
	requiredKeys := []string{"os", "arch", "num_cpu"}
	for _, key := range requiredKeys {
		if _, ok := health.Details[key]; !ok {
			t.Errorf("expected detail '%s' to be present", key)
		}
	}
}

// TestHealthCheckService 验证健康检查服务:
//  1. 创建服务
//  2. 注册自定义指标
//  3. 执行健康检查
//  4. 获取指标
func TestHealthCheckService(t *testing.T) {
	t.Parallel()
	service := NewHealthCheckService()

	// 验证内置指标已注册
	indicators := service.GetAllIndicators()
	if len(indicators) < 2 {
		t.Errorf("expected at least 2 built-in indicators, got %d", len(indicators))
	}

	// 注册自定义指标
	service.RegisterIndicator("custom", func(ctx context.Context) Health {
		return Up().WithDetail("status", "ok").Build()
	})

	// 验证指标已注册
	customInd, ok := service.GetIndicator("custom")
	if !ok {
		t.Error("expected to find 'custom' indicator")
	}
	if customInd.Name() != "custom" {
		t.Errorf("expected name 'custom', got '%s'", customInd.Name())
	}

	// 执行健康检查
	health := service.Check(context.Background())
	if health.Status != StatusUp {
		t.Errorf("expected status UP, got %s", health.Status)
	}

	// 验证包含自定义指标的详细信息
	if _, ok := health.Details["custom"]; !ok {
		t.Error("expected 'custom' indicator in health details")
	}
}

// TestGlobalRegistry 验证全局注册表功能:
//  1. 注册到全局
//  2. 从全局获取
func TestGlobalRegistry(t *testing.T) {
	t.Parallel()
	// 使用新的指标名称避免影响其他测试
	indicatorName := "test_global_indicator"

	RegisterIndicator(indicatorName, func(ctx context.Context) Health {
		return Up().WithDetail("test", "global").Build()
	})

	registry := GlobalIndicatorRegistry()
	got, ok := registry.Get(indicatorName)
	if !ok {
		t.Errorf("expected to find '%s' in global registry", indicatorName)
	}
	if got.Name() != indicatorName {
		t.Errorf("expected name '%s', got '%s'", indicatorName, got.Name())
	}
}

// TestDefaultHealthCheckService 验证默认健康检查服务:
//  1. CheckHealth 函数
//  2. RegisterCustomIndicator 函数
func TestDefaultHealthCheckService(t *testing.T) {
	t.Parallel()
	// 注册自定义指标
	customName := "test_default_service"
	RegisterCustomIndicator(customName, func(ctx context.Context) Health {
		return Up().WithDetail("version", "1.0").Build()
	})

	// 执行健康检查
	health := CheckHealth(context.Background())
	if health.Status != StatusUp {
		t.Errorf("expected status UP, got %s", health.Status)
	}

	// 验证包含自定义指标
	if _, ok := health.Details[customName]; !ok {
		t.Errorf("expected '%s' indicator in health details", customName)
	}
}
