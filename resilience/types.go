// Package resilience 提供弹性容错功能，用于 enhance 框架。
package resilience

import (
	"sync"
	"sync/atomic"
	"time"
)

// DefaultBreaker 默认熔断器实现（已弃用，请使用 Breaker 接口）。
//
// 使用 atomic 操作优化计数器的并发性能，
// 适用于高并发场景下的请求保护。
//
// 配置参数（只读，无需锁保护）：
//   - maxRequests: 半开状态下允许的最大试探请求数
//   - errorThreshold: 错误率阈值（0.0-1.0），超过此值触发熔断
//   - waitDuration: 打开状态等待时间，之后进入半开状态
//
// 状态管理（使用 atomic 操作）：
//   - state: 当前熔断器状态
//   - errors/successes/requests: 请求统计
//   - lastStateChange: 上次状态变更时间
//   - halfOpenRequests: 半开状态已处理请求数
type DefaultBreaker struct {
	// mu 仅用于保护状态转换的原子性。
	mu sync.Mutex

	// 配置参数（只读，无需锁保护）。
	maxRequests    int           // 半开状态下允许的最大请求数
	errorThreshold float64       // 错误率阈值
	waitDuration   time.Duration // 打开状态等待时间

	// 状态（使用 atomic 操作）。
	state            atomic.Int32 // State
	errors           atomic.Int64
	successes        atomic.Int64
	requests         atomic.Int64
	lastStateChange  atomic.Int64 // Unix 时间戳
	halfOpenRequests atomic.Int64
}

// String 返回状态的字符串表示。
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
