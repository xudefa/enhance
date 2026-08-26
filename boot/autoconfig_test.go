package boot

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/condition"
)

type mockAutoConfig struct{}

func (m *mockAutoConfig) Configure(ctx ApplicationContext) error { return nil }

type mockConditionContext struct{}

func (m *mockConditionContext) Environment() condition.EnvironmentAccessor { return nil }
func (m *mockConditionContext) Container() condition.ContainerAccessor     { return nil }
func (m *mockConditionContext) GetBeanByType(t reflect.Type) (any, bool)   { return nil, false }
func (m *mockConditionContext) HasProperty(key string) bool                { return false }
func (m *mockConditionContext) GetProperty(key string) (any, bool)         { return nil, false }

// ==================== AutoConfigEntry Options 测试 ====================

func TestWithDependsOn(t *testing.T) {
	t.Parallel()
	entry := &AutoConfigEntry{Config: &mockAutoConfig{}}

	opt := WithDependsOn("dbConfig", "cacheConfig")
	opt(entry)

	if len(entry.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(entry.Dependencies))
	}
	if entry.Dependencies[0] != "dbConfig" || entry.Dependencies[1] != "cacheConfig" {
		t.Errorf("expected [dbConfig, cacheConfig], got %v", entry.Dependencies)
	}
}

func TestWithDependsOn_Append(t *testing.T) {
	t.Parallel()
	entry := &AutoConfigEntry{
		Config:       &mockAutoConfig{},
		Dependencies: []string{"existing"},
	}

	opt := WithDependsOn("newDep")
	opt(entry)

	if len(entry.Dependencies) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(entry.Dependencies))
	}
	if entry.Dependencies[0] != "existing" || entry.Dependencies[1] != "newDep" {
		t.Errorf("expected [existing, newDep], got %v", entry.Dependencies)
	}
}

func TestWithOverride(t *testing.T) {
	t.Parallel()
	entry := &AutoConfigEntry{Config: &mockAutoConfig{}}

	opt := WithOverride("*GinAutoConfiguration")
	opt(entry)

	if !entry.Override {
		t.Error("expected Override to be true")
	}
	if entry.OverrideTarget != "*GinAutoConfiguration" {
		t.Errorf("expected OverrideTarget '*GinAutoConfiguration', got '%s'", entry.OverrideTarget)
	}
	if entry.Order != -100 {
		t.Errorf("expected Order -100, got %d", entry.Order)
	}
}

func TestWithOrder(t *testing.T) {
	t.Parallel()
	entry := &AutoConfigEntry{Config: &mockAutoConfig{}}

	opt := WithOrder(500)
	opt(entry)

	if entry.Order != 500 {
		t.Errorf("expected Order 500, got %d", entry.Order)
	}
}

func TestWithBefore(t *testing.T) {
	t.Parallel()
	entry := &AutoConfigEntry{Config: &mockAutoConfig{}}

	opt := WithBefore("webConfig", "securityConfig")
	opt(entry)

	if len(entry.Before) != 2 {
		t.Fatalf("expected 2 Before entries, got %d", len(entry.Before))
	}
	if entry.Before[0] != "webConfig" {
		t.Errorf("expected Before[0] 'webConfig', got '%s'", entry.Before[0])
	}
}

func TestWithAfter(t *testing.T) {
	t.Parallel()
	entry := &AutoConfigEntry{Config: &mockAutoConfig{}}

	opt := WithAfter("dbConfig")
	opt(entry)

	if len(entry.After) != 1 {
		t.Fatalf("expected 1 After entry, got %d", len(entry.After))
	}
	if entry.After[0] != "dbConfig" {
		t.Errorf("expected After[0] 'dbConfig', got '%s'", entry.After[0])
	}
}

func TestWithConditions(t *testing.T) {
	t.Parallel()
	entry := &AutoConfigEntry{Config: &mockAutoConfig{}}

	opt := WithConditions(condition.OnProperty("key", "value"))
	opt(entry)

	if len(entry.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(entry.Conditions))
	}
}

func TestRegisterAutoConfigWith(t *testing.T) {
	// 不并发，使用全局注册表
	cfg := &mockAutoConfig{}
	before := len(GlobalRegistry().GetAll())

	RegisterAutoConfigWith(cfg,
		WithOrder(100),
		WithDependsOn("db"),
		WithConditions(condition.OnProperty("app.enabled", "true")),
	)

	after := len(GlobalRegistry().GetAll())
	if after != before+1 {
		t.Fatalf("expected %d entries, got %d", before+1, after)
	}
}

// ==================== GetMatchingWithExclude 测试 ====================

type mockAutoConfig2 struct{}

func (m *mockAutoConfig2) Configure(ctx ApplicationContext) error { return nil }

func TestGetMatchingWithExclude_OverrideFiltering(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 原始配置
	registry.Add(AutoConfigEntry{
		Config: &mockAutoConfig{},
		Order:  1,
	})

	// 覆盖配置：OverrideTarget 使用完整包路径（reflect.TypeOf 的输出格式）
	// reflect.TypeOf(&mockAutoConfig{}).String() = "*boot.mockAutoConfig"
	registry.Add(AutoConfigEntry{
		Config:         &mockAutoConfig2{},
		Order:          0,
		Override:       true,
		OverrideTarget: "*boot.mockAutoConfig",
	})

	ctx := &mockConditionContext{}
	matched := registry.GetMatching(ctx)

	// 原始 *mockAutoConfig 被覆盖过滤掉，只剩覆盖配置
	if len(matched) != 1 {
		t.Fatalf("expected 1 matching config (override should filter original), got %d", len(matched))
	}
	if matched[0].Config != registry.GetAll()[1].Config {
		t.Error("expected the override config to survive")
	}
}

func TestGetMatchingWithExclude_ExcludedTypes(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	registry.Add(AutoConfigEntry{
		Config: &mockAutoConfig{},
		Order:  1,
	})

	// 排除列表使用完整包路径（同 reflect.TypeOf().String() 格式）
	typeName := "*boot.mockAutoConfig"

	ctx := &mockConditionContext{}
	matched := registry.GetMatchingWithExclude(ctx, []string{typeName})

	if len(matched) != 0 {
		t.Fatalf("expected 0 matching configs after exclusion, got %d", len(matched))
	}
}

func TestGetMatchingWithExclude_BeforeAfterSorting(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	cfgA := &mockAutoConfig{}
	cfgB := &mockAutoConfig{}
	cfgC := &mockAutoConfig{}

	// 使用完整包路径名称
	typeName := "boot.mockAutoConfig"

	registry.Add(AutoConfigEntry{Config: cfgB, Order: 2})
	registry.Add(AutoConfigEntry{
		Config: cfgA,
		Order:  1,
		Before: []string{typeName},
	})
	registry.Add(AutoConfigEntry{
		Config: cfgC,
		Order:  3,
		After:  []string{typeName},
	})

	ctx := &mockConditionContext{}
	matched := registry.GetMatching(ctx)

	if len(matched) < 2 {
		t.Fatalf("expected at least 2 matched configs, got %d", len(matched))
	}
}

// ==================== stripPackagePath 测试 ====================

func TestStripPackagePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"*github.com/xudefa/enhance/examples/gin.GinAutoConfiguration", "GinAutoConfiguration"},
		{"SimpleType", "SimpleType"},
		{"pkg.Type", "Type"},
		{"*pkg.Type", "Type"},
		{"", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := stripPackagePath(tt.input)
			if result != tt.expected {
				t.Errorf("stripPackagePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== Add Panics on Nil Config ====================

func TestAutoConfigRegistry_Add_NilConfig_Panics(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when adding nil config")
		}
	}()

	registry.Add(AutoConfigEntry{Config: nil})
}
