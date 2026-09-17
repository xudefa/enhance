package event

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryPolicy_CalculateDelay(t *testing.T) {
	t.Parallel()
	tests := testRetryPolicyCalculateDelayCases()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

type retryPolicyDelayCase struct {
	name     string
	policy   RetryPolicy
	attempt  int
	minDelay time.Duration
	maxDelay time.Duration
}

func testRetryPolicyCalculateDelayCases() []retryPolicyDelayCase {
	return []retryPolicyDelayCase{
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
	retried := bus.RetryDeadLetter()
	if !retried {
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
