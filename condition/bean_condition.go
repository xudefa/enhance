package condition

import "fmt"

// beanCondition 基于 Bean 存在性的条件实现
type beanCondition struct {
	beanID  string // Bean ID
	missing bool   // true 表示检查不存在
}

// OnBean 创建基于 Bean 存在的条件
//
// 当容器中存在指定 ID 的 Bean 时匹配。
func OnBean(beanID string) Condition {
	return &beanCondition{beanID: beanID, missing: false}
}

// OnMissingBean 创建基于 Bean 不存在的条件
//
// 当容器中不存在指定 ID 的 Bean 时匹配。
func OnMissingBean(beanID string) Condition {
	return &beanCondition{beanID: beanID, missing: true}
}

// Matches 实现 Condition 接口
func (b *beanCondition) Matches(ctx ConditionContext) bool {
	container := ctx.Container()
	has := container.Has(b.beanID)
	if b.missing {
		return !has
	}
	return has
}

// String 返回条件的字符串表示
func (b *beanCondition) String() string {
	if b.missing {
		return fmt.Sprintf("OnMissingBean(%s)", b.beanID)
	}
	return fmt.Sprintf("OnBean(%s)", b.beanID)
}
