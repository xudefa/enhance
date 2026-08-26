package retry

import (
	"context"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"time"
)

// BackoffStrategy 退避策略
type BackoffStrategy string

const (
	BackoffNone        BackoffStrategy = "none"        // 无退避，立即重试
	BackoffFixed       BackoffStrategy = "fixed"       // 固定间隔退避
	BackoffLinear      BackoffStrategy = "linear"      // 线性退避
	BackoffExponential BackoffStrategy = "exponential" // 指数退避
)

// RetryPolicy 重试策略配置
type RetryPolicy struct {
	MaxAttempts  int             // 最大尝试次数（包含首次），0 表示不重试
	Strategy     BackoffStrategy // 退避策略
	InitialDelay time.Duration   // 初始延迟
	MaxDelay     time.Duration   // 最大延迟（指数退避上限）
	Multiplier   float64         // 退避乘数（指数退避用）
	Jitter       float64         // 抖动比例 (0.0-1.0)，防止惊群效应
}

// NoRetry 创建不重试策略
func NoRetry() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 1,
	}
}

// FixedDelay 创建固定延迟重试策略
func FixedDelay(maxAttempts int, delay time.Duration) RetryPolicy {
	return RetryPolicy{
		MaxAttempts:  maxAttempts,
		Strategy:     BackoffFixed,
		InitialDelay: delay,
		Jitter:       0.1, // 默认 10% 抖动
	}
}

// LinearBackoff 创建线性退避策略
func LinearBackoff(maxAttempts int, initialDelay, maxDelay time.Duration) RetryPolicy {
	return RetryPolicy{
		MaxAttempts:  maxAttempts,
		Strategy:     BackoffLinear,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Jitter:       0.1,
	}
}

// ExponentialBackoff 创建指数退避策略（带 jitter 防惊群）
func ExponentialBackoff(maxAttempts int, initialDelay, maxDelay time.Duration) RetryPolicy {
	return RetryPolicy{
		MaxAttempts:  maxAttempts,
		Strategy:     BackoffExponential,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   2.0,
		Jitter:       0.2, // 默认 20% 抖动
	}
}

// Validate 验证策略配置是否合法
func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 1 {
		return fmt.Errorf("maxAttempts must be at least 1")
	}
	if p.InitialDelay < 0 {
		return fmt.Errorf("initialDelay must be non-negative")
	}
	if p.MaxDelay < 0 {
		return fmt.Errorf("maxDelay must be non-negative")
	}
	if p.Jitter < 0 || p.Jitter > 1 {
		return fmt.Errorf("jitter must be between 0 and 1")
	}
	return nil
}

// CalculateDelay 计算当前重试次数对应的延迟（含 jitter）
func (p RetryPolicy) CalculateDelay(attempt int) time.Duration {
	if p.MaxAttempts <= 1 || attempt >= p.MaxAttempts-1 {
		return 0
	}

	var delay time.Duration
	switch p.Strategy {
	case BackoffNone:
		return 0
	case BackoffFixed:
		delay = p.InitialDelay
	case BackoffLinear:
		delay = p.InitialDelay * time.Duration(attempt+1)
	case BackoffExponential:
		shift := attempt
		if shift > 62 {
			shift = 62
		}
		delay = p.InitialDelay * time.Duration(float64(int64(1)<<uint(shift))*p.Multiplier)
	default:
		delay = p.InitialDelay
	}

	// 应用 jitter 防惊群（在 MaxDelay 限制之前，避免抖动后超出上限）
	if p.Jitter > 0 {
		delay = applyJitter(delay, p.Jitter)
	}

	// 限制最大延迟
	const maxDuration = 24 * time.Hour
	if delay < 0 || delay > maxDuration {
		delay = maxDuration
	}

	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	return delay
}

// applyJitter 对延迟应用随机抖动
func applyJitter(delay time.Duration, jitterRatio float64) time.Duration {
	if jitterRatio <= 0 {
		return delay
	}
	if jitterRatio >= 1 {
		return time.Duration(rand.Int63n(int64(delay)))
	}

	// 在 [delay*(1-jitter), delay*(1+jitter)] 范围内随机
	jitter := time.Duration(float64(delay) * jitterRatio)
	minDelay := delay - jitter
	if minDelay < 0 {
		minDelay = 0
	}
	maxJitter := jitter * 2
	return minDelay + time.Duration(rand.Int63n(int64(maxJitter)))
}

// RetryableFunc 可重试的函数签名
type RetryableFunc[T any] func(ctx context.Context) (T, error)

// RetryInfo 重试信息（用于回调）
type RetryInfo struct {
	Attempt     int           // 当前尝试次数（从 0 开始）
	MaxAttempts int           // 最大尝试次数
	Delay       time.Duration // 下次重试延迟
	LastErr     error         // 最后一次错误
}

// OnRetryFunc 重试回调函数签名
type OnRetryFunc func(info RetryInfo)

// Executor 重试执行器
type Executor struct {
	policy      RetryPolicy
	isRetryable func(error) bool // 判断错误是否可重试
	onRetry     OnRetryFunc      // 重试回调
}

// NewExecutor 创建重试执行器
func NewExecutor(policy RetryPolicy, opts ...ExecutorOption) (*Executor, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}

	exec := &Executor{
		policy: policy,
		isRetryable: func(err error) bool {
			return err != nil // 默认所有错误都可重试
		},
	}

	for _, opt := range opts {
		opt(exec)
	}

	return exec, nil
}

// MustNewExecutor 创建重试执行器（失败时 panic）
func MustNewExecutor(policy RetryPolicy, opts ...ExecutorOption) *Executor {
	exec, err := NewExecutor(policy, opts...)
	if err != nil {
		panic(err)
	}
	return exec
}

// ExecutorOption 执行器选项
type ExecutorOption func(*Executor)

// WithRetryable 设置可重试错误判断函数
func WithRetryable(fn func(error) bool) ExecutorOption {
	return func(e *Executor) {
		e.isRetryable = fn
	}
}

// WithOnRetry 设置重试回调
func WithOnRetry(fn OnRetryFunc) ExecutorOption {
	return func(e *Executor) {
		e.onRetry = fn
	}
}

// Execute 执行带重试的函数
func Execute[T any](ctx context.Context, exec *Executor, fn RetryableFunc[T]) (T, error) {
	var lastErr error
	var lastResult T

	for attempt := 0; attempt < exec.policy.MaxAttempts; attempt++ {
		lastResult, lastErr = fn(ctx)

		// 成功或不可重试错误，直接返回
		if lastErr == nil || !exec.isRetryable(lastErr) {
			return lastResult, lastErr
		}

		// 最后一次尝试，不再延迟
		if attempt == exec.policy.MaxAttempts-1 {
			break
		}

		// 触发重试回调
		if exec.onRetry != nil {
			delay := exec.policy.CalculateDelay(attempt)
			exec.onRetry(RetryInfo{
				Attempt:     attempt + 1,
				MaxAttempts: exec.policy.MaxAttempts,
				Delay:       delay,
				LastErr:     lastErr,
			})
		}

		// 计算延迟并等待
		delay := exec.policy.CalculateDelay(attempt)
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return lastResult, ctx.Err()
			case <-timer.C:
				timer.Stop()
			}
		}
	}

	return lastResult, lastErr
}

// ExecuteVoid 执行无返回值的重试函数
func (e *Executor) ExecuteVoid(ctx context.Context, fn func(ctx context.Context) error) error {
	_, err := Execute(ctx, e, func(ctx context.Context) (any, error) {
		return nil, fn(ctx)
	})
	return err
}

// maxSafeShift 计算 baseDelay*(1<<shift) 不会溢出且不超过 maxDelay 的最大位移量
func maxSafeShift(baseDelay, maxDelay time.Duration) uint {
	if baseDelay <= 0 {
		return 0
	}
	limit := maxDelay
	if maxDiv := int64(math.MaxInt64) / int64(baseDelay); maxDiv > 0 && maxDiv < int64(limit) {
		limit = time.Duration(maxDiv)
	}
	ratio := limit / baseDelay
	if ratio <= 0 {
		return 0
	}
	return uint(bits.Len64(uint64(ratio))) - 1
}
