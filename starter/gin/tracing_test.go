package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/enhance/tracing"
)

func TestTracingMiddleware_WithTracer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracer := tracing.NewTracer(
		tracing.WithServiceName("test-service"),
	)

	router := gin.New()
	router.Use(TracingMiddleware(tracer))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("期望响应体 'ok'，实际 '%s'", w.Body.String())
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
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TracingMiddleware(nil))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("期望响应体 'ok'，实际 '%s'", w.Body.String())
	}
}

func TestTracingMiddleware_WithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracer := tracing.NewTracer(
		tracing.WithServiceName("test-service"),
	)

	router := gin.New()
	router.Use(TracingMiddleware(tracer))
	router.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusInternalServerError, w.Code)
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
	gin.SetMode(gin.TestMode)

	tracer := tracing.NewTracer(
		tracing.WithServiceName("test-service"),
	)

	router := gin.New()
	router.Use(TracingMiddleware(tracer))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Trace-ID", "test-trace-id-123")
	req.Header.Set("X-Span-ID", "test-span-id-456")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d", http.StatusOK, w.Code)
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

	respHeaders := w.Header()
	if respHeaders.Get("X-Trace-ID") == "" {
		t.Error("响应头中缺少 X-Trace-ID")
	}
	if respHeaders.Get("X-Span-ID") == "" {
		t.Error("响应头中缺少 X-Span-ID")
	}
}
