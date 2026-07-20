// Package server 提供 HTTP 服务器功能，用于 enhance 框架。
package server

import (
	"context"
	"time"
)

// NewCircuitBreaker 创建断路器。
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// NewCircuitBreakerClient 创建带断路器的 HTTP 客户端。
func NewCircuitBreakerClient(client *NetClient, opts ...CircuitBreakerOption) *CircuitBreakerClient {
	cfg := CircuitBreakerConfig{
		maxFailures:  5,
		resetTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return &CircuitBreakerClient{
		client:   client,
		breaker:  NewCircuitBreaker(cfg.maxFailures, cfg.resetTimeout),
		fallback: cfg.fallback,
	}
}

// WithCircuitMaxFailures 设置最大失败次数。
func WithCircuitMaxFailures(n int) CircuitBreakerOption {
	return func(c *CircuitBreakerConfig) {
		c.maxFailures = n
	}
}

// WithCircuitResetTimeout 设置重置超时。
func WithCircuitResetTimeout(d time.Duration) CircuitBreakerOption {
	return func(c *CircuitBreakerConfig) {
		c.resetTimeout = d
	}
}

// WithFallback 设置降级函数。
func WithFallback(fn func(ctx context.Context) (*HTTPResponse, error)) CircuitBreakerOption {
	return func(c *CircuitBreakerConfig) {
		c.fallback = fn
	}
}

// AllowRequest 判断是否允许请求。
func (cb *CircuitBreaker) AllowRequest() bool {
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess 记录成功。
func (cb *CircuitBreaker) RecordSuccess() {
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure 记录失败。
func (cb *CircuitBreaker) RecordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

// GetState 获取当前状态。
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// GetCircuitState 获取断路器状态。
func (c *CircuitBreakerClient) GetCircuitState() CircuitState {
	return c.breaker.GetState()
}

// Close 关闭客户端。
func (c *CircuitBreakerClient) Close() error {
	return c.client.Close()
}

// Close 关闭客户端。
func (c *RetryableClient) Close() error {
	return c.client.Close()
}

// Close 关闭客户端。
func (c *NetClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// WithMiddleware 添加中间件。
func (c *NetClient) WithMiddleware(m ClientMiddlewareFunc) *NetClient {
	c.mu.Lock()
	c.middleware = append(c.middleware, m)
	c.mu.Unlock()
	return c
}
