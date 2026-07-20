package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPClientBuilder HTTP 客户端构建器，支持链式配置
type HTTPClientBuilder struct {
	baseURL     string
	timeout     time.Duration
	headers     http.Header
	middleware  []ClientMiddlewareFunc
	retryConfig *RetryConfig
}

// NewHTTPClientBuilder 创建 HTTP 客户端构建器
func NewHTTPClientBuilder() *HTTPClientBuilder {
	return &HTTPClientBuilder{
		timeout: DefaultTimeout,
		headers: make(http.Header),
	}
}

// BaseURL 设置基础 URL
func (b *HTTPClientBuilder) BaseURL(url string) *HTTPClientBuilder {
	b.baseURL = url
	return b
}

// Timeout 设置请求超时时间
func (b *HTTPClientBuilder) Timeout(timeout time.Duration) *HTTPClientBuilder {
	b.timeout = timeout
	return b
}

// Header 添加默认请求头
func (b *HTTPClientBuilder) Header(key, value string) *HTTPClientBuilder {
	b.headers.Set(key, value)
	return b
}

// Middleware 添加客户端中间件
func (b *HTTPClientBuilder) Middleware(m ClientMiddlewareFunc) *HTTPClientBuilder {
	b.middleware = append(b.middleware, m)
	return b
}

// Retry 配置重试策略
func (b *HTTPClientBuilder) Retry(opts ...RetryOption) *HTTPClientBuilder {
	cfg := &RetryConfig{
		maxAttempts: 3,
		strategy:    NewExponentialBackoff(100*time.Millisecond, 10*time.Second),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	b.retryConfig = cfg
	return b
}

// Build 构建 HTTP 客户端
func (b *HTTPClientBuilder) Build() (HTTPClient, error) {
	if b.baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	client := NewClient(b.baseURL,
		WithClientTimeout(b.timeout),
		WithHeaders(b.headers),
	)

	for _, m := range b.middleware {
		client.WithMiddleware(m)
	}

	if b.retryConfig != nil {
		retryableClient := NewRetryableClient(client,
			WithMaxAttempts(b.retryConfig.maxAttempts),
			WithRetryStrategy(b.retryConfig.strategy),
		)
		if b.retryConfig.onRetry != nil {
			retryableClient = NewRetryableClient(client,
				WithMaxAttempts(b.retryConfig.maxAttempts),
				WithRetryStrategy(b.retryConfig.strategy),
				WithOnRetry(b.retryConfig.onRetry),
			)
		}
		return retryableClient, nil
	}

	return client, nil
}

// MustBuild 构建 HTTP 客户端，失败则 panic
func (b *HTTPClientBuilder) MustBuild() HTTPClient {
	client, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build HTTP client: %v", err))
	}
	return client
}

// RequestBuilder 请求构建器，简化 HTTP 请求构建
type RequestBuilder struct {
	ctx     context.Context
	method  string
	path    string
	body    any
	opts    []RequestOption
	headers http.Header
	query   map[string]string
}

// NewRequestBuilder 创建请求构建器
func NewRequestBuilder(ctx context.Context, method, path string) *RequestBuilder {
	return &RequestBuilder{
		ctx:     ctx,
		method:  method,
		path:    path,
		headers: make(http.Header),
		query:   make(map[string]string),
	}
}

// Body 设置请求体
func (b *RequestBuilder) Body(body any) *RequestBuilder {
	b.body = body
	return b
}

// Header 添加请求头
func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
	b.headers.Set(key, value)
	return b
}

// Query 添加查询参数
func (b *RequestBuilder) Query(key, value string) *RequestBuilder {
	b.query[key] = value
	return b
}

// AuthToken 设置认证令牌
func (b *RequestBuilder) AuthToken(token string) *RequestBuilder {
	b.opts = append(b.opts, WithAuthToken(token))
	return b
}

// ContentType 设置 Content-Type
func (b *RequestBuilder) ContentType(contentType string) *RequestBuilder {
	b.opts = append(b.opts, WithContentType(contentType))
	return b
}

// Timeout 设置请求超时
func (b *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder {
	b.opts = append(b.opts, WithTimeout(timeout))
	return b
}

// Build 构建请求选项
func (b *RequestBuilder) Build() []RequestOption {
	opts := make([]RequestOption, len(b.opts))
	copy(opts, b.opts)

	for key, values := range b.headers {
		for _, value := range values {
			k, v := key, value
			opts = append(opts, func(req *HTTPRequest) {
				if req.Header == nil {
					req.Header = make(http.Header)
				}
				req.Header.Add(k, v)
			})
		}
	}

	for key, value := range b.query {
		k, v := key, value
		opts = append(opts, WithQuery(k, v))
	}

	return opts
}

// Execute 执行请求
func (b *RequestBuilder) Execute(client HTTPClient) (*HTTPResponse, error) {
	opts := b.Build()

	switch b.method {
	case "GET":
		return client.Get(b.ctx, b.path, opts...)
	case "POST":
		return client.Post(b.ctx, b.path, b.body, opts...)
	case "PUT":
		return client.Put(b.ctx, b.path, b.body, opts...)
	case "PATCH":
		return client.Patch(b.ctx, b.path, b.body, opts...)
	case "DELETE":
		return client.Delete(b.ctx, b.path, opts...)
	case "HEAD":
		return client.Head(b.ctx, b.path, opts...)
	case "OPTIONS":
		return client.Options(b.ctx, b.path, opts...)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", b.method)
	}
}
