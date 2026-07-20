package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	_ "github.com/xudefa/enhance/starter/gin" // 触发 Gin 自动配置注册
	"github.com/xudefa/enhance/tracing"
)

func TestGinTracingIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-gin-tracing"),
		boot.WithProperty("gin.enabled", "true"),
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
	if err := ctx.Container().RegisterInstance(engine, reflect.TypeFor[*gin.Engine]()); err != nil {
		t.Fatalf("注册 Gin Engine 失败: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}

	engine.GET("/api/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	engine.GET("/api/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "test error"})
	})

	engine.GET("/api/spans", func(c *gin.Context) {
		spans := tracer.GetSpans()
		c.JSON(200, gin.H{
			"total_spans": len(spans),
			"spans":       spans,
		})
	})

	t.Run("测试正常请求", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("期望状态码 200，实际 %d", w.Code)
		}

		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("解码响应失败: %v", err)
		}
		if resp["message"] != "ok" {
			t.Errorf("期望响应 message='ok'，实际 '%s'", resp["message"])
		}
	})

	t.Run("测试错误请求", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/error", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != 500 {
			t.Errorf("期望状态码 500，实际 %d", w.Code)
		}
	})

	t.Run("测试链路数据", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/spans", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("期望状态码 200，实际 %d", w.Code)
		}

		body, _ := io.ReadAll(w.Body)
		if len(body) == 0 {
			t.Error("期望返回链路数据，实际为空")
		}
	})

	t.Run("测试链路传播", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
		req.Header.Set("X-Trace-ID", "test-trace-123")
		req.Header.Set("X-Span-ID", "test-span-456")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("期望状态码 200，实际 %d", w.Code)
		}

		traceID := w.Header().Get("X-Trace-ID")
		if traceID == "" {
			t.Error("响应头中缺少 X-Trace-ID")
		}

		spanID := w.Header().Get("X-Span-ID")
		if spanID == "" {
			t.Error("响应头中缺少 X-Span-ID")
		}
	})
}

func TestGinTracingHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-gin-tracing-http"),
		boot.WithProperty("gin.enabled", "true"),
		boot.WithProperty("gin.port", "18081"),
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
	if err := ctx.Container().RegisterInstance(engine, reflect.TypeFor[*gin.Engine]()); err != nil {
		t.Fatalf("注册 Gin Engine 失败: %v", err)
	}

	go app.Start()

	if err := waitForServer("http://localhost:18081/api/hello", 10, 500*time.Millisecond); err != nil {
		t.Fatalf("Server did not start in time: %v", err)
	}

	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		t.Fatalf("未找到 Tracer: %v", err)
	}

	engine.GET("/api/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	engine.GET("/api/spans", func(c *gin.Context) {
		spans := tracer.GetSpans()
		c.JSON(200, gin.H{
			"total_spans": len(spans),
			"spans":       spans,
		})
	})

	t.Run("HTTP 正常请求", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18081/api/hello")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
		}

		traceID := resp.Header.Get("X-Trace-ID")
		if traceID == "" {
			t.Error("响应头中缺少 X-Trace-ID")
		}
	})

	t.Run("HTTP 查看链路数据", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18081/api/spans")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) == 0 {
			t.Error("期望返回链路数据，实际为空")
		}
	})

	t.Run("HTTP 监控端点-健康检查", func(t *testing.T) {
		resp, err := http.Get("http://localhost:18081/actuator/health")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
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
		resp, err := http.Get("http://localhost:18081/actuator/metrics")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("期望状态码 200，实际 %d", resp.StatusCode)
		}

		traceID := resp.Header.Get("X-Trace-ID")
		if traceID == "" {
			t.Error("监控端点响应头中缺少 X-Trace-ID")
		}

		t.Logf("监控端点 /actuator/metrics 链路追踪: TraceID=%s", traceID)
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

func TestActuatorEndpointsWithTracing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-actuator-tracing"),
		boot.WithProperty("gin.enabled", "true"),
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
	if err := ctx.Container().RegisterInstance(engine, reflect.TypeFor[*gin.Engine]()); err != nil {
		t.Fatalf("注册 Gin Engine 失败: %v", err)
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
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("期望状态码 200，实际 %d", w.Code)
			}

			spansAfter := len(tracer.GetSpans())
			if spansAfter <= spansBefore {
				t.Errorf("期望产生新的链路追踪日志，实际未产生: spans before=%d, after=%d", spansBefore, spansAfter)
			}

			spans := tracer.GetSpans()
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

func TestActuatorEndpointsTracingContextPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()

	app, err := boot.NewApplication(
		boot.WithAppName("test-actuator-tracing-propagation"),
		boot.WithProperty("gin.enabled", "true"),
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
	if err := ctx.Container().RegisterInstance(engine, reflect.TypeFor[*gin.Engine]()); err != nil {
		t.Fatalf("注册 Gin Engine 失败: %v", err)
	}

	if err := app.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	t.Run("监控端点链路传播", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
		req.Header.Set("X-Trace-ID", "test-trace-actuator-123")
		req.Header.Set("X-Span-ID", "test-span-actuator-456")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("期望状态码 200，实际 %d", w.Code)
		}

		traceID := w.Header().Get("X-Trace-ID")
		if traceID == "" {
			t.Error("响应头中缺少 X-Trace-ID")
		}

		spanID := w.Header().Get("X-Span-ID")
		if spanID == "" {
			t.Error("响应头中缺少 X-Span-ID")
		}

		if traceID != "test-trace-actuator-123" {
			t.Errorf("期望 TraceID 传播 'test-trace-actuator-123'，实际 '%s'", traceID)
		}
	})
}
