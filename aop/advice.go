// Package aop 提供面向切面编程（AOP）支持。
package aop

import "context"

// AdviceFunc 通知函数类型。
type AdviceFunc func(ctx context.Context, joinPoint JoinPoint) (any, error)

// beforeAdvice 前置通知实现。
type beforeAdvice struct {
	fn    AdviceFunc
	order int
}

// NewBeforeAdvice 创建前置通知。
//
// 在目标方法执行之前执行增强逻辑。
// 前置通知无法修改方法参数或阻止方法执行。
//
// 参数:
//   - fn: 通知函数
//   - order: 执行顺序，值越小优先级越高
//
// 返回值:
//   - Advice: 前置通知实例
func NewBeforeAdvice(fn AdviceFunc, order int) Advice {
	return &beforeAdvice{
		fn:    fn,
		order: order,
	}
}

func (a *beforeAdvice) Type() AdviceType {
	return AdviceTypeBefore
}

func (a *beforeAdvice) Order() int {
	return a.order
}

func (a *beforeAdvice) Execute(ctx context.Context, joinPoint JoinPoint) (any, error) {
	if a.fn != nil {
		return a.fn(ctx, joinPoint)
	}
	return nil, nil
}

// afterAdvice 后置通知实现。
type afterAdvice struct {
	fn    AdviceFunc
	order int
}

// NewAfterAdvice 创建后置通知。
//
// 在目标方法执行之后执行增强逻辑，无论方法是否抛出异常。
// 后置通知在异常通知之前执行。
//
// 参数:
//   - fn: 通知函数
//   - order: 执行顺序
//
// 返回值:
//   - Advice: 后置通知实例
func NewAfterAdvice(fn AdviceFunc, order int) Advice {
	return &afterAdvice{
		fn:    fn,
		order: order,
	}
}

func (a *afterAdvice) Type() AdviceType {
	return AdviceTypeAfter
}

func (a *afterAdvice) Order() int {
	return a.order
}

func (a *afterAdvice) Execute(ctx context.Context, joinPoint JoinPoint) (any, error) {
	if a.fn != nil {
		return a.fn(ctx, joinPoint)
	}
	return nil, nil
}

// aroundAdvice 环绕通知实现。
type aroundAdvice struct {
	fn    func(ctx context.Context, joinPoint JoinPoint, proceed func() (any, error)) (any, error)
	order int
}

// NewAroundAdvice 创建环绕通知。
//
// 最强大的通知类型，完全控制目标方法的执行。
// 可以决定是否执行目标方法、何时执行、执行几次，甚至可以替换返回值。
//
// 重要：Around 通知必须调用 proceed 函数使调用链继续，否则目标方法不会执行。
//
// 参数:
//   - fn: 通知函数，接收 proceed 参数
//   - order: 执行顺序
//
// 返回值:
//   - Advice: 环绕通知实例
func NewAroundAdvice(fn func(ctx context.Context, joinPoint JoinPoint, proceed func() (any, error)) (any, error), order int) Advice {
	return &aroundAdvice{
		fn:    fn,
		order: order,
	}
}

func (a *aroundAdvice) Type() AdviceType {
	return AdviceTypeAround
}

func (a *aroundAdvice) Order() int {
	return a.order
}

func (a *aroundAdvice) Execute(ctx context.Context, joinPoint JoinPoint) (any, error) {
	if a.fn != nil {
		proceed := func() (any, error) {
			return joinPoint.Proceed()
		}
		return a.fn(ctx, joinPoint, proceed)
	}
	return joinPoint.Proceed()
}

// AfterReturningAdvice 返回后通知实现。
type afterReturningAdvice struct {
	fn    func(ctx context.Context, joinPoint JoinPoint, result any) (any, error)
	order int
}

// NewAfterReturningAdvice 创建返回后通知。
//
// 在目标方法正常返回后执行增强逻辑。
// 可以访问方法的返回值，适用于结果缓存、响应增强等场景。
// 如果方法抛出异常，此通知不会执行。
//
// 参数:
//   - fn: 通知函数，接收 result 参数
//   - order: 执行顺序
//
// 返回值:
//   - Advice: 返回后通知实例
func NewAfterReturningAdvice(fn func(ctx context.Context, joinPoint JoinPoint, result any) (any, error), order int) Advice {
	return &afterReturningAdvice{
		fn:    fn,
		order: order,
	}
}

func (a *afterReturningAdvice) Type() AdviceType {
	return AdviceTypeAfterReturning
}

func (a *afterReturningAdvice) Order() int {
	return a.order
}

func (a *afterReturningAdvice) Execute(ctx context.Context, joinPoint JoinPoint) (any, error) {
	result := joinPoint.GetResult()
	if a.fn != nil {
		return a.fn(ctx, joinPoint, result)
	}
	return result, nil
}

// afterThrowingAdvice 异常后通知实现。
type afterThrowingAdvice struct {
	fn    func(ctx context.Context, joinPoint JoinPoint, err error) (any, error)
	order int
}

// NewAfterThrowingAdvice 创建异常后通知。
//
// 在目标方法抛出异常后执行增强逻辑。
// 可以访问错误对象，适用于错误日志、异常转换、告警通知等场景。
// 如果方法正常返回，此通知不会执行。
//
// 参数:
//   - fn: 通知函数，接收 err 参数
//   - order: 执行顺序
//
// 返回值:
//   - Advice: 异常后通知实例
func NewAfterThrowingAdvice(fn func(ctx context.Context, joinPoint JoinPoint, err error) (any, error), order int) Advice {
	return &afterThrowingAdvice{
		fn:    fn,
		order: order,
	}
}

func (a *afterThrowingAdvice) Type() AdviceType {
	return AdviceTypeAfterThrowing
}

func (a *afterThrowingAdvice) Order() int {
	return a.order
}

func (a *afterThrowingAdvice) Execute(ctx context.Context, joinPoint JoinPoint) (any, error) {
	err := joinPoint.GetError()
	if err != nil && a.fn != nil {
		return a.fn(ctx, joinPoint, err)
	}
	return nil, err
}

// 向后兼容的函数式 API

// Before 创建前置通知（便捷函数）。
//
// 参数:
//   - fn: 通知函数（无 context 和返回值）
//
// 返回值:
//   - Advice: 前置通知实例
func Before(fn func(JoinPoint)) Advice {
	return NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
		fn(jp)
		return nil, nil
	}, 0)
}

// After 创建后置通知（便捷函数）。
func After(fn func(JoinPoint)) Advice {
	return NewAfterAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
		fn(jp)
		return nil, nil
	}, 0)
}

// Around 创建环绕通知（便捷函数）。
//
// fn 接收 JoinPoint 与 ProceedFunc。ProceedFunc 与 doc.go 中定义一致：
//   - 无参数调用：继续原方法（无参透传）
//   - 带参数调用：替换参数执行（proceed(args...)，调用 ProceedWithArgs）
func Around(fn func(JoinPoint, ProceedFunc) any) Advice {
	return NewAroundAdvice(func(ctx context.Context, jp JoinPoint, proceed func() (any, error)) (any, error) {
		var proceedErr error
		result := fn(jp, func(args ...any) any {
			var r any
			var err error
			if len(args) > 0 {
				r, err = jp.ProceedWithArgs(args)
			} else {
				r, err = proceed()
			}
			if err != nil {
				proceedErr = err
			}
			return r
		})
		if proceedErr != nil {
			return result, proceedErr
		}
		return result, nil
	}, 0)
}

// AfterReturning 创建返回后通知（便捷函数）。
func AfterReturning(fn func(JoinPoint, any)) Advice {
	return NewAfterReturningAdvice(func(ctx context.Context, jp JoinPoint, result any) (any, error) {
		fn(jp, result)
		return result, nil
	}, 0)
}

// AfterThrowing 创建异常后通知（便捷函数）。
func AfterThrowing(fn func(JoinPoint, error)) Advice {
	return NewAfterThrowingAdvice(func(ctx context.Context, jp JoinPoint, err error) (any, error) {
		fn(jp, err)
		return nil, err
	}, 0)
}
