// Package tracing 提供分布式链路追踪自动配置。
package tracing

import (
	"fmt"
	"reflect"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&TracingAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(TracingEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)
}

// TracingAutoConfiguration 分布式链路追踪自动配置类。
//
// 当配置文件中启用 tracing 时（tracing.enabled=true）自动生效。
// 负责创建 tracing.Tracer 实例并注册到 IoC 容器中。
// Web 框架（Gin/Fiber/Echo/Chi）会自动从容器获取 Tracer 并注册 tracing 中间件。
type TracingAutoConfiguration struct {
	logger log.Logger
	tracer *Tracer
	config *TracingConfig
}

// Configure 配置分布式链路追踪。
func (c *TracingAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load tracing config: %w", err)
	}

	c.config = cfg

	opts := []TracerOption{
		WithServiceName(cfg.ServiceName),
	}

	if cfg.SamplingRate < 1.0 {
		opts = append(opts, WithSampler(NewProbabilitySampler(cfg.SamplingRate)))
	}

	if cfg.MaxSpans > 0 {
		opts = append(opts, WithMaxSpans(cfg.MaxSpans))
	}

	c.tracer = NewTracer(opts...)

	if err := ctx.Container().RegisterInstance(c.tracer, reflect.TypeFor[*Tracer]()); err != nil {
		return fmt.Errorf("failed to register tracer: %w", err)
	}

	c.logger.Info(ctx.Context(), "tracing enabled",
		log.KeyValue{Key: "service_name", Value: cfg.ServiceName},
		log.KeyValue{Key: "sampling_rate", Value: cfg.SamplingRate},
	)

	return nil
}

// GetTracer 获取 Tracer 实例。
func (c *TracingAutoConfiguration) GetTracer() *Tracer {
	return c.tracer
}

// 配置常量。
const (
	TracingEnabled = "tracing.enabled"
	ConditionTrue  = "true"
)

// TracingConfig 链路追踪配置。
type TracingConfig struct {
	Enabled      bool    `json:"enabled" mapstructure:"enabled"`
	ServiceName  string  `json:"service_name" mapstructure:"service_name"`
	SamplingRate float64 `json:"sampling_rate" mapstructure:"sampling_rate"`
	MaxSpans     int     `json:"max_spans" mapstructure:"max_spans"`
}

// loadConfig 从 Environment 加载 Tracing 配置。
func (c *TracingAutoConfiguration) loadConfig(env *environment.Environment) (*TracingConfig, error) {
	cfg := &TracingConfig{
		ServiceName:  DefaultServiceName,
		SamplingRate: DefaultSamplingRate,
		MaxSpans:     DefaultMaxSpans,
	}

	if err := env.BindPrefix("tracing", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind tracing config: %w", err)
	}

	return cfg, nil
}
