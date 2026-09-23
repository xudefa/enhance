package authorization

import (
	"testing"
)

func TestExpressionBasedUrlRegistry_PermitAll(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/public/**").PermitAll()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if len(rules[0].Patterns) != 1 || rules[0].Patterns[0] != "/api/public/**" {
		t.Errorf("expected pattern /api/public/**, got %v", rules[0].Patterns)
	}
	if len(rules[0].Attributes) != 1 || rules[0].Attributes[0] != "permitAll" {
		t.Errorf("expected attribute permitAll, got %v", rules[0].Attributes)
	}
}

func TestExpressionBasedUrlRegistry_HasRole(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/admin/**").HasRole("ADMIN")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected hasRole('ADMIN'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyRole(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/manager/**").HasAnyRole("ADMIN", "MANAGER")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAnyRole('ADMIN','MANAGER')" {
		t.Errorf("expected hasAnyRole('ADMIN','MANAGER'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyAuthority_Single(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/data/**").HasAnyAuthority("read")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAuthority('read')" {
		t.Errorf("expected hasAuthority('read'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyAuthority(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/data/**").HasAnyAuthority("read", "write")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAnyAuthority('read','write')" {
		t.Errorf("expected hasAnyAuthority('read','write'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_DenyAll(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/internal/**").DenyAll()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "denyAll" {
		t.Errorf("expected denyAll, got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_Authenticated(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/secure/**").Authenticated()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "authenticated" {
		t.Errorf("expected authenticated, got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_MultipleRules(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/public/**").PermitAll()
	authz.RequestMatchers("/api/admin/**").HasRole("ADMIN")
	authz.RequestMatchers("**").Authenticated()

	rules := getRegistryRules(authz)
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "permitAll" {
		t.Errorf("expected first rule permitAll, got %v", rules[0].Attributes[0])
	}
	if rules[1].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected second rule hasRole('ADMIN'), got %v", rules[1].Attributes[0])
	}
	if rules[2].Attributes[0] != "authenticated" {
		t.Errorf("expected third rule authenticated, got %v", rules[2].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_AnyRequest(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.AnyRequest().Authenticated()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if len(rules[0].Patterns) != 1 || rules[0].Patterns[0] != "**" {
		t.Errorf("expected pattern **, got %v", rules[0].Patterns)
	}
}

func TestExpressionBasedUrlRegistry_EmptyGet(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	rules := getRegistryRules(authz)
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty registry, got %d", len(rules))
	}
}

func TestExpressionBasedUrlRegistry_SingleAuthority(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/data/**").HasAnyAuthority("read")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAuthority('read')" {
		t.Errorf("expected hasAuthority('read') for single authority, got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_SingleRole(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/admin/**").HasAnyRole("ADMIN")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected hasRole('ADMIN') for single role, got %v", rules[0].Attributes[0])
	}
}

func TestAuthorizeRequests(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/public/**").PermitAll()
	authz.RequestMatchers("/api/admin/**").HasRole("ADMIN")

	if authz == nil {
		t.Error("expected non-nil AuthorizeRequests")
	}
}

// getRegistryRules 通过类型断言从 AuthorizeRequests 获取规则。
func getRegistryRules(authz AuthorizeRequests) []UrlAuthorizationRule {
	ar := authz.(*authorizeRequests)
	return ar.registry.Get()
}
