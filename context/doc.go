// Package context 提供应用上下文管理，用于 enhance 框架。
//
// 该模块管理应用生命周期中的上下文信息，包括容器、环境配置、事件总线等核心组件。
// 是应用启动和运行的核心枢纽。
//
// # 架构设计
//
//   - ApplicationContext: 应用上下文接口，统一封装 IoC 容器、Environment、生命周期和事件系统
//   - EventPublisher: 事件发布器接口，解耦事件发布逻辑
//   - AsyncEventPublisher: 异步事件发布器接口，支持非阻塞事件发布
//   - EventBusAccess: 事件总线访问接口，支持事件的发布与订阅

//
// # 核心功能
//
//   - 容器管理: 提供 IoC 容器访问，管理 Bean 的注册和获取
//   - 环境配置: 提供环境配置访问，支持多级配置源
//   - 事件总线: 提供事件发布和订阅，支持优先级和异步
//   - 生命周期: 管理应用启动和关闭，控制状态流转
//   - 刷新作用域: 支持配置热更新时的 Bean 刷新
//
// # 使用方式
//
// 获取应用上下文：
//
//	ctx := boot.NewApplication()
//	container := ctx.Container()
//	env := ctx.Environment()
//
// 从容器获取 Bean：
//
//	service, err := container.Get("myService")
//
// 发布事件：
//
//	ctx.PublishEvent(event.NewEvent("myEvent"))
//
// # 设计模式
//
//   - Facade: ApplicationContext 作为框架核心组件的统一入口
//   - Adapter: asyncEventPublisherAdapter 适配 event.AsyncPublisher 为 AsyncEventPublisher
//   - Singleton: globalClassLoader 全局共享 ClassLoader 实例
package context

import (
	"context"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/config/refresh"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

// EventPublisher 事件发布器接口。
//
// 用于解耦事件发布逻辑，便于测试和替换实现。
type EventPublisher interface {
	// Publish 发布事件。
	Publish(event event.ApplicationEvent)
}

// AsyncEventPublisher 异步事件发布器接口。
//
// 支持异步事件发布，不阻塞调用者。
type AsyncEventPublisher interface {
	// PublishAsync 异步发布事件，使用背景上下文。
	PublishAsync(event event.ApplicationEvent)

	// PublishAsyncWithCtx 异步发布事件，使用指定上下文（支持超时控制）。
	PublishAsyncWithCtx(ctx context.Context, event event.ApplicationEvent)
}

// EventBusAccess 事件总线访问接口。
//
// 定义事件总线的公共操作，支持 EventBus 和 EventBusWithOrdering。
type EventBusAccess interface {
	// Publish 发布事件。
	Publish(event event.ApplicationEvent)

	// Subscribe 订阅指定类型的事件。
	Subscribe(eventType string, listener event.EventListener)

	// Unsubscribe 取消订阅指定类型的事件。
	Unsubscribe(eventType string, target event.EventListener)
}

// ApplicationContext 应用上下文接口。
//
// 作为框架核心组件的统一入口（Facade 模式），聚合对各个子系统的访问。
// 不重复定义已有方法，通过 Container() 获取 core.Container 来操作 Bean，
// 通过 Lifecycle() 获取 lifecycle.LifecycleManager 来控制生命周期。
//
// 该接口遵循小接口原则，仅包含 7 个访问方法：
//   - Container: 获取依赖注入容器
//   - Environment: 获取环境配置
//   - Lifecycle: 获取生命周期管理器
//   - EventBus: 获取事件总线访问
//   - EventPublisher: 获取事件发布器
//   - AsyncEventPublisher: 获取异步事件发布器
//   - RefreshScopeManager: 获取刷新作用域管理器
type ApplicationContext interface {
	// Container 返回 IoC 容器实例。
	Container() core.Container

	// Environment 返回环境配置实例。
	Environment() *environment.Environment

	// Lifecycle 返回生命周期管理器。
	Lifecycle() *lifecycle.LifecycleManager

	// EventBus 返回事件总线访问接口（支持 EventBus 和 EventBusWithOrdering）。
	EventBus() EventBusAccess

	// EventPublisher 返回事件发布器接口（解耦事件发布）。
	EventPublisher() EventPublisher

	// AsyncEventPublisher 返回异步事件发布器接口。
	AsyncEventPublisher() AsyncEventPublisher

	// RefreshScopeManager 返回刷新作用域管理器。
	RefreshScopeManager() *refresh.RefreshScopeManager
}
