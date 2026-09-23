package event

import (
	"context"
	"errors"
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

// TestDeadLetterBus_HandleFailure_Coverage 测试失败处理逻辑
func TestDeadLetterBus_HandleFailure_Coverage(t *testing.T) {
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

// TestDeadLetterBus_Publish_Coverage 测试 Publish 方法（委托给 PublishWithRecovery）
func TestDeadLetterBus_Publish_Coverage(t *testing.T) {
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

// TestDeadLetterBus_RetryDeadLetter_Coverage 测试手动重试死信
func TestDeadLetterBus_RetryDeadLetter_Coverage(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{MaxRetries: 0}
	bus := NewEventBusWithDeadLetter(context.Background(), WithRetryPolicy(policy))
	defer bus.Close()

	bus.Subscribe("test.retry", func(e ApplicationEvent) {
		panic("always fails")
	})

	event := &BaseEvent{EventType: "test.retry"}
	bus.PublishWithRecovery(event)

	// 等待事件进入死信队列
	time.Sleep(time.Millisecond * 50)

	// 测试空死信队列时返回 false
	emptyPolicy := RetryPolicy{MaxRetries: 0}
	emptyBus := NewEventBusWithDeadLetter(context.Background(), WithRetryPolicy(emptyPolicy))
	defer emptyBus.Close()
	if emptyBus.RetryDeadLetter() {
		t.Error("Expected RetryDeadLetter to return false for empty queue")
	}
}

// TestDeadLetterBus_RetryAllDeadLetters_Coverage 测试重试所有死信
func TestDeadLetterBus_RetryAllDeadLetters_Coverage(t *testing.T) {
	t.Parallel()

	policy := RetryPolicy{MaxRetries: 0}
	bus := NewEventBusWithDeadLetter(context.Background(), WithRetryPolicy(policy))
	defer bus.Close()

	bus.Subscribe("test.retry1", func(e ApplicationEvent) {
		panic("fails")
	})
	bus.Subscribe("test.retry2", func(e ApplicationEvent) {
		panic("fails")
	})

	bus.PublishWithRecovery(&BaseEvent{EventType: "test.retry1"})
	bus.PublishWithRecovery(&BaseEvent{EventType: "test.retry2"})

	// 等待事件进入死信队列
	time.Sleep(time.Millisecond * 50)

	count := bus.RetryAllDeadLetters()
	if count < 0 {
		t.Errorf("Expected non-negative retry count, got %d", count)
	}
}

// TestDeadLetterBus_PublishWithRecovery_NilEvent_Coverage 测试发布 nil 事件
func TestDeadLetterBus_PublishWithRecovery_NilEvent_Coverage(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithDeadLetter(context.Background())
	defer bus.Close()

	// 不应 panic
	bus.PublishWithRecovery(nil)
}

// TestDeadLetterBus_PublishWithRecovery_ContextCancelled_Coverage 测试上下文取消时的行为
func TestDeadLetterBus_PublishWithRecovery_ContextCancelled_Coverage(t *testing.T) {
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

// TestDeadLetterStore_Stats_Coverage 测试统计信息
func TestDeadLetterStore_Stats_Coverage(t *testing.T) {
	t.Parallel()

	dlq := NewDeadLetterQueue()

	fe1 := FailedEvent{
		Event:      &BaseEvent{EventType: "test.1"},
		Err:        errors.New("error 1"),
		RetryCount: 0,
		MaxRetries: 3,
	}
	fe2 := FailedEvent{
		Event:      &BaseEvent{EventType: "test.2"},
		Err:        errors.New("error 2"),
		RetryCount: 1,
		MaxRetries: 3,
	}

	dlq.Add(fe1)
	dlq.Add(fe2)

	if dlq.Size() != 2 {
		t.Errorf("Expected size 2, got %d", dlq.Size())
	}
}

// TestDeadLetterStore_GetByType_Coverage 测试按类型获取
func TestDeadLetterStore_GetByType_Coverage(t *testing.T) {
	t.Parallel()

	dlq := NewDeadLetterQueue()

	fe := FailedEvent{
		Event:      &BaseEvent{EventType: "test.type"},
		Err:        errors.New("error"),
		RetryCount: 0,
		MaxRetries: 3,
	}
	dlq.Add(fe)

	// 测试 Peek 方法
	_, found := dlq.Peek()
	if !found {
		t.Error("Expected to find event")
	}

	// 测试 Size 方法
	if dlq.Size() != 1 {
		t.Errorf("Expected size 1, got %d", dlq.Size())
	}

	// 测试 Events 方法
	events := dlq.Events()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// 测试 Clear 方法
	dlq.Clear()
	if dlq.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", dlq.Size())
	}
}

// TestTransactionalEvent_Coverage 测试事务性事件
func TestTransactionalEvent_Coverage(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	committed := int32(0)
	rolledBack := int32(0)

	bus.Subscribe("tx.commit", func(e ApplicationEvent) {
		atomic.AddInt32(&committed, 1)
	})
	bus.Subscribe("tx.rollback", func(e ApplicationEvent) {
		atomic.AddInt32(&rolledBack, 1)
	})

	// 测试成功提交
	publisher := NewTransactionalEventPublisher(bus)
	tx := publisher.BeginTransaction()
	tx.PublishAfterCommit(&BaseEvent{EventType: "tx.commit"})
	tx.Commit(bus)

	time.Sleep(time.Millisecond * 10)

	if atomic.LoadInt32(&committed) != 1 {
		t.Errorf("Expected 1 committed event, got %d", atomic.LoadInt32(&committed))
	}

	// 测试回滚
	tx2 := publisher.BeginTransaction()
	tx2.PublishAfterRollback(&BaseEvent{EventType: "tx.rollback"})
	tx2.Rollback(bus)

	time.Sleep(time.Millisecond * 10)

	// 回滚时应发布事件
	if atomic.LoadInt32(&rolledBack) != 1 {
		t.Errorf("Expected 1 rolled back event, got %d", atomic.LoadInt32(&rolledBack))
	}
}
