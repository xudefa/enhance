package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// eventKeyCounter 全局原子计数器，确保事件 key 唯一，避免同纳秒碰撞
var eventKeyCounter atomic.Uint64

// makeEventKey 生成唯一的事件 key：类型-时间戳-计数器
func makeEventKey(eventType string, ts int64) string {
	return fmt.Sprintf("%s-%d-%d", eventType, ts, eventKeyCounter.Add(1))
}

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
//	    context.Background(),
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
	wg          sync.WaitGroup // 跟踪所有异步 goroutine
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
//
// 参数:
//   - ctx: 父级 context，用于控制异步重试生命周期
//   - opts: 配置选项
func NewEventBusWithDeadLetter(ctx context.Context, opts ...DeadLetterOption) *EventBusWithDeadLetter {
	ctx, cancel := context.WithCancel(ctx)
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
	if event == nil {
		return
	}
	key := makeEventKey(event.Type(), event.Timestamp().UnixNano())

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

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
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

	// 带超时等待结果，防止监听器死锁或无限阻塞导致 goroutine 泄漏
	var result publishResult
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case result = <-resultCh:
	case <-timer.C:
		result = publishResult{err: fmt.Errorf("event handler timeout after 30s")}
	case <-b.ctx.Done():
		b.mu.Lock()
		delete(b.retrying, key)
		b.mu.Unlock()
		return
	}

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
			existing, ok := value.(FailedEvent)
			if ok && existing.Event != nil && existing.Event.Type() == event.Type() && existing.Event.Timestamp().Equal(event.Timestamp()) {
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
	}

	b.dlq.Add(fe)

	if attempt < b.retryPolicy.MaxRetries {
		delay := time.Until(fe.NextRetryAt)

		// 异步重试，使用 context 控制取消
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
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
	key := makeEventKey(fe.Event.Type(), fe.Event.Timestamp().UnixNano())
	b.publishWithRecoveryInternal(fe.Event, fe.RetryCount, key)
	return true
}

// RetryAllDeadLetters 重试所有可重试的死信事件
func (b *EventBusWithDeadLetter) RetryAllDeadLetters() int {
	// 收集可重试事件并删除
	retryable := make([]FailedEvent, 0)
	b.dlq.events.Range(func(key, value any) bool {
		fe, ok := value.(FailedEvent)
		if ok && fe.Event != nil && fe.ShouldRetry() {
			retryable = append(retryable, fe)
			b.dlq.events.Delete(key)
		}
		return true
	})

	// 重试（在锁外执行，避免长时间持有锁）
	for _, fe := range retryable {
		key := makeEventKey(fe.Event.Type(), fe.Event.Timestamp().UnixNano())
		b.publishWithRecoveryInternal(fe.Event, fe.RetryCount, key)
	}

	return len(retryable)
}

// Close 关闭事件总线，取消所有待处理的异步重试并等待 goroutine 退出。
//
// 调用后，所有正在进行的异步重试将收到 ctx.Done() 信号并终止。
// 此方法会阻塞直到所有异步 goroutine 完成（包括重试 goroutine 和异步事件处理器）。
// 死信队列中的数据不会被清除，可通过 RetryDeadLetter 或 RetryAllDeadLetters 手动处理。
func (b *EventBusWithDeadLetter) Close() {
	// 先取消 context，释放阻塞在 ctx.Done() 上的异步事件处理器和重试 goroutine，
	// 否则 WaitAsync/Wait 会因处理器永远阻塞而陷入死锁。
	if b.cancel != nil {
		b.cancel()
	}
	// 等待异步事件处理器完成
	b.WaitAsync()
	// 等待重试 goroutine 退出
	b.wg.Wait()
}
