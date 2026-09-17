package security

import (
	"testing"
)

// TestExpressionInterceptUrlRegistry 测试 URL 拦截注册的各种方法

func testRegistryPermitAll(t *testing.T) {
	t.Parallel()
	httpSec := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/public/**"}}
	registry.PermitAll()
	if len(httpSec.authorizeRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(httpSec.authorizeRules))
	}
	if httpSec.authorizeRules[0].attrs[0] != "permitAll" {
		t.Errorf("expected 'permitAll', got '%s'", httpSec.authorizeRules[0].attrs[0])
	}
}

func testRegistryAuthenticated(t *testing.T) {
	t.Parallel()
	httpSec := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/api/**"}}
	registry.Authenticated()
	if len(httpSec.authorizeRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(httpSec.authorizeRules))
	}
}

func testRegistryHasRole(t *testing.T) {
	t.Parallel()
	httpSec := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/admin/**"}}
	registry.HasRole("ADMIN")
	if len(httpSec.authorizeRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(httpSec.authorizeRules))
	}
}

func testRegistryHasAnyRole(t *testing.T) {
	t.Parallel()
	httpSec := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/api/**"}}
	registry.HasAnyRole("ADMIN", "USER")
	if len(httpSec.authorizeRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(httpSec.authorizeRules))
	}
}

func testRegistryHasAuthority(t *testing.T) {
	t.Parallel()
	httpSec := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/api/**"}}
	registry.HasAuthority("READ")
	if len(httpSec.authorizeRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(httpSec.authorizeRules))
	}
}

func testRegistryHasAnyAuthority(t *testing.T) {
	t.Parallel()
	httpSec := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/api/**"}}
	registry.HasAnyAuthority("READ", "WRITE")
	if len(httpSec.authorizeRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(httpSec.authorizeRules))
	}
}

func testRegistryDenyAll(t *testing.T) {
	t.Parallel()
	httpSec := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/secret/**"}}
	registry.DenyAll()
	if len(httpSec.authorizeRules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(httpSec.authorizeRules))
	}
}

func testRegistryAddRuleEmptyPatterns(t *testing.T, httpSec *httpSecurity) {
	t.Parallel()
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{}}
	addRuleErr := registry.addRule([]string{"permitAll"})
	if addRuleErr == nil {
		t.Error("expected non-nil result")
	}
}

func testRegistryAddRuleEmptyAttrs(t *testing.T, httpSec *httpSecurity) {
	t.Parallel()
	registry := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/test"}}
	addRuleErr := registry.addRule([]string{})
	if addRuleErr == nil {
		t.Error("expected non-nil result")
	}
}

func testRegistryAnyRequest(t *testing.T, httpSec *httpSecurity) {
	t.Parallel()
	authorizer := &httpSecurityAuthorizer{httpSecurity: httpSec}
	reg := authorizer.AnyRequest()
	if reg == nil {
		t.Error("expected non-nil registry")
	}
}

func TestExpressionInterceptUrlRegistry(t *testing.T) {
	t.Parallel()

	httpSec := &httpSecurity{
		authorizeRules: make([]authorizeRule, 0),
	}

	t.Run("PermitAll", testRegistryPermitAll)
	t.Run("Authenticated", testRegistryAuthenticated)
	t.Run("HasRole", testRegistryHasRole)
	t.Run("HasAnyRole", testRegistryHasAnyRole)
	t.Run("HasAuthority", testRegistryHasAuthority)
	t.Run("HasAnyAuthority", testRegistryHasAnyAuthority)
	t.Run("DenyAll", testRegistryDenyAll)
	t.Run("addRule with empty patterns", func(t *testing.T) { testRegistryAddRuleEmptyPatterns(t, httpSec) })
	t.Run("addRule with empty attrs", func(t *testing.T) { testRegistryAddRuleEmptyAttrs(t, httpSec) })
	t.Run("AnyRequest", func(t *testing.T) { testRegistryAnyRequest(t, httpSec) })
}
