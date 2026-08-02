package actuator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestHttpEndpointRegistryAdapter(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	registry := NewHttpEndpointRegistryAdapter(&StdHttpHandlerRegistry{Mux: mux})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	endpoints := []EndpointConfig{
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
	}

	registry.RegisterEndpoints(endpoints)

	// 验证路由已注册
	req := httptest.NewRequest(http.MethodGet, "/actuator/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// 验证 HasEndpoint
	if !registry.HasEndpoint("/actuator/health") {
		t.Error("Expected endpoint to be registered")
	}

	if registry.HasEndpoint("/actuator/nonexistent") {
		t.Error("Expected endpoint to not be registered")
	}
}

func TestHttpEndpointRegistryAdapter_Concurrent(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	registry := NewHttpEndpointRegistryAdapter(&StdHttpHandlerRegistry{Mux: mux})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	const workers = 32
	const pathsPerWorker = 100

	var wg sync.WaitGroup
	wg.Add(workers * 2)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < pathsPerWorker; j++ {
				path := fmt.Sprintf("/actuator/worker-%d/%d", id, j)
				registry.RegisterEndpoint(http.MethodGet, path, handler)
			}
		}(i)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < pathsPerWorker; j++ {
				path := fmt.Sprintf("/actuator/worker-%d/%d", id, j)
				_ = registry.HasEndpoint(path)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < workers; i++ {
		path := fmt.Sprintf("/actuator/worker-%d/%d", i, 0)
		if !registry.HasEndpoint(path) {
			t.Errorf("expected endpoint %s to be registered", path)
		}
	}
}

func TestPathNormalizer(t *testing.T) {
	t.Parallel()

	normalizer := PathNormalizer{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "/"},
		{"simple", "/health", "/health"},
		{"no leading slash", "health", "/health"},
		{"trailing slash", "/health/", "/health"},
		{"double slash", "/actuator//health", "/actuator/health"},
		{"root", "/", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		base     string
		path     string
		expected string
	}{
		{"simple", "/actuator", "/health", "/actuator/health"},
		{"base with trailing slash", "/actuator/", "/health", "/actuator/health"},
		{"path without leading slash", "/actuator", "health", "/actuator/health"},
		{"both with slashes", "/actuator/", "/health", "/actuator/health"},
		{"root base", "/", "/health", "/health"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinPath(tt.base, tt.path)
			if result != tt.expected {
				t.Errorf("JoinPath(%q, %q) = %q, want %q", tt.base, tt.path, result, tt.expected)
			}
		})
	}
}

func TestBuildEndpointConfigs(t *testing.T) {
	t.Parallel()

	// 创建测试用的 ActuatorHttpStarter
	starter := &ActuatorHttpStarter{
		actuator: &Actuator{},
		basePath: "/actuator",
	}

	env := environment.NewEnvironment()

	// 测试默认配置(全部暴露)
	endpoints := starter.buildEndpointConfigs(env)
	if len(endpoints) != 6 {
		t.Errorf("Expected 6 endpoints, got %d", len(endpoints))
	}

	// 验证端点路径
	expectedPaths := map[string]bool{
		"/actuator/health":  false,
		"/actuator/metrics": false,
		"/actuator/env":     false,
		"/actuator/beans":   false,
		"/actuator/info":    false,
		"/metrics":          false,
	}

	for _, ep := range endpoints {
		if _, ok := expectedPaths[ep.Path]; ok {
			expectedPaths[ep.Path] = true
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected endpoint %s not found", path)
		}
	}
}

func TestBuildEndpointConfigs_Disabled(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{
		actuator: &Actuator{},
		basePath: "/actuator",
	}

	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test-config", map[string]any{
		"actuator.expose.health":     false,
		"actuator.expose.metrics":    false,
		"actuator.expose.env":        false,
		"actuator.expose.beans":      false,
		"actuator.expose.info":       false,
		"actuator.expose.prometheus": false,
	}))

	endpoints := starter.buildEndpointConfigs(env)
	if len(endpoints) != 0 {
		t.Errorf("Expected 0 endpoints when all disabled, got %d", len(endpoints))
	}
}

func TestBuildEndpointConfigs_CustomBasePath(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{
		actuator: &Actuator{},
		basePath: "/monitor",
	}

	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test-config", map[string]any{
		"actuator.expose.metrics":    false,
		"actuator.expose.env":        false,
		"actuator.expose.beans":      false,
		"actuator.expose.info":       false,
		"actuator.expose.prometheus": false,
	}))

	endpoints := starter.buildEndpointConfigs(env)
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpoints))
	}

	if endpoints[0].Path != "/monitor/health" {
		t.Errorf("Expected path /monitor/health, got %s", endpoints[0].Path)
	}
}
