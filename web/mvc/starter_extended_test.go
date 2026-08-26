package mvc

import (
	"net/http"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

func TestWebStarter_Setters(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()

	// Test SetRouter
	mockRouter := &mockRouter{}
	starter.SetRouter(mockRouter)
	if starter.router != mockRouter {
		t.Error("expected router to be set")
	}

	// Test SetServer
	mockServer := &mockServer{}
	starter.SetServer(mockServer)
	if starter.server != mockServer {
		t.Error("expected server to be set")
	}

	// Test SetHandler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	starter.SetHandler(handler)
	if starter.handler == nil {
		t.Error("expected handler to be set")
	}

	// Test SetMiddlewares
	middlewares := []core.MiddlewareFunc{
		func(ctx core.Context) {},
	}
	starter.SetMiddlewares(middlewares)
	if len(starter.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(starter.middlewares))
	}

	// Test AddMiddleware
	starter.AddMiddleware(func(ctx core.Context) {})
	if len(starter.middlewares) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(starter.middlewares))
	}
}

func TestWebStarter_Name(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	if starter.Name() == "" {
		t.Error("expected non-empty name")
	}
}

func TestWebStarter_Dependencies(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	deps := starter.Dependencies()
	if deps != nil {
		t.Error("expected nil dependencies")
	}
}

func TestWebStarter_Configure_NilContext(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	err := starter.Configure(nil)
	if err != nil {
		t.Errorf("expected no error for nil context, got %v", err)
	}
}

func TestWebStarter_GetRouter(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	mockRouter := &mockRouter{}
	starter.router = mockRouter

	router := starter.GetRouter()
	if router != mockRouter {
		t.Error("expected router to match")
	}
}

func TestWebStarter_Use(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	middleware := func(ctx core.Context) {}

	result := starter.Use(middleware)
	if result != starter {
		t.Error("expected Use to return starter for chaining")
	}
	if len(starter.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(starter.middlewares))
	}
}

func TestWebStarter_WithServer(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	mockServer := &mockServer{}

	result := starter.WithServer(mockServer)
	if result != starter {
		t.Error("expected WithServer to return starter for chaining")
	}
	if starter.server != mockServer {
		t.Error("expected server to be set")
	}
}

func TestWebStarter_WithRouter(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	mockRouter := &mockRouter{}

	result := starter.WithRouter(mockRouter)
	if result != starter {
		t.Error("expected WithRouter to return starter for chaining")
	}
	if starter.router != mockRouter {
		t.Error("expected router to be set")
	}
}

func TestWebStarter_WithMiddlewares(t *testing.T) {
	t.Parallel()

	middlewares := []core.MiddlewareFunc{
		func(ctx core.Context) {},
	}
	starter := NewWebStarter(WithMiddlewares(middlewares))

	if len(starter.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(starter.middlewares))
	}
}

func TestWebStarter_WithHandler(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	starter := NewWebStarter(WithHandler(handler))

	if starter.handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestWebStarter_GetCondition(t *testing.T) {
	t.Parallel()

	starter := NewWebStarter()
	condition := starter.GetCondition()
	if condition != nil {
		t.Errorf("expected nil condition, got %v", condition)
	}
}
