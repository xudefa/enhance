package event

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryPolicy_CalculateDelay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   RetryPolicy
		attempt  int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{
			name:     "no backoff",
			policy:   RetryPolicy{MaxRetries: 3, Strategy: BackoffNone},
			attempt:  0,
			minDelay: 0,
			maxDelay: 0,
		},
		{
			name:     "fixed backoff",
			policy:   RetryPolicy{MaxRetries: 3, Strategy: BackoffFixed, InitialDelay: 100 * time.Millisecond},
			attempt:  2,
			minDelay: 100 * time.Millisecond,
			maxDelay: 100 * time.Millisecond,
		},
		{
			name:     "linear backoff",
			policy:   RetryPolicy{MaxRetries: 3, Strategy: BackoffLinear, InitialDelay: 100 * time.Millisecond},
			attempt:  2,
			minDelay: 300 * time.Millisecond,
			maxDelay: 300 * time.Millisecond,
		},
		{
			name:     "exponential backoff",
			policy:   RetryPolicy{MaxRetries: 3, Strategy: BackoffExponential, InitialDelay: 100 * time.Millisecond, Multiplier: 2.0},
			attempt:  2,
			minDelay: 400 * time.Millisecond,
			maxDelay: 10 * time.Second,
		},
		{
			name:     "exponential with max delay cap",
			policy:   RetryPolicy{MaxRetries: 10, Strategy: BackoffExponential, InitialDelay: 1 * time.Second, MaxDelay: 5 * time.Second, Multiplier: 2.0},
			attempt:  5,
			minDelay: 0,
			maxDelay: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := tt.policy.CalculateDelay(tt.attempt)
			if delay < tt.minDelay {
				t.Errorf("expected delay >= %v, got %v", tt.minDelay, delay)
			}
			if tt.maxDelay > 0 && delay > tt.maxDelay {
				t.Errorf("expected delay <= %v, got %v", tt.maxDelay, delay)
			}
		})
	}
}

func TestFailedEvent_IsExhausted(t *testing.T) {
	t.Parallel()
	fe := FailedEvent{RetryCount: 2, MaxRetries: 3}
	if fe.IsExhausted() {
		t.Error("expected not exhausted at 2/3")
	}

	fe.RetryCount = 3
	if !fe.IsExhausted() {
		t.Error("expected exhausted at 3/3")
	}
}

func TestFailedEvent_ShouldRetry(t *testing.T) {
	t.Parallel()
	fe := FailedEvent{
		RetryCount:  2,
		MaxRetries:  3,
		NextRetryAt: time.Now().Add(-time.Second),
	}
	if !fe.ShouldRetry() {
		t.Error("expected should retry")
	}

	fe.RetryCount = 3
	if fe.ShouldRetry() {
		t.Error("expected should NOT retry when exhausted")
	}

	fe.RetryCount = 2
	fe.NextRetryAt = time.Now().Add(time.Second)
	if fe.ShouldRetry() {
		t.Error("expected should NOT retry before next retry time")
	}
}

func TestDeadLetterQueue_AddAndPeek(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()

	fe := FailedEvent{
		Event:       &BaseEvent{EventType: "test"},
		RetryCount:  0,
		MaxRetries:  3,
		NextRetryAt: time.Now().Add(-time.Second),
	}

	dlq.Add(fe)

	if dlq.Size() != 1 {
		t.Errorf("expected size 1, got %d", dlq.Size())
	}

	peeked, ok := dlq.Peek()
	if !ok {
		t.Fatal("expected peek to return event")
	}

	if peeked.Event.Type() != "test" {
		t.Errorf("expected event type 'test', got %s", peeked.Event.Type())
	}
}

func TestDeadLetterQueue_PermanentFailure(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()
	var permanentCalled int32

	dlq.SetPermanentFailureHandler(func(fe FailedEvent) {
		atomic.AddInt32(&permanentCalled, 1)
	})

	fe := FailedEvent{
		Event:      &BaseEvent{EventType: "test"},
		RetryCount: 3,
		MaxRetries: 3,
	}

	dlq.Add(fe)

	if atomic.LoadInt32(&permanentCalled) != 1 {
		t.Error("expected permanent failure handler to be called")
	}

	if dlq.Size() != 1 {
		t.Error("exhausted event should be added to queue")
	}
}

func TestDeadLetterQueue_Remove(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()
	now := time.Now()

	fe := FailedEvent{
		Event:       &BaseEvent{EventType: "test", EventTime: now},
		RetryCount:  0,
		MaxRetries:  3,
		NextRetryAt: time.Now().Add(-time.Second),
	}

	dlq.Add(fe)
	dlq.Remove(fe.Event)

	if dlq.Size() != 0 {
		t.Errorf("expected size 0 after remove, got %d", dlq.Size())
	}
}

func TestDeadLetterQueue_Clear(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()

	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "test1"}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "test2"}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})

	dlq.Clear()

	if dlq.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", dlq.Size())
	}
}

func TestDeadLetterQueue_Events(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()

	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "test1"}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})
	dlq.Add(FailedEvent{Event: &BaseEvent{EventType: "test2"}, RetryCount: 0, MaxRetries: 3, NextRetryAt: time.Now().Add(-time.Second)})

	events := dlq.Events()

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// 验证返回的是快照，修改不影响内部状态
	events[0].Event = &BaseEvent{EventType: "modified"}
	if dlq.Events()[0].Event.Type() == "modified" {
		t.Error("expected Events() to return a copy")
	}
}

func TestEventBusWithDeadLetter_DefaultOptions(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background())

	if bus.retryPolicy.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", bus.retryPolicy.MaxRetries)
	}

	if bus.dlq == nil {
		t.Error("expected non-nil dead letter queue")
	}
}

func TestEventBusWithDeadLetter_CustomOptions(t *testing.T) {
	t.Parallel()
	var permanentCalled int32

	bus := NewEventBusWithDeadLetter(context.Background(),
		WithMaxRetries(5),
		WithBackoff(BackoffFixed, 200*time.Millisecond),
		WithDeadLetterHandler(func(fe FailedEvent) {
			atomic.AddInt32(&permanentCalled, 1)
		}),
	)

	if bus.retryPolicy.MaxRetries != 5 {
		t.Errorf("expected max retries 5, got %d", bus.retryPolicy.MaxRetries)
	}

	if bus.retryPolicy.Strategy != BackoffFixed {
		t.Errorf("expected backoff strategy fixed, got %s", bus.retryPolicy.Strategy)
	}

	if bus.retryPolicy.InitialDelay != 200*time.Millisecond {
		t.Errorf("expected initial delay 200ms, got %v", bus.retryPolicy.InitialDelay)
	}
}

func TestEventBusWithDeadLetter_FullRetryPolicy(t *testing.T) {
	t.Parallel()
	policy := RetryPolicy{
		MaxRetries:   5,
		Strategy:     BackoffExponential,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   1.5,
	}

	bus := NewEventBusWithDeadLetter(context.Background(), WithRetryPolicy(policy))

	if bus.retryPolicy != policy {
		t.Error("expected retry policy to match")
	}
}

func TestEventBusWithDeadLetter_PublishWithRecovery(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background(),
		WithMaxRetries(0), // 不重试，直接进入死信队列
	)

	var handlerCalled int32
	bus.Subscribe("test.event", func(e ApplicationEvent) {
		atomic.AddInt32(&handlerCalled, 1)
	})

	bus.PublishWithRecovery(&BaseEvent{EventType: "test.event"})

	if atomic.LoadInt32(&handlerCalled) != 1 {
		t.Error("expected handler to be called")
	}
}

func TestEventBusWithDeadLetter_DeadLetterQueueAccess(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background())

	dlq := bus.DeadLetterQueue()
	if dlq == nil {
		t.Error("expected non-nil dead letter queue")
	}

	policy := bus.RetryPolicy()
	if policy.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", policy.MaxRetries)
	}
}

func TestEventBusWithDeadLetter_RetryDeadLetter(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background(),
		WithMaxRetries(3),
		WithBackoff(BackoffNone, 0), // 无延迟
	)

	var handlerCalled int32
	var done chan struct{}

	bus.Subscribe("retry.event", func(e ApplicationEvent) {
		atomic.AddInt32(&handlerCalled, 1)
		if atomic.LoadInt32(&handlerCalled) == 1 {
			select {
			case <-done:
			default:
				close(done)
			}
		}
	})

	// 手动添加一个可重试的死信事件到队列
	fe := FailedEvent{
		Event:       &BaseEvent{EventType: "retry.event"},
		RetryCount:  0,
		MaxRetries:  3,
		NextRetryAt: time.Now().Add(-time.Second),
	}
	bus.dlq.Add(fe)

	done = make(chan struct{})
	result := bus.RetryDeadLetter()
	if !result {
		t.Error("expected retry to succeed")
	}

	// 使用 channel 等待异步重试完成
	select {
	case <-done:
		// 事件已处理
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for retry")
	}

	if atomic.LoadInt32(&handlerCalled) < 1 {
		t.Errorf("expected handler to be called at least once, got %d", handlerCalled)
	}
}

func TestEventBusWithDeadLetter_RetryAllDeadLetters(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background(),
		WithMaxRetries(3),
		WithBackoff(BackoffNone, 0),
	)

	var handlerCalled int32
	bus.Subscribe("retry.event", func(e ApplicationEvent) {
		atomic.AddInt32(&handlerCalled, 1)
	})

	// 添加多个可重试的死信事件
	for i := range 3 {
		fe := FailedEvent{
			Event:       &BaseEvent{EventType: "retry.event", EventTime: time.Now().Add(time.Duration(i) * time.Millisecond)},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(-time.Second),
		}
		bus.dlq.Add(fe)
	}

	count := bus.RetryAllDeadLetters()
	if count != 3 {
		t.Errorf("expected 3 retries, got %d", count)
	}
}

func TestRetryDeadLetter_EmptyQueue(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithDeadLetter(context.Background())

	result := bus.RetryDeadLetter()
	if result {
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
