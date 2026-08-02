// Package aop 提供面向切面编程（AOP）支持，用于 enhance 框架。
//
// 该模块提供完整的 AOP 实现，包括通知、切点、顾问器、动态代理等功能。
// 支持静态代理代码生成，提升运行时性能。
//
// # 架构设计
//
//   - Advice: 通知接口（Before/After/AfterReturning/AfterThrowing/Around）
//   - PointCut: 切点定义，支持方法名、包名、注解匹配
//   - Advisor: 顾问器，组合切点和通知
//   - Proxy: 动态代理生成
//   - Aspect: 切面定义和管理
//   - 代码生成: 支持静态代理代码生成
//
// # 使用方式
//
// 直接使用：
//
//	// 创建切面
//	aspect := aop.NewAspect("loggingAspect").
//	    AddAdvice(aop.BeforeAdvice(func(ctx aop.JoinPoint) {
//	        fmt.Println("Before:", ctx.MethodName())
//	    }))
//
//	// 创建代理
//	proxy := aop.NewProxy(target, aspect)
//
// # 代码生成
//
// 使用 go generate 生成静态代理代码：
//
//	//go:generate go run github.com/xudefa/enhance/cmd/goaop
//
// # 支持的匹配模式
//
//   - 方法名匹配：支持通配符（*）
//   - 包名匹配：支持包路径前缀匹配
//   - 注解匹配：基于方法注解进行匹配
package aop

import (
	"context"
	"reflect"
)

// ==================== 核心接口 ====================

// Advice 通知接口。
//
// 定义 AOP 通知的核心行为。通知是对目标方法的增强逻辑，
// 在方法执行的生命周期中的特定时机介入。
//
// 推荐使用 Before/After/Around 等函数式 API 替代直接实现此接口，
// 以获得更简洁的 Go 惯用法体验。此接口主要用于框架内部实现。
type Advice interface {
	// Type 返回通知类型。
	Type() AdviceType

	// Order 返回执行顺序。
	Order() int

	// Execute 执行通知。
	//
	// 执行通知的增强逻辑。对于 Around 通知，需要通过 joinPoint.Proceed()
	// 调用目标方法或下一个通知。
	//
	// 参数:
	//   - ctx: 调用上下文
	//   - joinPoint: 连接点，包含方法调用的上下文信息
	//
	// 返回值:
	//   - any: 通知的返回值
	//   - error: 执行错误
	Execute(ctx context.Context, joinPoint JoinPoint) (any, error)
}

// AdviceType 通知类型枚举。
//
// 定义 AOP 框架中的五种标准通知类型，对应 Spring AOP 的通知模型。
// 每种通知类型决定了增强逻辑在目标方法执行生命周期中的介入时机。
type AdviceType int

const (
	// AdviceTypeBefore 前置通知。
	//
	// 在目标方法执行之前调用增强逻辑。
	// 适用于日志记录、参数校验、权限检查等场景。
	// 前置通知无法阻止目标方法的执行。
	AdviceTypeBefore AdviceType = iota

	// AdviceTypeAfter 后置通知。
	//
	// 在目标方法执行之后调用，无论方法是否抛出异常。
	// 适用于资源清理、状态重置等场景。
	// 注意：后置通知在异常通知之前执行。
	AdviceTypeAfter

	// AdviceTypeAround 环绕通知。
	//
	// 最强大的通知类型，完全控制目标方法的执行。
	// 可以决定是否执行目标方法、何时执行、执行几次，甚至可以替换返回值。
	// 必须调用 joinPoint.Proceed() 使调用链继续，否则目标方法不会执行。
	// 适用于事务管理、性能监控、重试逻辑等场景。
	AdviceTypeAround

	// AdviceTypeAfterReturning 返回后通知。
	//
	// 在目标方法正常返回后调用（未抛出异常）。
	// 可以访问方法的返回值，适用于结果缓存、响应增强等场景。
	// 如果方法抛出异常，此通知不会执行。
	AdviceTypeAfterReturning

	// AdviceTypeAfterThrowing 异常后通知。
	//
	// 在目标方法抛出异常后调用。
	// 可以访问错误对象，适用于错误日志、异常转换、告警通知等场景。
	// 如果方法正常返回，此通知不会执行。
	AdviceTypeAfterThrowing
)

// PointCut 切点接口。
//
// 定义 AOP 中用于匹配目标方法的规则。切点决定了哪些类或方法需要被拦截，
// 是 AOP 框架的核心组件之一。
type PointCut interface {
	// Matches 是否匹配。
	//
	// 检查给定目标对象和方法名是否匹配切点条件。
	//
	// 参数:
	//   - target: 目标对象
	//   - methodName: 方法名
	//
	// 返回值:
	//   - bool: 是否匹配
	Matches(target any, methodName string) bool

	// MatchClass 是否匹配类。
	//
	// 检查给定类型是否匹配切点的类级别条件，不考虑方法名。
	// 用于预筛选切面，减少不必要的方法级匹配开销。
	//
	// 参数:
	//   - t: 目标类型
	//
	// 返回值:
	//   - bool: 是否匹配
	MatchClass(t reflect.Type) bool

	// Expression 返回切点表达式。
	//
	// 用于调试和日志输出，返回切点的匹配规则。
	Expression() string
}

// Advisor 通知器接口。
//
// 通知器是 AOP 中的基本单元，包含一个切点和一个通知。
// 类似于 Spring 中的 Advisor 概念。
//
// 使用场景:
//   - 细粒度的切面控制，为每个通知指定独立的切点和执行顺序
//   - 精确控制通知执行顺序和匹配规则
type Advisor interface {
	// Advice 获取通知。
	Advice() Advice

	// PointCut 获取切点。
	PointCut() PointCut

	// Order 获取执行顺序。
	Order() int
}

// JoinPoint 连接点接口。
//
// AOP 核心概念，代表程序执行的某个位置。在 enhance AOP 框架中，
// 连接点通常指方法调用，通知（Advice）可以通过 JoinPoint 访问方法调用的上下文。
type JoinPoint interface {
	// Target 获取目标对象。
	Target() any

	// Method 获取方法名。
	Method() string

	// Args 获取参数。
	Args() []any

	// Proceed 执行原方法。
	Proceed() (any, error)

	// ProceedWithArgs 带参数执行原方法。
	ProceedWithArgs(args []any) (any, error)

	// Context 获取上下文。
	Context() context.Context

	// GetResult 获取已执行的结果。
	GetResult() any

	// GetError 获取已执行的错误。
	GetError() error

	// SetResult 设置执行结果。
	SetResult(v any)

	// SetError 设置执行错误。
	SetError(err error)
}

// Invocation 调用接口。
//
// 用于在通知链中控制方法的执行流程。
type Invocation interface {
	// JoinPoint 获取连接点。
	JoinPoint() JoinPoint

	// Arguments 获取参数。
	Arguments() []any

	// Proceed 执行。
	Proceed() (any, error)
}

// ChainExecutor 通知链执行器接口。
//
// 定义通知链的执行策略。默认实现支持 panic 恢复、自定义拦截器和 context 传播。
// 可通过实现此接口自定义执行策略（如异步执行、限流等）。
type ChainExecutor interface {
	// Execute 执行通知链。
	//
	// 参数:
	//   - inv: 调用信息
	//   - aspects: 切面元数据列表
	//   - targetFunc: 目标方法调用函数
	//
	// 返回值:
	//   - any: 方法执行结果
	Execute(inv Invocation, aspects []*AspectMeta, targetFunc func(...any) any) any
}

// Weaver 织入器接口。
//
// 负责将切面织入目标对象，生成代理对象。
// 类似于 Spring 中的 AopProxyFactory。
//
// 工作流程:
//  1. 创建织入器: NewWeaver()
//  2. 添加切面: AddAspects(aspect1, aspect2, ...)
//  3. 织入目标: Weave(target) -> 返回代理对象
//
// 织入规则:
//   - 如果目标对象没有任何匹配的切面，返回原对象（不进行代理）
//   - 如果目标对象有匹配的切面，创建代理对象
//   - 代理对象的方法调用会触发匹配的通知
type Weaver interface {
	// Weave 织入目标对象。
	//
	// 将已注册的切面织入目标对象，返回代理对象。
	//
	// 参数:
	//   - target: 目标对象，可以是结构体指针或接口实现
	//
	// 返回值:
	//   - any: 代理对象（如果匹配到切面）或原对象
	//
	// 注意:
	//   - 返回的对象类型可能和原对象不同（代理类型）
	//   - 使用代理对象调用方法时，匹配的通知会自动执行
	Weave(target any) any

	// AddAspects 添加切面。
	//
	// 添加一个或多个切面到织入器。
	// 添加后，这些切面会应用于后续 Weave 调用。
	//
	// 参数:
	//   - aspects: 一个或多个切面元数据
	AddAspects(aspects ...*AspectMeta)
}

// ==================== 函数类型 ====================

// ProceedFunc 继续执行函数类型。
//
// 在 Around 通知中，调用此函数可以继续执行目标方法或下一个通知。
// 参数为传递给目标方法的参数，返回值为目标方法的返回值。
type ProceedFunc func(args ...any) any

// Interceptor 拦截器函数类型。
//
// 在通知链执行前后提供额外的处理逻辑，采用中间件模式。
// inv 为当前调用信息，next 为下一个处理函数。
//
// 示例:
//
//	aop.WithInterceptor(func(inv aop.Invocation, next func(aop.Invocation) any) any {
//	    start := time.Now()
//	    result := next(inv)
//	    slog.Info("method called", "name", inv.Signature().Name(), "duration", time.Since(start))
//	    return result
//	})
type Interceptor func(inv Invocation, next func(Invocation) any) any

// 以下类型定义在其他文件中，此处仅作文档说明：
// - ClassMatcher: point_cut.go
// - MethodMatcher: point_cut.go

// ==================== 枚举类型 ====================

// AopMode AOP 模式枚举。
type AopMode string

const (
	// AopModeRuntime 运行时模式，使用反射和动态代理。
	AopModeRuntime AopMode = "runtime"

	// AopModeGenerated 代码生成模式，使用编译时代码生成。
	AopModeGenerated AopMode = "generated"

	// AopModeMixed 混合模式，自动选择最优方案。
	AopModeMixed AopMode = "mixed"
)

// ==================== 通用结构体 ====================

// AspectMeta 切面元数据。
//
// 存储切面的实例、切点、通知和执行顺序。
// 切面是 AOP 的核心数据结构，定义了"在什么地方、做什么增强"。
type AspectMeta struct {
	Instance any      // 切面实例对象
	PointCut PointCut // 切点定义，匹配目标方法
	Advice   Advice   // 通知，包含增强逻辑
	Order    int      // 执行顺序，数字越小越先执行
}

// MethodInvocation 方法调用信息。
//
// 用于代码生成的代理类，包含方法调用所需的所有信息。
// 新增字段 Proxy、Ctx 可按需设置，未设置时保持零值兼容。
type MethodInvocation struct {
	MethodName string          // 方法名称
	Func       any             // 目标方法（函数值）
	Params     []any           // 方法调用参数列表
	Object     any             // 目标对象（被代理的原始对象）
	Proxy      any             // 代理对象本身，未设置时 This() 返回 nil
	Ctx        context.Context // 上下文信息，未设置时 Context() 返回 context.Background()
	proceed    ProceedFunc     // 继续执行函数，用于 Around 通知中控制执行流程
	result     any             // 方法执行结果
	lastErr    error           // 方法执行错误
}

// AopConfig AOP 配置。
type AopConfig struct {
	Mode        AopMode // AOP 模式
	Weaver      Weaver  // 织入器
	EnableCache bool    // 是否启用代理缓存
}

// 以下类型定义在其他文件中，此处仅作文档说明：
// - AopContainer: container.go（AOP 容器结构体）
// - AopManager: config.go（AOP 管理器结构体）
// - ClassMatcher: point_cut.go
// - MethodMatcher: point_cut.go
