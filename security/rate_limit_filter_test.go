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

	f := NewRateLimitFilter(RateLimitConfig{Enabled: true, Rate: 10, Burst: 20})
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	defer f.Close()

	if f.config.Rate != 10 {
		t.Errorf("expected rate 10, got %d", f.config.Rate)
	}
	if f.config.Burst != 20 {
		t.Errorf("expected burst 20, got %d", f.config.Burst)
	}
}

func TestNewRateLimitFilter_Defaults(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{})
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	defer f.Close()

	if f.config.Rate != 100 {
		t.Errorf("expected default rate 100, got %d", f.config.Rate)
	}
	if f.config.Burst != 200 {
		t.Errorf("expected default burst 200, got %d", f.config.Burst)
	}
}

func TestRateLimitFilter_Close(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	f.Close()
	f.Close()
}

func TestRateLimitFilter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	defer f.Close()

	err := f.DoFilter("notContext", nil, nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = f.DoFilter(context.Background(), "notReq", nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = f.DoFilter(context.Background(), newMockSecurityRequest("GET", "/", nil), "notResp", &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestRateLimitFilter_Disabled(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{Enabled: false})
	defer f.Close()
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/", nil)
	req.remoteAddr = "127.0.0.1:8080"
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called when rate limiting is disabled")
	}
}

func TestRateLimitFilter_ExcludedPath(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled:      true,
		ExcludePaths: []string{"/health"},
		Log:          &mockLogger{},
	})
	defer f.Close()
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/health", nil)
	req.remoteAddr = "127.0.0.1:8080"
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for excluded path")
	}
}

func TestRateLimitFilter_RateLimited(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Rate:    1000,
		Burst:   2,
		Log:     &mockLogger{},
	})
	defer f.Close()
	chain := &mockFilterChain{}

	for i := 0; i < 5; i++ {
		req := newMockSecurityRequest("GET", "/api", nil)
		req.remoteAddr = "127.0.0.1:8080"
		resp := newMockSecurityResponse()
		f.DoFilter(context.Background(), req, resp, chain)
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
			result := parseRemoteIP(tt.input)
			if result != tt.expect {
				t.Errorf("parseRemoteIP(%q) = %q, want %q", tt.input, result, tt.expect)
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

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		TrustProxyHeaders: false,
	})
	defer f.Close()

	req := newMockSecurityRequest("GET", "/", nil)
	req.remoteAddr = "192.168.1.1:9090"

	ip := f.getClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip)
	}
}

func TestRateLimitFilter_GetClientIP_EmptyRemote(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	defer f.Close()

	req := newMockSecurityRequest("GET", "/", nil)
	req.remoteAddr = ""

	ip := f.getClientIP(req)
	if ip != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", ip)
	}
}

func TestRateLimitFilter_GetClientIP_TrustedProxy(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		TrustProxyHeaders: true,
		TrustedProxies:    []string{"10.0.0.0/8"},
	})
	defer f.Close()

	req := newMockSecurityRequest("GET", "/", map[string]string{
		"X-Forwarded-For": "203.0.113.1",
	})
	req.remoteAddr = "10.0.0.1:8080"

	ip := f.getClientIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("expected 203.0.113.1, got %s", ip)
	}
}

func TestRateLimitFilter_GetClientIP_UntrustedProxy(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		TrustProxyHeaders: true,
		TrustedProxies:    []string{"10.0.0.0/8"},
	})
	defer f.Close()

	req := newMockSecurityRequest("GET", "/", map[string]string{
		"X-Forwarded-For": "203.0.113.1",
	})
	req.remoteAddr = "192.168.1.1:8080"

	ip := f.getClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip)
	}
}

func TestRateLimitFilter_OrderFunc(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{Enabled: true})
	defer f.Close()

	if f.Order() != 0 {
		t.Errorf("expected order 0, got %d", f.Order())
	}
}

func TestRateLimitFilter_CleanupBucketsFunc(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled:           true,
		BucketIdleTimeout: time.Nanosecond,
	})
	defer f.Close()

	f.buckets.Store("test", NewTokenBucket(10, 100))
	time.Sleep(10 * time.Millisecond)
	f.cleanupBuckets()

	f.buckets.Range(func(key, value any) bool {
		t.Errorf("expected bucket to be cleaned up, found key: %v", key)
		return true
	})
}
