package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	_ "github.com/xudefa/enhance/starter/echo" // 触发 Echo 自动配置注册
	"github.com/xudefa/enhance/tracing"
)

// monitoringEndpoint 描述一个监控端点及其展示名称。
type monitoringEndpoint struct {
	name string
	path string
}

func TestEchoTracingIntegration(t *testing.T) {
	echoServer, app, tracer := newTestServer(t, "test-echo-tracing")
	defer app.Stop()

	echoServer.GET("/api/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	})

	echoServer.GET("/api/error", func(c echo.Context) error {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "test error"})
	})

	echoServer.GET("/api/spans", func(c echo.Context) error {
		spans := tracer.GetSpans()
		return c.JSON(http.StatusOK, map[string]interface{}{
			"total_spans": len(spans),
			"spans":       spans,
		})
	})

	t.Run("测试正常请求", func(t *testing.T) {
		testHelloEndpoint(t, echoServer)
	})
	t.Run("测试错误请求", func(t *testing.T) {
		testErrorEndpoint(t, echoServer)
	})
	t.Run("测试链路数据", func(t *testing.T) {
		testSpansEndpoint(t, echoServer)
	})
	t.Run("测试链路传播", func(t *testing.T) {
		testTracePropagation(t, echoServer)
	})
}

func TestEchoTracingHTTP(t *testing.T) {
	echoServer, app, tracer := newHTTPServer(t, "test-echo-tracing-http")
	defer app.Stop()

	echoServer.GET("/api/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	})

	echoServer.GET("/api/spans", func(c echo.Context) error {
		spans := tracer.GetSpans()
		return c.JSON(http.StatusOK, map[string]interface{}{
			"total_spans": len(spans),
			"spans":       spans,
		})
	})

	t.Run("HTTP 正常请求", func(t *testing.T) {
		assertGetOKWithTraceID(t, "http://localhost:18083/api/hello")
	})
	t.Run("HTTP 查看链路数据", func(t *testing.T) {
		assertGetNonEmptyBody(t, "http://localhost:18083/api/spans")
	})
	t.Run("HTTP 监控端点-健康检查", func(t *testing.T) {
		assertHealthEndpointTracing(t, "http://localhost:18083/actuator/health")
	})
	t.Run("HTTP 监控端点-指标", func(t *testing.T) {
		assertMetricsEndpointTracing(t, "http://localhost:18083/actuator/metrics")
	})
}

func TestEchoActuatorEndpointsWithTracing(t *testing.T) {
	echoServer, app, tracer := newTestServer(t, "test-echo-actuator-tracing", boot.WithProperty("actuator.path", "/actuator"))
	defer app.Stop()

	monitoringEndpoints := []monitoringEndpoint{
		{"健康检查端点", "/actuator/health"},
		{"指标端点", "/actuator/metrics"},
		{"环境信息端点", "/actuator/env"},
		{"Bean列表端点", "/actuator/beans"},
		{"应用信息端点", "/actuator/info"},
	}

	for _, ep := range monitoringEndpoints {
		t.Run("监控端点_"+ep.name, func(t *testing.T) {
			assertMonitoringEndpoint(t, echoServer, tracer, ep)
		})
	}
}

func TestEchoActuatorEndpointsTracingContextPropagation(t *testing.T) {
	echoServer := echo.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-echo-actuator-tracing-propagation"),
		boot.WithProperty("echo.enabled", "true"),
		boot.WithProperty("tracing.enabled", "true"),
		boot.WithProperty("tracing.service_name", "test-service"),
		boot.WithProperty("tracing.sampling_rate", "1.0"),
		boot.WithProperty("actuator.enabled", "true"),
		boot.WithProperty("actuator.path", "/actuator"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}
	defer app.Stop()

	ctx := app.Context()
	if err := ctx.Container().RegisterInstance(echoServer, reflect.TypeFor[*echo.Echo]()); err != nil {
		t.Fatalf("注册 Echo Server 失败: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	t.Run("监控端点链路传播", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
		req.Header.Set("X-Trace-ID", "test-trace-actuator-123")
		req.Header.Set("X-Span-ID", "test-span-actuator-456")
		rec := httptest.NewRecorder()
		echoServer.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, rec.Code)
		}

		traceID := rec.Header().Get("X-Trace-ID")
		if traceID == "" {
			t.Error("响应头中缺少 X-Trace-ID")
		}

		spanID := rec.Header().Get("X-Span-ID")
		if spanID == "" {
			t.Error("响应头中缺少 X-Span-ID")
		}

		if traceID != "test-trace-actuator-123" {
			t.Errorf("期望 TraceID 传播 'test-trace-actuator-123'，实际 '%s'", traceID)
		}
	})
}

// newTestServer 构建并同步启动一个测试用 Echo 应用，返回 Echo Server、Boot 与 Tracer。
func newTestServer(t *testing.T, appName string, extra ...boot.BootOption) (*echo.Echo, *boot.Boot, *tracing.Tracer) {
	echoServer := echo.New()

	opts := []boot.BootOption{
		boot.WithAppName(appName),
		boot.WithProperty("echo.enabled", "true"),
		boot.WithProperty("tracing.enabled", "true"),
		boot.WithProperty("tracing.service_name", "test-service"),
		boot.WithProperty("tracing.sampling_rate", "1.0"),
		boot.WithProperty("actuator.enabled", "true"),
	}
	app, err := boot.NewApplication(append(opts, extra...)...)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	ctx := app.Context()
	if err := ctx.Container().RegisterInstance(echoServer, reflect.TypeFor[*echo.Echo]()); err != nil {
		t.Fatalf("注册 Echo Server 失败: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}
	return echoServer, app, tracer
}

// newHTTPServer 以异步方式启动 Echo 应用并等待服务就绪，供真实 HTTP 测试使用。
func newHTTPServer(t *testing.T, appName string) (*echo.Echo, *boot.Boot, *tracing.Tracer) {
	echoServer := echo.New()

	app, err := boot.NewApplication(
		boot.WithAppName(appName),
		boot.WithProperty("echo.enabled", "true"),
		boot.WithProperty("echo.port", "18083"),
		boot.WithProperty("tracing.enabled", "true"),
		boot.WithProperty("tracing.service_name", "test-service"),
		boot.WithProperty("tracing.sampling_rate", "1.0"),
		boot.WithProperty("actuator.enabled", "true"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}

	ctx := app.Context()
	if err := ctx.Container().RegisterInstance(echoServer, reflect.TypeFor[*echo.Echo]()); err != nil {
		t.Fatalf("注册 Echo Server 失败: %v", err)
	}

	go app.Start()

	if err := waitForServer("http://localhost:18083/api/hello", 10, 500*time.Millisecond); err != nil {
		t.Fatalf("Server did not start in time: %v", err)
	}

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}
	return echoServer, app, tracer
}

// testHelloEndpoint 验证 /api/hello 返回 200 与正确的响应体。
func testHelloEndpoint(t *testing.T, handler http.Handler) {
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("解码响应失败: %v", err)
	}
	if resp["message"] != "ok" {
		t.Errorf("期望响应 message='ok'，实际 '%s'", resp["message"])
	}
}

// testErrorEndpoint 验证 /api/error 返回 500 状态码。
func testErrorEndpoint(t *testing.T, handler http.Handler) {
	req := httptest.NewRequest(http.MethodGet, "/api/error", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusInternalServerError, rec.Code)
	}
}

// testSpansEndpoint 验证 /api/spans 返回链路追踪数据。
func testSpansEndpoint(t *testing.T, handler http.Handler) {
	req := httptest.NewRequest(http.MethodGet, "/api/spans", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("解码响应失败: %v", err)
	}
	if resp["total_spans"] == nil {
		t.Error("期望返回链路数据，实际为空")
	}
}

// testTracePropagation 验证 X-Trace-ID / X-Span-ID 请求头在响应中回传。
func testTracePropagation(t *testing.T, handler http.Handler) {
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	req.Header.Set("X-Trace-ID", "test-trace-123")
	req.Header.Set("X-Span-ID", "test-span-456")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, rec.Code)
	}

	traceID := rec.Header().Get("X-Trace-ID")
	if traceID == "" {
		t.Error("响应头中缺少 X-Trace-ID")
	}

	spanID := rec.Header().Get("X-Span-ID")
	if spanID == "" {
		t.Error("响应头中缺少 X-Span-ID")
	}
}

// assertGetOKWithTraceID 发起 GET 请求并断言返回 200 与 X-Trace-ID 响应头。
func assertGetOKWithTraceID(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, resp.StatusCode)
	}

	traceID := resp.Header.Get("X-Trace-ID")
	if traceID == "" {
		t.Error("响应头中缺少 X-Trace-ID")
	}
}

// assertGetNonEmptyBody 发起 GET 请求并断言响应体非空。
func assertGetNonEmptyBody(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("期望返回链路数据，实际为空")
	}
}

// assertHealthEndpointTracing 验证健康检查端点的链路追踪响应头。
func assertHealthEndpointTracing(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, resp.StatusCode)
	}

	traceID := resp.Header.Get("X-Trace-ID")
	if traceID == "" {
		t.Error("监控端点响应头中缺少 X-Trace-ID")
	}

	spanID := resp.Header.Get("X-Span-ID")
	if spanID == "" {
		t.Error("监控端点响应头中缺少 X-Span-ID")
	}

	t.Logf("监控端点 /actuator/health 链路追踪: TraceID=%s, SpanID=%s", traceID, spanID)
}

// assertMetricsEndpointTracing 验证指标端点的链路追踪响应头。
func assertMetricsEndpointTracing(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, resp.StatusCode)
	}

	traceID := resp.Header.Get("X-Trace-ID")
	if traceID == "" {
		t.Error("监控端点响应头中缺少 X-Trace-ID")
	}

	t.Logf("监控端点 /actuator/metrics 链路追踪: TraceID=%s", traceID)
}

// assertMonitoringEndpoint 请求监控端点并断言链路追踪日志符合预期。
func assertMonitoringEndpoint(t *testing.T, handler http.Handler, tracer *tracing.Tracer, ep monitoringEndpoint) {
	spansBefore := len(tracer.GetSpans())

	req := httptest.NewRequest(http.MethodGet, ep.path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, rec.Code)
	}

	spansAfter := len(tracer.GetSpans())
	if spansAfter <= spansBefore {
		t.Errorf("期望产生新的链路追踪日志，实际未产生: spans before=%d, after=%d", spansBefore, spansAfter)
	}

	spans := tracer.GetSpans()
	if len(spans) == 0 {
		t.Fatal("未产生任何链路追踪日志")
	}

	latestSpan := spans[len(spans)-1]

	if latestSpan.Name == "" {
		t.Error("链路追踪日志中 Span 名称为空")
	}

	if latestSpan.TraceID == "" {
		t.Error("链路追踪日志中 TraceID 为空")
	}

	if latestSpan.SpanID == "" {
		t.Error("链路追踪日志中 SpanID 为空")
	}

	if latestSpan.Tags["http.method"] != "GET" {
		t.Errorf("期望 http.method='GET'，实际 '%s'", latestSpan.Tags["http.method"])
	}

	if latestSpan.Tags["http.url"] != ep.path {
		t.Errorf("期望 http.url='%s'，实际 '%s'", ep.path, latestSpan.Tags["http.url"])
	}

	t.Logf("监控端点 [%s] 链路追踪日志: TraceID=%s, SpanID=%s, Name=%s, Status=%s",
		ep.name, latestSpan.TraceID, latestSpan.SpanID, latestSpan.Name, latestSpan.Status)
}

func waitForServer(url string, maxRetries int, delay time.Duration) error {
	for range maxRetries {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(delay)
	}
	return &httpError{url: url}
}

type httpError struct {
	url string
}

func (e *httpError) Error() string {
	return "server at " + e.url + " did not respond"
}
