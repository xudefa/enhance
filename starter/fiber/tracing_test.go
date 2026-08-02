package fiber

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/xudefa/enhance/tracing"
)

func TestTracingMiddleware_WithTracer(t *testing.T) {
	tracer := tracing.NewTracer(
		tracing.WithServiceName("test-service"),
	)

	app := fiber.New()
	app.Use(TracingMiddleware(tracer))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("期望响应体 'ok'，实际 '%s'", string(body))
	}

	spans := tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("期望 1 个 Span，实际 %d 个", len(spans))
	}
	if spans[0].Name != "GET /test" {
		t.Errorf("期望 Span 名称 'GET /test'，实际 '%s'", spans[0].Name)
	}
	if spans[0].Tags["http.status_code"] != "200" {
		t.Errorf("期望状态码标签 '200'，实际 '%s'", spans[0].Tags["http.status_code"])
	}
	if spans[0].Status != tracing.StatusOK {
		t.Errorf("期望状态 OK，实际 %s", spans[0].Status)
	}
}

func TestTracingMiddleware_WithoutTracer(t *testing.T) {
	app := fiber.New()
	app.Use(TracingMiddleware(nil))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("期望响应体 'ok'，实际 '%s'", string(body))
	}
}

func TestTracingMiddleware_WithError(t *testing.T) {
	tracer := tracing.NewTracer(
		tracing.WithServiceName("test-service"),
	)

	app := fiber.New()
	app.Use(TracingMiddleware(tracer))
	app.Get("/error", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != 500 {
		t.Errorf("期望状态码 500，实际 %d", resp.StatusCode)
	}

	spans := tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("期望 1 个 Span，实际 %d 个", len(spans))
	}
	if spans[0].Name != "GET /error" {
		t.Errorf("期望 Span 名称 'GET /error'，实际 '%s'", spans[0].Name)
	}
	if spans[0].Tags["http.status_code"] != "500" {
		t.Errorf("期望状态码标签 '500'，实际 '%s'", spans[0].Tags["http.status_code"])
	}
	if spans[0].Status != tracing.StatusError {
		t.Errorf("期望状态 ERROR，实际 %s", spans[0].Status)
	}
}

func TestTracingMiddleware_ContextPropagation(t *testing.T) {
	tracer := tracing.NewTracer(
		tracing.WithServiceName("test-service"),
	)

	app := fiber.New()
	app.Use(TracingMiddleware(tracer))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-ID", "test-trace-id-123")
	req.Header.Set("X-Span-ID", "test-span-id-456")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
	}

	spans := tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("期望 1 个 Span，实际 %d 个", len(spans))
	}
	if spans[0].TraceID != tracing.TraceID("test-trace-id-123") {
		t.Errorf("期望 TraceID 'test-trace-id-123'，实际 '%s'", spans[0].TraceID)
	}
	if spans[0].Status != tracing.StatusOK {
		t.Errorf("期望状态 OK，实际 %s", spans[0].Status)
	}

	if resp.Header.Get("X-Trace-ID") == "" {
		t.Error("响应头中缺少 X-Trace-ID")
	}
	if resp.Header.Get("X-Span-ID") == "" {
		t.Error("响应头中缺少 X-Span-ID")
	}
}
