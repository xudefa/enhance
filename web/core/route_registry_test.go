package core

import (
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
