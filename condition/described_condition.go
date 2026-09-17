package condition

import "fmt"

// describedCondition 带描述的条件实现
type describedCondition struct {
	description string
	fn          func(ctx ConditionContext) bool
}

// When 创建带描述的条件
//
// 在 ConditionFunc 基础上增加可读的 String() 输出。
//
// 示例:
//
//	cond := condition.When("feature flag is enabled", func(ctx condition.ConditionContext) bool {
//	    val, _ := ctx.GetProperty("feature.enabled")
//	    return val == "true"
//	})
func When(description string, fn func(ctx ConditionContext) bool) Condition {
	return &describedCondition{
		description: description,
		fn:          fn,
	}
}

// Matches 执行条件的判断函数并返回其结果。
func (d *describedCondition) Matches(ctx ConditionContext) bool {
	return d.fn(ctx)
}

// String 返回带描述条件的可读描述。
func (d *describedCondition) String() string {
	return fmt.Sprintf("When(%s)", d.description)
}
