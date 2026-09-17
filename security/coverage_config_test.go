package security

import (
	"testing"
)

// ============================================================
// filter_chain.go 测试
// ============================================================

// TestSecurityFilterChainAdapter 测试安全过滤器链适配器
func TestConfigRegistry_AddRule(t *testing.T) {
	t.Parallel()

	t.Run("empty patterns returns nil", func(t *testing.T) {
		registry := &configRegistry{cfg: &SecurityConfig{}, patterns: []string{}}
		addRuleErr := registry.addRule([]string{"permitAll"})
		if addRuleErr != nil {
			t.Error("expected nil for empty patterns")
		}
	})

	t.Run("empty attrs returns nil", func(t *testing.T) {
		registry := &configRegistry{cfg: &SecurityConfig{}, patterns: []string{"/test"}}
		addRuleErr := registry.addRule([]string{})
		if addRuleErr != nil {
			t.Error("expected nil for empty attrs")
		}
	})

	t.Run("valid rule added", func(t *testing.T) {
		cfg := &SecurityConfig{}
		registry := &configRegistry{cfg: cfg, patterns: []string{"/api/**"}}
		addRuleErr := registry.addRule([]string{"authenticated"})
		if addRuleErr != nil {
			t.Error("expected nil for valid rule")
		}
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})
}

// TestConfigAuthorizer 测试 configAuthorizer

func TestConfigAuthorizer(t *testing.T) {
	t.Parallel()

	cfg := &SecurityConfig{}
	authorizer := &configAuthorizer{cfg: cfg}

	registry := authorizer.AntMatchers("/api/**")
	if registry == nil {
		t.Error("expected non-nil registry")
	}

	anyRegistry := authorizer.AnyRequest()
	if anyRegistry == nil {
		t.Error("expected non-nil registry")
	}
}

// TestConfigRegistry_AllMethods 测试 configRegistry 的所有授权方法

func TestConfigRegistry_AllMethods(t *testing.T) {
	t.Parallel()

	t.Run("HasRole", func(t *testing.T) {
		cfg := &SecurityConfig{}
		registry := &configRegistry{cfg: cfg, patterns: []string{"/admin/**"}}
		registry.HasRole("ADMIN")
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})

	t.Run("HasAnyRole", func(t *testing.T) {
		cfg := &SecurityConfig{}
		registry := &configRegistry{cfg: cfg, patterns: []string{"/api/**"}}
		registry.HasAnyRole("ADMIN", "USER")
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})
}

// ============================================================
// rate_limit_filter.go 测试 - 补充覆盖
// ============================================================

// TestRateLimitFilter_DoFilter_InvalidTypes 测试 TokenBucket 的类型检查
