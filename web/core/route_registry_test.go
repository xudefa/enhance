package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockController 模拟控制器。
type mockController struct{}

func (m *mockController) HandleGet() {
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
