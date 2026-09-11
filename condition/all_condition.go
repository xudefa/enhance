package condition

import "strings"

// allCondition 逻辑与复合条件
type allCondition struct {
	conditions  []Condition // 子条件列表
	description string      // 条件描述
}

// All 返回一个复合条件，当所有子条件都匹配时返回 true（逻辑与运算）。
// 执行逻辑：遍历所有子条件，依次调用 Matches 方法：
//  1. 只要有一个子条件不匹配，立即短路返回 false（不再继续判断后续条件）
//  2. 所有子条件都匹配时返回 true
//
// 该行为与编程语言中的 && 运算符一致，支持短路优化。
func All(conditions ...Condition) Condition {
	return &allCondition{conditions: conditions}
}

// AllWithOptions 返回一个带选项的复合条件
func AllWithOptions(opts ...CompositeOption) func(conditions ...Condition) Condition {
	config := &compositeConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return func(conditions ...Condition) Condition {
		return &allCondition{
			conditions:  conditions,
			description: config.description,
		}
	}
}

func (a *allCondition) Matches(ctx ConditionContext) bool {
	for _, c := range a.conditions {
		if !c.Matches(ctx) {
			return false
		}
	}
	return true
}

func (a *allCondition) String() string {
	if a.description != "" {
		return "All(" + a.description + ")"
	}
	return "All(" + joinConditions(a.conditions, ", ") + ")"
}

// joinConditions 将条件列表格式化为可读字符串，用指定的分隔符连接每个条件的 String() 输出。
func joinConditions(conditions []Condition, sep string) string {
	var result strings.Builder
	for i, c := range conditions {
		if i > 0 {
			result.WriteString(sep)
		}
		result.WriteString(c.String())
	}
	return result.String()
}
