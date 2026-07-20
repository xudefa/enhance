package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/web/mvc"
)

// BenchmarkRouter_ServeHTTP_StaticRoute 测试静态路由匹配性能
func BenchmarkRouter_ServeHTTP_StaticRoute(b *testing.B) {
	router := NewRouter()
	router.GET("/api/users", func(ctx mvc.Context) {
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_ServeHTTP_PathParams 测试路径参数提取性能
func BenchmarkRouter_ServeHTTP_PathParams(b *testing.B) {
	router := NewRouter()
	router.GET("/api/users/{id}/posts/{postId}", func(ctx mvc.Context) {
		_ = ctx.PathParam("id")
		_ = ctx.PathParam("postId")
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_ServeHTTP_MultipleRoutes 测试多路由匹配性能
func BenchmarkRouter_ServeHTTP_MultipleRoutes(b *testing.B) {
	router := NewRouter()

	// 注册 50 个路由
	for i := 0; i < 50; i++ {
		path := "/api/resource" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		router.GET(path, func(ctx mvc.Context) {
			ctx.JSON(200, map[string]string{"status": "ok"})
		})
	}

	// 添加带参数的路由
	router.GET("/api/users/{id}", func(ctx mvc.Context) {
		_ = ctx.PathParam("id")
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_RegisterRoute 测试路由注册性能
func BenchmarkRouter_RegisterRoute(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router := NewRouter()
		router.GET("/api/test", func(ctx mvc.Context) {})
	}
}

// BenchmarkRouter_Group 测试路由组创建性能
func BenchmarkRouter_Group(b *testing.B) {
	router := NewRouter()
	router.Use(func(ctx mvc.Context) { ctx.Next() })

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = router.Group("/api/v1")
	}
}

// BenchmarkRouter_MiddlewareChain 测试中间件链性能
func BenchmarkRouter_MiddlewareChain(b *testing.B) {
	router := NewRouter()

	// 添加 10 个中间件
	for i := 0; i < 10; i++ {
		router.Use(func(ctx mvc.Context) { ctx.Next() })
	}

	router.GET("/api/test", func(ctx mvc.Context) {
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_ConcurrentRequests 测试并发请求性能
func BenchmarkRouter_ConcurrentRequests(b *testing.B) {
	router := NewRouter()
	router.GET("/api/users/{id}", func(ctx mvc.Context) {
		_ = ctx.PathParam("id")
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
		w := httptest.NewRecorder()

		for pb.Next() {
			w.Body.Reset()
			router.ServeHTTP(w, req)
		}
	})
}
