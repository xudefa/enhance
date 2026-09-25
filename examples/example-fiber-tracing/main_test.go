package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	_ "github.com/xudefa/enhance/starter/fiber" // 触发 Fiber 自动配置注册
	"github.com/xudefa/enhance/tracing"
)

// httpTestRoutesEnabled 控制真实 HTTP 测试路由自动配置的开关属性键。
const httpTestRoutesEnabled = "fiber.http-test-routes"

// httpTestRoutesConfig 在 Fiber 中间件注册完成后、监听启动前注册测试路由。
//
// Fiber v2.52 的 Use 中间件仅对注册在其之后的路由生效，且路由必须在 Listen 之前注册，
// 因此测试路由需通过自动配置注入到中间件之后、监听之前这个唯一合法的时机。
type httpTestRoutesConfig struct{}

func init() {
	boot.RegisterAutoConfigWith(&httpTestRoutesConfig{},
		boot.WithConditions(condition.OnProperty(httpTestRoutesEnabled, "true")),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)+10),
		boot.WithAfter("FiberAutoConfiguration"),
	)
}

// Configure 注册真实 HTTP 测试所需的业务路由。
func (c *httpTestRoutesConfig) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	fiberApp, err := core.GetByName[*fiber.App](container, "")
	if err != nil {
		return fmt.Errorf("获取 Fiber App 失败: %w", err)
	}
	tracer, err := core.GetByName[*tracing.Tracer](container, "")
	if err != nil {
		return fmt.Errorf("获取 Tracer 失败: %w", err)
	}

	fiberApp.Get("/api/hello", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{"message": "ok"})
	})
	fiberApp.Get("/api/spans", func(c *fiber.Ctx) error {
		spans := tracer.GetSpans()
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"total_spans": len(spans),
			"spans":       spans,
		})
	})
	return nil
}

// monitoringEndpoint 描述一个监控端点及其展示名称。
type monitoringEndpoint struct {
	name string
	path string
}

func TestFiberTracingIntegration(t *testing.T) {
	fiberApp, app, tracer := newTestServer(t, "test-fiber-tracing")
	defer app.Stop()

	fiberApp.Get("/api/hello", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{"message": "ok"})
	})

	fiberApp.Get("/api/error", func(c *fiber.Ctx) error {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "test error"})
	})

	fiberApp.Get("/api/spans", func(c *fiber.Ctx) error {
		spans := tracer.GetSpans()
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"total_spans": len(spans),
			"spans":       spans,
		})
	})

	t.Run("测试正常请求", func(t *testing.T) {
		testHelloEndpoint(t, fiberApp)
	})
	t.Run("测试错误请求", func(t *testing.T) {
		testErrorEndpoint(t, fiberApp)
	})
	t.Run("测试链路数据", func(t *testing.T) {
		testSpansEndpoint(t, fiberApp)
	})
	t.Run("测试链路传播", func(t *testing.T) {
		testTracePropagation(t, fiberApp)
	})
}

func TestFiberTracingHTTP(t *testing.T) {
	_, app := newTestApp(t, "test-fiber-tracing-http",
		boot.WithProperty("fiber.port", "18082"),
		boot.WithProperty(httpTestRoutesEnabled, "true"),
	)
	defer app.Stop()

	// 业务路由由 httpTestRoutesConfig 在监听启动前注册，监听后再注册路由会返回 404
	_ = startHTTPServer(t, app, "http://localhost:18082/api/hello")

	t.Run("HTTP 正常请求", func(t *testing.T) {
		assertGetOKWithTraceID(t, "http://localhost:18082/api/hello")
	})
	t.Run("HTTP 查看链路数据", func(t *testing.T) {
		assertGetNonEmptyBody(t, "http://localhost:18082/api/spans")
	})
	t.Run("HTTP 监控端点-健康检查", func(t *testing.T) {
		assertHealthEndpointTracing(t, "http://localhost:18082/actuator/health")
	})
	t.Run("HTTP 监控端点-指标", func(t *testing.T) {
		assertMetricsEndpointTracing(t, "http://localhost:18082/actuator/metrics")
	})
}

func TestFiberActuatorEndpointsWithTracing(t *testing.T) {
	fiberApp, app, tracer := newTestServer(t, "test-fiber-actuator-tracing", boot.WithProperty("actuator.path", "/actuator"))
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
			assertMonitoringEndpoint(t, fiberApp, tracer, ep)
		})
	}
}

func TestFiberActuatorEndpointsTracingContextPropagation(t *testing.T) {
	fiberApp := fiber.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-fiber-actuator-tracing-propagation"),
		boot.WithProperty("fiber.enabled", "true"),
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
	if err := ctx.Container().RegisterInstance(fiberApp, reflect.TypeFor[*fiber.App]()); err != nil {
		t.Fatalf("注册 Fiber App 失败: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	t.Run("监控端点链路传播", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
		req.Header.Set("X-Trace-ID", "test-trace-actuator-123")
		req.Header.Set("X-Span-ID", "test-span-actuator-456")
		resp, err := fiberApp.Test(req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
		}

		traceID := resp.Header.Get("X-Trace-ID")
		if traceID == "" {
			t.Error("响应头中缺少 X-Trace-ID")
		}

		spanID := resp.Header.Get("X-Span-ID")
		if spanID == "" {
			t.Error("响应头中缺少 X-Span-ID")
		}

		if traceID != "test-trace-actuator-123" {
			t.Errorf("期望 TraceID 传播 'test-trace-actuator-123'，实际 '%s'", traceID)
		}
	})
}

// newTestServer 构建并同步启动一个测试用 Fiber 应用，返回 Fiber App、Boot 与 Tracer。
func newTestServer(t *testing.T, appName string, extra ...boot.BootOption) (*fiber.App, *boot.Boot, *tracing.Tracer) {
	fiberApp := fiber.New()

	opts := []boot.BootOption{
		boot.WithAppName(appName),
		boot.WithProperty("fiber.enabled", "true"),
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
	if err := ctx.Container().RegisterInstance(fiberApp, reflect.TypeFor[*fiber.App]()); err != nil {
		t.Fatalf("注册 Fiber App 失败: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}
	return fiberApp, app, tracer
}

// newTestApp 创建测试用 Fiber 应用并注册到容器中，但不启动应用。
// Fiber 不支持在 Listen 之后注册路由，调用方可先注册路由再执行 startHTTPServer。
func newTestApp(t *testing.T, appName string, extra ...boot.BootOption) (*fiber.App, *boot.Boot) {
	fiberApp := fiber.New()

	opts := []boot.BootOption{
		boot.WithAppName(appName),
		boot.WithProperty("fiber.enabled", "true"),
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
	if err := ctx.Container().RegisterInstance(fiberApp, reflect.TypeFor[*fiber.App]()); err != nil {
		t.Fatalf("注册 Fiber App 失败: %v", err)
	}
	return fiberApp, app
}

// startHTTPServer 以异步方式启动 Fiber 应用并等待服务就绪，返回 Tracer。
func startHTTPServer(t *testing.T, app *boot.Boot, url string) *tracing.Tracer {
	go app.Start()

	if err := waitForServer(url, 10, 500*time.Millisecond); err != nil {
		t.Fatalf("Server did not start in time: %v", err)
	}

	tracer, err := core.GetByName[*tracing.Tracer](app.Context().Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}
	return tracer
}

// testHelloEndpoint 验证 /api/hello 返回 200 状态码。
func testHelloEndpoint(t *testing.T, app *fiber.App) {
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, resp.StatusCode)
	}
}

// testErrorEndpoint 验证 /api/error 返回 500 状态码。
func testErrorEndpoint(t *testing.T, app *fiber.App) {
	req := httptest.NewRequest(http.MethodGet, "/api/error", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// testSpansEndpoint 验证 /api/spans 返回链路追踪数据。
func testSpansEndpoint(t *testing.T, app *fiber.App) {
	req := httptest.NewRequest(http.MethodGet, "/api/spans", nil)
	resp, err := app.Test(req)
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

// testTracePropagation 验证 X-Trace-ID / X-Span-ID 请求头在响应中回传。
func testTracePropagation(t *testing.T, app *fiber.App) {
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	req.Header.Set("X-Trace-ID", "test-trace-123")
	req.Header.Set("X-Span-ID", "test-span-456")
	resp, err := app.Test(req)
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

	spanID := resp.Header.Get("X-Span-ID")
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
func assertMonitoringEndpoint(t *testing.T, app *fiber.App, tracer *tracing.Tracer, ep monitoringEndpoint) {
	spansBefore := len(tracer.GetSpans())

	req := httptest.NewRequest(http.MethodGet, ep.path, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
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
