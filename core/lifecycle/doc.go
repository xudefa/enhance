// Package lifecycle 定义了 Bean 的生命周期管理接口和实现。
//
// # 设计原则
//
//   - 接口优先：实现 LifecycleBean 接口，零反射开销
//   - 函数式回调：通过 Register[T] 的 WithInit/WithDestroy 选项注册回调
//   - 阶段监听：支持监听 Bean 生命周期阶段变化，用于日志、指标等横切关注点
//
// # 生命周期阶段
//
//	注册 → 实例化 → 依赖注入 → 初始化 → 使用中 → 销毁
//
// # 使用方式
//
// 方式一：实现 LifecycleBean 接口（推荐，零反射）
//
//	type MyService struct{}
//
//	func (s *MyService) Init() error {
//	    // 初始化逻辑
//	    return nil
//	}
//
//	func (s *MyService) Destroy() error {
//	    // 清理逻辑
//	    return nil
//	}
//
// 方式二：使用函数式选项
//
//	core.Register(container, "myService",
//	    func(c core.Container) *MyService { return &MyService{} },
//	    core.WithInit(func(s *MyService) error { ... }),
//	    core.WithDestroy(func(s *MyService) error { ... }),
//	)
//
// # 生命周期监听器
//
//	listener := lifecycle.LifecycleListenerFunc(func(name string, bean any, phase lifecycle.Phase) {
//	    log.Printf("Bean %s entered phase %v", name, phase)
//	})
//	lifecycleMgr.RegisterListener(listener)
package lifecycle

// Phase 定义 Bean 生命周期阶段。
type Phase int

const (
	// PhaseRegistered Bean 已注册到容器。
	PhaseRegistered Phase = iota

	// PhaseInstantiated Bean 实例已创建。
	PhaseInstantiated

	// PhaseDependenciesInjected Bean 依赖已注入。
	PhaseDependenciesInjected

	// PhaseInitialized Bean 初始化回调已执行。
	PhaseInitialized

	// PhaseReady Bean 就绪，可以正常使用。
	PhaseReady

	// PhaseDestroyed Bean 已销毁。
	PhaseDestroyed
)

// LifecycleBean Bean 生命周期接口。
//
// 实现此接口的 Bean 会在适当时机被容器调用对应方法。
// 这是推荐的方式，避免反射开销。
type LifecycleBean interface {
	// Init 初始化回调。
	// 在依赖注入完成后调用。
	Init() error

	// Destroy 销毁回调。
	// 在容器销毁时调用。
	Destroy() error
}

// LifecycleListener 生命周期监听器接口。
//
// 用于监听 Bean 生命周期阶段变化，
// 可用于日志记录、指标收集等横切关注点。
type LifecycleListener interface {
	// OnPhaseChange Bean 生命周期阶段变化时调用。
	//
	// 参数:
	//   - beanName: Bean 名称
	//   - bean: Bean 实例
	//   - phase: 当前阶段
	OnPhaseChange(beanName string, bean any, phase Phase)
}

// LifecycleListenerFunc 生命周期监听器函数类型。
//
// 方便使用函数式监听器，无需实现接口。
type LifecycleListenerFunc func(beanName string, bean any, phase Phase)

// LifecycleManager 生命周期管理器接口。
//
// 负责管理 Bean 的生命周期回调和阶段监听。
type LifecycleManager interface {
	// RegisterListener 注册生命周期监听器。
	RegisterListener(listener LifecycleListener)

	// NotifyPhaseChange 通知生命周期阶段变化。
	//
	// 参数:
	//   - beanName: Bean 名称
	//   - bean: Bean 实例
	//   - phase: 当前阶段
	NotifyPhaseChange(beanName string, bean any, phase Phase)

	// InvokeInit 调用 Bean 的初始化回调。
	// 支持 LifecycleBean 接口和函数式回调。
	//
	// 参数:
	//   - beanName: Bean 名称
	//   - bean: Bean 实例
	//   - initFunc: 函数式初始化回调（可选）
	//
	// 返回:
	//   - error: 错误信息
	InvokeInit(beanName string, bean any, initFunc func(any) error) error

	// InvokeDestroy 调用 Bean 的销毁回调。
	// 支持 LifecycleBean 接口和函数式回调。
	//
	// 参数:
	//   - beanName: Bean 名称
	//   - bean: Bean 实例
	//   - destroyFunc: 函数式销毁回调（可选）
	//
	// 返回:
	//   - error: 错误信息
	InvokeDestroy(beanName string, bean any, destroyFunc func(any) error) error

	// DestroyAll 销毁所有已注册的 Bean。
	// 按注册逆序调用 Destroy 回调。
	//
	// 返回:
	//   - error: 错误信息（首个错误）
	DestroyAll() error

	// RegisterBean 注册 Bean 的销毁回调。
	// 用于在容器销毁时调用 Bean 的 Destroy 方法。
	//
	// 参数:
	//   - name: Bean 名称
	//   - instance: Bean 实例
	//   - destroyFunc: 销毁回调函数
	RegisterBean(name string, instance any, destroyFunc func(any) error)
}
