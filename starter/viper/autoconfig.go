// Package viper 提供 Viper 配置管理增强自动配置。
package viper

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&ViperAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ViperEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)
}

// ViperAutoConfiguration Viper 配置管理自动配置类。
type ViperAutoConfiguration struct {
	logger log.Logger
	viper  *viper.Viper
	config *ViperConfig
}

// Configure 配置 Viper。
func (c *ViperAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载 Viper 配置失败: %w", err)
	}

	c.config = cfg

	v := viper.New()
	v.SetConfigName(cfg.ConfigName)
	v.SetConfigType(cfg.ConfigType)
	v.AddConfigPath(cfg.ConfigPath)

	if cfg.WatchChanges {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			c.logger.Info(context.Background(), "配置文件已更改",
				log.KeyValue{Key: "file", Value: e.Name},
			)
		})
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	c.viper = v

	if err := ctx.Container().RegisterInstance(c.viper, reflect.TypeFor[*viper.Viper]()); err != nil {
		return fmt.Errorf("注册 Viper 实例失败: %w", err)
	}

	c.logger.Info(context.Background(), "Viper 配置已加载",
		log.KeyValue{Key: "config_name", Value: cfg.ConfigName},
		log.KeyValue{Key: "config_path", Value: cfg.ConfigPath},
	)

	return nil
}

// GetViper 获取 Viper 实例。
func (c *ViperAutoConfiguration) GetViper() *viper.Viper {
	return c.viper
}

// GetString 获取字符串配置值。
func (c *ViperAutoConfiguration) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetInt 获取整数配置值。
func (c *ViperAutoConfiguration) GetInt(key string) int {
	return c.viper.GetInt(key)
}

// GetBool 获取布尔配置值。
func (c *ViperAutoConfiguration) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

// ViperConfig Viper 配置管理配置。
type ViperConfig struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	ConfigName   string `json:"config-name" mapstructure:"config-name"`
	ConfigType   string `json:"config-type" mapstructure:"config-type"`
	ConfigPath   string `json:"config-path" mapstructure:"config-path"`
	WatchChanges bool   `json:"watch-changes" mapstructure:"watch-changes"`
}

// 配置常量。
const (
	ViperEnabled        = "viper.enabled"
	DefaultConfigName   = "application"
	DefaultConfigType   = "yaml"
	DefaultConfigPath   = "."
	DefaultWatchChanges = false
	ConditionTrue       = "true"
)

// loadConfig 从 Environment 加载 Viper 配置。
func (c *ViperAutoConfiguration) loadConfig(env *environment.Environment) (*ViperConfig, error) {
	cfg := &ViperConfig{
		ConfigName:   DefaultConfigName,
		ConfigType:   DefaultConfigType,
		ConfigPath:   DefaultConfigPath,
		WatchChanges: DefaultWatchChanges,
	}

	if err := env.BindPrefix("viper", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Viper 配置失败: %w", err)
	}

	return cfg, nil
}
