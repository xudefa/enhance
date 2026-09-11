package condition

import "fmt"

// propertyCondition 基于配置属性的条件实现
type propertyCondition struct {
	key           string   // 配置键名
	expectedValue []string // 期望值（为空时仅检查键是否存在）
}

// OnProperty 创建基于配置属性的条件
//
// 如果只传 key，当该属性存在且不为空时匹配。
// 如果传 key 和 value，当该属性等于 value 时匹配。
func OnProperty(key string, expectedValue ...string) Condition {
	return &propertyCondition{
		key:           key,
		expectedValue: expectedValue,
	}
}

// Matches 实现 Condition 接口
func (p *propertyCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	val, ok := env.GetProperty(p.key)
	if !ok {
		return false
	}
	if len(p.expectedValue) == 0 {
		return val != nil && val != ""
	}
	return valAsString(val) == p.expectedValue[0]
}

// String 返回条件的字符串表示
func (p *propertyCondition) String() string {
	if len(p.expectedValue) > 0 {
		return fmt.Sprintf("OnProperty(%s=%s)", p.key, p.expectedValue[0])
	}
	return fmt.Sprintf("OnProperty(%s)", p.key)
}
