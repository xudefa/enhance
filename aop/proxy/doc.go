// Package proxy 提供 AOP 动态代理实现，用于 enhance 框架。
//
// 该模块提供两种代理实现和代理工厂：
//   - JdkDynamicProxy: 基于反射的 JDK 动态代理，支持接口代理
//   - CglibProxy: 基于结构体嵌入的 CGLIB 代理，支持非接口方法
//   - ProxyFactory: 代理工厂，统一创建代理对象
//
// # 架构设计
//
//   - Proxy: 代理接口，定义统一的调用方法
//   - ProxyFactory: 代理工厂接口，负责创建代理对象
//   - InvocationHandler: 调用处理器接口，定义方法拦截逻辑
//   - JdkDynamicProxy: JDK 动态代理实现
//   - CglibProxy: CGLIB 代理实现
//
// # 使用方式
//
// 创建 JDK 动态代理：
//
//	handler := &MyInvocationHandler{}
//	proxy := proxy.NewJdkDynamicProxy(target, iface, handler)
//	result, err := proxy.Invoke(target, "MethodName", []any{arg1, arg2})
//
// 创建 CGLIB 代理：
//
//	handler := &MyInvocationHandler{}
//	proxy := proxy.NewCglibProxy(target, handler)
//	result, err := proxy.Invoke(target, "MethodName", []any{arg1, arg2})
//
// 使用代理工厂：
//
//	factory := proxy.NewProxyFactory(target)
//	factory.SetHandler(handler)
//	proxy := factory.CreateProxy()
//
// # 设计模式
//
//   - Proxy 模式：封装目标对象，控制方法访问
//   - Factory 模式：统一代理对象创建
//   - Chain of Responsibility：通过 InvocationHandler 实现调用链
package proxy

import (
	"github.com/xudefa/enhance/aop"
)

// ==================== 核心接口 ====================

// Proxy 代理接口。
//
// 定义代理对象的核心行为。代理对象封装目标对象，控制方法访问。
// 所有代理实现都必须实现此接口，提供统一的调用方式。
type Proxy interface {
	// Invoke 调用代理方法。
	//
	// 通过代理对象调用目标方法，拦截器会自动执行匹配的通知。
	//
	// 参数:
	//   - target: 目标对象
	//   - method: 方法名
	//   - args: 方法参数列表
	//
	// 返回值:
	//   - any: 方法返回值
	//   - error: 调用错误
	Invoke(target any, method string, args []any) (any, error)

	// Target 获取原始目标对象。
	Target() any

	// AdvisedAdvisors 获取已应用的通知器列表。
	AdvisedAdvisors() []aop.Advisor
}

// ProxyFactory 代理工厂接口。
//
// 负责创建代理对象。根据目标对象的类型（接口或结构体），
// 自动选择最优的代理实现（JDK 动态代理或 CGLIB 代理）。
type ProxyFactory interface {
	// CreateProxy 创建代理对象。
	//
	// 根据目标对象类型自动选择代理实现。
	// 如果目标实现了接口，使用 JDK 动态代理；
	// 否则使用 CGLIB 代理。
	//
	// 参数:
	//   - target: 目标对象
	//   - handler: 调用处理器
	//
	// 返回值:
	//   - Proxy: 代理对象
	//   - error: 创建错误
	CreateProxy(target any, handler InvocationHandler) (Proxy, error)

	// CreateProxyWithAdvisors 创建带通知器的代理对象。
	//
	// 创建代理对象并应用指定的通知器列表。
	//
	// 参数:
	//   - target: 目标对象
	//   - handler: 调用处理器
	//   - advisors: 通知器列表
	//
	// 返回值:
	//   - Proxy: 代理对象
	//   - error: 创建错误
	CreateProxyWithAdvisors(target any, handler InvocationHandler, advisors []aop.Advisor) (Proxy, error)
}

// ProxyType 代理类型枚举。
type ProxyType int

const (
	// ProxyTypeJDK JDK 动态代理（基于接口）。
	ProxyTypeJDK ProxyType = iota

	// ProxyTypeCGLIB CGLIB 代理（基于类）。
	ProxyTypeCGLIB
)

// InvocationHandler 调用处理器接口。
//
// 定义方法拦截逻辑。在代理调用目标方法时，调用处理器可以：
//   - 在方法执行前执行前置逻辑
//   - 在方法执行后执行后置逻辑
//   - 替换方法执行结果
//   - 抛出异常阻止方法执行
//
// 类似于 java.lang.reflect.InvocationHandler。
type InvocationHandler interface {
	// Invoke 处理方法调用。
	//
	// 在目标方法调用前后执行自定义逻辑。
	// 必须调用 target 的实际方法（如果需要）。
	//
	// 参数:
	//   - target: 目标对象
	//   - method: 方法名
	//   - args: 方法参数列表
	//
	// 返回值:
	//   - any: 方法返回值
	//   - error: 调用错误
	Invoke(target any, method string, args []any) (any, error)
}

// ==================== 工厂函数 ====================
