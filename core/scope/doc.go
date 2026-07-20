// Package scope 定义了 Bean 的作用域管理接口和实现。
//
// # 设计原则
//
//   - 策略模式：不同作用域通过实现 Scope 接口扩展
//   - 并发安全：所有作用域实现必须保证线程安全
//   - 生命周期绑定：作用域与 Bean 生命周期紧密关联
//
// # 内置作用域
//
//   - Singleton：单例作用域，整个容器生命周期内只存在一个实例（默认）
//   - Prototype：原型作用域，每次获取都创建新实例
//
// # 并发优化
//
// 使用 sync.Map 替代 sync.RWMutex，优化读多写少的注册场景。
// 容器注册阶段发生在应用启动时，读取 Bean 实例的频率远高于创建新实例。
//
// # 扩展作用域
//
// 可以通过实现 Scope 接口自定义作用域，例如：
//   - Request：请求作用域，每个 HTTP 请求一个实例
//   - Session：会话作用域，每个用户会话一个实例
//
// # 使用示例
//
//	// 获取单例作用域
//	singletonScope := scopeRegistry.Get(scope.SingletonScope)
//
//	// 注册自定义作用域
//	scopeRegistry.Register("request", &RequestScope{})
package scope

// Scope 作用域接口，定义了 Bean 实例的获取和存储策略。
//
// 设计说明：
//   - 每种作用域实现此接口
//   - 容器根据 registry.BeanDef.Scope 选择对应的作用域实现
//   - 作用域实现负责管理 Bean 实例的生命周期
type Scope interface {
	// Get 获取 Bean 实例。
	// 如果实例不存在，调用 factory 创建。
	//
	// 参数:
	//   - beanID: Bean ID
	//   - factory: 工厂函数，用于创建 Bean 实例
	//
	// 返回:
	//   - any: Bean 实例
	//   - error: 错误信息
	Get(beanID string, factory func(c ...any) (any, error)) (any, error)

	// Remove 移除指定 ID的 Bean 实例。
	// 用于清理作用域内的 Bean 实例。
	//
	// 参数:
	//   - beanID: Bean ID
	Remove(beanID string)

	// Clear 清空作用域内所有 Bean 实例。
	// 通常在容器销毁时调用。
	Clear()
}

// ScopeRegistry 作用域注册表，用于管理和扩展作用域。
//
// 设计说明：
//   - 内置 Singleton 和 Prototype 作用域
//   - 支持自定义作用域注册
//   - 线程安全
type ScopeRegistry interface {
	// Register 注册自定义作用域。
	//
	// 参数:
	//   - name: 作用域名称
	//   - scope: 作用域实现
	Register(name string, scope Scope)

	// Get 获取指定名称的作用域。
	// 如果未找到，返回 nil。
	//
	// 参数:
	//   - name: 作用域名称
	//
	// 返回:
	//   - Scope: 作用域实现
	Get(name string) Scope

	// Has 检查是否存在指定名称的作用域。
	//
	// 参数:
	//   - name: 作用域名称
	//
	// 返回:
	//   - bool: 是否存在
	Has(name string) bool
}

// ScopeNames 内置作用域名称常量。
const (
	SingletonScope = "singleton"
	PrototypeScope = "prototype"
)
