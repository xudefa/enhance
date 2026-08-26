package engine

import (
	"context"
	"net/http"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

func TestServerAdapter_Start(t *testing.T) {
	t.Parallel()

	started := false
	startFunc := func() error {
		started = true
		return nil
	}
	stopFunc := func(ctx context.Context) error { return nil }
	setHandlerFunc := func(handler http.Handler) {}
	useFunc := func(middleware func(http.Handler) http.Handler) {}

	adapter := NewServerAdapter(startFunc, stopFunc, setHandlerFunc, useFunc)

	err := adapter.Start()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !started {
		t.Error("expected startFunc to be called")
	}
}

func TestServerAdapter_Stop(t *testing.T) {
	t.Parallel()

	stopped := false
	startFunc := func() error { return nil }
	stopFunc := func(ctx context.Context) error {
		stopped = true
		return nil
	}
	setHandlerFunc := func(handler http.Handler) {}
	useFunc := func(middleware func(http.Handler) http.Handler) {}

	adapter := NewServerAdapter(startFunc, stopFunc, setHandlerFunc, useFunc)

	ctx := context.Background()
	err := adapter.Stop(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !stopped {
		t.Error("expected stopFunc to be called")
	}
}

func TestServerAdapter_SetHandler(t *testing.T) {
	t.Parallel()

	var receivedHandler http.Handler
	startFunc := func() error { return nil }
	stopFunc := func(ctx context.Context) error { return nil }
	setHandlerFunc := func(handler http.Handler) {
		receivedHandler = handler
	}
	useFunc := func(middleware func(http.Handler) http.Handler) {}

	adapter := NewServerAdapter(startFunc, stopFunc, setHandlerFunc, useFunc)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	adapter.SetHandler(handler)

	if receivedHandler == nil {
		t.Error("expected handler to be set")
	}
}

func TestServerAdapter_Use(t *testing.T) {
	t.Parallel()

	var receivedMiddleware func(http.Handler) http.Handler
	startFunc := func() error { return nil }
	stopFunc := func(ctx context.Context) error { return nil }
	setHandlerFunc := func(handler http.Handler) {}
	useFunc := func(middleware func(http.Handler) http.Handler) {
		receivedMiddleware = middleware
	}

	adapter := NewServerAdapter(startFunc, stopFunc, setHandlerFunc, useFunc)

	middleware := func(next http.Handler) http.Handler {
		return next
	}
	adapter.Use(middleware)

	if receivedMiddleware == nil {
		t.Error("expected middleware to be registered")
	}
}

func TestRouterAdapter_GET(t *testing.T) {
	t.Parallel()

	var receivedPath string
	var receivedHandler core.HandlerFunc
	getFunc := func(path string, handler core.HandlerFunc) {
		receivedPath = path
		receivedHandler = handler
	}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	handleFunc := func(method, path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	handler := func(ctx core.Context) {}
	adapter.GET("/test", handler)

	if receivedPath != "/test" {
		t.Errorf("expected path /test, got %s", receivedPath)
	}
	if receivedHandler == nil {
		t.Error("expected handler to be set")
	}
}

func TestRouterAdapter_POST(t *testing.T) {
	t.Parallel()

	var receivedPath string
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {
		receivedPath = path
	}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	handleFunc := func(method, path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	handler := func(ctx core.Context) {}
	adapter.POST("/test", handler)

	if receivedPath != "/test" {
		t.Errorf("expected path /test, got %s", receivedPath)
	}
}

func TestRouterAdapter_PUT(t *testing.T) {
	t.Parallel()

	var receivedPath string
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {
		receivedPath = path
	}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	handleFunc := func(method, path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	handler := func(ctx core.Context) {}
	adapter.PUT("/test", handler)

	if receivedPath != "/test" {
		t.Errorf("expected path /test, got %s", receivedPath)
	}
}

func TestRouterAdapter_DELETE(t *testing.T) {
	t.Parallel()

	var receivedPath string
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {
		receivedPath = path
	}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	handleFunc := func(method, path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	handler := func(ctx core.Context) {}
	adapter.DELETE("/test", handler)

	if receivedPath != "/test" {
		t.Errorf("expected path /test, got %s", receivedPath)
	}
}

func TestRouterAdapter_PATCH(t *testing.T) {
	t.Parallel()

	var receivedPath string
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {
		receivedPath = path
	}
	handleFunc := func(method, path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	handler := func(ctx core.Context) {}
	adapter.PATCH("/test", handler)

	if receivedPath != "/test" {
		t.Errorf("expected path /test, got %s", receivedPath)
	}
}

func TestRouterAdapter_Handle(t *testing.T) {
	t.Parallel()

	var receivedMethod string
	var receivedPath string
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	handleFunc := func(method, path string, handler core.HandlerFunc) {
		receivedMethod = method
		receivedPath = path
	}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	handler := func(ctx core.Context) {}
	adapter.Handle("OPTIONS", "/test", handler)

	if receivedMethod != "OPTIONS" {
		t.Errorf("expected method OPTIONS, got %s", receivedMethod)
	}
	if receivedPath != "/test" {
		t.Errorf("expected path /test, got %s", receivedPath)
	}
}

func TestRouterAdapter_Handle_Nil(t *testing.T) {
	t.Parallel()

	// 测试 handleFunc 为 nil 的情况
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, nil, groupFunc, useFunc)

	handler := func(ctx core.Context) {}
	// 不应该 panic
	adapter.Handle("GET", "/test", handler)
}

func TestRouterAdapter_Group(t *testing.T) {
	t.Parallel()

	var receivedPrefix string
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	handleFunc := func(method, path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router {
		receivedPrefix = prefix
		return nil
	}
	useFunc := func(middleware core.MiddlewareFunc) {}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	adapter.Group("/api")

	if receivedPrefix != "/api" {
		t.Errorf("expected prefix /api, got %s", receivedPrefix)
	}
}

func TestRouterAdapter_Use(t *testing.T) {
	t.Parallel()

	var receivedMiddleware core.MiddlewareFunc
	getFunc := func(path string, handler core.HandlerFunc) {}
	postFunc := func(path string, handler core.HandlerFunc) {}
	putFunc := func(path string, handler core.HandlerFunc) {}
	deleteFunc := func(path string, handler core.HandlerFunc) {}
	patchFunc := func(path string, handler core.HandlerFunc) {}
	handleFunc := func(method, path string, handler core.HandlerFunc) {}
	groupFunc := func(prefix string) core.Router { return nil }
	useFunc := func(middleware core.MiddlewareFunc) {
		receivedMiddleware = middleware
	}

	adapter := NewRouterAdapter(getFunc, postFunc, putFunc, deleteFunc, patchFunc, handleFunc, groupFunc, useFunc)

	middleware := func(ctx core.Context) {
		ctx.Next()
	}
	adapter.Use(middleware)

	if receivedMiddleware == nil {
		t.Error("expected middleware to be registered")
	}
}
