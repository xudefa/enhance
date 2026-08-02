package event

import (
	"fmt"
	"time"
)

// EventBusBuilder 事件总线构建器，支持链式配置
type EventBusBuilder struct {
	listeners map[string][]EventListener
}

// NewEventBusBuilder 创建事件总线构建器
func NewEventBusBuilder() *EventBusBuilder {
	return &EventBusBuilder{
		listeners: make(map[string][]EventListener),
	}
}

// Subscribe 订阅事件
func (b *EventBusBuilder) Subscribe(eventType string, listener EventListener) *EventBusBuilder {
	b.listeners[eventType] = append(b.listeners[eventType], listener)
	return b
}

// OnApplicationStarted 注册应用启动事件监听器
func (b *EventBusBuilder) OnApplicationStarted(listener EventListener) *EventBusBuilder {
	return b.Subscribe(EventApplicationStarted, listener)
}

// OnApplicationReady 注册应用就绪事件监听器
func (b *EventBusBuilder) OnApplicationReady(listener EventListener) *EventBusBuilder {
	return b.Subscribe(EventApplicationReady, listener)
}

// OnApplicationStopped 注册应用停止事件监听器
func (b *EventBusBuilder) OnApplicationStopped(listener EventListener) *EventBusBuilder {
	return b.Subscribe(EventApplicationStopped, listener)
}

// Build 构建事件总线
func (b *EventBusBuilder) Build() *EventBus {
	bus := NewEventBus()

	// 注册所有监听器
	for eventType, listeners := range b.listeners {
		for _, listener := range listeners {
			bus.Subscribe(eventType, listener)
		}
	}

	return bus
}

// AsyncPublisherBuilder 异步事件发布器构建器
type AsyncPublisherBuilder struct {
	bus         *EventBus
	workerCount int
	errHandler  func(error, ApplicationEvent)
}

// NewAsyncPublisherBuilder 创建异步事件发布器构建器
func NewAsyncPublisherBuilder() *AsyncPublisherBuilder {
	return &AsyncPublisherBuilder{
		workerCount: 10, // 默认工作协程数
	}
}

// Bus 设置事件总线
func (b *AsyncPublisherBuilder) Bus(bus *EventBus) *AsyncPublisherBuilder {
	b.bus = bus
	return b
}

// WorkerCount 设置工作协程池大小
func (b *AsyncPublisherBuilder) WorkerCount(count int) *AsyncPublisherBuilder {
	b.workerCount = count
	return b
}

// ErrorHandler 设置错误处理器
func (b *AsyncPublisherBuilder) ErrorHandler(handler func(error, ApplicationEvent)) *AsyncPublisherBuilder {
	b.errHandler = handler
	return b
}

// Build 构建异步事件发布器
func (b *AsyncPublisherBuilder) Build() (*AsyncPublisher, error) {
	if b.bus == nil {
		return nil, fmt.Errorf("event bus cannot be nil")
	}

	if b.workerCount <= 0 {
		return nil, fmt.Errorf("worker count must be positive, got %d", b.workerCount)
	}

	publisher := NewAsyncPublisher(b.bus,
		WithWorkerCount(b.workerCount),
		WithErrorHandler(b.errHandler),
	)

	return publisher, nil
}

// MustBuild 构建异步事件发布器，失败则panic
func (b *AsyncPublisherBuilder) MustBuild() *AsyncPublisher {
	publisher, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build async publisher: %v", err))
	}
	return publisher
}

// BaseEventBuilder 基础事件构建器
type BaseEventBuilder struct {
	event BaseEvent
}

// NewBaseEventBuilder 创建基础事件构建器
func NewBaseEventBuilder() *BaseEventBuilder {
	return &BaseEventBuilder{}
}

// Type 设置事件类型
func (b *BaseEventBuilder) Type(eventType string) *BaseEventBuilder {
	b.event.EventType = eventType
	return b
}

// Timestamp 设置事件时间戳
func (b *BaseEventBuilder) Timestamp(timestamp time.Time) *BaseEventBuilder {
	b.event.EventTime = timestamp
	return b
}

// Now 设置事件时间戳为当前时间
func (b *BaseEventBuilder) Now() *BaseEventBuilder {
	b.event.EventTime = time.Now()
	return b
}

// Build 构建基础事件
func (b *BaseEventBuilder) Build() *BaseEvent {
	return &b.event
}

// Publish 发布事件到事件总线
func (b *BaseEventBuilder) Publish(bus *EventBus) {
	bus.Publish(b.Build())
}

// ApplicationStartedEvent 创建应用启动事件
func ApplicationStartedEvent() *BaseEventBuilder {
	return NewBaseEventBuilder().Type(EventApplicationStarted).Now()
}

// ApplicationReadyEvent 创建应用就绪事件
func ApplicationReadyEvent() *BaseEventBuilder {
	return NewBaseEventBuilder().Type(EventApplicationReady).Now()
}

// ApplicationStoppedEvent 创建应用停止事件
func ApplicationStoppedEvent() *BaseEventBuilder {
	return NewBaseEventBuilder().Type(EventApplicationStopped).Now()
}

// ContextRefreshedEvent 创建上下文刷新事件
func ContextRefreshedEvent() *BaseEventBuilder {
	return NewBaseEventBuilder().Type(EventContextRefreshed).Now()
}

// EnvironmentPreparedEvent 创建环境准备事件
func EnvironmentPreparedEvent() *BaseEventBuilder {
	return NewBaseEventBuilder().Type(EventEnvironmentPrepared).Now()
}
