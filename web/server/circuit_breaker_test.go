package server

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewCircuitBreaker_Defaults(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(0, 0)

	if cb.maxFailures != 5 {
		t.Errorf("expected maxFailures=5, got %d", cb.maxFailures)
	}
	if cb.resetTimeout != 30*time.Second {
		t.Errorf("expected resetTimeout=30s, got %v", cb.resetTimeout)
	}
	if cb.state != CircuitClosed {
		t.Errorf("expected state=CircuitClosed, got %v", cb.state)
	}
}

func TestNewCircuitBreaker_Custom(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(10, 60*time.Second)

	if cb.maxFailures != 10 {
		t.Errorf("expected maxFailures=10, got %d", cb.maxFailures)
	}
	if cb.resetTimeout != 60*time.Second {
		t.Errorf("expected resetTimeout=60s, got %v", cb.resetTimeout)
	}
}

func TestCircuitBreaker_AllowRequest_Closed(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(3, 1*time.Second)

	if !cb.AllowRequest() {
		t.Error("expected to allow request in closed state")
	}
}

func TestCircuitBreaker_RecordFailure(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(3, 1*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitClosed {
		t.Error("expected closed state after 2 failures (threshold=3)")
	}

	cb.RecordFailure()

	if cb.GetState() != CircuitOpen {
		t.Error("expected open state after 3 failures")
	}
}

func TestCircuitBreaker_AllowRequest_Open(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(3, 1*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.AllowRequest() {
		t.Error("expected to reject request in open state")
	}
}

func TestCircuitBreaker_AllowRequest_HalfOpen(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitOpen {
		t.Error("expected open state")
	}

	time.Sleep(150 * time.Millisecond)

	if !cb.AllowRequest() {
		t.Error("expected to allow request after timeout (half-open)")
	}

	if cb.GetState() != CircuitHalfOpen {
		t.Error("expected half-open state")
	}
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(3, 1*time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	if cb.failures != 0 {
		t.Errorf("expected failures=0 after success, got %d", cb.failures)
	}
	if cb.GetState() != CircuitClosed {
		t.Error("expected closed state after success")
	}
}

func TestNewCircuitBreakerClient(t *testing.T) {
	t.Parallel()

	client := NewClient("http://localhost:8080")
	cbClient := NewCircuitBreakerClient(client)

	if cbClient == nil {
		t.Fatal("expected non-nil CircuitBreakerClient")
	}
	if cbClient.breaker == nil {
		t.Fatal("expected non-nil breaker")
	}
}

func TestCircuitBreakerClient_WithOptions(t *testing.T) {
	t.Parallel()

	client := NewClient("http://localhost:8080")
	cbClient := NewCircuitBreakerClient(client,
		WithCircuitMaxFailures(10),
		WithCircuitResetTimeout(60*time.Second),
		WithFallback(func(ctx context.Context) (*HTTPResponse, error) {
			return &HTTPResponse{StatusCode: 503}, nil
		}),
	)

	if cbClient.breaker.maxFailures != 10 {
		t.Errorf("expected maxFailures=10, got %d", cbClient.breaker.maxFailures)
	}
	if cbClient.breaker.resetTimeout != 60*time.Second {
		t.Errorf("expected resetTimeout=60s, got %v", cbClient.breaker.resetTimeout)
	}
	if cbClient.fallback == nil {
		t.Error("expected non-nil fallback")
	}
}

func TestCircuitBreakerClient_GetCircuitState(t *testing.T) {
	t.Parallel()

	client := NewClient("http://localhost:8080")
	cbClient := NewCircuitBreakerClient(client)

	state := cbClient.GetCircuitState()
	if state != CircuitClosed {
		t.Errorf("expected CircuitClosed, got %v", state)
	}
}

func TestNetClient_Close(t *testing.T) {
	t.Parallel()

	client := NewClient("http://localhost:8080")

	err := client.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNetClient_WithMiddleware(t *testing.T) {
	t.Parallel()

	client := NewClient("http://localhost:8080")
	middleware := func(req *http.Request, resp *HTTPResponse) error {
		return nil
	}

	result := client.WithMiddleware(ClientMiddlewareFunc(middleware))
	if result != client {
		t.Error("expected WithMiddleware to return the same client")
	}
}

func TestRetryableClient_Close(t *testing.T) {
	t.Parallel()

	netClient := NewClient("http://localhost:8080")
	retryClient := NewRetryableClient(netClient)

	err := retryClient.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
