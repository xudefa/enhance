package actuator

import "net/http"

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
