package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClientBuilder_Build(t *testing.T) {
	t.Parallel()
	builder := NewHTTPClientBuilder()
	builder.BaseURL("http://localhost:8080")
	builder.Timeout(5 * time.Second)
	builder.Header("X-Custom", "value")

	client, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if client == nil {
		t.Fatal("Build() returned nil client")
	}
}

func TestHTTPClientBuilder_Build_MissingBaseURL(t *testing.T) {
	t.Parallel()
	builder := NewHTTPClientBuilder()

	_, err := builder.Build()
	if err == nil {
		t.Error("Build() should error when baseURL is missing")
	}
}

func TestHTTPClientBuilder_MustBuild(t *testing.T) {
	t.Parallel()
	builder := NewHTTPClientBuilder()
	builder.BaseURL("http://localhost:8080")

	defer func() {
		if r := recover(); r != nil {
			t.Error("MustBuild() should not panic with valid config")
		}
	}()

	client := builder.MustBuild()
	if client == nil {
		t.Fatal("MustBuild() returned nil client")
	}
}

func TestHTTPClientBuilder_MustBuild_Panic(t *testing.T) {
	t.Parallel()
	builder := NewHTTPClientBuilder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBuild() should panic when baseURL is missing")
		}
	}()

	builder.MustBuild()
}

func TestHTTPClientBuilder_WithMiddleware(t *testing.T) {
	t.Parallel()
	middlewareCalled := false

	builder := NewHTTPClientBuilder()
	builder.BaseURL("http://localhost:8080")
	builder.Middleware(func(req *http.Request, resp *HTTPResponse) error {
		middlewareCalled = true
		return nil
	})

	client, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// 应该是 NetClient
	netClient, ok := client.(*NetClient)
	if !ok {
		t.Fatal("client should be *NetClient")
	}

	// 发起请求以触发中间件
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	netClient.baseURL = ts.URL

	resp, err := netClient.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !middlewareCalled {
		t.Error("middleware should be called")
	}
}

func TestHTTPClientBuilder_WithRetry(t *testing.T) {
	t.Parallel()
	builder := NewHTTPClientBuilder()
	builder.BaseURL("http://localhost:8080")
	builder.Retry(
		WithMaxAttempts(5),
		WithRetryStrategy(NewFixedDelay(100*time.Millisecond)),
	)

	client, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// 应该返回 RetryableClient
	_, ok := client.(*RetryableClient)
	if !ok {
		t.Error("Build() should return *RetryableClient when retry is configured")
	}
}

func TestRequestBuilder_Execute_Get(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("page = %s, want 1", r.URL.Query().Get("page"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodGet, "/")
	builder.Query("page", "1")
	builder.Header("X-Custom", "value")

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestBuilder_Execute_Post(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodPost, "/")
	builder.Body(map[string]string{"name": "test"})
	builder.ContentType("application/json")

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestRequestBuilder_Execute_Put(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodPut, "/")
	builder.Body(map[string]string{"name": "test"})

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestBuilder_Execute_Patch(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodPatch, "/")
	builder.Body(map[string]string{"name": "test"})

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestBuilder_Execute_Delete(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodDelete, "/")

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestRequestBuilder_Execute_Head(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("method = %s, want HEAD", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodHead, "/")

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestBuilder_Execute_Options(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			t.Errorf("method = %s, want OPTIONS", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodOptions, "/")

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestBuilder_Execute_UnsupportedMethod(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:8080")
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, "INVALID", "/")

	_, err := builder.Execute(client)
	if err == nil {
		t.Error("Execute() should error for unsupported method")
	}
}

func TestRequestBuilder_AuthToken(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mytoken" {
			t.Errorf("Authorization = %s, want Bearer mytoken", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodGet, "/")
	builder.AuthToken("mytoken")

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestBuilder_Timeout(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	ctx := context.Background()

	builder := NewRequestBuilder(ctx, http.MethodGet, "/")
	builder.Timeout(5 * time.Second)

	resp, err := builder.Execute(client)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
