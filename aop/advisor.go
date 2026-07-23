// Package aop 提供面向切面编程（AOP）支持。
package aop

// defaultAdvisor 默认通知器实现。
type defaultAdvisor struct {
	advice   Advice
	pointCut PointCut
	order    int
}

// NewAdvisor 创建默认通知器。
//
// 参数:
//   - advice: 通知实例
//   - pointCut: 切点定义
//   - order: 执行顺序，值越小优先级越高
//
// 返回值:
//   - Advisor: 通知器实例
func NewAdvisor(advice Advice, pointCut PointCut, order int) Advisor {
	return &defaultAdvisor{
		advice:   advice,
		pointCut: pointCut,
		order:    order,
	}
}

func (a *defaultAdvisor) Advice() Advice {
	return a.advice
}

func (a *defaultAdvisor) PointCut() PointCut {
	return a.pointCut
}

func (a *defaultAdvisor) Order() int {
	return a.order
}
