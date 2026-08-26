package server

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPServerBuilder_DefaultValues(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()

	if builder.host != ":8080" {
		t.Errorf("expected default host ':8080', got %s", builder.host)
	}
	if builder.readTimeout != 30*time.Second {
		t.Errorf("expected default read timeout 30s, got %v", builder.readTimeout)
	}
	if builder.writeTimeout != 30*time.Second {
		t.Errorf("expected default write timeout 30s, got %v", builder.writeTimeout)
	}
	if builder.idleTimeout != 120*time.Second {
		t.Errorf("expected default idle timeout 120s, got %v", builder.idleTimeout)
	}
}

func TestHTTPServerBuilder_Host(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	result := builder.Host(":9090")

	if result != builder {
		t.Error("expected builder to return itself for chaining")
	}
	if builder.host != ":9090" {
		t.Errorf("expected host ':9090', got %s", builder.host)
	}
}

func TestHTTPServerBuilder_ReadTimeout(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	result := builder.ReadTimeout(10 * time.Second)

	if result != builder {
		t.Error("expected builder to return itself for chaining")
	}
	if builder.readTimeout != 10*time.Second {
		t.Errorf("expected read timeout 10s, got %v", builder.readTimeout)
	}
}

func TestHTTPServerBuilder_WriteTimeout(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	result := builder.WriteTimeout(15 * time.Second)

	if result != builder {
		t.Error("expected builder to return itself for chaining")
	}
	if builder.writeTimeout != 15*time.Second {
		t.Errorf("expected write timeout 15s, got %v", builder.writeTimeout)
	}
}

func TestHTTPServerBuilder_IdleTimeout(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	result := builder.IdleTimeout(60 * time.Second)

	if result != builder {
		t.Error("expected builder to return itself for chaining")
	}
	if builder.idleTimeout != 60*time.Second {
		t.Errorf("expected idle timeout 60s, got %v", builder.idleTimeout)
	}
}

func TestHTTPServerBuilder_TLS(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	result := builder.TLS("/path/to/cert.pem", "/path/to/key.pem")

	if result != builder {
		t.Error("expected builder to return itself for chaining")
	}
	if builder.certFile != "/path/to/cert.pem" {
		t.Errorf("expected cert file '/path/to/cert.pem', got %s", builder.certFile)
	}
	if builder.keyFile != "/path/to/key.pem" {
		t.Errorf("expected key file '/path/to/key.pem', got %s", builder.keyFile)
	}
}

func TestHTTPServerBuilder_Middleware(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	result := builder.Middleware(middleware)

	if result != builder {
		t.Error("expected builder to return itself for chaining")
	}
	if len(builder.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(builder.middlewares))
	}
}

func TestHTTPServerBuilder_Handler(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	result := builder.Handler(handler)

	if result != builder {
		t.Error("expected builder to return itself for chaining")
	}
	if builder.handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestHTTPServerBuilder_Build_Default(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder()
	server, err := builder.Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestHTTPServerBuilder_Build_WithHandler(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})

	builder := NewHTTPServerBuilder().Handler(handler)
	server, err := builder.Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
	if server.handler == nil {
		t.Error("expected handler to be set on server")
	}
}

func TestHTTPServerBuilder_Build_WithMiddleware(t *testing.T) {
	t.Parallel()

	var middlewareCalled bool
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	builder := NewHTTPServerBuilder().
		Handler(handler).
		Middleware(middleware)

	server, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(server.middlewares) != 1 {
		t.Errorf("expected 1 middleware on server, got %d", len(server.middlewares))
	}
	_ = middlewareCalled // 中间件在构建时不会调用
}

func TestHTTPServerBuilder_Build_WithTLS(t *testing.T) {
	t.Parallel()

	builder := NewHTTPServerBuilder().
		TLS("/path/to/cert.pem", "/path/to/key.pem")

	server, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
	// TLS 配置会被设置，但不会验证文件是否存在
}

func TestHTTPServerBuilder_MustBuild(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	server := NewHTTPServerBuilder().MustBuild()
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestHTTPServerBuilder_Chaining(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server, err := NewHTTPServerBuilder().
		Host(":9090").
		ReadTimeout(10 * time.Second).
		WriteTimeout(15 * time.Second).
		IdleTimeout(60 * time.Second).
		Handler(handler).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}
