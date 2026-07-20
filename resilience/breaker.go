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
	b.successes.Add(1)
	b.requests.Add(1)

	state := State(b.state.Load())
	switch state {
	case StateHalfOpen:
		successes := b.successes.Load()
		if successes >= int64(b.maxRequests) {
			b.mu.Lock()
			// 双重检查
			if State(b.state.Load()) == StateHalfOpen {
				b.transitionToClosed()
			}
			b.mu.Unlock()
		}
	case StateClosed:
		// 成功请求不需要检查阈值，但需要定期重置计数器防止误判
		// 使用更大的窗口避免高流量场景下频繁重置
		if b.requests.Load() >= 1000 {
			b.mu.Lock()
			// 双重检查：确保仍然是 Closed 状态且计数器达到阈值
			if State(b.state.Load()) == StateClosed && b.requests.Load() >= 1000 {
				b.resetCounters()
			}
			b.mu.Unlock()
		}
	}
}

func (b *breakerImpl) RecordFailure() {
	b.errors.Add(1)
	b.requests.Add(1)

	state := State(b.state.Load())
	switch state {
	case StateClosed:
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
		// 半开状态下任何失败都立即熔断
		b.mu.Lock()
		if State(b.state.Load()) == StateHalfOpen {
			b.transitionToOpen()
		}
		b.mu.Unlock()
	}
}

// transitionToClosed 转换到关闭状态（调用方必须持有锁）
func (b *breakerImpl) transitionToClosed() {
	b.state.Store(int32(StateClosed))
	b.resetCounters()
	b.lastStateChange.Store(time.Now().UnixNano())
}

// transitionToOpen 转换到打开状态（调用方必须持有锁）
func (b *breakerImpl) transitionToOpen() {
	b.state.Store(int32(StateOpen))
	b.lastStateChange.Store(time.Now().UnixNano())
}

// resetCounters 重置计数器（调用方必须持有锁）
func (b *breakerImpl) resetCounters() {
	b.errors.Store(0)
	b.successes.Store(0)
	b.requests.Store(0)
}

func (b *breakerImpl) State() State {
	return State(b.state.Load())
}
