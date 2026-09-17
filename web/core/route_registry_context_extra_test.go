package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ==================== RegisterToMux duplicate routes ====================

func TestSimpleContext_Next(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.Next()

	if ctx.IsAborted() {
		t.Error("expected Next() to not abort")
	}
}

func TestSimpleContext_QueryDefault(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test?name=John", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	defaultedValue := ctx.QueryDefault("name", "default")
	if defaultedValue != "John" {
		t.Errorf("expected 'John', got %s", defaultedValue)
	}

	defaultedValue = ctx.QueryDefault("missing", "default")
	if defaultedValue != "default" {
		t.Errorf("expected 'default', got %s", defaultedValue)
	}
}

func TestSimpleContext_AbortWithStatus(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.AbortWithStatus(http.StatusUnauthorized)

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestSimpleContext_AbortWithStatus_NoContent(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.AbortWithStatus(http.StatusNoContent)

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestSimpleContext_AbortWithStatusJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}

	var resultBody map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resultBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resultBody["error"] != "invalid request" {
		t.Errorf("expected error message 'invalid request', got %q", resultBody["error"])
	}
}

func TestSimpleContext_String(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.String(http.StatusOK, "Hello, %s!", "World")

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type to start with 'text/plain', got %q", ct)
	}

	body := rec.Body.String()
	if body != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", body)
	}
}

func TestSimpleContext_SetStatusCode(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.SetStatusCode(http.StatusCreated)

	if ctx.statusCode != http.StatusCreated {
		t.Errorf("expected status code 201, got %d", ctx.statusCode)
	}
}

func TestSimpleContext_SetHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.SetHeader("X-Custom-Header", "custom-value")

	header := rec.Header().Get("X-Custom-Header")
	if header != "custom-value" {
		t.Errorf("expected 'custom-value', got %q", header)
	}
}

func TestSimpleContext_SetContext(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	newCtx := context.WithValue(req.Context(), "key", "value")
	ctx.SetContext(newCtx)

	if ctx.Context() != newCtx {
		t.Error("expected context to be updated")
	}
}

func TestSimpleContext_RequestMethod(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	if ctx.RequestMethod() != "POST" {
		t.Errorf("expected 'POST', got %s", ctx.RequestMethod())
	}
}

func TestSimpleContext_RequestURI(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test?query=value", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	uri := ctx.RequestURI()
	if uri != "/test?query=value" {
		t.Errorf("expected '/test?query=value', got %s", uri)
	}
}

func TestSimpleContext_PathParam(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	paramValue := ctx.PathParam("id")
	if paramValue != "" {
		t.Errorf("expected empty string, got %s", paramValue)
	}
}

func TestSimpleContext_Query(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	name := ctx.Query("name")
	if name != "John" {
		t.Errorf("expected 'John', got %s", name)
	}

	age := ctx.Query("age")
	if age != "30" {
		t.Errorf("expected '30', got %s", age)
	}

	missing := ctx.Query("missing")
	if missing != "" {
		t.Errorf("expected empty string, got %s", missing)
	}
}

func TestSimpleContext_Header(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	auth := ctx.Header("Authorization")
	if auth != "Bearer token123" {
		t.Errorf("expected 'Bearer token123', got %s", auth)
	}

	missing := ctx.Header("X-Missing")
	if missing != "" {
		t.Errorf("expected empty string, got %s", missing)
	}
}

func TestSimpleContext_Request(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	if ctx.Request() != req {
		t.Error("expected request to match")
	}
}
