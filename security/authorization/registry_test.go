package authorization

import (
	"testing"
)

func TestNewExpressionBasedUrlRegistry(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestExpressionBasedUrlRegistry_PermitAllRule(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/public/**").PermitAll()

	rules := reg.Get()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "permitAll" {
		t.Errorf("expected permitAll, got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_DenyAllRule(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/internal/**").DenyAll()

	rules := reg.Get()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "denyAll" {
		t.Errorf("expected denyAll, got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_AuthenticatedRule(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/secure/**").Authenticated()

	rules := reg.Get()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "authenticated" {
		t.Errorf("expected authenticated, got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasRoleRule(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/admin/**").HasRole("ADMIN")

	rules := reg.Get()
	if rules[0].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected hasRole('ADMIN'), got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyRoleRule_Multiple(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/manager/**").HasAnyRole("ADMIN", "MANAGER")

	rules := reg.Get()
	if rules[0].Attributes[0] != "hasAnyRole('ADMIN','MANAGER')" {
		t.Errorf("expected hasAnyRole('ADMIN','MANAGER'), got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyRoleRule_Single(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/manager/**").HasAnyRole("ADMIN")

	rules := reg.Get()
	if rules[0].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected hasRole('ADMIN'), got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyAuthorityRule_Multiple(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/data/**").HasAnyAuthority("read", "write")

	rules := reg.Get()
	if rules[0].Attributes[0] != "hasAnyAuthority('read','write')" {
		t.Errorf("expected hasAnyAuthority('read','write'), got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyAuthorityRule_Single(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/data/**").HasAnyAuthority("read")

	rules := reg.Get()
	if rules[0].Attributes[0] != "hasAuthority('read')" {
		t.Errorf("expected hasAuthority('read'), got %s", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_MultipleRulesFunc(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/public/**").PermitAll()
	wrapper.requestMatchers("/api/admin/**").HasRole("ADMIN")
	wrapper.requestMatchers("**").Authenticated()

	rules := reg.Get()
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
}

func TestExpressionBasedUrlRegistry_EmptyGetFunc(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	rules := reg.Get()
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty registry, got %d", len(rules))
	}
}

func TestExpressionBasedUrlRegistry_AndFunc(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.requestMatchers("/api/**").PermitAll().And()
	wrapper.requestMatchers("/admin/**").HasRole("ADMIN")

	rules := reg.Get()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestNewAuthorizeRequestsFunc(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	if authz == nil {
		t.Fatal("expected non-nil AuthorizeRequests")
	}
}

func TestAuthorizeRequests_RequestMatchersFunc(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	builder := authz.RequestMatchers("/api/**")
	if builder == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestAuthorizeRequests_AnyRequestFunc(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	builder := authz.AnyRequest()
	if builder == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestJoinStringsFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{"empty", []string{}, ",", ""},
		{"single", []string{"a"}, ",", "a"},
		{"multiple", []string{"a", "b", "c"}, ",", "a,b,c"},
		{"different sep", []string{"a", "b"}, " | ", "a | b"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpressionBasedUrlRegistry_CommitNoAttrsFunc(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.registry.current = &currentRule{
		patterns: []string{"/api/**"},
		attrs:    []string{},
	}
	wrapper.registry.commit()

	rules := reg.Get()
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty attrs, got %d", len(rules))
	}
}

func TestExpressionBasedUrlRegistry_CommitNoPatternsFunc(t *testing.T) {
	t.Parallel()

	reg := NewExpressionBasedUrlRegistry()
	wrapper := reg.(*registryWrapper)
	wrapper.registry.current = &currentRule{
		patterns: []string{},
		attrs:    []string{"permitAll"},
	}
	wrapper.registry.commit()

	rules := reg.Get()
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty patterns, got %d", len(rules))
	}
}
