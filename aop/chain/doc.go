// Package chain 提供 AOP 拦截器链实现，用于 enhance 框架。
//
// 该模块用于管理多个 Advice 的执行链，支持拦截器链的构建和执行。
// 参考 Spring AOP 的拦截器链设计。
//
// # 架构设计
//
//   - InterceptorChain: 拦截器链接口，定义链式执行逻辑
//   - DefaultInterceptorChain: 默认拦截器链实现
//   - MethodInvocation: 方法调用接口，表示当前方法调用上下文
//
// # 核心功能
//
//   - 链式执行: 按顺序执行多个拦截器
//   - 拦截器管理: 支持添加、删除、插入拦截器
//   - 方法调用: 支持方法调用的拦截和增强
//   - 异常处理: 支持拦截器链中的异常传播
//
// # 使用方式
//
// 创建拦截器链：
//
//	chain := chain.NewInterceptorChain()
//	chain.AddInterceptor(loggingInterceptor)
//	chain.AddInterceptor(authInterceptor)
//	chain.AddInterceptor(transactionInterceptor)
//
// 执行方法调用：
//
//	result, err := chain.Proceed(methodInvocation)
//
// # 执行顺序
//
// 拦截器按添加顺序依次执行，支持 Around 类型的拦截器：
//
//   - Before: 在目标方法执行前调用
//   - After: 在目标方法执行后调用
//   - Around: 包裹目标方法，可在前后执行逻辑
package chain
