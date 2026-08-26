package security

import (
	"testing"
	"time"
)

func TestNewLeakyBucketRateLimiter(t *testing.T) {
	t.Parallel()

	l := NewLeakyBucketRateLimiter(10, 100*time.Millisecond)
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer l.Close()
}

func TestNewLeakyBucketRateLimiter_Defaults(t *testing.T) {
	t.Parallel()

	l := NewLeakyBucketRateLimiter(0, 0)
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer l.Close()

	if l.capacity != 100 {
		t.Errorf("expected default capacity 100, got %d", l.capacity)
	}
	if l.rate != 100*time.Millisecond {
		t.Errorf("expected default rate 100ms, got %v", l.rate)
	}
}

func TestLeakyBucketRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	l := NewLeakyBucketRateLimiter(5, 10*time.Millisecond)
	defer l.Close()

	if !l.Allow("key1") {
		t.Error("expected first request to be allowed")
	}
}

func TestLeakyBucketRateLimiter_Allow_Reject(t *testing.T) {
	t.Parallel()

	l := NewLeakyBucketRateLimiter(2, 1*time.Hour)
	defer l.Close()

	l.Allow("key1")
	l.Allow("key1")
	if l.Allow("key1") {
		t.Error("expected request to be rejected after capacity exceeded")
	}
}

func TestLeakyBucketRateLimiter_CloseMethod(t *testing.T) {
	t.Parallel()

	l := NewLeakyBucketRateLimiter(10, 100*time.Millisecond)
	l.Close()

	l.Close()
}

func TestNewFixedWindowCounterRateLimiter(t *testing.T) {
	t.Parallel()

	l := NewFixedWindowCounterRateLimiter(time.Minute, 10)
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer l.Close()
}

func TestNewFixedWindowCounterRateLimiter_Defaults(t *testing.T) {
	t.Parallel()

	l := NewFixedWindowCounterRateLimiter(0, 0)
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
	defer l.Close()

	if l.windowSize != time.Minute {
		t.Errorf("expected default windowSize 1m, got %v", l.windowSize)
	}
	if l.maxRequests != 100 {
		t.Errorf("expected default maxRequests 100, got %d", l.maxRequests)
	}
}

func TestFixedWindowCounterRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	l := NewFixedWindowCounterRateLimiter(time.Minute, 3)
	defer l.Close()

	if !l.Allow("key1") {
		t.Error("expected first request to be allowed")
	}
	if !l.Allow("key1") {
		t.Error("expected second request to be allowed")
	}
	if !l.Allow("key1") {
		t.Error("expected third request to be allowed")
	}
	if l.Allow("key1") {
		t.Error("expected fourth request to be rejected")
	}
}

func TestFixedWindowCounterRateLimiter_DifferentKeys(t *testing.T) {
	t.Parallel()

	l := NewFixedWindowCounterRateLimiter(time.Minute, 1)
	defer l.Close()

	if !l.Allow("key1") {
		t.Error("expected key1 to be allowed")
	}
	if !l.Allow("key2") {
		t.Error("expected key2 to be allowed")
	}
}

func TestFixedWindowCounterRateLimiter_CloseMethod(t *testing.T) {
	t.Parallel()

	l := NewFixedWindowCounterRateLimiter(time.Minute, 10)
	l.Close()
	l.Close()
}

func TestLeakyBucketRateLimiter_CleanupFunc(t *testing.T) {
	t.Parallel()

	l := NewLeakyBucketRateLimiter(10, time.Millisecond)
	defer l.Close()

	l.Allow("key1")
	l.Cleanup()
}

func TestFixedWindowCounterRateLimiter_CleanupFunc(t *testing.T) {
	t.Parallel()

	l := NewFixedWindowCounterRateLimiter(time.Millisecond, 10)
	defer l.Close()

	l.Allow("key1")
	l.Cleanup()
}
