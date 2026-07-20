// Package security 提供安全功能支持，用于 enhance 框架。
package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SlidingWindowRateLimiter 滑动窗口限流器
// 算法原理：记录每个请求的时间戳，只统计窗口时间范围内的请求数量
// 优点：避免固定窗口的临界问题，限流更平滑
// 适用场景：需要精确控制请求频率的场景
type SlidingWindowRateLimiter struct {
	windowSize  time.Duration
	maxRequests int
	mu          sync.RWMutex
	windows     map[string]*slidingWindow
}

type slidingWindow struct {
	requests []time.Time
}

// NewSlidingWindowRateLimiter 创建滑动窗口限流器。
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

// Allow 判断是否允许请求。
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

	// 清理窗口外的请求
	validIdx := 0
	for _, t := range window.requests {
		if t.After(windowStart) {
			window.requests[validIdx] = t
			validIdx++
		}
	}
	window.requests = window.requests[:validIdx]

	// 检查是否超过限制
	if len(window.requests) < r.maxRequests {
		window.requests = append(window.requests, now)
		return true
	}

	return false
}

// Cleanup 清理过期窗口。
func (r *SlidingWindowRateLimiter) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.windowSize)
	for key, window := range r.windows {
		// 清理窗口外的请求
		validIdx := 0
		for _, t := range window.requests {
			if t.After(windowStart) {
				window.requests[validIdx] = t
				validIdx++
			}
		}
		window.requests = window.requests[:validIdx]

		// 如果窗口为空，删除该 key
		if len(window.requests) == 0 {
			delete(r.windows, key)
		}
	}
}

// LeakyBucketRateLimiter 漏桶限流器
// 算法原理：请求像水滴进入桶中，桶以固定速率漏水（处理请求）
// 优点：平滑处理请求，防止突发流量
// 适用场景：需要均匀处理请求的场景
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

// NewLeakyBucketRateLimiter 创建漏桶限流器。
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

// Allow 判断是否允许请求。
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

// Cleanup 清理空桶。
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
// 算法原理：将时间划分为固定窗口，每个窗口内计数请求数量
// 优点：实现简单，内存占用小
// 缺点：存在窗口临界问题（窗口切换时可能允许2倍流量）
// 适用场景：对限流精度要求不高的场景
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

// NewFixedWindowCounterRateLimiter 创建固定窗口计数器限流器。
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

// Allow 判断是否允许请求。
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

// Cleanup 清理过期计数器。
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

// StrategyRateLimiterAdapter 限流器适配器。
type StrategyRateLimiterAdapter struct {
	limiter RateLimiter
}

// NewStrategyRateLimiterAdapter 创建适配器。
func NewStrategyRateLimiterAdapter(limiter RateLimiter) *StrategyRateLimiterAdapter {
	return &StrategyRateLimiterAdapter{limiter: limiter}
}

// Allow 判断是否允许请求。
func (a *StrategyRateLimiterAdapter) Allow(key string) bool {
	return a.limiter.Allow(key)
}

// EnhancedRateLimitFilter 增强版限流过滤器
// 职责：基于策略模式的限流过滤器，支持自定义限流策略和回调
// 与RateLimitFilter的区别：使用RateLimitStrategy接口，更灵活
type EnhancedRateLimitFilter struct {
	strategy     RateLimitStrategy
	excludePaths []string
	onRateLimit  func(ctx context.Context, request SecurityRequest, response SecurityResponse)
}

// EnhancedRateLimitOption 增强限流配置选项。
type EnhancedRateLimitOption func(*EnhancedRateLimitFilter)

// WithExcludePaths 设置排除路径。
func WithExcludePaths(paths ...string) EnhancedRateLimitOption {
	return func(f *EnhancedRateLimitFilter) {
		f.excludePaths = append(f.excludePaths, paths...)
	}
}

// WithOnRateLimit 设置限流回调。
func WithOnRateLimit(fn func(ctx context.Context, request SecurityRequest, response SecurityResponse)) EnhancedRateLimitOption {
	return func(f *EnhancedRateLimitFilter) {
		f.onRateLimit = fn
	}
}

// NewEnhancedRateLimitFilter 创建增强版限流过滤器。
func NewEnhancedRateLimitFilter(strategy RateLimitStrategy, opts ...EnhancedRateLimitOption) *EnhancedRateLimitFilter {
	f := &EnhancedRateLimitFilter{
		strategy: strategy,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// DoFilter 处理限流。
func (f *EnhancedRateLimitFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
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

// maxInt 辅助函数。
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
