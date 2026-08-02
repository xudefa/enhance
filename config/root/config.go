package config

import (
	"fmt"

	"github.com/xudefa/enhance/config/environment"
)

// configImpl 配置实现，委托给 environment.Environment。
type configImpl struct {
	env *environment.Environment
}

// NewConfig 创建配置对象。
//
// 参数:
//   - env: 环境配置实例
//
// 返回:
//   - Config: 配置接口实现
func NewConfig(env *environment.Environment) Config {
	return &configImpl{env: env}
}

// GetProperty 获取指定配置属性的字符串值。
func (c *configImpl) GetProperty(key string) string {
	val, ok := c.env.GetProperty(key)
	if !ok {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetPropertyWithDefault 获取配置属性的值，如果不存在返回默认值。
func (c *configImpl) GetPropertyWithDefault(key, defaultValue string) string {
	return c.env.GetPropertyWithDefault(key, defaultValue)
}

// ContainsProperty 检查配置中是否存在指定属性。
func (c *configImpl) ContainsProperty(key string) bool {
	return c.env.ContainsProperty(key)
}

// GetRequiredProperty 获取必需的配置属性值，如果不存在返回错误。
func (c *configImpl) GetRequiredProperty(key string) (string, error) {
	val, err := c.env.GetRequiredProperty(key)
	if err != nil {
		return "", err
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("property %q is not a string type", key)
}

// Environment 返回底层环境配置实例。
func (c *configImpl) Environment() *environment.Environment {
	return c.env
}
