// Package consul 提供 Consul 服务发现自动配置。
package consul

import (
	"context"
	"fmt"
	"reflect"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&ConsulAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ConsulEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityServiceDiscovery)),
	)
}

// ConsulAutoConfiguration Consul 服务发现自动配置类。
type ConsulAutoConfiguration struct {
	logger log.Logger
	client *consulapi.Client
	config *ConsulConfig
}

// Configure 配置 Consul 客户端。
func (c *ConsulAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 Consul 配置失败: %w", err)
	}

	c.config = cfg

	consulCfg := consulapi.DefaultConfig()
	consulCfg.Address = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if cfg.Token != "" {
		consulCfg.Token = cfg.Token
	}

	client, err := consulapi.NewClient(consulCfg)
	if err != nil {
		return fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}

	_, err = client.Status().Leader()
	if err != nil {
		return fmt.Errorf("Consul 连接失败: %w", err)
	}

	c.client = client

	if err := ctx.Container().RegisterInstance(c.client, reflect.TypeFor[*consulapi.Client]()); err != nil {
		return fmt.Errorf("注册 Consul Client 失败: %w", err)
	}

	c.logger.Info(context.Background(), "Consul 连接成功",
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "port", Value: cfg.Port},
	)

	return nil
}

// GetClient 获取 Consul Client 实例。
func (c *ConsulAutoConfiguration) GetClient() *consulapi.Client {
	return c.client
}

// RegisterService 注册服务到 Consul。
func (c *ConsulAutoConfiguration) RegisterService(service *consulapi.AgentServiceRegistration) error {
	return c.client.Agent().ServiceRegister(service)
}

// DeregisterService 从 Consul 注销服务。
func (c *ConsulAutoConfiguration) DeregisterService(serviceID string) error {
	return c.client.Agent().ServiceDeregister(serviceID)
}

// GetHealthyServices 获取健康服务列表。
func (c *ConsulAutoConfiguration) GetHealthyServices(serviceName string) ([]*consulapi.ServiceEntry, error) {
	entries, _, err := c.client.Health().Service(serviceName, "", true, nil)
	return entries, err
}

// ConsulConfig Consul 服务发现配置。
type ConsulConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Host    string `json:"host" mapstructure:"host"`
	Port    int    `json:"port" mapstructure:"port"`
	Token   string `json:"token" mapstructure:"token"`
}

// loadConfig 从 Environment 加载 Consul 配置。
func (c *ConsulAutoConfiguration) loadConfig(env *environment.Environment) (*ConsulConfig, error) {
	cfg := &ConsulConfig{
		Host: DefaultConsulHost,
		Port: DefaultConsulPort,
	}

	if err := env.BindPrefix("consul", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Consul 配置失败: %w", err)
	}

	return cfg, nil
}
