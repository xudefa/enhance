package security

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

func TestCsrfFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tokenRepo := NewCookieCsrfTokenRepository()
	csrfFilter := MustNewCsrfFilter(tokenRepo)

	t.Run("GET request should generate token", func(t *testing.T) {
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
	})

	t.Run("POST request without token should fail", func(t *testing.T) {
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
	})

	t.Run("Exclude path should skip CSRF", func(t *testing.T) {
		csrfFilterWithExclude := MustNewCsrfFilter(tokenRepo)
		_ = csrfFilterWithExclude.AddExcludePath("/public")

		req := &mockSecurityRequest{method: "POST", uri: "/public/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		err := csrfFilterWithExclude.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error for excluded path, got %v", err)
		}
	})
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

func TestLogoutFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Logout success", func(t *testing.T) {
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
		logoutCtx := ContextWithAuthentication(ctx, auth)

		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		logoutFilter := MustNewLogoutFilter("/logout", []LogoutHandler{NewSecurityContextLogoutHandler()})

		err := logoutFilter.DoFilter(logoutCtx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if chain.called {
			t.Error("Filter chain should not be called after logout")
		}
	})

	t.Run("Non-logout URL should skip", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		logoutFilter := MustNewLogoutFilter("/logout", []LogoutHandler{})

		err := logoutFilter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !chain.called {
			t.Error("Filter chain should be called for non-logout URL")
		}
	})

	t.Run("Custom success handler", func(t *testing.T) {
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
		logoutCtx := ContextWithAuthentication(ctx, auth)

		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		customHandler := &mockLogoutSuccessHandler{targetCalled: "/custom"}
		logoutFilter := MustNewLogoutFilter("/logout", []LogoutHandler{})
		logoutFilter.SetSuccessHandler(customHandler)

		err := logoutFilter.DoFilter(logoutCtx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if resp.headers["Location"] != "/custom" {
			t.Errorf("Expected Location '/custom', got '%s'", resp.headers["Location"])
		}
	})
}

func TestLogoutSuccessHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("DefaultLogoutSuccessHandler", func(t *testing.T) {
		handler := NewDefaultLogoutSuccessHandler("/login?logout")
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		handler.OnLogoutSuccess(ctx, req, resp, auth)

		if resp.statusCode != 302 {
			t.Errorf("Expected status code 302, got %d", resp.statusCode)
		}
		if resp.headers["Location"] != "/login?logout" {
			t.Errorf("Expected Location '/login?logout', got '%s'", resp.headers["Location"])
		}
	})

	t.Run("SimpleLogoutSuccessHandler", func(t *testing.T) {
		handler := NewSimpleLogoutSuccessHandler("/home")
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		handler.OnLogoutSuccess(ctx, req, resp, auth)

		if resp.statusCode != 302 {
			t.Errorf("Expected status code 302, got %d", resp.statusCode)
		}
		if resp.headers["Location"] != "/home" {
			t.Errorf("Expected Location '/home', got '%s'", resp.headers["Location"])
		}
	})
}

func TestCookieClearingLogoutHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}
	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

	handler := NewCookieClearingLogoutHandler("session_id")

	handler.Logout(ctx, req, resp, auth)

	cookie := resp.headers["Set-Cookie"]
	if cookie == "" {
		t.Error("Expected Set-Cookie header")
	}
	if !strings.Contains(cookie, "session_id") {
		t.Errorf("Expected cookie 'session_id' to be cleared, got %s", cookie)
	}
	if !strings.Contains(cookie, "Max-Age=0") {
		t.Errorf("Expected cookie to have Max-Age=0, got %s", cookie)
	}
}

func TestUsernamePasswordAuthenticationFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("Successful authentication", func(t *testing.T) {
		req := &mockSecurityRequest{
			method: "POST",
			uri:    "/login",
		}
		req.SetHeader("username", "admin")
		req.SetHeader("password", "admin123")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/home", "/login?error", authManager, log.Build())

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
	})

	t.Run("Failed authentication", func(t *testing.T) {
		req := &mockSecurityRequest{
			method: "POST",
			uri:    "/login",
		}
		req.SetHeader("username", "admin")
		req.SetHeader("password", "wrongpassword")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/home", "/login?error", authManager, log.Build())

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if resp.statusCode != 401 {
			t.Errorf("Expected status code 401, got %d", resp.statusCode)
		}
	})

	t.Run("Non-login URL should skip", func(t *testing.T) {
		req := &mockSecurityRequest{
			method: "POST",
			uri:    "/api/test",
		}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		filter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/home", "/login?error", authManager, log.Build())

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !chain.called {
			t.Error("Filter chain should be called for non-login URL")
		}
	})
}

func TestBasicAuthenticationFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("Successful Basic auth", func(t *testing.T) {
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
	})

	t.Run("Missing Authorization header returns error", func(t *testing.T) {
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
	})

	t.Run("Invalid credentials returns error", func(t *testing.T) {
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
	})
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

func TestHttpSecurityConfiguration(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("Build with all features", func(t *testing.T) {
		chain, err := NewHttpSecurity().
			AuthenticationManager(authManager).
			Csrf().
			FormLogin("/api/login", "/dashboard").
			Logout("/api/logout").
			HttpBasic().
			Build()

		if err != nil {
			t.Fatalf("Failed to build security chain: %v", err)
		}
		if chain == nil {
			t.Error("Expected security chain to be non-nil")
		}
	})

	t.Run("Custom FormLogin URL", func(t *testing.T) {
		chain, err := NewHttpSecurity().
			AuthenticationManager(authManager).
			FormLogin("/custom-login").
			Build()

		if err != nil {
			t.Fatalf("Failed to build security chain: %v", err)
		}
		if chain == nil {
			t.Error("Expected security chain to be non-nil")
		}
	})

	t.Run("Custom Logout URL", func(t *testing.T) {
		chain, err := NewHttpSecurity().
			AuthenticationManager(authManager).
			Logout("/custom-logout", NewSimpleLogoutSuccessHandler("/goodbye")).
			Build()

		if err != nil {
			t.Fatalf("Failed to build security chain: %v", err)
		}
		if chain == nil {
			t.Error("Expected security chain to be non-nil")
		}
	})
}

func TestHttpSecurity_AuthorizeRequests_AppliesRules(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("AntMatchers and AnyRequest rules applied", func(t *testing.T) {
		t.Parallel()
		h := NewHttpSecurity().
			AuthenticationManager(authManager).
			AuthorizeRequests(func(authz AuthorizeRequests) {
				authz.AntMatchers("/api/**").HasRole("ROLE_API")
				authz.AntMatchers("/admin/**").DenyAll()
				authz.AnyRequest().Authenticated()
			})

		sec := h.(*httpSecurity)
		if len(sec.authorizeRules) != 3 {
			t.Fatalf("expected 3 collected rules, got %d", len(sec.authorizeRules))
		}

		chain, err := h.Build()
		if err != nil {
			t.Fatalf("failed to build: %v", err)
		}
		if chain == nil {
			t.Fatal("expected non-nil chain")
		}

		source, ok := sec.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource)
		if !ok {
			t.Fatal("expected expression based metadata source")
		}
		ctx := context.Background()

		attrs, err := source.GetAttributes(ctx, &mockSecurityRequest{method: "GET", uri: "/api/users"})
		if err != nil || len(attrs) != 1 || attrs[0] != "hasRole('ROLE_API')" {
			t.Errorf("expected hasRole('ROLE_API') for /api/users, got %v (err=%v)", attrs, err)
		}

		attrs, err = source.GetAttributes(ctx, &mockSecurityRequest{method: "GET", uri: "/admin/panel"})
		if err != nil || len(attrs) != 1 || attrs[0] != "denyAll" {
			t.Errorf("expected denyAll for /admin/panel, got %v (err=%v)", attrs, err)
		}

		attrs, err = source.GetAttributes(ctx, &mockSecurityRequest{method: "GET", uri: "/other"})
		if err != nil || len(attrs) != 1 || attrs[0] != "authenticated" {
			t.Errorf("expected authenticated for /other, got %v (err=%v)", attrs, err)
		}
	})

	t.Run("no rules still builds with empty source", func(t *testing.T) {
		t.Parallel()
		h := NewHttpSecurity().AuthenticationManager(authManager)
		chain, err := h.Build()
		if err != nil {
			t.Fatalf("failed to build: %v", err)
		}
		if chain == nil {
			t.Fatal("expected non-nil chain")
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

func TestSecurityContextLogoutHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	handler := NewSecurityContextLogoutHandler()

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}

	handler.Logout(ctx, req, resp, auth)

	// SecurityContextLogoutHandler is a no-op; authentication clearing
	// is handled by LogoutFilter which reads from context.
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
	val, ok := r.attributes[key]
	return val, ok
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
