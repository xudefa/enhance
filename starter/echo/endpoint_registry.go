package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xudefa/enhance/actuator"
)

// EchoEndpointRegistry Echo 框架的 HttpEndpointRegistry 实现
//
// 该实现将 Actuator 的标准 http.Handler 适配为 Echo 的 HandlerFunc,
// 并注册到 Echo 引擎中。
type EchoEndpointRegistry struct {
	engine    *echo.Echo
	endpoints map[string]bool
}

// NewEchoEndpointRegistry 创建 Echo 端点注册表
func NewEchoEndpointRegistry(engine *echo.Echo) *EchoEndpointRegistry {
	return &EchoEndpointRegistry{
		engine:    engine,
		endpoints: make(map[string]bool),
	}
}

// RegisterEndpoint 注册单个端点
func (r *EchoEndpointRegistry) RegisterEndpoint(method, path string, handler http.Handler) {
	if r.engine == nil || handler == nil {
		return
	}

	wrapper := func(c echo.Context) error {
		handler.ServeHTTP(c.Response(), c.Request())
		return nil
	}

	if method != "" {
		r.engine.Add(method, path, wrapper)
	} else {
		r.engine.Any(path, wrapper)
	}

	r.endpoints[path] = true
}

// RegisterEndpoints 批量注册端点
func (r *EchoEndpointRegistry) RegisterEndpoints(endpoints []actuator.EndpointConfig) {
	for _, ep := range endpoints {
		r.RegisterEndpoint(ep.Method, ep.Path, ep.Handler)
	}
}

// HasEndpoint 检查是否已注册指定路径的端点
func (r *EchoEndpointRegistry) HasEndpoint(path string) bool {
	_, exists := r.endpoints[path]
	return exists
}

// 确保 EchoEndpointRegistry 实现了 actuator.HttpEndpointRegistry 接口
var _ actuator.HttpEndpointRegistry = (*EchoEndpointRegistry)(nil)
