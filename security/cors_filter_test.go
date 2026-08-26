package security

import (
	"context"
	"testing"
)

func TestNewCorsFilter_Defaults(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{})

	if f == nil {
		t.Fatal("expected non-nil CorsFilter")
	}
	if f.Order() != -100 {
		t.Errorf("expected order -100, got %d", f.Order())
	}
	if len(f.config.AllowedMethods) == 0 {
		t.Error("expected default AllowedMethods")
	}
	if len(f.config.AllowedHeaders) == 0 {
		t.Error("expected default AllowedHeaders")
	}
	if f.config.MaxAge != 3600 {
		t.Errorf("expected default MaxAge 3600, got %d", f.config.MaxAge)
	}
}

func TestNewCorsFilter_CustomConfig(t *testing.T) {
	t.Parallel()

	cfg := CorsConfig{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Custom"},
		AllowCredentials: true,
		MaxAge:           7200,
	}
	f := NewCorsFilter(cfg)

	if len(f.config.AllowedOrigins) != 1 || f.config.AllowedOrigins[0] != "http://example.com" {
		t.Errorf("expected AllowedOrigins [http://example.com], got %v", f.config.AllowedOrigins)
	}
	if f.config.AllowCredentials != true {
		t.Error("expected AllowCredentials to be true")
	}
	if f.config.MaxAge != 7200 {
		t.Errorf("expected MaxAge 7200, got %d", f.config.MaxAge)
	}
}

func TestCorsFilter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
	chain := &mockFilterChain{}

	err := f.DoFilter("notContext", nil, nil, chain)
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = f.DoFilter(context.Background(), "notReq", nil, chain)
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = f.DoFilter(context.Background(), newMockSecurityRequest("GET", "/", nil), "notResp", chain)
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestCorsFilter_NoOrigin_PassesThrough(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/", nil)
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called when no Origin header")
	}
}

func TestCorsFilter_AllowedOrigin(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{
		AllowedOrigins: []string{"http://example.com"},
	})
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/", map[string]string{"Origin": "http://example.com"})
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.headers["Access-Control-Allow-Origin"] != "http://example.com" {
		t.Errorf("expected Allow-Origin header, got %v", resp.headers["Access-Control-Allow-Origin"])
	}
	if !chain.called {
		t.Error("expected chain to be called")
	}
}

func TestCorsFilter_AllowedOriginWithCredentials(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{
		AllowedOrigins:   []string{"http://example.com"},
		AllowCredentials: true,
	})
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/", map[string]string{"Origin": "http://example.com"})
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.headers["Access-Control-Allow-Credentials"] != "true" {
		t.Error("expected Allow-Credentials header")
	}
}

func TestCorsFilter_PreflightRequest(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Custom"},
		AllowCredentials: true,
		MaxAge:           1800,
	})
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("OPTIONS", "/", map[string]string{"Origin": "http://example.com"})
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 204 {
		t.Errorf("expected status 204, got %d", resp.statusCode)
	}
	if resp.headers["Access-Control-Allow-Methods"] != "GET, POST" {
		t.Errorf("expected Allow-Methods header, got %v", resp.headers["Access-Control-Allow-Methods"])
	}
	if resp.headers["Access-Control-Allow-Headers"] != "Content-Type" {
		t.Errorf("expected Allow-Headers header, got %v", resp.headers["Access-Control-Allow-Headers"])
	}
	if resp.headers["Access-Control-Expose-Headers"] != "X-Custom" {
		t.Errorf("expected Expose-Headers header, got %v", resp.headers["Access-Control-Expose-Headers"])
	}
	if resp.headers["Access-Control-Max-Age"] != "1800" {
		t.Errorf("expected Max-Age header, got %v", resp.headers["Access-Control-Max-Age"])
	}
	if chain.called {
		t.Error("chain should not be called for preflight")
	}
}

func TestCorsFilter_DisallowedOrigin(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{
		AllowedOrigins: []string{"http://example.com"},
	})
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/", map[string]string{"Origin": "http://evil.com"})
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.headers["Access-Control-Allow-Origin"] != "" {
		t.Error("expected no Allow-Origin header for disallowed origin")
	}
	if !chain.called {
		t.Error("expected chain to be called")
	}
}

func TestMatchOriginPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		origin  string
		want    bool
	}{
		{"wildcard", "*", "http://any.com", true},
		{"exact match", "http://example.com", "http://example.com", true},
		{"exact mismatch", "http://example.com", "http://other.com", false},
		{"wildcard subdomain", "http://*.example.com", "http://sub.example.com", true},
		{"wildcard subdomain mismatch", "http://*.example.com", "http://sub.other.com", false},
		{"trailing slash mismatch", "http://example.com/", "http://example.com", false},
		{"no star no match", "http://a.com", "http://b.com", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchOriginPattern(tt.pattern, tt.origin)
			if got != tt.want {
				t.Errorf("matchOriginPattern(%q, %q) = %v, want %v", tt.pattern, tt.origin, got, tt.want)
			}
		})
	}
}

func TestCorsFilter_EmptyAllowedOrigins(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{AllowedOrigins: []string{}})
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/", map[string]string{"Origin": "http://example.com"})
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.headers["Access-Control-Allow-Origin"] != "" {
		t.Error("expected no Allow-Origin header when AllowedOrigins is empty")
	}
}
