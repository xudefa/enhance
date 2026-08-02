// Package resilience 提供弹性容错功能，用于 enhance 框架。
package resilience

import (
	"sync"
	"sync/atomic"
	"time"
)

// breakerImpl Breaker 接口的默认实现。
type breakerImpl struct {
	state            atomic.Int32
	lastStateChange  atomic.Int64
	errors           atomic.Int64
	successes        atomic.Int64
	requests         atomic.Int64
	halfOpenRequests atomic.Int64
	maxRequests      int
	errorThreshold   float64
	waitDuration     time.Duration
	mu               sync.Mutex
}

// WithMaxRequests 设置半开状态最大请求数。
func WithMaxRequests(max int) BreakerOption {
	return func(b Breaker) {
		if impl, ok := b.(*breakerImpl); ok {
			impl.maxRequests = max
		}
	}
}

// WithErrorThreshold 设置错误率阈值。
func WithErrorThreshold(threshold float64) BreakerOption {
	return func(b Breaker) {
		if impl, ok := b.(*breakerImpl); ok {
			impl.errorThreshold = threshold
		}
	}
}

// WithWaitDuration 设置打开状态等待时间。
func WithWaitDuration(duration time.Duration) BreakerOption {
	return func(b Breaker) {
		if impl, ok := b.(*breakerImpl); ok {
			impl.waitDuration = duration
		}
	}
}

// NewBreaker 创建熔断器。
func NewBreaker(opts ...BreakerOption) Breaker {
	b := &breakerImpl{
		maxRequests:    DefaultMaxRequests,
		errorThreshold: DefaultErrorThreshold,
		waitDuration:   DefaultWaitDuration,
	}
	b.state.Store(int32(StateClosed))
	b.lastStateChange.Store(time.Now().UnixNano())

	for _, opt := range opts {
		opt(b)
	}

	return b
}

func (b *breakerImpl) Allow() error {
	state := State(b.state.Load())

	switch state {
	case StateClosed:
		return nil
	case StateOpen:
		lastChange := time.Unix(0, b.lastStateChange.Load())
		if time.Since(lastChange) > b.waitDuration {
			// 尝试转换为半开状态
			b.mu.Lock()
			// 双重检查
			if State(b.state.Load()) == StateOpen && time.Since(lastChange) > b.waitDuration {
				b.state.Store(int32(StateHalfOpen))
				b.halfOpenRequests.Store(0)
				b.successes.Store(0)
				b.errors.Store(0)
				b.requests.Store(0)
				b.lastStateChange.Store(time.Now().UnixNano())
			}
			b.mu.Unlock()
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		halfOpen := b.halfOpenRequests.Add(1)
		if halfOpen <= int64(b.maxRequests) {
			return nil
		}
		return ErrCircuitHalfOpen
	}

	return ErrCircuitOpen
}

func (b *breakerImpl) RecordSuccess() {
	state := State(b.state.Load())
	switch state {
	case StateHalfOpen:
		b.successes.Add(1)
		b.requests.Add(1)
		b.mu.Lock()
		if State(b.state.Load()) == StateHalfOpen {
			// 释放半开试探请求的许可证，避免并发试探数累积
			b.halfOpenRequests.Add(-1)
			if b.successes.Load() >= int64(b.maxRequests) {
				b.transitionToClosed()
			}
		}
		b.mu.Unlock()
	case StateClosed:
		b.successes.Add(1)
		b.requests.Add(1)
		// 成功请求不需要检查阈值，但需要定期重置计数器防止误判
		// 使用更大的窗口避免高流量场景下频繁重置
		const resetWindow = 1000
		if b.requests.Load() >= resetWindow {
			b.mu.Lock()
			// 双重检查：确保仍然是 Closed 状态且计数器达到阈值
			if State(b.state.Load()) == StateClosed && b.requests.Load() >= resetWindow {
				b.resetCounters()
			}
			b.mu.Unlock()
		}
	}
}

func (b *breakerImpl) RecordFailure() {
	state := State(b.state.Load())
	// 仅在 CLOSED / HALF_OPEN 状态统计失败，避免 OPEN 状态的计数
	// 在恢复后污染统计，导致错误率失真。
	switch state {
	case StateClosed:
		b.errors.Add(1)
		b.requests.Add(1)

		total := b.requests.Load()
		if total >= 10 {
			errorRate := float64(b.errors.Load()) / float64(total)
			if errorRate >= b.errorThreshold {
				b.mu.Lock()
				// 双重检查
				if State(b.state.Load()) == StateClosed {
					b.transitionToOpen()
				}
				b.mu.Unlock()
			}
		}
	case StateHalfOpen:
		b.errors.Add(1)
		b.requests.Add(1)
		// 半开状态下任何失败都立即熔断
		b.mu.Lock()
		if State(b.state.Load()) == StateHalfOpen {
			b.halfOpenRequests.Add(-1)
			b.transitionToOpen()
		}
		b.mu.Unlock()
	}
}

// transitionToClosed 转换到关闭状态（调用方必须持有锁）
func (b *breakerImpl) transitionToClosed() {
	b.state.Store(int32(StateClosed))
	b.resetCounters()
	b.halfOpenRequests.Store(0)
	b.lastStateChange.Store(time.Now().UnixNano())
}

// transitionToOpen 转换到打开状态（调用方必须持有锁）
func (b *breakerImpl) transitionToOpen() {
	b.state.Store(int32(StateOpen))
	b.halfOpenRequests.Store(0)
	b.lastStateChange.Store(time.Now().UnixNano())
}

// resetCounters 重置计数器（调用方必须持有锁）。
//
// 注意：由于 errors/requests 使用 atomic 操作在锁外递增，
// 重置可能丢失锁获取与重置之间由其他 goroutine 递增的少量计数。
// 这是可接受的，因为重置仅用于防止计数器溢出，不影响熔断器正确性。
func (b *breakerImpl) resetCounters() {
	b.errors.Store(0)
	b.successes.Store(0)
	b.requests.Store(0)
}

func (b *breakerImpl) State() State {
	return State(b.state.Load())
}
