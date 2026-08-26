package aop

import (
	"context"
	"fmt"
)

// typedTarget 将 JoinPoint 目标断言为类型 T。
//
// 支持结构体指针（*T）与接口类型（T）两种目标形式。
// 当类型不匹配时返回明确的错误信息。
func typedTarget[T any](jp JoinPoint) (T, error) {
	var zero T
	t, ok := jp.Target().(T)
	if !ok {
		return zero, fmt.Errorf("aop: target %T does not match expected type %T", jp.Target(), zero)
	}
	return t, nil
}

// NewBefore 创建绑定目标类型的 Before 通知。
//
// fn 的 target 参数已被类型断言为 T，编译期类型安全。
// 当目标类型不匹配时，通知会返回错误并阻止方法执行。
//
// 示例:
//
//	type UserService struct { ... }
//
//	advice := aop.NewBefore[*UserService](
//	    func(ctx context.Context, target *UserService, jp aop.JoinPoint) error {
//	        log.Printf("calling method: %s", jp.Method())
//	        return nil
//	    },
//	    0,
//	)
func NewBefore[T any](fn func(ctx context.Context, target T, jp JoinPoint) error, order int) Advice {
	return NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
		t, err := typedTarget[T](jp)
		if err != nil {
			return nil, err
		}
		if err := fn(ctx, t, jp); err != nil {
			return nil, err
		}
		return nil, nil
	}, order)
}

// NewAfter 创建绑定目标类型的 After 通知。
//
// fn 的 target 参数已被类型断言为 T，编译期类型安全。
// After 通知在目标方法执行后运行，无论方法是否返回错误。
//
// 示例:
//
//	advice := aop.NewAfter[*UserService](
//	    func(ctx context.Context, target *UserService, jp aop.JoinPoint) error {
//	        log.Printf("method %s completed", jp.Method())
//	        return nil
//	    },
//	    0,
//	)
func NewAfter[T any](fn func(ctx context.Context, target T, jp JoinPoint) error, order int) Advice {
	return NewAfterAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
		t, err := typedTarget[T](jp)
		if err != nil {
			return nil, err
		}
		if err := fn(ctx, t, jp); err != nil {
			return nil, err
		}
		return nil, nil
	}, order)
}

// NewAround 创建绑定目标类型的 Around 通知。
//
// fn 的 target 参数已被类型断言为 T，编译期类型安全。
// proceed 函数必须被调用以继续执行链，否则目标方法不会执行。
//
// 示例:
//
//	advice := aop.NewAround[*UserService](
//	    func(ctx context.Context, target *UserService, jp aop.JoinPoint, proceed func() (any, error)) (any, error) {
//	        log.Printf("before %s", jp.Method())
//	        result, err := proceed()
//	        log.Printf("after %s", jp.Method())
//	        return result, err
//	    },
//	    0,
//	)
func NewAround[T any](fn func(ctx context.Context, target T, jp JoinPoint, proceed func() (any, error)) (any, error), order int) Advice {
	return NewAroundAdvice(func(ctx context.Context, jp JoinPoint, proceed func() (any, error)) (any, error) {
		t, err := typedTarget[T](jp)
		if err != nil {
			return nil, err
		}
		return fn(ctx, t, jp, proceed)
	}, order)
}

// NewAfterReturning 创建绑定目标类型的 AfterReturning 通知。
//
// fn 的 target 参数已被类型断言为 T，编译期类型安全。
// 仅在目标方法正常返回（无错误）时执行。
//
// 示例:
//
//	advice := aop.NewAfterReturning[*UserService](
//	    func(ctx context.Context, target *UserService, jp aop.JoinPoint, result any) error {
//	        log.Printf("method %s returned: %v", jp.Method(), result)
//	        return nil
//	    },
//	    0,
//	)
func NewAfterReturning[T any](fn func(ctx context.Context, target T, jp JoinPoint, result any) error, order int) Advice {
	return NewAfterReturningAdvice(func(ctx context.Context, jp JoinPoint, result any) (any, error) {
		t, err := typedTarget[T](jp)
		if err != nil {
			return nil, err
		}
		if err := fn(ctx, t, jp, result); err != nil {
			return nil, err
		}
		return result, nil
	}, order)
}

// NewAfterThrowing 创建绑定目标类型的 AfterThrowing 通知。
//
// fn 的 target 参数已被类型断言为 T，编译期类型安全。
// 仅在目标方法返回错误时执行。
//
// 示例:
//
//	advice := aop.NewAfterThrowing[*UserService](
//	    func(ctx context.Context, target *UserService, jp aop.JoinPoint, err error) error {
//	        log.Printf("method %s failed: %v", jp.Method(), err)
//	        return nil
//	    },
//	    0,
//	)
func NewAfterThrowing[T any](fn func(ctx context.Context, target T, jp JoinPoint, err error) error, order int) Advice {
	return NewAfterThrowingAdvice(func(ctx context.Context, jp JoinPoint, err error) (any, error) {
		t, terr := typedTarget[T](jp)
		if terr != nil {
			return nil, terr
		}
		if err := fn(ctx, t, jp, err); err != nil {
			return nil, err
		}
		return nil, err
	}, order)
}
