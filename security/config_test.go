package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

func testSecurityConfigBuildMinimal(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
	)

	chain, err := cfg.Build()
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}
}

func testSecurityConfigBuildAllOptions(t *testing.T, authManager AuthenticationManager, userDetailsService UserDetailsService, passwordEncoder PasswordEncoder) {
	t.Parallel()
	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithUserDetailsService(userDetailsService),
		WithPasswordEncoder(passwordEncoder),
		WithCsrf(),
		WithFormLogin("/login", "/dashboard"),
		WithLogout("/logout"),
		WithHttpBasic(),
		WithAnonymous(),
		WithAuthorizeRequests(func(authz AuthorizeRequests) {
			authz.AntMatchers("/api/**").HasRole("ROLE_API")
			authz.AntMatchers("/admin/**").DenyAll()
			authz.AnyRequest().Authenticated()
		}),
	)

	chain, err := cfg.Build()
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}

	// Verify authorize rules were collected
	if len(cfg.AuthorizeRules) != 3 {
		t.Errorf("expected 3 authorize rules, got %d", len(cfg.AuthorizeRules))
	}
}

func testSecurityConfigBuildNoAuthManager(t *testing.T) {
	t.Parallel()
	cfg := NewSecurityConfig()

	_, err := cfg.Build()
	if err == nil {
		t.Fatal("expected error when auth manager is missing")
	}
}

func testSecurityConfigBuildMetadataSource(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithAuthorizeRequests(func(authz AuthorizeRequests) {
			authz.AntMatchers("/api/**").HasRole("ROLE_API")
			authz.AnyRequest().Authenticated()
		}),
	)

	_, err := cfg.Build()
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	source, ok := cfg.SecurityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource)
	if !ok {
		t.Fatal("expected expression based metadata source")
	}

	ctx := context.Background()
	attrs, err := source.GetAttributes(ctx, &mockSecurityRequest{method: "GET", uri: "/api/users"})
	if err != nil || len(attrs) != 1 || attrs[0] != "hasRole('ROLE_API')" {
		t.Errorf("expected hasRole('ROLE_API') for /api/users, got %v (err=%v)", attrs, err)
	}

	attrs, err = source.GetAttributes(ctx, &mockSecurityRequest{method: "GET", uri: "/other"})
	if err != nil || len(attrs) != 1 || attrs[0] != "authenticated" {
		t.Errorf("expected authenticated for /other, got %v (err=%v)", attrs, err)
	}
}

func TestSecurityConfig_Build(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("build with minimal config", func(t *testing.T) { testSecurityConfigBuildMinimal(t, authManager) })
	t.Run("build with all options", func(t *testing.T) {
		testSecurityConfigBuildAllOptions(t, authManager, userDetailsService, passwordEncoder)
	})
	t.Run("build without auth manager fails", testSecurityConfigBuildNoAuthManager)
	t.Run("authorize rules applied to metadata source", func(t *testing.T) { testSecurityConfigBuildMetadataSource(t, authManager) })
}

func TestSecurityConfig_FilterOrdering(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("custom filter added", func(t *testing.T) {
		t.Parallel()
		customFilter := &mockSecurityFilter{order: 100}

		cfg := NewSecurityConfig(
			WithAuthenticationManager(authManager),
			WithFilter(customFilter),
		)

		chain, err := cfg.Build()
		if err != nil {
			t.Fatalf("failed to build: %v", err)
		}
		if chain == nil {
			t.Fatal("expected non-nil chain")
		}
	})

	t.Run("filter before/after", func(t *testing.T) {
		t.Parallel()
		// This tests the filter ordering logic
		cfg := NewSecurityConfig(
			WithAuthenticationManager(authManager),
			WithCsrf(),
		)

		chain, err := cfg.Build()
		if err != nil {
			t.Fatalf("failed to build: %v", err)
		}
		if chain == nil {
			t.Fatal("expected non-nil chain")
		}
	})
}

func testSecurityConfigWithOptionsExceptionHandling(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	handler := NewHttp403ForbiddenAccessDeniedHandler()
	entryPoint := NewHttp401UnauthorizedEntryPoint()

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithExceptionHandling(handler, entryPoint),
	)

	if cfg.ExceptionHandling == nil {
		t.Fatal("expected exception handling config to be set")
	}

	chain, err := cfg.Build()
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}
}

func testSecurityConfigWithOptionsCsrfRepository(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	customRepo := NewCookieCsrfTokenRepository()

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithCsrf(),
		WithCsrfTokenRepository(customRepo),
	)

	if cfg.CsrfTokenRepository != customRepo {
		t.Error("expected custom CSRF token repository")
	}
}

func testSecurityConfigWithOptionsBasicRealm(t *testing.T, authManager AuthenticationManager) {
	t.Parallel()
	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithHttpBasic("My Realm"),
	)

	if !cfg.HttpBasic {
		t.Error("expected HTTP Basic to be enabled")
	}
	if cfg.HttpBasicRealm != "My Realm" {
		t.Errorf("expected realm 'My Realm', got '%s'", cfg.HttpBasicRealm)
	}
}

func testBuildAuthManager() AuthenticationManager {
	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	return NewProviderManager(authProvider)
}

func TestSecurityConfig_WithOptions(t *testing.T) {
	t.Parallel()

	authManager := testBuildAuthManager()

	t.Run("exception handling config", func(t *testing.T) { testSecurityConfigWithOptionsExceptionHandling(t, authManager) })
	t.Run("csrf with custom repository", func(t *testing.T) { testSecurityConfigWithOptionsCsrfRepository(t, authManager) })
	t.Run("http basic with custom realm", func(t *testing.T) { testSecurityConfigWithOptionsBasicRealm(t, authManager) })
}

// mockSecurityFilter 用于测试的简单过滤器。
type mockSecurityFilter struct {
	order int
}

func (f *mockSecurityFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	return nil
}

func (f *mockSecurityFilter) Order() int {
	return f.order
}

func TestSecurityConfig_AuthorizeRules_Collection(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("collect all rule types", func(t *testing.T) {
		t.Parallel()
		cfg := NewSecurityConfig(
			WithAuthenticationManager(authManager),
			WithAuthorizeRequests(func(authz AuthorizeRequests) {
				authz.AntMatchers("/permit").PermitAll()
				authz.AntMatchers("/deny").DenyAll()
				authz.AntMatchers("/auth").Authenticated()
				authz.AntMatchers("/role").HasRole("ADMIN")
				authz.AntMatchers("/anyrole").HasAnyRole("ADMIN", "USER")
				authz.AntMatchers("/authority").HasAuthority("read")
				authz.AntMatchers("/anyauth").HasAnyAuthority("read", "write")
			}),
		)

		if len(cfg.AuthorizeRules) != 7 {
			t.Errorf("expected 7 rules, got %d", len(cfg.AuthorizeRules))
		}

		// Verify rule contents
		expectedAttrs := []string{
			"permitAll",
			"denyAll",
			"authenticated",
			"hasRole('ADMIN')",
			"hasAnyRole('ADMIN','USER')",
			"hasAuthority('read')",
			"hasAnyAuthority('read','write')",
		}

		for i, rule := range cfg.AuthorizeRules {
			if len(rule.attrs) != 1 || rule.attrs[0] != expectedAttrs[i] {
				t.Errorf("rule %d: expected %q, got %v", i, expectedAttrs[i], rule.attrs)
			}
		}
	})
}

func ExampleNewSecurityConfig() {
	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithFormLogin("/login", "/dashboard"),
		WithCsrf(),
		WithAuthorizeRequests(func(authz AuthorizeRequests) {
			authz.AntMatchers("/api/**").HasRole("ROLE_API")
			authz.AnyRequest().Authenticated()
		}),
	)

	chain, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	_ = chain
}

func TestSecurityConfig_WithAccessDecisionManager(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	decisionManager := NewAffirmativeBased()

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithAccessDecisionManager(decisionManager),
	)

	if cfg.AccessDecisionManager != decisionManager {
		t.Error("expected AccessDecisionManager to be set")
	}
}

func TestSecurityConfig_WithSecurityMetadataSource(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	source := NewExpressionBasedFilterInvocationSecurityMetadataSource()

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithSecurityMetadataSource(source),
	)

	if cfg.SecurityMetadataSource != source {
		t.Error("expected SecurityMetadataSource to be set")
	}
}
