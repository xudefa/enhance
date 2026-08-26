package actuator

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
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

func TestActuatorHttpStarter_Name(t *testing.T) {
	t.Parallel()
	starter := &ActuatorHttpStarter{}
	if starter.Name() != "actuator-http" {
		t.Errorf("Expected name 'actuator-http', got %s", starter.Name())
	}
}

func TestActuatorHttpStarter_Dependencies(t *testing.T) {
	t.Parallel()
	starter := &ActuatorHttpStarter{}
	deps := starter.Dependencies()
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestActuatorHttpStarter_GetCondition(t *testing.T) {
	t.Parallel()
	starter := &ActuatorHttpStarter{}
	cond := starter.GetCondition()
	if cond == nil {
		t.Error("Expected non-nil condition")
	}
}

func TestActuatorHttpStarter_Stop(t *testing.T) {
	t.Parallel()
	starter := &ActuatorHttpStarter{}

	// 测试没有standaloneSrv的情况
	err := starter.Stop(nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestActuatorHttpStarter_Configure_NilActuator(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{}
	ctx := &mockActuatorContext{}

	err := starter.Configure(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestActuatorHttpStarter_Start_NilActuator(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{}
	ctx := &mockActuatorContext{}

	err := starter.Start(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestActuatorHttpStarter_Start_NoEndpoints(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{
		actuator: &Actuator{},
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

	ctx := &mockActuatorContext{
		env: env,
	}

	err := starter.Start(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// mockActuatorContext 模拟ApplicationContext
type mockActuatorContext struct {
	env       *environment.Environment
	container core.Container
	ctx       context.Context
}

func (m *mockActuatorContext) Environment() *environment.Environment {
	if m.env == nil {
		return environment.NewEnvironment()
	}
	return m.env
}

func (m *mockActuatorContext) Container() core.Container {
	if m.container == nil {
		return core.NewContainer()
	}
	return m.container
}

func (m *mockActuatorContext) Context() context.Context {
	if m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func (m *mockActuatorContext) EventBus() boot.EventBusResult {
	return nil
}

func (m *mockActuatorContext) GetByType(t reflect.Type) (any, error) {
	return nil, nil
}

func (m *mockActuatorContext) Register(t reflect.Type, opts ...core.BeanOption) error {
	return nil
}

func TestActuatorHttpStarter_Configure_NoActuator(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	ctx := &mockActuatorContext{
		container: container,
	}

	starter := &ActuatorHttpStarter{}
	err := starter.Configure(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// actuator应该为nil因为容器中没有Actuator bean
	if starter.actuator != nil {
		t.Error("Expected actuator to be nil")
	}
}

func TestActuatorHttpStarter_Configure_CustomPath(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"actuator.path": "/monitor",
	}))

	ctx := &mockActuatorContext{
		container: container,
		env:       env,
	}

	starter := &ActuatorHttpStarter{basePath: "/actuator"} // 设置默认值
	err := starter.Configure(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 注意：Configure在没有actuator时会提前返回，不会修改basePath
	// 这个测试验证Configure在没有actuator时不会报错
}

func TestActuatorHttpStarter_Start_NilActuator_Extended(t *testing.T) {
	t.Parallel()

	ctx := &mockActuatorContext{}
	starter := &ActuatorHttpStarter{}

	err := starter.Start(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestActuatorHttpStarter_Stop_NoStandaloneServer_Extended(t *testing.T) {
	t.Parallel()

	ctx := &mockActuatorContext{}
	starter := &ActuatorHttpStarter{}

	err := starter.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestActuatorHttpStarter_Name_Extended(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{}
	if starter.Name() != "actuator-http" {
		t.Errorf("Expected name 'actuator-http', got '%s'", starter.Name())
	}
}

func TestActuatorHttpStarter_Dependencies_Extended(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{}
	deps := starter.Dependencies()
	if len(deps) != 0 {
		t.Errorf("Expected no dependencies, got %d", len(deps))
	}
}

func TestActuatorHttpStarter_GetCondition_Extended(t *testing.T) {
	t.Parallel()

	starter := &ActuatorHttpStarter{}
	cond := starter.GetCondition()
	if cond == nil {
		t.Error("Expected non-nil condition")
	}
}

func TestActuatorHttpStarter_RegisterViaEndpointRegistry(t *testing.T) {
	t.Parallel()

	// 创建端点配置
	endpoints := []EndpointConfig{
		{Path: "/test", Method: "GET", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
	}

	// 测试没有注册器时应该返回false
	starter := &ActuatorHttpStarter{}
	container := core.NewContainer()
	ctx := &mockActuatorContext{container: container}

	result := starter.registerViaEndpointRegistry(ctx, endpoints)
	if result {
		t.Error("expected false when no HttpEndpointRegistry is registered")
	}
}

func TestActuatorHttpStarter_RegisterViaHandlerRegistry(t *testing.T) {
	t.Parallel()

	// 创建端点配置
	endpoints := []EndpointConfig{
		{Path: "/test", Method: "GET", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
	}

	// 测试没有注册器时应该返回false
	starter := &ActuatorHttpStarter{}
	container := core.NewContainer()
	ctx := &mockActuatorContext{container: container}

	result := starter.registerViaHandlerRegistry(ctx, endpoints)
	if result {
		t.Error("expected false when no HttpHandlerRegistry is registered")
	}
}

func TestActuatorHttpStarter_RegisterViaRouteRegistrar(t *testing.T) {
	t.Parallel()

	// 创建端点配置
	endpoints := []EndpointConfig{
		{Path: "/test", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
	}

	// 测试没有注册器时应该返回false
	starter := &ActuatorHttpStarter{}
	container := core.NewContainer()
	ctx := &mockActuatorContext{container: container}

	result := starter.registerViaRouteRegistrar(ctx, endpoints)
	if result {
		t.Error("expected false when no RouteRegistrar is registered")
	}
}

func TestActuatorHttpStarter_StartStandaloneServer(t *testing.T) {
	t.Parallel()

	// 创建端点配置
	endpoints := []EndpointConfig{
		{Path: "/health", Method: "GET", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"UP"}`))
		})},
	}

	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewDefaultPropertySource("test", map[string]any{
		"actuator.port": "18081", // 使用非标准端口避免冲突
		"actuator.host": "127.0.0.1",
	}))

	starter := &ActuatorHttpStarter{}

	// 启动独立服务器
	starter.startStandaloneServer(env, endpoints)

	// 验证服务器已启动
	if starter.standaloneSrv == nil {
		t.Error("expected standalone server to be created")
	}

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 测试服务器是否响应
	resp, err := http.Get("http://127.0.0.1:18081/health")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	}

	// 停止服务器
	starter.Stop(nil)
}
