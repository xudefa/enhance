// Package event 提供应用事件驱动支持，用于 enhance 框架。
//
// 该模块提供完整的事件发布/订阅机制，参考 Spring 的 ApplicationEvent/ApplicationListener 模式。
// 支持同步和异步事件处理、事务性事件、死信队列等高级功能。
//
// # 设计原则
//
//   - 解耦组件：通过事件机制实现组件间的松耦合通信
//   - 灵活扩展：支持自定义事件类型和监听器，易于扩展
//   - 并发安全：事件总线支持并发安全的发布/订阅操作
//   - 零外部依赖：核心实现仅使用 Go 标准库
//
// # 架构设计
//
//   - ApplicationEvent: 应用事件接口
//   - EventBus: 事件总线，支持发布/订阅
//   - BaseEvent: 基础事件实现
//   - EventListener: 事件监听器函数类型
//   - ListenerConfig: 监听器配置
//   - EventBusWithOrdering: 支持优先级和过滤条件的事件总线
//
// # 核心功能
//
//   - 事件发布: 支持同步和异步事件发布
//   - 事件订阅: 支持按事件类型订阅
//   - 异步处理: 支持异步事件处理，提升性能
//   - 事务性事件: 支持事务提交后触发事件
//   - 死信队列: 处理失败事件，支持重试
//
// # 使用方式
//
// 定义事件：
//
//	type UserCreatedEvent struct {
//	    *event.BaseEvent
//	    UserID int64
//	}
//
// 发布事件：
//
//	bus := event.NewEventBus()
//	bus.Publish(&UserCreatedEvent{
//	    BaseEvent: &event.BaseEvent{EventType: "user.created"},
//	    UserID:    123,
//	})
//
// 订阅事件：
//
//	bus.Subscribe(func(e event.ApplicationEvent) {
//	    if evt, ok := e.(*UserCreatedEvent); ok {
//	        fmt.Println("User created:", evt.UserID)
//	    }
//	})
//
// # 异步事件
//
// 使用异步监听器处理事件：
//
//	bus.SubscribeAsync(func(e event.ApplicationEvent) {
//	    // 异步处理
//	})
//
// # 事务性事件
//
// 使用事务性监听器，在事务提交后触发：
//
//	bus.SubscribeTransactional(func(e event.ApplicationEvent) {
//	    // 事务提交后触发
//	})
package event

import (
	"sync"
	"time"
)

// ApplicationEvent 应用事件接口。
//
// 所有应用事件必须实现此接口。事件通过 Type() 返回的类型字符串进行路由和分发。
//
// # 设计原则
//
//   - 使用字符串类型标识事件，而非反射类型，提高灵活性和可读性
//   - 支持任意结构体实现事件接口，无需继承基类
//   - 时间戳用于事件排序和审计
type ApplicationEvent interface {
	// Type 返回事件类型字符串，用于事件路由和匹配。
	Type() string

	// Timestamp 返回事件发生的时间戳。
	Timestamp() time.Time
}

// EventListener 事件监听器函数类型。
//
// 接收 ApplicationEvent 参数，处理事件通知。
// 监听器在事件发布时同步调用。
type EventListener func(event ApplicationEvent)

// ListenerConfig 事件监听器配置。
//
// 提供比简单函数签名更丰富的监听器控制能力。
//
// # 字段说明
//
//   - Handler: 事件处理函数（必填）
//   - Order: 执行优先级，数值越小越先执行，默认 0
//   - Condition: 过滤条件函数，返回 false 时跳过该监听器
//   - Async: 是否异步执行，默认 false
//
// # 使用示例
//
//	bus.SubscribeWithConfig("MyEvent", event.ListenerConfig{
//	    Handler: func(e event.ApplicationEvent) {
//	        fmt.Println("处理事件:", e.Type())
//	    },
//	    Order: 10,
//	    Condition: func(e event.ApplicationEvent) bool {
//	        return e.Type() == "MyEvent"
//	    },
//	})
type ListenerConfig struct {
	Handler   EventListener               // 事件处理函数（必填）
	Order     int                         // 执行优先级，数值越小越先执行
	Condition func(ApplicationEvent) bool // 过滤条件，返回 false 时跳过
	Async     bool                        // 是否异步执行
}

// BaseEvent 基础事件实现。
//
// 可直接使用，也支持嵌入到自定义事件结构体中。
// 如果 EventTime 未设置，Timestamp() 会自动返回当前时间。
//
// # 使用示例
//
//	// 直接使用
//	evt := &event.BaseEvent{EventType: "user.created"}
//
//	// 嵌入到自定义事件
//	type UserCreatedEvent struct {
//	    event.BaseEvent
//	    UserID int
//	}
type BaseEvent struct {
	EventType string    // 事件类型
	EventTime time.Time // 事件发生时间（可选，为空时自动使用当前时间）
}

// listenerList 监听器列表包装器，用于 sync.Map 存储。
//
// sync.Map.CompareAndSwap 要求值可比较，切片不可比较，
// 因此使用指针包装器来支持 CAS 操作。
type listenerList struct {
	listeners []EventListener
}

// listenerSlice 包装器，使 slice 可用于 atomic.Value 的 CAS 操作。
type listenerSlice struct {
	list []orderedListener
}

// orderedListener 内部排序后的监听器表示。
type orderedListener struct {
	config    ListenerConfig
	original  EventListener // 用于 Unsubscribe 比较
	wrapperID int           // 唯一标识，用于区分相同函数多次注册
}

// EventBus 事件总线。
//
// 负责事件的发布与订阅管理，支持多监听器注册。
// 线程安全，支持并发发布和订阅。
// 使用 sync.Map 优化读多写少场景的性能。
//
// # 性能优化
//
//   - 使用 sync.Map 存储监听器，无锁读取
//   - 使用 CAS 操作实现无锁订阅更新
//   - 避免 range 分配迭代器，使用索引遍历
//
// # 并发安全
//
// EventBus 的所有方法都是并发安全的。
// 订阅和取消订阅使用 CAS 操作实现无锁更新，
// 发布事件使用无锁读取，性能优异。
type EventBus struct {
	listeners sync.Map // map[string]*listenerList
}

// EventBusWithOrdering 支持优先级和过滤条件的事件总线。
//
// 在原有 EventBus 基础上扩展，支持：
//   - 监听器优先级排序
//   - 条件过滤
//   - 异步执行
//   - 向后兼容原有的 Subscribe/Publish API
//
// 使用 sync.Map 优化读多写少场景的并发性能。
type EventBusWithOrdering struct {
	mu        sync.Mutex // 仅用于写操作（Subscribe/Unsubscribe）
	listeners sync.Map   // map[string]*listenerSlice
	nextID    int        // 监听器 ID 计数器
	wg        sync.WaitGroup
	closeMu   sync.Mutex // 保护 wg.Add/Wait 之间的竞态
}

// LegacyEventBusAdapter 将 EventBusWithOrdering 适配为 EventBus 接口。
// 用于需要 *EventBus 类型的场景。
type LegacyEventBusAdapter struct {
	bus *EventBusWithOrdering
}

// 内置事件类型常量。
//
// 这些是 enhance 框架生命周期中的标准事件类型，
// 应用可以在这些事件发生时注册监听器执行自定义逻辑。
const (
	// EventEnvironmentPrepared 环境配置准备完成事件。
	// 在环境配置加载完成后触发。
	EventEnvironmentPrepared = "EnvironmentPrepared"

	// EventContextRefreshed 应用上下文刷新完成事件。
	// 在应用上下文刷新完成后触发。
	EventContextRefreshed = "ContextRefreshed"

	// EventApplicationStarted 应用启动事件。
	// 在应用开始启动时触发。
	EventApplicationStarted = "ApplicationStarted"

	// EventApplicationReady 应用就绪事件。
	// 在应用完全启动并准备好处理请求时触发。
	EventApplicationReady = "ApplicationReady"

	// EventApplicationStopped 应用停止事件。
	// 在应用停止时触发。
	EventApplicationStopped = "ApplicationStopped"
)

// AsyncPublisherBus 事件发布器接口。
//
// 定义事件发布的最小接口，供 AsyncPublisher 使用。
type AsyncPublisherBus interface {
	// Publish 发布事件。
	Publish(event ApplicationEvent)
}
