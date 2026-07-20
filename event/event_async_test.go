package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEventBus_AsyncPublish 测试异步发布事件
func TestEventBus_AsyncPublish(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var called int32
	var wg sync.WaitGroup
	wg.Add(100)

	bus.Subscribe("async.event", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
		wg.Done()
	})

	for range 100 {
		go func() {
			bus.Publish(&BaseEvent{
				EventType: "async.event",
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
		t.Fatalf("timeout waiting for events, got %d/%d", atomic.LoadInt32(&called), 100)
	}

	if atomic.LoadInt32(&called) != 100 {
		t.Errorf("Expected 100 calls, got %d", called)
	}
}

// TestEventBus_ConcurrentSubscribe 测试并发订阅
func TestEventBus_ConcurrentSubscribe(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var called int32
	var subscribeWg sync.WaitGroup
	var eventWg sync.WaitGroup

	for range 100 {
		subscribeWg.Add(1)
		go func() {
			defer subscribeWg.Done()
			bus.Subscribe("concurrent.event", func(e ApplicationEvent) {
				atomic.AddInt32(&called, 1)
				eventWg.Done()
			})
		}()
	}

	subscribeWg.Wait()

	// 在发布前设置好 WaitGroup
	eventWg.Add(100)
	bus.Publish(&BaseEvent{
		EventType: "concurrent.event",
		EventTime: time.Now(),
	})

	// 使用 WaitGroup 等待所有监听器被调用
	done := make(chan struct{})
	go func() {
		eventWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 所有监听器已被调用
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for listeners, got %d/%d", atomic.LoadInt32(&called), 100)
	}

	if atomic.LoadInt32(&called) != 100 {
		t.Errorf("Expected 100 calls, got %d", called)
	}
}

// TestEventBus_MixedAsyncOperations 测试混合异步操作
func TestEventBus_MixedAsyncOperations(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var subscribeCount int32
	var publishCount int32
	var subscribeWg sync.WaitGroup
	var publishWg sync.WaitGroup

	// 先完成所有订阅
	for range 50 {
		subscribeWg.Add(1)
		go func() {
			defer subscribeWg.Done()
			bus.Subscribe("mixed.event", func(e ApplicationEvent) {
				atomic.AddInt32(&subscribeCount, 1)
			})
		}()
	}

	subscribeWg.Wait()

	// 发布事件
	for range 50 {
		publishWg.Add(1)
		go func() {
			defer publishWg.Done()
			bus.Publish(&BaseEvent{
				EventType: "mixed.event",
				EventTime: time.Now(),
			})
			atomic.AddInt32(&publishCount, 1)
		}()
	}

	publishWg.Wait()

	// 等待一小段时间确保所有事件处理完成
	done := make(chan struct{})
	go func() {
		// 使用轮询等待订阅计数稳定
		lastCount := atomic.LoadInt32(&subscribeCount)
		for range 20 {
			time.Sleep(5 * time.Millisecond)
			currentCount := atomic.LoadInt32(&subscribeCount)
			if currentCount == lastCount {
				break
			}
			lastCount = currentCount
		}
		close(done)
	}()

	select {
	case <-done:
		// 所有事件已处理完成
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for events, subscribeCount=%d, publishCount=%d",
			atomic.LoadInt32(&subscribeCount), atomic.LoadInt32(&publishCount))
	}

	if atomic.LoadInt32(&publishCount) != 50 {
		t.Errorf("Expected 50 publishes, got %d", publishCount)
	}
}
