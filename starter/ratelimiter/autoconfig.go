// Package ratelimiter 提供限流器自动配置。
//
// RateLimiter 是 Go 标准库提供的令牌桶限流器。
// 本模块提供自动配置支持，支持动态调整限流参数。
//
// 功能特性：
//   - 自动配置限流器
//   - 支持令牌桶算法
//   - 支持可配置速率和突发
//   - 支持阻塞和非阻塞模式
//   - 支持动态调整参数
//
// 配置示例：
//
//	{
//	  "ratelimiter": {
//	    "enabled": true,
//	    "rate": 10.0,
//	    "burst": 20
//	  }
//	}
//
// 使用示例：
//
//	limiter := core.MustGetBean[*rate.Limiter](app.Container())
//	if !limiter.Allow() {
//	    return errors.New("请求过于频繁")
//	}
package ratelimiter

import (
	"context"
	"fmt"
	"reflect"

	"golang.org/x/time/rate"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// init 注册 RateLimiter 自动配置类。
// 当配置 ratelimiter.enabled=true 时自动触发配置。
func init() {
	boot.RegisterAutoConfigWith(&RateLimiterAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(RateLimiterEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityMiddleware)),
	)
}

// RateLimiterAutoConfiguration 限流器自动配置类。
// 负责初始化 rate.Limiter 实例并注册到 IoC 容器。
type RateLimiterAutoConfiguration struct {
	logger  log.Logger         // 日志记录器
	limiter *rate.Limiter      // 限流器实例
	config  *RateLimiterConfig // 限流器配置信息
}

// Configure 配置限流器。
// 创建 rate.Limiter 实例，使用令牌桶算法进行限流。
func (c *RateLimiterAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 获取日志记录器，如果不存在则使用默认日志器
	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	// 加载配置
	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 RateLimiter 配置失败: %w", err)
	}

	c.config = cfg

	// 创建限流器实例
	// rate: 每秒允许的请求数
	// burst: 最大突发请求数（令牌桶容量）
	c.limiter = rate.NewLimiter(rate.Limit(cfg.Rate), cfg.Burst)

	// 注册限流器实例到 IoC 容器
	if err := ctx.Container().RegisterInstance(c.limiter, reflect.TypeFor[*rate.Limiter]()); err != nil {
		return fmt.Errorf("注册 RateLimiter 实例失败: %w", err)
	}

	c.logger.Info(context.Background(), "RateLimiter 限流器已配置",
		log.KeyValue{Key: "rate", Value: cfg.Rate},
		log.KeyValue{Key: "burst", Value: cfg.Burst},
	)

	return nil
}

// GetLimiter 获取限流器实例。
// 返回底层的 *rate.Limiter 实例，可用于高级操作。
func (c *RateLimiterAutoConfiguration) GetLimiter() *rate.Limiter {
	return c.limiter
}

// Allow 检查是否允许请求。
// 非阻塞模式，立即返回是否允许请求。
// 如果令牌桶中有可用令牌，返回 true 并消耗一个令牌。
// 如果令牌桶为空，返回 false。
//
// 使用示例：
//
//	if !limiter.Allow() {
//	    return errors.New("请求过于频繁")
//	}
func (c *RateLimiterAutoConfiguration) Allow() bool {
	return c.limiter.Allow()
}

// Wait 等待直到允许请求。
// 阻塞模式，等待直到获得令牌。
// 如果上下文被取消，返回错误。
//
// 使用示例：
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	err := limiter.Wait(ctx)
func (c *RateLimiterAutoConfiguration) Wait(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

// SetRate 动态设置速率。
// 在运行时调整每秒允许的请求数。
//
// 使用示例：
//
//	limiter.SetRate(50.0) // 调整为每秒 50 个请求
func (c *RateLimiterAutoConfiguration) SetRate(r float64) {
	c.limiter.SetLimit(rate.Limit(r))
	c.logger.Info(context.Background(), "RateLimiter 速率已更新",
		log.KeyValue{Key: "rate", Value: r},
	)
}

// SetBurst 动态设置突发值。
// 在运行时调整令牌桶的最大容量。
//
// 使用示例：
//
//	limiter.SetBurst(100) // 调整为最大突发 100 个请求
func (c *RateLimiterAutoConfiguration) SetBurst(b int) {
	c.limiter.SetBurst(b)
	c.logger.Info(context.Background(), "RateLimiter 突发值已更新",
		log.KeyValue{Key: "burst", Value: b},
	)
}

// Reserve 预约令牌。
// 返回一个 Reservation 对象，表示何时可以获得令牌。
// 适用于需要精确控制请求时间的场景。
//
// 使用示例：
//
//	res := limiter.Reserve()
//	if !res.OK() {
//	    // 无法获得令牌
//	    return
//	}
//	time.Sleep(res.Delay()) // 等待延迟
//	// 执行请求
func (c *RateLimiterAutoConfiguration) Reserve() *rate.Reservation {
	return c.limiter.Reserve()
}

// RateLimiterConfig 限流器配置。
// 包含限流器的所有可配置参数。
type RateLimiterConfig struct {
	Enabled bool    `json:"enabled" mapstructure:"enabled"` // 是否启用限流器
	Rate    float64 `json:"rate" mapstructure:"rate"`       // 每秒允许的请求数
	Burst   int     `json:"burst" mapstructure:"burst"`     // 最大突发请求数
}

// 配置常量。
const (
	RateLimiterEnabled = "ratelimiter.enabled" // 启用条件配置键
	DefaultRate        = 10.0                  // 默认每秒请求数
	DefaultBurst       = 20                    // 默认最大突发数
	ConditionTrue      = "true"                // 条件真值
)

// loadConfig 从 Environment 加载 RateLimiter 配置。
// 使用默认值初始化配置，然后从配置中心绑定用户自定义值。
func (c *RateLimiterAutoConfiguration) loadConfig(env *environment.Environment) (*RateLimiterConfig, error) {
	cfg := &RateLimiterConfig{
		Rate:  DefaultRate,
		Burst: DefaultBurst,
	}

	if err := env.BindPrefix("ratelimiter", cfg); err != nil {
		return nil, fmt.Errorf("绑定 RateLimiter 配置失败: %w", err)
	}

	return cfg, nil
}
