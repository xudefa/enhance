package core

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestSimpleContextRequestMethod(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"POST", http.MethodPost},
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(tt.method, "/test", "")
			if got := ctx.RequestMethod(); got != tt.method {
				t.Errorf("RequestMethod() = %q, want %q", got, tt.method)
			}
		})
	}
}

func TestSimpleContextRequestURI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{"simple path", "/users", "/users"},
		{"with query", "/users?page=1&size=10", "/users?page=1&size=10"},
		{"root path", "/", "/"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodGet, tt.uri, "")
			if got := ctx.RequestURI(); got != tt.want {
				t.Errorf("RequestURI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSimpleContextPathParam(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/users", "")
	if got := ctx.PathParam("id"); got != "" {
		t.Errorf("PathParam(\"id\") = %q, want empty", got)
	}
}

func TestSimpleContextQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		key  string
		want string
	}{
		{"existing param", "/test?key=value", "key", "value"},
		{"missing param", "/test", "key", ""},
		{"multiple params", "/test?a=1&b=2&c=3", "b", "2"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodGet, tt.url, "")
			if got := ctx.Query(tt.key); got != tt.want {
				t.Errorf("Query(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSimpleContextQueryDefault(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		url        string
		key        string
		defaultVal string
		want       string
	}{
		{"returns default when missing", "/test", "size", "10", "10"},
		{"returns value when present", "/test?size=20", "size", "10", "20"},
		{"returns default for empty value", "/test?size=", "size", "10", "10"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodGet, tt.url, "")
			if got := ctx.QueryDefault(tt.key, tt.defaultVal); got != tt.want {
				t.Errorf("QueryDefault(%q, %q) = %q, want %q", tt.key, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestSimpleContextHeader(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	ctx.req.Header.Set("X-Custom", "hello")
	ctx.req.Header.Set("Content-Type", "application/json")

	tests := []struct {
		key  string
		want string
	}{
		{"X-Custom", "hello"},
		{"Content-Type", "application/json"},
		{"Missing", ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			if got := ctx.Header(tt.key); got != tt.want {
				t.Errorf("Header(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

type bindJSONUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestSimpleContextBindJSON(t *testing.T) {
	t.Parallel()

	for _, tt := range testSimpleContextBindJSONCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodPost, "/test", tt.body)
			var got bindJSONUser
			err := ctx.BindJSON(&got)
			if (err != nil) != tt.wantErr {
				t.Errorf("BindJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("BindJSON() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("BindJSON() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func testSimpleContextBindJSONCases() []struct {
	name    string
	body    string
	wantErr bool
	errMsg  string
	want    bindJSONUser
} {
	return []struct {
		name    string
		body    string
		wantErr bool
		errMsg  string
		want    bindJSONUser
	}{
		{
			name:    "valid JSON",
			body:    `{"name":"Alice","age":30}`,
			wantErr: false,
			want:    bindJSONUser{Name: "Alice", Age: 30},
		},
		{
			name:    "empty body",
			body:    "",
			wantErr: true,
			errMsg:  "request body is empty",
		},
		{
			name:    "invalid JSON",
			body:    `{not json}`,
			wantErr: true,
		},
		{
			name:    "empty object",
			body:    `{}`,
			wantErr: false,
			want:    bindJSONUser{},
		},
	}
}

func TestSimpleContextJSON(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	payload := map[string]string{"message": "hello"}

	err := ctx.JSON(http.StatusOK, payload)
	if err != nil {
		t.Fatalf("JSON() unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["message"] != "hello" {
		t.Errorf("body message = %q, want %q", got["message"], "hello")
	}
}

func TestSimpleContextString(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.String(http.StatusOK, "hello %s %d", "world", 42)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
	}

	if got := rec.Body.String(); got != "hello world 42" {
		t.Errorf("body = %q, want %q", got, "hello world 42")
	}
}

func TestSimpleContextAbortWithStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		code    int
		wantMsg string
	}{
		{"normal error", http.StatusBadRequest, "Bad Request"},
		{"not found", http.StatusNotFound, "Not Found"},
		{"internal error", http.StatusInternalServerError, "Internal Server Error"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, rec := newTestContext(http.MethodGet, "/test", "")
			ctx.AbortWithStatus(tt.code)

			if !ctx.IsAborted() {
				t.Error("expected aborted = true")
			}
			if ctx.statusCode != tt.code {
				t.Errorf("statusCode = %d, want %d", ctx.statusCode, tt.code)
			}
			if rec.Code != tt.code {
				t.Errorf("response status = %d, want %d", rec.Code, tt.code)
			}
		})
	}
}

func TestSimpleContextAbortWithStatusNoContent(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.AbortWithStatus(http.StatusNoContent)

	if !ctx.IsAborted() {
		t.Error("expected aborted = true")
	}
	if ctx.statusCode != http.StatusNoContent {
		t.Errorf("statusCode = %d, want %d", ctx.statusCode, http.StatusNoContent)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("response status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestSimpleContextAbortWithStatusBelow200(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.AbortWithStatus(199)

	if !ctx.IsAborted() {
		t.Error("expected aborted = true")
	}
	if ctx.statusCode != 199 {
		t.Errorf("statusCode = %d, want 199", ctx.statusCode)
	}
	if rec.Code != 199 {
		t.Errorf("response status = %d, want 199", rec.Code)
	}
}

func TestSimpleContextAbortWithStatusJSON(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	body := map[string]string{"error": "unauthorized"}
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, body)

	if !ctx.IsAborted() {
		t.Error("expected aborted = true")
	}
	if ctx.statusCode != http.StatusUnauthorized {
		t.Errorf("statusCode = %d, want %d", ctx.statusCode, http.StatusUnauthorized)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("response status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if got["error"] != "unauthorized" {
		t.Errorf("body error = %q, want %q", got["error"], "unauthorized")
	}
}

func TestSimpleContextSetContext(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	if ctx.Context() == nil {
		t.Fatal("Context() should not be nil initially")
	}

	type ctxKey struct{}
	newCtx := context.WithValue(context.Background(), ctxKey{}, "value")
	ctx.SetContext(newCtx)

	if ctx.Context() != newCtx {
		t.Error("Context() should return the new context")
	}
	if ctx.Context().Value(ctxKey{}) != "value" {
		t.Error("context value should be preserved")
	}
}

func TestSimpleContextSetStatusCode(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	ctx.SetStatusCode(http.StatusTeapot)
	if ctx.statusCode != http.StatusTeapot {
		t.Errorf("statusCode = %d, want %d", ctx.statusCode, http.StatusTeapot)
	}
}

func TestSimpleContextSetHeader(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.SetHeader("X-Request-Id", "abc-123")
	ctx.SetHeader("X-Custom", "value")

	if rec.Header().Get("X-Request-Id") != "abc-123" {
		t.Errorf("X-Request-Id = %q, want %q", rec.Header().Get("X-Request-Id"), "abc-123")
	}
	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("X-Custom = %q, want %q", rec.Header().Get("X-Custom"), "value")
	}
}

func TestSimpleContextIsAborted(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	if ctx.IsAborted() {
		t.Error("expected IsAborted() = false initially")
	}
	ctx.AbortWithStatus(http.StatusBadRequest)
	if !ctx.IsAborted() {
		t.Error("expected IsAborted() = true after abort")
	}
}

func TestSimpleContextContext(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	got := ctx.Context()
	if got == nil {
		t.Fatal("Context() should not be nil")
	}
	if got != ctx.req.Context() {
		t.Error("Context() should return req's context")
	}
}

func TestSimpleContextRequest(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	got := ctx.Request()
	if got == nil {
		t.Fatal("Request() should not be nil")
	}
	if got != ctx.req {
		t.Error("Request() should return the underlying request")
	}
}

func TestSimpleContextNext(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	// Next() is a no-op, just verify it doesn't panic
	ctx.Next()
}
