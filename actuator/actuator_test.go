package actuator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/actuator/health"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

// testAppContext 测试用的应用上下文
type testAppContext struct {
	container core.Container
	env       *environment.Environment
}

func (t *testAppContext) Container() core.Container             { return t.container }
func (t *testAppContext) Environment() *environment.Environment { return t.env }

// newTestEnvironment 创建测试环境，移除默认的属性源
func newTestEnvironment() *environment.Environment {
	env := environment.NewEnvironment()
	env.RemovePropertySource("args")
	env.RemovePropertySource("env")
	return env
}

// testIndicator 测试用的健康指标
type testIndicator struct{}

func (t *testIndicator) Name() string {
	return "test"
}

func (t *testIndicator) Health(ctx context.Context) health.Health {
	return health.Health{
		Status: health.StatusUp,
		Details: map[string]any{
			"version": "1.0.0",
		},
	}
}

// downIndicator 测试用的 DOWN 健康指标
type downIndicator struct{}

func (d *downIndicator) Name() string {
	return "db"
}

func (d *downIndicator) Health(ctx context.Context) health.Health {
	return health.Health{
		Status: health.StatusDown,
		Details: map[string]any{
			"error": "connection refused",
		},
	}
}

// TestHealthHandler_DownReturns503 测试健康状态非 UP 时返回 503
func TestHealthHandler_DownReturns503(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	agg := health.NewAggregator()
	agg.AddIndicator(&downIndicator{})
	act.SetHealthAggregator(agg)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	act.HealthHandler(recorder, r)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}

	var healthResult health.Health
	if err := json.NewDecoder(recorder.Body).Decode(&healthResult); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if healthResult.Status != health.StatusDown {
		t.Fatalf("expected DOWN, got %s", healthResult.Status)
	}
}

// TestNew 测试创建 Actuator 实例
func TestNew(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}

	act := New(ctx)
	if act == nil || act.healthAggregator == nil || act.metricsRegistry == nil || act.appContext != ctx {
		t.Fatalf("Actuator not properly initialized: act=%v, health=%v, metrics=%v", act != nil, act != nil, act != nil)
	}
}

// TestActuator_NilContext 测试当应用上下文为 nil 时的错误处理
func TestActuator_NilContext(t *testing.T) {
	t.Parallel()
	act := &Actuator{}

	t.Run("EnviHandler returns 500", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/actuator/env", nil)
		act.EnvHandler(recorder, req)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("BeansHandler returns 500", func(t *testing.T) {
		t.Parallel()
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
		act.BeansHandler(recorder, req)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})
}

// TestHealthHandler_NilAggregator 测试健康聚合器为 nil 时的错误处理
func TestHealthHandler_NilAggregator(t *testing.T) {
	t.Parallel()
	act := &Actuator{}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	act.HealthHandler(recorder, r)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

// TestMetricsHandler_NilRegistry 测试指标注册表为 nil 时的错误处理
func TestMetricsHandler_NilRegistry(t *testing.T) {
	t.Parallel()
	act := &Actuator{}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/metrics", nil)
	act.MetricsHandler(recorder, r)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

// TestHealthHandler_EmptyAggregator 测试空健康聚合器的处理
func TestHealthHandler_EmptyAggregator(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	act.HealthHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var healthResult health.Health
	if err := json.NewDecoder(recorder.Body).Decode(&healthResult); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if healthResult.Status != health.StatusUp {
		t.Fatalf("expected UP, got %s", healthResult.Status)
	}
}

// TestHealthHandler_WithIndicators 测试带健康指标的处理
func TestHealthHandler_WithIndicators(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	agg := health.NewAggregator()
	agg.AddIndicator(&testIndicator{})
	act.SetHealthAggregator(agg)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	act.HealthHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var healthResult health.Health
	if err := json.NewDecoder(recorder.Body).Decode(&healthResult); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if healthResult.Status != health.StatusUp {
		t.Fatalf("expected UP, got %s", healthResult.Status)
	}
	if healthResult.Details == nil {
		t.Fatal("expected details, got nil")
	}
}

func TestDatabaseHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("nil check function returns unknown", func(t *testing.T) {
		t.Parallel()
		ind := NewDatabaseHealthIndicator(nil)
		healthResult := ind.Health(context.Background())
		if healthResult.Status != health.StatusUnknown {
			t.Fatalf("expected UNKNOWN, got %s", healthResult.Status)
		}
	})

	t.Run("successful check returns up", func(t *testing.T) {
		t.Parallel()
		ind := NewDatabaseHealthIndicator(func(ctx context.Context) error {
			return nil
		})
		healthResult := ind.Health(context.Background())
		if healthResult.Status != health.StatusUp {
			t.Fatalf("expected UP, got %s", healthResult.Status)
		}
	})

	t.Run("failed check returns down", func(t *testing.T) {
		t.Parallel()
		ind := NewDatabaseHealthIndicator(func(ctx context.Context) error {
			return assertAnError("connection refused")
		})
		healthResult := ind.Health(context.Background())
		if healthResult.Status != health.StatusDown {
			t.Fatalf("expected DOWN, got %s", healthResult.Status)
		}
		if healthResult.Details == nil {
			t.Fatal("expected details")
		}
	})
}

func TestRedisHealthIndicator(t *testing.T) {
	t.Parallel()
	t.Run("nil check function returns unknown", func(t *testing.T) {
		t.Parallel()
		ind := NewRedisHealthIndicator(nil)
		healthResult := ind.Health(context.Background())
		if healthResult.Status != health.StatusUnknown {
			t.Fatalf("expected UNKNOWN, got %s", healthResult.Status)
		}
	})

	t.Run("successful check returns up", func(t *testing.T) {
		t.Parallel()
		ind := NewRedisHealthIndicator(func(ctx context.Context) error {
			return nil
		})
		healthResult := ind.Health(context.Background())
		if healthResult.Status != health.StatusUp {
			t.Fatalf("expected UP, got %s", healthResult.Status)
		}
	})

	t.Run("failed check returns down", func(t *testing.T) {
		t.Parallel()
		ind := NewRedisHealthIndicator(func(ctx context.Context) error {
			return assertAnError("timeout")
		})
		healthResult := ind.Health(context.Background())
		if healthResult.Status != health.StatusDown {
			t.Fatalf("expected DOWN, got %s", healthResult.Status)
		}
		if healthResult.Details == nil {
			t.Fatal("expected details")
		}
	})
}

func TestDatabaseHealthIndicator_Name(t *testing.T) {
	t.Parallel()
	ind := NewDatabaseHealthIndicator(nil)
	if ind.Name() != "database" {
		t.Fatalf("expected 'database', got %s", ind.Name())
	}
}

func TestRedisHealthIndicator_Name(t *testing.T) {
	t.Parallel()
	ind := NewRedisHealthIndicator(nil)
	if ind.Name() != "redis" {
		t.Fatalf("expected 'redis', got %s", ind.Name())
	}
}

// assertAnError 返回一个固定的 error，用于测试
type assertAnError string

func (e assertAnError) Error() string { return string(e) }
