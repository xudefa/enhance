package actuator

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type testHandlerRegistry struct {
	mu       sync.Mutex
	handlers map[string]http.Handler
}

func newTestHandlerRegistry() *testHandlerRegistry {
	return &testHandlerRegistry{handlers: make(map[string]http.Handler)}
}

func (r *testHandlerRegistry) Handle(pattern string, handler http.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[pattern] = handler
}

func (r *testHandlerRegistry) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.handlers)
}

func TestNewHttpEndpointRegistryAdapter(t *testing.T) {
	t.Parallel()
	reg := newTestHandlerRegistry()
	adapter := NewHttpEndpointRegistryAdapter(reg)
	if adapter == nil {
		t.Fatal("NewHttpEndpointRegistryAdapter returned nil")
	}
}

func TestHttpEndpointRegistryAdapter_RegisterEndpoint_WithMethod(t *testing.T) {
	t.Parallel()
	reg := newTestHandlerRegistry()
	adapter := NewHttpEndpointRegistryAdapter(reg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	adapter.RegisterEndpoint("GET", "/health", handler)

	if reg.count() != 1 {
		t.Errorf("expected 1 handler, got %d", reg.count())
	}
	if !adapter.HasEndpoint("/health") {
		t.Error("expected HasEndpoint to return true")
	}
}

func TestHttpEndpointRegistryAdapter_RegisterEndpoint_NoMethod(t *testing.T) {
	t.Parallel()
	reg := newTestHandlerRegistry()
	adapter := NewHttpEndpointRegistryAdapter(reg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	adapter.RegisterEndpoint("", "/health", handler)

	if reg.count() != 1 {
		t.Errorf("expected 1 handler, got %d", reg.count())
	}
	if !adapter.HasEndpoint("/health") {
		t.Error("expected HasEndpoint to return true")
	}
}

func TestHttpEndpointRegistryAdapter_HasEndpoint_NotFound(t *testing.T) {
	t.Parallel()
	reg := newTestHandlerRegistry()
	adapter := NewHttpEndpointRegistryAdapter(reg)

	if adapter.HasEndpoint("/nonexistent") {
		t.Error("expected HasEndpoint to return false")
	}
}

func TestHttpEndpointRegistryAdapter_NilRegistry(t *testing.T) {
	t.Parallel()
	adapter := NewHttpEndpointRegistryAdapter(nil)
	adapter.RegisterEndpoint("GET", "/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	if adapter.HasEndpoint("/health") {
		t.Error("should not register endpoint with nil registry")
	}
}

func TestHttpEndpointRegistryAdapter_RegisterEndpoints(t *testing.T) {
	t.Parallel()
	reg := newTestHandlerRegistry()
	adapter := NewHttpEndpointRegistryAdapter(reg)

	endpoints := []EndpointConfig{
		{Method: "GET", Path: "/a", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
		{Method: "POST", Path: "/b", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
		{Path: "/c", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})},
	}
	adapter.RegisterEndpoints(endpoints)

	if reg.count() != 3 {
		t.Errorf("expected 3 handlers, got %d", reg.count())
	}
	if !adapter.HasEndpoint("/a") || !adapter.HasEndpoint("/b") || !adapter.HasEndpoint("/c") {
		t.Error("expected all endpoints to be registered")
	}
}

func TestStdHttpHandlerRegistry_Handle(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	reg := &StdHttpHandlerRegistry{Mux: mux}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	reg.Handle("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestStdHttpHandlerRegistry_NilMux(t *testing.T) {
	t.Parallel()
	reg := &StdHttpHandlerRegistry{Mux: nil}
	reg.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func TestPathNormalizer_NormalizePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "/"},
		{"already root", "/", "/"},
		{"without leading slash", "health", "/health"},
		{"with trailing slash", "health/", "/health"},
		{"double slash", "/a//b", "/a/b"},
		{"multiple double slash", "/a///b//c", "/a/b/c"},
		{"single segment", "/health", "/health"},
		{"nested", "/api/v1/health", "/api/v1/health"},
		{"trailing slash nested", "/api/v1/health/", "/api/v1/health"},
	}

	normalizer := PathNormalizer{}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizer.NormalizePath(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnsureLeadingSlash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"health", "/health"},
		{"/health", "/health"},
		{"", "/"},
		{"/", "/"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := EnsureLeadingSlash(tt.input)
			if got != tt.want {
				t.Errorf("EnsureLeadingSlash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJoinPath_TableDriven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		base string
		path string
		want string
	}{
		{"basic", "/api", "/health", "/api/health"},
		{"base with trailing slash", "/api/", "/health", "/api/health"},
		{"path without leading slash", "/api", "health", "/api/health"},
		{"both slashes", "/api/", "/health", "/api/health"},
		{"nested", "/api/v1", "/users", "/api/v1/users"},
		{"root base", "/", "/health", "/health"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := JoinPath(tt.base, tt.path)
			if got != tt.want {
				t.Errorf("JoinPath(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
			}
		})
	}
}
