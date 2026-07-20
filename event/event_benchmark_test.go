package event

import (
	"fmt"
	"testing"
)

// TestEvent 测试用事件
type TestEvent struct {
	BaseEvent
	TypeStr string
	Data    string
}

func (e *TestEvent) Type() string {
	return e.TypeStr
}

// BenchmarkEventBus_Publish_NoListeners 测试无监听器时的发布性能
func BenchmarkEventBus_Publish_NoListeners(b *testing.B) {
	bus := NewEventBus()
	event := &TestEvent{TypeStr: "test.event", Data: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

// BenchmarkEventBus_Publish_SingleListener 测试单监听器发布性能
func BenchmarkEventBus_Publish_SingleListener(b *testing.B) {
	bus := NewEventBus()
	bus.Subscribe("test.event", func(e ApplicationEvent) {})
	event := &TestEvent{TypeStr: "test.event", Data: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

// BenchmarkEventBus_Publish_MultipleListeners 测试多监听器发布性能
func BenchmarkEventBus_Publish_MultipleListeners(b *testing.B) {
	bus := NewEventBus()
	for i := 0; i < 10; i++ {
		bus.Subscribe("test.event", func(e ApplicationEvent) {})
	}
	event := &TestEvent{TypeStr: "test.event", Data: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

// BenchmarkEventBus_Subscribe 测试订阅性能
func BenchmarkEventBus_Subscribe(b *testing.B) {
	bus := NewEventBus()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Subscribe(fmt.Sprintf("event.%d", i), func(e ApplicationEvent) {})
	}
}

// BenchmarkEventBus_ConcurrentPublish 测试并发发布性能
func BenchmarkEventBus_ConcurrentPublish(b *testing.B) {
	bus := NewEventBus()
	bus.Subscribe("test.event", func(e ApplicationEvent) {})
	event := &TestEvent{TypeStr: "test.event", Data: "test"}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish(event)
		}
	})
}

// BenchmarkEventBus_ConcurrentSubscribe 测试并发订阅性能
func BenchmarkEventBus_ConcurrentSubscribe(b *testing.B) {
	bus := NewEventBus()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			bus.Subscribe("test.event", func(e ApplicationEvent) {})
			i++
		}
	})
}

// BenchmarkEventBus_DifferentListenerCounts 测试不同监听器数量的性能对比
func BenchmarkEventBus_DifferentListenerCounts(b *testing.B) {
	b.Run("1-Listener", func(b *testing.B) {
		bus := NewEventBus()
		bus.Subscribe("test.event", func(e ApplicationEvent) {})
		event := &TestEvent{TypeStr: "test.event", Data: "test"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bus.Publish(event)
		}
	})

	b.Run("5-Listeners", func(b *testing.B) {
		bus := NewEventBus()
		for i := 0; i < 5; i++ {
			bus.Subscribe("test.event", func(e ApplicationEvent) {})
		}
		event := &TestEvent{TypeStr: "test.event", Data: "test"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bus.Publish(event)
		}
	})

	b.Run("10-Listeners", func(b *testing.B) {
		bus := NewEventBus()
		for i := 0; i < 10; i++ {
			bus.Subscribe("test.event", func(e ApplicationEvent) {})
		}
		event := &TestEvent{TypeStr: "test.event", Data: "test"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bus.Publish(event)
		}
	})

	b.Run("50-Listeners", func(b *testing.B) {
		bus := NewEventBus()
		for i := 0; i < 50; i++ {
			bus.Subscribe("test.event", func(e ApplicationEvent) {})
		}
		event := &TestEvent{TypeStr: "test.event", Data: "test"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bus.Publish(event)
		}
	})
}
