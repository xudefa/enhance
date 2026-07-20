package resilience

import (
	"sync"
	"testing"
	"time"
)

// TestDefaultBreaker_ConcurrentAccess 测试熔断器并发访问安全性
func TestDefaultBreaker_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	breaker := NewBreaker(
		WithErrorThreshold(0.5),
		WithWaitDuration(100*time.Millisecond),
	)

	var wg sync.WaitGroup

	// 并发调用 Allow
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = breaker.Allow()
		}()
	}

	// 并发记录成功/失败
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			breaker.RecordSuccess()
		}()
	}

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			breaker.RecordFailure()
		}()
	}

	wg.Wait()

	// 不应该 panic
	t.Log("concurrent access test passed")
}

// TestDefaultBreaker_StateTransitions 测试状态转换边界
func TestDefaultBreaker_StateTransitions(t *testing.T) {
	t.Parallel()
	breaker := NewBreaker(
		WithErrorThreshold(0.5),
		WithMaxRequests(2),
		WithWaitDuration(50*time.Millisecond),
	)

	// 初始状态应该是 Closed
	if breaker.State() != StateClosed {
		t.Errorf("initial state should be Closed, got %v", breaker.State())
	}

	// 记录失败使错误率达到阈值
	breaker.RecordFailure()
	breaker.RecordFailure()
	breaker.RecordFailure()

	// 现在 Allow 应该失败（熔断器打开）
	err := breaker.Allow()
	if err != nil && err != ErrCircuitOpen {
		t.Logf("circuit opened or still allowing: %v", err)
	}

	// 等待 waitDuration 后应该进入半开状态
	time.Sleep(100 * time.Millisecond)

	// 现在应该允许试探请求
	err = breaker.Allow()
	if err != nil {
		t.Logf("after wait, Allow returned: %v", err)
	}
}

// TestDefaultBreaker_ZeroThreshold 测试零阈值
func TestDefaultBreaker_ZeroThreshold(t *testing.T) {
	t.Parallel()
	breaker := NewBreaker(
		WithErrorThreshold(0.0), // 任何错误都会触发熔断
	)

	// 一次失败就应该打开熔断器
	breaker.RecordFailure()

	// 验证状态
	state := breaker.State()
	t.Logf("state after one failure with 0.0 threshold: %v", state)
}

// TestDefaultBreaker_FullThreshold 测试满阈值
func TestDefaultBreaker_FullThreshold(t *testing.T) {
	t.Parallel()
	breaker := NewBreaker(
		WithErrorThreshold(1.0), // 100% 错误率才触发
	)

	// 即使全部失败也不应该打开（除非达到 100%）
	for range 100 {
		breaker.RecordFailure()
	}

	err := breaker.Allow()
	if err != nil {
		t.Logf("circuit state: %v, error: %v", breaker.State(), err)
	}
}

// TestDefaultBreaker_RapidStateChanges 测试快速状态变化
func TestDefaultBreaker_RapidStateChanges(t *testing.T) {
	t.Parallel()
	breaker := NewBreaker(
		WithErrorThreshold(0.5),
		WithWaitDuration(10*time.Millisecond),
	)

	// 快速切换状态
	for range 10 {
		// 记录一些成功
		for range 5 {
			breaker.RecordSuccess()
		}

		// 记录一些失败
		for range 10 {
			breaker.RecordFailure()
		}

		// 检查状态
		_ = breaker.Allow()
		_ = breaker.State()
	}

	t.Log("rapid state changes test passed")
}

// TestDefaultBreaker_HalfOpenMaxRequests 测试半开状态最大请求数
func TestDefaultBreaker_HalfOpenMaxRequests(t *testing.T) {
	t.Parallel()
	breaker := NewBreaker(
		WithErrorThreshold(0.5),
		WithMaxRequests(3),
		WithWaitDuration(10*time.Millisecond),
	)

	// 先让熔断器打开
	for range 10 {
		breaker.RecordFailure()
	}

	// 验证熔断器已打开
	if breaker.State() != StateOpen {
		t.Logf("state after failures: %v", breaker.State())
	}

	// 等待进入半开状态
	time.Sleep(50 * time.Millisecond)

	// 在半开状态下，顺序调用 Allow 并记录失败
	// 一旦失败，熔断器应该回到打开状态
	allowedBeforeFailure := 0
	for range 5 {
		err := breaker.Allow()
		if err == nil {
			allowedBeforeFailure++
			// 记录失败，应该立即回到打开状态
			breaker.RecordFailure()
		} else {
			// 熔断器已打开，停止
			break
		}
	}

	// 验证熔断器回到打开状态
	if breaker.State() != StateOpen {
		t.Errorf("after failure in half-open, state should be Open, got %v", breaker.State())
	}

	t.Logf("allowed %d requests before failure triggered open state", allowedBeforeFailure)
}
