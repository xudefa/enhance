// Package spel 提供 Spring Expression Language (SpEL) 表达式支持，用于 enhance 框架。
package spel

import (
	"fmt"
)

// AddInterceptor 添加拦截器。
func (c *interceptorChainImpl) AddInterceptor(interceptor MethodInterceptor) {
	c.interceptors = append(c.interceptors, interceptor)
}

// Proceed 执行下一个拦截器或目标方法。
func (c *interceptorChainImpl) Proceed() (any, error) {
	if c.index >= len(c.interceptors) {
		return c.invocation.Proceed()
	}

	interceptor := c.interceptors[c.index]
	c.index++
	return interceptor.Invoke(c)
}

// SetInvocation 设置方法调用上下文。
func (c *interceptorChainImpl) SetInvocation(invocation MethodInvocation) {
	c.invocation = invocation
}

// GetMethod 获取方法名（委托给 invocation）。
func (c *interceptorChainImpl) GetMethod() string {
	return c.invocation.GetMethod()
}

// GetArguments 获取方法参数（委托给 invocation）。
func (c *interceptorChainImpl) GetArguments() []any {
	return c.invocation.GetArguments()
}

// GetTarget 获取目标对象（委托给 invocation）。
func (c *interceptorChainImpl) GetTarget() any {
	return c.invocation.GetTarget()
}

// Invoke 执行拦截逻辑。
func (c *interceptorChainImpl) Invoke(invocation MethodInvocation) (any, error) {
	c.index = 0
	c.invocation = invocation
	return c.Proceed()
}

func (m *simpleMethodInvocationImpl) GetMethod() string {
	return m.method
}

func (m *simpleMethodInvocationImpl) GetArguments() []any {
	return m.args
}

func (m *simpleMethodInvocationImpl) GetTarget() any {
	return m.target
}

func (m *simpleMethodInvocationImpl) Proceed() (any, error) {
	if m.handler == nil {
		return nil, fmt.Errorf("no proceed function defined")
	}
	return m.handler()
}

func (l *loggingInterceptorImpl) Invoke(invocation MethodInvocation) (any, error) {
	method := invocation.GetMethod()
	args := invocation.GetArguments()

	fmt.Printf("[LOG] Calling %s with args: %v\n", method, args)

	result, err := invocation.Proceed()

	if err != nil {
		fmt.Printf("[LOG] %s returned error: %v\n", method, err)
		return result, err
	}
	fmt.Printf("[LOG] %s returned: %v\n", method, result)

	return result, err
}
