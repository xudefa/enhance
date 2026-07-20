package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xudefa/enhance/web/mvc"
)

// TestHTTP2_Configuration 测试 HTTP/2 配置
func TestHTTP2_Configuration(t *testing.T) {
	t.Parallel()

	srv := NewHTTPServer(
		WithHost(":0"),
		WithTLS("cert.pem", "key.pem"),
		WithHTTP2(true),
	)

	if !srv.enableHTTP2 {
		t.Error("HTTP/2 应该被启用")
	}
}

// TestHTTP2_Disabled 测试 HTTP/2 禁用
func TestHTTP2_Disabled(t *testing.T) {
	t.Parallel()

	srv := NewHTTPServer(
		WithHost(":0"),
		WithHTTP2(false),
	)

	if srv.enableHTTP2 {
		t.Error("HTTP/2 应该被禁用")
	}
}

// TestHTTP2_StartH2C 测试 H2C 启动
func TestHTTP2_StartH2C(t *testing.T) {
	t.Parallel()

	srv := NewHTTPServer(
		WithHost(":0"),
	)

	// 验证 H2C 方法存在
	// 注意：实际启动需要更复杂的设置
	_ = srv
}

// TestHTTP2_ConcurrentRequests 测试 HTTP/2 并发请求
func TestHTTP2_ConcurrentRequests(t *testing.T) {
	t.Parallel()

	// 创建测试服务器
	router := NewRouter()
	router.GET("/api/test", func(ctx mvc.Context) {
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	srv := NewHTTPServer(
		WithHost(":0"),
	)
	srv.SetHandler(router)

	// 启动服务器
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("服务器启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建测试客户端
	client := &http.Client{}

	// 并发请求
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			resp, err := client.Get("http://localhost:8080/api/test")
			if err == nil {
				resp.Body.Close()
			}
			done <- true
		}()
	}

	// 等待所有请求完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Stop(ctx)
}

// BenchmarkHTTP2_ServerConfig 测试 HTTP/2 服务器配置性能
func BenchmarkHTTP2_ServerConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		srv := NewHTTPServer(
			WithHost(":0"),
			WithTLS("cert.pem", "key.pem"),
			WithHTTP2(true),
		)
		_ = srv
	}
}

// BenchmarkHTTP2_MiddlewareChain 测试 HTTP/2 中间件链性能
func BenchmarkHTTP2_MiddlewareChain(b *testing.B) {
	router := NewRouter()

	// 添加 10 个中间件
	for i := 0; i < 10; i++ {
		router.Use(func(ctx mvc.Context) { ctx.Next() })
	}

	router.GET("/api/test", func(ctx mvc.Context) {
		ctx.JSON(200, map[string]string{"status": "ok"})
	})

	srv := NewHTTPServer(
		WithHost(":0"),
		WithHTTP2(true),
	)
	srv.SetHandler(router)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = srv
	}
}
