// Package lifecycle 提供生命周期管理功能，用于 enhance 框架。
//
// 该模块管理应用和组件的生命周期回调，包括初始化、启动、停止等阶段。
// 参考 Spring 的 Lifecycle 接口设计。
//
// # 架构设计
//
//   - ApplicationPhase: 应用生命周期阶段枚举
//   - PhaseListener: 生命周期阶段监听器接口
//   - LifecycleManager: 生命周期管理器
//   - LifecycleBuilder: 生命周期构建器，支持链式配置
//   - PhaseTransition: 阶段转换辅助结构
//   - BeanInitFunc: Bean 初始化钩子函数类型
//   - BeanDestroyFunc: Bean 销毁钩子函数类型
//
// # 核心功能
//
//   - 阶段管理: 支持 INIT → RUNNING → STOPPED 三阶段转换
//   - 监听器: 支持注册阶段变更监听器
//   - Bean 生命周期: 支持 Bean 初始化和销毁钩子
//   - 错误处理: 支持自定义错误处理回调
//
// # 使用方式
//
// 实现 PhaseListener 接口：
//
//	type MyListener struct{}
//
//	func (m *MyListener) OnPhaseChange(oldPhase, newPhase lifecycle.ApplicationPhase) error {
//	    fmt.Printf("Phase changed: %s -> %s\n", oldPhase, newPhase)
//	    return nil
//	}
//
// 注册到生命周期管理器：
//
//	manager := lifecycle.NewLifecycleManager()
//	manager.AddListener(&MyListener{})
//	manager.SetPhase(lifecycle.PhaseRunning)
//
// # 执行顺序
//
// 阶段转换必须是正向的：INIT → RUNNING → STOPPED。
// 反向转换会返回错误。
package lifecycle

import (
	"context"
	"sync"
)

// ApplicationPhase 应用生命周期阶段。
//
// 简化为 3 阶段：PhaseInit → PhaseRunning → PhaseStopped
type ApplicationPhase int

// 应用生命周期阶段定义。
const (
	PhaseInit    ApplicationPhase = iota // 初始化阶段：创建容器、加载配置、注册 Bean
	PhaseRunning                         // 运行阶段：应用正常运行，处理请求
	PhaseStopped                         // 已停止：应用完全停止
)

// PhaseListener 生命周期阶段监听器。
type PhaseListener interface {
	// OnPhaseChange 处理阶段变更事件。
	OnPhaseChange(oldPhase, newPhase ApplicationPhase) error
}

// Hook 生命周期钩子接口。
type Hook interface {
	OnInit(ctx context.Context) error
	OnStart(ctx context.Context) error
	OnStop(ctx context.Context) error
}

// HookFunc 函数式钩子。
type HookFunc struct {
	onInit  func(context.Context) error
	onStart func(context.Context) error
	onStop  func(context.Context) error
}

// BeanInitFunc Bean 初始化钩子函数类型。
//
// 在 Bean 创建后调用，替代 Spring 风格的 InitializingBean 接口。
// 使用函数类型而非接口，更符合 Go 惯用法。
type BeanInitFunc func(bean any) error

// BeanDestroyFunc Bean 销毁钩子函数类型。
//
// 在 Bean 销毁前调用，替代 Spring 风格的 DisposableBean 接口。
// 使用函数类型而非接口，更符合 Go 惯用法。
type BeanDestroyFunc func(bean any) error

// LifecycleManager 生命周期管理器。
type LifecycleManager struct {
	mu        sync.RWMutex
	phase     ApplicationPhase
	listeners []PhaseListener
	onError   func(oldPhase, newPhase ApplicationPhase, err error)
}

// LifecycleBuilder 生命周期构建器，支持链式配置。
type LifecycleBuilder struct {
	initialPhase ApplicationPhase
	listeners    []PhaseListener
	onError      func(oldPhase, newPhase ApplicationPhase, err error)
	mu           sync.Mutex
}

// PhaseTransition 阶段转换辅助结构。
type PhaseTransition struct {
	OldPhase ApplicationPhase
	NewPhase ApplicationPhase
}

// isForwardTransition 检查是否为有效的正向阶段转换。
func isForwardTransition(oldPhase, newPhase ApplicationPhase) bool {
	return oldPhase >= PhaseInit && newPhase > oldPhase && newPhase <= PhaseStopped
}

var phaseNames = map[ApplicationPhase]string{
	PhaseInit:    "INIT",
	PhaseRunning: "RUNNING",
	PhaseStopped: "STOPPED",
}
