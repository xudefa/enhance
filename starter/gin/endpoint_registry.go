package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/enhance/actuator"
)

// GinEndpointRegistry Gin 框架的 HttpEndpointRegistry 实现
//
// 该实现将 Actuator 的标准 http.Handler 适配为 Gin 的 HandlerFunc,
// 并注册到 Gin 引擎中。
type GinEndpointRegistry struct {
	engine    *gin.Engine
	endpoints map[string]bool
}

// NewGinEndpointRegistry 创建 Gin 端点注册表
func NewGinEndpointRegistry(engine *gin.Engine) *GinEndpointRegistry {
	return &GinEndpointRegistry{
		engine:    engine,
		endpoints: make(map[string]bool),
	}
}

// RegisterEndpoint 注册单个端点
func (r *GinEndpointRegistry) RegisterEndpoint(method, path string, handler http.Handler) {
	if r.engine == nil || handler == nil {
		return
	}

	wrapper := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}

	// 使用 engine.Group().Handle() 确保路由经过全局中间件
	// Gin 的 engine.Use() 添加的全局中间件会影响所有通过 engine.Group() 注册的路由
	group := r.engine.Group("/")
	if method == "" {
		// 空 method 注册为 Any（支持所有 HTTP 方法）
		group.Any(path, wrapper)
	} else {
		group.Handle(method, path, wrapper)
	}

	r.endpoints[path] = true
}

// RegisterEndpoints 批量注册端点
func (r *GinEndpointRegistry) RegisterEndpoints(endpoints []actuator.EndpointConfig) {
	for _, ep := range endpoints {
		r.RegisterEndpoint(ep.Method, ep.Path, ep.Handler)
	}
}

// HasEndpoint 检查是否已注册指定路径的端点
func (r *GinEndpointRegistry) HasEndpoint(path string) bool {
	_, exists := r.endpoints[path]
	return exists
}
