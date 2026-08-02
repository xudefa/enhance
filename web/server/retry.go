// Package server 提供 HTTP 服务器功能，用于 enhance 框架。
package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"time"
)

// Get 发送 GET 请求（支持重试）。
func (c *RetryableClient) Get(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Get(ctx, url, opts...)
	})
}

// Post 发送 POST 请求（支持重试）。
func (c *RetryableClient) Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Post(ctx, url, body, opts...)
	})
}

// Put 发送 PUT 请求（支持重试）。
func (c *RetryableClient) Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Put(ctx, url, body, opts...)
	})
}

// Delete 发送 DELETE 请求（支持重试）。
func (c *RetryableClient) Delete(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Delete(ctx, url, opts...)
	})
}

// Patch 发送 PATCH 请求（支持重试）。
func (c *RetryableClient) Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Patch(ctx, url, body, opts...)
	})
}

// Head 发送 HEAD 请求（支持重试）。
func (c *RetryableClient) Head(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Head(ctx, url, opts...)
	})
}

// Options 发送 OPTIONS 请求（支持重试）。
func (c *RetryableClient) Options(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Options(ctx, url, opts...)
	})
}

// Do 执行自定义请求（支持重试）。
func (c *RetryableClient) Do(ctx context.Context, req any) (*HTTPResponse, error) {
	return c.doWithRetry(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Do(ctx, req)
	})
}

// doWithRetry 执行带重试的请求。
func (c *RetryableClient) doWithRetry(ctx context.Context, fn func(context.Context) (*HTTPResponse, error)) (*HTTPResponse, error) {
	if c.config.maxAttempts < 1 {
		return nil, fmt.Errorf("maxAttempts must be at least 1")
	}
	if c.config.strategy == nil {
		return nil, fmt.Errorf("retry strategy must not be nil")
	}

	var lastResp *HTTPResponse
	var lastErr error

	for attempt := 0; attempt < c.config.maxAttempts; attempt++ {
		lastResp, lastErr = fn(ctx)

		if !c.config.strategy.ShouldRetry(lastResp, lastErr, attempt) {
			return lastResp, lastErr
		}

		if c.config.onRetry != nil {
			c.config.onRetry(attempt, lastResp, lastErr)
		}

		if attempt == c.config.maxAttempts-1 {
			break
		}

		delay := c.config.strategy.Delay(attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return lastResp, ctx.Err()
		case <-timer.C:
			timer.Stop()
		}
	}

	return lastResp, lastErr
}

// Get 发送 GET 请求（带断路器保护）。
func (c *CircuitBreakerClient) Get(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Get(ctx, url, opts...)
	})
}

// Post 发送 POST 请求（带断路器保护）。
func (c *CircuitBreakerClient) Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Post(ctx, url, body, opts...)
	})
}

// Put 发送 PUT 请求（带断路器保护）。
func (c *CircuitBreakerClient) Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Put(ctx, url, body, opts...)
	})
}

// Delete 发送 DELETE 请求（带断路器保护）。
func (c *CircuitBreakerClient) Delete(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Delete(ctx, url, opts...)
	})
}

// Head 发送 HEAD 请求（带断路器保护）。
func (c *CircuitBreakerClient) Head(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Head(ctx, url, opts...)
	})
}

// Patch 发送 PATCH 请求（带断路器保护）。
func (c *CircuitBreakerClient) Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Patch(ctx, url, body, opts...)
	})
}

// Options 发送 OPTIONS 请求（带断路器保护）。
func (c *CircuitBreakerClient) Options(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Options(ctx, url, opts...)
	})
}

// Do 执行自定义请求（带断路器保护）。
func (c *CircuitBreakerClient) Do(ctx context.Context, req any) (*HTTPResponse, error) {
	return c.execute(ctx, func(ctx context.Context) (*HTTPResponse, error) {
		return c.client.Do(ctx, req)
	})
}

// execute 执行带断路器保护的请求。
func (c *CircuitBreakerClient) execute(ctx context.Context, fn func(context.Context) (*HTTPResponse, error)) (*HTTPResponse, error) {
	if !c.breaker.AllowRequest() {
		if c.fallback != nil {
			return c.fallback(ctx)
		}
		return nil, fmt.Errorf("circuit breaker is open")
	}

	resp, err := fn(ctx)

	if err != nil {
		// 客户端自身的取消/超时不是服务故障，不应计入熔断（与 ShouldRetry 保持一致）
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return resp, err
		}
		c.breaker.RecordFailure()
		return resp, err
	}

	if resp != nil && resp.IsServerError() {
		c.breaker.RecordFailure()
		return resp, err
	}
	c.breaker.RecordSuccess()

	return resp, err
}

// ShouldRetry 判断是否应该重试。
func (e *ExponentialBackoff) ShouldRetry(resp *HTTPResponse, err error, attempt int) bool {
	if err != nil {
		// 上下文取消/超时不重试
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		return true
	}

	if resp != nil {
		for _, status := range e.retryableStatus {
			if resp.StatusCode == status {
				return true
			}
		}
	}

	return false
}

// Delay 计算延迟时间（指数退避 + 抖动）。
func (e *ExponentialBackoff) Delay(attempt int) time.Duration {
	// 限制位移上限，避免 baseDelay * (1<<shift) 溢出 int64 导致负延迟
	shift := uint(attempt)
	if shift > 62 {
		shift = 62
	}
	if maxShift := e.maxSafeShift(); shift > maxShift {
		shift = maxShift
	}

	delay := e.baseDelay * time.Duration(1<<shift)

	if delay > e.maxDelay {
		delay = e.maxDelay
	}

	jitter := delay / 2
	delay = delay - jitter + time.Duration(randInt64(int64(jitter*2)))

	if delay > e.maxDelay {
		delay = e.maxDelay
	}

	return delay
}

// maxSafeShift 计算 baseDelay*(1<<shift) 不会溢出且不超过 maxDelay 的最大位移量。
func (e *ExponentialBackoff) maxSafeShift() uint {
	if e.baseDelay <= 0 {
		return 0
	}
	limit := e.maxDelay
	if maxDiv := int64(math.MaxInt64) / int64(e.baseDelay); maxDiv > 0 && maxDiv < int64(limit) {
		limit = time.Duration(maxDiv)
	}
	ratio := limit / e.baseDelay
	if ratio <= 0 {
		return 0
	}
	return uint(bits.Len64(uint64(ratio))) - 1
}

// ShouldRetry 判断是否应该重试。
func (f *FixedDelay) ShouldRetry(resp *HTTPResponse, err error, attempt int) bool {
	if err != nil {
		// 上下文取消/超时不重试
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		return true
	}
	if resp != nil {
		for _, status := range f.retryableStatus {
			if resp.StatusCode == status {
				return true
			}
		}
	}
	return false
}

// Delay 返回固定延迟。
func (f *FixedDelay) Delay(attempt int) time.Duration {
	return f.delay
}

// randInt64 生成 [0, max) 范围内的随机整数。
func randInt64(max int64) int64 {
	if max <= 0 {
		return 0
	}
	return rand.Int63n(max)
}
