// Package spel 提供 Spring Expression Language (SpEL) 表达式支持，用于 enhance 框架。
//
// 该模块提供表达式解析、求值上下文、属性访问器和方法拦截器等功能。
// 参考 Spring Framework 的 SpEL 设计。
//
// # 架构设计
//
//   - Expression: 表达式接口，定义表达式求值和设置操作
//   - ExpressionParser: 表达式解析器接口，解析表达式字符串
//   - EvaluationContext: 表达式求值上下文接口，管理根对象和变量
//   - PropertyAccessor: 属性访问器接口，提供属性读写操作
//   - MethodInterceptor: 方法拦截器接口，用于方法调用拦截
//   - MethodInvocation: 方法调用上下文接口，提供方法调用信息
//
// # 核心功能
//
//   - 表达式解析: 支持属性访问、方法调用、运算符等表达式语法
//   - 动态求值: 在运行时根据上下文计算表达式值
//   - 属性访问: 基于反射的属性读写支持
//   - 变量管理: 支持在上下文中设置和获取命名变量
//   - 方法拦截: 支持方法调用前后的拦截逻辑
//
// # 使用方式
//
// 解析和求值表达式：
//
//	parser := spel.NewSpelParser()
//	expr, err := parser.ParseExpression("user.name")
//	if err != nil {
//	    // 处理解析错误
//	}
//
//	context := spel.NewStandardEvaluationContext(user)
//	value, err := expr.GetValue(context)
//
// 设置变量：
//
//	context.SetVariable("role", "admin")
//	expr, _ := parser.ParseExpression("#role")
//	value, _ := expr.GetValue(context)
//
// 方法拦截：
//
//	interceptor := spel.NewLoggingInterceptor()
//	chain := spel.NewInterceptorChain([]spel.MethodInterceptor{interceptor})
//	result, err := chain.Proceed()
package spel

// Expression 表达式接口。
//
// 定义表达式求值和设置操作的标准接口。
type Expression interface {
	// GetValue 在给定上下文中计算表达式值。
	GetValue(context EvaluationContext) (any, error)

	// SetValue 在给定上下文中设置表达式值。
	SetValue(context EvaluationContext, value any) error

	// String 返回表达式字符串。
	String() string
}

// ExpressionParser 表达式解析器接口。
//
// 解析表达式字符串为可执行的 Expression 对象。
type ExpressionParser interface {
	// ParseExpression 解析表达式字符串为 Expression。
	ParseExpression(expression string) (Expression, error)
}

// EvaluationContext 表达式求值上下文接口。
//
// 管理表达式的根对象、命名变量和属性访问器。
type EvaluationContext interface {
	// GetRootObject 获取根对象。
	GetRootObject() any

	// SetRootObject 设置根对象。
	SetRootObject(root any)

	// GetVariable 获取命名变量。
	GetVariable(name string) (any, bool)

	// SetVariable 设置命名变量。
	SetVariable(name string, value any)

	// GetPropertyAccessor 获取属性访问器。
	GetPropertyAccessor() PropertyAccessor
}

// PropertyAccessor 属性访问器接口。
//
// 提供从目标对象读写属性的标准方法。
type PropertyAccessor interface {
	// GetProperty 从目标对象获取指定属性。
	GetProperty(target any, name string) (any, error)

	// SetProperty 设置目标对象的指定属性。
	SetProperty(target any, name string, value any) error
}

// MethodInterceptor 方法拦截器接口。
//
// 用于在方法调用前后执行额外逻辑，如日志、权限检查、缓存等。
type MethodInterceptor interface {
	// Invoke 执行拦截逻辑，可以选择调用原方法或返回自定义结果。
	Invoke(invocation MethodInvocation) (any, error)
}

// MethodInvocation 方法调用上下文接口。
//
// 提供对被调用方法、目标对象和参数的访问，
// 拦截器可以通过 Proceed() 继续执行原方法。
type MethodInvocation interface {
	// GetMethod 获取方法名。
	GetMethod() string

	// GetArguments 获取方法参数。
	GetArguments() []any

	// GetTarget 获取目标对象。
	GetTarget() any

	// Proceed 继续执行原方法。
	Proceed() (any, error)
}
