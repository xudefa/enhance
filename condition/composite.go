package condition

// CompositeOption 复合条件选项函数
type CompositeOption func(*compositeConfig)

// compositeConfig 复合条件配置
type compositeConfig struct {
	description string
	lazy        bool
}

// WithDescription 设置条件描述
func WithDescription(desc string) CompositeOption {
	return func(c *compositeConfig) {
		c.description = desc
	}
}

// WithLazyEvaluation 启用惰性求值（短路优化）
// All 和 Any 默认已经支持短路优化，此选项用于显式声明
func WithLazyEvaluation() CompositeOption {
	return func(c *compositeConfig) {
		c.lazy = true
	}
}
