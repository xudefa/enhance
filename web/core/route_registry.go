// Package core 提供 Web 框架核心接口定义。
package core

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

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

		mux.Handle(route.Path, handler)
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

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != route.Method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if route.Produces != "" {
			w.Header().Set("Content-Type", route.Produces)
		}

		args := []reflect.Value{controllerVal}
		method.Call(args)
	}), nil
}
