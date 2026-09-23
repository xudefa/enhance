package server

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/xudefa/enhance/web/core"
)

// DefaultRouter 默认路由器实现
type DefaultRouter struct {
	middlewares []core.MiddlewareFunc
	handlers    map[string]core.HandlerFunc
	prefix      string
	mu          *sync.RWMutex // 共享的mutex，父子router共享
	// 性能优化：缓存路由模式，避免每次请求都遍历
	routePatterns *[]routePattern
	// 性能优化：按HTTP方法索引路由模式，将O(n)全量扫描降为O(k)方法内扫描
	methodPatterns *map[string][]int // method -> indices into routePatterns
	// 路由注册时的中间件链（子路由组的中间件在 handle 时绑定到路由）
	routeMiddleware map[string][]core.MiddlewareFunc
}

// routePattern 预编译的路由模式
type routePattern struct {
	method      string
	parts       []string
	paramNames  []string // 参数名称列表
	paramIdxs   []int    // 参数在路径中的索引
	handler     core.HandlerFunc
	hasParams   bool
	patternPath string // 原始路由模式路径（含 {param}），用于中间件查找
}

// NewRouter 创建新的路由器
func NewRouter() *DefaultRouter {
	patterns := make([]routePattern, 0)
	methodIdx := make(map[string][]int)
	return &DefaultRouter{
		handlers:        make(map[string]core.HandlerFunc),
		mu:              &sync.RWMutex{},
		routePatterns:   &patterns,
		methodPatterns:  &methodIdx,
		routeMiddleware: make(map[string][]core.MiddlewareFunc),
	}
}

// GET 注册 GET 路由
func (r *DefaultRouter) GET(path string, handler core.HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

// POST 注册 POST 路由
func (r *DefaultRouter) POST(path string, handler core.HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

// PUT 注册 PUT 路由
func (r *DefaultRouter) PUT(path string, handler core.HandlerFunc) {
	r.handle(http.MethodPut, path, handler)
}

// DELETE 注册 DELETE 路由
func (r *DefaultRouter) DELETE(path string, handler core.HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

// PATCH 注册 PATCH 路由
func (r *DefaultRouter) PATCH(path string, handler core.HandlerFunc) {
	r.handle(http.MethodPatch, path, handler)
}

// Handle 注册任意 HTTP 方法的路由
func (r *DefaultRouter) Handle(method, path string, handler core.HandlerFunc) {
	r.handle(method, path, handler)
}

// Group 创建路由组
func (r *DefaultRouter) Group(prefix string) core.Router {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 子router与父router共享handlers映射、mutex、routePatterns和routeMiddleware
	return &DefaultRouter{
		middlewares:     append([]core.MiddlewareFunc{}, r.middlewares...),
		handlers:        r.handlers,
		prefix:          r.prefix + prefix,
		mu:              r.mu,
		routePatterns:   r.routePatterns,
		methodPatterns:  r.methodPatterns,
		routeMiddleware: r.routeMiddleware,
	}
}

// Use 注册中间件
func (r *DefaultRouter) Use(middleware core.MiddlewareFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middleware)
}

// ServeHTTP 实现 http.Handler 接口
func (r *DefaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 查找匹配的路由
	matchPath := r.resolveMatchPath(req.URL.Path)
	if matchPath == "" {
		http.NotFound(w, req)
		return
	}

	handler, params, middlewaresCopy, _, ok := r.resolveRoute(req.Method, matchPath)
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

// resolveMatchPath 处理路由前缀，返回去前缀后的完整匹配路径；前缀校验失败返回空字符串。
func (r *DefaultRouter) resolveMatchPath(path string) string {
	matchPath := path
	if r.prefix == "" {
		return matchPath
	}

	// 校验请求路径确实以路由前缀开头
	if !strings.HasPrefix(path, r.prefix) {
		return ""
	}
	// 前缀必须落在路径分段边界上：/api 不能匹配 /apix
	if len(path) > len(r.prefix) && path[len(r.prefix)] != '/' {
		return ""
	}
	path = strings.TrimPrefix(path, r.prefix)
	if path == "" {
		path = "/"
	}
	return r.prefix + path
}

// resolveRoute 在读锁内查找路由处理器及其路径参数与中间件。
func (r *DefaultRouter) resolveRoute(method, matchPath string) (handler core.HandlerFunc, params map[string]string, middlewares []core.MiddlewareFunc, patternKey string, ok bool) {
	key := method + " " + matchPath
	effectiveMethod := method

	// 使用读锁保护 handlers 和 routeMiddleware 的读取
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, ok = r.handlers[key]
	if !ok && method == http.MethodHead {
		// HEAD 请求回退到 GET 路由
		effectiveMethod = http.MethodGet
		key = effectiveMethod + " " + matchPath
		handler, ok = r.handlers[key]
	}
	if !ok {
		handler, params, ok, patternKey = r.findHandlerWithParamsLocked(effectiveMethod, matchPath)
	}
	if patternKey != "" {
		key = patternKey
	}
	middlewares = r.routeMiddleware[key]
	return
}

// handle 注册路由
func (r *DefaultRouter) handle(method, path string, handler core.HandlerFunc) {
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
	idx := len(*r.routePatterns)
	*r.routePatterns = append(*r.routePatterns, pattern)

	// 性能优化：维护方法索引，将O(n)全量扫描降为O(k)方法内扫描
	if pattern.hasParams {
		(*r.methodPatterns)[method] = append((*r.methodPatterns)[method], idx)
	}

	// 绑定路由注册时的中间件链（组中间件在此成为处理链的一部分）
	r.routeMiddleware[key] = append([]core.MiddlewareFunc{}, r.middlewares...)
}

// compileRoutePattern 预编译路由模式
func (r *DefaultRouter) compileRoutePattern(method, path string, handler core.HandlerFunc) routePattern {
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
func (r *DefaultRouter) findHandlerWithParamsLocked(method, path string) (core.HandlerFunc, map[string]string, bool, string) {
	pathParts := strings.Split(path, "/")
	patterns := *r.routePatterns

	// 性能优化：使用方法索引只扫描对应HTTP方法的路由模式，O(k)而非O(n)
	indices, hasMethodIndex := (*r.methodPatterns)[method]
	if hasMethodIndex {
		for _, i := range indices {
			pattern := &patterns[i]
			if len(pattern.parts) != len(pathParts) {
				continue
			}

			matched := true
			for j, part := range pattern.parts {
				if part != "" && part != pathParts[j] {
					matched = false
					break
				}
			}

			if matched {
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

	// 降级：全量扫描（兼容无方法索引的场景）
	for i := range patterns {
		pattern := &patterns[i]
		if pattern.method != method {
			continue
		}
		if !pattern.hasParams {
			continue
		}
		if len(pattern.parts) != len(pathParts) {
			continue
		}

		matched := true
		for j, part := range pattern.parts {
			if part != "" && part != pathParts[j] {
				matched = false
				break
			}
		}

		if matched {
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
