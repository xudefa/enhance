package security

import (
	"testing"
	"time"
)

func TestNewLeakyBucketRateLimiter(t *testing.T) {
	t.Parallel()

	limiter := NewLeakyBucketRateLimiter(10, 100*time.Millisecond)
	if limiter == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer limiter.Close()
}

func TestNewLeakyBucketRateLimiter_Defaults(t *testing.T) {
	t.Parallel()

	limiter := NewLeakyBucketRateLimiter(0, 0)
	if limiter == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer limiter.Close()

	if limiter.capacity != 100 {
		t.Errorf("expected default capacity 100, got %d", limiter.capacity)
	}
	if limiter.rate != 100*time.Millisecond {
		t.Errorf("expected default rate 100ms, got %v", limiter.rate)
	}
}

func TestLeakyBucketRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	limiter := NewLeakyBucketRateLimiter(5, 10*time.Millisecond)
	defer limiter.Close()

	if !limiter.Allow("key1") {
		t.Error("expected first request to be allowed")
	}
}

func TestLeakyBucketRateLimiter_Allow_Reject(t *testing.T) {
	t.Parallel()

	limiter := NewLeakyBucketRateLimiter(2, 1*time.Hour)
	defer limiter.Close()

	limiter.Allow("key1")
	limiter.Allow("key1")
	if limiter.Allow("key1") {
		t.Error("expected request to be rejected after capacity exceeded")
	}
}

func TestLeakyBucketRateLimiter_CloseMethod(t *testing.T) {
	t.Parallel()

	limiter := NewLeakyBucketRateLimiter(10, 100*time.Millisecond)
	limiter.Close()

	limiter.Close()
}

func TestNewFixedWindowCounterRateLimiter(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowCounterRateLimiter(time.Minute, 10)
	if limiter == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer limiter.Close()
}

func TestNewFixedWindowCounterRateLimiter_Defaults(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowCounterRateLimiter(0, 0)
	if limiter == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer limiter.Close()

	if limiter.windowSize != time.Minute {
		t.Errorf("expected default windowSize 1m, got %v", limiter.windowSize)
	}
	if limiter.maxRequests != 100 {
		t.Errorf("expected default maxRequests 100, got %d", limiter.maxRequests)
	}
}

func TestFixedWindowCounterRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowCounterRateLimiter(time.Minute, 3)
	defer limiter.Close()

	if !limiter.Allow("key1") {
		t.Error("expected first request to be allowed")
	}
	if !limiter.Allow("key1") {
		t.Error("expected second request to be allowed")
	}
	if !limiter.Allow("key1") {
		t.Error("expected third request to be allowed")
	}
	if limiter.Allow("key1") {
		t.Error("expected fourth request to be rejected")
	}
}

func TestFixedWindowCounterRateLimiter_DifferentKeys(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowCounterRateLimiter(time.Minute, 1)
	defer limiter.Close()

	if !limiter.Allow("key1") {
		t.Error("expected key1 to be allowed")
	}
	if !limiter.Allow("key2") {
		t.Error("expected key2 to be allowed")
	}
}

func TestFixedWindowCounterRateLimiter_CloseMethod(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowCounterRateLimiter(time.Minute, 10)
	limiter.Close()
	limiter.Close()
}

func TestLeakyBucketRateLimiter_CleanupFunc(t *testing.T) {
	t.Parallel()

	limiter := NewLeakyBucketRateLimiter(10, time.Millisecond)
	defer limiter.Close()

	limiter.Allow("key1")
	limiter.Cleanup()
}

func TestFixedWindowCounterRateLimiter_CleanupFunc(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowCounterRateLimiter(time.Millisecond, 10)
	defer limiter.Close()

	limiter.Allow("key1")
	limiter.Cleanup()
}
