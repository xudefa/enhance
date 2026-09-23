package security

import (
	"context"
	"testing"
	"time"
)

func TestNewRateLimitFilter(t *testing.T) {
	t.Parallel()
	config := RateLimitConfig{
		Enabled: true,
		Rate:    10,
		Burst:   20,
	}

	f := NewRateLimitFilter(config)
	if f == nil {
		t.Error("expected non-nil filter")
	}
}

func TestRateLimitFilter_Order(t *testing.T) {
	t.Parallel()
	config := RateLimitConfig{
		Enabled: true,
		Rate:    10,
		Burst:   20,
	}

	f := NewRateLimitFilter(config)
	order := f.Order()
	if order != 0 {
		t.Errorf("expected order 0, got %d", order)
	}
}

func TestRateLimitFilter_CloseAndOrder(t *testing.T) {
	t.Parallel()

	filter := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Burst:   10,
		Rate:    1.0,
	})

	// 验证Close不会panic
	filter.Close()

	// 多次调用Close应该安全
	filter.Close()

	// 验证Order
	if filter.Order() != 0 {
		t.Errorf("expected order 0, got %d", filter.Order())
	}
}

func TestRateLimitFilter_ExcludePaths(t *testing.T) {
	t.Parallel()

	filter := NewRateLimitFilter(RateLimitConfig{
		Enabled:      true,
		Burst:        10,
		Rate:         1.0,
		ExcludePaths: []string{"/health", "/metrics"},
		Log:          &mockLogger{},
	})

	req := &mockSecurityRequest{
		uri:    "/health",
		method: "GET",
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected excluded path to bypass rate limiting")
	}
}

func TestRateLimitFilter_CleanupBuckets(t *testing.T) {
	t.Parallel()

	// 创建一个带很短空闲超时的过滤器，以便快速清理
	filter := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		Burst:             10,
		Rate:              1.0,
		BucketIdleTimeout: 50 * time.Millisecond,
		CleanupInterval:   100 * time.Millisecond,
		Log:               &mockLogger{},
	})

	// 手动添加一些桶到buckets中
	filter.buckets.Store("192.168.1.1", NewTokenBucket(10, 1.0))
	filter.buckets.Store("192.168.1.2", NewTokenBucket(10, 1.0))

	// 等待超过空闲超时
	time.Sleep(100 * time.Millisecond)

	// 手动调用cleanupBuckets
	filter.cleanupBuckets()

	// 验证桶已被清理
	count := 0
	filter.buckets.Range(func(key, value any) bool {
		count++
		return true
	})

	if count != 0 {
		t.Errorf("expected 0 buckets after cleanup, got %d", count)
	}

	filter.Close()
}

func testRateLimitFilterGetClientIPFromRemoteAddr(t *testing.T) {
	t.Parallel()

	filter := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Burst:   10,
		Rate:    1.0,
		Log:     &mockLogger{},
	})

	req := &mockSecurityRequest{}

	// mockSecurityRequest.RemoteAddress()返回"127.0.0.1:8080"
	ip := filter.getClientIP(req)
	if ip != "127.0.0.1" {
		t.Errorf("expected IP '127.0.0.1', got '%s'", ip)
	}

	filter.Close()
}

func testRateLimitFilterGetClientIPFromForwardedFor(t *testing.T) {
	t.Parallel()

	filter := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		Burst:             10,
		Rate:              1.0,
		TrustProxyHeaders: true,
		TrustedProxies:    []string{"127.0.0.1"},
		Log:               &mockLogger{},
	})

	req := &mockSecurityRequest{
		headers: map[string]string{
			"X-Forwarded-For": "203.0.113.50, 127.0.0.1",
		},
	}

	// RemoteAddress返回"127.0.0.1:8080"，在TrustedProxies中
	ip := filter.getClientIP(req)
	if ip != "203.0.113.50" {
		t.Errorf("expected IP '203.0.113.50', got '%s'", ip)
	}

	filter.Close()
}

func testRateLimitFilterGetClientIPIgnoreHeadersWithoutProxy(t *testing.T) {
	t.Parallel()

	filter := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Burst:   10,
		Rate:    1.0,
		Log:     &mockLogger{},
	})

	req := &mockSecurityRequest{
		headers: map[string]string{
			"X-Forwarded-For": "203.0.113.50",
		},
	}

	// 没有启用TrustProxyHeaders，应该忽略X-Forwarded-For
	ip := filter.getClientIP(req)
	if ip != "127.0.0.1" {
		t.Errorf("expected IP '127.0.0.1', got '%s'", ip)
	}

	filter.Close()
}

func TestRateLimitFilter_GetClientIP(t *testing.T) {
	t.Parallel()
	t.Run("from remote address", testRateLimitFilterGetClientIPFromRemoteAddr)
	t.Run("from X-Forwarded-For with trusted proxy", testRateLimitFilterGetClientIPFromForwardedFor)
	t.Run("ignore headers without trusted proxy", testRateLimitFilterGetClientIPIgnoreHeadersWithoutProxy)
}
