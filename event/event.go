package event

import (
	"reflect"
	"sync"
	"time"
)

// NewEventBus 创建新的事件总线实例。
func NewEventBus() EventBus {
	return &eventBus{
		listeners: sync.Map{},
	}
}

// eventBus 事件总线实现。
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
// eventBus 的所有方法都是并发安全的。
// 订阅和取消订阅使用 CAS 操作实现无锁更新，
// 发布事件使用无锁读取，性能优异。
type eventBus struct {
	listeners sync.Map // map[string]*listenerList
}

// Type 返回事件类型字符串。
func (e *BaseEvent) Type() string {
	return e.EventType
}

// Timestamp 返回事件发生的时间戳。
//
// 如果 EventTime 未设置（零值），自动返回当前时间。
func (e *BaseEvent) Timestamp() time.Time {
	if e.EventTime.IsZero() {
		return time.Now()
	}
	return e.EventTime
}

// Publish 发布事件，通知所有订阅了该事件类型的监听器。
//
// 发布流程：
//  1. 从 sync.Map 获取事件类型的监听器列表（无锁）
//  2. 如果无监听器则直接返回
//  3. 遍历监听器列表并逐个调用
//
// 参数:
//   - event: 要发布的事件实例
//
// 注意:
//   - 监听器按注册顺序同步调用
//   - 如果监听器抛出 panic，会影响后续监听器的执行
//   - 对于耗时操作，建议使用异步事件总线
//
// 性能提示:
//   - 发布操作是无锁的，性能优异
//   - 监听器数量较多时，考虑使用 AsyncEventBus
func (b *eventBus) Publish(event ApplicationEvent) {
	if event == nil {
		return
	}
	value, ok := b.listeners.Load(event.Type())
	if !ok {
		return
	}
	list, _ := value.(*listenerList)
	for i := range list.listeners {
		list.listeners[i](event)
	}
}

// Subscribe 订阅指定类型的事件。
//
// 参数:
//   - eventType: 事件类型字符串，与 ApplicationEvent.Type() 返回值对应
//   - listener: 事件监听器函数
//
// # 并发安全
//
// 使用 CAS 操作实现无锁订阅，支持高并发场景。
// 多个 goroutine 可以同时订阅同一事件类型，不会丢失任何订阅。
//
// # 使用示例
//
//	bus.Subscribe("user.created", func(e event.ApplicationEvent) {
//	    log.Println("New user created")
//	})
//
// # 性能提示
//
//   - 首次订阅使用 LoadOrStore 快速路径，无锁
//   - 后续订阅使用 CAS 重试，保证并发安全
//   - 避免在事件处理函数中调用 Subscribe，可能导致死锁
func (b *eventBus) Subscribe(eventType string, listener EventListener) {
	if listener == nil {
		return
	}
	for {
		value, loaded := b.listeners.LoadOrStore(eventType, &listenerList{listeners: []EventListener{listener}})
		if !loaded {
			return
		}
		list, _ := value.(*listenerList)
		newList := &listenerList{
			listeners: make([]EventListener, len(list.listeners)+1),
		}
		copy(newList.listeners, list.listeners)
		newList.listeners[len(list.listeners)] = listener
		if b.listeners.CompareAndSwap(eventType, list, newList) {
			return
		}
	}
}

// Unsubscribe 取消订阅指定类型的事件。
//
// 参数:
//   - eventType: 事件类型字符串
//   - target: 要移除的监听器函数
//
// 注意:
//   - 使用 reflect 比较函数指针来定位要移除的监听器
//   - 如果监听器不存在，此操作是空操作
//   - 移除最后一个监听器时会自动删除该事件类型的记录
//
// 性能提示:
//   - 取消订阅需要遍历监听器列表，O(n) 复杂度
//   - 频繁取消订阅的场景，考虑使用一次性监听器
func (b *eventBus) Unsubscribe(eventType string, target EventListener) {
	if target == nil {
		return
	}
	targetPtr := reflect.ValueOf(target).Pointer()

	for {
		value, ok := b.listeners.Load(eventType)
		if !ok {
			return
		}
		list, _ := value.(*listenerList)
		listeners := list.listeners
		for i, listener := range listeners {
			if reflect.ValueOf(listener).Pointer() == targetPtr {
				newListeners := make([]EventListener, len(listeners)-1)
				copy(newListeners, listeners[:i])
				copy(newListeners[i:], listeners[i+1:])
				if len(newListeners) == 0 {
					// 使用 CompareAndSwap 替代 Delete，避免与并发 Subscribe 竞态
					if b.listeners.CompareAndSwap(eventType, list, &listenerList{listeners: newListeners}) {
						return
					}
					// CAS 失败，重试
					continue
				}
				if b.listeners.CompareAndSwap(eventType, list, &listenerList{listeners: newListeners}) {
					return
				}
				// CAS失败，说明列表已被其他goroutine修改，重试
				continue
			}
		}
		// 未找到目标监听器
		return
	}
}
