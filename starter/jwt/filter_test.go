package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/xudefa/enhance/security"
	"github.com/xudefa/enhance/security/filter"
)

func TestExtractBearerToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		authHeader string
		want       string
	}{
		{"valid bearer", "Bearer abc123", "abc123"},
		{"no bearer prefix", "Basic abc123", ""},
		{"empty header", "", ""},
		{"bearer only", "Bearer ", ""},
		{"token with dots", "Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig", "eyJhbGciOiJIUzI1NiJ9.payload.sig"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractBearerToken(tt.authHeader)
			if got != tt.want {
				t.Errorf("extractBearerToken(%q) = %q, want %q", tt.authHeader, got, tt.want)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		uri     string
		want    bool
	}{
		{"exact match", "/api/users", "/api/users", true},
		{"exact no match", "/api/users", "/api/posts", false},
		{"double star matches all", "/**", "/anything/at/all", true},
		{"double star prefix match", "/api/**", "/api/users/123", true},
		{"double star prefix no match", "/api/**", "/other/path", false},
		{"single star match", "/api/*", "/api/users", true},
		{"single star no match nested", "/api/*", "/api/users/123", false},
		{"single star match empty path segment", "/api/*", "/api/", true},
		{"different patterns", "/login", "/register", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchPattern(tt.pattern, tt.uri)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.uri, got, tt.want)
			}
		})
	}
}

func TestNewJwtAuthenticationFilter(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	if filter == nil {
		t.Fatal("NewJwtAuthenticationFilter() returned nil")
	}
	if filter.tokenProvider == nil {
		t.Error("tokenProvider is nil")
	}
	if len(filter.excludePaths) != 0 {
		t.Errorf("expected no exclude paths, got %v", filter.excludePaths)
	}
}

func TestNewJwtAuthenticationFilter_WithOptions(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))

	mockService := &mockUserDetailsService{}
	filter := NewJwtAuthenticationFilter(provider,
		WithExcludePaths("/login", "/register"),
		WithUserDetailsService(mockService),
	)

	if len(filter.excludePaths) != 2 {
		t.Errorf("expected 2 exclude paths, got %d", len(filter.excludePaths))
	}
	if filter.userDetailsService != mockService {
		t.Error("userDetailsService not set correctly")
	}
}

func TestJwtAuthenticationFilter_Order(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	if got := filter.Order(); got != -500 {
		t.Errorf("Order() = %d, want -500", got)
	}
}

func TestJwtAuthenticationFilter_IsJwtFilter(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	if !filter.IsJwtFilter() {
		t.Error("IsJwtFilter() should return true")
	}
}

func TestJwtAuthenticationFilter_DoFilter_InvalidContext(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	err := filter.DoFilter("not-a-context", nil, nil, nil)
	if err == nil {
		t.Error("expected error for invalid context type")
	}
}

func TestJwtAuthenticationFilter_DoFilter_InvalidRequest(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	err := filter.DoFilter(context.Background(), "not-a-request", nil, nil)
	if err == nil {
		t.Error("expected error for invalid request type")
	}
}

func TestJwtAuthenticationFilter_DoFilter_InvalidResponse(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	err := filter.DoFilter(context.Background(), &mockSecurityRequest{}, "not-a-response", nil)
	if err == nil {
		t.Error("expected error for invalid response type")
	}
}

func TestJwtAuthenticationFilter_DoFilter_ExcludedPath(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider,
		WithExcludePaths("/health", "/public/**"),
	)

	req := &mockSecurityRequest{uri: "/health", headers: map[string]string{}}
	resp := &mockSecurityResponse{}
	chain := &mockFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain.DoFilter to be called for excluded path")
	}
}

func TestJwtAuthenticationFilter_DoFilter_NoAuthHeader(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	req := &mockSecurityRequest{uri: "/api/users", headers: map[string]string{}}
	resp := &mockSecurityResponse{}
	chain := &mockFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain.DoFilter to be called when no auth header")
	}
}

func TestJwtAuthenticationFilter_DoFilter_InvalidTokenFormat(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	req := &mockSecurityRequest{
		uri:     "/api/users",
		headers: map[string]string{HeaderAuthorization: "Basic abc123"},
	}
	resp := &mockSecurityResponse{}
	chain := &mockFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain.DoFilter to be called for non-bearer token")
	}
}

func TestJwtAuthenticationFilter_DoFilter_ValidToken(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(
		WithSecretKey("test-secret"),
		WithExpiration(time.Hour),
	)

	token, err := provider.GenerateToken(context.Background(), "alice", []string{"ROLE_USER"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	filter := NewJwtAuthenticationFilter(provider)

	req := &mockSecurityRequest{
		uri:     "/api/users",
		headers: map[string]string{HeaderAuthorization: "Bearer " + token},
	}
	resp := &mockSecurityResponse{}
	chain := &mockFilterChain{}

	err = filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain.DoFilter to be called")
	}
	if req.attributes["security.currentAuthentication"] == nil {
		t.Error("expected security.currentAuthentication to be set")
	}
}

func TestJwtAuthenticationFilter_DoFilter_InvalidToken(t *testing.T) {
	t.Parallel()
	provider := NewTokenProvider(WithSecretKey("test-secret"))
	filter := NewJwtAuthenticationFilter(provider)

	req := &mockSecurityRequest{
		uri:     "/api/users",
		headers: map[string]string{HeaderAuthorization: "Bearer invalid-token"},
	}
	resp := &mockSecurityResponse{}
	chain := &mockFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err == nil {
		t.Error("expected error for invalid token")
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status code 401, got %d", resp.statusCode)
	}
}

func TestNewSimpleUserDetails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		username     string
		authorities  []string
		wantUsername string
	}{
		{"with authorities", "alice", []string{"ROLE_ADMIN"}, "alice"},
		{"no authorities", "bob", nil, "bob"},
		{"empty authorities", "charlie", []string{}, "charlie"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := NewSimpleUserDetails(tt.username, tt.authorities)
			if d.Username() != tt.wantUsername {
				t.Errorf("Username() = %q, want %q", d.Username(), tt.wantUsername)
			}
			if d.Password() != "" {
				t.Errorf("Password() = %q, want empty", d.Password())
			}
			if d.Enabled() != true {
				t.Error("Enabled() should return true")
			}
			if d.AccountNonExpired() != true {
				t.Error("AccountNonExpired() should return true")
			}
			if d.CredentialsNonExpired() != true {
				t.Error("CredentialsNonExpired() should return true")
			}
			if d.AccountNonLocked() != true {
				t.Error("AccountNonLocked() should return true")
			}
			if len(d.Authorities()) != len(tt.authorities) {
				t.Errorf("Authorities() length = %d, want %d", len(d.Authorities()), len(tt.authorities))
			}
		})
	}
}

type mockUserDetailsService struct{}

func (m *mockUserDetailsService) LoadUserByUsername(_ context.Context, _ string) (security.UserDetails, error) {
	return NewSimpleUserDetails("mock-user", []string{"ROLE_USER"}), nil
}

type mockSecurityRequest struct {
	uri        string
	headers    map[string]string
	attributes map[string]any
}

func (m *mockSecurityRequest) GetMethod() string              { return "GET" }
func (m *mockSecurityRequest) GetURI() string                 { return m.uri }
func (m *mockSecurityRequest) GetHeader(key string) string    { return m.headers[key] }
func (m *mockSecurityRequest) RemoteAddress() string          { return "127.0.0.1:8080" }
func (m *mockSecurityRequest) SetAttribute(key string, v any) {
	if m.attributes == nil {
		m.attributes = make(map[string]any)
	}
	m.attributes[key] = v
}
func (m *mockSecurityRequest) GetAttribute(key string) (any, bool) {
	if m.attributes == nil {
		return nil, false
	}
	v, ok := m.attributes[key]
	return v, ok
}

type mockSecurityResponse struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (m *mockSecurityResponse) SetStatusCode(code int) { m.statusCode = code }
func (m *mockSecurityResponse) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}
func (m *mockSecurityResponse) Write(data []byte) error {
	m.body = append(m.body, data...)
	return nil
}

type mockFilterChain struct {
	called bool
}

func (m *mockFilterChain) DoFilter(_, _, _ interface{}) error {
	m.called = true
	return nil
}
func (m *mockFilterChain) AddFilter(_ filter.Filter) {}
func (m *mockFilterChain) GetFilters() []filter.Filter { return nil }
