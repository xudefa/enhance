package security

import (
	"context"
	"sync"
	"testing"

	"github.com/xudefa/enhance/log"
)

func TestSecurityContextComprehensive(t *testing.T) {
	t.Parallel()

	t.Run("SetAuthentication and Authentication", func(t *testing.T) {
		t.Parallel()
		ctx := NewSecurityContext()

		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user1", []string{"ROLE_USER"})
		ctx.SetAuthentication(auth)

		got := ctx.Authentication()
		if got == nil {
			t.Fatal("expected non-nil authentication")
		}
		if extractPrincipalName(got) != "user1" {
			t.Errorf("expected principal 'user1', got '%s'", extractPrincipalName(got))
		}
	})

	t.Run("ClearAuthentication", func(t *testing.T) {
		t.Parallel()
		ctx := NewSecurityContext()

		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user1", []string{"ROLE_USER"})
		ctx.SetAuthentication(auth)

		ctx.ClearAuthentication()
		got := ctx.Authentication()
		if got != nil {
			t.Errorf("expected nil after clear, got %v", got)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		t.Parallel()
		ctx := NewSecurityContext()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				auth := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})
				ctx.SetAuthentication(auth)
				_ = ctx.Authentication()
			}(i)
		}
		wg.Wait()
	})
}

func TestContextWithAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("stores and retrieves authentication", func(t *testing.T) {
		t.Parallel()
		baseCtx := context.Background()
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		ctx := ContextWithAuthentication(baseCtx, auth)
		got := GetAuthenticationFromContext(ctx)

		if got == nil {
			t.Fatal("expected non-nil authentication")
		}
		if extractPrincipalName(got) != "admin" {
			t.Errorf("expected principal 'admin', got '%s'", extractPrincipalName(got))
		}
	})

	t.Run("returns nil when no authentication", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		got := GetAuthenticationFromContext(ctx)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestCorsFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("AllowAll wildcard origin", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"*"},
		})
		req := &mockSecurityRequest{method: "GET", uri: "/api", headers: map[string]string{"Origin": "http://example.com"}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.headers["Access-Control-Allow-Origin"] != "http://example.com" {
			t.Errorf("expected Allow-Origin header, got %s", resp.headers["Access-Control-Allow-Origin"])
		}
	})

	t.Run("preflight OPTIONS request", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"http://example.com"},
		})
		req := &mockSecurityRequest{method: "OPTIONS", uri: "/api", headers: map[string]string{"Origin": "http://example.com"}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if chain.called {
			t.Error("filter chain should not be called for preflight")
		}
		if resp.statusCode != 204 {
			t.Errorf("expected status 204, got %d", resp.statusCode)
		}
		if resp.headers["Access-Control-Allow-Methods"] == "" {
			t.Error("expected Allow-Methods header")
		}
	})

	t.Run("blocked origin", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"http://allowed.com"},
		})
		req := &mockSecurityRequest{method: "GET", uri: "/api", headers: map[string]string{"Origin": "http://evil.com"}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.headers["Access-Control-Allow-Origin"] == "http://evil.com" {
			t.Error("blocked origin should not be allowed")
		}
	})

	t.Run("no Origin header skips CORS", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"*"},
		})
		req := &mockSecurityRequest{method: "GET", uri: "/api", headers: map[string]string{}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("chain should be called when no Origin")
		}
	})

	t.Run("invalid ctx type returns error", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
		err := filter.DoFilter("not-a-context", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context type")
		}
	})

	t.Run("invalid request type returns error", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
		err := filter.DoFilter(ctx, "not-a-request", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request type")
		}
	})

	t.Run("invalid response type returns error", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
		err := filter.DoFilter(ctx, &mockSecurityRequest{}, "not-a-response", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response type")
		}
	})

	t.Run("AllowCredentials header", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins:   []string{"http://example.com"},
			AllowCredentials: true,
		})
		req := &mockSecurityRequest{method: "GET", uri: "/api", headers: map[string]string{"Origin": "http://example.com"}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		_ = filter.DoFilter(ctx, req, resp, chain)
		if resp.headers["Access-Control-Allow-Credentials"] != "true" {
			t.Error("expected Allow-Credentials header")
		}
	})

	t.Run("exact origin match", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"https://example.com"},
		})
		req := &mockSecurityRequest{method: "GET", uri: "/api", headers: map[string]string{"Origin": "https://example.com"}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		_ = filter.DoFilter(ctx, req, resp, chain)
		if resp.headers["Access-Control-Allow-Origin"] != "https://example.com" {
			t.Errorf("expected origin allowed, got %s", resp.headers["Access-Control-Allow-Origin"])
		}
	})

	t.Run("Order returns -100", func(t *testing.T) {
		t.Parallel()
		filter := NewCorsFilter(CorsConfig{})
		if filter.Order() != -100 {
			t.Errorf("expected order -100, got %d", filter.Order())
		}
	})
}

func TestSecurityBuilder(t *testing.T) {
	t.Parallel()

	t.Run("Build returns SecurityConfig", func(t *testing.T) {
		t.Parallel()
		b := NewSecurityBuilder()
		config := b.Build()
		if config == nil {
			t.Fatal("expected non-nil config")
		}
	})

	t.Run("builder setters return self", func(t *testing.T) {
		t.Parallel()
		b := NewSecurityBuilder()
		result := b.
			AuthenticationManager(nil).
			UserDetailsService(nil).
			PasswordEncoder(nil).
			AccessDecisionManager(nil).
			EnableAnonymous().
			EnableCsrf().
			EnableFormLogin("/login").
			EnableHttpBasic().
			EnableLogout("/logout")
		if result == nil {
			t.Fatal("builder setters should return self")
		}
	})

	t.Run("AddFilter and filter positioning", func(t *testing.T) {
		t.Parallel()
		b := NewSecurityBuilder()
		b.AddFilter(nil).
			AddFilterBefore(nil, nil).
			AddFilterAfter(nil, nil)
		if b == nil {
			t.Fatal("builder should remain valid")
		}
	})
}

func TestCsrfFilterEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("POST with valid token via header", func(t *testing.T) {
		t.Parallel()
		tokenRepo := NewCookieCsrfTokenRepository()
		csrfFilter := NewCsrfFilter(tokenRepo)

		req := &mockSecurityRequest{method: "GET", uri: "/api/test"}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		_ = csrfFilter.DoFilter(ctx, req, resp, chain)

		tokenVal, _ := req.GetAttribute("csrf.token")
		tokenStr, _ := tokenVal.(string)

		postReq := &mockSecurityRequest{method: "POST", uri: "/api/test", headers: map[string]string{"X-CSRF-Token": tokenStr}}
		postReq.SetAttribute("csrf.token", tokenStr)
		postResp := &mockSecurityResponse{headers: map[string]string{}}
		postChain := &mockSecurityFilterChain{}

		err := csrfFilter.DoFilter(ctx, postReq, postResp, postChain)
		if err != nil {
			t.Errorf("POST with valid token should succeed, got error: %v", err)
		}
	})

	t.Run("generateSecureToken produces different values", func(t *testing.T) {
		t.Parallel()
		token1, err := generateSecureToken(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		token2, err := generateSecureToken(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token1 == token2 {
			t.Error("expected different tokens")
		}
		if len(token1) == 0 {
			t.Error("expected non-empty token")
		}
	})
}

func TestUserDetailsService(t *testing.T) {
	t.Parallel()

	t.Run("InMemoryUserDetails", func(t *testing.T) {
		t.Parallel()
		details := NewInMemoryUserDetails("user1", "pass1", []string{"ROLE_USER"})
		if details.Username() != "user1" {
			t.Errorf("expected username 'user1', got '%s'", details.Username())
		}
		if details.Password() != "pass1" {
			t.Errorf("expected password 'pass1', got '%s'", details.Password())
		}
		roles := details.Authorities()
		if len(roles) != 1 || roles[0] != "ROLE_USER" {
			t.Errorf("expected [ROLE_USER], got %v", roles)
		}
		if !details.Enabled() {
			t.Error("expected enabled")
		}
		if !details.AccountNonExpired() {
			t.Error("expected account non-expired")
		}
		if !details.CredentialsNonExpired() {
			t.Error("expected credentials non-expired")
		}
		if !details.AccountNonLocked() {
			t.Error("expected account non-locked")
		}
	})

	t.Run("InMemoryUserDetailsService CRUD", func(t *testing.T) {
		t.Parallel()
		svc := NewInMemoryUserDetailsService()

		svc.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

		ctx := context.Background()
		details, err := svc.LoadUserByUsername(ctx, "admin")
		if err != nil {
			t.Fatalf("failed to load user: %v", err)
		}
		if details.Username() != "admin" {
			t.Errorf("expected 'admin', got '%s'", details.Username())
		}

		_, err = svc.LoadUserByUsername(ctx, "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent user")
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		t.Parallel()
		svc := NewInMemoryUserDetailsService()
		svc.CreateUser("user1", "pass", []string{"ROLE_USER"})

		if svc.UserCount() != 1 {
			t.Errorf("expected 1 user, got %d", svc.UserCount())
		}

		svc.DeleteUser("user1")
		if svc.UserCount() != 0 {
			t.Errorf("expected 0 users after delete, got %d", svc.UserCount())
		}
	})
}

func TestUsernamePasswordAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("NewUsernamePasswordAuthenticationToken", func(t *testing.T) {
		t.Parallel()
		auth := NewUsernamePasswordAuthenticationToken("user", "pass")
		if auth.Principal() != "user" {
			t.Errorf("expected principal 'user', got '%v'", auth.Principal())
		}
		if auth.Credentials() != "pass" {
			t.Errorf("expected credentials 'pass', got '%v'", auth.Credentials())
		}
		if auth.Authenticated() {
			t.Error("expected not authenticated")
		}
	})

	t.Run("NewAuthenticatedUsernamePasswordAuthenticationToken", func(t *testing.T) {
		t.Parallel()
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
		if !auth.Authenticated() {
			t.Error("expected authenticated")
		}
		if len(auth.Authorities()) != 1 {
			t.Errorf("expected 1 authority, got %d", len(auth.Authorities()))
		}
	})

	t.Run("SetAuthenticated and SetAuthorities", func(t *testing.T) {
		t.Parallel()
		auth := NewUsernamePasswordAuthenticationToken("user", "pass")
		auth.SetAuthenticated(true)
		if !auth.Authenticated() {
			t.Error("expected authenticated after SetAuthenticated(true)")
		}
		auth.SetAuthorities([]string{"ROLE_USER", "ROLE_ADMIN"})
		if len(auth.Authorities()) != 2 {
			t.Errorf("expected 2 authorities, got %d", len(auth.Authorities()))
		}
	})

	t.Run("Name returns principal string", func(t *testing.T) {
		t.Parallel()
		auth := NewUsernamePasswordAuthenticationToken("user", "pass")
		if auth.Name() != "user" {
			t.Errorf("expected name 'user', got '%s'", auth.Name())
		}
	})
}

func TestBasicAuthFilterEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("empty Authorization header", func(t *testing.T) {
		t.Parallel()
		svc := NewInMemoryUserDetailsService()
		svc.CreateUser("admin", "pass", []string{"ROLE_ADMIN"})
		enc := NewNoOpPasswordEncoder()
		provider := NewDaoAuthenticationProvider(svc, enc, log.Build())
		mgr := NewProviderManager(provider)

		filter := NewBasicAuthenticationFilterWithRealm(mgr, "Test", log.Build())
		req := &mockSecurityRequest{method: "GET", uri: "/api", headers: map[string]string{"Authorization": "Basic "}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		err := filter.DoFilter(ctx, req, resp, chain)
		if err == nil {
			t.Error("expected error for empty credentials")
		}
	})

	t.Run("non-Basic authorization header", func(t *testing.T) {
		t.Parallel()
		svc := NewInMemoryUserDetailsService()
		svc.CreateUser("admin", "pass", []string{"ROLE_ADMIN"})
		enc := NewNoOpPasswordEncoder()
		provider := NewDaoAuthenticationProvider(svc, enc, log.Build())
		mgr := NewProviderManager(provider)

		filter := NewBasicAuthenticationFilterWithRealm(mgr, "Test", log.Build())
		req := &mockSecurityRequest{method: "GET", uri: "/api", headers: map[string]string{"Authorization": "Bearer token123"}}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		err := filter.DoFilter(ctx, req, resp, chain)
		if err == nil {
			t.Error("expected error for non-Basic auth")
		}
		if resp.statusCode != 401 {
			t.Errorf("expected status 401, got %d", resp.statusCode)
		}
	})
}

func TestPasswordEncoderInterface(t *testing.T) {
	t.Parallel()

	t.Run("NoOpPasswordEncoder roundtrip", func(t *testing.T) {
		t.Parallel()
		enc := NewNoOpPasswordEncoder()
		encoded := enc.Encode("mypassword")
		if encoded != "mypassword" {
			t.Errorf("NoOp should return same string, got '%s'", encoded)
		}
		if !enc.Matches("mypassword", "mypassword") {
			t.Error("expected match")
		}
		if enc.Matches("mypassword", "wrongpassword") {
			t.Error("expected no match")
		}
	})
}

func TestSecurityRequestResponseMock(t *testing.T) {
	t.Parallel()

	t.Run("mock request methods", func(t *testing.T) {
		t.Parallel()
		req := &mockSecurityRequest{method: "POST", uri: "/test"}
		req.SetHeader("X-Custom", "value123")
		req.SetAttribute("key", "val")

		if req.GetMethod() != "POST" {
			t.Errorf("expected POST, got %s", req.GetMethod())
		}
		if req.GetHeader("X-Custom") != "value123" {
			t.Errorf("expected value123, got %s", req.GetHeader("X-Custom"))
		}
		if req.GetHeader("X-Missing") != "" {
			t.Error("expected empty string for missing header")
		}

		val, ok := req.GetAttribute("key")
		if !ok || val != "val" {
			t.Errorf("expected attribute 'val', got %v (exists=%v)", val, ok)
		}
	})

	t.Run("mock response methods", func(t *testing.T) {
		t.Parallel()
		resp := &mockSecurityResponse{headers: map[string]string{}}
		resp.SetStatusCode(404)
		resp.SetHeader("X-Test", "hello")

		if resp.statusCode != 404 {
			t.Errorf("expected 404, got %d", resp.statusCode)
		}
		if resp.headers["X-Test"] != "hello" {
			t.Errorf("expected 'hello', got '%s'", resp.headers["X-Test"])
		}
	})
}

func TestFilterChainProxy(t *testing.T) {
	t.Parallel()

	t.Run("SecurityFilterChainAdapter adapts FilterChainProxy", func(t *testing.T) {
		t.Parallel()
		chain := NewFilterChainProxy([]SecurityFilter{}, &DefaultSecurityFilterChain{})
		adapter := securityFilterChainAdapter{proxy: chain}

		filters := adapter.GetFilters()
		if len(filters) != 0 {
			t.Errorf("expected empty filters, got %d", len(filters))
		}
	})
}

func TestLogoutFilterEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("empty logout URL causes panic", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty logout URL")
			}
		}()
		NewLogoutFilter("", []LogoutHandler{})
	})

	t.Run("GET to logout URL does not logout", func(t *testing.T) {
		t.Parallel()
		req := &mockSecurityRequest{method: "GET", uri: "/logout"}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		filter := NewLogoutFilter("/logout", []LogoutHandler{})
		err := filter.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("chain should be called for GET to logout URL")
		}
	})
}

func TestCsrfTokenManagerEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("validate with wrong user", func(t *testing.T) {
		t.Parallel()
		m := NewCsrfTokenManager()
		token, _ := m.GenerateToken("user1")
		if m.ValidateToken("user2", token) {
			t.Error("token should not validate for different user")
		}
	})

	t.Run("generate token is non-empty", func(t *testing.T) {
		t.Parallel()
		m := NewCsrfTokenManager()
		token, err := m.GenerateToken("user1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(token) == 0 {
			t.Error("expected non-empty token")
		}
	})
}
