package server

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/xudefa/enhance/web/mvc"
)

// DefaultRouter 默认路由器实现
type DefaultRouter struct {
	middlewares []mvc.MiddlewareFunc
	handlers    map[string]mvc.HandlerFunc
	prefix      string
	mu          *sync.RWMutex // 共享的mutex，父子router共享
	// 性能优化：缓存路由模式，避免每次请求都遍历
	routePatterns *[]routePattern
	// 路由注册时的中间件链（子路由组的中间件在 handle 时绑定到路由）
	routeMiddleware map[string][]mvc.MiddlewareFunc
}

// routePattern 预编译的路由模式
type routePattern struct {
	method      string
	parts       []string
	paramNames  []string // 参数名称列表
	paramIdxs   []int    // 参数在路径中的索引
	handler     mvc.HandlerFunc
	hasParams   bool
	patternPath string // 原始路由模式路径（含 {param}），用于中间件查找
}

// NewRouter 创建新的路由器
func NewRouter() *DefaultRouter {
	patterns := make([]routePattern, 0)
	return &DefaultRouter{
		handlers:        make(map[string]mvc.HandlerFunc),
		mu:              &sync.RWMutex{},
		routePatterns:   &patterns,
		routeMiddleware: make(map[string][]mvc.MiddlewareFunc),
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

	// 子router与父router共享handlers映射、mutex、routePatterns和routeMiddleware
	return &DefaultRouter{
		middlewares:     append([]mvc.MiddlewareFunc{}, r.middlewares...),
		handlers:        r.handlers,
		prefix:          r.prefix + prefix,
		mu:              r.mu,
		routePatterns:   r.routePatterns,
		routeMiddleware: r.routeMiddleware,
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
	matchPath := path
	if r.prefix != "" {
		// 校验请求路径确实以路由前缀开头
		if !strings.HasPrefix(path, r.prefix) {
			http.NotFound(w, req)
			return
		}
		// 前缀必须落在路径分段边界上：/api 不能匹配 /apix
		if len(path) > len(r.prefix) && path[len(r.prefix)] != '/' {
			http.NotFound(w, req)
			return
		}
		path = strings.TrimPrefix(path, r.prefix)
		if path == "" {
			path = "/"
		}
		matchPath = r.prefix + path
	}

	// 构建完整路径用于查找（路由模式使用完整前缀+路径编译）
	// RFC 7231 §4.3.2：无显式 HEAD 路由时，HEAD 请求由 GET 路由处理
	method := req.Method
	key := method + " " + matchPath

	// 使用读锁保护 handlers 和 routeMiddleware 的读取
	r.mu.RLock()
	var params map[string]string
	var patternKey string
	handler, ok := r.handlers[key]
	if !ok {
		// HEAD 请求回退到 GET 路由
		if method == http.MethodHead {
			method = http.MethodGet
			key = method + " " + matchPath
			handler, ok = r.handlers[key]
		}
	}
	if !ok {
		handler, params, ok, patternKey = r.findHandlerWithParamsLocked(method, matchPath)
	}
	if patternKey != "" {
		key = patternKey
	}
	middlewaresCopy := r.routeMiddleware[key]
	r.mu.RUnlock()

	if !ok {
		http.NotFound(w, req)
		return
	}

	// 创建上下文
	ctx := NewContext(w, req)

	// 设置路径参数
	if params != nil {
		ctx.WithParams(params)
	}

	// 执行中间件链和处理器
	ctx.WithMiddleware(middlewaresCopy, handler)
	ctx.Next()
}

// handle 注册路由
func (r *DefaultRouter) handle(method, path string, handler mvc.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPath := r.prefix + path
	key := method + " " + fullPath

	// 拒绝重复注册，避免静默覆盖导致静态路由与参数路由行为不一致
	if _, exists := r.handlers[key]; exists {
		slog.Error("duplicate route registration ignored",
			"method", method,
			"path", fullPath,
		)
		return
	}

	r.handlers[key] = handler

	// 性能优化：预编译路由模式
	pattern := r.compileRoutePattern(method, fullPath, handler)
	*r.routePatterns = append(*r.routePatterns, pattern)

	// 绑定路由注册时的中间件链（组中间件在此成为处理链的一部分）
	r.routeMiddleware[key] = append([]mvc.MiddlewareFunc{}, r.middlewares...)
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
		method:      method,
		parts:       parts,
		paramNames:  paramNames,
		paramIdxs:   paramIdxs,
		handler:     handler,
		hasParams:   hasParams,
		patternPath: path,
	}
}

// findHandlerWithParamsLocked 查找带路径参数的路由（使用预编译模式，调用方须持有读锁）
func (r *DefaultRouter) findHandlerWithParamsLocked(method, path string) (mvc.HandlerFunc, map[string]string, bool, string) {
	pathParts := strings.Split(path, "/")

	// 使用预编译的路由模式进行匹配，避免每次都解析
	for i := range *r.routePatterns {
		pattern := &(*r.routePatterns)[i]
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
			for pi, idx := range pattern.paramIdxs {
				if idx < len(pathParts) {
					params[pattern.paramNames[pi]] = pathParts[idx]
				}
			}
			return pattern.handler, params, true, method + " " + pattern.patternPath
		}
	}

	return nil, nil, false, ""
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
