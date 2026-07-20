// Package condition 提供条件化注册功能，用于 enhance 框架。
//
// 该模块支持根据条件动态决定是否注册 Bean 或执行自动配置。
// 参考 Spring Boot 的 @Conditional 注解设计。
//
// # 架构设计
//
//   - Condition: 条件接口，定义条件判断逻辑
//   - ConditionContext: 条件上下文接口
//   - EnvironmentAccessor: 环境配置访问接口
//   - ContainerAccessor: 容器访问接口
//   - OnProperty: 基于配置属性的条件
//   - OnModuleLoaded: 基于模块加载的条件（替代 Java 的 OnClass）
//   - OnBean: 基于 Bean 存在的条件
//   - OnMissingBean: 基于 Bean 缺失的条件
//   - 条件组合: 支持 AND、OR、NOT 逻辑组合
//
// # 内置条件
//
//   - OnProperty: 检查配置属性是否存在或等于特定值
//   - OnModuleLoaded: 检查指定模块是否已加载（Go 替代方案）
//   - OnBean: 检查指定 Bean 是否已注册
//   - OnMissingBean: 检查指定 Bean 是否未注册
//   - OnMissingModule: 检查指定模块是否未加载
//
// # 使用方式
//
// 在自动配置中使用条件：
//
//	type MyAutoConfiguration struct{}
//
//	func (m *MyAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
//	    // 配置逻辑
//	    return nil
//	}
//
//	func init() {
//	    boot.RegisterAutoConfig(
//	        &MyAutoConfiguration{},
//	        condition.OnProperty("my.feature.enabled", "true"),
//	    )
//	}
//
// # 条件组合
//
// 支持多个条件的逻辑组合：
//
//	condition.And(
//	    condition.OnProperty("feature.enabled", "true"),
//	    condition.OnMissingBean("customService"),
//	)
//
//	condition.Or(
//	    condition.OnProperty("env", "dev"),
//	    condition.OnProperty("env", "test"),
//	)
package condition

import "reflect"

// keyLister 可列举键的接口，用于 OnPropertyPrefix 条件
type keyLister interface {
	Keys() []string
}

// Condition 条件接口。
//
// 参考 Spring Boot 的 @Conditional 注解体系。
// 用于在 AutoConfiguration 中控制 Bean 是否注册。
type Condition interface {
	// Matches 评估条件是否匹配。
	Matches(ctx ConditionContext) bool
	// String 返回条件的可读描述，用于日志输出。
	String() string
}

// EnvironmentAccessor 环境配置访问接口。
type EnvironmentAccessor interface {
	GetProperty(key string) (any, bool)
}

// ContainerAccessor 容器访问接口。
type ContainerAccessor interface {
	Has(id string) bool
}

// ConditionContext 条件上下文。
//
// 提供条件判断所需的环境、容器和一些便捷方法。
type ConditionContext interface {
	// Environment 返回环境配置，用于读取属性值。
	Environment() EnvironmentAccessor
	// Container 返回 DI 容器，用于检查 Bean 是否存在。
	Container() ContainerAccessor

	// GetBeanByType 从容器中按类型获取 Bean 实例。
	GetBeanByType(t reflect.Type) (any, bool)
	// HasProperty 检查属性是否存在。
	HasProperty(key string) bool
	// GetProperty 获取属性值。
	GetProperty(key string) (any, bool)
}
