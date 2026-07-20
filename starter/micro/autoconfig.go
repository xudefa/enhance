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

func init() {
	boot.RegisterAutoConfigWith(&MicroAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(MicroEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityBusinessLayer)),
	)
}

// MicroAutoConfiguration go-micro 微服务框架自动配置类。
type MicroAutoConfiguration struct {
	logger  log.Logger
	service micro.Service
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
		return fmt.Errorf("加载 go-micro 配置失败: %w", err)
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

	if err := ctx.Container().RegisterInstance(c.service, reflect.TypeFor[micro.Service]()); err != nil {
		return fmt.Errorf("注册 go-micro Service 失败: %w", err)
	}

	c.logger.Info(context.Background(), "go-micro 微服务已配置",
		log.KeyValue{Key: "service_name", Value: cfg.ServiceName},
		log.KeyValue{Key: "version", Value: cfg.Version},
	)

	return nil
}

// Start 启动 go-micro 微服务。
func (c *MicroAutoConfiguration) Start() error {
	c.logger.Info(context.Background(), "go-micro 微服务启动中",
		log.KeyValue{Key: "service_name", Value: c.service.Server().Options().Name},
	)
	return c.service.Run()
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
		return nil, fmt.Errorf("绑定 go-micro 配置失败: %w", err)
	}

	return cfg, nil
}
