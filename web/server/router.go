package server

import (
	"net/http"
	"strings"
	"sync"

	"github.com/xudefa/enhance/web/mvc"
)

// DefaultRouter 默认路由器实现
type DefaultRouter struct {
	mux         *http.ServeMux
	middlewares []mvc.MiddlewareFunc
	handlers    map[string]mvc.HandlerFunc
	prefix      string
	mu          sync.RWMutex
	// 性能优化：缓存路由模式，避免每次请求都遍历
	routePatterns []routePattern
}

// routePattern 预编译的路由模式
type routePattern struct {
	method     string
	parts      []string
	paramNames []string // 参数名称列表
	paramIdxs  []int    // 参数在路径中的索引
	handler    mvc.HandlerFunc
	hasParams  bool
}

// NewRouter 创建新的路由器
func NewRouter() *DefaultRouter {
	return &DefaultRouter{
		mux:      http.NewServeMux(),
		handlers: make(map[string]mvc.HandlerFunc),
	}
}

// GET 注册 GET 路由
func (r *DefaultRouter) GET(path string, handler mvc.HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

// POST 注册 POST 路由
func (r *DefaultRouter) POST(path string, handler mvc.HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

// PUT 注册 PUT 路由
func (r *DefaultRouter) PUT(path string, handler mvc.HandlerFunc) {
	r.handle(http.MethodPut, path, handler)
}

// DELETE 注册 DELETE 路由
func (r *DefaultRouter) DELETE(path string, handler mvc.HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

// PATCH 注册 PATCH 路由
func (r *DefaultRouter) PATCH(path string, handler mvc.HandlerFunc) {
	r.handle(http.MethodPatch, path, handler)
}

// Group 创建路由组
func (r *DefaultRouter) Group(prefix string) mvc.Router {
	r.mu.Lock()
	defer r.mu.Unlock()

	return &DefaultRouter{
		mux:         r.mux,
		middlewares: append([]mvc.MiddlewareFunc{}, r.middlewares...),
		handlers:    r.handlers,
		prefix:      r.prefix + prefix,
	}
}

// Use 注册中间件
func (r *DefaultRouter) Use(middleware mvc.MiddlewareFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middleware)
}

// ServeHTTP 实现 http.Handler 接口
func (r *DefaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 查找匹配的路由
	path := req.URL.Path
	if r.prefix != "" {
		path = strings.TrimPrefix(path, r.prefix)
		if path == "" {
			path = "/"
		}
	}

	// 构建完整路径用于查找
	fullPath := req.Method + " " + path
	if r.prefix != "" {
		fullPath = req.Method + " " + r.prefix + path
	}

	handler, ok := r.handlers[fullPath]
	var params map[string]string
	if !ok {
		// 尝试查找带路径参数的路由
		handler, params, ok = r.findHandlerWithParams(req.Method, path)
		if !ok {
			http.NotFound(w, req)
			return
		}
	}

	// 创建上下文
	ctx := NewContext(w, req)

	// 设置路径参数
	if params != nil {
		ctx.WithParams(params)
	}

	// 构建中间件链
	allMiddleware := append([]mvc.MiddlewareFunc{}, r.middlewares...)

	// 执行中间件链和处理器
	ctx.WithMiddleware(allMiddleware, handler)
	ctx.Next()
}

// handle 注册路由
func (r *DefaultRouter) handle(method, path string, handler mvc.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPath := r.prefix + path
	key := method + " " + fullPath
	r.handlers[key] = handler

	// 性能优化：预编译路由模式
	pattern := r.compileRoutePattern(method, fullPath, handler)
	r.routePatterns = append(r.routePatterns, pattern)

	// 注意：不再注册到标准 mux，因为 Go 1.25 不允许重复注册
	// 所有请求都通过 ServeHTTP 处理
}

// compileRoutePattern 预编译路由模式
func (r *DefaultRouter) compileRoutePattern(method, path string, handler mvc.HandlerFunc) routePattern {
	parts := strings.Split(path, "/")
	paramNames := make([]string, 0, len(parts))
	paramIdxs := make([]int, 0, len(parts))
	hasParams := false

	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramNames = append(paramNames, part[1:len(part)-1])
			paramIdxs = append(paramIdxs, i)
			parts[i] = "" // 标记为参数
			hasParams = true
		}
	}

	return routePattern{
		method:     method,
		parts:      parts,
		paramNames: paramNames,
		paramIdxs:  paramIdxs,
		handler:    handler,
		hasParams:  hasParams,
	}
}

// findHandlerWithParams 查找带路径参数的路由（使用预编译模式）
func (r *DefaultRouter) findHandlerWithParams(method, path string) (mvc.HandlerFunc, map[string]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pathParts := strings.Split(path, "/")

	// 使用预编译的路由模式进行匹配，避免每次都解析
	for i := range r.routePatterns {
		pattern := &r.routePatterns[i]
		if pattern.method != method {
			continue
		}
		if !pattern.hasParams {
			continue
		}
		if len(pattern.parts) != len(pathParts) {
			continue
		}

		// 快速路径匹配
		matched := true
		for j, part := range pattern.parts {
			if part != "" && part != pathParts[j] {
				matched = false
				break
			}
		}

		if matched {
			// 提取参数
			params := make(map[string]string, len(pattern.paramNames))
			for i, idx := range pattern.paramIdxs {
				if idx < len(pathParts) {
					params[pattern.paramNames[i]] = pathParts[idx]
				}
			}
			return pattern.handler, params, true
		}
	}

	return nil, nil, false
}

// matchPath 匹配路径（支持 {param} 语法）
func (r *DefaultRouter) matchPath(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if part == "" && pathParts[i] == "" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}

	return true
}

// handleRequest 处理请求
func (r *DefaultRouter) handleRequest(w http.ResponseWriter, req *http.Request, handler mvc.HandlerFunc, pattern string) {
	ctx := NewContext(w, req)

	// 提取路径参数（使用已知的路由模式）
	params := r.extractParamsForPattern(pattern, req.URL.Path)
	ctx.WithParams(params)

	// 构建中间件链
	allMiddleware := append([]mvc.MiddlewareFunc{}, r.middlewares...)

	ctx.WithMiddleware(allMiddleware, handler)
	ctx.Next()
}

// extractParamsForPattern 根据给定模式提取路径参数
func (r *DefaultRouter) extractParamsForPattern(pattern, path string) map[string]string {
	params := make(map[string]string)

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return params
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := part[1 : len(part)-1]
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// extractParams 提取路径参数（使用预编译模式）
// 注意：此方法主要用于测试，实际请求处理使用 extractParamsForPattern
func (r *DefaultRouter) extractParams(path string) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pathParts := strings.Split(path, "/")

	// 使用预编译的路由模式进行匹配
	for i := range r.routePatterns {
		pattern := &r.routePatterns[i]
		if !pattern.hasParams {
			continue
		}
		if len(pattern.parts) != len(pathParts) {
			continue
		}

		// 快速路径匹配
		matched := true
		for j, part := range pattern.parts {
			if part != "" && part != pathParts[j] {
				matched = false
				break
			}
		}

		if matched {
			// 提取参数
			params := make(map[string]string, len(pattern.paramNames))
			for i, idx := range pattern.paramIdxs {
				if idx < len(pathParts) {
					params[pattern.paramNames[i]] = pathParts[idx]
				}
			}
			return params
		}
	}

	return nil
}
