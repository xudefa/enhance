package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
