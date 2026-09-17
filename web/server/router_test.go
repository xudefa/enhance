package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	if router == nil || router.handlers == nil {
		t.Fatalf("NewRouter() failed: router=%v, handlers=%v", router != nil, router != nil)
	}
}

func TestRouter_HEAD_FallsBackToGET(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	router.GET("/hello", func(ctx core.Context) {
		ctx.String(http.StatusOK, "hello")
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Head(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("HEAD request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("HEAD status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) != 0 {
		t.Errorf("HEAD response body should be empty, got %q", body)
	}
}

func TestRouter_HEAD_FallsBackToGET_ParamRoute(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	router.GET("/users/{id}", func(ctx core.Context) {
		ctx.String(http.StatusOK, ctx.PathParam("id"))
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Head(ts.URL + "/users/42")
	if err != nil {
		t.Fatalf("HEAD request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("HEAD status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRouter_HEAD_ExplicitRouteTakesPrecedence(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	router.GET("/x", func(ctx core.Context) {
		ctx.String(http.StatusOK, "get")
	})
	router.handle(http.MethodHead, "/x", func(ctx core.Context) {
		ctx.SetStatusCode(http.StatusNoContent)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Head(ts.URL + "/x")
	if err != nil {
		t.Fatalf("HEAD request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("HEAD status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestRouter_GroupPrefix_RespectsSegmentBoundary(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	api := router.Group("/api").(*DefaultRouter)
	sibling := router.Group("/apix")
	sibling.GET("/", func(ctx core.Context) {
		handlerCalled = true
	})

	// /apix/ must NOT be claimed by the /api group (shared handlers map collision)
	req := httptest.NewRequest(http.MethodGet, "/apix/", nil)
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /apix/ status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if handlerCalled {
		t.Error("handler must not be called for a path outside the /api boundary")
	}
}

func TestRouter_GroupPrefix_SegmentBoundaryMatch(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	api := router.Group("/api").(*DefaultRouter)
	api.GET("/x", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("handler should be called for /api/x")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/x status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRouter_DuplicateRegistration_Rejected(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	firstCalled := false
	secondCalled := false

	router.GET("/dup", func(ctx core.Context) {
		firstCalled = true
	})
	router.GET("/dup", func(ctx core.Context) {
		secondCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/dup", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !firstCalled {
		t.Error("first handler should be used")
	}
	if secondCalled {
		t.Error("second (duplicate) handler must not override the first")
	}
}

func TestRouter_DuplicateParamRegistration_Rejected(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	firstCalled := false
	secondCalled := false

	router.GET("/items/{id}", func(ctx core.Context) {
		firstCalled = true
	})
	router.GET("/items/{id}", func(ctx core.Context) {
		secondCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/items/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !firstCalled {
		t.Error("first handler should be used")
	}
	if secondCalled {
		t.Error("second (duplicate) handler must not override the first")
	}
}

func TestRouter_GET(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	router.GET("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("GET handler was not called")
	}
}

func TestRouter_POST(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	router.POST("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("POST handler was not called")
	}
}

func TestRouter_PUT(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	router.PUT("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("PUT handler was not called")
	}
}

func TestRouter_DELETE(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	router.DELETE("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("DELETE handler was not called")
	}
}

func TestRouter_PATCH(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	router.PATCH("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodPatch, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("PATCH handler was not called")
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	router.GET("/test", func(ctx core.Context) {})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRouter_NotFound(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	router.GET("/test", func(ctx core.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRouter_Group(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	api := router.Group("/api")
	api.GET("/users", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Group handler was not called")
	}
}

func TestRouter_Group_Prefix(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	v1 := router.Group("/v1")
	api := v1.Group("/api")
	api.GET("/users", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Nested group handler was not called")
	}
}

func TestRouter_Use(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	middlewareCalled := false

	router.Use(func(ctx core.Context) {
		middlewareCalled = true
		ctx.Next()
	})

	router.GET("/test", func(ctx core.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}
}

func TestRouter_MiddlewareChain(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	executionOrder := []string{}

	router.Use(func(ctx core.Context) {
		executionOrder = append(executionOrder, "mw1")
		ctx.Next()
	})

	router.Use(func(ctx core.Context) {
		executionOrder = append(executionOrder, "mw2")
		ctx.Next()
	})

	router.GET("/test", func(ctx core.Context) {
		executionOrder = append(executionOrder, "handler")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	expected := []string{"mw1", "mw2", "handler"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("executionOrder length = %d, want %d", len(executionOrder), len(expected))
	}
	for i, v := range expected {
		if executionOrder[i] != v {
			t.Errorf("executionOrder[%d] = %s, want %s", i, executionOrder[i], v)
		}
	}
}

func TestRouter_ParamRouteMiddleware(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	middlewareCalled := false
	var capturedID string

	router.Use(func(ctx core.Context) {
		middlewareCalled = true
		ctx.Next()
	})

	router.GET("/users/{id}", func(ctx core.Context) {
		capturedID = ctx.PathParam("id")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Middleware was not called for parameterized route")
	}
	if capturedID != "123" {
		t.Errorf("PathParam = %q, want %q", capturedID, "123")
	}
}
