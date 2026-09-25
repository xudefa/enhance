package security

import (
	"fmt"
	"sync"
	"time"
)

// LeakyBucketRateLimiter 漏桶限流器
type LeakyBucketRateLimiter struct {
	capacity  int
	rate      time.Duration
	mu        sync.RWMutex
	buckets   map[string]*leakyBucket
	done      chan struct{}
	closeOnce sync.Once
}

type leakyBucket struct {
	tokens   int
	lastLeak time.Time
}

// NewLeakyBucketRateLimiter 创建漏桶限流器，自动补齐默认参数并启动后台清理。
func NewLeakyBucketRateLimiter(capacity int, rate time.Duration) *LeakyBucketRateLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	if rate <= 0 {
		rate = 100 * time.Millisecond
	}
	limiter := &LeakyBucketRateLimiter{
		capacity: capacity,
		rate:     rate,
		buckets:  make(map[string]*leakyBucket),
		done:     make(chan struct{}),
	}
	newLeakyBucketCleanup(limiter)
	return limiter
}

func newLeakyBucketCleanup(limiter *LeakyBucketRateLimiter) {
	go limiter.leakyBucketCleanupLoop()
}

func (r *LeakyBucketRateLimiter) leakyBucketCleanupLoop() {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("[rate_limit] leaky bucket cleanup panic: %v\n", rec)
		}
	}()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.Cleanup()
		case <-r.done:
			return
		}
	}
}

// Allow 检查指定 key 的请求是否允许通过（漏桶算法）。
func (r *LeakyBucketRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.capacity <= 0 {
		return false
	}

	now := time.Now()
	bucket, exists := r.buckets[key]

	if !exists {
		r.buckets[key] = &leakyBucket{
			tokens:   0,
			lastLeak: now,
		}
		if r.capacity > 0 {
			r.buckets[key].tokens = 1
			return true
		}
		return false
	}

	if r.rate <= 0 {
		r.rate = 100 * time.Millisecond
	}
	elapsed := now.Sub(bucket.lastLeak)
	leaked := int(elapsed / r.rate)
	if leaked > 0 {
		bucket.tokens = max(0, bucket.tokens-leaked)
		bucket.lastLeak = now
	}

	if bucket.tokens < r.capacity {
		bucket.tokens++
		return true
	}

	return false
}

// Cleanup 清理漏桶中过期的桶数据，释放内存。
func (r *LeakyBucketRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, bucket := range r.buckets {
		if now.Sub(bucket.lastLeak) > r.rate*time.Duration(r.capacity) {
			delete(r.buckets, key)
		}
	}
}

// Close 关闭漏桶限流器，停止后台清理协程。
func (r *LeakyBucketRateLimiter) Close() {
	r.closeOnce.Do(func() {
		close(r.done)
	})
}

// FixedWindowCounterRateLimiter 固定窗口计数器限流器
type FixedWindowCounterRateLimiter struct {
	windowSize  time.Duration
	maxRequests int
	mu          sync.RWMutex
	counters    map[string]*fixedWindowCounter
	done        chan struct{}
	closeOnce   sync.Once
}

type fixedWindowCounter struct {
	count       int
	windowStart time.Time
}

// NewFixedWindowCounterRateLimiter 创建固定窗口计数器限流器，自动补齐默认参数并启动后台清理。
func NewFixedWindowCounterRateLimiter(windowSize time.Duration, maxRequests int) *FixedWindowCounterRateLimiter {
	if windowSize <= 0 {
		windowSize = 1 * time.Minute
	}
	if maxRequests <= 0 {
		maxRequests = 100
	}
	limiter := &FixedWindowCounterRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		counters:    make(map[string]*fixedWindowCounter),
		done:        make(chan struct{}),
	}
	newFixedWindowCleanup(limiter)
	return limiter
}

func newFixedWindowCleanup(limiter *FixedWindowCounterRateLimiter) {
	go limiter.fixedWindowCleanupLoop()
}

func (r *FixedWindowCounterRateLimiter) fixedWindowCleanupLoop() {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("[rate_limit] fixed window cleanup panic: %v\n", rec)
		}
	}()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.Cleanup()
		case <-r.done:
			return
		}
	}
}

// Allow 检查指定 key 的请求是否允许通过（固定窗口计数算法）。
func (r *FixedWindowCounterRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.maxRequests <= 0 {
		return false
	}

	now := time.Now()
	counter, exists := r.counters[key]

	if !exists || now.Sub(counter.windowStart) > r.windowSize {
		r.counters[key] = &fixedWindowCounter{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if counter.count < r.maxRequests {
		counter.count++
		return true
	}

	return false
}

// Cleanup 清理固定窗口中过期的计数器数据，释放内存。
func (r *FixedWindowCounterRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, counter := range r.counters {
		if now.Sub(counter.windowStart) > r.windowSize {
			delete(r.counters, key)
		}
	}
}

// Close 关闭固定窗口计数器限流器，停止后台清理协程。
func (r *FixedWindowCounterRateLimiter) Close() {
	r.closeOnce.Do(func() {
		close(r.done)
	})
}
