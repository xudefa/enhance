// Package config 提供配置管理功能，作为 config/environment 的上层封装。
//
// 该包提供简化的配置访问接口，委托给 environment.Environment 执行实际操作。
// 依赖方向：config/root → config/environment（正确方向）。
//
// # 架构设计
//
//   - Config: 配置接口，提供简化的配置访问
//   - 委托给 environment.Environment 执行实际操作
//
// # 使用方式
//
//	env := environment.NewMapEnvironment(map[string]string{"app.name": "my-app"})
//	cfg := config.NewConfig(env)
//	name := cfg.GetProperty("app.name")
package config

import "github.com/xudefa/enhance/config/environment"

// Config 配置接口。
//
// 提供对配置的简化访问，委托给 environment.Environment 执行实际操作。
// 依赖方向：config/root → config/environment（正确方向）。
type Config interface {
	// GetProperty 获取属性值，不存在时返回空字符串。
	GetProperty(key string) string

	// GetPropertyWithDefault 获取属性值，不存在时返回默认值。
	GetPropertyWithDefault(key, defaultValue string) string

	// ContainsProperty 检查是否包含指定属性。
	ContainsProperty(key string) bool

	// GetRequiredProperty 获取必需属性，不存在时返回错误。
	GetRequiredProperty(key string) (string, error)

	// Environment 返回底层的环境配置实例。
	Environment() *environment.Environment
}
