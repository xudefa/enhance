package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

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
				fetched := scope.Get("shared", func() any {
					return id
				})
				mu.Lock()
				results = append(results, fetched)
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

func TestRequestScopeMiddleware_GetWithNilFactory(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := GetRequestScope(r.Context())
		fetched := scope.Get("nil_key", func() any {
			return nil
		})
		if fetched != nil {
			t.Errorf("expected nil, got %v", fetched)
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
