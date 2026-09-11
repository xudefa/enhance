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
		r := &configRegistry{cfg: &SecurityConfig{}, patterns: []string{}}
		result := r.addRule([]string{"permitAll"})
		if result != nil {
			t.Error("expected nil for empty patterns")
		}
	})

	t.Run("empty attrs returns nil", func(t *testing.T) {
		r := &configRegistry{cfg: &SecurityConfig{}, patterns: []string{"/test"}}
		result := r.addRule([]string{})
		if result != nil {
			t.Error("expected nil for empty attrs")
		}
	})

	t.Run("valid rule added", func(t *testing.T) {
		cfg := &SecurityConfig{}
		r := &configRegistry{cfg: cfg, patterns: []string{"/api/**"}}
		result := r.addRule([]string{"authenticated"})
		if result != nil {
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
		r := &configRegistry{cfg: cfg, patterns: []string{"/admin/**"}}
		r.HasRole("ADMIN")
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})

	t.Run("HasAnyRole", func(t *testing.T) {
		cfg := &SecurityConfig{}
		r := &configRegistry{cfg: cfg, patterns: []string{"/api/**"}}
		r.HasAnyRole("ADMIN", "USER")
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})
}

// ============================================================
// rate_limit_filter.go 测试 - 补充覆盖
// ============================================================

// TestRateLimitFilter_DoFilter_InvalidTypes 测试 TokenBucket 的类型检查
