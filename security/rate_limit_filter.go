package security

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

// TokenBucket 令牌桶
type TokenBucket struct {
	capacity int
	tokens   int
	rate     int
	mu       sync.Mutex
	lastTime time.Time
}

func NewTokenBucket(capacity, rate int) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens:   capacity,
		rate:     rate,
		lastTime: time.Now(),
	}
}

func (b *TokenBucket) Take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled      bool
	Rate         int
	Burst        int
	ExcludePaths []string
	Log          log.Logger
}

// RateLimitFilter 限流过滤器
type RateLimitFilter struct {
	config       RateLimitConfig
	buckets      sync.Map
	globalBucket *TokenBucket
	logger       log.Logger
}

func NewRateLimitFilter(config RateLimitConfig) *RateLimitFilter {
	if config.Rate == 0 {
		config.Rate = 100
	}
	if config.Burst == 0 {
		config.Burst = 200
	}
	return &RateLimitFilter{
		config:       config,
		globalBucket: NewTokenBucket(config.Burst, config.Rate),
		logger:       config.Log,
	}
}

// DoFilter 实现 filter.Filter 接口
func (f *RateLimitFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, _ := ctx.(context.Context)
	req, _ := request.(SecurityRequest)
	resp, _ := response.(SecurityResponse)
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
	if !bucket.(*TokenBucket).Take() {
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
	return ""
}
