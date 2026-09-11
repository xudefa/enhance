package actuator

import "net/http"

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

// 实现类已按"一实现一文件"原则拆分到独立文件中：
//   - http_endpoint_registry_adapter.go: HttpEndpointRegistryAdapter
//   - std_http_handler_registry.go: StdHttpHandlerRegistry
//   - path_normalizer.go: PathNormalizer + 辅助函数
