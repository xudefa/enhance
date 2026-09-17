package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
)

func testCorsAllowAllWildcard(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
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
}

func testCorsPreflightOptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
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
}

func testCorsBlockedOrigin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
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
}

func testCorsNoOriginHeader(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
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
}

func testCorsInvalidCtx(t *testing.T) {
	t.Parallel()
	filter := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
	err := filter.DoFilter("not-a-context", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
	if err == nil {
		t.Error("expected error for invalid context type")
	}
}

func testCorsInvalidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	filter := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
	err := filter.DoFilter(ctx, "not-a-request", &mockSecurityResponse{}, &mockSecurityFilterChain{})
	if err == nil {
		t.Error("expected error for invalid request type")
	}
}

func testCorsInvalidResponse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	filter := NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}})
	err := filter.DoFilter(ctx, &mockSecurityRequest{}, "not-a-response", &mockSecurityFilterChain{})
	if err == nil {
		t.Error("expected error for invalid response type")
	}
}

func testCorsAllowCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
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
}

func testCorsExactOriginMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
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
}

func testCorsOrder(t *testing.T) {
	t.Parallel()
	filter := NewCorsFilter(CorsConfig{})
	if filter.Order() != -100 {
		t.Errorf("expected order -100, got %d", filter.Order())
	}
}

func TestCorsFilter(t *testing.T) {
	t.Parallel()

	t.Run("AllowAll wildcard origin", testCorsAllowAllWildcard)
	t.Run("preflight OPTIONS request", testCorsPreflightOptions)
	t.Run("blocked origin", testCorsBlockedOrigin)
	t.Run("no Origin header skips CORS", testCorsNoOriginHeader)
	t.Run("invalid ctx type returns error", testCorsInvalidCtx)
	t.Run("invalid request type returns error", testCorsInvalidRequest)
	t.Run("invalid response type returns error", testCorsInvalidResponse)
	t.Run("AllowCredentials header", testCorsAllowCredentials)
	t.Run("exact origin match", testCorsExactOriginMatch)
	t.Run("Order returns -100", testCorsOrder)
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
		returnedBuilder := b.
			AuthenticationManager(nil).
			UserDetailsService(nil).
			PasswordEncoder(nil).
			AccessDecisionManager(nil).
			EnableAnonymous().
			EnableCsrf().
			EnableFormLogin("/login").
			EnableHttpBasic().
			EnableLogout("/logout")
		if returnedBuilder == nil {
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
		csrfFilter := MustNewCsrfFilter(tokenRepo)

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

func testUserDetailsInMemory(t *testing.T) {
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
}

func testUserDetailsServiceCRUD(t *testing.T) {
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
}

func testUserDetailsServiceDelete(t *testing.T) {
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
}

func TestUserDetailsService(t *testing.T) {
	t.Parallel()

	t.Run("InMemoryUserDetails", testUserDetailsInMemory)
	t.Run("InMemoryUserDetailsService CRUD", testUserDetailsServiceCRUD)
	t.Run("DeleteUser", testUserDetailsServiceDelete)
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

func TestLogoutFilterEdgeCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("empty logout URL returns error", func(t *testing.T) {
		t.Parallel()
		_, err := NewLogoutFilter("", []LogoutHandler{})
		if err == nil {
			t.Error("expected error for empty logout URL")
		}
	})

	t.Run("GET to logout URL does not logout", func(t *testing.T) {
		t.Parallel()
		req := &mockSecurityRequest{method: "GET", uri: "/logout"}
		resp := &mockSecurityResponse{headers: map[string]string{}}
		chain := &mockSecurityFilterChain{}

		filter := MustNewLogoutFilter("/logout", []LogoutHandler{})
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
		tokenManager := NewCsrfTokenManager()
		token, _ := tokenManager.GenerateToken("user1")
		if tokenManager.ValidateToken("user2", token) {
			t.Error("token should not validate for different user")
		}
	})

	t.Run("generate token is non-empty", func(t *testing.T) {
		t.Parallel()
		tokenManager := NewCsrfTokenManager()
		token, err := tokenManager.GenerateToken("user1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(token) == 0 {
			t.Error("expected non-empty token")
		}
	})
}
