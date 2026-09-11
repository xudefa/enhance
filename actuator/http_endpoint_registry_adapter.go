package actuator

import (
	"net/http"
	"sync"
)

// HttpEndpointRegistryAdapter HttpEndpointRegistry 的基础实现
//
// 该适配器将 HttpEndpointRegistry 接口委托给底层的 HttpHandlerRegistry,
// 为不同 HTTP 框架提供统一的注册方式。
//
// 框架集成者可以实现 HttpHandlerRegistry 接口,然后使用此适配器
// 快速获得 HttpEndpointRegistry 的完整功能。
type HttpEndpointRegistryAdapter struct {
	registry  HttpHandlerRegistry
	mu        sync.RWMutex
	endpoints map[string]bool
}

// NewHttpEndpointRegistryAdapter 创建 HttpEndpointRegistry 适配器
func NewHttpEndpointRegistryAdapter(registry HttpHandlerRegistry) *HttpEndpointRegistryAdapter {
	return &HttpEndpointRegistryAdapter{
		registry:  registry,
		endpoints: make(map[string]bool),
	}
}

// RegisterEndpoint 注册单个端点
func (a *HttpEndpointRegistryAdapter) RegisterEndpoint(method, path string, handler http.Handler) {
	if a.registry == nil {
		return
	}

	// 如果指定了 method,注册为 method:path 格式
	if method != "" {
		a.registry.Handle(method+" "+path, handler)
	} else {
		a.registry.Handle(path, handler)
	}

	a.mu.Lock()
	a.endpoints[path] = true
	a.mu.Unlock()
}

// RegisterEndpoints 批量注册端点
func (a *HttpEndpointRegistryAdapter) RegisterEndpoints(endpoints []EndpointConfig) {
	for _, ep := range endpoints {
		a.RegisterEndpoint(ep.Method, ep.Path, ep.Handler)
	}
}

// HasEndpoint 检查是否已注册指定路径的端点
func (a *HttpEndpointRegistryAdapter) HasEndpoint(path string) bool {
	a.mu.RLock()
	_, exists := a.endpoints[path]
	a.mu.RUnlock()
	return exists
}
