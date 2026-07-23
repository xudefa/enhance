package event

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventBusWithDeadLetter 支持死信队列的事件总线
//
// 在 EventBusWithOrdering 基础上增加：
//   - 监听器 panic 捕获
//   - 自动重试机制
//   - 死信队列
//
// 使用示例：
//
//	bus := event.NewEventBusWithDeadLetter(
//	    event.WithMaxRetries(3),
//	    event.WithBackoff(event.BackoffExponential, time.Second),
//	)
//
//	bus.Subscribe("MyEvent", func(e event.ApplicationEvent) {
//	    // 可能失败的处理逻辑
//	})
//
//	bus.Publish(&event.BaseEvent{EventType: "MyEvent"})
type EventBusWithDeadLetter struct {
	*EventBusWithOrdering
	dlq         *DeadLetterQueue
	retryPolicy RetryPolicy
	mu          sync.Mutex
	retrying    map[string]bool // 防止重试递归
	ctx         context.Context // 上下文，用于取消异步重试 goroutine
	cancel      context.CancelFunc
}

// DeadLetterOption 死信队列配置选项
type DeadLetterOption func(*EventBusWithDeadLetter)

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(n int) DeadLetterOption {
	return func(b *EventBusWithDeadLetter) {
		b.retryPolicy.MaxRetries = n
	}
}

// WithBackoff 设置退避策略
func WithBackoff(strategy BackoffStrategy, initialDelay time.Duration) DeadLetterOption {
	return func(b *EventBusWithDeadLetter) {
		b.retryPolicy.Strategy = strategy
		b.retryPolicy.InitialDelay = initialDelay
	}
}

// WithRetryPolicy 设置完整重试策略
func WithRetryPolicy(policy RetryPolicy) DeadLetterOption {
	return func(b *EventBusWithDeadLetter) {
		b.retryPolicy = policy
	}
}

// WithDeadLetterHandler 设置永久失败处理器
func WithDeadLetterHandler(handler func(FailedEvent)) DeadLetterOption {
	return func(b *EventBusWithDeadLetter) {
		b.dlq.SetPermanentFailureHandler(handler)
	}
}

// NewEventBusWithDeadLetter 创建支持死信队列的事件总线
func NewEventBusWithDeadLetter(opts ...DeadLetterOption) *EventBusWithDeadLetter {
	ctx, cancel := context.WithCancel(context.Background())
	bus := &EventBusWithDeadLetter{
		EventBusWithOrdering: NewEventBusWithOrdering(),
		dlq:                  NewDeadLetterQueue(),
		retryPolicy:          DefaultRetryPolicy(),
		retrying:             make(map[string]bool),
		ctx:                  ctx,
		cancel:               cancel,
	}

	for _, opt := range opts {
		opt(bus)
	}

	return bus
}

// DeadLetterQueue 返回死信队列实例
func (b *EventBusWithDeadLetter) DeadLetterQueue() *DeadLetterQueue {
	return b.dlq
}

// RetryPolicy 返回重试策略
func (b *EventBusWithDeadLetter) RetryPolicy() RetryPolicy {
	return b.retryPolicy
}

// PublishWithRecovery 发布事件并捕获错误，失败时进入死信队列
func (b *EventBusWithDeadLetter) PublishWithRecovery(event ApplicationEvent) {
	key := fmt.Sprintf("%s-%d", event.Type(), event.Timestamp().UnixNano())

	b.mu.Lock()
	if b.retrying[key] {
		b.mu.Unlock()
		return
	}
	b.retrying[key] = true
	b.mu.Unlock()

	b.publishWithRecoveryInternal(event, 0, key)
}

// publishResult 携带 goroutine 执行结果，避免共享变量竞争
type publishResult struct {
	err error
}

func (b *EventBusWithDeadLetter) publishWithRecoveryInternal(event ApplicationEvent, attempt int, key string) {
	resultCh := make(chan publishResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					resultCh <- publishResult{err: err}
					return
				}
				resultCh <- publishResult{err: fmt.Errorf("event handler panic: %v", r)}
				return
			}
			resultCh <- publishResult{}
		}()

		b.EventBusWithOrdering.Publish(event)
	}()

	result := <-resultCh

	b.mu.Lock()
	delete(b.retrying, key)
	b.mu.Unlock()

	if result.err != nil {
		b.handleFailure(event, result.err, attempt, key)
	}
}

func (b *EventBusWithDeadLetter) handleFailure(event ApplicationEvent, err error, attempt int, key string) {
	fe := FailedEvent{
		Event:        event,
		Err:          err,
		RetryCount:   attempt,
		MaxRetries:   b.retryPolicy.MaxRetries,
		LastFailedAt: time.Now(),
	}

	// 首次失败时记录 FirstFailedAt
	if attempt == 0 {
		fe.FirstFailedAt = time.Now()
	} else {
		// 从死信队列中查找原始事件的首次失败时间
		b.dlq.events.Range(func(k, value any) bool {
			existing := value.(FailedEvent)
			if existing.Event.Type() == event.Type() && existing.Event.Timestamp().Equal(event.Timestamp()) {
				fe.FirstFailedAt = existing.FirstFailedAt
				return false
			}
			return true
		})
		// 如果未找到（理论上不应发生），使用当前时间
		if fe.FirstFailedAt.IsZero() {
			fe.FirstFailedAt = time.Now()
		}
	}

	if attempt < b.retryPolicy.MaxRetries {
		delay := b.retryPolicy.CalculateDelay(attempt)
		fe.NextRetryAt = time.Now().Add(delay)

		// 异步重试，使用 context 控制取消
		go func() {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-timer.C:
				b.dlq.Remove(event)
				b.publishWithRecoveryInternal(event, attempt+1, key)
			case <-b.ctx.Done():
				// 调度器已关闭，取消重试
				return
			}
		}()
	}

	b.dlq.Add(fe)
}

// Publish 覆盖原有 Publish 方法，使用带恢复的发布
func (b *EventBusWithDeadLetter) Publish(event ApplicationEvent) {
	b.PublishWithRecovery(event)
}

// RetryDeadLetter 手动重试死信队列中的下一个事件
func (b *EventBusWithDeadLetter) RetryDeadLetter() bool {
	fe, ok := b.dlq.Peek()
	if !ok {
		return false
	}

	b.dlq.Remove(fe.Event)
	key := fmt.Sprintf("%s-%d", fe.Event.Type(), fe.Event.Timestamp().UnixNano())
	b.publishWithRecoveryInternal(fe.Event, fe.RetryCount, key)
	return true
}

// RetryAllDeadLetters 重试所有可重试的死信事件
func (b *EventBusWithDeadLetter) RetryAllDeadLetters() int {
	// 收集可重试事件并删除
	retryable := make([]FailedEvent, 0)
	b.dlq.events.Range(func(key, value any) bool {
		fe := value.(FailedEvent)
		if fe.ShouldRetry() {
			retryable = append(retryable, fe)
			b.dlq.events.Delete(key)
		}
		return true
	})

	// 重试（在锁外执行，避免长时间持有锁）
	for _, fe := range retryable {
		key := fmt.Sprintf("%s-%d", fe.Event.Type(), fe.Event.Timestamp().UnixNano())
		b.publishWithRecoveryInternal(fe.Event, fe.RetryCount, key)
	}

	return len(retryable)
}

// Close 关闭事件总线，取消所有待处理的异步重试。
//
// 调用后，所有正在进行的异步重试将收到 ctx.Done() 信号并终止。
// 死信队列中的数据不会被清除，可通过 RetryDeadLetter 或 RetryAllDeadLetters 手动处理。
//
// 注意：Close 应在应用关闭时调用，确保 goroutine 正确退出。
func (b *EventBusWithDeadLetter) Close() {
	if b.cancel != nil {
		b.cancel()
	}
}
