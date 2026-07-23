package event

import (
	"reflect"
	"sort"
	"time"
)

// NewEventBusWithOrdering 创建支持优先级和过滤条件的事件总线。
func NewEventBusWithOrdering() *EventBusWithOrdering {
	return &EventBusWithOrdering{}
}

// Subscribe 订阅事件（向后兼容，等价于 Order=0 无条件的监听器）
func (b *EventBusWithOrdering) Subscribe(eventType string, listener EventListener) {
	b.SubscribeWithConfig(eventType, ListenerConfig{
		Handler: listener,
		Order:   0,
	})
}

// SubscribeWithConfig 带配置的订阅
func (b *EventBusWithOrdering) SubscribeWithConfig(eventType string, config ListenerConfig) {
	b.mu.Lock()
	b.nextID++
	id := b.nextID
	b.mu.Unlock()

	// 使用 CAS 无锁算法添加监听器
	for {
		oldValue, _ := b.listeners.LoadOrStore(eventType, &listenerSlice{})
		old := oldValue.(*listenerSlice)

		// 在锁外创建新列表，减少锁持有时间
		newList := make([]orderedListener, len(old.list)+1)
		copy(newList, old.list)
		newList[len(old.list)] = orderedListener{
			config:    config,
			original:  config.Handler,
			wrapperID: id,
		}
		newSlice := &listenerSlice{list: newList}

		if b.listeners.CompareAndSwap(eventType, old, newSlice) {
			return
		}
	}
}

// SubscribeOnce 订阅事件，仅消费一次后自动取消订阅
func (b *EventBusWithOrdering) SubscribeOnce(eventType string, listener EventListener) {
	b.mu.Lock()
	b.nextID++
	id := b.nextID
	b.mu.Unlock()

	var wrapper EventListener = func(e ApplicationEvent) {
		listener(e)
		b.Unsubscribe(eventType, listener)
	}

	// 使用 CAS 无锁算法添加监听器
	for {
		oldValue, _ := b.listeners.LoadOrStore(eventType, &listenerSlice{})
		old := oldValue.(*listenerSlice)

		// 在锁外创建新列表，减少锁持有时间
		newList := make([]orderedListener, len(old.list)+1)
		copy(newList, old.list)
		newList[len(old.list)] = orderedListener{
			config:    ListenerConfig{Handler: wrapper},
			original:  listener,
			wrapperID: id,
		}
		newSlice := &listenerSlice{list: newList}

		if b.listeners.CompareAndSwap(eventType, old, newSlice) {
			return
		}
	}
}

// Unsubscribe 取消订阅
func (b *EventBusWithOrdering) Unsubscribe(eventType string, target EventListener) {
	if target == nil {
		return
	}

	targetPtr := reflect.ValueOf(target).Pointer()

	// 使用 CAS 无锁算法移除监听器
	for {
		oldValue, ok := b.listeners.Load(eventType)
		if !ok {
			return
		}
		old := oldValue.(*listenerSlice)

		// 在锁外查找索引
		found := -1
		for i, ol := range old.list {
			if reflect.ValueOf(ol.original).Pointer() == targetPtr {
				found = i
				break
			}
		}

		if found == -1 {
			return
		}

		// 在锁外创建新列表
		newList := make([]orderedListener, len(old.list)-1)
		copy(newList, old.list[:found])
		copy(newList[found:], old.list[found+1:])
		newSlice := &listenerSlice{list: newList}

		if b.listeners.CompareAndSwap(eventType, old, newSlice) {
			return
		}
	}
}

// Publish 发布事件，按优先级排序并应用过滤条件
//
// 性能优化：
//   - 使用预分配切片避免动态扩容
//   - 快照后释放锁，减少锁持有时间
func (b *EventBusWithOrdering) Publish(event ApplicationEvent) {
	oldValue, ok := b.listeners.Load(event.Type())
	if !ok {
		return
	}
	listeners := oldValue.(*listenerSlice).list

	if len(listeners) == 0 {
		return
	}

	// 预分配切片容量，避免动态扩容
	snapshot := make([]orderedListener, len(listeners))
	copy(snapshot, listeners)

	sort.SliceStable(snapshot, func(i, j int) bool {
		return snapshot[i].config.Order < snapshot[j].config.Order
	})

	// 执行监听器
	for i := range snapshot {
		ol := &snapshot[i]
		// 应用过滤条件
		if ol.config.Condition != nil && !ol.config.Condition(event) {
			continue
		}

		if ol.config.Async {
			go ol.config.Handler(event)
			continue
		}
		ol.config.Handler(event)
	}
}

// Listeners 返回指定事件类型的监听器数量
func (b *EventBusWithOrdering) Listeners(eventType string) int {
	oldValue, ok := b.listeners.Load(eventType)
	if !ok {
		return 0
	}
	return len(oldValue.(*listenerSlice).list)
}

// Clear 清除指定事件类型的所有监听器
func (b *EventBusWithOrdering) Clear(eventType string) {
	b.listeners.Delete(eventType)
}

// ClearAll 清除所有监听器
func (b *EventBusWithOrdering) ClearAll() {
	b.listeners.Range(func(key, value any) bool {
		b.listeners.Delete(key)
		return true
	})
}

// NewLegacyEventBusAdapter 创建适配器。
func NewLegacyEventBusAdapter(bus *EventBusWithOrdering) *LegacyEventBusAdapter {
	return &LegacyEventBusAdapter{bus: bus}
}

// Publish 转发到 EventBusWithOrdering
func (a *LegacyEventBusAdapter) Publish(event ApplicationEvent) {
	a.bus.Publish(event)
}

// Subscribe 转发到 EventBusWithOrdering
func (a *LegacyEventBusAdapter) Subscribe(eventType string, listener EventListener) {
	a.bus.Subscribe(eventType, listener)
}

// Unsubscribe 转发到 EventBusWithOrdering
func (a *LegacyEventBusAdapter) Unsubscribe(eventType string, target EventListener) {
	a.bus.Unsubscribe(eventType, target)
}

// 便捷构造函数

// NewListenerConfig 创建监听器配置
func NewListenerConfig(handler EventListener) ListenerConfig {
	return ListenerConfig{Handler: handler}
}

// WithOrder 设置优先级
func (c ListenerConfig) WithOrder(order int) ListenerConfig {
	c.Order = order
	return c
}

// WithCondition 设置过滤条件
func (c ListenerConfig) WithCondition(cond func(ApplicationEvent) bool) ListenerConfig {
	c.Condition = cond
	return c
}

// WithAsync 设置异步执行
func (c ListenerConfig) WithAsync(async bool) ListenerConfig {
	c.Async = async
	return c
}

// 便捷过滤条件函数

// ConditionAlways 始终返回 true 的过滤条件
func ConditionAlways() func(ApplicationEvent) bool {
	return func(ApplicationEvent) bool { return true }
}

// ConditionType 按事件类型过滤
func ConditionType(types ...string) func(ApplicationEvent) bool {
	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	return func(e ApplicationEvent) bool {
		return typeSet[e.Type()]
	}
}

// ConditionAfter 按时间戳过滤，仅处理指定时间之后的事件
func ConditionAfter(t time.Time) func(ApplicationEvent) bool {
	return func(e ApplicationEvent) bool {
		return e.Timestamp().After(t)
	}
}

// ConditionBefore 按时间戳过滤，仅处理指定时间之前的事件
func ConditionBefore(t time.Time) func(ApplicationEvent) bool {
	return func(e ApplicationEvent) bool {
		return e.Timestamp().Before(t)
	}
}
