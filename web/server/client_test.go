package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:8080")
	if client == nil || client.baseURL != "http://localhost:8080" {
		t.Fatalf("NewClient() failed: client=%v, baseURL=%v", client != nil, client != nil)
	}
}

func TestNewClient_Options(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:8080",
		WithClientTimeout(5*time.Second),
	)

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", client.httpClient.Timeout)
	}
}

func TestClient_Get(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Post(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Post(context.Background(), "/", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestClient_Put(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Put(context.Background(), "/", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Delete(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Delete(context.Background(), "/")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestClient_Patch(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Patch(context.Background(), "/", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("Patch() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Head(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Head(context.Background(), "/")
	if err != nil {
		t.Fatalf("Head() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Options(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Options(context.Background(), "/")
	if err != nil {
		t.Fatalf("Options() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Timeout(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithClientTimeout(50*time.Millisecond))
	_, err := client.Get(context.Background(), "/")
	if err == nil {
		t.Error("Get() expected timeout error, got nil")
	}
}

func TestClient_QueryParams(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("page = %s, want 1", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("size") != "10" {
			t.Errorf("size = %s, want 10", r.URL.Query().Get("size"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Get(context.Background(), "/", WithQuery("page", "1"), WithQuery("size", "10"))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_Headers(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("Authorization = %s, want Bearer token", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Get(context.Background(), "/", WithAuthToken("token"))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestClient_WithMiddleware(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	middlewareCalled := false
	client := NewClient(ts.URL)
	client.WithMiddleware(func(req *http.Request, resp *HTTPResponse) error {
		middlewareCalled = true
		return nil
	})

	resp, err := client.Get(context.Background(), "/")
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

func TestClient_Close(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:8080")
	if err := client.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestRequestOption_WithAuthToken(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mytoken" {
			t.Errorf("Authorization = %s, want Bearer mytoken", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Get(context.Background(), "/", WithAuthToken("mytoken"))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestOption_WithContentType(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/xml" {
			t.Errorf("Content-Type = %s, want application/xml", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Post(context.Background(), "/", nil, WithContentType("application/xml"))
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestOption_WithTimeout(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	resp, err := client.Get(context.Background(), "/", WithTimeout(5*time.Second))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHTTPResponse_IsSuccess(t *testing.T) {
	t.Parallel()
	tests := []struct {
		statusCode int
		want       bool
	}{
		{http.StatusOK, true},
		{http.StatusCreated, true},
		{http.StatusNoContent, true},
		{http.StatusBadRequest, false},
		{http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		resp := &HTTPResponse{StatusCode: tt.statusCode}
		if got := resp.IsSuccess(); got != tt.want {
			t.Errorf("IsSuccess(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestHTTPResponse_IsServerError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		statusCode int
		want       bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		resp := &HTTPResponse{StatusCode: tt.statusCode}
		if got := resp.IsServerError(); got != tt.want {
			t.Errorf("IsServerError(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestHTTPResponse_IsClientError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		statusCode int
		want       bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, true},
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, true},
		{http.StatusNotFound, true},
		{http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		resp := &HTTPResponse{StatusCode: tt.statusCode}
		if got := resp.IsClientError(); got != tt.want {
			t.Errorf("IsClientError(%d) = %v, want %v", tt.statusCode, got, tt.want)
		}
	}
}

func TestHTTPResponse_Bind(t *testing.T) {
	t.Parallel()
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	resp := &HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"name":"alice","age":30}`),
	}

	var user User
	if err := resp.Bind(&user); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if user.Name != "alice" {
		t.Errorf("Name = %s, want alice", user.Name)
	}
	if user.Age != 30 {
		t.Errorf("Age = %d, want 30", user.Age)
	}
}

func TestHTTPResponse_Bind_InvalidBody(t *testing.T) {
	t.Parallel()
	resp := &HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       []byte(`invalid json`),
	}

	var result map[string]string
	if err := resp.Bind(&result); err == nil {
		t.Error("Bind() expected error for invalid JSON, got nil")
	}
}

func TestHTTPResponse_String(t *testing.T) {
	t.Parallel()
	resp := &HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("hello world"),
	}

	if got := resp.String(); got != "hello world" {
		t.Errorf("String() = %s, want hello world", got)
	}
}

func TestDo(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/test", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}

	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestDo_WithBody(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		if body["name"] != "test" {
			t.Errorf("name = %s, want test", body["name"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+"/test", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader([]byte(`{"name":"test"}`)))

	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestBuildURL(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:8080")

	tests := []struct {
		path string
		want string
	}{
		{"/api/users", "http://localhost:8080/api/users"},
		{"api/users", "http://localhost:8080/api/users"},
		{"http://example.com/test", "http://example.com/test"},
	}

	for _, tt := range tests {
		if got := client.buildURL(tt.path, nil); got != tt.want {
			t.Errorf("buildURL(%s) = %s, want %s", tt.path, got, tt.want)
		}
	}
}

func TestBuildURL_WithQuery(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:8080")
	query := url.Values{}
	query.Set("page", "1")

	got := client.buildURL("/api/users", query)
	if !strings.Contains(got, "page=1") {
		t.Errorf("buildURL() = %s, should contain page=1", got)
	}
}
