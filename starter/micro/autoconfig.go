// Package micro 提供 go-micro 微服务框架自动配置。
package micro

import (
	"context"
	"fmt"
	"reflect"

	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

var microAutoConfig = &MicroAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(microAutoConfig,
		boot.WithConditions(
			condition.OnProperty(MicroEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityBusinessLayer)),
	)
	boot.RegisterStarter(microAutoConfig)
}

// MicroAutoConfiguration go-micro 微服务框架自动配置类。
type MicroAutoConfiguration struct {
	logger  log.Logger
	service micro.Service
	ctx     context.Context
}

// Configure 配置 go-micro 微服务。
func (c *MicroAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load go-micro config: %w", err)
	}

	opts := []micro.Option{
		micro.Name(cfg.ServiceName),
		micro.Version(cfg.Version),
	}

	if cfg.RegistryAddr != "" {
		opts = append(opts, micro.Registry(
			registry.NewRegistry(
				registry.Addrs(cfg.RegistryAddr),
			),
		))
	}

	c.service = micro.NewService(opts...)

	// 存储应用上下文
	c.ctx = ctx.Context()

	if err := ctx.Container().RegisterInstance(c.service, reflect.TypeFor[micro.Service]()); err != nil {
		return fmt.Errorf("failed to register go-micro Service: %w", err)
	}

	c.logger.Info(ctx.Context(), "go-micro service configured",
		log.KeyValue{Key: "service_name", Value: cfg.ServiceName},
		log.KeyValue{Key: "version", Value: cfg.Version},
	)

	return nil
}

// Start 启动 go-micro 微服务。
func (c *MicroAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	c.logger.Info(c.ctx, "starting go-micro service...",
		log.KeyValue{Key: "service_name", Value: c.service.Server().Options().Name},
	)
	return c.service.Run()
}

// Stop 停止 go-micro 微服务。
func (c *MicroAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	var sCtx context.Context
	if ctx != nil {
		sCtx = ctx.Context()
	} else {
		sCtx = context.Background()
	}
	c.logger.Info(sCtx, "go-micro service stopped")
	return nil
}

// Name 返回启动器名称。
func (c *MicroAutoConfiguration) Name() string {
	return "MicroStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *MicroAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *MicroAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnProperty(MicroEnabled, ConditionTrue)
}

// GetService 获取 go-micro 服务实例。
func (c *MicroAutoConfiguration) GetService() micro.Service {
	return c.service
}

// MicroConfig go-micro 微服务配置。
type MicroConfig struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	ServiceName  string `json:"service_name" mapstructure:"service_name"`
	Version      string `json:"version" mapstructure:"version"`
	RegistryAddr string `json:"registry_addr" mapstructure:"registry_addr"`
}

// 配置常量。
const (
	MicroEnabled       = "micro.enabled"
	DefaultServiceName = "enhance-service"
	DefaultVersion     = "latest"
	ConditionTrue      = "true"
)

// loadConfig 从 Environment 加载 go-micro 配置。
func (c *MicroAutoConfiguration) loadConfig(env *environment.Environment) (*MicroConfig, error) {
	cfg := &MicroConfig{
		ServiceName: DefaultServiceName,
		Version:     DefaultVersion,
	}

	if err := env.BindPrefix("micro", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind go-micro config: %w", err)
	}

	return cfg, nil
}
