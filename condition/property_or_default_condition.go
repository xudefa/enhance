package condition

import "fmt"

// propertyOrDefaultCondition 基于配置属性（带默认值）的条件实现
type propertyOrDefaultCondition struct {
	key           string   // 配置键名
	defaultValue  string   // 默认值（当配置键不存在时使用）
	expectedValue []string // 期望值
}

// OnPropertyOrDefault 创建基于配置属性的条件（带默认值）
//
// 当配置键存在且值匹配 expectedValue 时返回 true。
// 当配置键不存在时，使用 defaultValue 进行匹配。
//
// 这是"约定优于配置"的核心条件，允许框架提供合理的默认行为，
// 用户无需显式配置即可启用功能。
//
// 示例:
//
//	// 当 gin.enabled 未配置时，默认为 true（即默认启用）
//	condition.OnPropertyOrDefault("gin.enabled", "true", "true")
//
//	// 当 gin.enabled 未配置时，默认为 false（即默认禁用）
//	condition.OnPropertyOrDefault("gin.enabled", "false", "false")
func OnPropertyOrDefault(key string, defaultValue string, expectedValue ...string) Condition {
	return &propertyOrDefaultCondition{
		key:           key,
		defaultValue:  defaultValue,
		expectedValue: expectedValue,
	}
}

// Matches 实现 Condition 接口
func (p *propertyOrDefaultCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	propValue, ok := env.GetProperty(p.key)
	if !ok {
		// 配置键不存在，使用默认值
		propValue = p.defaultValue
	}
	if len(p.expectedValue) == 0 {
		return propValue != nil && propValue != ""
	}
	return valAsString(propValue) == p.expectedValue[0]
}

// String 返回条件的字符串表示
func (p *propertyOrDefaultCondition) String() string {
	if len(p.expectedValue) > 0 {
		return fmt.Sprintf("OnPropertyOrDefault(%s=%s, default=%s)", p.key, p.expectedValue[0], p.defaultValue)
	}
	return fmt.Sprintf("OnPropertyOrDefault(%s, default=%s)", p.key, p.defaultValue)
}
