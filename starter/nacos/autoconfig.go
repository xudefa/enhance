// Package nacos provides Nacos configuration center auto-configuration.
package nacos

import (
	"fmt"
	"reflect"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&NacosAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(NacosEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)
}

// NacosAutoConfiguration Nacos configuration center auto-configuration.
type NacosAutoConfiguration struct {
	logger       log.Logger
	configClient config_client.IConfigClient
}

// Configure configures Nacos configuration center client.
func (c *NacosAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Nacos config: %w", err)
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(cfg.ServerAddr, cfg.Port),
	}

	cc := constant.ClientConfig{
		AppName:     cfg.AppName,
		NamespaceId: cfg.NamespaceID,
		Username:    cfg.Username,
		Password:    cfg.Password,
		TimeoutMs:   uint64(cfg.TimeoutMs),
		LogDir:      cfg.LogDir,
		CacheDir:    cfg.CacheDir,
		LogLevel:    cfg.LogLevel,
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create Nacos config client: %w", err)
	}

	c.configClient = configClient

	if err := ctx.Container().RegisterInstance(configClient, reflect.TypeFor[config_client.IConfigClient]()); err != nil {
		return fmt.Errorf("failed to register Nacos ConfigClient: %w", err)
	}

	c.logger.Info(ctx.Context(), "Nacos configuration center client configured",
		log.KeyValue{Key: "server_addr", Value: cfg.ServerAddr},
		log.KeyValue{Key: "namespace", Value: cfg.NamespaceID},
	)

	return nil
}

// GetConfig 从 Nacos 获取配置。
func (c *NacosAutoConfiguration) GetConfig(dataID, group string) (string, error) {
	content, err := c.configClient.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get Nacos config: %w", err)
	}
	return content, nil
}

// ListenConfig 监听 Nacos 配置变更。
func (c *NacosAutoConfiguration) ListenConfig(dataID, group string, onChange func(string)) error {
	return c.configClient.ListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
		OnChange: func(namespace, group, dataId, dataId2 string) {
			onChange(dataId2)
		},
	})
}

// PublishConfig 发布配置到 Nacos。
func (c *NacosAutoConfiguration) PublishConfig(dataID, group, content string) (bool, error) {
	return c.configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
		Group:   group,
		Content: content,
	})
}

// NacosConfig Nacos configuration center config.
type NacosConfig struct {
	Enabled     bool   `json:"enabled" mapstructure:"enabled"`
	ServerAddr  string `json:"server_addr" mapstructure:"server_addr"`
	Port        uint64 `json:"port" mapstructure:"port"`
	NamespaceID string `json:"namespace_id" mapstructure:"namespace_id"`
	AppName     string `json:"app_name" mapstructure:"app_name"`
	Username    string `json:"username" mapstructure:"username"`
	Password    string `json:"password" mapstructure:"password"`
	TimeoutMs   int    `json:"timeout_ms" mapstructure:"timeout_ms"`
	LogDir      string `json:"log_dir" mapstructure:"log_dir"`
	CacheDir    string `json:"cache_dir" mapstructure:"cache_dir"`
	LogLevel    string `json:"log_level" mapstructure:"log_level"`
}

// Configuration constants.
const (
	NacosEnabled       = "nacos.enabled"
	DefaultNacosAddr   = "127.0.0.1"
	DefaultNacosPort   = 8848
	DefaultNamespaceID = "public"
	DefaultTimeoutMs   = 10000
	DefaultLogLevel    = "info"
	ConditionTrue      = "true"
)

// loadConfig loads Nacos config from Environment.
func (c *NacosAutoConfiguration) loadConfig(env *environment.Environment) (*NacosConfig, error) {
	cfg := &NacosConfig{
		ServerAddr:  DefaultNacosAddr,
		Port:        DefaultNacosPort,
		NamespaceID: DefaultNamespaceID,
		TimeoutMs:   DefaultTimeoutMs,
		LogLevel:    DefaultLogLevel,
	}

	if err := env.BindPrefix("nacos", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Nacos config: %w", err)
	}

	return cfg, nil
}
