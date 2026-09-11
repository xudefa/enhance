package security

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

// ============================================================
// filter_chain.go 测试
// ============================================================

// TestSecurityFilterChainAdapter 测试安全过滤器链适配器
func TestSlidingWindowRateLimiter_Close(t *testing.T) {
	t.Parallel()

	limiter := NewSlidingWindowRateLimiter(100*time.Millisecond, 10)
	limiter.Close()
	// Closing twice should not panic
	limiter.Close()
}

// TestEnhancedRateLimitFilter_ClientKey 测试客户端键计算

func TestEnhancedRateLimitFilter_ClientKey(t *testing.T) {
	t.Parallel()

	strategy := &mockRateLimitStrategy{}
	filter := NewEnhancedRateLimitFilter(strategy)

	t.Run("without trusted proxies", func(t *testing.T) {
		req := &mockSecurityRequest{method: "GET", uri: "/test"}
		req.SetHeader("X-Forwarded-For", "10.0.0.1")
		key := filter.clientKey(req)
		if key == "" {
			t.Error("expected non-empty key")
		}
	})

	t.Run("with trusted proxies", func(t *testing.T) {
		filterWithProxy := NewEnhancedRateLimitFilter(strategy, WithTrustedProxies("192.168.0.0/16"))
		req := &mockSecurityRequest{method: "GET", uri: "/test", remoteAddr: "192.168.1.1:8080"}
		req.SetHeader("X-Forwarded-For", "10.0.0.1")
		key := filterWithProxy.clientKey(req)
		if key == "" {
			t.Error("expected non-empty key")
		}
	})
}

// TestEnhancedRateLimitFilter_DoFilter 测试增强限流过滤器的各种场景

func TestEnhancedRateLimitFilter_DoFilter(t *testing.T) {
	t.Parallel()

	t.Run("exclude path", func(t *testing.T) {
		filterWithExclude := NewEnhancedRateLimitFilter(&mockRateLimitStrategy{}, WithExcludePaths("/health"))
		req := &mockSecurityRequest{method: "GET", uri: "/health"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterWithExclude.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for excluded path")
		}
	})

	t.Run("rate limit triggered", func(t *testing.T) {
		denyStrategy := &mockRateLimitStrategy{allow: false}
		filterDeny := NewEnhancedRateLimitFilter(denyStrategy)
		req := &mockSecurityRequest{method: "GET", uri: "/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterDeny.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.statusCode != 429 {
			t.Errorf("expected status 429, got %d", resp.statusCode)
		}
	})

	t.Run("custom onRateLimit callback", func(t *testing.T) {
		customCalled := false
		denyStrategy := &mockRateLimitStrategy{allow: false}
		filterCustom := NewEnhancedRateLimitFilter(denyStrategy, WithOnRateLimit(func(ctx context.Context, request SecurityRequest, response SecurityResponse) {
			customCalled = true
			response.SetStatusCode(429)
		}))
		req := &mockSecurityRequest{method: "GET", uri: "/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterCustom.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !customCalled {
			t.Error("expected custom onRateLimit to be called")
		}
	})

	t.Run("allowed request passes through", func(t *testing.T) {
		allowStrategy := &mockRateLimitStrategy{allow: true}
		filterAllow := NewEnhancedRateLimitFilter(allowStrategy)
		req := &mockSecurityRequest{method: "GET", uri: "/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterAllow.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for allowed request")
		}
	})
}

// TestEnhancedRateLimitFilter_InvalidTypes 测试 EnhancedRateLimitFilter 的类型检查

func TestEnhancedRateLimitFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	f := NewEnhancedRateLimitFilter(&mockRateLimitStrategy{})

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// ============================================================
// cors_filter.go 测试
// ============================================================

// TestCorsFilter_IsOriginAllowed 测试 CORS 来源验证

func TestRateLimitFilter_DoFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Rate:    10,
		Burst:   20,
		Log:     newTestLogger(),
	})

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestRateLimitFilter_DoFilter_RateLimited 测试令牌桶限流

func TestRateLimitFilter_DoFilter_RateLimited(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Rate:    1,
		Burst:   1,
		Log:     newTestLogger(),
	})

	req := &mockSecurityRequest{method: "GET", uri: "/api", remoteAddr: "10.0.0.1:8080"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	// First request should be allowed
	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for first request")
	}

	// Second request should be rate limited
	chain.called = false
	err = f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.called {
		t.Error("expected chain not to be called for rate limited request")
	}
}

// ============================================================
// 辅助 Mock 类型
// ============================================================

// filterFuncFilter 用于测试的自定义过滤器
type filterFuncFilter struct {
	doFilter func(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error
}

func (f *filterFuncFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	if f.doFilter != nil {
		return f.doFilter(ctx, request, response, chain)
	}
	return chain.DoFilter(ctx, request, response)
}

func (f *filterFuncFilter) Order() int { return 0 }

// chainFuncFilter 用于测试的 SecurityFilterChain 实现
type chainFuncFilter struct {
	doFilter func(ctx interface{}, request interface{}, response interface{}) error
}

func (c *chainFuncFilter) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	if c.doFilter != nil {
		return c.doFilter(ctx, request, response)
	}
	return nil
}

func (c *chainFuncFilter) Matches(request interface{}) bool { return true }
func (c *chainFuncFilter) GetFilters() []filter.Filter      { return nil }

// mockLogoutHandler 用于测试的登出处理器 Mock
type mockLogoutHandler struct {
	called bool
}

func (m *mockLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	m.called = true
}

// mockRateLimitStrategy 用于测试的限流策略 Mock
type mockRateLimitStrategy struct {
	allow bool
}

func (m *mockRateLimitStrategy) Allow(key string) bool {
	return m.allow
}

// mockResponseWriter 用于测试的 http.ResponseWriter Mock
type mockResponseWriter struct {
	header http.Header
	code   int
	body   []byte
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{header: make(http.Header)}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(code int) {
	m.code = code
}

// newTestLogger 创建静默测试日志器（输出丢弃，避免污染测试输出）
func newTestLogger() log.Logger {
	return log.NewSlogLogger(log.WithLevel(log.ErrorLevel), log.WithOutput(io.Discard))
}
