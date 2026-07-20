package condition

import "fmt"

// ConditionFunc 条件函数类型
//
// 这是 Condition 接口的函数式替代方案，适用于简单条件判断。
// 无需实现接口，直接写函数即可。
//
// 示例:
//
//	// 简单条件函数
//	cond := condition.ConditionFunc(func(ctx condition.ConditionContext) bool {
//	    val, ok := ctx.GetProperty("feature.enabled")
//	    return ok && val == "true"
//	})
//
//	// 或使用内置的 OnProperty 等函数（它们已经返回 Condition）
//	cond := condition.OnProperty("feature.enabled", "true")
type ConditionFunc func(ctx ConditionContext) bool

// Matches 实现 Condition 接口
func (f ConditionFunc) Matches(ctx ConditionContext) bool {
	return f(ctx)
}

// String 返回条件的可读描述
func (f ConditionFunc) String() string {
	return "ConditionFunc(...)"
}

// Func 将普通函数包装为 Condition
//
// 便捷函数，用于快速创建条件。
//
// 示例:
//
//	cond := condition.Func(func(ctx condition.ConditionContext) bool {
//	    return os.Getenv("ENABLE_FEATURE") == "1"
//	})
func Func(fn func(ctx ConditionContext) bool) Condition {
	return ConditionFunc(fn)
}

// Always 创建始终匹配的条件
//
// 用于测试或强制启用的场景。
func Always() Condition {
	return ConditionFunc(func(ctx ConditionContext) bool { return true })
}

// Never 创建始终不匹配的条件
//
// 用于禁用某个自动配置。
func Never() Condition {
	return ConditionFunc(func(ctx ConditionContext) bool { return false })
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

// describedCondition 带描述的条件实现
type describedCondition struct {
	description string
	fn          func(ctx ConditionContext) bool
}

func (d *describedCondition) Matches(ctx ConditionContext) bool {
	return d.fn(ctx)
}

func (d *describedCondition) String() string {
	return fmt.Sprintf("When(%s)", d.description)
}

// AllFunc 创建函数类型的逻辑与条件
//
// 与 All() 等效，但直接接受函数而非 Condition 接口。
//
// 示例:
//
//	cond := condition.AllFunc(
//	    func(ctx condition.ConditionContext) bool { return ctx.HasProperty("db.url") },
//	    func(ctx condition.ConditionContext) bool { return ctx.HasProperty("db.password") },
//	)
func AllFunc(fns ...func(ctx ConditionContext) bool) Condition {
	return ConditionFunc(func(ctx ConditionContext) bool {
		for _, fn := range fns {
			if !fn(ctx) {
				return false
			}
		}
		return true
	})
}

// AnyFunc 创建函数类型的逻辑或条件
//
// 与 Any() 等效，但直接接受函数而非 Condition 接口。
//
// 示例:
//
//	cond := condition.AnyFunc(
//	    func(ctx condition.ConditionContext) bool { return ctx.HasProperty("cache.redis.url") },
//	    func(ctx condition.ConditionContext) bool { return ctx.HasProperty("cache.memcached.url") },
//	)
func AnyFunc(fns ...func(ctx ConditionContext) bool) Condition {
	return ConditionFunc(func(ctx ConditionContext) bool {
		for _, fn := range fns {
			if fn(ctx) {
				return true
			}
		}
		return false
	})
}
