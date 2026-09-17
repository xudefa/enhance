package security

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/xudefa/enhance/log"
)

func testUsernamePasswordAuthFilterSuccessful(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{
		method: "POST",
		uri:    "/login",
	}
	req.SetHeader("username", "admin")
	req.SetHeader("password", "admin123")
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", authManager, log.Build(), WithDefaultSuccessURL("/home"), WithFailureURL("/login?error"))

	err := filter.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.statusCode != 302 {
		t.Errorf("Expected status code 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/home" {
		t.Errorf("Expected Location '/home', got '%s'", resp.headers["Location"])
	}

	authVal, exists := req.GetAttribute("security.currentAuthentication")
	if !exists || authVal == nil {
		t.Error("Expected authentication to be set in request attribute")
	}
	if auth, ok := authVal.(Authentication); ok {
		if extractPrincipalName(auth) != "admin" {
			t.Errorf("Expected username 'admin', got '%s'", extractPrincipalName(auth))
		}
	}
}

func testUsernamePasswordAuthFilterFailed(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{
		method: "POST",
		uri:    "/login",
	}
	req.SetHeader("username", "admin")
	req.SetHeader("password", "wrongpassword")
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", authManager, log.Build(), WithDefaultSuccessURL("/home"), WithFailureURL("/login?error"))

	err := filter.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.statusCode != 401 {
		t.Errorf("Expected status code 401, got %d", resp.statusCode)
	}
}

func testUsernamePasswordAuthFilterNonLoginURL(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{
		method: "POST",
		uri:    "/api/test",
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", authManager, log.Build(), WithDefaultSuccessURL("/home"), WithFailureURL("/login?error"))

	err := filter.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !chain.called {
		t.Error("Filter chain should be called for non-login URL")
	}
}

func TestUsernamePasswordAuthenticationFilter(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("Successful authentication", func(t *testing.T) { testUsernamePasswordAuthFilterSuccessful(t, authManager) })
	t.Run("Failed authentication", func(t *testing.T) { testUsernamePasswordAuthFilterFailed(t, authManager) })
	t.Run("Non-login URL should skip", func(t *testing.T) { testUsernamePasswordAuthFilterNonLoginURL(t, authManager) })
}

func testBasicAuthFilterSuccessful(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	ctx := context.Background()
	encoded := base64.StdEncoding.EncodeToString([]byte("admin:admin123"))
	req := &mockSecurityRequest{
		method: "GET",
		uri:    "/api/test",
	}
	req.SetHeader("Authorization", "Basic "+encoded)
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	filter := NewBasicAuthenticationFilterWithRealm(authManager, "Test Realm", log.Build())

	err := filter.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !chain.called {
		t.Error("Expected chain to be called")
	}
}

func testBasicAuthFilterMissingHeader(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{
		method: "GET",
		uri:    "/api/test",
	}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	filter := NewBasicAuthenticationFilterWithRealm(authManager, "Test Realm", log.Build())

	err := filter.DoFilter(ctx, req, resp, chain)
	if err == nil {
		t.Error("Expected error for missing Authorization header")
	}
	if resp.statusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.statusCode)
	}
}

func testBasicAuthFilterInvalidCredentials(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	ctx := context.Background()
	encoded := base64.StdEncoding.EncodeToString([]byte("admin:wrongpassword"))
	req := &mockSecurityRequest{
		method: "GET",
		uri:    "/api/test",
	}
	req.SetHeader("Authorization", "Basic "+encoded)
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	filter := NewBasicAuthenticationFilterWithRealm(authManager, "Test Realm", log.Build())

	err := filter.DoFilter(ctx, req, resp, chain)
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
	if resp.statusCode != 401 {
		t.Errorf("Expected status 401, got %d", resp.statusCode)
	}
}

func TestBasicAuthenticationFilter(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("Successful Basic auth", func(t *testing.T) { testBasicAuthFilterSuccessful(t, authManager) })
	t.Run("Missing Authorization header returns error", func(t *testing.T) { testBasicAuthFilterMissingHeader(t, authManager) })
	t.Run("Invalid credentials returns error", func(t *testing.T) { testBasicAuthFilterInvalidCredentials(t, authManager) })
}

func TestBasicAuthenticationEntryPoint(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Send challenge", func(t *testing.T) {
		entryPoint := NewBasicAuthenticationEntryPointWithRealm("Test Realm", log.Build())
		req := &mockSecurityRequest{method: "GET", uri: "/api/test"}
		resp := &mockSecurityResponse{}

		err := entryPoint.Commence(ctx, req, resp, ErrBadCredentials)
		if err == nil {
			t.Error("Expected error to be returned")
		}

		if resp.statusCode != 401 {
			t.Errorf("Expected status code 401, got %d", resp.statusCode)
		}

		wwwAuth := resp.headers["WWW-Authenticate"]
		if !strings.Contains(wwwAuth, `Basic realm="Test Realm"`) {
			t.Errorf("Expected WWW-Authenticate header with realm, got '%s'", wwwAuth)
		}
	})
}
