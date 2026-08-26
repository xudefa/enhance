package server

import (
	"context"
	"net/http"
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
func (a *HttpServerAdapter) SetHandler(handler http.Handler) {
	a.server.SetHandler(handler)
}

// Use 向服务器注册一个中间件
func (a *HttpServerAdapter) Use(middleware func(http.Handler) http.Handler) {
	a.server.Use(middleware)
}

// GetServer 获取底层的 HttpServer 实例
func (a *HttpServerAdapter) GetServer() *HttpServer {
	return a.server
}
