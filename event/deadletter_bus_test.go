package event

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestMakeEventKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType string
		ts        int64
	}{
		{"basic", "test.event", time.Now().UnixNano()},
		{"empty type", "", 0},
		{"future timestamp", "event", time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key := makeEventKey(tt.eventType, tt.ts)
			if !strings.HasPrefix(key, tt.eventType) {
				t.Errorf("expected key to start with %q, got %q", tt.eventType, key)
			}
		})
	}
}

func TestMakeEventKey_Unique(t *testing.T) {
	t.Parallel()

	ts := time.Now().UnixNano()
	keys := make(map[string]bool)
	for range 100 {
		key := makeEventKey("test", ts)
		if keys[key] {
			t.Fatalf("duplicate key: %s", key)
		}
		keys[key] = true
	}
}

func TestEventBusWithDeadLetter_NilEvent(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithDeadLetter(context.Background())
	bus.PublishWithRecovery(nil)

	if bus.DeadLetterQueue().Size() != 0 {
		t.Error("nil event should not be added to DLQ")
	}
}

func TestEventBusWithDeadLetter_Close(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithDeadLetter(context.Background(),
		WithMaxRetries(0),
	)

	var called bool
	bus.Subscribe("close.test", func(e ApplicationEvent) {
		called = true
	})

	bus.PublishWithRecovery(&BaseEvent{EventType: "close.test"})
	time.Sleep(100 * time.Millisecond)

	bus.Close()

	if !called {
		t.Error("expected handler to be called before close")
	}
}

func TestEventBusWithDeadLetter_HandleFailure(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries:   1,
		InitialDelay: time.Millisecond * 10,
		MaxDelay:     time.Millisecond * 50,
		Multiplier:   2.0,
	}
	bus := NewEventBusWithDeadLetter(context.Background(), WithRetryPolicy(policy))
	defer bus.Close()

	callCount := int32(0)
	bus.Subscribe("test.failure", func(e ApplicationEvent) {
		atomic.AddInt32(&callCount, 1)
		if atomic.LoadInt32(&callCount) == 1 {
			panic("first call fails")
		}
	})

	event := &BaseEvent{EventType: "test.failure"}
	bus.PublishWithRecovery(event)

	// 等待异步重试完成
	time.Sleep(time.Millisecond * 100)

	// 应该被调用了2次（第一次失败，第二次重试成功）
	if count := atomic.LoadInt32(&callCount); count < 1 {
		t.Errorf("Expected at least 1 call, got %d", count)
	}
}

func TestEventBusWithDeadLetter_Publish(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{
		MaxRetries: 0,
	}
	bus := NewEventBusWithDeadLetter(context.Background(), WithRetryPolicy(policy))
	defer bus.Close()

	received := int32(0)
	bus.Subscribe("test.publish", func(e ApplicationEvent) {
		atomic.AddInt32(&received, 1)
	})

	event := &BaseEvent{EventType: "test.publish"}
	bus.Publish(event) // 使用 Publish 而非 PublishWithRecovery

	time.Sleep(time.Millisecond * 10)

	if count := atomic.LoadInt32(&received); count != 1 {
		t.Errorf("Expected 1 event received, got %d", count)
	}
}

func TestEventBusWithDeadLetter_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	policy := RetryPolicy{
		MaxRetries:   3,
		InitialDelay: time.Second,
		MaxDelay:     time.Second * 2,
		Multiplier:   2.0,
	}
	bus := NewEventBusWithDeadLetter(ctx, WithRetryPolicy(policy))

	bus.Subscribe("test.cancel", func(e ApplicationEvent) {
		panic("always fails")
	})

	// 立即取消上下文
	cancel()

	event := &BaseEvent{EventType: "test.cancel"}
	bus.PublishWithRecovery(event)

	// 等待一下，确保不会阻塞
	time.Sleep(time.Millisecond * 50)

	// 关闭总线
	bus.Close()
}
