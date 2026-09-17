package event

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestDeadLetterQueue_GetByType(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()
	now := time.Now()

	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "type1", EventTime: now}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "type2", EventTime: now.Add(time.Millisecond)}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "type1", EventTime: now.Add(2 * time.Millisecond)}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})

	// 获取 type1 的事件
	type1Events := dlq.GetByType("type1")
	if len(type1Events) != 2 {
		t.Errorf("expected 2 type1 events, got %d", len(type1Events))
	}

	// 获取 type2 的事件
	type2Events := dlq.GetByType("type2")
	if len(type2Events) != 1 {
		t.Errorf("expected 1 type2 event, got %d", len(type2Events))
	}

	// 获取不存在的事件类型
	type3Events := dlq.GetByType("type3")
	if len(type3Events) != 0 {
		t.Errorf("expected 0 type3 events, got %d", len(type3Events))
	}

	// 验证返回的是快照
	type1Events[0].Event = &BaseEvent{EventType: "modified"}
	if dlq.GetByType("type1")[0].Event.Type() == "modified" {
		t.Error("expected GetByType() to return a copy")
	}
}

func TestDeadLetterQueue_RemoveByType(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()
	now := time.Now()

	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "type1", EventTime: now}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "type2", EventTime: now.Add(time.Millisecond)}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "type1", EventTime: now.Add(2 * time.Millisecond)}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})

	// 移除 type1 的事件
	removed := dlq.RemoveByType("type1")
	if removed != 2 {
		t.Errorf("expected to remove 2 type1 events, got %d", removed)
	}

	// 验证剩余事件数量
	if dlq.Size() != 1 {
		t.Errorf("expected size 1 after removal, got %d", dlq.Size())
	}

	// 验证剩余的是 type2 事件
	events := dlq.Events()
	if len(events) != 1 || events[0].Event.Type() != "type2" {
		t.Errorf("expected remaining event to be type2")
	}

	// 移除不存在的事件类型
	removed = dlq.RemoveByType("type3")
	if removed != 0 {
		t.Errorf("expected to remove 0 type3 events, got %d", removed)
	}
}

func TestDeadLetterQueue_Stats(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()
	now := time.Now()

	// 添加可重试事件
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "retryable", EventTime: now}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "retryable", EventTime: now.Add(time.Millisecond)}, RetryCount: 1, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})

	// 添加未到重试时间的事件
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "waiting", EventTime: now.Add(2 * time.Millisecond)}, RetryCount: 1, MaxRetries: 3, NextRetryAt: time.Now().Add(time.Hour)})

	// 注意：已耗尽的事件（RetryCount >= MaxRetries）不会加入队列，而是触发永久失败回调
	// 因此这里不添加耗尽事件

	stats := dlq.Stats()

	if stats.Total != 3 {
		t.Errorf("expected total 3, got %d", stats.Total)
	}

	if stats.Retryable != 2 {
		t.Errorf("expected retryable 2, got %d", stats.Retryable)
	}

	if stats.Exhausted != 0 {
		t.Errorf("expected exhausted 0, got %d", stats.Exhausted)
	}

	if len(stats.EventTypeCount) != 2 {
		t.Errorf("expected 2 event types, got %d", len(stats.EventTypeCount))
	}

	if stats.EventTypeCount["retryable"] != 2 {
		t.Errorf("expected 2 retryable events, got %d", stats.EventTypeCount["retryable"])
	}

	if stats.EventTypeCount["waiting"] != 1 {
		t.Errorf("expected 1 waiting event, got %d", stats.EventTypeCount["waiting"])
	}
}

func TestDeadLetterQueue_Stats_Empty(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()

	stats := dlq.Stats()

	if stats.Total != 0 {
		t.Errorf("expected total 0, got %d", stats.Total)
	}

	if stats.Retryable != 0 {
		t.Errorf("expected retryable 0, got %d", stats.Retryable)
	}

	if stats.Exhausted != 0 {
		t.Errorf("expected exhausted 0, got %d", stats.Exhausted)
	}

	if len(stats.EventTypeCount) != 0 {
		t.Errorf("expected 0 event types, got %d", len(stats.EventTypeCount))
	}
}

func TestRetryPolicy_DefaultAndNoRetry(t *testing.T) {
	t.Parallel()
	defaultPolicy := DefaultRetryPolicy()
	if defaultPolicy.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", defaultPolicy.MaxRetries)
	}

	noRetry := NoRetryPolicy()
	if noRetry.MaxRetries != 0 {
		t.Errorf("expected no retry max retries 0, got %d", noRetry.MaxRetries)
	}
}

func TestFailedEvent_Error(t *testing.T) {
	t.Parallel()
	testErr := errors.New("test error")
	fe := FailedEvent{
		Event:      &BaseEvent{EventType: "test"},
		Err:        testErr,
		RetryCount: 1,
		MaxRetries: 3,
	}

	if fe.Err.Error() != testErr.Error() {
		t.Errorf("expected error %v, got %v", testErr, fe.Err)
	}
}

func TestRetryDeadLetter_EmptyQueue(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background())

	retried := bus.RetryDeadLetter()
	if retried {
		t.Error("expected retry to return false for empty queue")
	}
}

func TestEventBusWithDeadLetter_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background(),
		WithMaxRetries(0),
	)

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			bus.PublishWithRecovery(&BaseEvent{EventType: "concurrent.event"})
		}(i)
	}
	wg.Wait()

	// 不应该 panic
}
