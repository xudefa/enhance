package mvc

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/web/core"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host '0.0.0.0', got %s", cfg.Host)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.Logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestDefaultWebConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultWebConfig()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
}

func TestWebConfig_Alias(t *testing.T) {
	t.Parallel()
	var cfg WebConfig = DefaultConfig()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
}

func TestWithConfig(t *testing.T) {
	t.Parallel()
	cfg := Config{Port: 9090, Host: "localhost"}
	starter := NewWebStarter(WithConfig(cfg))
	if starter.config.Port != 9090 {
		t.Errorf("expected port 9090, got %d", starter.config.Port)
	}
	if starter.config.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %s", starter.config.Host)
	}
}

func TestWithName(t *testing.T) {
	t.Parallel()
	starter := NewWebStarter(WithName("test-web"))
	if starter.name != "test-web" {
		t.Errorf("expected name 'test-web', got %s", starter.name)
	}
}

func TestWithLogger(t *testing.T) {
	t.Parallel()
	logger := log.Build()
	starter := NewWebStarter(WithLogger(logger))
	if starter.logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestWithServer(t *testing.T) {
	t.Parallel()
	mockServer := &mockServer{}
	starter := NewWebStarter(WithServer(mockServer))
	if starter.server == nil {
		t.Error("expected server to be set")
	}
}

func TestWithRouter(t *testing.T) {
	t.Parallel()
	mockRouter := &mockRouter{}
	starter := NewWebStarter(WithRouter(mockRouter))
	if starter.router == nil {
		t.Error("expected router to be set")
	}
}

func TestWithMiddlewares(t *testing.T) {
	t.Parallel()
	middleware := func(ctx core.Context) {
		ctx.Next()
	}
	starter := NewWebStarter(WithMiddlewares([]core.MiddlewareFunc{middleware}))
	if len(starter.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(starter.middlewares))
	}
}

func TestWithHandler(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	starter := NewWebStarter(WithHandler(handler))
	if starter.handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestNewWebStarter_Defaults(t *testing.T) {
	t.Parallel()
	starter := NewWebStarter()
	if starter.config.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", starter.config.Port)
	}
	if starter.name != "web" {
		t.Errorf("expected default name 'web', got %s", starter.name)
	}
	if starter.logger == nil {
		t.Error("expected default logger to be set")
	}
}

// Mock implementations

type mockServer struct{}

func (m *mockServer) Start() error                                   { return nil }
func (m *mockServer) Stop(ctx context.Context) error                 { return nil }
func (m *mockServer) SetHandler(handler http.Handler)                {}
func (m *mockServer) Use(middleware func(http.Handler) http.Handler) {}

type mockRouter struct{}

func (m *mockRouter) GET(path string, handler core.HandlerFunc)            {}
func (m *mockRouter) POST(path string, handler core.HandlerFunc)           {}
func (m *mockRouter) PUT(path string, handler core.HandlerFunc)            {}
func (m *mockRouter) DELETE(path string, handler core.HandlerFunc)         {}
func (m *mockRouter) PATCH(path string, handler core.HandlerFunc)          {}
func (m *mockRouter) Group(path string) core.Router                        { return m }
func (m *mockRouter) Use(middleware core.MiddlewareFunc)                   {}
func (m *mockRouter) Handle(method, path string, handler core.HandlerFunc) {}
