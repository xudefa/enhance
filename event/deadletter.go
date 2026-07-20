package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// BackoffStrategy 退避策略类型
type BackoffStrategy string

const (
	BackoffNone        BackoffStrategy = "none"        // 无退避，立即重试
	BackoffFixed       BackoffStrategy = "fixed"       // 固定间隔退避
	BackoffExponential BackoffStrategy = "exponential" // 指数退避
	BackoffLinear      BackoffStrategy = "linear"      // 线性退避
)

// RetryPolicy 重试策略配置
type RetryPolicy struct {
	MaxRetries   int             // 最大重试次数，0 表示不重试
	Strategy     BackoffStrategy // 退避策略
	InitialDelay time.Duration   // 初始延迟
	MaxDelay     time.Duration   // 最大延迟（指数退避上限）
	Multiplier   float64         // 退避乘数（指数退避用）
}

// DefaultRetryPolicy 默认重试策略
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:   3,
		Strategy:     BackoffExponential,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
}

// NoRetryPolicy 不重试策略
func NoRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries: 0,
	}
}

// CalculateDelay 计算当前重试次数对应的延迟
func (p RetryPolicy) CalculateDelay(attempt int) time.Duration {
	if p.Strategy == BackoffNone || p.MaxRetries == 0 {
		return 0
	}

	var delay time.Duration
	switch p.Strategy {
	case BackoffFixed:
		delay = p.InitialDelay
	case BackoffLinear:
		delay = p.InitialDelay * time.Duration(attempt+1)
	case BackoffExponential:
		delay = p.InitialDelay * time.Duration(float64(int(1)<<uint(attempt))*p.Multiplier)
	default:
		delay = p.InitialDelay
	}

	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	return delay
}

// FailedEvent 失败事件记录
type FailedEvent struct {
	Event         ApplicationEvent // 原始事件
	Err           error            // 最后一次错误
	RetryCount    int              // 已重试次数
	MaxRetries    int              // 最大重试次数
	FirstFailedAt time.Time        // 首次失败时间
	LastFailedAt  time.Time        // 最后失败时间
	NextRetryAt   time.Time        // 下次重试时间
}

// IsExhausted 返回是否已达到最大重试次数
func (fe FailedEvent) IsExhausted() bool {
	return fe.RetryCount >= fe.MaxRetries
}

// ShouldRetry 返回是否应该重试
func (fe FailedEvent) ShouldRetry() bool {
	return fe.RetryCount < fe.MaxRetries && time.Now().After(fe.NextRetryAt)
}

// DeadLetterQueue 死信队列
//
// 存储处理失败的事件，支持重试和永久失败处理。
// 线程安全，使用 sync.Map 优化并发访问，atomic.Int64 跟踪大小。
type DeadLetterQueue struct {
	events             sync.Map // map[string]FailedEvent，key 为事件唯一标识
	size               atomic.Int64
	onPermanentFailure func(FailedEvent) // 永久失败回调
}

// NewDeadLetterQueue 创建死信队列
func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{}
}

// SetPermanentFailureHandler 设置永久失败处理器
func (dlq *DeadLetterQueue) SetPermanentFailureHandler(handler func(FailedEvent)) {
	dlq.onPermanentFailure = handler
}

// Add 添加失败事件到死信队列
func (dlq *DeadLetterQueue) Add(fe FailedEvent) {
	key := fmt.Sprintf("%s-%d", fe.Event.Type(), fe.Event.Timestamp().UnixNano())

	if fe.IsExhausted() {
		// 已达到最大重试次数，触发永久失败回调
		if dlq.onPermanentFailure != nil {
			dlq.onPermanentFailure(fe)
		}
		return
	}

	// 检查是否是新 key，避免重复计数
	if _, loaded := dlq.events.LoadOrStore(key, fe); !loaded {
		dlq.size.Add(1)
		return
	}
	// key 已存在，更新值
	dlq.events.Store(key, fe)
}

// Peek 获取下一个可重试的事件（不移除）
func (dlq *DeadLetterQueue) Peek() (FailedEvent, bool) {
	var found FailedEvent
	var ok bool

	dlq.events.Range(func(key, value any) bool {
		fe := value.(FailedEvent)
		if fe.ShouldRetry() {
			found = fe
			ok = true
			return false
		}
		return true
	})

	return found, ok
}

// Remove 移除指定事件（重试成功后调用）
func (dlq *DeadLetterQueue) Remove(event ApplicationEvent) {
	key := fmt.Sprintf("%s-%d", event.Type(), event.Timestamp().UnixNano())
	if _, loaded := dlq.events.LoadAndDelete(key); loaded {
		dlq.size.Add(-1)
	}
}

// Size 返回死信队列中的事件数量
func (dlq *DeadLetterQueue) Size() int {
	return int(dlq.size.Load())
}

// Events 返回所有死信事件（快照）
func (dlq *DeadLetterQueue) Events() []FailedEvent {
	result := make([]FailedEvent, 0)
	dlq.events.Range(func(key, value any) bool {
		result = append(result, value.(FailedEvent))
		return true
	})
	return result
}

// Clear 清空死信队列
func (dlq *DeadLetterQueue) Clear() {
	count := int64(0)
	dlq.events.Range(func(key, value any) bool {
		dlq.events.Delete(key)
		count++
		return true
	})
	dlq.size.Add(-count)
}

// GetByType 获取指定类型的所有失败事件（快照）
func (dlq *DeadLetterQueue) GetByType(eventType string) []FailedEvent {
	result := make([]FailedEvent, 0)
	dlq.events.Range(func(key, value any) bool {
		fe := value.(FailedEvent)
		if fe.Event.Type() == eventType {
			result = append(result, fe)
		}
		return true
	})
	return result
}

// RemoveByType 移除指定类型的所有事件
func (dlq *DeadLetterQueue) RemoveByType(eventType string) int {
	removed := 0
	dlq.events.Range(func(key, value any) bool {
		fe := value.(FailedEvent)
		if fe.Event.Type() == eventType {
			dlq.events.Delete(key)
			removed++
		}
		return true
	})
	dlq.size.Add(-int64(removed))
	return removed
}

// Stats 返回死信队列统计信息
func (dlq *DeadLetterQueue) Stats() DeadLetterStats {
	stats := DeadLetterStats{
		EventTypeCount: make(map[string]int),
	}

	dlq.events.Range(func(key, value any) bool {
		fe := value.(FailedEvent)
		stats.Total++
		if fe.ShouldRetry() {
			stats.Retryable++
		}
		if fe.IsExhausted() {
			stats.Exhausted++
		}
		stats.EventTypeCount[fe.Event.Type()]++
		return true
	})

	return stats
}

// DeadLetterStats 死信队列统计信息
type DeadLetterStats struct {
	Total          int            // 总事件数
	Retryable      int            // 可重试事件数
	Exhausted      int            // 已耗尽事件数
	EventTypeCount map[string]int // 按事件类型统计
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

// doneChanPool 复用 done channel，减少高频场景下的 GC 压力
var doneChanPool = sync.Pool{
	New: func() any {
		ch := make(chan struct{}, 1)
		return &ch
	},
}

func getDoneChan() *chan struct{} {
	return doneChanPool.Get().(*chan struct{})
}

func putDoneChan(ch *chan struct{}) {
	// 确保 channel 是空的再放回 pool
	select {
	case <-*ch:
	default:
	}
	doneChanPool.Put(ch)
}

func (b *EventBusWithDeadLetter) publishWithRecoveryInternal(event ApplicationEvent, attempt int, key string) {
	// 复用 done channel，减少 GC 压力
	donePtr := getDoneChan()
	defer putDoneChan(donePtr)

	var capturedErr error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					capturedErr = err
					*donePtr <- struct{}{}
					return
				}
				capturedErr = fmt.Errorf("event handler panic: %v", r)
			}
			*donePtr <- struct{}{}
		}()

		b.EventBusWithOrdering.Publish(event)
	}()

	<-(*donePtr)

	b.mu.Lock()
	delete(b.retrying, key)
	b.mu.Unlock()

	if capturedErr != nil {
		b.handleFailure(event, capturedErr, attempt, key)
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

	// 首次失败时记录 FirstFailedAt，重试时保留原始值
	if attempt == 0 {
		fe.FirstFailedAt = time.Now()
		return
	}
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
