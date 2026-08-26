package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

func TestSecurityConfig_Build(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("build with minimal config", func(t *testing.T) {
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
	})

	t.Run("build with all options", func(t *testing.T) {
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
	})

	t.Run("build without auth manager fails", func(t *testing.T) {
		t.Parallel()
		cfg := NewSecurityConfig()

		_, err := cfg.Build()
		if err == nil {
			t.Fatal("expected error when auth manager is missing")
		}
	})

	t.Run("authorize rules applied to metadata source", func(t *testing.T) {
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
	})
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

func TestSecurityConfig_WithOptions(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	t.Run("exception handling config", func(t *testing.T) {
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
	})

	t.Run("csrf with custom repository", func(t *testing.T) {
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
	})

	t.Run("http basic with custom realm", func(t *testing.T) {
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
	})
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

func TestSecurityConfig_WithFilterBefore(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	customFilter := &mockSecurityFilter{}
	authFilter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/dashboard", "/login?error", authManager, log.Build())

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithFilterBefore(customFilter, authFilter),
	)

	if len(cfg.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(cfg.Filters))
	}
	if cfg.Filters[0].Before == nil {
		t.Error("expected Before filter to be set")
	}
}

func TestSecurityConfig_WithFilterAfter(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	passwordEncoder := NewNoOpPasswordEncoder()
	authProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, log.Build())
	authManager := NewProviderManager(authProvider)

	customFilter := &mockSecurityFilter{}
	authFilter := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/dashboard", "/login?error", authManager, log.Build())

	cfg := NewSecurityConfig(
		WithAuthenticationManager(authManager),
		WithFilterAfter(customFilter, authFilter),
	)

	if len(cfg.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(cfg.Filters))
	}
	if cfg.Filters[0].After == nil {
		t.Error("expected After filter to be set")
	}
}

func TestInsertFilterBefore(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filter3 := &mockSecurityFilter{order: 3}
	filters := []SecurityFilter{filter1, filter2, filter3}

	newFilter := &mockSecurityFilter{order: 0}
	result := insertFilterBefore(filters, newFilter, filter2)

	if len(result) != 4 {
		t.Fatalf("expected 4 filters, got %d", len(result))
	}
	// 验证newFilter在filter2之前
	if result[1] != newFilter {
		t.Error("expected newFilter to be before filter2")
	}
	if result[2] != filter2 {
		t.Error("expected filter2 to be at index 2")
	}
}

func TestInsertFilterBefore_NotFound(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filters := []SecurityFilter{filter1, filter2}

	newFilter := &mockSecurityFilter{order: 0}
	notFound := &mockSecurityFilter{order: 99}
	result := insertFilterBefore(filters, newFilter, notFound)

	if len(result) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(result))
	}
	// 验证newFilter被添加到末尾
	if result[2] != newFilter {
		t.Error("expected newFilter to be appended at the end")
	}
}

func TestInsertFilterAfter(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filter3 := &mockSecurityFilter{order: 3}
	filters := []SecurityFilter{filter1, filter2, filter3}

	newFilter := &mockSecurityFilter{order: 0}
	result := insertFilterAfter(filters, newFilter, filter2)

	if len(result) != 4 {
		t.Fatalf("expected 4 filters, got %d", len(result))
	}
	// 验证newFilter在filter2之后
	if result[2] != newFilter {
		t.Error("expected newFilter to be after filter2")
	}
	if result[1] != filter2 {
		t.Error("expected filter2 to be at index 1")
	}
}

func TestInsertFilterAfter_NotFound(t *testing.T) {
	t.Parallel()

	filter1 := &mockSecurityFilter{order: 1}
	filter2 := &mockSecurityFilter{order: 2}
	filters := []SecurityFilter{filter1, filter2}

	newFilter := &mockSecurityFilter{order: 0}
	notFound := &mockSecurityFilter{order: 99}
	result := insertFilterAfter(filters, newFilter, notFound)

	if len(result) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(result))
	}
	// 验证newFilter被添加到末尾
	if result[2] != newFilter {
		t.Error("expected newFilter to be appended at the end")
	}
}
