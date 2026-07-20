// Package apollo 提供 Apollo 配置中心自动配置。
package apollo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&ApolloAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ApolloEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)
}

// ApolloAutoConfiguration Apollo 配置中心自动配置类。
type ApolloAutoConfiguration struct {
	logger log.Logger
	client agollo.Client
}

// Configure 配置 Apollo 配置中心客户端。
func (c *ApolloAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 Apollo 配置失败: %w", err)
	}

	if !cfg.Enabled {
		c.logger.Info(context.Background(), "Apollo 配置中心未启用，跳过配置")
		return nil
	}

	apolloConfig := &config.AppConfig{
		AppID:          cfg.AppID,
		Cluster:        cfg.Cluster,
		IP:             cfg.MetaAddr,
		NamespaceName:  cfg.Namespace,
		IsBackupConfig: cfg.IsBackupConfig,
		Secret:         cfg.Secret,
	}

	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return apolloConfig, nil
	})
	if err != nil {
		return fmt.Errorf("启动 Apollo 客户端失败: %w", err)
	}

	c.client = client

	if err := ctx.Container().RegisterInstance(client, reflect.TypeFor[agollo.Client]()); err != nil {
		return fmt.Errorf("注册 Apollo Client 失败: %w", err)
	}

	c.logger.Info(context.Background(), "Apollo 配置中心客户端已配置",
		log.KeyValue{Key: "app_id", Value: cfg.AppID},
		log.KeyValue{Key: "meta_addr", Value: cfg.MetaAddr},
		log.KeyValue{Key: "namespace", Value: cfg.Namespace},
	)

	return nil
}

// GetConfig 从 Apollo 获取配置。
func (c *ApolloAutoConfiguration) GetConfig(key, namespace string) (string, error) {
	cache := c.client.GetConfigCache(namespace)
	if cache == nil {
		return "", fmt.Errorf("namespace %s 不存在", namespace)
	}

	val, err := cache.Get(key)
	if err != nil {
		return "", fmt.Errorf("key %s 不存在: %w", key, err)
	}

	if str, ok := val.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", val), nil
}

// ApolloConfig Apollo 配置中心配置。
type ApolloConfig struct {
	Enabled        bool   `json:"enabled" mapstructure:"enabled"`
	AppID          string `json:"app_id" mapstructure:"app_id"`
	Cluster        string `json:"cluster" mapstructure:"cluster"`
	MetaAddr       string `json:"meta_addr" mapstructure:"meta_addr"`
	Namespace      string `json:"namespace" mapstructure:"namespace"`
	IsBackupConfig bool   `json:"is_backup_config" mapstructure:"is_backup_config"`
	Secret         string `json:"secret" mapstructure:"secret"`
}

// 配置常量。
const (
	ApolloEnabled         = "apollo.enabled"
	DefaultCluster        = "default"
	DefaultNamespace      = "application"
	DefaultIsBackupConfig = true
	ConditionTrue         = "true"
)

// loadConfig 从 Environment 加载 Apollo 配置。
func (c *ApolloAutoConfiguration) loadConfig(env *environment.Environment) (*ApolloConfig, error) {
	cfg := &ApolloConfig{
		Cluster:        DefaultCluster,
		Namespace:      DefaultNamespace,
		IsBackupConfig: DefaultIsBackupConfig,
	}

	if err := env.BindPrefix("apollo", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Apollo 配置失败: %w", err)
	}

	return cfg, nil
}
