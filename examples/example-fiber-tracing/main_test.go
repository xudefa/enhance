package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	_ "github.com/xudefa/enhance/starter/fiber" // 触发 Fiber 自动配置注册
	"github.com/xudefa/enhance/tracing"
)

func TestFiberTracingIntegration(t *testing.T) {
	fiberApp := fiber.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-fiber-tracing"),
		boot.WithProperty("fiber.enabled", "true"),
		boot.WithProperty("tracing.enabled", "true"),
		boot.WithProperty("tracing.service_name", "test-service"),
		boot.WithProperty("tracing.sampling_rate", "1.0"),
		boot.WithProperty("actuator.enabled", "true"),
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

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}

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
		req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
		resp, err := fiberApp.Test(req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("测试错误请求", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/error", nil)
		resp, err := fiberApp.Test(req)
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("期望状态码 %d，实际 %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("测试链路数据", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/spans", nil)
		resp, err := fiberApp.Test(req)
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
	})

	t.Run("测试链路传播", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
		req.Header.Set("X-Trace-ID", "test-trace-123")
		req.Header.Set("X-Span-ID", "test-span-456")
		resp, err := fiberApp.Test(req)
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
	})
}

func TestFiberTracingHTTP(t *testing.T) {
	fiberApp := fiber.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-fiber-tracing-http"),
		boot.WithProperty("fiber.enabled", "true"),
		boot.WithProperty("fiber.port", "18082"),
		boot.WithProperty("tracing.enabled", "true"),
		boot.WithProperty("tracing.service_name", "test-service"),
		boot.WithProperty("tracing.sampling_rate", "1.0"),
		boot.WithProperty("actuator.enabled", "true"),
	)
	if err != nil {
		t.Fatalf("创建应用失败: %v", err)
	}
	defer app.Stop()

	ctx := app.Context()
	if err := ctx.Container().RegisterInstance(fiberApp, reflect.TypeFor[*fiber.App]()); err != nil {
		t.Fatalf("注册 Fiber App 失败: %v", err)
	}

	go app.Start()

	if err := waitForServer("http://localhost:18082/api/hello", 10, 500*time.Millisecond); err != nil {
		t.Fatalf("Server did not start in time: %v", err)
	}

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
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

	t.Run("HTTP 正常请求", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18082/api/hello")
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
	})

	t.Run("HTTP 查看链路数据", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18082/api/spans")
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
	})

	t.Run("HTTP 监控端点-健康检查", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18082/actuator/health")
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
	})

	t.Run("HTTP 监控端点-指标", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18082/actuator/metrics")
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
	})
}

func TestFiberActuatorEndpointsWithTracing(t *testing.T) {
	fiberApp := fiber.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-fiber-actuator-tracing"),
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

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}

	monitoringEndpoints := []struct {
		name string
		path string
	}{
		{"健康检查端点", "/actuator/health"},
		{"指标端点", "/actuator/metrics"},
		{"环境信息端点", "/actuator/env"},
		{"Bean列表端点", "/actuator/beans"},
		{"应用信息端点", "/actuator/info"},
	}

	for _, ep := range monitoringEndpoints {
		t.Run("监控端点_"+ep.name, func(t *testing.T) {
			spansBefore := len(tracer.GetSpans())

			req := httptest.NewRequest(http.MethodGet, ep.path, nil)
			resp, err := fiberApp.Test(req)
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
