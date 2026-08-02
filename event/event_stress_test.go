package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEventBus_StressHighConcurrencyPublish 测试高并发发布事件
func TestEventBus_StressHighConcurrencyPublish(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var called int32
	var wg sync.WaitGroup
	wg.Add(10000)

	bus.Subscribe("stress.event", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	})

	for range 10000 {
		go func() {
			bus.Publish(&BaseEvent{
				EventType: "stress.event",
				EventTime: time.Now(),
			})
		}()
	}

	// 使用 WaitGroup 等待所有事件处理完成
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 所有事件已处理完成
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for events, got %d/%d", atomic.LoadInt32(&called), 10000)
	}

	if atomic.LoadInt32(&called) != 10000 {
		t.Errorf("Expected 10000 calls, got %d", called)
	}
}

// TestEventBus_StressHighConcurrencySubscribe 测试高并发订阅
func TestEventBus_StressHighConcurrencySubscribe(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var called int32
	var wg sync.WaitGroup
	var subscribeWg sync.WaitGroup

	for i := range 1000 {
		subscribeWg.Add(1)
		go func(id int) {
			defer subscribeWg.Done()
			bus.Subscribe("stress.subscribe", func(e ApplicationEvent) {
				atomic.AddInt32(&called, 1)
				wg.Done()
			})
		}(i)
	}

	subscribeWg.Wait()

	// 在发布前设置好 WaitGroup
	wg.Add(1000)
	bus.Publish(&BaseEvent{
		EventType: "stress.subscribe",
		EventTime: time.Now(),
	})

	// 使用 WaitGroup 等待所有监听器被调用
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 所有监听器已被调用
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for listeners, got %d/%d", atomic.LoadInt32(&called), 1000)
	}

	if atomic.LoadInt32(&called) != 1000 {
		t.Errorf("Expected 1000 calls, got %d", called)
	}
}

// TestEventBus_StressMultipleEventTypes 测试多种事件类型的压力测试
func TestEventBus_StressMultipleEventTypes(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var called int32
	var wg sync.WaitGroup
	wg.Add(10000)

	for i := range 10 {
		eventType := fmt.Sprintf("stress.event.%d", i)
		bus.Subscribe(eventType, func(e ApplicationEvent) {
			atomic.AddInt32(&called, 1)
			wg.Done()
		})
	}

	for i := range 10000 {
		go func(id int) {
			eventType := fmt.Sprintf("stress.event.%d", id%10)
			bus.Publish(&BaseEvent{
				EventType: eventType,
				EventTime: time.Now(),
			})
		}(i)
	}

	// 使用 WaitGroup 等待所有事件处理完成
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 所有事件已处理完成
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for events, got %d/%d", atomic.LoadInt32(&called), 10000)
	}

	if atomic.LoadInt32(&called) != 10000 {
		t.Errorf("Expected 10000 calls, got %d", called)
	}
}
