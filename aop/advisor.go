// Package aop 提供面向切面编程（AOP）支持。
package aop

// advisor 顾问器内部实现。
//
// 存储切点、通知和执行顺序信息。
type advisor struct {
	pointCut PointCut // 切点定义
	advice   Advice   // 通知实例
	order    int      // 执行顺序，值越小优先级越高
}

func (a *advisor) GetPointCut() PointCut {
	return a.pointCut
}

func (a *advisor) GetAdvice() Advice {
	return a.advice
}

func (a *advisor) Order() int {
	return a.order
}

// NewAdvisor 创建顾问
//
// 参数:
//   - pointCut: 切点
//   - advice: 通知
//   - order: 可选的执行顺序，默认 0
//
// 返回值:
//   - Advisor: 顾问实例
//
// 示例:
//
//	advisor := aop.NewAdvisor(
//	    aop.MatchByName("DoSomething"),
//	    aop.Before(func(jp aop.JoinPoint) { fmt.Println("before") }),
//	    1, // order
//	)
func NewAdvisor(pointCut PointCut, advice Advice, order ...int) Advisor {
	o := 0
	if len(order) > 0 {
		o = order[0]
	}
	return &advisor{
		pointCut: pointCut,
		advice:   advice,
		order:    o,
	}
}
