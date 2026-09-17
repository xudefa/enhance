package actuator

import (
	"net/http"
	"testing"
	"time"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

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

	registered := starter.registerViaEndpointRegistry(ctx, endpoints)
	if registered {
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

	registered := starter.registerViaHandlerRegistry(ctx, endpoints)
	if registered {
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

	registered := starter.registerViaRouteRegistrar(ctx, endpoints)
	if registered {
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
