package condition

// notCondition 逻辑非复合条件
type notCondition struct {
	condition Condition // 被取反的子条件
}

// Not 返回一个复合条件，对子条件的匹配结果取反（逻辑非运算）。
// 执行逻辑：调用子条件的 Matches 方法，然后对其结果取反：
//  1. 子条件匹配时返回 false
//  2. 子条件不匹配时返回 true
//
// 该行为与编程语言中的 ! 运算符一致。
func Not(condition Condition) Condition {
	return &notCondition{condition: condition}
}

// Matches 返回子条件匹配结果的取反值。
func (n *notCondition) Matches(ctx ConditionContext) bool {
	return !n.condition.Matches(ctx)
}

// String 返回逻辑非复合条件的可读描述。
func (n *notCondition) String() string {
	return "Not(" + n.condition.String() + ")"
}
