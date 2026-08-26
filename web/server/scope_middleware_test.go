package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestRequestScopeMiddleware_ScopeInContext(t *testing.T) {
	t.Parallel()

	var gotScope *RequestScope
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotScope = GetRequestScope(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if gotScope == nil {
		t.Fatal("RequestScope should be in context")
	}
}

func TestRequestScopeMiddleware_IsolatedPerRequest(t *testing.T) {
	t.Parallel()

	var scopes []*RequestScope
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scopes = append(scopes, scope)
		scope.Set("request_id", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
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

func TestRequestScopeMiddleware_ClearOnFinish(t *testing.T) {
	t.Parallel()

	var capturedScope *RequestScope
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedScope = GetRequestScope(r.Context())
		capturedScope.Set("user", "alice")
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if len(capturedScope.cache) != 0 {
		t.Errorf("cache should be cleared after request, got %d items", len(capturedScope.cache))
	}
}

func TestRequestScopeMiddleware_ScopeFunc(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("user", "alice")
		val := scope.Get("user", func() any { return "default" })
		if val != "alice" {
			t.Errorf("expected alice, got %v", val)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		wg.Add(1)
		go func() {
			defer wg.Done()
			scope.Set("key", "value")
			scope.Get("key", func() any { return "factory" })
		}()
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}

	wg.Wait()
}

func TestGetRequestScope_NilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetRequestScope(nil) should panic")
		}
	}()
	GetRequestScope(nil)
}

func TestGetRequestScope_Background(t *testing.T) {
	t.Parallel()
	scope := GetRequestScope(context.Background())
	if scope != nil {
		t.Error("expected nil for background context")
	}
}

func TestGetRequestScope_WithScope(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()
	ctx := context.WithValue(context.Background(), ScopeContextKey{}, scope)

	got := GetRequestScope(ctx)
	if got == nil {
		t.Fatal("expected scope from context")
	}
	if got != scope {
		t.Error("should return the same scope instance")
	}
}

func TestGetRequestScope_WrongType(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ScopeContextKey{}, "not a scope")

	got := GetRequestScope(ctx)
	if got != nil {
		t.Error("expected nil when context value has wrong type")
	}
}

func TestMustGetRequestScope_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetRequestScope should panic when scope is missing")
		}
	}()

	MustGetRequestScope(context.Background())
}

func TestMustGetRequestScope_Success(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()
	ctx := context.WithValue(context.Background(), ScopeContextKey{}, scope)
	got := MustGetRequestScope(ctx)
	if got != scope {
		t.Error("should return the same scope")
	}
}

func TestMustGetRequestScope_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetRequestScope should panic when context value has wrong type")
		}
	}()

	ctx := context.WithValue(context.Background(), ScopeContextKey{}, "not a scope")
	MustGetRequestScope(ctx)
}

func TestRequestScopeMiddleware_DifferentHTTPMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"POST", http.MethodPost},
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
		{"PATCH", http.MethodPatch},
		{"HEAD", http.MethodHead},
		{"OPTIONS", http.MethodOptions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotScope *RequestScope
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotScope = GetRequestScope(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequestScopeMiddleware(handler)
			req := httptest.NewRequest(tt.method, "/test", nil)
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rr.Code)
			}
			if gotScope == nil {
				t.Error("RequestScope should be in context")
			}
		})
	}
}

func TestRequestScopeMiddleware_MultipleSetAndGet(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("user_id", 123)
		scope.Set("user_name", "alice")
		scope.Set("is_admin", true)

		userID := scope.Get("user_id", func() any { return 0 })
		userName := scope.Get("user_name", func() any { return "" })
		isAdmin := scope.Get("is_admin", func() any { return false })

		if userID != 123 {
			t.Errorf("user_id = %v, want 123", userID)
		}
		if userName != "alice" {
			t.Errorf("user_name = %v, want alice", userName)
		}
		if isAdmin != true {
			t.Errorf("is_admin = %v, want true", isAdmin)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_HandlerPanic(t *testing.T) {
	t.Parallel()

	var capturedScope *RequestScope
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedScope = GetRequestScope(r.Context())
		capturedScope.Set("user", "alice")
		panic("handler panic")
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from handler")
		}
		if len(capturedScope.cache) != 0 {
			t.Errorf("cache should be cleared after panic, got %d items", len(capturedScope.cache))
		}
	}()

	middleware.ServeHTTP(rr, req)
}

func TestRequestScopeMiddleware_ResponseBody(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("user", "alice")
		user := scope.Get("user", func() any { return "default" })
		w.Write([]byte(user.(string)))
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if rr.Body.String() != "alice" {
		t.Errorf("body = %s, want alice", rr.Body.String())
	}
}

func TestRequestScopeMiddleware_ChainMultiple(t *testing.T) {
	t.Parallel()

	var innerScope *RequestScope
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerScope = GetRequestScope(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware1 := RequestScopeMiddleware(handler)
	middleware2 := RequestScopeMiddleware(middleware1)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware2.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if innerScope == nil {
		t.Error("handler should have received a scope")
	}
}

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

func TestRequestScopeMiddleware_RequestWithBody(t *testing.T) {
	t.Parallel()

	var gotScope *RequestScope
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotScope = GetRequestScope(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("POST", "/test", nil)
	req.Body = nil
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if gotScope == nil {
		t.Error("RequestScope should be in context")
	}
}

func TestRequestScopeMiddleware_DifferentURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{"root", "/"},
		{"simple", "/test"},
		{"nested", "/api/v1/users"},
		{"with_query", "/search?q=test&page=1"},
		{"with_fragment", "/page#section"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotScope *RequestScope
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotScope = GetRequestScope(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequestScopeMiddleware(handler)
			req := httptest.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rr.Code)
			}
			if gotScope == nil {
				t.Error("RequestScope should be in context")
			}
		})
	}
}

func TestRequestScopeMiddleware_ScopeDataIsolation(t *testing.T) {
	t.Parallel()

	var scope1, scope2 *RequestScope
	var data1, data2 string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		if r.URL.Path == "/req1" {
			scope1 = scope
			scope.Set("data", "value1")
			data1 = scope.Get("data", func() any { return "" }).(string)
		} else {
			scope2 = scope
			scope.Set("data", "value2")
			data2 = scope.Get("data", func() any { return "" }).(string)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)

	req1 := httptest.NewRequest("GET", "/req1", nil)
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req1)

	req2 := httptest.NewRequest("GET", "/req2", nil)
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req2)

	if scope1 == scope2 {
		t.Error("scopes should be different instances")
	}
	if data1 != "value1" {
		t.Errorf("data1 = %s, want value1", data1)
	}
	if data2 != "value2" {
		t.Errorf("data2 = %s, want value2", data2)
	}
}

func TestRequestScopeMiddleware_ContextValuePreserved(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}
	originalValue := "original"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(ctxKey{})
		if val != originalValue {
			t.Errorf("context value = %v, want %s", val, originalValue)
		}
		scope := GetRequestScope(r.Context())
		if scope == nil {
			t.Error("scope should be in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	ctx := context.WithValue(context.Background(), ctxKey{}, originalValue)
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
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

func TestRequestScopeMiddleware_ConcurrentStress(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				key := "key"
				scope.Get(key, func() any { return i })
				scope.Set("temp", i)
			}
		}()
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)

	for i := 0; i < 50; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}

	wg.Wait()
}

func TestRequestScopeMiddleware_NestedContextValues(t *testing.T) {
	t.Parallel()

	type complexData struct {
		ID   int
		Name string
		Tags []string
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		data := &complexData{
			ID:   1,
			Name: "test",
			Tags: []string{"a", "b", "c"},
		}
		scope.Set("complex", data)

		got := scope.Get("complex", func() any { return nil }).(*complexData)
		if got.ID != 1 || got.Name != "test" || len(got.Tags) != 3 {
			t.Errorf("complex data mismatch: %+v", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_EmptyHandler(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_OverwriteValue(t *testing.T) {
	t.Parallel()

	var capturedValue string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("key", "first")
		scope.Set("key", "second")
		capturedValue = scope.Get("key", func() any { return "default" }).(string)
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if capturedValue != "second" {
		t.Errorf("value = %s, want second", capturedValue)
	}
}

func TestRequestScopeMiddleware_FactoryNotCalledForExistingKey(t *testing.T) {
	t.Parallel()

	factoryCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("key", "existing")
		scope.Get("key", func() any {
			factoryCalled = true
			return "new"
		})
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if factoryCalled {
		t.Error("factory should not be called for existing key")
	}
}

func TestRequestScopeMiddleware_GetDoubleCheck(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		var results []any
		var mu sync.Mutex
		var start sync.WaitGroup
		start.Add(1)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				start.Wait()
				result := scope.Get("shared", func() any {
					return id
				})
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}(i)
		}
		start.Done()
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	wg.Wait()
}

func TestRequestScope_Get_DoubleCheckRace(t *testing.T) {
	t.Parallel()

	scope := NewRequestScope()
	var wg sync.WaitGroup
	var start sync.WaitGroup
	start.Add(1)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			start.Wait()
			result := scope.Get("key", func() any {
				return id
			})
			if result == nil {
				t.Errorf("result should not be nil")
			}
		}(i)
	}

	start.Done()
	wg.Wait()
}

func TestRequestScopeMiddleware_GetWithNilFactory(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		result := scope.Get("nil_key", func() any {
			return nil
		})
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_SetMultipleKeys(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		for i := 0; i < 100; i++ {
			scope.Set("key_"+string(rune(i)), i)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_GetAfterClear(t *testing.T) {
	t.Parallel()

	var scope *RequestScope
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope = GetRequestScope(r.Context())
		scope.Set("key", "value")
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}

	newVal := scope.Get("key", func() any { return "new_value" })
	if newVal != "new_value" {
		t.Errorf("expected new_value after clear, got %v", newVal)
	}
}

func TestRequestScopeMiddleware_ContextChain(t *testing.T) {
	t.Parallel()

	type ctxKey1 struct{}
	type ctxKey2 struct{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val1 := r.Context().Value(ctxKey1{})
		val2 := r.Context().Value(ctxKey2{})
		scope := GetRequestScope(r.Context())

		if val1 != "value1" {
			t.Errorf("ctxKey1 = %v, want value1", val1)
		}
		if val2 != "value2" {
			t.Errorf("ctxKey2 = %v, want value2", val2)
		}
		if scope == nil {
			t.Error("scope should be in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	ctx := context.WithValue(context.Background(), ctxKey1{}, "value1")
	ctx = context.WithValue(ctx, ctxKey2{}, "value2")
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
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

func TestRequestScopeMiddleware_MiddlewareWithResponse(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("user", "alice")
		user := scope.Get("user", func() any { return "" })
		w.Header().Set("X-User", user.(string))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if rr.Header().Get("X-User") != "alice" {
		t.Errorf("X-User header = %s, want alice", rr.Header().Get("X-User"))
	}
	if rr.Body.String() != "OK" {
		t.Errorf("body = %s, want OK", rr.Body.String())
	}
}

func TestRequestScopeMiddleware_HandlerWithError(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("error", "something went wrong")
		http.Error(w, "error", http.StatusInternalServerError)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestRequestScopeMiddleware_HandlerWithRedirect(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("redirect", true)
		http.Redirect(w, r, "/new-location", http.StatusFound)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", rr.Code)
	}
	if rr.Header().Get("Location") != "/new-location" {
		t.Errorf("Location = %s, want /new-location", rr.Header().Get("Location"))
	}
}

func TestRequestScopeMiddleware_NilHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil handler")
		}
	}()

	middleware := RequestScopeMiddleware(nil)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
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

func TestRequestScopeMiddleware_RequestHeaders(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("content_type", r.Header.Get("Content-Type"))
		scope.Set("auth", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_RequestWithQueryParams(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("query", r.URL.RawQuery)
		scope.Set("page", r.URL.Query().Get("page"))
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test?page=1&size=10", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_RequestWithHost(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("host", r.Host)
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRequestScopeMiddleware_RequestWithRemoteAddr(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		scope.Set("remote", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestScopeMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}
