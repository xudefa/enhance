// Package aop 提供面向切面编程（AOP）支持。
package aop

// adviceFunc 函数式通知适配器。
//
// 将简单的函数适配为 Advice 接口，是 Go 惯用法的体现。
// 类似于 http.HandlerFunc 的设计模式，避免用户必须实现完整接口。
type adviceFunc struct {
	adviceType AdviceType
	fn         func(JoinPoint, ProceedFunc) any
}

// Type 返回通知类型
func (a *adviceFunc) Type() AdviceType {
	return a.adviceType
}

// Apply 应用通知
func (a *adviceFunc) Apply(jp JoinPoint, proceed ProceedFunc) any {
	if a.fn != nil {
		return a.fn(jp, proceed)
	}
	return nil
}

// Before 创建前置通知
//
// 在目标方法执行之前执行增强逻辑。
// 前置通知无法修改方法参数或阻止方法执行。
//
// 参数:
//   - fn: 前置通知函数，接收 JoinPoint 参数，可访问方法信息
//
// 返回值:
//   - Advice: 前置通知实例
//
// 示例:
//
//	aop.Before(func(jp aop.JoinPoint) {
//	    fmt.Println("方法执行前:", jp.Method().Name)
//	})
func Before(fn func(JoinPoint)) Advice {
	return &adviceFunc{
		adviceType: AdviceBefore,
		fn: func(jp JoinPoint, _ ProceedFunc) any {
			fn(jp)
			return nil
		},
	}
}

// After 创建后置通知
//
// 在目标方法执行之后执行增强逻辑，无论方法是否抛出异常。
// 后置通知在异常通知之前执行。
//
// 参数:
//   - fn: 后置通知函数，接收 JoinPoint 参数
//
// 返回值:
//   - Advice: 后置通知实例
//
// 示例:
//
//	aop.After(func(jp aop.JoinPoint) {
//	    fmt.Println("方法执行后:", jp.Method().Name)
//	})
func After(fn func(JoinPoint)) Advice {
	return &adviceFunc{
		adviceType: AdviceAfter,
		fn: func(jp JoinPoint, _ ProceedFunc) any {
			fn(jp)
			return nil
		},
	}
}

// AfterReturning 创建返回通知
//
// 在目标方法正常返回后执行增强逻辑。
// 可以访问方法的返回值，适用于结果缓存、响应增强等场景。
// 如果方法抛出异常，此通知不会执行。
//
// 参数:
//   - fn: 返回通知函数，接收 JoinPoint 和方法返回值
//
// 返回值:
//   - Advice: 返回通知实例
//
// 示例:
//
//	aop.AfterReturning(func(jp aop.JoinPoint, result any) {
//	    fmt.Println("方法返回:", result)
//	})
func AfterReturning(fn func(JoinPoint, any)) Advice {
	return &adviceFunc{
		adviceType: AdviceAfterReturning,
		fn: func(jp JoinPoint, proceed ProceedFunc) any {
			var result any
			if proceed != nil {
				result = proceed()
			}
			fn(jp, result)
			return result
		},
	}
}

// AfterThrowing 创建异常通知
//
// 在目标方法抛出异常后执行增强逻辑。
// 可以访问错误对象，适用于错误日志、异常转换、告警通知等场景。
// 如果方法正常返回，此通知不会执行。
//
// 参数:
//   - fn: 异常通知函数，接收 JoinPoint 和错误对象
//
// 返回值:
//   - Advice: 异常通知实例
//
// 示例:
//
//	aop.AfterThrowing(func(jp aop.JoinPoint, err error) {
//	    fmt.Println("方法异常:", err)
//	})
func AfterThrowing(fn func(JoinPoint, error)) Advice {
	return &adviceFunc{
		adviceType: AdviceAfterThrowing,
		fn: func(jp JoinPoint, proceed ProceedFunc) any {
			var err error
			if proceed != nil {
				result := proceed()
				if result != nil {
					// 从返回值中提取 error（通常 error 是最后一个返回值）
					if multiResult, ok := result.([]any); ok && len(multiResult) > 0 {
						// 从后往前查找，优先取最后一个 error
						for i := len(multiResult) - 1; i >= 0; i-- {
							if e, ok := multiResult[i].(error); ok {
								err = e
								break
							}
						}
					} else if e, ok := result.(error); ok {
						err = e
					}
				}
			}
			fn(jp, err)
			return nil
		},
	}
}

// Around 创建环绕通知
//
// 最强大的通知类型，完全控制目标方法的执行。
// 可以决定是否执行目标方法、何时执行、执行几次，甚至可以替换返回值。
//
// 重要：Around 通知必须调用 proceed 函数使调用链继续，否则目标方法不会执行。
//
// 参数:
//   - fn: 环绕通知函数，接收 JoinPoint 和 ProceedFunc
//
// 返回值:
//   - Advice: 环绕通知实例
//
// 示例:
//
//	aop.Around(func(jp aop.JoinPoint, proceed aop.ProceedFunc) any {
//	    fmt.Println("方法执行前:", jp.Method().Name)
//	    result := proceed()
//	    fmt.Println("方法执行后:", result)
//	    return result
//	})
func Around(fn func(JoinPoint, ProceedFunc) any) Advice {
	return &adviceFunc{
		adviceType: AdviceAround,
		fn:         fn,
	}
}
