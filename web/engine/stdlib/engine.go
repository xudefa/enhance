// Package stdlib 提供基于标准库 net/http 的引擎实现。
package stdlib

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/xudefa/enhance/web/core"
	"github.com/xudefa/enhance/web/engine"
	"github.com/xudefa/enhance/web/server"
)

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
func (s *Server) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handler = handler
}

// Use 注册中间件。
func (s *Server) Use(middleware func(http.Handler) http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middlewares = append(s.middlewares, middleware)
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
	return server.NewRouter(), nil
}

// CreateServer 创建服务器。
func (f *Factory) CreateServer(opts ...engine.ServerOption) (core.Server, error) {
	return NewServer(opts...), nil
}

// init 自动注册标准库引擎。
func init() {
	engine.GlobalRegistry.Register(&Factory{})
}
