package server

import (
	"crypto/tls"
	goNet "net"
	"net/http"
	"time"
)

// TLSClientOption 是 HTTPS 客户端配置选项。
type TLSClientOption func(*TLSClientBuilder)

// TLSClientBuilder 是 HTTPS 客户端构建器，封装 NetClient 并配置 TLS。
type TLSClientBuilder struct {
	baseURL   string
	tlsConfig *tls.Config
	timeout   time.Duration
	transport *http.Transport
	headers   map[string]string
}

// NewTLSClient 创建 HTTPS 客户端构建器。
//
// 参数:
//   - baseURL: 基础 URL，如 "https://example.com"
//   - opts: 可选配置选项
//
// 返回值:
//   - *NetClient: 配置好的 HTTPS 客户端实例
func NewTLSClient(baseURL string, opts ...TLSClientOption) *NetClient {
	b := &TLSClientBuilder{
		baseURL: baseURL,
		timeout: DefaultTimeout,
	}

	for _, opt := range opts {
		opt(b)
	}

	transport := b.transport
	if transport == nil {
		transport = &http.Transport{
			TLSClientConfig: b.tlsConfig,
			DialContext: (&goNet.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	header := make(http.Header)
	for k, v := range b.headers {
		header.Set(k, v)
	}

	return NewClient(baseURL,
		WithClientTimeout(b.timeout),
		WithHeaders(header),
		func(c *NetClient) {
			c.httpClient.Transport = transport
		},
	)
}

// WithTLSConfig 设置 TLS 配置。
func WithTLSConfig(tlsConfig *tls.Config) TLSClientOption {
	return func(b *TLSClientBuilder) {
		b.tlsConfig = tlsConfig
	}
}

// WithInsecureTLS 跳过 TLS 证书验证（仅用于开发测试）。
func WithInsecureTLS() TLSClientOption {
	return func(b *TLSClientBuilder) {
		b.tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}
	}
}

// WithTLSRequestTimeout 设置请求超时时间。
func WithTLSRequestTimeout(timeout time.Duration) TLSClientOption {
	return func(b *TLSClientBuilder) {
		b.timeout = timeout
	}
}

// WithTLSDefaultHeader 设置默认请求头。
func WithTLSDefaultHeader(key, value string) TLSClientOption {
	return func(b *TLSClientBuilder) {
		if b.headers == nil {
			b.headers = make(map[string]string)
		}
		b.headers[key] = value
	}
}

// WithTLSTransport 设置自定义 HTTP Transport。
func WithTLSTransport(transport *http.Transport) TLSClientOption {
	return func(b *TLSClientBuilder) {
		b.transport = transport
	}
}
