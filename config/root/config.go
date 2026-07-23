package config

import "github.com/xudefa/enhance/config/environment"

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

func (c *configImpl) GetPropertyWithDefault(key, defaultValue string) string {
	return c.env.GetPropertyWithDefault(key, defaultValue)
}

func (c *configImpl) ContainsProperty(key string) bool {
	return c.env.ContainsProperty(key)
}

func (c *configImpl) GetRequiredProperty(key string) (string, error) {
	val, err := c.env.GetRequiredProperty(key)
	if err != nil {
		return "", err
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", nil
}

func (c *configImpl) Environment() *environment.Environment {
	return c.env
}
