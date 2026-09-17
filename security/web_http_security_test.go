package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
)

func testHttpSecConfigBuildAllFeatures(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
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
}

func testHttpSecConfigCustomFormLoginURL(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
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
}

func testHttpSecConfigCustomLogoutURL(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
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
}

func TestHttpSecurityConfiguration(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("Build with all features", func(t *testing.T) { testHttpSecConfigBuildAllFeatures(t, authManager) })
	t.Run("Custom FormLogin URL", func(t *testing.T) { testHttpSecConfigCustomFormLoginURL(t, authManager) })
	t.Run("Custom Logout URL", func(t *testing.T) { testHttpSecConfigCustomLogoutURL(t, authManager) })
}

func testHttpSecAuthorizeRulesApplied(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	httpSec := NewHttpSecurity().
		AuthenticationManager(authManager).
		AuthorizeRequests(func(authz AuthorizeRequests) {
			authz.AntMatchers("/api/**").HasRole("ROLE_API")
			authz.AntMatchers("/admin/**").DenyAll()
			authz.AnyRequest().Authenticated()
		})

	sec := httpSec.(*httpSecurity)
	if len(sec.authorizeRules) != 3 {
		t.Fatalf("expected 3 collected rules, got %d", len(sec.authorizeRules))
	}

	chain, err := httpSec.Build()
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
}

func testHttpSecAuthorizeNoRulesBuilds(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	httpSec := NewHttpSecurity().AuthenticationManager(authManager)
	chain, err := httpSec.Build()
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}
}

func TestHttpSecurity_AuthorizeRequests_AppliesRules(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("AntMatchers and AnyRequest rules applied", func(t *testing.T) { testHttpSecAuthorizeRulesApplied(t, authManager) })
	t.Run("no rules still builds with empty source", func(t *testing.T) { testHttpSecAuthorizeNoRulesBuilds(t, authManager) })
}
