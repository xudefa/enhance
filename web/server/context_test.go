package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

func TestNewContext(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test?page=1", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}
	if ctx.RequestMethod() != http.MethodGet {
		t.Errorf("RequestMethod() = %s, want %s", ctx.RequestMethod(), http.MethodGet)
	}
	if ctx.RequestURI() != "/test?page=1" {
		t.Errorf("RequestURI() = %s, want /test?page=1", ctx.RequestURI())
	}
}

func TestContext_PathParam(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req).WithParams(map[string]string{"id": "123"})

	if got := ctx.PathParam("id"); got != "123" {
		t.Errorf("PathParam(id) = %s, want 123", got)
	}
	if got := ctx.PathParam("name"); got != "" {
		t.Errorf("PathParam(name) = %s, want empty string", got)
	}
}

func TestContext_Query(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test?name=alice&age=25", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	if got := ctx.Query("name"); got != "alice" {
		t.Errorf("Query(name) = %s, want alice", got)
	}
	if got := ctx.Query("missing"); got != "" {
		t.Errorf("Query(missing) = %s, want empty", got)
	}
}

func TestContext_QueryDefault(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test?name=alice", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	if got := ctx.QueryDefault("name", "default"); got != "alice" {
		t.Errorf("QueryDefault(name) = %s, want alice", got)
	}
	if got := ctx.QueryDefault("missing", "default"); got != "default" {
		t.Errorf("QueryDefault(missing) = %s, want default", got)
	}
}

func TestContext_Header(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "value")
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	if got := ctx.Header("Content-Type"); got != "application/json" {
		t.Errorf("Header(Content-Type) = %s, want application/json", got)
	}
	if got := ctx.Header("Missing"); got != "" {
		t.Errorf("Header(Missing) = %s, want empty", got)
	}
}

func TestContext_BindJSON(t *testing.T) {
	t.Parallel()
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body := User{Name: "alice", Age: 30}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	var got User
	if err := ctx.BindJSON(&got); err != nil {
		t.Fatalf("BindJSON() error = %v", err)
	}
	if got.Name != "alice" {
		t.Errorf("BindJSON().Name = %s, want alice", got.Name)
	}
	if got.Age != 30 {
		t.Errorf("BindJSON().Age = %d, want 30", got.Age)
	}
}

func TestContext_BindJSON_InvalidJSON(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	var result map[string]string
	if err := ctx.BindJSON(&result); err == nil {
		t.Error("BindJSON() expected error for invalid JSON, got nil")
	}
}

func TestContext_SetStatusCode(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)
	ctx.SetStatusCode(http.StatusCreated)

	if w.Code != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestContext_SetHeader(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)
	ctx.SetHeader("X-Custom", "test-value")

	if got := w.Header().Get("X-Custom"); got != "test-value" {
		t.Errorf("Header(X-Custom) = %s, want test-value", got)
	}
}

func TestContext_JSON(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)
	data := map[string]string{"message": "hello"}

	if err := ctx.JSON(http.StatusOK, data); err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", w.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}
	if result["message"] != "hello" {
		t.Errorf("response message = %s, want hello", result["message"])
	}
}

func TestContext_String(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)
	ctx.String(http.StatusOK, "Hello %s", "World")

	if w.Body.String() != "Hello World" {
		t.Errorf("Body = %s, want Hello World", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Content-Type = %s, want text/plain", w.Header().Get("Content-Type"))
	}
}

func TestContext_AbortWithStatus(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)
	ctx.AbortWithStatus(http.StatusUnauthorized)

	if !ctx.IsAborted() {
		t.Error("IsAborted() = false, want true")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestContext_AbortWithStatusJSON(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)
	ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"error": "bad request"})

	if !ctx.IsAborted() {
		t.Error("IsAborted() = false, want true")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", w.Header().Get("Content-Type"))
	}
}

func TestContext_Next(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	executionOrder := []string{}

	mw1 := func(ctx core.Context) {
		executionOrder = append(executionOrder, "mw1-before")
		ctx.Next()
		executionOrder = append(executionOrder, "mw1-after")
	}

	mw2 := func(ctx core.Context) {
		executionOrder = append(executionOrder, "mw2-before")
		ctx.Next()
		executionOrder = append(executionOrder, "mw2-after")
	}

	handler := func(ctx core.Context) {
		executionOrder = append(executionOrder, "handler")
	}

	ctx.WithMiddleware([]core.MiddlewareFunc{mw1, mw2}, handler)
	ctx.Next()

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("executionOrder length = %d, want %d", len(executionOrder), len(expected))
	}
	for i, v := range expected {
		if executionOrder[i] != v {
			t.Errorf("executionOrder[%d] = %s, want %s", i, executionOrder[i], v)
		}
	}
}

func TestContext_Next_Abort(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	executionOrder := []string{}

	mw1 := func(ctx core.Context) {
		executionOrder = append(executionOrder, "mw1")
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}

	mw2 := func(ctx core.Context) {
		executionOrder = append(executionOrder, "mw2")
	}

	handler := func(ctx core.Context) {
		executionOrder = append(executionOrder, "handler")
	}

	ctx.WithMiddleware([]core.MiddlewareFunc{mw1, mw2}, handler)
	ctx.Next()

	if len(executionOrder) != 1 || executionOrder[0] != "mw1" {
		t.Errorf("executionOrder = %v, want [mw1]", executionOrder)
	}
}

func TestContext_Context(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	if ctx.Context() != req.Context() {
		t.Error("Context() should return request.Context()")
	}
}

func TestContext_SetContext(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	type testKey struct{}
	newCtx := context.WithValue(req.Context(), testKey{}, "value")
	ctx.SetContext(newCtx)

	if ctx.Context().Value(testKey{}) != "value" {
		t.Error("SetContext() should update request context")
	}
}

func TestContext_WithMiddleware(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	handler := func(ctx core.Context) {}
	middlewares := []core.MiddlewareFunc{func(ctx core.Context) {}}

	result := ctx.WithMiddleware(middlewares, handler)

	if result != ctx {
		t.Error("WithMiddleware should return the context for chaining")
	}
}

func TestContext_WithParams(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ctx := NewContext(w, req)

	params := map[string]string{"id": "123"}
	result := ctx.WithParams(params)

	if result != ctx {
		t.Error("WithParams should return the context for chaining")
	}
	if ctx.PathParam("id") != "123" {
		t.Errorf("PathParam(id) = %s, want 123", ctx.PathParam("id"))
	}
}
