package engine

import (
	"context"
	"net/http"

	"github.com/xudefa/enhance/web/core"
)

// ServerAdapter 将不同框架的服务器适配为统一的 Server 接口。
type ServerAdapter struct {
	startFunc  func() error
	stopFunc   func(ctx context.Context) error
	setHandler func(handler http.Handler)
	useFunc    func(middleware func(http.Handler) http.Handler)
}

// NewServerAdapter 创建服务器适配器。
func NewServerAdapter(
	start func() error,
	stop func(ctx context.Context) error,
	setHandler func(handler http.Handler),
	use func(middleware func(http.Handler) http.Handler),
) *ServerAdapter {
	return &ServerAdapter{
		startFunc:  start,
		stopFunc:   stop,
		setHandler: setHandler,
		useFunc:    use,
	}
}

// Start 启动服务器。
func (a *ServerAdapter) Start() error {
	return a.startFunc()
}

// Stop 停止服务器。
func (a *ServerAdapter) Stop(ctx context.Context) error {
	return a.stopFunc(ctx)
}

// SetHandler 设置处理器。
func (a *ServerAdapter) SetHandler(handler http.Handler) {
	a.setHandler(handler)
}

// Use 注册中间件。
func (a *ServerAdapter) Use(middleware func(http.Handler) http.Handler) {
	a.useFunc(middleware)
}

// RouterAdapter 将不同框架的路由器适配为统一的 Router 接口。
type RouterAdapter struct {
	getFunc    func(path string, handler core.HandlerFunc)
	postFunc   func(path string, handler core.HandlerFunc)
	putFunc    func(path string, handler core.HandlerFunc)
	deleteFunc func(path string, handler core.HandlerFunc)
	patchFunc  func(path string, handler core.HandlerFunc)
	handleFunc func(method, path string, handler core.HandlerFunc)
	groupFunc  func(prefix string) core.Router
	useFunc    func(middleware core.MiddlewareFunc)
}

// RouterAdapterConfig 路由器适配器的配置项。
type RouterAdapterConfig struct {
	Get    func(path string, handler core.HandlerFunc)
	Post   func(path string, handler core.HandlerFunc)
	Put    func(path string, handler core.HandlerFunc)
	Delete func(path string, handler core.HandlerFunc)
	Patch  func(path string, handler core.HandlerFunc)
	Handle func(method, path string, handler core.HandlerFunc)
	Group  func(prefix string) core.Router
	Use    func(middleware core.MiddlewareFunc)
}

// NewRouterAdapter 创建路由器适配器。
func NewRouterAdapter(config RouterAdapterConfig) *RouterAdapter {
	return &RouterAdapter{
		getFunc:    config.Get,
		postFunc:   config.Post,
		putFunc:    config.Put,
		deleteFunc: config.Delete,
		patchFunc:  config.Patch,
		handleFunc: config.Handle,
		groupFunc:  config.Group,
		useFunc:    config.Use,
	}
}

// GET 注册 GET 路由。
func (a *RouterAdapter) GET(path string, handler core.HandlerFunc) {
	a.getFunc(path, handler)
}

// POST 注册 POST 路由。
func (a *RouterAdapter) POST(path string, handler core.HandlerFunc) {
	a.postFunc(path, handler)
}

// PUT 注册 PUT 路由。
func (a *RouterAdapter) PUT(path string, handler core.HandlerFunc) {
	a.putFunc(path, handler)
}

// DELETE 注册 DELETE 路由。
func (a *RouterAdapter) DELETE(path string, handler core.HandlerFunc) {
	a.deleteFunc(path, handler)
}

// PATCH 注册 PATCH 路由。
func (a *RouterAdapter) PATCH(path string, handler core.HandlerFunc) {
	a.patchFunc(path, handler)
}

// Handle 注册任意 HTTP 方法的路由。
func (a *RouterAdapter) Handle(method, path string, handler core.HandlerFunc) {
	if a.handleFunc != nil {
		a.handleFunc(method, path, handler)
	}
}

// Group 创建路由组。
func (a *RouterAdapter) Group(prefix string) core.Router {
	return a.groupFunc(prefix)
}

// Use 注册中间件。
func (a *RouterAdapter) Use(middleware core.MiddlewareFunc) {
	a.useFunc(middleware)
}
