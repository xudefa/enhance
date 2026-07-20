package config

import (
	"strings"
	"time"
)

// WithDataID 设置配置中心的数据 ID（Nacos 使用）。
func WithDataID(dataID string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.DataID = dataID
	}
}

// WithGroup 设置配置中心的分组（Nacos 使用）。
func WithGroup(group string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.Group = group
	}
}

// WithFormat 设置配置中心的前缀（Etcd/Consul 使用）。
func WithFormat(format string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.Prefix = format
	}
}

// WithProfiles 设置配置中心的命名空间（将 Profile 列表以逗号连接）。
func WithProfiles(profiles []string) func(*ConfigCenterConfig) {
	return func(cfg *ConfigCenterConfig) {
		cfg.Namespace = strings.Join(profiles, ",")
	}
}

// WithRemoteSource 创建配置中心配置
//
// 便捷函数，用于创建 ConfigCenterConfig 实例。
//
// 参数:
//   - centerType: 配置中心类型 (nacos/etcd/consul)
//   - endpoints: 配置中心地址列表
//   - opts: 可选的配置中心选项
//
// 示例:
//
//	cfg := config.WithRemoteSource("nacos", []string{"127.0.0.1:8848"},
//	    config.WithDataID("app-config"),
//	    config.WithGroup("DEFAULT_GROUP"),
//	)
//
//	cfg := config.WithRemoteSource("etcd", []string{"127.0.0.1:2379"},
//	    config.WithFormat("/config/myapp/"),
//	)
func WithRemoteSource(centerType string, endpoints []string, opts ...func(*ConfigCenterConfig)) *ConfigCenterConfig {
	cfg := &ConfigCenterConfig{
		Endpoints: endpoints,
		Timeout:   10 * time.Second,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
