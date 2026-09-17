package security

import (
	"context"
	"strings"
	"testing"

	"github.com/xudefa/enhance/security/filter"
)

func testCsrfFilterGetRequestGeneratesToken(t *testing.T, csrfFilter *CsrfFilter) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{method: "GET", uri: "/api/test"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := csrfFilter.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !chain.called {
		t.Error("Expected filter chain to be called")
	}

	token, exists := req.GetAttribute("csrf.token")
	if !exists || token == nil {
		t.Error("Expected CSRF token to be generated")
	}
}

func testCsrfFilterPostWithoutTokenFails(t *testing.T, csrfFilter *CsrfFilter) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := csrfFilter.DoFilter(ctx, req, resp, chain)
	if err == nil {
		t.Error("Expected error for missing CSRF token")
	}
	if !strings.Contains(err.Error(), "missing CSRF token") {
		t.Errorf("Expected 'missing CSRF token' error, got %v", err)
	}
}

func testCsrfFilterExcludePathSkipsCSRF(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokenRepo := NewCookieCsrfTokenRepository()
	csrfFilter := MustNewCsrfFilter(tokenRepo)
	_ = csrfFilter.AddExcludePath("/public")

	req := &mockSecurityRequest{method: "POST", uri: "/public/api"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := csrfFilter.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error for excluded path, got %v", err)
	}
}

func TestCsrfFilter(t *testing.T) {
	t.Parallel()

	tokenRepo := NewCookieCsrfTokenRepository()
	csrfFilter := MustNewCsrfFilter(tokenRepo)

	t.Run("GET request should generate token", func(t *testing.T) { testCsrfFilterGetRequestGeneratesToken(t, csrfFilter) })
	t.Run("POST request without token should fail", func(t *testing.T) { testCsrfFilterPostWithoutTokenFails(t, csrfFilter) })
	t.Run("Exclude path should skip CSRF", testCsrfFilterExcludePathSkipsCSRF)
}

func TestCookieCsrfTokenRepository(t *testing.T) {
	t.Parallel()
	repo := NewCookieCsrfTokenRepository()
	ctx := context.Background()

	t.Run("Generate token", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		token, err := repo.GenerateToken(ctx, req)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		if token.Value == "" {
			t.Error("Expected non-empty token value")
		}
		if token.Identifier != "/api/test" {
			t.Errorf("Expected identifier '/api/test', got '%s'", token.Identifier)
		}
	})

	t.Run("Validate token", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		req.SetAttribute("csrf.token", "test-token")
		valid := repo.ValidateToken(ctx, req, "test-token")
		if !valid {
			t.Error("Expected token to be valid")
		}

		valid = repo.ValidateToken(ctx, req, "wrong-token")
		if valid {
			t.Error("Expected token to be invalid")
		}
	})

	t.Run("Save token", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		token := &CsrfToken{Identifier: "/api/test", Value: "test-token-value"}
		repo.SaveToken(ctx, req, resp, token)

		cookie := resp.headers["Set-Cookie"]
		if cookie == "" {
			t.Error("Expected Set-Cookie header")
		}
		if !strings.Contains(cookie, "test-token-value") {
			t.Errorf("Expected token value in cookie, got %s", cookie)
		}
	})
}

func TestCsrfTokenManager(t *testing.T) {
	t.Parallel()
	manager := NewCsrfTokenManager()

	t.Run("Generate and validate token", func(t *testing.T) {
		token, err := manager.GenerateToken("user1")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		if token == "" {
			t.Error("Expected non-empty token")
		}

		if !manager.ValidateToken("user1", token) {
			t.Error("Expected token to be valid")
		}

		if manager.ValidateToken("user1", "invalid-token") {
			t.Error("Expected invalid token to fail validation")
		}
	})

	t.Run("Remove token", func(t *testing.T) {
		token, err := manager.GenerateToken("user2")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		manager.RemoveToken("user2")

		if manager.ValidateToken("user2", token) {
			t.Error("Expected token to be invalid after removal")
		}
	})

	t.Run("Different users have different tokens", func(t *testing.T) {
		token1, err := manager.GenerateToken("user3")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
		token2, err := manager.GenerateToken("user4")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		if token1 == token2 {
			t.Error("Expected different tokens for different users")
		}
	})
}

type mockLogoutSuccessHandler struct {
	targetCalled string
	called       bool
}

func (h *mockLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	h.called = true
	response.SetStatusCode(302)
	response.SetHeader("Location", h.targetCalled)
}

type mockSecurityRequest struct {
	uri        string
	method     string
	headers    map[string]string
	attributes map[string]any
	remoteAddr string
}

func (r *mockSecurityRequest) GetMethod() string { return r.method }
func (r *mockSecurityRequest) GetURI() string    { return r.uri }
func (r *mockSecurityRequest) GetHeader(key string) string {
	if r.headers == nil {
		return ""
	}
	return r.headers[key]
}
func (r *mockSecurityRequest) RemoteAddress() string {
	if r.remoteAddr != "" {
		return r.remoteAddr
	}
	return "127.0.0.1:8080"
}
func (r *mockSecurityRequest) SetHeader(key, value string) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
}
func (r *mockSecurityRequest) SetAttribute(key string, value any) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	if r.attributes == nil {
		r.attributes = make(map[string]any)
	}
	r.attributes[key] = value
}
func (r *mockSecurityRequest) GetAttribute(key string) (any, bool) {
	attributeValue, ok := r.attributes[key]
	return attributeValue, ok
}

func newMockSecurityRequest(method, uri string, headers map[string]string) *mockSecurityRequest {
	if headers == nil {
		headers = make(map[string]string)
	}
	return &mockSecurityRequest{
		method:  method,
		uri:     uri,
		headers: headers,
	}
}

func newMockSecurityResponse() *mockSecurityResponse {
	return &mockSecurityResponse{headers: make(map[string]string)}
}

type mockSecurityResponse struct {
	statusCode int
	headers    map[string]string
}

func (r *mockSecurityResponse) StatusCode() int        { return r.statusCode }
func (r *mockSecurityResponse) SetStatusCode(code int) { r.statusCode = code }
func (r *mockSecurityResponse) Header(name string) string {
	if r.headers == nil {
		return ""
	}
	return r.headers[name]
}
func (r *mockSecurityResponse) SetHeader(name, value string) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[name] = value
}
func (r *mockSecurityResponse) Write(data []byte) error { return nil }

type mockSecurityFilterChain struct {
	called bool
}

func (m *mockSecurityFilterChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	m.called = true
	return nil
}

func (m *mockSecurityFilterChain) AddFilter(filter filter.Filter) {}

func (m *mockSecurityFilterChain) GetFilters() []filter.Filter {
	return nil
}
