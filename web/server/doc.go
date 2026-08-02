// Package server 提供 HTTP 服务器功能，用于 enhance 框架。
//
// 该模块提供高性能 HTTP 服务器、路由器、中间件、HTTP 客户端、TLS 支持等网络通信功能。
// 参考 Spring Boot 的嵌入式服务器设计。
//
// # 架构设计
//
//   - HTTPServer: HTTP 服务器接口，定义服务器操作
//   - Router: 路由器接口，负责路由注册和匹配
//   - Middleware: 中间件接口，处理请求和响应
//   - HTTPClient: HTTP 客户端接口，发送 HTTP 请求
//   - RetryStrategy: 重试策略接口
//   - CircuitBreaker: 断路器接口
//   - TLSConfig: TLS 配置，支持 HTTPS
//
// # 核心功能
//
//   - HTTP 服务器: 支持高性能 HTTP 服务器
//   - 路由器: 支持 RESTful 路由和注解驱动路由
//   - 中间件: 支持请求拦截、日志、认证等中间件
//   - HTTP 客户端: 提供简单易用的 HTTP 客户端
//   - TLS 支持: 支持 HTTPS 和 TLS 配置
//   - 优雅关闭: 支持服务器优雅关闭
//   - 重试机制: 支持指数退避和固定延迟重试
//   - 断路器: 支持熔断保护
//
// # 使用方式
//
// 创建服务器：
//
//	srv := server.NewServer(":8080")
//	srv.GET("/api/users", userHandler)
//	srv.POST("/api/users", createUserHandler)
//
// 添加中间件：
//
//	srv.Use(loggingMiddleware)
//	srv.Use(authMiddleware)
//
// 启动服务器：
//
//	srv.Start()
//
// # TLS 配置
//
// 支持 HTTPS：
//
//	srv := server.NewServer(":8443")
//	srv.SetTLSConfig("/path/to/cert.pem", "/path/to/key.pem")
//	srv.StartTLS()
//
// # 优雅关闭
//
// 支持优雅关闭：
//
//	srv.Shutdown(ctx)
package server

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/xudefa/enhance/log"
)

// HTTPServer HTTP 服务器接口。
//
// 定义 HTTP 服务器的生命周期管理。
type HTTPServer interface {
	// Start 启动 HTTP 服务器并开始监听请求。
	Start() error
	// Stop 优雅地停止服务器，等待正在处理的请求完成。
	Stop(ctx context.Context) error
	// SetHandler 设置自定义的 HTTP 处理器。
	SetHandler(handler http.Handler)
	// Use 向服务器注册一个中间件。
	Use(middleware func(http.Handler) http.Handler)
}

// Router 路由器接口。
//
// 提供路由注册和路由组功能。
type Router interface {
	// GET 注册 GET 路由。
	GET(path string, handler http.HandlerFunc)
	// POST 注册 POST 路由。
	POST(path string, handler http.HandlerFunc)
	// PUT 注册 PUT 路由。
	PUT(path string, handler http.HandlerFunc)
	// DELETE 注册 DELETE 路由。
	DELETE(path string, handler http.HandlerFunc)
	// PATCH 注册 PATCH 路由。
	PATCH(path string, handler http.HandlerFunc)
	// Group 创建路由组。
	Group(prefix string) Router
	// Use 注册中间件。
	Use(middleware func(http.Handler) http.Handler)
	// ServeHTTP 实现 http.Handler 接口。
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// HTTPClient HTTP 客户端接口。
//
// 提供统一的 HTTP 请求发送接口。
type HTTPClient interface {
	// Get 发送 GET 请求。
	Get(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error)
	// Head 发送 HEAD 请求。
	Head(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error)
	// Post 发送 POST 请求。
	Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error)
	// Put 发送 PUT 请求。
	Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error)
	// Patch 发送 PATCH 请求。
	Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error)
	// Delete 发送 DELETE 请求。
	Delete(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error)
	// Options 发送 OPTIONS 请求。
	Options(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error)
	// Do 发送自定义 HTTP 请求。
	Do(ctx context.Context, req any) (*HTTPResponse, error)
	// Close 关闭客户端并释放资源。
	Close() error
}

// RequestOption 请求选项配置函数。
type RequestOption func(*HTTPRequest)

// HTTPRequest HTTP 请求封装。
type HTTPRequest struct {
	Header      http.Header   // 请求头
	Query       url.Values    // 查询参数
	Timeout     time.Duration // 请求超时
	AuthToken   string        // Bearer 认证令牌
	ContentType string        // 请求 Content-Type
	BasicAuth   BasicAuth     // 基本认证凭据
}

// HTTPResponse HTTP 响应封装。
type HTTPResponse struct {
	StatusCode int         // HTTP 状态码
	Header     http.Header // 响应头
	Body       []byte      // 响应体
}

// BasicAuth HTTP 基本认证凭据。
type BasicAuth struct {
	Username string // 用户名
	Password string // 密码
}

// ClientMiddlewareFunc HTTP 客户端中间件函数类型。
type ClientMiddlewareFunc func(*http.Request, *HTTPResponse) error

// RetryStrategy 重试策略接口。
type RetryStrategy interface {
	// ShouldRetry 判断是否应该重试。
	ShouldRetry(resp *HTTPResponse, err error, attempt int) bool
	// Delay 计算下次重试的延迟时间。
	Delay(attempt int) time.Duration
}

// RetryOption 重试配置选项。
type RetryOption func(*RetryConfig)

// RetryConfig 重试配置。
type RetryConfig struct {
	maxAttempts int
	strategy    RetryStrategy
	onRetry     func(attempt int, resp *HTTPResponse, err error)
}

// CircuitBreakerOption 断路器配置选项。
type CircuitBreakerOption func(*CircuitBreakerConfig)

// CircuitBreakerConfig 断路器配置。
type CircuitBreakerConfig struct {
	maxFailures  int
	resetTimeout time.Duration
	fallback     func(ctx context.Context) (*HTTPResponse, error)
}

// CircuitState 断路器状态。
type CircuitState int

const (
	// DefaultTimeout 默认请求超时时间。
	DefaultTimeout = 30 * time.Second
	// DefaultMaxResponseBodySize 默认最大响应体大小（50 MB）。
	DefaultMaxResponseBodySize = 50 << 20
)

const (
	// CircuitClosed 关闭状态（正常请求）。
	CircuitClosed CircuitState = iota
	// CircuitOpen 打开状态（拒绝请求）。
	CircuitOpen
	// CircuitHalfOpen 半开状态（允许探测请求）。
	CircuitHalfOpen
)

// ClientOption 客户端配置选项。
type ClientOption func(*NetClient)

// NetClient HTTP 客户端实现。
type NetClient struct {
	mu         sync.RWMutex           // 读写锁
	baseURL    string                 // 基础 URL
	httpClient *http.Client           // 底层 HTTP 客户端
	headers    http.Header            // 默认请求头
	middleware []ClientMiddlewareFunc // 中间件列表
	logger     log.Logger             // 日志记录器
}

// ExponentialBackoff 指数退避重试策略。
type ExponentialBackoff struct {
	baseDelay       time.Duration
	maxDelay        time.Duration
	retryableStatus []int
}

// FixedDelay 固定延迟重试策略。
type FixedDelay struct {
	delay           time.Duration
	retryableStatus []int
}

// RetryableClient 支持重试的 HTTP 客户端。
type RetryableClient struct {
	client HTTPClient
	config RetryConfig
}

// CircuitBreaker 断路器。
type CircuitBreaker struct {
	mu           sync.Mutex // 并发锁
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailure  time.Time
	state        CircuitState
}

// CircuitBreakerClient 带断路器的 HTTP 客户端。
type CircuitBreakerClient struct {
	client   HTTPClient
	breaker  *CircuitBreaker
	fallback func(ctx context.Context) (*HTTPResponse, error)
}
