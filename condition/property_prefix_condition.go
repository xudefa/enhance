package condition

import (
	"fmt"
	"strings"

	"github.com/xudefa/enhance/config/environment"
)

// propertyPrefixCondition 基于配置前缀的条件实现
type propertyPrefixCondition struct {
	prefix string // 配置键前缀
}

// OnPropertyPrefix 创建基于配置前缀存在的条件
//
// 当存在以指定前缀开头的配置键时匹配。
func OnPropertyPrefix(prefix string) Condition {
	return &propertyPrefixCondition{prefix: prefix}
}

// Matches 实现 Condition 接口
func (p *propertyPrefixCondition) Matches(ctx ConditionContext) bool {
	env := ctx.Environment()
	type propertySourceLister interface {
		GetPropertySources() []environment.PropertySource
	}
	if lister, ok := env.(propertySourceLister); ok {
		for _, src := range lister.GetPropertySources() {
			for _, key := range listKeys(src) {
				if strings.HasPrefix(key, p.prefix) {
					return true
				}
			}
		}
	}
	return false
}

// String 返回条件的字符串表示
func (p *propertyPrefixCondition) String() string {
	return fmt.Sprintf("OnPropertyPrefix(%s)", p.prefix)
}
