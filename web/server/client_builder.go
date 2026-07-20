// Package server 提供 HTTP 服务器功能，用于 enhance 框架。
package server

import (
	"net/http"
	"net/url"
	"time"

	"github.com/xudefa/enhance/log"
)

// NewExponentialBackoff 创建指数退避策略。
func NewExponentialBackoff(baseDelay, maxDelay time.Duration, retryableStatus ...int) *ExponentialBackoff {
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 10 * time.Second
	}
	if len(retryableStatus) == 0 {
		retryableStatus = []int{500, 502, 503, 504}
	}
	return &ExponentialBackoff{
		baseDelay:       baseDelay,
		maxDelay:        maxDelay,
		retryableStatus: retryableStatus,
	}
}

// NewFixedDelay 创建固定延迟策略。
func NewFixedDelay(delay time.Duration, retryableStatus ...int) *FixedDelay {
	if delay <= 0 {
		delay = 1 * time.Second
	}
	if len(retryableStatus) == 0 {
		retryableStatus = []int{500, 502, 503, 504}
	}
	return &FixedDelay{
		delay:           delay,
		retryableStatus: retryableStatus,
	}
}

// NewClient 创建新的客户端。
func NewClient(baseURL string, opts ...ClientOption) *NetClient {
	c := &NetClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(http.Header),
		logger:  log.Build(), // 这里是默认值，从容器中获取日志记录器， 在下面for循环会覆盖
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewRetryableClient 创建可重试的 HTTP 客户端。
func NewRetryableClient(client *NetClient, opts ...RetryOption) *RetryableClient {
	cfg := RetryConfig{
		maxAttempts: 3,
		strategy:    NewExponentialBackoff(100*time.Millisecond, 10*time.Second),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &RetryableClient{
		client: client,
		config: cfg,
	}
}

// WithMaxAttempts 设置最大重试次数。
func WithMaxAttempts(n int) RetryOption {
	return func(c *RetryConfig) {
		c.maxAttempts = n
	}
}

// WithRetryStrategy 设置重试策略。
func WithRetryStrategy(strategy RetryStrategy) RetryOption {
	return func(c *RetryConfig) {
		c.strategy = strategy
	}
}

// WithOnRetry 设置重试回调。
func WithOnRetry(fn func(attempt int, resp *HTTPResponse, err error)) RetryOption {
	return func(c *RetryConfig) {
		c.onRetry = fn
	}
}

// WithClientTimeout 设置客户端请求超时时间。
func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(c *NetClient) {
		c.httpClient.Timeout = timeout
	}
}

// WithHeaders 设置默认请求头。
func WithHeaders(headers http.Header) ClientOption {
	return func(c *NetClient) {
		c.headers = headers.Clone()
	}
}

// WithLog 设置日志记录器。
func WithLog(logger log.Logger) ClientOption {
	return func(c *NetClient) {
		c.logger = logger
	}
}

// WithHeader 设置请求头。
func WithHeader(key, value string) RequestOption {
	return func(c *HTTPRequest) {
		if c.Header == nil {
			c.Header = make(http.Header)
		}
		c.Header.Set(key, value)
	}
}

// WithQuery 设置查询参数。
func WithQuery(key, value string) RequestOption {
	return func(c *HTTPRequest) {
		if c.Query == nil {
			c.Query = make(url.Values)
		}
		c.Query.Set(key, value)
	}
}

// WithTimeout 设置请求超时。
func WithTimeout(timeout time.Duration) RequestOption {
	return func(c *HTTPRequest) {
		c.Timeout = timeout
	}
}

// WithAuthToken 设置认证令牌。
func WithAuthToken(token string) RequestOption {
	return func(c *HTTPRequest) {
		c.AuthToken = token
	}
}

// WithContentType 设置请求 Content-Type。
func WithContentType(contentType string) RequestOption {
	return func(c *HTTPRequest) {
		c.ContentType = contentType
	}
}

// WithBasicAuth 设置基本认证。
func WithBasicAuth(username, password string) RequestOption {
	return func(c *HTTPRequest) {
		c.BasicAuth = BasicAuth{Username: username, Password: password}
	}
}
