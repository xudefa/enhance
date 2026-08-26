package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xudefa/enhance/retry"
)

var dlqKeyCounter atomic.Uint64

// BackoffStrategy 退避策略类型（保留向后兼容，委托给 retry 包）
type BackoffStrategy = retry.BackoffStrategy

const (
	BackoffNone        = retry.BackoffNone
	BackoffFixed       = retry.BackoffFixed
	BackoffExponential = retry.BackoffExponential
	BackoffLinear      = retry.BackoffLinear
)

// RetryPolicy 重试策略配置（保留向后兼容，委托给 retry 包）
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

// CalculateDelay 计算当前重试次数对应的延迟（委托给 retry 包）
func (p RetryPolicy) CalculateDelay(attempt int) time.Duration {
	rp := retry.RetryPolicy{
		MaxAttempts:  p.MaxRetries + 1,
		Strategy:     p.Strategy,
		InitialDelay: p.InitialDelay,
		MaxDelay:     p.MaxDelay,
		Multiplier:   p.Multiplier,
	}
	return rp.CalculateDelay(attempt)
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
	dlqKey        string           // DLQ 内部 key（由 Add 方法设置）
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
	mu                 sync.Mutex
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
	key := fmt.Sprintf("%s-%d", fe.Event.Type(), dlqKeyCounter.Add(1))
	fe.dlqKey = key

	// 检查是否是新 key，避免重复计数
	if _, loaded := dlq.events.LoadOrStore(key, fe); !loaded {
		dlq.size.Add(1)
		// 仅在新事件且已耗尽时触发永久失败回调
		if fe.IsExhausted() {
			if dlq.onPermanentFailure != nil {
				dlq.onPermanentFailure(fe)
			}
		}
		return
	}
	// key 已存在，更新值
	dlq.events.Store(key, fe)
}

// Peek 获取下一个可重试的事件（不移除）
func (dlq *DeadLetterQueue) Peek() (FailedEvent, bool) {
	var found FailedEvent
	var foundOk bool

	dlq.events.Range(func(key, value any) bool {
		fe, ok := value.(FailedEvent)
		if !ok {
			return true
		}
		if fe.ShouldRetry() {
			found = fe
			foundOk = true
			return false
		}
		return true
	})

	return found, foundOk
}

// Remove 移除指定事件（重试成功后调用）
//
// 匹配规则：按事件类型和发生时间匹配，而非接口指针地址。
func (dlq *DeadLetterQueue) Remove(event ApplicationEvent) {
	eventType := event.Type()
	eventTime := event.Timestamp()
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.events.Range(func(key, value any) bool {
		fe, ok := value.(FailedEvent)
		if !ok {
			return true
		}
		if fe.Event.Type() == eventType && fe.Event.Timestamp().Equal(eventTime) {
			dlq.events.Delete(key)
			dlq.size.Add(-1)
			return false
		}
		return true
	})
}

// Size 返回死信队列中的事件数量
func (dlq *DeadLetterQueue) Size() int {
	return int(dlq.size.Load())
}

// Events 返回所有死信事件（快照）
func (dlq *DeadLetterQueue) Events() []FailedEvent {
	result := make([]FailedEvent, 0)
	dlq.events.Range(func(key, value any) bool {
		if fe, ok := value.(FailedEvent); ok {
			result = append(result, fe)
		}
		return true
	})
	return result
}

// Clear 清空死信队列
func (dlq *DeadLetterQueue) Clear() {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.events.Range(func(key, value any) bool {
		dlq.events.Delete(key)
		dlq.size.Add(-1)
		return true
	})
}

// GetByType 获取指定类型的所有失败事件（快照）
func (dlq *DeadLetterQueue) GetByType(eventType string) []FailedEvent {
	result := make([]FailedEvent, 0)
	dlq.events.Range(func(key, value any) bool {
		fe, ok := value.(FailedEvent)
		if !ok {
			return true
		}
		if fe.Event.Type() == eventType {
			result = append(result, fe)
		}
		return true
	})
	return result
}

// RemoveByType 移除指定类型的所有事件
func (dlq *DeadLetterQueue) RemoveByType(eventType string) int {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	removed := 0
	dlq.events.Range(func(key, value any) bool {
		fe, ok := value.(FailedEvent)
		if !ok {
			return true
		}
		if fe.Event.Type() == eventType {
			dlq.events.Delete(key)
			dlq.size.Add(-1)
			removed++
		}
		return true
	})
	return removed
}

// Stats 返回死信队列统计信息
func (dlq *DeadLetterQueue) Stats() DeadLetterStats {
	stats := DeadLetterStats{
		EventTypeCount: make(map[string]int),
	}

	dlq.events.Range(func(key, value any) bool {
		fe, ok := value.(FailedEvent)
		if !ok {
			return true
		}
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
