package chi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xudefa/enhance/actuator"
)

// ChiEndpointRegistry Chi 框架的 HttpEndpointRegistry 实现
//
// 该实现将 Actuator 的标准 http.Handler 适配为 Chi 的 HandlerFunc,
// 并注册到 Chi Router 中。
type ChiEndpointRegistry struct {
	router    *chi.Mux
	endpoints map[string]bool
}

// NewChiEndpointRegistry 创建 Chi 端点注册表
func NewChiEndpointRegistry(router *chi.Mux) *ChiEndpointRegistry {
	return &ChiEndpointRegistry{
		router:    router,
		endpoints: make(map[string]bool),
	}
}

// RegisterEndpoint 注册单个端点
func (r *ChiEndpointRegistry) RegisterEndpoint(method, path string, handler http.Handler) {
	if r.router == nil || handler == nil {
		return
	}

	wrapper := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req)
	})

	if method != "" {
		r.router.MethodFunc(method, path, wrapper)
	} else {
		r.router.Handle(path, wrapper)
	}

	r.endpoints[path] = true
}

// RegisterEndpoints 批量注册端点
func (r *ChiEndpointRegistry) RegisterEndpoints(endpoints []actuator.EndpointConfig) {
	for _, ep := range endpoints {
		r.RegisterEndpoint(ep.Method, ep.Path, ep.Handler)
	}
}

// HasEndpoint 检查是否已注册指定路径的端点
func (r *ChiEndpointRegistry) HasEndpoint(path string) bool {
	_, exists := r.endpoints[path]
	return exists
}

// 确保 ChiEndpointRegistry 实现了 actuator.HttpEndpointRegistry 接口
var _ actuator.HttpEndpointRegistry = (*ChiEndpointRegistry)(nil)
