// Package security 提供安全功能支持，用于 enhance 框架。
package security

import (
	"context"
	"fmt"
	"net"
	"strings"
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
	done        chan struct{}
	closeOnce   sync.Once
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
	l := &SlidingWindowRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		windows:     make(map[string]*slidingWindow),
		done:        make(chan struct{}),
	}
	newSlidingWindowCleanup(l)
	return l
}

func newSlidingWindowCleanup(l *SlidingWindowRateLimiter) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[rate_limit] sliding window cleanup panic: %v\n", r)
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

// Allow 检查指定 key 的请求是否允许通过（滑动窗口算法）。
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

// Cleanup 清理滑动窗口中过期的请求记录，释放内存。
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

// Close 关闭滑动窗口限流器，停止后台清理协程。
func (r *SlidingWindowRateLimiter) Close() {
	r.closeOnce.Do(func() {
		close(r.done)
	})
}

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

// StrategyRateLimiterAdapter 限流器适配器
type StrategyRateLimiterAdapter struct {
	limiter RateLimiter
}

func NewStrategyRateLimiterAdapter(limiter RateLimiter) *StrategyRateLimiterAdapter {
	return &StrategyRateLimiterAdapter{limiter: limiter}
}

// Allow 委托内部限流器检查指定 key 的请求是否允许通过。
func (a *StrategyRateLimiterAdapter) Allow(key string) bool {
	return a.limiter.Allow(key)
}

// EnhancedRateLimitFilter 增强版限流过滤器
type EnhancedRateLimitFilter struct {
	strategy          RateLimitStrategy
	excludePaths      []string
	onRateLimit       func(ctx context.Context, request SecurityRequest, response SecurityResponse)
	trustProxyHeaders bool
	trustedProxyNets  []*net.IPNet
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

// WithTrustedProxies 设置可信代理 IP/CIDR 列表。
// 启用后仅信任来自这些代理的转发头，防止客户端伪造 IP 绕过限流。
func WithTrustedProxies(proxies ...string) EnhancedRateLimitOption {
	return func(f *EnhancedRateLimitFilter) {
		f.trustProxyHeaders = true
		f.trustedProxyNets = parseTrustedProxies(proxies)
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
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("rate_limit: expected context.Context, got %T", ctx)
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("rate_limit: expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("rate_limit: expected SecurityResponse, got %T", response)
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *EnhancedRateLimitFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	uri := request.GetURI()
	for _, path := range f.excludePaths {
		if uri == path || (len(path) > 0 && len(uri) >= len(path) && uri[:len(path)] == path) {
			return chain.DoFilter(ctx, request, response)
		}
	}

	key := f.clientKey(request)

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

// clientKey 计算限流键，仅信任来自可信代理的转发头，避免伪造。
func (f *EnhancedRateLimitFilter) clientKey(request SecurityRequest) string {
	remote := parseRemoteIP(request.RemoteAddress())

	if f.trustProxyHeaders && isTrustedProxy(remote, f.trustedProxyNets) {
		headers := []string{"X-Real-IP", "X-Forwarded-For"}
		for _, header := range headers {
			if ip := request.GetHeader(header); ip != "" {
				parts := strings.Split(ip, ",")
				clientIP := strings.TrimSpace(parts[0])
				if net.ParseIP(clientIP) != nil {
					return clientIP
				}
			}
		}
	}

	if remote != "" {
		return remote
	}
	return "global"
}
