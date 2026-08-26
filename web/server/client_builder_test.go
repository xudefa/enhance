package server

import (
	"testing"
	"time"
)

func TestNewExponentialBackoff_Helper(t *testing.T) {
	t.Parallel()
	b := NewExponentialBackoff(0, 0)
	if b.baseDelay != 100*time.Millisecond {
		t.Errorf("baseDelay = %v, want 100ms", b.baseDelay)
	}
	if b.maxDelay != 10*time.Second {
		t.Errorf("maxDelay = %v, want 10s", b.maxDelay)
	}
	if len(b.retryableStatus) == 0 {
		t.Error("retryableStatus should have default values")
	}
}

func TestNewExponentialBackoff_CustomHelper(t *testing.T) {
	t.Parallel()
	b := NewExponentialBackoff(200*time.Millisecond, 5*time.Second, 500, 503)
	if b.baseDelay != 200*time.Millisecond {
		t.Errorf("baseDelay = %v, want 200ms", b.baseDelay)
	}
	if b.maxDelay != 5*time.Second {
		t.Errorf("maxDelay = %v, want 5s", b.maxDelay)
	}
	if len(b.retryableStatus) != 2 {
		t.Errorf("retryableStatus len = %d, want 2", len(b.retryableStatus))
	}
}

func TestNewFixedDelay_Helper(t *testing.T) {
	t.Parallel()
	d := NewFixedDelay(0)
	if d.delay != 1*time.Second {
		t.Errorf("delay = %v, want 1s", d.delay)
	}
	if len(d.retryableStatus) == 0 {
		t.Error("retryableStatus should have default values")
	}
}

func TestNewFixedDelay_CustomHelper(t *testing.T) {
	t.Parallel()
	d := NewFixedDelay(500*time.Millisecond, 500, 502)
	if d.delay != 500*time.Millisecond {
		t.Errorf("delay = %v, want 500ms", d.delay)
	}
	if len(d.retryableStatus) != 2 {
		t.Errorf("retryableStatus len = %d, want 2", len(d.retryableStatus))
	}
}

func TestNewClient_Helper(t *testing.T) {
	t.Parallel()
	c := NewClient("http://localhost:8080")
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %s, want http://localhost:8080", c.baseURL)
	}
}

func TestNewClient_WithTimeoutHelper(t *testing.T) {
	t.Parallel()
	c := NewClient("http://localhost:8080", WithClientTimeout(5*time.Second))
	if c.httpClient.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", c.httpClient.Timeout)
	}
}

func TestNewClient_WithHeadersHelper(t *testing.T) {
	t.Parallel()
	h := make(map[string][]string)
	h["X-Custom"] = []string{"value"}

	c := NewClient("http://localhost:8080", WithHeaders(h))
	if c.headers.Get("X-Custom") != "value" {
		t.Errorf("header X-Custom = %s, want value", c.headers.Get("X-Custom"))
	}
}

func TestNewRetryableClient_Helper(t *testing.T) {
	t.Parallel()
	c := NewRetryableClient(nil)
	if c == nil {
		t.Fatal("NewRetryableClient returned nil")
	}
	if c.config.maxAttempts != 3 {
		t.Errorf("maxAttempts = %d, want 3", c.config.maxAttempts)
	}
}

func TestNewRetryableClient_WithMaxAttemptsHelper(t *testing.T) {
	t.Parallel()
	c := NewRetryableClient(nil, WithMaxAttempts(5))
	if c.config.maxAttempts != 5 {
		t.Errorf("maxAttempts = %d, want 5", c.config.maxAttempts)
	}
}

func TestWithOnRetry_Helper(t *testing.T) {
	t.Parallel()
	cfg := &RetryConfig{}
	called := false
	opt := WithOnRetry(func(attempt int, resp *HTTPResponse, err error) {
		called = true
	})
	opt(cfg)

	if cfg.onRetry == nil {
		t.Fatal("onRetry should be set")
	}
	cfg.onRetry(1, nil, nil)
	if !called {
		t.Error("onRetry callback should be called")
	}
}

func TestWithHeader_Helper(t *testing.T) {
	t.Parallel()
	req := &HTTPRequest{}
	opt := WithHeader("X-Key", "X-Val")
	opt(req)

	if req.Header == nil {
		t.Fatal("Header should be initialized")
	}
	if req.Header.Get("X-Key") != "X-Val" {
		t.Errorf("header = %s, want X-Val", req.Header.Get("X-Key"))
	}
}

func TestWithQuery_Helper(t *testing.T) {
	t.Parallel()
	req := &HTTPRequest{}
	opt := WithQuery("page", "1")
	opt(req)

	if req.Query == nil {
		t.Fatal("Query should be initialized")
	}
	if req.Query.Get("page") != "1" {
		t.Errorf("query page = %s, want 1", req.Query.Get("page"))
	}
}

func TestWithTimeout_Helper(t *testing.T) {
	t.Parallel()
	req := &HTTPRequest{}
	opt := WithTimeout(5 * time.Second)
	opt(req)

	if req.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", req.Timeout)
	}
}

func TestWithAuthToken_Helper(t *testing.T) {
	t.Parallel()
	req := &HTTPRequest{}
	opt := WithAuthToken("my-token")
	opt(req)

	if req.AuthToken != "my-token" {
		t.Errorf("AuthToken = %s, want my-token", req.AuthToken)
	}
}

func TestWithContentType_Helper(t *testing.T) {
	t.Parallel()
	req := &HTTPRequest{}
	opt := WithContentType("application/json")
	opt(req)

	if req.ContentType != "application/json" {
		t.Errorf("ContentType = %s, want application/json", req.ContentType)
	}
}

func TestWithBasicAuth_Helper(t *testing.T) {
	t.Parallel()
	req := &HTTPRequest{}
	opt := WithBasicAuth("user", "pass")
	opt(req)

	if req.BasicAuth.Username != "user" {
		t.Errorf("Username = %s, want user", req.BasicAuth.Username)
	}
	if req.BasicAuth.Password != "pass" {
		t.Errorf("Password = %s, want pass", req.BasicAuth.Password)
	}
}

func TestWithRetryStrategy_Helper(t *testing.T) {
	t.Parallel()
	cfg := &RetryConfig{}
	strategy := NewFixedDelay(2 * time.Second)
	opt := WithRetryStrategy(strategy)
	opt(cfg)

	if cfg.strategy != strategy {
		t.Error("strategy should be set")
	}
}
