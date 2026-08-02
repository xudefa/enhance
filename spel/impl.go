// Package spel 提供 Spring Expression Language (SpEL) 表达式支持，用于 enhance 框架。
package spel

import "sync"

// standardEvaluationContextImpl EvaluationContext 接口的默认实现。
type standardEvaluationContextImpl struct {
	rootObject       any
	mu               sync.RWMutex
	variables        map[string]any
	propertyAccessor PropertyAccessor
}

// NewStandardEvaluationContext 创建标准求值上下文。
func NewStandardEvaluationContext(root any) EvaluationContext {
	return &standardEvaluationContextImpl{
		rootObject:       root,
		variables:        make(map[string]any),
		propertyAccessor: NewReflectPropertyAccessor(),
	}
}

// reflectPropertyAccessorImpl PropertyAccessor 接口的默认实现。
type reflectPropertyAccessorImpl struct{}

// NewReflectPropertyAccessor 创建反射属性访问器。
func NewReflectPropertyAccessor() PropertyAccessor {
	return &reflectPropertyAccessorImpl{}
}

// spelParserImpl ExpressionParser 接口的默认实现。
type spelParserImpl struct{}

// NewSpelParser 创建 SpEL 表达式解析器。
func NewSpelParser() ExpressionParser {
	return &spelParserImpl{}
}

// propertyExpressionImpl Expression 接口的简单属性实现。
type propertyExpressionImpl struct {
	property string
}

// complexExpressionImpl Expression 接口的复杂表达式实现。
type complexExpressionImpl struct {
	raw string
}

// interceptorChainImpl MethodInterceptor 接口的拦截器链实现。
type interceptorChainImpl struct {
	mu           sync.Mutex
	interceptors []MethodInterceptor
	index        int
	invocation   MethodInvocation
}

// NewInterceptorChain 创建拦截器链。
func NewInterceptorChain(interceptors []MethodInterceptor) MethodInterceptor {
	return &interceptorChainImpl{
		interceptors: interceptors,
		index:        0,
	}
}

// simpleMethodInvocationImpl MethodInvocation 接口的简单实现。
type simpleMethodInvocationImpl struct {
	method  string
	args    []any
	target  any
	handler func() (any, error)
}

// NewSimpleMethodInvocation 创建简单方法调用。
func NewSimpleMethodInvocation(method string, arguments []any, target any, proceedFn func() (any, error)) MethodInvocation {
	return &simpleMethodInvocationImpl{
		method:  method,
		args:    arguments,
		target:  target,
		handler: proceedFn,
	}
}

// loggingInterceptorImpl MethodInterceptor 接口的日志实现。
type loggingInterceptorImpl struct{}

// NewLoggingInterceptor 创建日志拦截器。
func NewLoggingInterceptor() MethodInterceptor {
	return &loggingInterceptorImpl{}
}
