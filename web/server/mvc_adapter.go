package server

import (
	"context"
	"net/http"

	"github.com/xudefa/enhance/web/mvc"
)

// HttpServerAdapter 将 web/server.HttpServer 适配到 mvc.Server 接口
// 这样可以让 HttpServer 与 mvc.WebStarter 配合使用
type HttpServerAdapter struct {
	server *HttpServer
}

// NewHttpServerAdapter 创建一个新的 HTTP 服务器适配器
func NewHttpServerAdapter(opts ...ServerOption) *HttpServerAdapter {
	return &HttpServerAdapter{
		server: NewHTTPServer(opts...),
	}
}

// Start 启动 HTTP 服务器
func (a *HttpServerAdapter) Start() error {
	return a.server.Start()
}

// Stop 优雅地停止服务器
func (a *HttpServerAdapter) Stop(ctx context.Context) error {
	return a.server.Stop(ctx)
}

// SetHandler 设置自定义的 HTTP 处理器
func (a *HttpServerAdapter) SetHandler(handler any) {
	// mvc.Server 接口的 SetHandler 接受 any 类型
	// HttpServer.SetHandler 现在接受 http.Handler 类型
	switch h := handler.(type) {
	case http.Handler:
		// 直接传递 http.Handler（DefaultRouter 实现了 http.Handler）
		a.server.SetHandler(h)
	default:
		_ = h
	}
}

// Use 向服务器注册一个中间件
func (a *HttpServerAdapter) Use(m any) {
	// mvc.Server 接口的 Use 接受 any 类型
	// HttpServer.Use 接受 func(http.Handler) http.Handler 类型
	switch mw := m.(type) {
	case func(http.Handler) http.Handler:
		a.server.Use(mw)
	case mvc.MiddlewareFunc:
		// 将 mvc.MiddlewareFunc 转换为 func(http.Handler) http.Handler
		// 使用已有的 DefaultContext 实现
		adapter := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// 创建 DefaultContext（使用已有的实现）
				ctx := NewContext(w, r).WithMiddleware([]mvc.MiddlewareFunc{mw}, func(c mvc.Context) {
					// 中间件执行完后，调用下一个 http.Handler
					next.ServeHTTP(w, r)
				})
				// 执行中间件链
				ctx.Next()
			})
		}
		a.server.Use(adapter)
	default:
		_ = mw
	}
}

// GetServer 获取底层的 HttpServer 实例
func (a *HttpServerAdapter) GetServer() *HttpServer {
	return a.server
}
