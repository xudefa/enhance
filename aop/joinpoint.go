// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"context"
	"reflect"
)

// joinPointImpl 连接点实现。
type joinPointImpl struct {
	target          any
	method          string
	args            []any
	proceed         func() (any, error)
	proceedWithArgs func([]any) (any, error)
	ctx             context.Context
}

// NewJoinPoint 创建连接点。
func NewJoinPoint(target any, method string, args []any, proceed func() (any, error), proceedWithArgs func([]any) (any, error)) JoinPoint {
	return &joinPointImpl{
		target:          target,
		method:          method,
		args:            args,
		proceed:         proceed,
		proceedWithArgs: proceedWithArgs,
		ctx:             context.Background(),
	}
}

func (j *joinPointImpl) Target() any {
	return j.target
}

func (j *joinPointImpl) Method() string {
	return j.method
}

func (j *joinPointImpl) Args() []any {
	return j.args
}

func (j *joinPointImpl) Proceed() (any, error) {
	if j.proceed != nil {
		return j.proceed()
	}
	return nil, nil
}

func (j *joinPointImpl) ProceedWithArgs(args []any) (any, error) {
	if j.proceedWithArgs != nil {
		return j.proceedWithArgs(args)
	}
	return nil, nil
}

// invocationImpl 调用实现。
type invocationImpl struct {
	joinPoint JoinPoint
	args      []any
	proceed   func() (any, error)
	lastErr   error
}

// NewInvocation 创建调用。
func NewInvocation(joinPoint JoinPoint, proceed func() (any, error)) Invocation {
	return &invocationImpl{
		joinPoint: joinPoint,
		args:      joinPoint.Args(),
		proceed:   proceed,
	}
}

func (i *invocationImpl) JoinPoint() JoinPoint {
	return i.joinPoint
}

func (i *invocationImpl) Arguments() []any {
	return i.args
}

func (i *invocationImpl) Proceed() (any, error) {
	if i.proceed != nil {
		return i.proceed()
	}
	return nil, nil
}

// SetError 在调用上存储错误信息。
func (i *invocationImpl) SetError(err error) {
	i.lastErr = err
}

// Error 返回已存储的错误信息。
func (i *invocationImpl) Error() error {
	return i.lastErr
}

// methodSignature 方法签名实现。
type methodSignature struct {
	name          string
	declaringType reflect.Type
}

func (m *methodSignature) Name() string {
	return m.name
}

func (m *methodSignature) DeclaringType() reflect.Type {
	return m.declaringType
}

// NewMethodSignature 创建方法签名。
func NewMethodSignature(name string, t reflect.Type) *methodSignature {
	return &methodSignature{
		name:          name,
		declaringType: t,
	}
}
