// Package core 提供 Web 框架核心接口定义。
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
)

// contextInterfaceType 用于检测方法参数是否为 Context 接口类型
var contextInterfaceType = reflect.TypeOf((*Context)(nil)).Elem()

// RouteInfo 路由信息。
type RouteInfo struct {
	// Method HTTP 方法
	Method string
	// Path 路由路径
	Path string
	// HandlerFunc 处理函数
	HandlerFunc reflect.Method
	// Controller 控制器实例
	Controller any
	// Consumes 请求内容类型
	Consumes string
	// Produces 响应内容类型
	Produces string
	// StructName 结构体名称
	StructName string
	// MethodName 方法名称
	MethodName string
}

// RouteRegistry 路由注册表。
//
// 管理控制器的路由信息,支持在扫描完成后统一注册到 HTTP 服务器。
type RouteRegistry struct {
	mu              sync.RWMutex
	controllers     map[string]string
	routes          []RouteInfo
	controllerCache map[string]any
}

// NewRouteRegistry 创建路由注册表。
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		controllers:     make(map[string]string),
		controllerCache: make(map[string]any),
	}
}

// RegisterController 注册控制器。
func (r *RouteRegistry) RegisterController(structName string, controller any, basePath string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.controllers[structName] = basePath
	r.controllerCache[structName] = controller
}

// GetControllerBasePath 获取控制器基础路径。
func (r *RouteRegistry) GetControllerBasePath(structName string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.controllers[structName]
}

// RegisterRoute 注册路由。
func (r *RouteRegistry) RegisterRoute(route RouteInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, route)
}

// GetRoutes 获取所有路由。
func (r *RouteRegistry) GetRoutes() []RouteInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]RouteInfo, len(r.routes))
	copy(routes, r.routes)

	for i := range routes {
		if routes[i].Controller == nil && routes[i].StructName != "" {
			if cached, ok := r.controllerCache[routes[i].StructName]; ok {
				routes[i].Controller = cached
			}
		}
	}

	return routes
}

// RegisterToMux 注册路由到 http.ServeMux。
func (r *RouteRegistry) RegisterToMux(mux *http.ServeMux) error {
	routes := r.GetRoutes()

	seen := make(map[string]string)
	for _, route := range routes {
		key := route.Method + " " + route.Path
		if prevKey, ok := seen[key]; ok {
			return fmt.Errorf("duplicate route %s: %s and %s",
				key, prevKey, route.Method+" "+route.Path)
		}
		seen[key] = route.Method
	}

	for _, route := range routes {
		if route.Controller == nil {
			return fmt.Errorf("controller not found for route %s %s (struct: %s, method: %s)",
				route.Method, route.Path, route.StructName, route.MethodName)
		}

		handler, err := r.createHandler(route)
		if err != nil {
			return fmt.Errorf("failed to create handler for route %s %s: %w",
				route.Method, route.Path, err)
		}

		// 使用方法限定的模式注册，允许同一路径下存在多个 HTTP 方法
		mux.Handle(route.Method+" "+route.Path, handler)
	}

	return nil
}

// createHandler 创建 HTTP 处理器适配器。
func (r *RouteRegistry) createHandler(route RouteInfo) (http.Handler, error) {
	controllerVal := reflect.ValueOf(route.Controller)
	method := controllerVal.MethodByName(route.MethodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found on controller %T",
			route.MethodName, route.Controller)
	}

	methodType := method.Type()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != route.Method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if route.Produces != "" {
			w.Header().Set("Content-Type", route.Produces)
		}

		args := make([]reflect.Value, 0, methodType.NumIn())

		for i := 0; i < methodType.NumIn(); i++ {
			paramType := methodType.In(i)
			if paramType.Implements(contextInterfaceType) {
				args = append(args, reflect.ValueOf(newSimpleContext(w, req)))
			} else {
				args = append(args, reflect.Zero(paramType))
			}
		}

		results := method.Call(args)
		if len(results) > 0 {
			lastResult := results[len(results)-1]
			if lastResult.Type().Implements(reflect.TypeFor[error]()) {
				if err, _ := lastResult.Interface().(error); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if len(results) == 1 || (len(results) == 2 && lastResult.Type().Implements(reflect.TypeFor[error]())) {
				result := results[0]
				switch result.Kind() {
				case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
					if result.IsNil() {
						return
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(result.Interface())
			}
		}
	}), nil
}

// simpleContext 是 core.Context 的简单实现，基于标准库 http.ResponseWriter 和 *http.Request。
type simpleContext struct {
	w          http.ResponseWriter
	req        *http.Request
	statusCode int
	aborted    bool
	ctx        context.Context
}

func newSimpleContext(w http.ResponseWriter, req *http.Request) *simpleContext {
	return &simpleContext{
		w:   w,
		req: req,
		ctx: req.Context(),
	}
}

func (c *simpleContext) RequestMethod() string        { return c.req.Method }
func (c *simpleContext) RequestURI() string           { return c.req.URL.RequestURI() }
func (c *simpleContext) PathParam(name string) string { return "" }
func (c *simpleContext) Query(name string) string     { return c.req.URL.Query().Get(name) }
func (c *simpleContext) Header(key string) string     { return c.req.Header.Get(key) }
func (c *simpleContext) Next()                        {}
func (c *simpleContext) IsAborted() bool              { return c.aborted }
func (c *simpleContext) Context() context.Context     { return c.ctx }
func (c *simpleContext) Request() *http.Request       { return c.req }

func (c *simpleContext) SetContext(ctx context.Context) { c.ctx = ctx }
func (c *simpleContext) SetStatusCode(code int)         { c.statusCode = code }
func (c *simpleContext) SetHeader(key, value string)    { c.w.Header().Set(key, value) }

func (c *simpleContext) QueryDefault(name, defaultVal string) string {
	if v := c.req.URL.Query().Get(name); v != "" {
		return v
	}
	return defaultVal
}

func (c *simpleContext) BindJSON(target any) error {
	c.req.Body = http.MaxBytesReader(nil, c.req.Body, 32<<20)
	body, err := io.ReadAll(c.req.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return fmt.Errorf("request body is empty")
	}
	c.req.Body = io.NopCloser(bytes.NewReader(body))
	return json.Unmarshal(body, target)
}

func (c *simpleContext) JSON(code int, data any) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	return json.NewEncoder(c.w).Encode(data)
}

func (c *simpleContext) String(code int, format string, args ...any) {
	c.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.w.WriteHeader(code)
	_, _ = fmt.Fprintf(c.w, format, args...)
}

func (c *simpleContext) AbortWithStatus(code int) {
	c.aborted = true
	c.statusCode = code
	if code == http.StatusNoContent || code < 200 {
		c.w.WriteHeader(code)
		return
	}
	http.Error(c.w, http.StatusText(code), code)
}

func (c *simpleContext) AbortWithStatusJSON(code int, body any) {
	c.aborted = true
	c.statusCode = code
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	_ = json.NewEncoder(c.w).Encode(body)
}
