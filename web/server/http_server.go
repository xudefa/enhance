package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/web/mvc"
)

// HttpServer 是基于 net/http 原生包的 Server 接口实现。
type HttpServer struct {
	server       *http.Server                      // 底层 HTTP 服务器
	host         string                            // 监听地址
	readTimeout  time.Duration                     // 读取超时
	writeTimeout time.Duration                     // 写入超时
	idleTimeout  time.Duration                     // 空闲超时
	certFile     string                            // TLS 证书文件路径
	keyFile      string                            // TLS 密钥文件路径
	middlewares  []func(http.Handler) http.Handler // 中间件列表
	mu           sync.RWMutex                      // 读写锁
	handler      http.Handler                      // 自定义处理器
	enableHTTP2  bool                              // 是否启用 HTTP/2
	logger       log.Logger                        // 日志器
}

// ServerOption 是服务器配置选项函数。
type ServerOption func(*HttpServer)

// WithHost 设置服务器监听地址。
func WithHost(host string) ServerOption {
	return func(c *HttpServer) {
		c.host = host
	}
}

// WithReadTimeout 设置读取超时。
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(c *HttpServer) {
		c.readTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时。
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(c *HttpServer) {
		c.writeTimeout = timeout
	}
}

// WithIdleTimeout 设置空闲超时。
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(c *HttpServer) {
		c.idleTimeout = timeout
	}
}

// WithTLS 设置 TLS 证书和密钥文件路径。
func WithTLS(certFile, keyFile string) ServerOption {
	return func(c *HttpServer) {
		c.certFile = certFile
		c.keyFile = keyFile
	}
}

// WithHTTP2 启用 HTTP/2 支持（需要 TLS）。
func WithHTTP2(enabled bool) ServerOption {
	return func(c *HttpServer) {
		c.enableHTTP2 = enabled
	}
}

// WithLogger 设置日志记录器。
func WithLogger(logger log.Logger) ServerOption {
	return func(c *HttpServer) {
		c.logger = logger
	}
}

// NewHTTPServer 创建一个新的基于 net/http 的 HTTP 服务器。
func NewHTTPServer(opts ...ServerOption) *HttpServer {
	s := &HttpServer{
		host:         ":8080",
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		idleTimeout:  120 * time.Second,
		middlewares:  make([]func(http.Handler) http.Handler, 0),
		logger:       log.Build(), // 这里是默认值，从容器中获取日志记录器， 在下面for循环会覆盖
	}

	for _, opt := range opts {
		opt(s)
	}

	s.server = &http.Server{
		Addr:         s.host,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}

	return s
}

// Start 启动 HTTP 服务器并开始监听请求。
func (s *HttpServer) Start() error {
	s.mu.Lock()

	if s.handler == nil {
		s.server.Handler = s.wrapHTTPHandler(http.DefaultServeMux)
	} else {
		// 根据处理器类型包装
		switch h := s.handler.(type) {
		case http.Handler:
			s.server.Handler = s.wrapHTTPHandler(h)
		case mvc.Router:
			s.server.Handler = s.wrapRouter(h)
		default:
			s.mu.Unlock()
			return fmt.Errorf("unsupported handler type: %T", s.handler)
		}
	}

	// 配置 HTTP/2
	if s.enableHTTP2 && s.certFile != "" && s.keyFile != "" {
		s.configureHTTP2()
	}

	s.mu.Unlock()

	s.logger.Info(context.Background(), "HTTP 服务器启动中",
		log.KeyValue{Key: "addr", Value: s.host},
		log.KeyValue{Key: "http2", Value: s.enableHTTP2},
	)
	if s.certFile != "" && s.keyFile != "" {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	return s.server.ListenAndServe()
}

// configureHTTP2 配置 HTTP/2 支持
func (s *HttpServer) configureHTTP2() {
	// 配置 TLS 以支持 HTTP/2
	if s.server.TLSConfig == nil {
		s.server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	// 设置 ALPN 协议以支持 HTTP/2
	s.server.TLSConfig.NextProtos = append(s.server.TLSConfig.NextProtos, "h2", "http/1.1")

	s.logger.Info(context.Background(), "HTTP/2 已启用",
		log.KeyValue{Key: "protocols", Value: s.server.TLSConfig.NextProtos},
	)
}

// StartTLS 启动带 TLS 的 HTTP 服务器（支持 HTTP/2）。
func (s *HttpServer) StartTLS(certFile, keyFile string) error {
	s.certFile = certFile
	s.keyFile = keyFile

	if s.enableHTTP2 {
		s.configureHTTP2()
	}

	s.logger.Info(context.Background(), "HTTPS 服务器启动中",
		log.KeyValue{Key: "addr", Value: s.host},
		log.KeyValue{Key: "http2", Value: s.enableHTTP2},
	)
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

// StartH2C 启动 HTTP/2 明文支持（不需要 TLS）。
func (s *HttpServer) StartH2C() error {
	s.logger.Info(context.Background(), "H2C 服务器启动中",
		log.KeyValue{Key: "addr", Value: s.host},
	)

	// 创建 H2C 服务器
	h2s := &http.Server{
		Addr:         s.host,
		Handler:      s.server.Handler,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}

	// 创建监听器
	lis, err := net.Listen("tcp", s.host)
	if err != nil {
		s.logger.Error(context.Background(), "H2C 监听失败", log.KeyValue{Key: "error", Value: err.Error()})
		return fmt.Errorf("failed to listen: %w", err)
	}

	// 启动 H2C 服务器
	return h2s.Serve(lis)
}

// Use 向服务器注册一个中间件。
func (s *HttpServer) Use(middleware func(http.Handler) http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.middlewares = append(s.middlewares, middleware)
}

// Stop 优雅地停止服务器，等待正在处理的请求完成。
func (s *HttpServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info(context.Background(), "HTTP 服务器停止中...")
	return s.server.Shutdown(ctx)
}

// SetHandler 设置自定义的 HTTP 处理器。
func (s *HttpServer) SetHandler(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handler = handler
}

// wrapHTTPHandler 将中间件包装到 http.Handler 中。
func (s *HttpServer) wrapHTTPHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 构建中间件链
		allMiddleware := append([]func(http.Handler) http.Handler{}, s.middlewares...)

		// 创建最终处理器
		var currentHandler http.Handler = handler
		for i := len(allMiddleware) - 1; i >= 0; i-- {
			currentHandler = allMiddleware[i](currentHandler)
		}

		currentHandler.ServeHTTP(w, r)
	})
}

// wrapRouter 将中间件包装到 mvc.Router 中。
func (s *HttpServer) wrapRouter(router mvc.Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 路由器自己处理中间件和请求
		switch rt := router.(type) {
		case http.Handler:
			rt.ServeHTTP(w, r)
		default:
			http.Error(w, "router does not implement http.Handler", http.StatusInternalServerError)
		}
	})
}
