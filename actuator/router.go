package actuator

import "net/http"

// StdRouteRegistrar 标准库 HTTP 路由注册器
//
// 该实现包装 http.ServeMux，提供 RouteRegistrar 接口的功能。
type StdRouteRegistrar struct {
	Mux interface {
		Handle(pattern string, handler http.Handler)
	}
}

// Handle 注册路由处理器
func (r *StdRouteRegistrar) Handle(pattern string, handler http.Handler) {
	if r.Mux != nil {
		r.Mux.Handle(pattern, handler)
	}
}

// RouteConfig 路由配置
type RouteConfig struct {
	BasePath    string
	ExposeDebug bool
	Prefix      string
}

// DefaultRouteConfig 返回默认路由配置
func DefaultRouteConfig() RouteConfig {
	return RouteConfig{
		BasePath:    "/actuator",
		ExposeDebug: false,
		Prefix:      "",
	}
}
