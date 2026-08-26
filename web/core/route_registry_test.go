package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockController 模拟控制器。
type mockController struct{}

func (m *mockController) HandleGet() {
}

func (m *mockController) HandleWithContext(ctx Context) {
}

type mockControllerWithReturn struct{}

func (m *mockControllerWithReturn) GetData() map[string]string {
	return map[string]string{"message": "success"}
}

func (m *mockControllerWithReturn) GetError() (map[string]string, error) {
	return nil, fmt.Errorf("test error")
}

func TestNewRouteRegistry(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()

	if registry == nil {
		t.Fatal("expected registry to be created")
	}
}

func TestRegisterController(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("UserController", controller, "/users")
	registry.RegisterController("OrderController", controller, "/orders")

	if path := registry.GetControllerBasePath("UserController"); path != "/users" {
		t.Errorf("expected /users, got %s", path)
	}

	if path := registry.GetControllerBasePath("OrderController"); path != "/orders" {
		t.Errorf("expected /orders, got %s", path)
	}

	if path := registry.GetControllerBasePath("UnknownController"); path != "" {
		t.Errorf("expected empty path, got %s", path)
	}
}

func TestRegisterRoute(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()

	route := RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "GetUsers",
	}

	registry.RegisterRoute(route)

	routes := registry.GetRoutes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	if routes[0].Method != "GET" {
		t.Errorf("expected GET, got %s", routes[0].Method)
	}

	if routes[0].Path != "/users" {
		t.Errorf("expected /users, got %s", routes[0].Path)
	}
}

func TestGetRoutesWithController(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("UserController", controller, "/users")

	route := RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "HandleGet",
	}

	registry.RegisterRoute(route)

	routes := registry.GetRoutes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	if routes[0].Controller == nil {
		t.Fatal("expected controller to be resolved")
	}

	if _, ok := routes[0].Controller.(*mockController); !ok {
		t.Errorf("expected *mockController, got %T", routes[0].Controller)
	}
}

func TestRegisterToMux(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("UserController", controller, "/users")

	route := RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "HandleGet",
		Produces:   "application/json",
	}

	registry.RegisterRoute(route)

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterToMuxMissingController(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()

	route := RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "HandleGet",
	}

	registry.RegisterRoute(route)

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err == nil {
		t.Fatal("expected error for missing controller")
	}
}

func TestRegisterToMuxMissingMethod(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("UserController", controller, "/users")

	route := RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "NonExistentMethod",
	}

	registry.RegisterRoute(route)

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err == nil {
		t.Fatal("expected error for missing method")
	}
}

func TestCreateHandlerMethodNotAllowed(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("UserController", controller, "/users")

	route := RouteInfo{
		Method:     "POST",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "HandleGet",
	}

	registry.RegisterRoute(route)

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

// ==================== simpleContext tests ====================

func newTestContext(method, url string, body string) (*simpleContext, *httptest.ResponseRecorder) {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, url, bodyReader)
	rec := httptest.NewRecorder()
	return newSimpleContext(rec, req), rec
}

func TestSimpleContextRequestMethod(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"POST", http.MethodPost},
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(tt.method, "/test", "")
			if got := ctx.RequestMethod(); got != tt.method {
				t.Errorf("RequestMethod() = %q, want %q", got, tt.method)
			}
		})
	}
}

func TestSimpleContextRequestURI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{"simple path", "/users", "/users"},
		{"with query", "/users?page=1&size=10", "/users?page=1&size=10"},
		{"root path", "/", "/"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodGet, tt.uri, "")
			if got := ctx.RequestURI(); got != tt.want {
				t.Errorf("RequestURI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSimpleContextPathParam(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/users", "")
	if got := ctx.PathParam("id"); got != "" {
		t.Errorf("PathParam(\"id\") = %q, want empty", got)
	}
}

func TestSimpleContextQuery(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		key  string
		want string
	}{
		{"existing param", "/test?key=value", "key", "value"},
		{"missing param", "/test", "key", ""},
		{"multiple params", "/test?a=1&b=2&c=3", "b", "2"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodGet, tt.url, "")
			if got := ctx.Query(tt.key); got != tt.want {
				t.Errorf("Query(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSimpleContextQueryDefault(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		url        string
		key        string
		defaultVal string
		want       string
	}{
		{"returns default when missing", "/test", "size", "10", "10"},
		{"returns value when present", "/test?size=20", "size", "10", "20"},
		{"returns default for empty value", "/test?size=", "size", "10", "10"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodGet, tt.url, "")
			if got := ctx.QueryDefault(tt.key, tt.defaultVal); got != tt.want {
				t.Errorf("QueryDefault(%q, %q) = %q, want %q", tt.key, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestSimpleContextHeader(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	ctx.req.Header.Set("X-Custom", "hello")
	ctx.req.Header.Set("Content-Type", "application/json")

	tests := []struct {
		key  string
		want string
	}{
		{"X-Custom", "hello"},
		{"Content-Type", "application/json"},
		{"Missing", ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			if got := ctx.Header(tt.key); got != tt.want {
				t.Errorf("Header(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSimpleContextBindJSON(t *testing.T) {
	t.Parallel()
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		body    string
		wantErr bool
		errMsg  string
		want    user
	}{
		{
			name:    "valid JSON",
			body:    `{"name":"Alice","age":30}`,
			wantErr: false,
			want:    user{Name: "Alice", Age: 30},
		},
		{
			name:    "empty body",
			body:    "",
			wantErr: true,
			errMsg:  "request body is empty",
		},
		{
			name:    "invalid JSON",
			body:    `{not json}`,
			wantErr: true,
		},
		{
			name:    "empty object",
			body:    `{}`,
			wantErr: false,
			want:    user{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := newTestContext(http.MethodPost, "/test", tt.body)
			var got user
			err := ctx.BindJSON(&got)
			if (err != nil) != tt.wantErr {
				t.Errorf("BindJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("BindJSON() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("BindJSON() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSimpleContextJSON(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	data := map[string]string{"message": "hello"}

	err := ctx.JSON(http.StatusOK, data)
	if err != nil {
		t.Fatalf("JSON() unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["message"] != "hello" {
		t.Errorf("body message = %q, want %q", got["message"], "hello")
	}
}

func TestSimpleContextString(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.String(http.StatusOK, "hello %s %d", "world", 42)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
	}

	if got := rec.Body.String(); got != "hello world 42" {
		t.Errorf("body = %q, want %q", got, "hello world 42")
	}
}

func TestSimpleContextAbortWithStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		code    int
		wantMsg string
	}{
		{"normal error", http.StatusBadRequest, "Bad Request"},
		{"not found", http.StatusNotFound, "Not Found"},
		{"internal error", http.StatusInternalServerError, "Internal Server Error"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, rec := newTestContext(http.MethodGet, "/test", "")
			ctx.AbortWithStatus(tt.code)

			if !ctx.IsAborted() {
				t.Error("expected aborted = true")
			}
			if ctx.statusCode != tt.code {
				t.Errorf("statusCode = %d, want %d", ctx.statusCode, tt.code)
			}
			if rec.Code != tt.code {
				t.Errorf("response status = %d, want %d", rec.Code, tt.code)
			}
		})
	}
}

func TestSimpleContextAbortWithStatusNoContent(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.AbortWithStatus(http.StatusNoContent)

	if !ctx.IsAborted() {
		t.Error("expected aborted = true")
	}
	if ctx.statusCode != http.StatusNoContent {
		t.Errorf("statusCode = %d, want %d", ctx.statusCode, http.StatusNoContent)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("response status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestSimpleContextAbortWithStatusBelow200(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.AbortWithStatus(199)

	if !ctx.IsAborted() {
		t.Error("expected aborted = true")
	}
	if ctx.statusCode != 199 {
		t.Errorf("statusCode = %d, want 199", ctx.statusCode)
	}
	if rec.Code != 199 {
		t.Errorf("response status = %d, want 199", rec.Code)
	}
}

func TestSimpleContextAbortWithStatusJSON(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	body := map[string]string{"error": "unauthorized"}
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, body)

	if !ctx.IsAborted() {
		t.Error("expected aborted = true")
	}
	if ctx.statusCode != http.StatusUnauthorized {
		t.Errorf("statusCode = %d, want %d", ctx.statusCode, http.StatusUnauthorized)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("response status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if got["error"] != "unauthorized" {
		t.Errorf("body error = %q, want %q", got["error"], "unauthorized")
	}
}

func TestSimpleContextSetContext(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	if ctx.Context() == nil {
		t.Fatal("Context() should not be nil initially")
	}

	type ctxKey struct{}
	newCtx := context.WithValue(context.Background(), ctxKey{}, "value")
	ctx.SetContext(newCtx)

	if ctx.Context() != newCtx {
		t.Error("Context() should return the new context")
	}
	if ctx.Context().Value(ctxKey{}) != "value" {
		t.Error("context value should be preserved")
	}
}

func TestSimpleContextSetStatusCode(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	ctx.SetStatusCode(http.StatusTeapot)
	if ctx.statusCode != http.StatusTeapot {
		t.Errorf("statusCode = %d, want %d", ctx.statusCode, http.StatusTeapot)
	}
}

func TestSimpleContextSetHeader(t *testing.T) {
	t.Parallel()
	ctx, rec := newTestContext(http.MethodGet, "/test", "")
	ctx.SetHeader("X-Request-Id", "abc-123")
	ctx.SetHeader("X-Custom", "value")

	if rec.Header().Get("X-Request-Id") != "abc-123" {
		t.Errorf("X-Request-Id = %q, want %q", rec.Header().Get("X-Request-Id"), "abc-123")
	}
	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("X-Custom = %q, want %q", rec.Header().Get("X-Custom"), "value")
	}
}

func TestSimpleContextIsAborted(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	if ctx.IsAborted() {
		t.Error("expected IsAborted() = false initially")
	}
	ctx.AbortWithStatus(http.StatusBadRequest)
	if !ctx.IsAborted() {
		t.Error("expected IsAborted() = true after abort")
	}
}

func TestSimpleContextContext(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	got := ctx.Context()
	if got == nil {
		t.Fatal("Context() should not be nil")
	}
	if got != ctx.req.Context() {
		t.Error("Context() should return req's context")
	}
}

func TestSimpleContextRequest(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	got := ctx.Request()
	if got == nil {
		t.Fatal("Request() should not be nil")
	}
	if got != ctx.req {
		t.Error("Request() should return the underlying request")
	}
}

func TestSimpleContextNext(t *testing.T) {
	t.Parallel()
	ctx, _ := newTestContext(http.MethodGet, "/test", "")
	// Next() is a no-op, just verify it doesn't panic
	ctx.Next()
}

// ==================== RegisterToMux duplicate routes ====================

func TestRegisterToMuxDuplicateRoutes(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("UserController", controller, "/users")

	registry.RegisterRoute(RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "HandleGet",
	})
	registry.RegisterRoute(RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "HandleGet",
	})

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err == nil {
		t.Fatal("expected error for duplicate routes")
	}
	if !strings.Contains(err.Error(), "duplicate route") {
		t.Errorf("error should contain 'duplicate route', got: %v", err)
	}
}

// ==================== GetRoutes lazy resolution ====================

func TestGetRoutesLazyResolution(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterRoute(RouteInfo{
		Method:     "GET",
		Path:       "/users",
		StructName: "UserController",
		MethodName: "HandleGet",
	})

	routes := registry.GetRoutes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Controller != nil {
		t.Error("controller should be nil before registration")
	}

	registry.RegisterController("UserController", controller, "/users")

	routes = registry.GetRoutes()
	if routes[0].Controller == nil {
		t.Fatal("controller should be resolved from cache after registration")
	}
	if _, ok := routes[0].Controller.(*mockController); !ok {
		t.Errorf("expected *mockController, got %T", routes[0].Controller)
	}
}

func TestCreateHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("TestController", controller, "/test")

	route := RouteInfo{
		Method:     "POST",
		Path:       "/test",
		StructName: "TestController",
		MethodName: "HandleGet",
	}

	registry.RegisterRoute(route)

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestCreateHandler_WithProduces(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("TestController", controller, "/test")

	route := RouteInfo{
		Method:     "GET",
		Path:       "/test",
		StructName: "TestController",
		MethodName: "HandleWithContext",
		Produces:   "application/json",
	}

	registry.RegisterRoute(route)

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

func TestCreateHandler_MethodNotFound(t *testing.T) {
	t.Parallel()
	registry := NewRouteRegistry()
	controller := &mockController{}

	registry.RegisterController("TestController", controller, "/test")

	route := RouteInfo{
		Method:     "GET",
		Path:       "/test",
		StructName: "TestController",
		MethodName: "NonExistentMethod",
	}

	registry.RegisterRoute(route)

	mux := http.NewServeMux()
	err := registry.RegisterToMux(mux)
	if err == nil {
		t.Error("expected error for non-existent method")
	}
}

func TestSimpleContext_Next(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.Next()

	if ctx.IsAborted() {
		t.Error("expected Next() to not abort")
	}
}

func TestSimpleContext_QueryDefault(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test?name=John", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	val := ctx.QueryDefault("name", "default")
	if val != "John" {
		t.Errorf("expected 'John', got %s", val)
	}

	val = ctx.QueryDefault("missing", "default")
	if val != "default" {
		t.Errorf("expected 'default', got %s", val)
	}
}

func TestSimpleContext_AbortWithStatus(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.AbortWithStatus(http.StatusUnauthorized)

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestSimpleContext_AbortWithStatus_NoContent(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.AbortWithStatus(http.StatusNoContent)

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}
}

func TestSimpleContext_AbortWithStatusJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})

	if !ctx.IsAborted() {
		t.Error("expected context to be aborted")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}

	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["error"] != "invalid request" {
		t.Errorf("expected error message 'invalid request', got %q", result["error"])
	}
}

func TestSimpleContext_String(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.String(http.StatusOK, "Hello, %s!", "World")

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type to start with 'text/plain', got %q", ct)
	}

	body := rec.Body.String()
	if body != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", body)
	}
}

func TestSimpleContext_SetStatusCode(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.SetStatusCode(http.StatusCreated)

	if ctx.statusCode != http.StatusCreated {
		t.Errorf("expected status code 201, got %d", ctx.statusCode)
	}
}

func TestSimpleContext_SetHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	ctx.SetHeader("X-Custom-Header", "custom-value")

	header := rec.Header().Get("X-Custom-Header")
	if header != "custom-value" {
		t.Errorf("expected 'custom-value', got %q", header)
	}
}

func TestSimpleContext_SetContext(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	newCtx := context.WithValue(req.Context(), "key", "value")
	ctx.SetContext(newCtx)

	if ctx.Context() != newCtx {
		t.Error("expected context to be updated")
	}
}

func TestSimpleContext_RequestMethod(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	if ctx.RequestMethod() != "POST" {
		t.Errorf("expected 'POST', got %s", ctx.RequestMethod())
	}
}

func TestSimpleContext_RequestURI(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test?query=value", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	uri := ctx.RequestURI()
	if uri != "/test?query=value" {
		t.Errorf("expected '/test?query=value', got %s", uri)
	}
}

func TestSimpleContext_PathParam(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	val := ctx.PathParam("id")
	if val != "" {
		t.Errorf("expected empty string, got %s", val)
	}
}

func TestSimpleContext_Query(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test?name=John&age=30", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	name := ctx.Query("name")
	if name != "John" {
		t.Errorf("expected 'John', got %s", name)
	}

	age := ctx.Query("age")
	if age != "30" {
		t.Errorf("expected '30', got %s", age)
	}

	missing := ctx.Query("missing")
	if missing != "" {
		t.Errorf("expected empty string, got %s", missing)
	}
}

func TestSimpleContext_Header(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	auth := ctx.Header("Authorization")
	if auth != "Bearer token123" {
		t.Errorf("expected 'Bearer token123', got %s", auth)
	}

	missing := ctx.Header("X-Missing")
	if missing != "" {
		t.Errorf("expected empty string, got %s", missing)
	}
}

func TestSimpleContext_Request(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newSimpleContext(rec, req)

	if ctx.Request() != req {
		t.Error("expected request to match")
	}
}
