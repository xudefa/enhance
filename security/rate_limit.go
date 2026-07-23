// Package security 提供安全功能支持，用于 enhance 框架。
package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xudefa/enhance/security/filter"
)

// SlidingWindowRateLimiter 滑动窗口限流器
type SlidingWindowRateLimiter struct {
	windowSize  time.Duration
	maxRequests int
	mu          sync.RWMutex
	windows     map[string]*slidingWindow
}

type slidingWindow struct {
	requests []time.Time
}

func NewSlidingWindowRateLimiter(windowSize time.Duration, maxRequests int) *SlidingWindowRateLimiter {
	if windowSize <= 0 {
		windowSize = 1 * time.Minute
	}
	if maxRequests < 0 {
		maxRequests = 100
	}
	return &SlidingWindowRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		windows:     make(map[string]*slidingWindow),
	}
}

func (r *SlidingWindowRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.maxRequests <= 0 {
		return false
	}

	now := time.Now()
	windowStart := now.Add(-r.windowSize)

	window, exists := r.windows[key]
	if !exists {
		window = &slidingWindow{
			requests: make([]time.Time, 0, r.maxRequests),
		}
		r.windows[key] = window
	}

	validIdx := 0
	for _, t := range window.requests {
		if t.After(windowStart) {
			window.requests[validIdx] = t
			validIdx++
		}
	}
	window.requests = window.requests[:validIdx]

	if len(window.requests) < r.maxRequests {
		window.requests = append(window.requests, now)
		return true
	}

	return false
}

func (r *SlidingWindowRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.windowSize)
	for key, window := range r.windows {
		validIdx := 0
		for _, t := range window.requests {
			if t.After(windowStart) {
				window.requests[validIdx] = t
				validIdx++
			}
		}
		window.requests = window.requests[:validIdx]
		if len(window.requests) == 0 {
			delete(r.windows, key)
		}
	}
}

// LeakyBucketRateLimiter 漏桶限流器
type LeakyBucketRateLimiter struct {
	capacity int
	rate     time.Duration
	mu       sync.RWMutex
	buckets  map[string]*leakyBucket
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
	return &LeakyBucketRateLimiter{
		capacity: capacity,
		rate:     rate,
		buckets:  make(map[string]*leakyBucket),
	}
}

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

	elapsed := now.Sub(bucket.lastLeak)
	leaked := int(elapsed / r.rate)
	if leaked > 0 {
		bucket.tokens = maxInt(0, bucket.tokens-leaked)
		bucket.lastLeak = now
	}

	if bucket.tokens < r.capacity {
		bucket.tokens++
		return true
	}

	return false
}

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

// FixedWindowCounterRateLimiter 固定窗口计数器限流器
type FixedWindowCounterRateLimiter struct {
	windowSize  time.Duration
	maxRequests int
	mu          sync.RWMutex
	counters    map[string]*fixedWindowCounter
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
	return &FixedWindowCounterRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		counters:    make(map[string]*fixedWindowCounter),
	}
}

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

// StrategyRateLimiterAdapter 限流器适配器
type StrategyRateLimiterAdapter struct {
	limiter RateLimiter
}

func NewStrategyRateLimiterAdapter(limiter RateLimiter) *StrategyRateLimiterAdapter {
	return &StrategyRateLimiterAdapter{limiter: limiter}
}

func (a *StrategyRateLimiterAdapter) Allow(key string) bool {
	return a.limiter.Allow(key)
}

// EnhancedRateLimitFilter 增强版限流过滤器
type EnhancedRateLimitFilter struct {
	strategy     RateLimitStrategy
	excludePaths []string
	onRateLimit  func(ctx context.Context, request SecurityRequest, response SecurityResponse)
}

type EnhancedRateLimitOption func(*EnhancedRateLimitFilter)

func WithExcludePaths(paths ...string) EnhancedRateLimitOption {
	return func(f *EnhancedRateLimitFilter) {
		f.excludePaths = append(f.excludePaths, paths...)
	}
}

func WithOnRateLimit(fn func(ctx context.Context, request SecurityRequest, response SecurityResponse)) EnhancedRateLimitOption {
	return func(f *EnhancedRateLimitFilter) {
		f.onRateLimit = fn
	}
}

func NewEnhancedRateLimitFilter(strategy RateLimitStrategy, opts ...EnhancedRateLimitOption) *EnhancedRateLimitFilter {
	f := &EnhancedRateLimitFilter{
		strategy: strategy,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// DoFilter 实现 filter.Filter 接口
func (f *EnhancedRateLimitFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, _ := ctx.(context.Context)
	req, _ := request.(SecurityRequest)
	resp, _ := response.(SecurityResponse)
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *EnhancedRateLimitFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	uri := request.GetURI()
	for _, path := range f.excludePaths {
		if uri == path || (len(path) > 0 && uri[:len(path)] == path) {
			return chain.DoFilter(ctx, request, response)
		}
	}

	key := request.GetHeader("X-Real-IP")
	if key == "" {
		key = request.GetHeader("X-Forwarded-For")
	}
	if key == "" {
		key = "global"
	}

	if !f.strategy.Allow(key) {
		if f.onRateLimit == nil {
			response.SetStatusCode(429)
			if writeErr := response.Write([]byte(`{"error":"rate_limited","message":"too many requests"}`)); writeErr != nil {
				fmt.Printf("[enhance] failed to write rate limit response: %v\n", writeErr)
			}
			return nil
		}
		f.onRateLimit(ctx, request, response)
		return nil
	}

	return chain.DoFilter(ctx, request, response)
}

// Order 实现 filter.Filter 接口
func (f *EnhancedRateLimitFilter) Order() int { return 0 }

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
