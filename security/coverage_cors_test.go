package security

import (
	"context"
	"testing"
)

// ============================================================
// filter_chain.go 测试
// ============================================================

// TestSecurityFilterChainAdapter 测试安全过滤器链适配器
func TestCorsFilter_IsOriginAllowed(t *testing.T) {
	t.Parallel()

	t.Run("empty allowed origins returns false", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{})
		if filter.isOriginAllowed("http://example.com") {
			t.Error("expected false for empty allowed origins")
		}
	})

	t.Run("exact match", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"http://example.com"},
		})
		if !filter.isOriginAllowed("http://example.com") {
			t.Error("expected true for exact match")
		}
		if filter.isOriginAllowed("http://other.com") {
			t.Error("expected false for non-matching origin")
		}
	})

	t.Run("wildcard suffix match", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"http://*.example.com"},
		})
		if !filter.isOriginAllowed("http://sub.example.com") {
			t.Error("expected true for wildcard match")
		}
		if !filter.isOriginAllowed("http://sub.example.com/") {
			t.Error("expected true for wildcard match with trailing slash")
		}
		if filter.isOriginAllowed("http://evil.com") {
			t.Error("expected false for non-matching origin")
		}
	})

	t.Run("star wildcard", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"*"},
		})
		if !filter.isOriginAllowed("http://any.com") {
			t.Error("expected true for * wildcard")
		}
	})
}

// TestCorsFilter_InvalidTypes 测试 CorsFilter 的类型检查

func TestCorsFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	corsFilter := NewCorsFilter(CorsConfig{})
	t.Run("invalid context", func(t *testing.T) {
		err := corsFilter.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})
	t.Run("invalid request", func(t *testing.T) {
		err := corsFilter.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})
	t.Run("invalid response", func(t *testing.T) {
		err := corsFilter.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestCorsFilter_DoFilter_OPTIONS 测试 CORS 预检请求

func TestCorsFilter_DoFilter_OPTIONS(t *testing.T) {
	t.Parallel()

	filter := NewCorsFilter(CorsConfig{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Custom"},
		AllowCredentials: true,
		MaxAge:           7200,
	})

	req := &mockSecurityRequest{method: "OPTIONS", uri: "/api/test"}
	req.SetHeader("Origin", "http://example.com")
	resp := &mockSecurityResponse{headers: make(map[string]string)}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.statusCode != 204 {
		t.Errorf("expected status 204, got %d", resp.statusCode)
	}
	if resp.headers["Access-Control-Allow-Origin"] != "http://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin header")
	}
	if resp.headers["Access-Control-Allow-Credentials"] != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials header")
	}
}

// ============================================================
// csrf.go 测试
// ============================================================

// TestCsrfFilter_NewCsrfFilter_Error 测试 NewCsrfFilter 的错误路径
