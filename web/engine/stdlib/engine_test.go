package stdlib

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

func TestNewContext(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	if ctx == nil {
		t.Fatal("expected context to be created")
	}
}

func TestContext_RequestMethod(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	if ctx.RequestMethod() != "POST" {
		t.Errorf("expected POST, got %s", ctx.RequestMethod())
	}
}

func TestContext_RequestURI(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	if ctx.RequestURI() != "/test?param=value" {
		t.Errorf("expected /test?param=value, got %s", ctx.RequestURI())
	}
}

func TestContext_PathParam(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	ctx.WithParams(map[string]string{"id": "123"})

	if ctx.PathParam("id") != "123" {
		t.Errorf("expected 123, got %s", ctx.PathParam("id"))
	}

	if ctx.PathParam("unknown") != "" {
		t.Errorf("expected empty string, got %s", ctx.PathParam("unknown"))
	}
}

func TestContext_Query(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	if ctx.Query("name") != "John" {
		t.Errorf("expected John, got %s", ctx.Query("name"))
	}

	if ctx.Query("age") != "30" {
		t.Errorf("expected 30, got %s", ctx.Query("age"))
	}

	if ctx.Query("unknown") != "" {
		t.Errorf("expected empty string, got %s", ctx.Query("unknown"))
	}
}

func TestContext_QueryDefault(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test?name=John", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	if ctx.QueryDefault("name", "Default") != "John" {
		t.Errorf("expected John, got %s", ctx.QueryDefault("name", "Default"))
	}

	if ctx.QueryDefault("unknown", "Default") != "Default" {
		t.Errorf("expected Default, got %s", ctx.QueryDefault("unknown", "Default"))
	}
}

func TestContext_Header(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	if ctx.Header("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", ctx.Header("Content-Type"))
	}

	if ctx.Header("Unknown") != "" {
		t.Errorf("expected empty string, got %s", ctx.Header("Unknown"))
	}
}

func TestContext_BindJSON(t *testing.T) {
	t.Parallel()
	type testData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body := `{"name":"John","age":30}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	var data testData
	err := ctx.BindJSON(&data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Name != "John" {
		t.Errorf("expected John, got %s", data.Name)
	}

	if data.Age != 30 {
		t.Errorf("expected 30, got %d", data.Age)
	}
}

func TestContext_SetStatusCode(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	ctx.SetStatusCode(201)

	if rec.Code != 201 {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestContext_SetHeader(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	ctx.SetHeader("X-Custom", "value")

	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("expected value, got %s", rec.Header().Get("X-Custom"))
	}
}

func TestContext_JSON(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	err := ctx.JSON(200, map[string]string{"message": "success"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected application/json, got %s", contentType)
	}
}

func TestContext_String(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	ctx.String(200, "Hello, %s!", "World")

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	if rec.Body.String() != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %s", rec.Body.String())
	}
}

func TestContext_AbortWithStatus(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	ctx.AbortWithStatus(403)

	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
}

func TestContext_AbortWithStatusJSON(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	ctx.AbortWithStatusJSON(500, map[string]string{"error": "Internal Server Error"})

	if rec.Code != 500 {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
}

func TestContext_Next(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	called := false
	ctx.WithMiddleware([]core.MiddlewareFunc{
		func(c core.Context) {
			called = true
			c.Next()
		},
	}, func(c core.Context) {
		called = true
	})

	ctx.Next()

	if !called {
		t.Error("expected middleware to be called")
	}
}

func TestContext_Context(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	if ctx.Context() == nil {
		t.Error("expected context to be set")
	}

	newCtx := context.WithValue(req.Context(), "key", "value")
	ctx.SetContext(newCtx)

	if ctx.Context().Value("key") != "value" {
		t.Error("expected context value to be set")
	}
}

func TestNewRouter(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	if router == nil {
		t.Fatal("expected router to be created")
	}
}

func TestRouter_GET(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	handlerCalled := false
	router.GET("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestRouter_POST(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	handlerCalled := false
	router.POST("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestRouter_PUT(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	handlerCalled := false
	router.PUT("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestRouter_DELETE(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	handlerCalled := false
	router.DELETE("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestRouter_PATCH(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	handlerCalled := false
	router.PATCH("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestRouter_Group(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	group := router.Group("/api")

	handlerCalled := false
	group.GET("/users", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
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

	handlerCalled := false
	router.GET("/test", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestRouter_NotFound(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	router.GET("/test", func(ctx core.Context) {})

	req := httptest.NewRequest("GET", "/unknown", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	router.GET("/test", func(ctx core.Context) {})

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestNewServer(t *testing.T) {
	t.Parallel()
	server := NewServer()

	if server == nil {
		t.Fatal("expected server to be created")
	}
}

func TestServer_SetHandler(t *testing.T) {
	t.Parallel()
	server := NewServer()

	router := NewRouter()
	server.SetHandler(router)

	if server.handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestServer_Use(t *testing.T) {
	t.Parallel()
	server := NewServer()

	server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})

	if len(server.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(server.middlewares))
	}
}

func TestFactory_Type(t *testing.T) {
	t.Parallel()
	factory := &Factory{}

	if factory.Type() != "stdlib" {
		t.Errorf("expected stdlib, got %s", factory.Type())
	}
}

func TestFactory_CreateRouter(t *testing.T) {
	t.Parallel()
	factory := &Factory{}

	router, err := factory.CreateRouter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if router == nil {
		t.Error("expected router to be created")
	}
}

func TestFactory_CreateServer(t *testing.T) {
	t.Parallel()
	factory := &Factory{}

	server, err := factory.CreateServer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server == nil {
		t.Error("expected server to be created")
	}
}

func BenchmarkContext_Query(b *testing.B) {
	req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.Query("name")
	}
}

func BenchmarkContext_PathParam(b *testing.B) {
	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()

	ctx := NewContext(rec, req)
	ctx.WithParams(map[string]string{"id": "123"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.PathParam("id")
	}
}

func BenchmarkRouter_HandleRequest(b *testing.B) {
	router := NewRouter()

	router.GET("/test", func(ctx core.Context) {})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_WithMiddleware(b *testing.B) {
	router := NewRouter()

	router.Use(func(ctx core.Context) {
		ctx.Next()
	})

	router.GET("/test", func(ctx core.Context) {})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rec, req)
	}
}
