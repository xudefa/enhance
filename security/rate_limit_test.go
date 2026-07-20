package security

import (
	"context"
	"testing"
	"time"
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

	limiter.mu.RLock()
	count := len(limiter.windows)
	limiter.mu.RUnlock()

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

// 测试模拟实现

type testAuth struct {
	principal   string
	credentials string
}

func (a *testAuth) Principal() any        { return a.principal }
func (a *testAuth) Credentials() any      { return a.credentials }
func (a *testAuth) Authorities() []string { return []string{"ROLE_USER"} }
func (a *testAuth) Authenticated() bool   { return true }
func (a *testAuth) Name() string          { return a.principal }

type testAuthManager struct{}

func (m *testAuthManager) Authenticate(ctx context.Context, auth Authentication) (Authentication, error) {
	return &testAuth{principal: "testuser", credentials: "testpass"}, nil
}

type testUserDetailsService struct{}

func (s *testUserDetailsService) LoadUserByUsername(ctx context.Context, username string) (UserDetails, error) {
	return nil, nil
}

type testAccessDecisionManager struct{}

func (m *testAccessDecisionManager) Decide(ctx context.Context, auth Authentication, object any, attrs []string) error {
	return nil
}

type testSecurityFilter struct{}

func (f *testSecurityFilter) DoFilter(ctx context.Context, req SecurityRequest, resp SecurityResponse, chain SecurityFilterChain) error {
	return chain.DoFilter(ctx, req, resp)
}

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

	limiter.mu.RLock()
	count := len(limiter.windows)
	limiter.mu.RUnlock()

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
