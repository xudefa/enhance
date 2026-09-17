package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

func TestRouter_PathParams(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	var capturedID string

	router.GET("/users/{id}", func(ctx core.Context) {
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
	router.GET("/users/{id}/posts/{postId}", func(ctx core.Context) {})

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

	router.Use(func(ctx core.Context) {
		middlewareCalled = true
		ctx.Next()
	})

	api := router.Group("/api")
	api.GET("/test", func(ctx core.Context) {})

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

	router.Use(func(ctx core.Context) {
		executionOrder = append(executionOrder, "root")
		ctx.Next()
	})

	// 创建组 - 它在创建时继承根中间件
	api := router.Group("/api")

	// 在创建后向组添加中间件
	api.Use(func(ctx core.Context) {
		executionOrder = append(executionOrder, "api")
		ctx.Next()
	})

	api.GET("/test", func(ctx core.Context) {
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
