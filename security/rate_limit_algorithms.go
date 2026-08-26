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

func NewLeakyBucketRateLimiter(capacity int, rate time.Duration) *LeakyBucketRateLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	if rate <= 0 {
		rate = 100 * time.Millisecond
	}
	l := &LeakyBucketRateLimiter{
		capacity: capacity,
		rate:     rate,
		buckets:  make(map[string]*leakyBucket),
		done:     make(chan struct{}),
	}
	newLeakyBucketCleanup(l)
	return l
}

func newLeakyBucketCleanup(l *LeakyBucketRateLimiter) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[rate_limit] leaky bucket cleanup panic: %v\n", r)
			}
		}()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Cleanup()
			case <-l.done:
				return
			}
		}
	}()
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
			tokens:   1,
			lastLeak: now,
		}
		return true
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

func NewFixedWindowCounterRateLimiter(windowSize time.Duration, maxRequests int) *FixedWindowCounterRateLimiter {
	if windowSize <= 0 {
		windowSize = 1 * time.Minute
	}
	if maxRequests <= 0 {
		maxRequests = 100
	}
	l := &FixedWindowCounterRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		counters:    make(map[string]*fixedWindowCounter),
		done:        make(chan struct{}),
	}
	newFixedWindowCleanup(l)
	return l
}

func newFixedWindowCleanup(l *FixedWindowCounterRateLimiter) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[rate_limit] fixed window cleanup panic: %v\n", r)
			}
		}()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Cleanup()
			case <-l.done:
				return
			}
		}
	}()
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
