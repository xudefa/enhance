// Package boot 提供应用启动器功能，用于 enhance 框架。
package boot

import (
	"context"
	"time"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/lifecycle"
)

// defaultBootConfig 返回默认启动配置。
func defaultBootConfig() *BootConfig {
	return &BootConfig{
		AppName:     "enhance-app",
		Version:     "1.0.0",
		ConfigType:  "json",
		AutoExecute: true,
		Starters:    true,
	}
}

// WithConfigLocation 设置配置文件路径
func WithConfigLocation(location string) BootOption {
	return func(cfg *BootConfig) {
		cfg.ConfigLocation = location
	}
}

// WithProfiles 设置激活的 Profile
func WithProfiles(profiles ...string) BootOption {
	return func(cfg *BootConfig) {
		cfg.Profiles = append(cfg.Profiles, profiles...)
	}
}

// WithAppName 设置应用名称
func WithAppName(name string) BootOption {
	return func(cfg *BootConfig) {
		cfg.AppName = name
	}
}

// WithVersion 设置版本号
func WithVersion(version string) BootOption {
	return func(cfg *BootConfig) {
		cfg.Version = version
	}
}

// WithoutAutoConfig 禁用自动配置执行
func WithoutAutoConfig() BootOption {
	return func(cfg *BootConfig) {
		cfg.AutoExecute = false
	}
}

// WithoutStarters 禁用启动器自动管理
func WithoutStarters() BootOption {
	return func(cfg *BootConfig) {
		cfg.Starters = false
	}
}

// WithExclude 排除指定的自动配置
//
// 参数:
//   - configs: 需要排除的自动配置类型名列表（如 "*DatabaseAutoConfiguration"）
//
// 示例:
//
//	boot.Run(
//	    boot.WithExclude("DatabaseStarter"),
//	    boot.WithAutoConfigRegistry(registry),
//	)
func WithExclude(configs ...string) BootOption {
	return func(cfg *BootConfig) {
		cfg.ExcludedAutoConfigs = append(cfg.ExcludedAutoConfigs, configs...)
	}
}

// WithConfigType 设置配置文件类型（如 json），为空时使用默认值。
func WithConfigType(configType string) BootOption {
	return func(cfg *BootConfig) {
		if configType != "" {
			cfg.ConfigType = configType
		}
	}
}

// WithPropertySource 添加自定义配置源，优先级最高。
func WithPropertySource(source environment.PropertySource) BootOption {
	return func(cfg *BootConfig) {
		cfg.CustomPropertySources = append(cfg.CustomPropertySources, source)
	}
}

// WithProperty 添加单个配置属性，使用内置的 MapPropertySource。
//
// 参数:
//   - key: 配置键（如 "tracing.enabled"）
//   - value: 配置值（任意类型）
//
// 示例:
//
//	app, err := boot.NewApplication(
//	    boot.WithAppName("my-app"),
//	    boot.WithProperty("tracing.enabled", "true"),
//	    boot.WithProperty("server.port", "8080"),
//	)
func WithProperty(key string, value any) BootOption {
	return func(cfg *BootConfig) {
		source := environment.NewMapPropertySource(
			"inline-property",
			environment.PriorityHigh,
			map[string]any{key: value},
		)
		cfg.CustomPropertySources = append(cfg.CustomPropertySources, source)
	}
}

// WithProperties 批量添加配置属性。
//
// 参数:
//   - props: 配置键值对（必须为偶数个参数，key-value 交替）
//
// 示例:
//
//	app, err := boot.NewApplication(
//	    boot.WithAppName("my-app"),
//	    boot.WithProperties(
//	        "tracing.enabled", "true",
//	        "server.port", 8080,
//	    ),
//	)
func WithProperties(props ...any) BootOption {
	if len(props)%2 != 0 {
		panic("WithProperties requires an even number of arguments (key-value pairs)")
	}
	propMap := make(map[string]any)
	for i := 0; i < len(props); i += 2 {
		key, ok := props[i].(string)
		if !ok {
			panic("WithProperties keys must be strings")
		}
		propMap[key] = props[i+1]
	}
	return func(cfg *BootConfig) {
		source := environment.NewMapPropertySource(
			"inline-properties",
			environment.PriorityHigh,
			propMap,
		)
		cfg.CustomPropertySources = append(cfg.CustomPropertySources, source)
	}
}

// WithConfigCenter 启用配置中心
func WithConfigCenter(centerType string, addr []string, opts ...ConfigCenterOption) BootOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterEnabled = true
		cfg.ConfigCenterType = centerType
		cfg.ConfigCenterAddr = addr
		cfg.ConfigCenterTimeout = 5 * time.Second
		for _, opt := range opts {
			opt(cfg)
		}
	}
}

// ConfigCenterOption 配置中心选项函数
type ConfigCenterOption func(*BootConfig)

// WithConfigCenterDataID 设置配置中心数据ID
func WithConfigCenterDataID(dataID string) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterDataID = dataID
	}
}

// WithConfigCenterGroup 设置配置中心分组
func WithConfigCenterGroup(group string) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterGroup = group
	}
}

// WithConfigCenterPrefix 设置配置中心前缀
func WithConfigCenterPrefix(prefix string) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterPrefix = prefix
	}
}

// WithConfigCenterTimeout 设置配置中心超时时间
func WithConfigCenterTimeout(timeout time.Duration) ConfigCenterOption {
	return func(cfg *BootConfig) {
		cfg.ConfigCenterTimeout = timeout
	}
}

// WithHook 添加生命周期钩子
func WithHook(hook lifecycle.Hook) BootOption {
	return func(cfg *BootConfig) {
		cfg.Hooks = append(cfg.Hooks, hook)
	}
}

// WithHookFunc 通过函数添加生命周期钩子
func WithHookFunc(onInit, onStart, onStop func(context.Context) error) BootOption {
	return func(cfg *BootConfig) {
		cfg.Hooks = append(cfg.Hooks, lifecycle.NewHookFunc(onInit, onStart, onStop))
	}
}
