package security

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

// TokenBucket 令牌桶
type TokenBucket struct {
	capacity   int
	tokens     int
	rate       int
	mu         sync.Mutex
	lastTime   time.Time
	lastAccess time.Time
}

func NewTokenBucket(capacity, rate int) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastTime:   now,
		lastAccess: now,
	}
}

func (b *TokenBucket) Take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	b.lastAccess = now
	elapsed := now.Sub(b.lastTime).Seconds()
	newTokens := int(elapsed * float64(b.rate))

	if newTokens > 0 {
		b.tokens = min(b.capacity, b.tokens+newTokens)
		b.lastTime = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// IsExpired 判断令牌桶是否已超过 ttl 未被访问。
func (b *TokenBucket) IsExpired(ttl time.Duration) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return time.Since(b.lastAccess) > ttl
}

const (
	defaultBucketIdleTimeout = 5 * time.Minute
	defaultCleanupInterval   = 1 * time.Minute
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool
	Rate              int
	Burst             int
	ExcludePaths      []string
	Log               log.Logger
	TrustProxyHeaders bool          // 是否信任 X-Forwarded-For 等代理头
	TrustedProxies    []string      // 可信代理 IP 或 CIDR 列表，仅从这些代理获取真实客户端 IP
	BucketIdleTimeout time.Duration // 空闲桶清理超时，防止内存泄漏
	CleanupInterval   time.Duration // 过期桶清理周期
}

// RateLimitFilter 限流过滤器
type RateLimitFilter struct {
	config           RateLimitConfig
	buckets          sync.Map
	globalBucket     *TokenBucket
	logger           log.Logger
	trustedProxyNets []*net.IPNet
	done             chan struct{}
	closeOnce        sync.Once
}

func NewRateLimitFilter(config RateLimitConfig) *RateLimitFilter {
	if config.Rate == 0 {
		config.Rate = 100
	}
	if config.Burst == 0 {
		config.Burst = 200
	}
	if config.BucketIdleTimeout <= 0 {
		config.BucketIdleTimeout = defaultBucketIdleTimeout
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = defaultCleanupInterval
	}
	f := &RateLimitFilter{
		config:           config,
		globalBucket:     NewTokenBucket(config.Burst, config.Rate),
		logger:           config.Log,
		trustedProxyNets: parseTrustedProxies(config.TrustedProxies),
		done:             make(chan struct{}),
	}
	newBucketCleanup(f)
	return f
}

// newBucketCleanup 启动过期限流桶后台清理协程。
func newBucketCleanup(f *RateLimitFilter) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[rate_limit] bucket cleanup panic: %v\n", r)
			}
		}()
		ticker := time.NewTicker(f.config.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f.cleanupBuckets()
			case <-f.done:
				return
			}
		}
	}()
}

// cleanupBuckets 删除超过空闲超时未使用的限流桶。
func (f *RateLimitFilter) cleanupBuckets() {
	f.buckets.Range(func(key, value any) bool {
		if tb, ok := value.(*TokenBucket); ok && tb.IsExpired(f.config.BucketIdleTimeout) {
			f.buckets.Delete(key)
		}
		return true
	})
}

// Close 停止后台清理协程。
func (f *RateLimitFilter) Close() {
	f.closeOnce.Do(func() {
		close(f.done)
	})
}

// DoFilter 实现 filter.Filter 接口
func (f *RateLimitFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("RateLimitFilter: ctx must be context.Context")
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("RateLimitFilter: request must be SecurityRequest")
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("RateLimitFilter: response must be SecurityResponse")
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *RateLimitFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	if !f.config.Enabled {
		return chain.DoFilter(ctx, request, response)
	}

	uri := request.GetURI()
	for _, path := range f.config.ExcludePaths {
		if strings.HasPrefix(uri, path) {
			f.logger.Debug(ctx, "请求路径被排除在限流之外", log.KeyValue{Key: "uri", Value: uri})
			return chain.DoFilter(ctx, request, response)
		}
	}

	clientIP := f.getClientIP(request)

	if clientIP == "" {
		if !f.globalBucket.Take() {
			f.logger.Warn(ctx, "全局限流触发", log.KeyValue{Key: "uri", Value: uri})
			response.SetStatusCode(429)
			if writeErr := response.Write([]byte(`{"error":"rate limited","message":"too many requests"}`)); writeErr != nil {
				f.logger.Error(ctx, "写入限流响应失败", log.KeyValue{Key: "error", Value: writeErr.Error()})
			}
			return nil
		}
		return chain.DoFilter(ctx, request, response)
	}

	bucket, _ := f.buckets.LoadOrStore(clientIP, NewTokenBucket(f.config.Burst, f.config.Rate))
	tb, _ := bucket.(*TokenBucket)
	if !tb.Take() {
		f.logger.Warn(ctx, "客户端限流触发", log.KeyValue{Key: "ip", Value: clientIP}, log.KeyValue{Key: "uri", Value: uri})
		response.SetStatusCode(429)
		if writeErr := response.Write([]byte(`{"error":"rate limited","message":"too many requests"}`)); writeErr != nil {
			f.logger.Error(ctx, "写入限流响应失败", log.KeyValue{Key: "error", Value: writeErr.Error()})
		}
		return nil
	}

	f.logger.Debug(ctx, "请求通过限流检查", log.KeyValue{Key: "ip", Value: clientIP}, log.KeyValue{Key: "uri", Value: uri})
	return chain.DoFilter(ctx, request, response)
}

// Order 实现 filter.Filter 接口
func (f *RateLimitFilter) Order() int { return 0 }

func (f *RateLimitFilter) getClientIP(request SecurityRequest) string {
	remote := parseRemoteIP(request.RemoteAddress())

	// 仅信任来自已知代理的转发头，防止客户端伪造 IP 绕过限流
	if f.config.TrustProxyHeaders && isTrustedProxy(remote, f.trustedProxyNets) {
		headers := []string{"X-Forwarded-For", "X-Real-IP", "Proxy-Client-IP", "WL-Proxy-Client-IP"}
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

	return remote
}

// parseTrustedProxies 将可信代理配置解析为 IPNet 列表，支持 IP 与 CIDR。
func parseTrustedProxies(proxies []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(proxies))
	for _, proxy := range proxies {
		proxy = strings.TrimSpace(proxy)
		if proxy == "" {
			continue
		}
		if strings.Contains(proxy, "/") {
			if _, ipnet, err := net.ParseCIDR(proxy); err == nil {
				nets = append(nets, ipnet)
			}
			continue
		}
		ip := net.ParseIP(proxy)
		if ip == nil {
			continue
		}
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		nets = append(nets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
	}
	return nets
}

// isTrustedProxy 判断 remote IP 是否在可信代理列表中。
func isTrustedProxy(ip string, trustedProxyNets []*net.IPNet) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, n := range trustedProxyNets {
		if n.Contains(parsed) {
			return true
		}
	}
	return false
}

// parseRemoteIP 从 "host:port" 中提取客户端 IP，非合法 IP 时返回空串。
func parseRemoteIP(remote string) string {
	if remote == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(remote); err == nil {
		remote = host
	}
	ip := net.ParseIP(remote)
	if ip == nil {
		return ""
	}
	return ip.String()
}
