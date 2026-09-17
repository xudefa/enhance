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
)

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

	var payload map[string]string
	if err := resp.Bind(&payload); err == nil {
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
