// Package generator 提供 AOP 代码生成功能，用于 enhance 框架。
//
// 该模块用于生成静态代理代码，提高运行时性能。
// 通过代码生成避免运行时动态代理的性能开销。
//
// # 架构设计
//
//   - CodeGenerator: 代码生成器接口，定义代码生成逻辑
//   - ProxyGenerator: 代理代码生成器，生成静态代理类
//   - TemplateManager: 模板管理器，管理代码生成模板
//   - MetadataParser: 元数据解析器，解析 AOP 元数据
//
// # 核心功能
//
//   - 静态代理生成: 在编译时生成代理代码
//   - 模板引擎: 支持自定义代码生成模板
//   - 元数据解析: 解析注解和切面定义
//   - 代码格式化: 生成格式化的 Go 代码
//
// # 使用方式
//
// 使用 go generate 生成代码：
//
//	//go:generate go run github.com/xudefa/enhance/cmd/goaop
//
// 或使用代码生成器 API：
//
//	generator := generator.NewProxyGenerator()
//	generator.SetOutputDir("./generated")
//	generator.Generate(targetType, aspects)
//
// # 生成的代码
//
// 生成的代理代码包含：
//
//   - 代理结构体: 包装目标对象
//   - 方法拦截: 拦截目标方法调用
//   - 切面执行: 按顺序执行切面逻辑
//   - 异常处理: 处理切面执行中的异常
//
// # 性能优势
//
// 静态代理相比动态代理的优势：
//
//   - 无运行时反射开销
//   - 编译器优化更有效
//   - 代码可调试性更好
//   - 启动时间更短
package generator
