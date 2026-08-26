package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/xudefa/enhance/web/core"
)

func TestHTTPResponse_IsRedirect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		statusCode int
		want       bool
	}{
		{http.StatusMovedPermanently, true},
		{http.StatusFound, true},
		{http.StatusNotModified, true},
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusInternalServerError, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			t.Parallel()
			resp := &HTTPResponse{StatusCode: tt.statusCode}
			if got := resp.IsRedirect(); got != tt.want {
				t.Errorf("IsRedirect(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_Bind_EmptyBody(t *testing.T) {
	t.Parallel()
	resp := &HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       nil,
	}
	var result map[string]string
	if err := resp.Bind(&result); err != nil {
		t.Errorf("Bind() with empty body should not error, got %v", err)
	}
}

func TestHTTPResponse_Unmarshal_EmptyBody(t *testing.T) {
	t.Parallel()
	resp := &HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       nil,
	}
	var result map[string]string
	if err := resp.Unmarshal(&result); err != nil {
		t.Errorf("Unmarshal() with empty body should not error, got %v", err)
	}
}

func TestHTTPResponse_Unmarshal_ValidJSON(t *testing.T) {
	t.Parallel()
	type user struct {
		Name string `json:"name"`
	}
	resp := &HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"name":"alice"}`),
	}
	var u user
	if err := resp.Unmarshal(&u); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if u.Name != "alice" {
		t.Errorf("Name = %s, want alice", u.Name)
	}
}

func TestContext_Request(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	got := ctx.Request()
	if got != req {
		t.Error("Request() should return the original request")
	}
}

func TestContext_JSON_MarshalFailure(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	err := ctx.JSON(http.StatusOK, make(chan int))
	if err == nil {
		t.Error("JSON() should error for non-marshalable data")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestContext_BindJSON_ReadError(t *testing.T) {
	t.Parallel()
	body := strings.NewReader(strings.Repeat("x", 100))
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	req.Body = http.MaxBytesReader(w, io.NopCloser(body), 1)

	var result map[string]string
	err := ctx.BindJSON(&result)
	if err == nil {
		t.Error("BindJSON() should error when body exceeds limit")
	}
}

func TestContext_Next_NoMiddlewareNoHandler(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)
	ctx.Next()
}

func TestContext_Next_AbortInSecondMiddleware(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	var order []string
	mw1 := func(c core.Context) {
		order = append(order, "mw1")
		c.Next()
	}
	mw2 := func(c core.Context) {
		order = append(order, "mw2")
		c.AbortWithStatus(http.StatusForbidden)
	}
	handler := func(c core.Context) {
		order = append(order, "handler")
	}

	ctx.WithMiddleware([]core.MiddlewareFunc{mw1, mw2}, handler)
	ctx.Next()

	if len(order) != 2 || order[0] != "mw1" || order[1] != "mw2" {
		t.Errorf("order = %v, want [mw1 mw2]", order)
	}
}

func TestHttpServerAdapter_GetServer(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(
		WithHost(":0"),
	)
	if adapter == nil {
		t.Fatal("NewHttpServerAdapter() returned nil")
	}
	if adapter.GetServer() == nil {
		t.Error("GetServer() returned nil")
	}
}

func TestHttpServerAdapter_SetHandler(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	adapter.SetHandler(handler)
	if adapter.GetServer().handler == nil {
		t.Error("SetHandler() should set the handler")
	}
}

func TestHttpServerAdapter_Use(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter()
	mw := func(next http.Handler) http.Handler {
		return next
	}
	adapter.Use(mw)
	if len(adapter.GetServer().middlewares) != 1 {
		t.Errorf("Use() should add middleware, got %d", len(adapter.GetServer().middlewares))
	}
}

func TestHttpServerAdapter_Stop(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(WithHost(":0"))
	err := adapter.Stop(context.Background())
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestHttpServerAdapter_StartAndStop(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(WithHost("127.0.0.1:0"))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	adapter.SetHandler(handler)

	errCh := make(chan error, 1)
	go func() {
		errCh <- adapter.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := adapter.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	select {
	case startErr := <-errCh:
		if startErr != nil && startErr != http.ErrServerClosed {
			t.Errorf("Start() error = %v", startErr)
		}
	case <-time.After(3 * time.Second):
		t.Error("Start() did not return")
	}
}
