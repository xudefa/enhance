package event

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAsyncPublisher_Publish(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus)
	defer publisher.Close()

	var mu sync.Mutex
	received := 0
	done := make(chan struct{})

	bus.Subscribe("TestEvent", func(e ApplicationEvent) {
		mu.Lock()
		received++
		mu.Unlock()
		close(done)
	})

	ctx := context.Background()
	publisher.Publish(ctx, &BaseEvent{EventType: "TestEvent"})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}

	mu.Lock()
	if received != 1 {
		t.Errorf("expected 1, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_MultipleEvents(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus)
	defer publisher.Close()

	var mu sync.Mutex
	received := 0
	done := make(chan struct{})

	bus.Subscribe("MultiEvent", func(e ApplicationEvent) {
		mu.Lock()
		received++
		if received == 10 {
			mu.Unlock()
			close(done)
			return
		}
		mu.Unlock()
	})

	ctx := context.Background()
	for range 10 {
		publisher.Publish(ctx, &BaseEvent{EventType: "MultiEvent"})
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for events, received %d", received)
	}

	mu.Lock()
	if received != 10 {
		t.Errorf("expected 10, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_ContextTimeout(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var errHandled error
	var mu sync.Mutex
	done := make(chan struct{})

	publisher := NewAsyncPublisher(bus,
		WithErrorHandler(func(err error, e ApplicationEvent) {
			mu.Lock()
			errHandled = err
			mu.Unlock()
			close(done)
		}),
	)
	defer publisher.Close()

	// 创建一个已经取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	publisher.Publish(ctx, &BaseEvent{EventType: "TimeoutEvent"})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for error handler")
	}

	mu.Lock()
	defer mu.Unlock()
	if errHandled == nil {
		t.Error("expected context canceled error")
	}
}

func TestAsyncPublisher_PanicRecovery(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var mu sync.Mutex
	panicHandled := false
	done := make(chan struct{})

	bus.Subscribe("PanicEvent", func(e ApplicationEvent) {
		panic("test panic")
	})

	publisher := NewAsyncPublisher(bus,
		WithErrorHandler(func(err error, e ApplicationEvent) {
			mu.Lock()
			panicHandled = true
			mu.Unlock()
			close(done)
		}),
	)
	defer publisher.Close()

	ctx := context.Background()
	publisher.Publish(ctx, &BaseEvent{EventType: "PanicEvent"})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for panic recovery")
	}

	mu.Lock()
	defer mu.Unlock()
	if !panicHandled {
		t.Error("expected panic to be handled")
	}
}

func TestAsyncPublisher_WorkerCount(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus, WithWorkerCount(5))
	defer publisher.Close()

	var mu sync.Mutex
	received := 0
	done := make(chan struct{})

	bus.Subscribe("WorkerEvent", func(e ApplicationEvent) {
		mu.Lock()
		received++
		if received == 20 {
			mu.Unlock()
			close(done)
			return
		}
		mu.Unlock()
	})

	ctx := context.Background()
	for range 20 {
		publisher.Publish(ctx, &BaseEvent{EventType: "WorkerEvent"})
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for events, received %d", received)
	}

	mu.Lock()
	if received != 20 {
		t.Errorf("expected 20, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_Close(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus)

	var mu sync.Mutex
	received := 0
	done := make(chan struct{})

	bus.Subscribe("CloseEvent", func(e ApplicationEvent) {
		mu.Lock()
		received++
		if received == 5 {
			mu.Unlock()
			close(done)
			return
		}
		mu.Unlock()
	})

	ctx := context.Background()
	for range 5 {
		publisher.Publish(ctx, &BaseEvent{EventType: "CloseEvent"})
	}

	// 等待所有事件被处理
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for events, received %d", received)
	}

	// 关闭发布器
	publisher.Close()

	mu.Lock()
	if received != 5 {
		t.Errorf("expected 5, got %d", received)
	}
	mu.Unlock()
}

func TestAsyncPublisher_NoErrorHandler(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	// 不设置错误处理器
	publisher := NewAsyncPublisher(bus)
	defer publisher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// 等待上下文超时，使用 channel 替代 time.Sleep
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(done)
	}()

	select {
	case <-done:
		// 上下文已取消
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for context cancellation")
	}

	// 应该不会 panic
	publisher.Publish(ctx, &BaseEvent{EventType: "NoHandlerEvent"})
}

// TestAsyncPublisher_WorkerCountGreaterThanBuffer 测试 workerCount > channel buffer 时的行为
func TestAsyncPublisher_WorkerCountGreaterThanBuffer(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	// workerCount=5, queueSize=2，worker 数量大于缓冲
	publisher := NewAsyncPublisher(bus,
		WithWorkerCount(5),
		WithWorkerQueueSize(2),
	)
	defer publisher.Close()

	var mu sync.Mutex
	received := 0
	done := make(chan struct{})

	bus.Subscribe("BufferEvent", func(e ApplicationEvent) {
		mu.Lock()
		received++
		if received == 10 {
			mu.Unlock()
			close(done)
			return
		}
		mu.Unlock()
	})

	ctx := context.Background()
	// 发送超过 buffer 的事件
	for range 10 {
		publisher.Publish(ctx, &BaseEvent{EventType: "BufferEvent"})
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for events, received %d", received)
	}

	mu.Lock()
	if received != 10 {
		t.Errorf("expected 10, got %d", received)
	}
	mu.Unlock()
}

// TestAsyncPublisher_IndependentQueueSize 测试 workerCount 和 queueSize 独立配置
func TestAsyncPublisher_IndependentQueueSize(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	publisher := NewAsyncPublisher(bus,
		WithWorkerCount(2),
		WithWorkerQueueSize(100),
	)
	defer publisher.Close()

	if publisher.workerCount != 2 {
		t.Errorf("expected workerCount=2, got %d", publisher.workerCount)
	}
	if publisher.queueSize != 100 {
		t.Errorf("expected queueSize=100, got %d", publisher.queueSize)
	}
}
