package security

import (
	"context"
	"testing"
	"time"
)

func TestEnhancedRateLimitFilter_Allow(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 100)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter)

	req := &mockSecurityRequest{
		uri:    "/api/test",
		method: "GET",
		headers: map[string]string{
			"X-Real-IP": "192.168.1.1",
		},
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called")
	}
}

func TestEnhancedRateLimitFilter_ExcludePath(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 0)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter,
		WithExcludePaths("/health", "/metrics"),
	)

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

func TestEnhancedRateLimitFilter_RateLimited(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 0)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter)

	req := &mockSecurityRequest{
		uri:    "/api/test",
		method: "GET",
		headers: map[string]string{
			"X-Real-IP": "192.168.1.1",
		},
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.statusCode != 429 {
		t.Errorf("expected status 429, got %d", resp.statusCode)
	}
	if chain.called {
		t.Error("expected chain not to be called when rate limited")
	}
}

func TestEnhancedRateLimitFilter_CustomCallback(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 0)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	callbackCalled := false
	filter := NewEnhancedRateLimitFilter(adapter,
		WithOnRateLimit(func(ctx context.Context, request SecurityRequest, response SecurityResponse) {
			callbackCalled = true
			response.SetStatusCode(503)
		}),
	)

	req := &mockSecurityRequest{
		uri:    "/api/test",
		method: "GET",
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	_ = filter.DoFilter(context.Background(), req, resp, chain)

	if !callbackCalled {
		t.Error("expected custom callback to be called")
	}
	if resp.statusCode != 503 {
		t.Errorf("expected status 503 from callback, got %d", resp.statusCode)
	}
}

func TestEnhancedRateLimitFilter_WithTrustedProxies(t *testing.T) {
	t.Parallel()

	limiter := NewSlidingWindowRateLimiter(1*time.Second, 100)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter,
		WithTrustedProxies("10.0.0.0/8", "192.168.1.100"),
	)

	if !filter.trustProxyHeaders {
		t.Error("expected trustProxyHeaders to be true")
	}
	if len(filter.trustedProxyNets) == 0 {
		t.Error("expected trustedProxyNets to be populated")
	}
}

func TestEnhancedRateLimitFilter_Order(t *testing.T) {
	t.Parallel()

	limiter := NewSlidingWindowRateLimiter(1*time.Second, 100)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	filter := NewEnhancedRateLimitFilter(adapter)

	// Order应该返回0
	if filter.Order() != 0 {
		t.Errorf("expected order 0, got %d", filter.Order())
	}
}
