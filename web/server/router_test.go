package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/web/mvc"
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
	router.GET("/hello", func(ctx mvc.Context) {
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
	router.GET("/users/{id}", func(ctx mvc.Context) {
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
	router.GET("/x", func(ctx mvc.Context) {
		ctx.String(http.StatusOK, "get")
	})
	router.handle(http.MethodHead, "/x", func(ctx mvc.Context) {
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
	sibling.GET("/", func(ctx mvc.Context) {
		handlerCalled = true
	})

	// /apix/ must NOT be claimed by the /api group (shared handlers map collision)
	req := httptest.NewRequest(http.MethodGet, "/apix/", nil)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /apix/ status = %d, want %d", w.Code, http.StatusNotFound)
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
	api.GET("/x", func(ctx mvc.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	w := httptest.NewRecorder()
	api.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("handler should be called for /api/x")
	}
	if w.Code != http.StatusOK {
		t.Errorf("GET /api/x status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRouter_DuplicateRegistration_Rejected(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	firstCalled := false
	secondCalled := false

	router.GET("/dup", func(ctx mvc.Context) {
		firstCalled = true
	})
	router.GET("/dup", func(ctx mvc.Context) {
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

	router.GET("/items/{id}", func(ctx mvc.Context) {
		firstCalled = true
	})
	router.GET("/items/{id}", func(ctx mvc.Context) {
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

	router.GET("/test", func(ctx mvc.Context) {
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

	router.POST("/test", func(ctx mvc.Context) {
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

	router.PUT("/test", func(ctx mvc.Context) {
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

	router.DELETE("/test", func(ctx mvc.Context) {
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

	router.PATCH("/test", func(ctx mvc.Context) {
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

	router.GET("/test", func(ctx mvc.Context) {})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRouter_NotFound(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	router.GET("/test", func(ctx mvc.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRouter_Group(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	api := router.Group("/api")
	api.GET("/users", func(ctx mvc.Context) {
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
	api.GET("/users", func(ctx mvc.Context) {
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

	router.Use(func(ctx mvc.Context) {
		middlewareCalled = true
		ctx.Next()
	})

	router.GET("/test", func(ctx mvc.Context) {})

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

	router.Use(func(ctx mvc.Context) {
		executionOrder = append(executionOrder, "mw1")
		ctx.Next()
	})

	router.Use(func(ctx mvc.Context) {
		executionOrder = append(executionOrder, "mw2")
		ctx.Next()
	})

	router.GET("/test", func(ctx mvc.Context) {
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

	router.Use(func(ctx mvc.Context) {
		middlewareCalled = true
		ctx.Next()
	})

	router.GET("/users/{id}", func(ctx mvc.Context) {
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

func TestRouter_PathParams(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	var capturedID string

	router.GET("/users/{id}", func(ctx mvc.Context) {
		capturedID = ctx.PathParam("id")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if capturedID != "123" {
		t.Errorf("PathParam(id) = %s, want 123", capturedID)
	}
}

func TestRouter_MatchPath(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"/users/{id}", "/users/123", true},
		{"/users/{id}", "/users/456", true},
		{"/users/{id}", "/users", false},
		{"/users/{id}/posts", "/users/123/posts", true},
		{"/users", "/users", true},
		{"/users", "/admins", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.path, func(t *testing.T) {
			got := router.matchPath(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("matchPath(%s, %s) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestRouter_ExtractParams(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	router.GET("/users/{id}/posts/{postId}", func(ctx mvc.Context) {})

	// 通过 ServeHTTP 触发路由匹配和参数提取
	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证参数提取（通过路由内部处理）
	// 这里我们直接测试 extractParamsForPattern
	params := router.extractParamsForPattern("/users/{id}/posts/{postId}", "/users/123/posts/456")

	if params["id"] != "123" {
		t.Errorf("params[id] = %s, want 123", params["id"])
	}
	if params["postId"] != "456" {
		t.Errorf("params[postId] = %s, want 456", params["postId"])
	}
}

func TestRouter_GroupInheritsMiddleware(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	middlewareCalled := false

	router.Use(func(ctx mvc.Context) {
		middlewareCalled = true
		ctx.Next()
	})

	api := router.Group("/api")
	api.GET("/test", func(ctx mvc.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Group should inherit parent middleware")
	}
}

func TestRouter_GroupMiddleware(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	executionOrder := []string{}

	router.Use(func(ctx mvc.Context) {
		executionOrder = append(executionOrder, "root")
		ctx.Next()
	})

	// 创建组 - 它在创建时继承根中间件
	api := router.Group("/api")

	// 在创建后向组添加中间件
	api.Use(func(ctx mvc.Context) {
		executionOrder = append(executionOrder, "api")
		ctx.Next()
	})

	api.GET("/test", func(ctx mvc.Context) {
		executionOrder = append(executionOrder, "handler")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 组中间件在 handle 时绑定到路由，应随处理链执行
	expected := []string{"root", "api", "handler"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("executionOrder length = %d, want %d", len(executionOrder), len(expected))
	}
	for i, v := range expected {
		if executionOrder[i] != v {
			t.Errorf("executionOrder[%d] = %s, want %s", i, executionOrder[i], v)
		}
	}
}
