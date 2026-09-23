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
		fetched := scope.Get("user", func() any { return "default" })
		if fetched != "alice" {
			t.Errorf("expected alice, got %v", fetched)
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
		if panicValue := recover(); panicValue == nil {
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
		contextValue := r.Context().Value(ctxKey{})
		if contextValue != originalValue {
			t.Errorf("context value = %v, want %s", contextValue, originalValue)
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
		payload := &complexData{
			ID:   1,
			Name: "test",
			Tags: []string{"a", "b", "c"},
		}
		scope.Set("complex", payload)

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
