package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestRequestScope_Basic(t *testing.T) {
	t.Parallel()

	scope := NewRequestScope()
	if scope == nil {
		t.Fatal("NewRequestScope() returned nil")
	}
}

func TestRequestScope_Get(t *testing.T) {
	t.Parallel()

	scope := NewRequestScope()

	// First call should use factory
	val1 := scope.Get("user", func() any {
		return "user1"
	})
	if val1 != "user1" {
		t.Errorf("expected 'user1', got %v", val1)
	}

	// Second call should return cached value
	val2 := scope.Get("user", func() any {
		return "user2"
	})
	if val2 != "user1" {
		t.Errorf("expected cached value 'user1', got %v", val2)
	}
}

func TestRequestScope_Set(t *testing.T) {
	t.Parallel()

	scope := NewRequestScope()
	scope.Set("user", "testuser")

	val := scope.Get("user", func() any {
		return "factory_user"
	})
	if val != "testuser" {
		t.Errorf("expected 'testuser', got %v", val)
	}
}

func TestRequestScope_Clear(t *testing.T) {
	t.Parallel()

	scope := NewRequestScope()
	scope.Set("user", "testuser")
	scope.Clear()

	val := scope.Get("user", func() any {
		return "new_user"
	})
	if val != "new_user" {
		t.Errorf("expected 'new_user' after clear, got %v", val)
	}
}

func TestRequestScope_Concurrent(t *testing.T) {
	t.Parallel()

	scope := NewRequestScope()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key"
			scope.Get(key, func() any {
				return id
			})
		}(i)
	}

	wg.Wait()
}

func TestRequestScopeMiddleware(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		if scope == nil {
			http.Error(w, "scope not found", http.StatusInternalServerError)
			return
		}

		user := scope.Get("user", func() any {
			return "testuser"
		})
		w.Write([]byte(user.(string)))
	})

	wrapped := RequestScopeMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != "testuser" {
		t.Errorf("expected 'testuser', got %s", rr.Body.String())
	}
}

func TestGetRequestScope_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	scope := GetRequestScope(ctx)
	if scope != nil {
		t.Error("expected nil scope for empty context")
	}
}

func TestMustGetRequestScope(t *testing.T) {
	t.Parallel()

	t.Run("with scope", func(t *testing.T) {
		t.Parallel()

		scope := NewRequestScope()
		ctx := context.WithValue(context.Background(), ScopeContextKey{}, scope)

		result := MustGetRequestScope(ctx)
		if result != scope {
			t.Error("expected same scope")
		}
	})

	t.Run("without scope", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for missing scope")
			}
		}()

		ctx := context.Background()
		MustGetRequestScope(ctx)
	})
}

func TestRequestScopeMiddlewareFunc(t *testing.T) {
	t.Parallel()

	middleware := RequestScopeMiddlewareFunc()
	if middleware == nil {
		t.Fatal("RequestScopeMiddlewareFunc() returned nil")
	}

	// Test with mock context
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

// mockCoreContext 模拟core.Context接口
type mockCoreContext struct {
	ctx     context.Context
	next    func()
	aborted bool
}

func (m *mockCoreContext) RequestMethod() string {
	return "GET"
}

func (m *mockCoreContext) RequestURI() string {
	return "/test"
}

func (m *mockCoreContext) PathParam(name string) string {
	return ""
}

func (m *mockCoreContext) Query(name string) string {
	return ""
}

func (m *mockCoreContext) QueryDefault(name, defaultVal string) string {
	return defaultVal
}

func (m *mockCoreContext) Header(key string) string {
	return ""
}

func (m *mockCoreContext) BindJSON(target any) error {
	return nil
}

func (m *mockCoreContext) SetStatusCode(code int) {}

func (m *mockCoreContext) SetHeader(key, value string) {}

func (m *mockCoreContext) JSON(code int, data any) error {
	return nil
}

func (m *mockCoreContext) String(code int, format string, args ...any) {}

func (m *mockCoreContext) AbortWithStatus(code int) {
	m.aborted = true
}

func (m *mockCoreContext) AbortWithStatusJSON(code int, body any) {
	m.aborted = true
}

func (m *mockCoreContext) Next() {
	if m.next != nil {
		m.next()
	}
}

func (m *mockCoreContext) IsAborted() bool {
	return m.aborted
}

func (m *mockCoreContext) Context() context.Context {
	return m.ctx
}

func (m *mockCoreContext) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *mockCoreContext) Request() *http.Request {
	return nil
}

func (m *mockCoreContext) ResponseWriter() http.ResponseWriter {
	return nil
}

func (m *mockCoreContext) Get(key string) any {
	return nil
}

func (m *mockCoreContext) Set(key string, value any) {}
