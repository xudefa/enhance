package retry

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestNoRetry(t *testing.T) {
	t.Parallel()
	policy := NoRetry()
	if policy.MaxAttempts != 1 {
		t.Errorf("expected MaxAttempts=1, got %d", policy.MaxAttempts)
	}
}

func TestFixedDelay(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 100*time.Millisecond)
	if policy.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", policy.MaxAttempts)
	}
	if policy.Strategy != BackoffFixed {
		t.Errorf("expected Strategy=BackoffFixed, got %s", policy.Strategy)
	}
}

func TestExponentialBackoff(t *testing.T) {
	t.Parallel()
	policy := ExponentialBackoff(5, 100*time.Millisecond, 10*time.Second)
	if policy.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", policy.MaxAttempts)
	}
	if policy.Strategy != BackoffExponential {
		t.Errorf("expected Strategy=BackoffExponential, got %s", policy.Strategy)
	}
	if policy.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", policy.Multiplier)
	}
}

func TestCalculateDelay_Fixed(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 100*time.Millisecond)

	delay0 := policy.CalculateDelay(0)
	if delay0 == 0 {
		t.Error("expected non-zero delay for attempt 0")
	}

	delay1 := policy.CalculateDelay(1)
	if delay1 == 0 {
		t.Error("expected non-zero delay for attempt 1")
	}

	delay2 := policy.CalculateDelay(2)
	if delay2 != 0 {
		t.Error("expected zero delay for last attempt")
	}
}

func TestCalculateDelay_Exponential(t *testing.T) {
	t.Parallel()
	policy := ExponentialBackoff(5, 100*time.Millisecond, 10*time.Second)

	delay0 := policy.CalculateDelay(0)
	if delay0 == 0 {
		t.Error("expected non-zero delay for attempt 0")
	}

	// 指数增长：attempt 1 应该比 attempt 0 大（考虑 jitter）
	delay1 := policy.CalculateDelay(1)
	if delay1 < delay0/2 {
		t.Errorf("expected delay1 >= delay0/2, got delay0=%v, delay1=%v", delay0, delay1)
	}
}

func TestCalculateDelay_MaxDelay(t *testing.T) {
	t.Parallel()
	policy := ExponentialBackoff(10, 100*time.Millisecond, 1*time.Second)

	// 大 attempt 应该被 MaxDelay 限制
	delay := policy.CalculateDelay(8)
	if delay > 1*time.Second {
		t.Errorf("expected delay <= 1s, got %v", delay)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		policy    RetryPolicy
		wantError bool
	}{
		{
			name:      "valid policy",
			policy:    FixedDelay(3, 100*time.Millisecond),
			wantError: false,
		},
		{
			name:      "invalid maxAttempts",
			policy:    RetryPolicy{MaxAttempts: 0},
			wantError: true,
		},
		{
			name:      "invalid jitter",
			policy:    RetryPolicy{MaxAttempts: 3, Jitter: 1.5},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.policy.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestExecutor_Success(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 10*time.Millisecond)
	exec, err := NewExecutor(policy)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := context.Background()
	var callCount atomic.Int32

	result, err := Execute(ctx, exec, func(ctx context.Context) (string, error) {
		callCount.Add(1)
		return "success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected result='success', got %s", result)
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call, got %d", callCount.Load())
	}
}

func TestExecutor_RetryThenSuccess(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 10*time.Millisecond)
	var retryCount atomic.Int32
	exec, err := NewExecutor(policy, WithOnRetry(func(info RetryInfo) {
		retryCount.Add(1)
	}))
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := context.Background()
	var callCount atomic.Int32

	result, err := Execute(ctx, exec, func(ctx context.Context) (string, error) {
		count := callCount.Add(1)
		if count < 3 {
			return "", fmt.Errorf("transient error %d", count)
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected result='success', got %s", result)
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", callCount.Load())
	}
	if retryCount.Load() != 2 {
		t.Errorf("expected 2 retries, got %d", retryCount.Load())
	}
}

func TestExecutor_ExhaustedRetries(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 10*time.Millisecond)
	exec, err := NewExecutor(policy)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := context.Background()
	var callCount atomic.Int32
	expectedErr := errors.New("persistent error")

	_, err = Execute(ctx, exec, func(ctx context.Context) (string, error) {
		callCount.Add(1)
		return "", expectedErr
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", callCount.Load())
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(5, 100*time.Millisecond)
	exec, err := NewExecutor(policy)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var callCount atomic.Int32

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err = Execute(ctx, exec, func(ctx context.Context) (string, error) {
		callCount.Add(1)
		return "", errors.New("transient error")
	})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if callCount.Load() < 1 {
		t.Errorf("expected at least 1 call, got %d", callCount.Load())
	}
}

func TestExecutor_CustomRetryable(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 10*time.Millisecond)
	exec, err := NewExecutor(policy, WithRetryable(func(err error) bool {
		// 只有特定错误才重试
		return err.Error() == "retryable"
	}))
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := context.Background()
	var callCount atomic.Int32

	_, err = Execute(ctx, exec, func(ctx context.Context) (string, error) {
		callCount.Add(1)
		return "", errors.New("non-retryable")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call (non-retryable), got %d", callCount.Load())
	}
}

func TestExecutor_ExecuteVoid(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 10*time.Millisecond)
	exec, err := NewExecutor(policy)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := context.Background()
	var callCount atomic.Int32

	err = exec.ExecuteVoid(ctx, func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount.Load() != 1 {
		t.Errorf("expected 1 call, got %d", callCount.Load())
	}
}

func TestMustNewExecutor(t *testing.T) {
	t.Parallel()
	policy := FixedDelay(3, 100*time.Millisecond)
	exec := MustNewExecutor(policy)
	if exec == nil {
		t.Error("expected non-nil executor")
	}
}

func TestMustNewExecutor_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid policy")
		}
	}()

	policy := RetryPolicy{MaxAttempts: 0}
	MustNewExecutor(policy)
}

func TestApplyJitter(t *testing.T) {
	t.Parallel()
	delay := 100 * time.Millisecond

	// jitter=0 应该返回原值
	result0 := applyJitter(delay, 0)
	if result0 != delay {
		t.Errorf("expected %v with jitter=0, got %v", delay, result0)
	}

	// jitter>0 应该产生随机值
	result1 := applyJitter(delay, 0.5)
	if result1 < 0 || result1 > delay*2 {
		t.Errorf("jitter result out of range: %v", result1)
	}
}

func TestLinearBackoff(t *testing.T) {
	t.Parallel()
	policy := LinearBackoff(5, 100*time.Millisecond, 1*time.Second)

	if policy.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", policy.MaxAttempts)
	}
	if policy.Strategy != BackoffLinear {
		t.Errorf("expected Strategy=BackoffLinear, got %s", policy.Strategy)
	}
	if policy.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected InitialDelay=100ms, got %v", policy.InitialDelay)
	}
	if policy.MaxDelay != 1*time.Second {
		t.Errorf("expected MaxDelay=1s, got %v", policy.MaxDelay)
	}

	// 测试线性增长的延迟
	delay0 := policy.CalculateDelay(0)
	delay1 := policy.CalculateDelay(1)

	// 线性退避：delay1应该比delay0大
	if delay1 < delay0 {
		t.Errorf("expected delay1 >= delay0 for linear backoff, got delay0=%v, delay1=%v", delay0, delay1)
	}
}

func TestMaxSafeShift(t *testing.T) {
	t.Parallel()

	// 测试正常情况
	shift := maxSafeShift(100*time.Millisecond, 10*time.Second)
	if shift == 0 {
		t.Error("expected non-zero shift for valid inputs")
	}

	// 测试baseDelay为0
	shift0 := maxSafeShift(0, 10*time.Second)
	if shift0 != 0 {
		t.Errorf("expected 0 shift for zero baseDelay, got %d", shift0)
	}

	// 测试baseDelay为负
	shiftNeg := maxSafeShift(-100*time.Millisecond, 10*time.Second)
	if shiftNeg != 0 {
		t.Errorf("expected 0 shift for negative baseDelay, got %d", shiftNeg)
	}
}
