package condition

import "fmt"

// missingPropertyCondition 基于属性不存在的条件实现
type missingPropertyCondition struct {
	key string // 配置键名
}

// OnMissingProperty 创建基于属性不存在的条件
//
// 当指定的配置键不存在时匹配。
func OnMissingProperty(key string) Condition {
	return &missingPropertyCondition{key: key}
}

// Matches 实现 Condition 接口
func (m *missingPropertyCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	_, ok := env.GetProperty(m.key)
	return !ok
}

// String 返回条件的字符串表示
func (m *missingPropertyCondition) String() string {
	return fmt.Sprintf("OnMissingProperty(%s)", m.key)
}
