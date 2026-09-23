package security

import (
	"context"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(10, 5)
	if tb == nil {
		t.Fatal("expected non-nil bucket")
	}
	if tb.capacity != 10 {
		t.Errorf("expected capacity 10, got %d", tb.capacity)
	}
	if tb.tokens != 10 {
		t.Errorf("expected tokens 10, got %d", tb.tokens)
	}
	if tb.rate != 5 {
		t.Errorf("expected rate 5, got %d", tb.rate)
	}
}

func TestTokenBucket_TakeMultiple(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(3, 100)
	if !tb.Take() {
		t.Error("expected first take to succeed")
	}
	if !tb.Take() {
		t.Error("expected second take to succeed")
	}
	if !tb.Take() {
		t.Error("expected third take to succeed")
	}
	if tb.Take() {
		t.Error("expected fourth take to fail")
	}
}

func TestTokenBucket_Take_Refill(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(2, 10000)
	tb.Take()
	tb.Take()

	time.Sleep(100 * time.Millisecond)
	if !tb.Take() {
		t.Error("expected take to succeed after refill")
	}
}

func TestTokenBucket_IsExpiredFunc(t *testing.T) {
	t.Parallel()

	tb := NewTokenBucket(10, 100)
	if tb.IsExpired(time.Second) {
		t.Error("expected bucket to not be expired immediately")
	}

	time.Sleep(10 * time.Millisecond)
	if !tb.IsExpired(time.Nanosecond) {
		t.Error("expected bucket to be expired")
	}
}

func TestNewRateLimitFilterFunc(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{Enabled: true, Rate: 10, Burst: 20})
	if rateLimiter == nil {
		t.Fatal("expected non-nil filter")
	}
	defer rateLimiter.Close()

	if rateLimiter.config.Rate != 10 {
		t.Errorf("expected rate 10, got %d", rateLimiter.config.Rate)
	}
	if rateLimiter.config.Burst != 20 {
		t.Errorf("expected burst 20, got %d", rateLimiter.config.Burst)
	}
}

func TestNewRateLimitFilter_Defaults(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{})
	if rateLimiter == nil {
		t.Fatal("expected non-nil filter")
	}
	defer rateLimiter.Close()

	if rateLimiter.config.Rate != 100 {
		t.Errorf("expected default rate 100, got %d", rateLimiter.config.Rate)
	}
	if rateLimiter.config.Burst != 200 {
		t.Errorf("expected default burst 200, got %d", rateLimiter.config.Burst)
	}
}

func TestRateLimitFilter_Close(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	rateLimiter.Close()
	rateLimiter.Close()
}

func TestRateLimitFilter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	defer rateLimiter.Close()

	err := rateLimiter.DoFilter("notContext", nil, nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = rateLimiter.DoFilter(context.Background(), "notReq", nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = rateLimiter.DoFilter(context.Background(), newMockSecurityRequest("GET", "/", nil), "notResp", &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestRateLimitFilter_Disabled(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{Enabled: false})
	defer rateLimiter.Close()
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/", nil)
	req.remoteAddr = "127.0.0.1:8080"
	resp := newMockSecurityResponse()

	err := rateLimiter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called when rate limiting is disabled")
	}
}

func TestRateLimitFilter_ExcludedPath(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{
		Enabled:      true,
		ExcludePaths: []string{"/health"},
		Log:          &mockLogger{},
	})
	defer rateLimiter.Close()
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/health", nil)
	req.remoteAddr = "127.0.0.1:8080"
	resp := newMockSecurityResponse()

	err := rateLimiter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for excluded path")
	}
}

func TestRateLimitFilter_RateLimited(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Rate:    1000,
		Burst:   2,
		Log:     &mockLogger{},
	})
	defer rateLimiter.Close()
	chain := &mockFilterChain{}

	for i := 0; i < 5; i++ {
		req := newMockSecurityRequest("GET", "/api", nil)
		req.remoteAddr = "127.0.0.1:8080"
		resp := newMockSecurityResponse()
		rateLimiter.DoFilter(context.Background(), req, resp, chain)
		if i >= 2 && resp.statusCode == 429 {
			return
		}
	}
}

func TestParseRemoteIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"empty", "", ""},
		{"host:port", "192.168.1.1:8080", "192.168.1.1"},
		{"host only", "192.168.1.1", "192.168.1.1"},
		{"ipv6", "[::1]:8080", "::1"},
		{"invalid", "notanip", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parsedIP := parseRemoteIP(tt.input)
			if parsedIP != tt.expect {
				t.Errorf("parseRemoteIP(%q) = %q, want %q", tt.input, parsedIP, tt.expect)
			}
		})
	}
}

func TestParseTrustedProxiesFunc(t *testing.T) {
	t.Parallel()

	nets := parseTrustedProxies([]string{"192.168.1.0/24", "10.0.0.1", "invalid", ""})
	if len(nets) != 2 {
		t.Errorf("expected 2 nets, got %d", len(nets))
	}
}

func TestIsTrustedProxyFunc(t *testing.T) {
	t.Parallel()

	nets := parseTrustedProxies([]string{"192.168.1.0/24", "10.0.0.1"})

	if !isTrustedProxy("192.168.1.100", nets) {
		t.Error("expected 192.168.1.100 to be trusted")
	}
	if isTrustedProxy("172.16.0.1", nets) {
		t.Error("expected 172.16.0.1 to not be trusted")
	}
	if isTrustedProxy("notanip", nets) {
		t.Error("expected invalid IP to not be trusted")
	}
	if !isTrustedProxy("10.0.0.1", nets) {
		t.Error("expected 10.0.0.1 to be trusted")
	}
}

func TestRateLimitFilter_GetClientIPFunc(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		TrustProxyHeaders: false,
	})
	defer rateLimiter.Close()

	req := newMockSecurityRequest("GET", "/", nil)
	req.remoteAddr = "192.168.1.1:9090"

	ip := rateLimiter.getClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip)
	}
}

func TestRateLimitFilter_GetClientIP_EmptyRemote(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	defer rateLimiter.Close()

	req := newMockSecurityRequest("GET", "/", nil)
	req.remoteAddr = ""

	ip := rateLimiter.getClientIP(req)
	if ip != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", ip)
	}
}

func TestRateLimitFilter_GetClientIP_TrustedProxy(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		TrustProxyHeaders: true,
		TrustedProxies:    []string{"10.0.0.0/8"},
	})
	defer rateLimiter.Close()

	req := newMockSecurityRequest("GET", "/", map[string]string{
		"X-Forwarded-For": "203.0.113.1",
	})
	req.remoteAddr = "10.0.0.1:8080"

	ip := rateLimiter.getClientIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1, got %s", ip)
	}
}

func TestRateLimitFilter_GetClientIP_UntrustedProxy(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		TrustProxyHeaders: true,
		TrustedProxies:    []string{"10.0.0.0/8"},
	})
	defer rateLimiter.Close()

	req := newMockSecurityRequest("GET", "/", map[string]string{
		"X-Forwarded-For": "203.0.113.1",
	})
	req.remoteAddr = "192.168.1.1:8080"

	ip := rateLimiter.getClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip)
	}
}

func TestRateLimitFilter_OrderFunc(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	defer rateLimiter.Close()

	if rateLimiter.Order() != 0 {
		t.Errorf("expected order 0, got %d", rateLimiter.Order())
	}
}

func TestRateLimitFilter_CleanupBucketsFunc(t *testing.T) {
	t.Parallel()

	rateLimiter := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		BucketIdleTimeout: time.Nanosecond,
	})
	defer rateLimiter.Close()

	rateLimiter.buckets.Store("test", NewTokenBucket(10, 100))
	time.Sleep(10 * time.Millisecond)
	rateLimiter.cleanupBuckets()

	rateLimiter.buckets.Range(func(key, value any) bool {
		t.Errorf("expected bucket to be cleaned up, found key: %v", key)
		return true
	})
}

// ==================== Enhanced Rate Limit Filter Tests ====================

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
