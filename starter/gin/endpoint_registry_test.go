package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/enhance/actuator"
)

func TestGinEndpointRegistry_RegisterEndpoint(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := NewGinEndpointRegistry(engine)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	registry.RegisterEndpoint(http.MethodGet, "/actuator/health", handler)

	// 验证端点已注册
	if !registry.HasEndpoint("/actuator/health") {
		t.Error("Expected endpoint to be registered")
	}

	// 验证端点可以正常访问
	req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", rec.Body.String())
	}
}

func TestGinEndpointRegistry_RegisterEndpoints(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := NewGinEndpointRegistry(engine)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	endpoints := []actuator.EndpointConfig{
		{
			Method:  http.MethodGet,
			Path:    "/actuator/health",
			Handler: handler,
		},
		{
			Method:  http.MethodGet,
			Path:    "/actuator/metrics",
			Handler: handler,
		},
		{
			Method:  http.MethodGet,
			Path:    "/actuator/env",
			Handler: handler,
		},
	}

	registry.RegisterEndpoints(endpoints)

	// 验证所有端点都已注册
	if !registry.HasEndpoint("/actuator/health") {
		t.Error("Expected /actuator/health to be registered")
	}
	if !registry.HasEndpoint("/actuator/metrics") {
		t.Error("Expected /actuator/metrics to be registered")
	}
	if !registry.HasEndpoint("/actuator/env") {
		t.Error("Expected /actuator/env to be registered")
	}

	// 验证端点可以正常访问
	req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGinEndpointRegistry_RegisterEndpoint_Any(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := NewGinEndpointRegistry(engine)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// 注册空 method 的端点(应该注册为 Any)
	registry.RegisterEndpoint("", "/actuator/info", handler)

	if !registry.HasEndpoint("/actuator/info") {
		t.Error("Expected endpoint to be registered")
	}

	// 验证 GET 请求
	req := httptest.NewRequest(http.MethodGet, "/actuator/info", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// 验证 POST 请求也应该能访问
	req = httptest.NewRequest(http.MethodPost, "/actuator/info", nil)
	rec = httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d for POST, got %d", http.StatusOK, rec.Code)
	}
}

func TestGinEndpointRegistry_NilHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := NewGinEndpointRegistry(engine)

	// 不应该 panic
	registry.RegisterEndpoint(http.MethodGet, "/test", nil)

	if registry.HasEndpoint("/test") {
		t.Error("Expected endpoint not to be registered with nil handler")
	}
}

func TestGinEndpointRegistry_NilEngine(t *testing.T) {
	t.Parallel()

	registry := &GinEndpointRegistry{
		engine:    nil,
		endpoints: make(map[string]bool),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 不应该 panic
	registry.RegisterEndpoint(http.MethodGet, "/test", handler)

	if registry.HasEndpoint("/test") {
		t.Error("Expected endpoint not to be registered with nil engine")
	}
}

func TestGinEndpointRegistry_Interface(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := NewGinEndpointRegistry(engine)

	// 验证实现了 actuator.HttpEndpointRegistry 接口
	var _ actuator.HttpEndpointRegistry = registry
}
