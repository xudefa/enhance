package actuator

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/metrics"
)

// TestPrometheusHandler 测试 Prometheus 格式指标处理器
func TestPrometheusHandler(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := metrics.NewSimpleRegistry()
	reg.Counter("requests_total", "method", "GET").Inc()
	reg.Counter("requests_total", "method", "POST").Add(2)
	reg.Gauge("memory_usage").Set(1024.5)
	act.SetMetricsRegistry(reg)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	act.PrometheusHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "text/plain; version=0.0.4; charset=utf-8" {
		t.Fatalf("expected Content-Type 'text/plain; version=0.0.4; charset=utf-8', got %s", contentType)
	}

	body := recorder.Body.String()
	if body == "" {
		t.Fatal("expected non-empty response body")
	}

	// 验证包含指标名称
	if !contains(body, "requests_total") {
		t.Fatal("expected 'requests_total' in response")
	}
	if !contains(body, "memory_usage") {
		t.Fatal("expected 'memory_usage' in response")
	}
}

// TestPrometheusHandler_EscapedLabelsAndType 测试标签转义和 TYPE 输出
func TestPrometheusHandler_EscapedLabelsAndType(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	reg := metrics.NewSimpleRegistry()
	reg.Counter("http_requests", "path", `/api/"quoted"`).Add(1)
	act.SetMetricsRegistry(reg)

	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	act.PrometheusHandler(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "# TYPE http_requests counter") {
		t.Errorf("expected TYPE line, got:\n%s", body)
	}
	if !strings.Contains(body, `path="/api/\"quoted\""`) {
		t.Errorf("expected escaped label value, got:\n%s", body)
	}
}

// TestPrometheusHandler_NilRegistry 测试指标注册表为 nil 时的错误处理
func TestPrometheusHandler_NilRegistry(t *testing.T) {
	t.Parallel()
	act := &Actuator{}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	act.PrometheusHandler(recorder, r)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

// TestPprofHandlers 测试 pprof 端点处理器
func TestPprofHandlers(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	handlers := act.PprofHandlers()
	if handlers == nil {
		t.Fatal("PprofHandlers() returned nil")
	}

	expectedPaths := []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
	}

	for _, path := range expectedPaths {
		if _, ok := handlers[path]; !ok {
			t.Fatalf("expected path %s not found in handlers", path)
		}
	}

	if len(handlers) != len(expectedPaths) {
		t.Fatalf("expected %d handlers, got %d", len(expectedPaths), len(handlers))
	}
}

// TestRegisterDebugRoutes 测试注册调试路由
func TestRegisterDebugRoutes(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := newTestEnvironment()
	ctx := &testAppContext{container: container, env: env}
	act := New(ctx)

	mux := http.NewServeMux()
	act.RegisterDebugRoutes(&StdRouteRegistrar{Mux: mux})

	expectedPaths := []string{
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/symbol",
		"/debug/pprof/profile",
		"/debug/pprof/trace",
	}

	for _, path := range expectedPaths {
		handler, _ := mux.Handler(httptest.NewRequest(http.MethodGet, path, nil))
		if handler == nil {
			t.Fatalf("path %s has no handler registered", path)
		}
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
