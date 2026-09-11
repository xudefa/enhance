package condition

import (
	"fmt"
	"strings"
)

// profileCondition 基于 Profile 的条件实现
type profileCondition struct {
	profile string // Profile 名称，支持 "!" 否定前缀
}

// OnProfile 创建基于 Profile 的条件
//
// 当指定 Profile 被激活时匹配。支持否定前缀 "!"，如 "!dev" 表示非 dev 环境时匹配。
func OnProfile(profile string) Condition {
	return &profileCondition{profile: profile}
}

// Matches 实现 Condition 接口
func (p *profileCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	type profileAcceptor interface {
		AcceptsProfile(profile string) bool
	}
	if pa, ok := env.(profileAcceptor); ok {
		return pa.AcceptsProfile(p.profile)
	}
	negate := strings.HasPrefix(p.profile, "!")
	return negate
}

// String 返回条件的字符串表示
func (p *profileCondition) String() string {
	return fmt.Sprintf("OnProfile(%s)", p.profile)
}
