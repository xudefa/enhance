package actuator

import (
	"net/http"
	"strings"
	"sync"
)

// HttpEndpointRegistry HTTP 端点注册表接口
//
// 该接口作为 Web 框架和 Actuator 之间的桥梁,允许 Actuator 将端点
// 挂载到任意 HTTP 框架,而无需关心框架的具体实现细节。
//
// Web 框架(如 Gin、Fiber、默认 Router 等)应在启动时向容器注册
// 此接口的实现,Actuator 通过查找此接口来自动挂载端点。
//
// 使用示例(Gin 框架):
//
//	registry := &GinEndpointRegistry{engine: ginEngine}
//	ctx.Container().RegisterInstance(registry, reflect.TypeFor[actuator.HttpEndpointRegistry]())
//
// 使用示例(Fiber 框架):
//
//	registry := &FiberEndpointRegistry{app: fiberApp}
//	ctx.Container().RegisterInstance(registry, reflect.TypeFor[actuator.HttpEndpointRegistry]())
type HttpEndpointRegistry interface {
	// RegisterEndpoint 注册单个端点
	// method: HTTP 方法(GET, POST 等),空字符串表示所有方法
	// path: 路由路径
	// handler: HTTP 处理器
	RegisterEndpoint(method, path string, handler http.Handler)

	// RegisterEndpoints 批量注册端点
	// endpoints: 端点配置列表
	RegisterEndpoints(endpoints []EndpointConfig)

	// HasEndpoint 检查是否已注册指定路径的端点
	HasEndpoint(path string) bool
}

// EndpointConfig 端点配置
type EndpointConfig struct {
	// Method HTTP 方法,空字符串表示所有方法
	Method string

	// Path 路由路径
	Path string

	// Handler HTTP 处理器
	Handler http.Handler

	// Description 端点描述(可选,用于日志和文档)
	Description string
}

// HttpHandlerRegistry HTTP Handler 注册表
//
// 这是 HttpEndpointRegistry 的简化版本,仅支持注册 http.Handler。
// 适用于只需要基本路由注册功能的场景。
type HttpHandlerRegistry interface {
	// Handle 注册路由处理器
	// pattern: 路由模式,如 "/actuator/health"
	// handler: HTTP 处理器
	Handle(pattern string, handler http.Handler)
}

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

// StdHttpHandlerRegistry 标准库 http.Handler 注册表实现
//
// 该实现包装 http.ServeMux 或其他实现了 Handle 方法的类型,
// 提供 HttpHandlerRegistry 接口的功能。
type StdHttpHandlerRegistry struct {
	Mux interface {
		Handle(pattern string, handler http.Handler)
	}
}

// Handle 注册路由处理器
func (r *StdHttpHandlerRegistry) Handle(pattern string, handler http.Handler) {
	if r.Mux != nil {
		r.Mux.Handle(pattern, handler)
	}
}

// PathNormalizer 路径标准化工具
type PathNormalizer struct{}

// NormalizePath 标准化路径,确保路径格式正确
func (PathNormalizer) NormalizePath(path string) string {
	if path == "" {
		return "/"
	}

	// 确保路径以 / 开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 移除末尾的 / (除非路径就是 /)
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// 将连续的 // 替换为单个 /
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	return path
}

// EnsureLeadingSlash 确保路径以 / 开头
func EnsureLeadingSlash(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// JoinPath 拼接路径
func JoinPath(base, path string) string {
	base = strings.TrimSuffix(base, "/")
	path = EnsureLeadingSlash(path)
	return base + path
}
