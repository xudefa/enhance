package condition

// anyCondition 逻辑或复合条件
type anyCondition struct {
	conditions  []Condition // 子条件列表
	description string      // 条件描述
}

// Any 返回一个复合条件，当任一子条件匹配时返回 true（逻辑或运算）。
// 执行逻辑：遍历所有子条件，依次调用 Matches 方法：
//  1. 只要有一个子条件匹配，立即短路返回 true（不再继续判断后续条件）
//  2. 所有子条件都不匹配时返回 false
//
// 该行为与编程语言中的 || 运算符一致，支持短路优化。
func Any(conditions ...Condition) Condition {
	return &anyCondition{conditions: conditions}
}

func (a *anyCondition) Matches(ctx ConditionContext) bool {
	for _, c := range a.conditions {
		if c.Matches(ctx) {
			return true
		}
	}
	return false
}

func (a *anyCondition) String() string {
	if a.description != "" {
		return "Any(" + a.description + ")"
	}
	return "Any(" + joinConditions(a.conditions, ", ") + ")"
}
