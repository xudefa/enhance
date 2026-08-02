package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestExponentialBackoff_ShouldRetry_Error(t *testing.T) {
	t.Parallel()
	strategy := NewExponentialBackoff(100*time.Millisecond, 10*time.Second)

	if strategy.ShouldRetry(nil, context.DeadlineExceeded, 0) {
		t.Error("ShouldRetry should return false for context.DeadlineExceeded")
	}
	if strategy.ShouldRetry(nil, context.Canceled, 0) {
		t.Error("ShouldRetry should return false for context.Canceled")
	}
	if !strategy.ShouldRetry(nil, fmt.Errorf("network error"), 0) {
		t.Error("ShouldRetry should return true for other errors")
	}
}

func TestExponentialBackoff_ShouldRetry_StatusCodes(t *testing.T) {
	t.Parallel()
	strategy := NewExponentialBackoff(100*time.Millisecond, 10*time.Second)

	tests := []struct {
		statusCode int
		want       bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		resp := &HTTPResponse{StatusCode: tt.statusCode}
		if got := strategy.ShouldRetry(resp, nil, 0); got != tt.want {
			t.Errorf("ShouldRetry(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestExponentialBackoff_Delay(t *testing.T) {
	t.Parallel()
	strategy := NewExponentialBackoff(100*time.Millisecond, 10*time.Second)

	// 延迟应该随每次尝试增加（平均而言）
	delay0 := strategy.Delay(0)
	delay1 := strategy.Delay(1)
	delay2 := strategy.Delay(2)

	// 有抖动的情况下，我们只能检查延迟是正数且在范围内
	if delay0 <= 0 {
		t.Errorf("delay(0) = %v, should be > 0", delay0)
	}
	if delay1 <= 0 {
		t.Errorf("delay(1) = %v, should be > 0", delay1)
	}
	if delay2 <= 0 {
		t.Errorf("delay(2) = %v, should be > 0", delay2)
	}
}

func TestExponentialBackoff_MaxDelay(t *testing.T) {
	t.Parallel()
	strategy := NewExponentialBackoff(100*time.Millisecond, 500*time.Millisecond)

	// 足够多的尝试后，延迟应该被限制在 maxDelay
	delay := strategy.Delay(10)
	if delay > 500*time.Millisecond {
		t.Errorf("delay(10) = %v should be <= 500ms", delay)
	}
}

func TestFixedDelay_ShouldRetry(t *testing.T) {
	t.Parallel()
	strategy := NewFixedDelay(1 * time.Second)

	if strategy.ShouldRetry(nil, context.DeadlineExceeded, 0) {
		t.Error("ShouldRetry should return false for context.DeadlineExceeded")
	}
	if strategy.ShouldRetry(nil, context.Canceled, 0) {
		t.Error("ShouldRetry should return false for context.Canceled")
	}
	if !strategy.ShouldRetry(nil, fmt.Errorf("network error"), 0) {
		t.Error("ShouldRetry should return true for other errors")
	}

	resp := &HTTPResponse{StatusCode: http.StatusInternalServerError}
	if !strategy.ShouldRetry(resp, nil, 0) {
		t.Error("ShouldRetry should return true for 500 status")
	}

	resp2 := &HTTPResponse{StatusCode: http.StatusOK}
	if strategy.ShouldRetry(resp2, nil, 0) {
		t.Error("ShouldRetry should return false for 200 status")
	}
}

func TestFixedDelay_Delay(t *testing.T) {
	t.Parallel()
	strategy := NewFixedDelay(500 * time.Millisecond)

	delay0 := strategy.Delay(0)
	delay1 := strategy.Delay(1)
	delay2 := strategy.Delay(2)

	if delay0 != 500*time.Millisecond {
		t.Errorf("delay(0) = %v, want 500ms", delay0)
	}
	if delay1 != 500*time.Millisecond {
		t.Errorf("delay(1) = %v, want 500ms", delay1)
	}
	if delay2 != 500*time.Millisecond {
		t.Errorf("delay(2) = %v, want 500ms", delay2)
	}
}

func TestRetryableClient_Success(t *testing.T) {
	t.Parallel()
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
	)

	resp, err := retryableClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("attempts = %d, want 1", atomic.LoadInt32(&attempts))
	}
}

func TestRetryableClient_RetryOnFailure(t *testing.T) {
	t.Parallel()
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
	)

	resp, err := retryableClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestRetryableClient_ExhaustedRetries(t *testing.T) {
	t.Parallel()
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
	)

	resp, err := retryableClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
	}
}

func TestRetryableClient_OnRetry(t *testing.T) {
	t.Parallel()
	var attempts int32
	var retryCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond)),
		WithOnRetry(func(attempt int, resp *HTTPResponse, err error) {
			atomic.AddInt32(&retryCount, 1)
		}),
	)

	_, err := retryableClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if atomic.LoadInt32(&retryCount) != 1 {
		t.Errorf("retryCount = %d, want 1", atomic.LoadInt32(&retryCount))
	}
}

func TestRetryableClient_ContextCancellation(t *testing.T) {
	t.Parallel()
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(5),
		WithRetryStrategy(NewExponentialBackoff(100*time.Millisecond, 1*time.Second)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := retryableClient.Get(ctx, "/")
	if err != context.DeadlineExceeded {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestRetryableClient_Post(t *testing.T) {
	t.Parallel()
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewFixedDelay(10*time.Millisecond)),
	)

	resp, err := retryableClient.Post(context.Background(), "/", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestRetryableClient_Put(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewFixedDelay(10*time.Millisecond)),
	)

	resp, err := retryableClient.Put(context.Background(), "/", nil)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRetryableClient_Delete(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewFixedDelay(10*time.Millisecond)),
	)

	resp, err := retryableClient.Delete(context.Background(), "/")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestRetryableClient_Patch(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewFixedDelay(10*time.Millisecond)),
	)

	resp, err := retryableClient.Patch(context.Background(), "/", nil)
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRetryableClient_Head(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewFixedDelay(10*time.Millisecond)),
	)

	resp, err := retryableClient.Head(context.Background(), "/")
	if err != nil {
		t.Fatalf("Head() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRetryableClient_Options(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	retryableClient := NewRetryableClient(client,
		WithMaxAttempts(3),
		WithRetryStrategy(NewFixedDelay(10*time.Millisecond)),
	)

	resp, err := retryableClient.Options(context.Background(), "/")
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// 初始状态应该是关闭的
	if cb.GetState() != CircuitClosed {
		t.Errorf("initial state = %v, want CircuitClosed", cb.GetState())
	}

	// 关闭时应该允许请求
	if !cb.AllowRequest() {
		t.Error("should allow requests when closed")
	}

	// 记录失败
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != CircuitClosed {
		t.Errorf("state after 2 failures = %v, want CircuitClosed", cb.GetState())
	}

	// 第三次失败应该打开断路器
	cb.RecordFailure()
	if cb.GetState() != CircuitOpen {
		t.Errorf("state after 3 failures = %v, want CircuitOpen", cb.GetState())
	}

	// 打开时不应该允许请求
	if cb.AllowRequest() {
		t.Error("should not allow requests when open (before timeout)")
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(2, 50*time.Millisecond)

	// 打开断路器
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitOpen {
		t.Errorf("state = %v, want CircuitOpen", cb.GetState())
	}

	// 等待重置超时
	time.Sleep(100 * time.Millisecond)

	// 应该允许请求（转换到半开状态）
	if !cb.AllowRequest() {
		t.Error("should allow request after timeout (half-open)")
	}

	if cb.GetState() != CircuitHalfOpen {
		t.Errorf("state = %v, want CircuitHalfOpen", cb.GetState())
	}

	// 成功应该关闭断路器
	cb.RecordSuccess()
	if cb.GetState() != CircuitClosed {
		t.Errorf("state after success = %v, want CircuitClosed", cb.GetState())
	}
}

func TestCircuitBreaker_FailureResetsOnSuccess(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// 记录一些失败
	cb.RecordFailure()
	cb.RecordFailure()

	// 成功应该重置失败计数
	cb.RecordSuccess()

	// 再两次失败不应该打开断路器（计数已重置）
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitClosed {
		t.Errorf("state = %v, want CircuitClosed", cb.GetState())
	}
}

func TestCircuitBreakerClient_Success(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(3),
		WithCircuitResetTimeout(30*time.Second),
	)

	resp, err := cbClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if cbClient.GetCircuitState() != CircuitClosed {
		t.Errorf("circuit state = %v, want CircuitClosed", cbClient.GetCircuitState())
	}
}

func TestCircuitBreakerClient_Failure(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(2),
		WithCircuitResetTimeout(30*time.Second),
	)

	// 第一次失败
	_, err := cbClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// 第二次失败应该打开断路器
	_, err = cbClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if cbClient.GetCircuitState() != CircuitOpen {
		t.Errorf("circuit state = %v, want CircuitOpen", cbClient.GetCircuitState())
	}
}

func TestCircuitBreakerClient_Fallback(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)

	fallbackCalled := false
	fallback := func(ctx context.Context) (*HTTPResponse, error) {
		fallbackCalled = true
		return &HTTPResponse{StatusCode: http.StatusServiceUnavailable}, nil
	}

	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(1),
		WithCircuitResetTimeout(30*time.Second),
		WithFallback(fallback),
	)

	// 第一次请求失败并打开断路器
	_, err := cbClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// 第二个请求应该使用降级
	resp, err := cbClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !fallbackCalled {
		t.Error("fallback should be called")
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
}

func TestCircuitBreakerClient_Post(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(3),
		WithCircuitResetTimeout(30*time.Second),
	)

	resp, err := cbClient.Post(context.Background(), "/", nil)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestCircuitBreakerClient_Put(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(3),
		WithCircuitResetTimeout(30*time.Second),
	)

	resp, err := cbClient.Put(context.Background(), "/", nil)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCircuitBreakerClient_Delete(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(3),
		WithCircuitResetTimeout(30*time.Second),
	)

	resp, err := cbClient.Delete(context.Background(), "/")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestCircuitBreakerClient_Close(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:8080")
	cbClient := NewCircuitBreakerClient(client)

	if err := cbClient.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// contextErrClient 返回固定错误的 HTTP 客户端桩，用于测试客户端取消/超时。
type contextErrClient struct {
	err error
}

func (c *contextErrClient) Get(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Head(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Post(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Put(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Patch(ctx context.Context, url string, body any, opts ...RequestOption) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Delete(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Options(ctx context.Context, url string, opts ...RequestOption) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Do(ctx context.Context, req any) (*HTTPResponse, error) {
	return nil, c.err
}

func (c *contextErrClient) Close() error {
	return nil
}

func TestCircuitBreakerClient_ClientCancellationDoesNotTripBreaker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{"canceled", context.Canceled},
		{"deadline exceeded", context.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cbClient := NewCircuitBreakerClient(&contextErrClient{err: tt.err},
				WithCircuitMaxFailures(1),
				WithCircuitResetTimeout(30*time.Second),
			)

			// 即使 maxFailures=1，客户端自身的取消/超时也不应触发熔断
			for i := 0; i < 5; i++ {
				_, err := cbClient.Get(context.Background(), "/")
				if !errors.Is(err, tt.err) {
					t.Fatalf("Get() error = %v, want %v", err, tt.err)
				}
			}

			if state := cbClient.GetCircuitState(); state != CircuitClosed {
				t.Errorf("circuit state = %v, want CircuitClosed (client %s must not trip breaker)", state, tt.name)
			}
		})
	}
}

func TestCircuitBreakerClient_ServerErrorStillTripsBreaker(t *testing.T) {
	t.Parallel()
	// 真正的服务端错误（5xx）仍应触发熔断
	cbClient := NewCircuitBreakerClient(&contextErrClient{err: fmt.Errorf("upstream 502")},
		WithCircuitMaxFailures(1),
		WithCircuitResetTimeout(30*time.Second),
	)

	_, err := cbClient.Get(context.Background(), "/")
	if err == nil {
		t.Fatal("expected error")
	}

	if state := cbClient.GetCircuitState(); state != CircuitOpen {
		t.Errorf("circuit state = %v, want CircuitOpen (service error must trip breaker)", state)
	}
}
