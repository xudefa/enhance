package actuator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/actuator/health"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/metrics"
)

func TestEnvHandler(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"server.port": "8080",
		"app.name":    "test-app",
	}))
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/env", nil)
	act.EnvHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var envSourceList []map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&envSourceList); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(envSourceList) == 0 {
		t.Fatal("expected at least one property source")
	}

	found := false
	for _, src := range envSourceList {
		if src["name"] == "test" {
			found = true
			props, ok := src["properties"].([]any)
			if !ok {
				t.Fatal("expected properties array")
			}
			if len(props) != 2 {
				t.Fatalf("expected 2 properties, got %d", len(props))
			}
			break
		}
	}
	if !found {
		t.Fatal("expected to find 'test' property source")
	}
}

func TestEnvHandler_NoAppContext(t *testing.T) {
	t.Parallel()
	act := &Actuator{}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/env", nil)
	act.EnvHandler(recorder, r)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestBeansHandler_EmptyContainer(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
	act.BeansHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var beanListResult map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&beanListResult); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	beans, ok := beanListResult["beans"].([]any)
	if !ok {
		t.Fatal("expected beans array")
	}
	if len(beans) != 0 {
		t.Fatalf("expected 0 beans, got %d", len(beans))
	}
}

func TestBeansHandler_WithBeans(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	type testService struct{ Name string }
	type testDB struct{}
	svc := &testService{Name: "test"}
	db := &testDB{}

	_ = container.RegisterInstance(svc, reflect.TypeOf(svc))
	_ = container.RegisterInstance(db, reflect.TypeOf(db))

	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
	act.BeansHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var beanListResult map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&beanListResult); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	beans, ok := beanListResult["beans"].([]any)
	if !ok {
		t.Fatal("expected beans array")
	}
	if len(beans) != 2 {
		t.Fatalf("expected 2 beans, got %d", len(beans))
	}
}

func TestBeansHandler_NoAppContext(t *testing.T) {
	t.Parallel()
	act := &Actuator{}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/beans", nil)
	act.BeansHandler(recorder, r)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	mux := http.NewServeMux()
	config := RouteConfig{
		BasePath:    "/actuator",
		ExposeDebug: false,
	}
	act.RegisterRoutes(&StdRouteRegistrar{Mux: mux}, config)

	routes := []string{
		"/actuator/health",
		"/actuator/metrics",
		"/actuator/env",
		"/actuator/beans",
	}

	for _, route := range routes {
		recorder := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, route, nil)
		mux.ServeHTTP(recorder, r)
		if recorder.Code != http.StatusOK {
			t.Fatalf("route %s returned %d", route, recorder.Code)
		}
	}
}

func TestSetHealthAggregator(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	agg := health.NewAggregator()
	agg.AddIndicator(&testIndicator{})
	act.SetHealthAggregator(agg)

	if act.healthAggregator == nil {
		t.Fatal("healthAggregator should not be nil")
	}
	healthResult := act.healthAggregator.Aggregate(context.Background())
	if healthResult.Status != health.StatusUp {
		t.Fatalf("expected UP, got %s", healthResult.Status)
	}
}

func TestSetMetricsRegistry(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := metrics.NewSimpleRegistry()
	reg.Counter("test").Inc()
	act.SetMetricsRegistry(reg)

	got := act.MetricsRegistry()
	if got == nil {
		t.Fatal("MetricsRegistry() returned nil")
	}
	collected := got.Collect()
	if len(collected) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(collected))
	}
}

func TestMetricsRegistry(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := act.MetricsRegistry()
	if reg == nil {
		t.Fatal("MetricsRegistry() returned nil")
	}
}

func TestMetricsHandler_Empty(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/metrics", nil)
	act.MetricsHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var metricsList []metrics.Metric
	if err := json.NewDecoder(recorder.Body).Decode(&metricsList); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(metricsList) != 0 {
		t.Fatalf("expected 0 metrics, got %d", len(metricsList))
	}
}

func TestMetricsHandler_WithData(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := metrics.NewSimpleRegistry()
	reg.Counter("requests_total").Inc()
	reg.Gauge("memory_usage").Set(1024.5)
	act.SetMetricsRegistry(reg)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/metrics", nil)
	act.MetricsHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var metricsList []metrics.Metric
	if err := json.NewDecoder(recorder.Body).Decode(&metricsList); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(metricsList) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metricsList))
	}
}

func TestInfoHandler(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"app.name":    "test-app",
		"app.version": "2.0.0",
		"build.time":  "2024-01-01T00:00:00Z",
	}))
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/info", nil)
	act.InfoHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var infoResult map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&infoResult); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	app, ok := infoResult["app"].(map[string]any)
	if !ok {
		t.Fatal("expected app object")
	}
	if app["name"] != "test-app" {
		t.Fatalf("expected app.name 'test-app', got %v", app["name"])
	}
	if app["version"] != "2.0.0" {
		t.Fatalf("expected app.version '2.0.0', got %v", app["version"])
	}
}

// TestInfoHandler_NoAppContext 测试应用上下文为 nil 时的错误处理
func TestInfoHandler_NoAppContext(t *testing.T) {
	t.Parallel()
	act := &Actuator{}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/info", nil)
	act.InfoHandler(recorder, r)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

// TestInfoHandler_NilEnvironment 测试环境为 nil 时的错误处理
func TestInfoHandler_NilEnvironment(t *testing.T) {
	t.Parallel()
	act := &Actuator{appContext: &testAppContext{container: core.NewContainer(), env: nil}}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/actuator/info", nil)
	act.InfoHandler(recorder, r)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
