package fiber

import (
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/xudefa/enhance/actuator"
)

// FiberEndpointRegistry Fiber 框架的 HttpEndpointRegistry 实现
//
// 该实现将 Actuator 的标准 http.Handler 适配为 Fiber 的 Handler,
// 并注册到 Fiber App 中。
type FiberEndpointRegistry struct {
	app       *fiber.App
	endpoints map[string]bool
}

// NewFiberEndpointRegistry 创建 Fiber 端点注册表
func NewFiberEndpointRegistry(app *fiber.App) *FiberEndpointRegistry {
	return &FiberEndpointRegistry{
		app:       app,
		endpoints: make(map[string]bool),
	}
}

// RegisterEndpoint 注册单个端点
func (r *FiberEndpointRegistry) RegisterEndpoint(method, path string, handler http.Handler) {
	if r.app == nil || handler == nil {
		return
	}

	wrapper := func(c *fiber.Ctx) error {
		// 创建适配的 ResponseWriter
		rw := &fiberResponseWriter{
			ctx:    c,
			header: make(http.Header),
		}

		// 创建标准 http.Request
		req := &http.Request{
			Method: c.Method(),
			URL: &url.URL{
				Path: c.Path(),
			},
			Header: rw.header,
			Body:   nil,
		}

		// 调用标准 HTTP handler
		handler.ServeHTTP(rw, req)
		return nil
	}

	if method != "" {
		r.app.Add(method, path, wrapper)
	} else {
		// 注册所有常用方法
		r.app.Get(path, wrapper)
		r.app.Post(path, wrapper)
		r.app.Put(path, wrapper)
		r.app.Delete(path, wrapper)
		r.app.Patch(path, wrapper)
	}

	r.endpoints[path] = true
}

// RegisterEndpoints 批量注册端点
func (r *FiberEndpointRegistry) RegisterEndpoints(endpoints []actuator.EndpointConfig) {
	for _, ep := range endpoints {
		r.RegisterEndpoint(ep.Method, ep.Path, ep.Handler)
	}
}

// HasEndpoint 检查是否已注册指定路径的端点
func (r *FiberEndpointRegistry) HasEndpoint(path string) bool {
	_, exists := r.endpoints[path]
	return exists
}

// fiberResponseWriter 实现 http.ResponseWriter 接口以桥接 Fiber 和标准 http 包
type fiberResponseWriter struct {
	ctx    *fiber.Ctx
	header http.Header
	status int
}

func (w *fiberResponseWriter) Header() http.Header {
	return w.header
}

func (w *fiberResponseWriter) Write(b []byte) (int, error) {
	if err := w.ctx.Send(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ctx.Status(statusCode)
}

// 确保 fiberResponseWriter 实现了 http.ResponseWriter 接口
var _ http.ResponseWriter = (*fiberResponseWriter)(nil)

// 确保 FiberEndpointRegistry 实现了 actuator.HttpEndpointRegistry 接口
var _ actuator.HttpEndpointRegistry = (*FiberEndpointRegistry)(nil)
