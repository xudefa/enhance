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

// slidingWindowShard 分片结构
type slidingWindowShard struct {
	mu      sync.Mutex
	windows map[string]*slidingWindow
}

// SlidingWindowRateLimiter 滑动窗口限流器（分片锁优化）
type SlidingWindowRateLimiter struct {
	windowSize  time.Duration
	maxRequests int
	shards      []*slidingWindowShard
	shardCount  int
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

	shardCount := 16 // 16 分片
	shards := make([]*slidingWindowShard, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = &slidingWindowShard{
			windows: make(map[string]*slidingWindow),
		}
	}

	l := &SlidingWindowRateLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		shards:      shards,
		shardCount:  shardCount,
		done:        make(chan struct{}),
	}
	newSlidingWindowCleanup(l)
	return l
}

// getShard 根据 key 获取对应的分片
func (r *SlidingWindowRateLimiter) getShard(key string) *slidingWindowShard {
	// 简单哈希分片
	hash := uint64(0)
	for _, c := range key {
		hash = hash*31 + uint64(c)
	}
	return r.shards[hash%uint64(r.shardCount)]
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
	shard := r.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if r.maxRequests <= 0 {
		return false
	}

	now := time.Now()
	windowStart := now.Add(-r.windowSize)

	window, exists := shard.windows[key]
	if !exists {
		window = &slidingWindow{
			requests: make([]time.Time, 0, r.maxRequests),
		}
		shard.windows[key] = window
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
	now := time.Now()
	windowStart := now.Add(-r.windowSize)

	for _, shard := range r.shards {
		shard.mu.Lock()
		for key, window := range shard.windows {
			validIdx := 0
			for _, t := range window.requests {
				if t.After(windowStart) {
					window.requests[validIdx] = t
					validIdx++
				}
			}
			window.requests = window.requests[:validIdx]
			if len(window.requests) == 0 {
				delete(shard.windows, key)
			}
		}
		shard.mu.Unlock()
	}
}

// Close 关闭滑动窗口限流器，停止后台清理协程。
func (r *SlidingWindowRateLimiter) Close() {
	r.closeOnce.Do(func() {
		close(r.done)
	})
}

// WindowCount 返回当前活跃窗口数量（用于测试和监控）
func (r *SlidingWindowRateLimiter) WindowCount() int {
	count := 0
	for _, shard := range r.shards {
		shard.mu.Lock()
		count += len(shard.windows)
		shard.mu.Unlock()
	}
	return count
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
