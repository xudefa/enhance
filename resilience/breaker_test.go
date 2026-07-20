package resilience

import (
	"testing"
	"time"
)

func TestNewBreaker(t *testing.T) {
	t.Parallel()
	b := NewBreaker()
	if b == nil {
		t.Fatal("expected breaker to be created")
	}
	if b.State() != StateClosed {
		t.Errorf("expected initial state to be closed, got %v", b.State())
	}
}

func TestBreakerAllow(t *testing.T) {
	t.Parallel()
	b := NewBreaker()
	if err := b.Allow(); err != nil {
		t.Errorf("expected no error in closed state, got %v", err)
	}
}

func TestBreakerOpenAfterFailures(t *testing.T) {
	t.Parallel()
	b := NewBreaker(
		WithErrorThreshold(0.5),
		WithMaxRequests(5),
	)

	// 记录足够的失败请求以触发熔断
	for range 10 {
		_ = b.Allow()
		b.RecordFailure()
	}

	if b.State() != StateOpen {
		t.Errorf("expected state to be open, got %v", b.State())
	}

	if err := b.Allow(); err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreakerHalfOpenAfterWait(t *testing.T) {
	t.Parallel()
	b := NewBreaker(
		WithErrorThreshold(0.5),
		WithWaitDuration(50*time.Millisecond),
	)

	// 触发熔断
	for range 10 {
		_ = b.Allow()
		b.RecordFailure()
	}

	if b.State() != StateOpen {
		t.Fatalf("expected state to be open, got %v", b.State())
	}

	// 等待恢复时间
	time.Sleep(60 * time.Millisecond)

	if err := b.Allow(); err != nil {
		t.Errorf("expected no error in half-open state, got %v", err)
	}

	if b.State() != StateHalfOpen {
		t.Errorf("expected state to be half-open, got %v", b.State())
	}
}

func TestBreakerRecovery(t *testing.T) {
	t.Parallel()
	b := NewBreaker(
		WithErrorThreshold(0.5),
		WithWaitDuration(50*time.Millisecond),
		WithMaxRequests(3),
	)

	// 触发熔断
	for range 10 {
		_ = b.Allow()
		b.RecordFailure()
	}

	// 等待恢复
	time.Sleep(60 * time.Millisecond)

	// 成功请求恢复
	for range 3 {
		_ = b.Allow()
		b.RecordSuccess()
	}

	if b.State() != StateClosed {
		t.Errorf("expected state to be closed after recovery, got %v", b.State())
	}
}

func TestBreakerStateString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestBreakerOptions(t *testing.T) {
	t.Parallel()
	// 验证选项函数可以正常工作
	b := NewBreaker(
		WithMaxRequests(20),
		WithErrorThreshold(0.3),
		WithWaitDuration(60*time.Second),
	)

	// 验证熔断器可以正常工作
	if b.State() != StateClosed {
		t.Errorf("expected initial state to be closed, got %v", b.State())
	}
}
