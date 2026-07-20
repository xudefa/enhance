package actuator

import (
	"fmt"
	"net/http"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

// ActuatorHttpStarter Actuator HTTP 启动器
//
// 负责将 Actuator 端点自动挂载到现有的 HTTP 服务器上。
// 通过 HttpEndpointRegistry 接口实现框架无关的端点注册,
// 支持任意 HTTP 框架(Gin、Fiber、Echo、Chi 等)。
//
// 挂载策略(按优先级):
// 1. HttpEndpointRegistry 接口(推荐,框架无关)
// 2. HttpHandlerRegistry 接口(简化版)
// 3. RouteRegistrar 接口(向后兼容)
// 4. 独立 HTTP 服务器(降级方案)
//
// 通过配置项控制各端点的暴露:
//   - actuator.expose.health: 健康检查端点(默认 true)
//   - actuator.expose.metrics: 指标端点(默认 true)
//   - actuator.expose.env: 环境信息端点(默认 true)
//   - actuator.expose.beans: Bean 列表端点(默认 true)
//   - actuator.expose.info: 应用信息端点(默认 true)
//   - actuator.expose.prometheus: Prometheus 端点(默认 true)
type ActuatorHttpStarter struct {
	actuator *Actuator
	basePath string
}

// Name 返回启动器名称
func (s *ActuatorHttpStarter) Name() string {
	return "actuator-http"
}

// Dependencies 返回依赖的其他启动器名称
func (s *ActuatorHttpStarter) Dependencies() []string {
	return []string{}
}

// GetCondition 返回启动条件
func (s *ActuatorHttpStarter) GetCondition() condition.Condition {
	return condition.OnProperty(ActuatorEnabled, ConditionTrue)
}

// Configure 配置阶段:从容器中获取 Actuator 实例
func (s *ActuatorHttpStarter) Configure(ctx boot.ApplicationContext) error {
	bean, err := core.GetByName[*Actuator](ctx.Container(), "")
	if err != nil {
		return nil
	}
	s.actuator = bean

	// 从配置读取 basePath,默认为 /actuator
	s.basePath = ctx.Environment().GetString("actuator.path", "/actuator")
	return nil
}

// Start 启动阶段:将 Actuator 端点挂载到 HTTP 服务器
func (s *ActuatorHttpStarter) Start(ctx boot.ApplicationContext) error {
	if s.actuator == nil {
		return nil
	}

	env := ctx.Environment()

	// 构建端点配置列表
	endpoints := s.buildEndpointConfigs(env)
	if len(endpoints) == 0 {
		return nil
	}

	// 策略 1: 尝试通过 HttpEndpointRegistry 接口注册(推荐,框架无关)
	if s.registerViaEndpointRegistry(ctx, endpoints) {
		return nil
	}

	// 策略 2: 尝试通过 HttpHandlerRegistry 接口注册(简化版)
	if s.registerViaHandlerRegistry(ctx, endpoints) {
		return nil
	}

	// 策略 3: 尝试通过 RouteRegistrar 接口注册(向后兼容)
	if s.registerViaRouteRegistrar(ctx, endpoints) {
		return nil
	}

	// 策略 4: 降级方案 - 启动独立的 HTTP 服务器
	s.startStandaloneServer(env, endpoints)
	return nil
}

// Stop 停止阶段:无需特殊处理
func (s *ActuatorHttpStarter) Stop(ctx boot.ApplicationContext) error {
	return nil
}

// buildEndpointConfigs 构建端点配置列表
func (s *ActuatorHttpStarter) buildEndpointConfigs(env *environment.Environment) []EndpointConfig {
	exposeHealth := env.GetBool("actuator.expose.health", true)
	exposeMetrics := env.GetBool("actuator.expose.metrics", true)
	exposeEnv := env.GetBool("actuator.expose.env", true)
	exposeBeans := env.GetBool("actuator.expose.beans", true)
	exposeInfo := env.GetBool("actuator.expose.info", true)
	exposePrometheus := env.GetBool("actuator.expose.prometheus", true)

	endpoints := make([]EndpointConfig, 0, 6)

	if exposeHealth {
		endpoints = append(endpoints, EndpointConfig{
			Method:      http.MethodGet,
			Path:        JoinPath(s.basePath, "/health"),
			Handler:     http.HandlerFunc(s.actuator.HealthHandler),
			Description: "Health Check",
		})
	}

	if exposeMetrics {
		endpoints = append(endpoints, EndpointConfig{
			Method:      http.MethodGet,
			Path:        JoinPath(s.basePath, "/metrics"),
			Handler:     http.HandlerFunc(s.actuator.MetricsHandler),
			Description: "Application Metrics",
		})
	}

	if exposeEnv {
		endpoints = append(endpoints, EndpointConfig{
			Method:      http.MethodGet,
			Path:        JoinPath(s.basePath, "/env"),
			Handler:     http.HandlerFunc(s.actuator.EnvHandler),
			Description: "Environment Information",
		})
	}

	if exposeBeans {
		endpoints = append(endpoints, EndpointConfig{
			Method:      http.MethodGet,
			Path:        JoinPath(s.basePath, "/beans"),
			Handler:     http.HandlerFunc(s.actuator.BeansHandler),
			Description: "Spring Beans",
		})
	}

	if exposeInfo {
		endpoints = append(endpoints, EndpointConfig{
			Method:      http.MethodGet,
			Path:        JoinPath(s.basePath, "/info"),
			Handler:     http.HandlerFunc(s.actuator.InfoHandler),
			Description: "Application Info",
		})
	}

	if exposePrometheus {
		endpoints = append(endpoints, EndpointConfig{
			Method:      http.MethodGet,
			Path:        "/metrics",
			Handler:     http.HandlerFunc(s.actuator.PrometheusHandler),
			Description: "Prometheus Metrics",
		})
	}

	return endpoints
}

// registerViaEndpointRegistry 通过 HttpEndpointRegistry 接口注册端点
//
// 这是推荐的注册方式,框架无关,支持任意 HTTP 框架。
// Web 框架应在启动时向容器注册 HttpEndpointRegistry 实现。
func (s *ActuatorHttpStarter) registerViaEndpointRegistry(ctx boot.ApplicationContext, endpoints []EndpointConfig) bool {
	registry, err := core.GetByName[HttpEndpointRegistry](ctx.Container(), "")
	if err != nil {
		return false
	}

	if registry == nil {
		return false
	}

	registry.RegisterEndpoints(endpoints)
	return true
}

// registerViaHandlerRegistry 通过 HttpHandlerRegistry 接口注册端点
//
// 这是简化版的注册方式,适用于只需要基本路由注册功能的场景。
func (s *ActuatorHttpStarter) registerViaHandlerRegistry(ctx boot.ApplicationContext, endpoints []EndpointConfig) bool {
	registry, err := core.GetByName[HttpHandlerRegistry](ctx.Container(), "")
	if err != nil {
		return false
	}

	if registry == nil {
		return false
	}

	for _, ep := range endpoints {
		// 对于标准 http.ServeMux,需要包含 method 前缀
		pattern := ep.Path
		if ep.Method != "" {
			pattern = ep.Method + " " + ep.Path
		}
		registry.Handle(pattern, ep.Handler)
	}

	return true
}

// registerViaRouteRegistrar 通过 RouteRegistrar 接口注册端点(向后兼容)
//
// 这是旧版的注册方式,保留以兼容使用 RouteRegistrar 的现有代码。
func (s *ActuatorHttpStarter) registerViaRouteRegistrar(ctx boot.ApplicationContext, endpoints []EndpointConfig) bool {
	registrar, err := core.GetByName[RouteRegistrar](ctx.Container(), "")
	if err != nil {
		return false
	}

	if registrar == nil {
		return false
	}

	for _, ep := range endpoints {
		registrar.Handle(ep.Path, ep.Handler)
	}

	return true
}

// startStandaloneServer 启动独立的 HTTP 服务器(降级方案)
//
// 当无法找到任何已注册的 HTTP 服务器时,使用此方法启动独立的服务器。
// 这应该作为最后的降级方案,因为会增加额外的端口和服务器实例。
func (s *ActuatorHttpStarter) startStandaloneServer(env *environment.Environment, endpoints []EndpointConfig) {
	mux := http.NewServeMux()

	for _, ep := range endpoints {
		pattern := ep.Path
		if ep.Method != "" {
			pattern = ep.Method + " " + ep.Path
		}
		mux.Handle(pattern, ep.Handler)
	}

	// 从配置读取端口,默认 8081
	port := env.GetString("actuator.port", "8081")
	host := env.GetString("actuator.host", "0.0.0.0")
	addr := fmt.Sprintf("%s:%s", host, port)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			fmt.Printf("[Actuator] Failed to start standalone server on %s: %v\n", addr, err)
		}
	}()

	fmt.Printf("[Actuator] Started standalone server on %s (no HTTP server found in container)\n", addr)
}

// ==================== 框架特定的 HttpEndpointRegistry 实现 ====================
// 注意: 框架特定的实现已移至各自的 starter 包中:
//   - Gin: starter/gin/endpoint_registry.go
//   - Fiber: starter/fiber/endpoint_registry.go
//   - 默认 Router: web/mvc/starter.go (handlerBasedEndpointRegistry)
//
// 这样可以避免 actuator 包对具体框架的依赖,保持清晰的依赖关系。
//
// 如果您需要为其他框架(如 Echo、Chi 等)集成 Actuator,
// 请在相应的 starter 包中实现 HttpEndpointRegistry 接口,
// 并在框架的 autoconfig 中注册到容器。

func init() {
	boot.RegisterStarter(&ActuatorHttpStarter{})
}
