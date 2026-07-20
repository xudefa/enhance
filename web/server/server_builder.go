package server

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPServerBuilder HTTP 服务器构建器，支持链式配置
type HTTPServerBuilder struct {
	host         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	certFile     string
	keyFile      string
	middlewares  []func(http.Handler) http.Handler
	handler      http.Handler
}

// NewHTTPServerBuilder 创建 HTTP 服务器构建器
func NewHTTPServerBuilder() *HTTPServerBuilder {
	return &HTTPServerBuilder{
		host:         ":8080",
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
		idleTimeout:  120 * time.Second,
	}
}

// Host 设置监听地址
func (b *HTTPServerBuilder) Host(host string) *HTTPServerBuilder {
	b.host = host
	return b
}

// ReadTimeout 设置读取超时
func (b *HTTPServerBuilder) ReadTimeout(timeout time.Duration) *HTTPServerBuilder {
	b.readTimeout = timeout
	return b
}

// WriteTimeout 设置写入超时
func (b *HTTPServerBuilder) WriteTimeout(timeout time.Duration) *HTTPServerBuilder {
	b.writeTimeout = timeout
	return b
}

// IdleTimeout 设置空闲超时
func (b *HTTPServerBuilder) IdleTimeout(timeout time.Duration) *HTTPServerBuilder {
	b.idleTimeout = timeout
	return b
}

// TLS 设置 TLS 证书
func (b *HTTPServerBuilder) TLS(certFile, keyFile string) *HTTPServerBuilder {
	b.certFile = certFile
	b.keyFile = keyFile
	return b
}

// Middleware 添加中间件
func (b *HTTPServerBuilder) Middleware(middleware func(http.Handler) http.Handler) *HTTPServerBuilder {
	b.middlewares = append(b.middlewares, middleware)
	return b
}

// Handler 设置 HTTP 处理器
func (b *HTTPServerBuilder) Handler(handler http.Handler) *HTTPServerBuilder {
	b.handler = handler
	return b
}

// Build 构建 HTTP 服务器
func (b *HTTPServerBuilder) Build() (*HttpServer, error) {
	server := NewHTTPServer(
		WithHost(b.host),
		WithReadTimeout(b.readTimeout),
		WithWriteTimeout(b.writeTimeout),
		WithIdleTimeout(b.idleTimeout),
	)

	if b.certFile != "" && b.keyFile != "" {
		WithTLS(b.certFile, b.keyFile)(server)
	}

	if b.handler != nil {
		server.SetHandler(b.handler)
	}

	for _, m := range b.middlewares {
		server.Use(m)
	}

	return server, nil
}

// MustBuild 构建 HTTP 服务器，失败则 panic
func (b *HTTPServerBuilder) MustBuild() *HttpServer {
	server, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build HTTP server: %v", err))
	}
	return server
}
