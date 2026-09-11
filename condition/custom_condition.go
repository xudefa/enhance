package condition

import "fmt"

// customCondition 自定义条件实现
type customCondition struct {
	name      string                          // 条件名称
	evaluator func(ctx ConditionContext) bool // 评估函数
}

// Custom 创建自定义条件
//
// 通过传入的评估函数决定是否匹配，name 用于日志和调试输出。
func Custom(name string, evaluator func(ctx ConditionContext) bool) Condition {
	return &customCondition{name: name, evaluator: evaluator}
}

// Matches 实现 Condition 接口
func (c *customCondition) Matches(ctx ConditionContext) bool {
	return c.evaluator(ctx)
}

// String 返回条件的字符串表示
func (c *customCondition) String() string {
	return fmt.Sprintf("Custom(%s)", c.name)
}
