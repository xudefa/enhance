package ratelimiter

import (
	"testing"
	"time"

	"github.com/xudefa/enhance/config/environment"
	"golang.org/x/time/rate"
)

func TestRateLimiterConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-ratelimiter", environment.PriorityNormal, map[string]any{
		"ratelimiter.enabled": "true",
		"ratelimiter.rate":    "50",
		"ratelimiter.burst":   "100",
	}))

	cfg := &RateLimiterConfig{
		Rate:  DefaultRate,
		Burst: DefaultBurst,
	}

	err := env.BindPrefix("ratelimiter", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected ratelimiter.enabled to be true")
	}
	if cfg.Rate != 50.0 {
		t.Errorf("expected rate 50.0, got %f", cfg.Rate)
	}
	if cfg.Burst != 100 {
		t.Errorf("expected burst 100, got %d", cfg.Burst)
	}
}

func TestRateLimiterConfig_DefaultValues(t *testing.T) {
	cfg := &RateLimiterConfig{
		Rate:  DefaultRate,
		Burst: DefaultBurst,
	}

	if cfg.Rate != 10.0 {
		t.Errorf("expected default rate 10.0, got %f", cfg.Rate)
	}
	if cfg.Burst != 20 {
		t.Errorf("expected default burst 20, got %d", cfg.Burst)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	cfg := &RateLimiterConfig{
		Rate:  100.0,
		Burst: 10,
	}

	limiter := rate.NewLimiter(rate.Limit(cfg.Rate), cfg.Burst)

	// 前 10 个请求应该被允许
	for i := 0; i < 10; i++ {
		if !limiter.Allow() {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}

	// 第 11 个请求应该被拒绝（因为 burst 为 10）
	if limiter.Allow() {
		t.Error("expected 11th request to be denied")
	}

	// 等待令牌补充
	time.Sleep(100 * time.Millisecond)
	if !limiter.Allow() {
		t.Error("expected request after wait to be allowed")
	}
}
