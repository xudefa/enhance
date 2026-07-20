package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xudefa/enhance/web/mvc"
)

func TestNewHTTPServer(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer()
	if server == nil || server.host != ":8080" || server.readTimeout != 30*time.Second || server.writeTimeout != 30*time.Second || server.idleTimeout != 120*time.Second {
		t.Fatalf("NewHTTPServer() failed: server=%v, host=%v", server != nil, server != nil)
	}
}

func TestNewHTTPServer_Options(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer(
		WithHost(":9090"),
		WithReadTimeout(10*time.Second),
		WithWriteTimeout(20*time.Second),
		WithIdleTimeout(60*time.Second),
	)

	if server.host != ":9090" {
		t.Errorf("host = %s, want :9090", server.host)
	}
	if server.readTimeout != 10*time.Second {
		t.Errorf("readTimeout = %v, want 10s", server.readTimeout)
	}
	if server.writeTimeout != 20*time.Second {
		t.Errorf("writeTimeout = %v, want 20s", server.writeTimeout)
	}
	if server.idleTimeout != 60*time.Second {
		t.Errorf("idleTimeout = %v, want 60s", server.idleTimeout)
	}
}

func TestHTTPServer_Use(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer()

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	server.Use(middleware)

	if len(server.middlewares) != 1 {
		t.Errorf("middlewares length = %d, want 1", len(server.middlewares))
	}
}

func TestHTTPServer_Use_PanicOnInvalidType(t *testing.T) {
	t.Parallel()
	// 由于 Use 现在接受 func(http.Handler) http.Handler 类型
	// 编译器会在编译时捕获类型错误，因此此测试不再需要
	// 保留此测试作为文档说明
	t.Skip("Type checking is now done at compile time")
}

func TestHTTPServer_SetHandler(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server.SetHandler(handler)

	if server.handler == nil {
		t.Error("SetHandler() should set the handler")
	}
}

func TestHTTPServer_Start_Stop(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer(WithHost(":0"))

	router := NewRouter()
	router.GET("/test", func(ctx mvc.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	server.SetHandler(router)

	// 在 goroutine 中启动服务器
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// 给服务器启动时间
	time.Sleep(100 * time.Millisecond)

	// 停止服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// 检查启动错误（应该为 nil 或服务器已关闭）
	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Start() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Stop() did not complete in time")
	}
}

func TestHTTPServer_WrapHTTPHandler(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer()

	middlewareCalled := false
	server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := server.wrapHTTPHandler(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("wrapHTTPHandler should call middleware")
	}
}

func TestHTTPServer_WrapRouter(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer()

	router := NewRouter()
	router.GET("/test", func(ctx mvc.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	wrapped := server.wrapRouter(router)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHTTPServer_WrapRouter_UnsupportedType(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer()

	// 创建一个模拟路由器，实现 mvc.Router 但不实现 http.Handler
	router := &mockRouterNoHTTPHandler{}

	wrapped := server.wrapRouter(router)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// mockRouterNoHTTPHandler implements mvc.Router but not http.Handler
type mockRouterNoHTTPHandler struct{}

func (m *mockRouterNoHTTPHandler) GET(path string, handler mvc.HandlerFunc)    {}
func (m *mockRouterNoHTTPHandler) POST(path string, handler mvc.HandlerFunc)   {}
func (m *mockRouterNoHTTPHandler) PUT(path string, handler mvc.HandlerFunc)    {}
func (m *mockRouterNoHTTPHandler) DELETE(path string, handler mvc.HandlerFunc) {}
func (m *mockRouterNoHTTPHandler) PATCH(path string, handler mvc.HandlerFunc)  {}
func (m *mockRouterNoHTTPHandler) Group(prefix string) mvc.Router              { return m }
func (m *mockRouterNoHTTPHandler) Use(middleware mvc.MiddlewareFunc)           {}

func TestHTTPServer_Start_UnsupportedHandler(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer(WithHost(":0"))

	// HttpServer.SetHandler 现在接受 http.Handler 类型
	// 传入无效类型会在编译时报错
	// 此测试仅验证 SetHandler 不会 panic
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server.SetHandler(handler)

	if server.handler == nil {
		t.Error("SetHandler() should set the handler")
	}
}
