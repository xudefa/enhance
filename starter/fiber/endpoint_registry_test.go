package fiber

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/xudefa/enhance/actuator"
)

func TestFiberEndpointRegistry_RegisterEndpoint(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	registry := NewFiberEndpointRegistry(app)

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
	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/actuator/health",
		},
		Header: make(http.Header),
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestFiberEndpointRegistry_RegisterEndpoints(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	registry := NewFiberEndpointRegistry(app)

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
	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/actuator/health",
		},
		Header: make(http.Header),
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestFiberEndpointRegistry_RegisterEndpoint_Any(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	registry := NewFiberEndpointRegistry(app)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// 注册空 method 的端点(应该注册所有常用方法)
	registry.RegisterEndpoint("", "/actuator/info", handler)

	if !registry.HasEndpoint("/actuator/info") {
		t.Error("Expected endpoint to be registered")
	}

	// 验证 GET 请求
	getReq := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/actuator/info",
		},
		Header: make(http.Header),
	}
	resp, err := app.Test(getReq)
	if err != nil {
		t.Fatalf("Failed to test GET request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d for GET, got %d", http.StatusOK, resp.StatusCode)
	}

	// 验证 POST 请求
	postReq := &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Path: "/actuator/info",
		},
		Header: make(http.Header),
	}
	resp, err = app.Test(postReq)
	if err != nil {
		t.Fatalf("Failed to test POST request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d for POST, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestFiberEndpointRegistry_NilHandler(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	registry := NewFiberEndpointRegistry(app)

	// 不应该 panic
	registry.RegisterEndpoint(http.MethodGet, "/test", nil)

	if registry.HasEndpoint("/test") {
		t.Error("Expected endpoint not to be registered with nil handler")
	}
}

func TestFiberEndpointRegistry_NilApp(t *testing.T) {
	t.Parallel()

	registry := &FiberEndpointRegistry{
		app:       nil,
		endpoints: make(map[string]bool),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 不应该 panic
	registry.RegisterEndpoint(http.MethodGet, "/test", handler)

	if registry.HasEndpoint("/test") {
		t.Error("Expected endpoint not to be registered with nil app")
	}
}

func TestFiberEndpointRegistry_Interface(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	registry := NewFiberEndpointRegistry(app)

	// 验证实现了 actuator.HttpEndpointRegistry 接口
	var _ actuator.HttpEndpointRegistry = registry
}
