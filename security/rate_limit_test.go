package security

import (
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
