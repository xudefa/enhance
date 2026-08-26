package security

import (
	"context"
	"testing"
	"time"

	"github.com/xudefa/enhance/security/authorization"
	"github.com/xudefa/enhance/security/filter"
)

func TestSlidingWindowRateLimiter(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 3)

	for range 3 {
		if !limiter.Allow("user1") {
			t.Error("expected request to be allowed")
		}
	}

	if limiter.Allow("user1") {
		t.Error("expected 4th request to be denied")
	}

	if !limiter.Allow("user2") {
		t.Error("expected user2 to be allowed")
	}
}

func TestSlidingWindowRateLimiter_Cleanup(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(100*time.Millisecond, 10)

	limiter.Allow("user1")
	limiter.Allow("user2")

	time.Sleep(150 * time.Millisecond)
	limiter.Cleanup()

	count := limiter.WindowCount()
	if count != 0 {
		t.Errorf("expected 0 windows after cleanup, got %d", count)
	}
}

func TestSlidingWindowRateLimiter_DefaultValues(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(0, -1)

	if limiter.windowSize <= 0 {
		t.Error("expected default window size")
	}
	if limiter.maxRequests <= 0 {
		t.Error("expected default max requests")
	}
}

func TestLeakyBucketRateLimiter(t *testing.T) {
	t.Parallel()
	limiter := NewLeakyBucketRateLimiter(2, 10*time.Millisecond)

	if !limiter.Allow("user1") {
		t.Error("expected 1st request to be allowed")
	}
	if !limiter.Allow("user1") {
		t.Error("expected 2nd request to be allowed")
	}

	if limiter.Allow("user1") {
		t.Error("expected 3rd request to be denied")
	}

	time.Sleep(20 * time.Millisecond)

	if !limiter.Allow("user1") {
		t.Error("expected request after leak to be allowed")
	}
}

func TestLeakyBucketRateLimiter_Cleanup(t *testing.T) {
	t.Parallel()
	limiter := NewLeakyBucketRateLimiter(2, 50*time.Millisecond)

	limiter.Allow("user1")

	time.Sleep(200 * time.Millisecond)
	limiter.Cleanup()

	limiter.mu.RLock()
	count := len(limiter.buckets)
	limiter.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 buckets after cleanup, got %d", count)
	}
}

func TestLeakyBucketRateLimiter_DefaultValues(t *testing.T) {
	t.Parallel()
	limiter := NewLeakyBucketRateLimiter(0, 0)

	if limiter.capacity <= 0 {
		t.Error("expected default capacity")
	}
	if limiter.rate <= 0 {
		t.Error("expected default rate")
	}
}

func TestFixedWindowCounterRateLimiter(t *testing.T) {
	t.Parallel()
	limiter := NewFixedWindowCounterRateLimiter(1*time.Second, 2)

	if !limiter.Allow("user1") {
		t.Error("expected 1st request to be allowed")
	}
	if !limiter.Allow("user1") {
		t.Error("expected 2nd request to be allowed")
	}

	if limiter.Allow("user1") {
		t.Error("expected 3rd request to be denied")
	}
}

func TestFixedWindowCounterRateLimiter_Cleanup(t *testing.T) {
	t.Parallel()
	limiter := NewFixedWindowCounterRateLimiter(100*time.Millisecond, 10)

	limiter.Allow("user1")

	time.Sleep(150 * time.Millisecond)
	limiter.Cleanup()

	limiter.mu.RLock()
	count := len(limiter.counters)
	limiter.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 counters after cleanup, got %d", count)
	}
}

func TestFixedWindowCounterRateLimiter_DefaultValues(t *testing.T) {
	t.Parallel()
	limiter := NewFixedWindowCounterRateLimiter(0, 0)

	if limiter.windowSize <= 0 {
		t.Error("expected default window size")
	}
	if limiter.maxRequests <= 0 {
		t.Error("expected default max requests")
	}
}

func TestStrategyRateLimiterAdapter(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(1*time.Second, 100)
	adapter := NewStrategyRateLimiterAdapter(limiter)

	if !adapter.Allow("user1") {
		t.Error("expected adapter to allow request")
	}
}

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

func TestSecurityBuilder_BasicConfig(t *testing.T) {
	t.Parallel()
	authManager := &testAuthManager{}
	userDetailsService := &testUserDetailsService{}
	passwordEncoder := NewNoOpPasswordEncoder()

	config := NewSecurityBuilder().
		AuthenticationManager(authManager).
		UserDetailsService(userDetailsService).
		PasswordEncoder(passwordEncoder).
		EnableAnonymous().
		EnableHttpBasic().
		Build()

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_WithFilters(t *testing.T) {
	t.Parallel()
	filter1 := &testSecurityFilter{}
	filter2 := &testSecurityFilter{}

	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AddFilter(filter1).
		AddFilterAfter(filter2, filter1).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_FormLogin(t *testing.T) {
	t.Parallel()
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		EnableFormLogin("/login", "/home").
		EnableCsrf().
		EnableLogout("/logout").
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	chain, err := http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
	if chain == nil {
		t.Error("expected non-nil filter chain")
	}
}

func TestSecurityBuilder_LogoutWithHandler(t *testing.T) {
	t.Parallel()
	handler := &testLogoutSuccessHandler{}
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		EnableLogout("/logout", handler).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	chain, err := http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
	if chain == nil {
		t.Error("expected non-nil filter chain")
	}
}

func TestSecurityBuilder_AddFilterBefore(t *testing.T) {
	t.Parallel()
	filter1 := &testSecurityFilter{}
	filter2 := &testSecurityFilter{}

	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AddFilter(filter1).
		AddFilterBefore(filter2, filter1).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_AccessDecisionManager(t *testing.T) {
	t.Parallel()
	accessDecisionMgr := &testAccessDecisionManager{}
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AccessDecisionManager(accessDecisionMgr).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_EmptyBuild(t *testing.T) {
	t.Parallel()
	config := NewSecurityBuilder().Build()
	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err == nil {
		t.Error("expected build to fail without auth manager")
	}
}

func TestTokenBucket_Take(t *testing.T) {
	t.Parallel()
	bucket := NewTokenBucket(5, 10)

	// 应该允许前5个请求
	for range 5 {
		if !bucket.Take() {
			t.Error("expected token to be available")
		}
	}

	// 第6个请求应该被拒绝（令牌已用完）
	if bucket.Take() {
		t.Error("expected token to be exhausted")
	}
}

func TestTokenBucket_IsExpired(t *testing.T) {
	t.Parallel()
	bucket := NewTokenBucket(5, 10)

	// 刚创建不应该过期
	if bucket.IsExpired(time.Second) {
		t.Error("expected new bucket to not be expired")
	}

	// 模拟过期（通过修改lastAccess）
	bucket.mu.Lock()
	bucket.lastAccess = time.Now().Add(-time.Minute)
	bucket.mu.Unlock()

	// 现在应该过期
	if !bucket.IsExpired(time.Second) {
		t.Error("expected bucket to be expired after timeout")
	}
}

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

func TestParseTrustedProxies(t *testing.T) {
	t.Parallel()
	// 测试解析可信代理列表
	proxies := parseTrustedProxies([]string{"192.168.1.1", "10.0.0.0/8"})
	if len(proxies) != 2 {
		t.Errorf("expected 2 proxies, got %d", len(proxies))
	}
}

func TestIsTrustedProxy(t *testing.T) {
	t.Parallel()
	proxies := parseTrustedProxies([]string{"192.168.1.1", "10.0.0.0/8"})

	// 测试可信代理
	if !isTrustedProxy("192.168.1.1", proxies) {
		t.Error("expected 192.168.1.1 to be trusted")
	}

	// 测试CIDR范围内的IP
	if !isTrustedProxy("10.0.1.1", proxies) {
		t.Error("expected 10.0.1.1 to be trusted (in 10.0.0.0/8)")
	}

	// 测试不可信代理
	if isTrustedProxy("172.16.0.1", proxies) {
		t.Error("expected 172.16.0.1 to not be trusted")
	}
}

// 测试模拟实现

type testAuth struct {
	principal   string
	credentials string
}

func (a *testAuth) Principal() any        { return a.principal }
func (a *testAuth) Credentials() any      { return a.credentials }
func (a *testAuth) Authorities() []string { return []string{"ROLE_USER"} }
func (a *testAuth) Authenticated() bool   { return true }

type testAuthManager struct{}

func (m *testAuthManager) Authenticate(ctx context.Context, auth AuthenticationToken) (Authentication, error) {
	return &testAuth{principal: "testuser", credentials: "testpass"}, nil
}

type testUserDetailsService struct{}

func (s *testUserDetailsService) LoadUserByUsername(ctx context.Context, username string) (UserDetails, error) {
	return nil, nil
}

type testAccessDecisionManager struct{}

func (m *testAccessDecisionManager) Decide(ctx context.Context, auth authorization.Authentication, resource string, attrs []string) error {
	return nil
}

func (m *testAccessDecisionManager) Supports(attribute string) bool {
	return true
}

type testSecurityFilter struct{}

func (f *testSecurityFilter) DoFilter(ctx interface{}, req interface{}, resp interface{}, chain filter.FilterChain) error {
	return chain.DoFilter(ctx, req, resp)
}

func (f *testSecurityFilter) Order() int { return 0 }

type testLogoutSuccessHandler struct{}

func (h *testLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, req SecurityRequest, resp SecurityResponse, auth Authentication) {
	resp.SetStatusCode(302)
	resp.SetHeader("Location", "/login")
}

// TestLeakyBucketRateLimiter_ZeroCapacity 测试 capacity=0 时拒绝所有请求
func TestLeakyBucketRateLimiter_ZeroCapacity(t *testing.T) {
	t.Parallel()
	// 直接构造 capacity=0 的 limiter，绕过 NewLeakyBucketRateLimiter 的默认值
	limiter := &LeakyBucketRateLimiter{
		capacity: 0,
		rate:     10 * time.Millisecond,
		buckets:  make(map[string]*leakyBucket),
	}

	// capacity=0 时所有请求都应被拒绝
	if limiter.Allow("user1") {
		t.Error("expected request to be denied when capacity=0")
	}
	if limiter.Allow("user2") {
		t.Error("expected request to be denied when capacity=0")
	}
}

// TestSlidingWindowRateLimiter_LongRunning 测试长时间运行后的内存泄漏
func TestSlidingWindowRateLimiter_LongRunning(t *testing.T) {
	t.Parallel()
	limiter := NewSlidingWindowRateLimiter(50*time.Millisecond, 100)

	// 模拟大量请求
	for range 100 {
		limiter.Allow("user1")
	}

	// 等待窗口过期
	time.Sleep(100 * time.Millisecond)

	// 触发清理
	limiter.Cleanup()

	count := limiter.WindowCount()
	if count != 0 {
		t.Errorf("expected 0 windows after cleanup, got %d", count)
	}
}

// TestFixedWindowCounterRateLimiter_WindowBoundary 测试窗口边界突变行为
func TestFixedWindowCounterRateLimiter_WindowBoundary(t *testing.T) {
	t.Parallel()
	limiter := NewFixedWindowCounterRateLimiter(100*time.Millisecond, 2)

	// 在第一个窗口内用完配额
	if !limiter.Allow("user1") {
		t.Error("expected 1st request to be allowed")
	}
	if !limiter.Allow("user1") {
		t.Error("expected 2nd request to be allowed")
	}
	if limiter.Allow("user1") {
		t.Error("expected 3rd request to be denied")
	}

	// 等待窗口切换
	time.Sleep(150 * time.Millisecond)

	// 新窗口应允许请求
	if !limiter.Allow("user1") {
		t.Error("expected request in new window to be allowed")
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

func TestLeakyBucketRateLimiter_Close(t *testing.T) {
	t.Parallel()

	limiter := NewLeakyBucketRateLimiter(10, 1*time.Second)

	// 验证Close不会panic
	limiter.Close()
}

func TestFixedWindowCounterRateLimiter_Close(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowCounterRateLimiter(1*time.Second, 10)

	// 验证Close不会panic
	limiter.Close()
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

func TestRateLimitFilter_GetClientIP(t *testing.T) {
	t.Parallel()

	t.Run("from remote address", func(t *testing.T) {
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
	})

	t.Run("from X-Forwarded-For with trusted proxy", func(t *testing.T) {
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
	})

	t.Run("ignore headers without trusted proxy", func(t *testing.T) {
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
	})
}
