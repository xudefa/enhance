package server

import (
	"context"
	"testing"
)

func TestRequestScopeMiddlewareFunc_Basic(t *testing.T) {
	t.Parallel()

	middleware := RequestScopeMiddlewareFunc()
	if middleware == nil {
		t.Fatal("RequestScopeMiddlewareFunc() returned nil")
	}

	mockCtx := &mockCoreContext{
		ctx: context.Background(),
	}

	var scope *RequestScope
	mockCtx.next = func() {
		scope = GetRequestScope(mockCtx.ctx)
	}

	middleware(mockCtx)

	if scope == nil {
		t.Error("expected scope to be set")
	}
}

func TestRequestScopeMiddlewareFunc_ClearOnFinish(t *testing.T) {
	t.Parallel()

	middleware := RequestScopeMiddlewareFunc()
	mockCtx := &mockCoreContext{
		ctx: context.Background(),
	}

	var capturedScope *RequestScope
	mockCtx.next = func() {
		capturedScope = GetRequestScope(mockCtx.ctx)
		capturedScope.Set("user", "alice")
	}

	middleware(mockCtx)

	if len(capturedScope.cache) != 0 {
		t.Errorf("cache should be cleared after request, got %d items", len(capturedScope.cache))
	}
}

func TestRequestScopeMiddlewareFunc_MultipleCalls(t *testing.T) {
	t.Parallel()

	middleware := RequestScopeMiddlewareFunc()
	var scopes []*RequestScope

	for i := 0; i < 3; i++ {
		mockCtx := &mockCoreContext{
			ctx: context.Background(),
		}
		mockCtx.next = func() {
			scope := GetRequestScope(mockCtx.ctx)
			scopes = append(scopes, scope)
		}
		middleware(mockCtx)
	}

	if len(scopes) != 3 {
		t.Fatalf("expected 3 scopes, got %d", len(scopes))
	}

	for i := 0; i < len(scopes); i++ {
		for j := i + 1; j < len(scopes); j++ {
			if scopes[i] == scopes[j] {
				t.Errorf("scope[%d] and scope[%d] should be different instances", i, j)
			}
		}
	}
}

func TestRequestScopeMiddlewareFunc_PanicRecovery(t *testing.T) {
	t.Parallel()

	middleware := RequestScopeMiddlewareFunc()
	var capturedScope *RequestScope

	mockCtx := &mockCoreContext{
		ctx: context.Background(),
	}
	mockCtx.next = func() {
		capturedScope = GetRequestScope(mockCtx.ctx)
		capturedScope.Set("user", "alice")
		panic("handler panic")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
		if len(capturedScope.cache) != 0 {
			t.Errorf("cache should be cleared after panic, got %d items", len(capturedScope.cache))
		}
	}()

	middleware(mockCtx)
}

func TestRequestScopeMiddlewareFunc_ContextOperations(t *testing.T) {
	t.Parallel()

	middleware := RequestScopeMiddlewareFunc()
	var capturedScope *RequestScope

	mockCtx := &mockCoreContext{
		ctx: context.Background(),
	}
	mockCtx.next = func() {
		capturedScope = GetRequestScope(mockCtx.ctx)
		capturedScope.Set("key1", "value1")
		capturedScope.Set("key2", 123)
	}

	middleware(mockCtx)

	if capturedScope == nil {
		t.Fatal("expected scope to be set")
	}
	if len(capturedScope.cache) != 0 {
		t.Errorf("cache should be cleared, got %d items", len(capturedScope.cache))
	}
}

func TestRequestScopeMiddlewareFunc_NilNext(t *testing.T) {
	t.Parallel()

	middleware := RequestScopeMiddlewareFunc()
	mockCtx := &mockCoreContext{
		ctx:  context.Background(),
		next: nil,
	}

	middleware(mockCtx)
}
