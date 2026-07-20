// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"cmp"
	"slices"
)

// SortAspectsByOrder 按 Order 升序排序切面列表。
//
// 使用标准库 slices.SortFunc 按 Order 升序排列切面，Order 小的在前。
// 当多个切点匹配同一方法时，通过 Order 控制通知的执行顺序。
func SortAspectsByOrder(aspects []*AspectMeta) {
	if len(aspects) < 2 {
		return
	}
	slices.SortFunc(aspects, func(a, b *AspectMeta) int {
		if a == nil && b == nil {
			return 0
		}
		if a == nil {
			return 1
		}
		if b == nil {
			return -1
		}
		return cmp.Compare(a.Order, b.Order)
	})
}
