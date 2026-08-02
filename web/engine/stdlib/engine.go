// Package stdlib 提供基于标准库 net/http 的引擎实现。
package stdlib

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xudefa/enhance/web/core"
	"github.com/xudefa/enhance/web/engine"
)

// Router 标准库路由器实现。
type Router struct {
	mux           *http.ServeMux
	middlewares   []core.MiddlewareFunc
	handlers      map[string]core.HandlerFunc
	prefix        string
	mu            *sync.RWMutex // 共享的mutex，父子router共享
	routePatterns *[]routePattern
}

type routePattern struct {
	method     string
	parts      []string
	paramNames []string
	paramIdxs  []int
	handler    core.HandlerFunc
	hasParams  bool
}

// NewRouter 创建新的路由器。
func NewRouter() *Router {
	patterns := make([]routePattern, 0)
	return &Router{
		mux:           http.NewServeMux(),
		handlers:      make(map[string]core.HandlerFunc),
		mu:            &sync.RWMutex{},
		routePatterns: &patterns,
	}
}

// GET 注册 GET 路由。
func (r *Router) GET(path string, handler core.HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

// POST 注册 POST 路由。
func (r *Router) POST(path string, handler core.HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

// PUT 注册 PUT 路由。
func (r *Router) PUT(path string, handler core.HandlerFunc) {
	r.handle(http.MethodPut, path, handler)
}

// DELETE 注册 DELETE 路由。
func (r *Router) DELETE(path string, handler core.HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

// PATCH 注册 PATCH 路由。
func (r *Router) PATCH(path string, handler core.HandlerFunc) {
	r.handle(http.MethodPatch, path, handler)
}

// Group 创建路由组。
func (r *Router) Group(prefix string) core.Router {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 子router与父router共享handlers映射、mutex和routePatterns
	return &Router{
		mux:           r.mux,
		middlewares:   append([]core.MiddlewareFunc{}, r.middlewares...),
		handlers:      r.handlers,
		prefix:        r.prefix + prefix,
		mu:            r.mu,
		routePatterns: r.routePatterns,
	}
}

// Use 注册中间件。
func (r *Router) Use(middleware core.MiddlewareFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middleware)
}

// ServeHTTP 实现 http.Handler 接口。
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if r.prefix != "" {
		path = strings.TrimPrefix(path, r.prefix)
		if path == "" {
			path = "/"
		}
	}

	matchPath := path
	if r.prefix != "" {
		matchPath = r.prefix + path
	}

	fullPath := req.Method + " " + matchPath

	// 使用读锁保护 handlers 和 middlewares 的读取
	r.mu.RLock()
	var params map[string]string
	handler, ok := r.handlers[fullPath]
	if !ok {
		handler, params, ok = r.findHandlerWithParamsLocked(req.Method, matchPath)
	}
	middlewaresCopy := make([]core.MiddlewareFunc, len(r.middlewares))
	copy(middlewaresCopy, r.middlewares)
	r.mu.RUnlock()

	if !ok {
		http.NotFound(w, req)
		return
	}

	ctx := NewContext(w, req)
	if params != nil {
		ctx = ctx.WithParams(params)
	}

	ctx.WithMiddleware(middlewaresCopy, handler)
	ctx.Next()
}

func (r *Router) handle(method, path string, handler core.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fullPath := r.prefix + path
	key := method + " " + fullPath
	r.handlers[key] = handler

	pattern := r.compileRoutePattern(method, fullPath, handler)
	*r.routePatterns = append(*r.routePatterns, pattern)
}

func (r *Router) compileRoutePattern(method, path string, handler core.HandlerFunc) routePattern {
	parts := strings.Split(path, "/")
	paramNames := make([]string, 0, len(parts))
	paramIdxs := make([]int, 0, len(parts))
	hasParams := false

	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramNames = append(paramNames, part[1:len(part)-1])
			paramIdxs = append(paramIdxs, i)
			parts[i] = ""
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

func (r *Router) findHandlerWithParamsLocked(method, path string) (core.HandlerFunc, map[string]string, bool) {
	pathParts := strings.Split(path, "/")

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

		matched := true
		for j, part := range pattern.parts {
			if part != "" && part != pathParts[j] {
				matched = false
				break
			}
		}

		if matched {
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

// Server 标准库服务器实现。
type Server struct {
	server       *http.Server
	host         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	certFile     string
	keyFile      string
	middlewares  []func(http.Handler) http.Handler
	mu           sync.RWMutex
	handler      http.Handler
}

// NewServer 创建新的 HTTP 服务器。
func NewServer(opts ...engine.ServerOption) *Server {
	config := engine.DefaultServerConfig()
	for _, opt := range opts {
		opt(config)
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	readTimeout := time.Duration(config.ReadTimeout) * time.Second
	writeTimeout := time.Duration(config.WriteTimeout) * time.Second
	idleTimeout := time.Duration(config.IdleTimeout) * time.Second

	s := &Server{
		host:         addr,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		idleTimeout:  idleTimeout,
		certFile:     config.TLSCertFile,
		keyFile:      config.TLSKeyFile,
		middlewares:  make([]func(http.Handler) http.Handler, 0),
	}

	s.server = &http.Server{
		Addr:         addr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return s
}

// Start 启动 HTTP 服务器。
func (s *Server) Start() error {
	s.mu.Lock()

	if s.handler == nil {
		s.server.Handler = s.wrapHTTPHandler(http.DefaultServeMux)
	} else {
		s.server.Handler = s.wrapHTTPHandler(s.handler)
	}

	s.mu.Unlock()

	slog.Info("HTTP server starting", "addr", s.host)
	if s.certFile != "" && s.keyFile != "" {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	return s.server.ListenAndServe()
}

// Stop 优雅地停止服务器。
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()

	// 在锁外执行 Shutdown：Shutdown 会等待正在处理的请求完成，
	// 而请求处理器会获取读锁，若持写锁会形成死锁。
	slog.Info("HTTP server stopping...")
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

// SetHandler 设置处理器。
func (s *Server) SetHandler(handler any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch h := handler.(type) {
	case http.Handler:
		s.handler = h
	case core.Router:
		if httpHandler, ok := h.(http.Handler); ok {
			s.handler = httpHandler
		}
	default:
		// 不支持的处理器类型，保持 handler 为 nil
	}
}

// Use 注册中间件。
func (s *Server) Use(m any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch mw := m.(type) {
	case func(http.Handler) http.Handler:
		s.middlewares = append(s.middlewares, mw)
	case core.MiddlewareFunc:
		adapter := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := NewContext(w, r).WithMiddleware([]core.MiddlewareFunc{mw}, func(c core.Context) {
					next.ServeHTTP(w, r)
				})
				ctx.Next()
			})
		}
		s.middlewares = append(s.middlewares, adapter)
	}
}

func (s *Server) wrapHTTPHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		allMiddleware := append([]func(http.Handler) http.Handler{}, s.middlewares...)
		s.mu.RUnlock()

		currentHandler := handler
		for i := len(allMiddleware) - 1; i >= 0; i-- {
			currentHandler = allMiddleware[i](currentHandler)
		}

		currentHandler.ServeHTTP(w, r)
	})
}

// Factory 标准库引擎工厂。
type Factory struct{}

// Type 返回引擎类型。
func (f *Factory) Type() engine.Type {
	return engine.StdLib
}

// CreateRouter 创建路由器。
func (f *Factory) CreateRouter() (core.Router, error) {
	return NewRouter(), nil
}

// CreateServer 创建服务器。
func (f *Factory) CreateServer(opts ...engine.ServerOption) (core.Server, error) {
	return NewServer(opts...), nil
}

// init 自动注册标准库引擎。
func init() {
	engine.GlobalRegistry.Register(&Factory{})
}
