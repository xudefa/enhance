package boot

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/condition"
)

// TestConfigA 测试用配置 A
type TestConfigA struct{}

func (c *TestConfigA) Configure(ctx ApplicationContext) error { return nil }

// TestConfigB 测试用配置 B
type TestConfigB struct{}

func (c *TestConfigB) Configure(ctx ApplicationContext) error { return nil }

// TestConfigC 测试用配置 C
type TestConfigC struct{}

func (c *TestConfigC) Configure(ctx ApplicationContext) error { return nil }

// TestAutoConfigRegistry_BeforeAfter 测试 Before/After 排序
func TestAutoConfigRegistry_BeforeAfter(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 注册配置：A 应该在 B 之前，B 应该在 C 之前
	registry.Add(AutoConfigEntry{
		Config: &TestConfigA{},
		Before: []string{"*boot.TestConfigB"},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigB{},
		Before: []string{"*boot.TestConfigC"},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigC{},
		Order:  100,
	})

	// 获取匹配的配置
	matched := registry.GetMatching(nil)

	if len(matched) != 3 {
		t.Fatalf("Expected 3 matched configs, got %d", len(matched))
	}

	// 验证顺序：A -> B -> C
	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	if typeName(matched[0]) != "*boot.TestConfigA" {
		t.Errorf("Expected TestConfigA first, got %s", typeName(matched[0]))
	}
	if typeName(matched[1]) != "*boot.TestConfigB" {
		t.Errorf("Expected TestConfigB second, got %s", typeName(matched[1]))
	}
	if typeName(matched[2]) != "*boot.TestConfigC" {
		t.Errorf("Expected TestConfigC third, got %s", typeName(matched[2]))
	}
}

// TestDatabaseConfig 测试用数据库配置
type TestDatabaseConfig struct{}

func (c *TestDatabaseConfig) Configure(ctx ApplicationContext) error { return nil }

// TestWebConfig 测试用 Web 配置
type TestWebConfig struct{}

func (c *TestWebConfig) Configure(ctx ApplicationContext) error { return nil }

// TestCacheConfig 测试用缓存配置
type TestCacheConfig struct{}

func (c *TestCacheConfig) Configure(ctx ApplicationContext) error { return nil }

// TestAutoConfigRegistry_After 测试 After 排序
func TestAutoConfigRegistry_After(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 注册配置：Web 和 Cache 都应在 Database 之后
	registry.Add(AutoConfigEntry{
		Config: &TestDatabaseConfig{},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestWebConfig{},
		After:  []string{"*boot.TestDatabaseConfig"},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestCacheConfig{},
		After:  []string{"*boot.TestDatabaseConfig"},
		Order:  100,
	})

	// 获取匹配的配置
	matched := registry.GetMatching(nil)

	if len(matched) != 3 {
		t.Fatalf("Expected 3 matched configs, got %d", len(matched))
	}

	// 验证 Database 在第一位
	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	if typeName(matched[0]) != "*boot.TestDatabaseConfig" {
		t.Errorf("Expected TestDatabaseConfig first, got %s", typeName(matched[0]))
	}
}

// TestAutoConfigRegistry_CircularDependency 测试循环依赖回退
func TestAutoConfigRegistry_CircularDependency(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 创建循环依赖：A 在 B 之前，B 在 A 之前
	registry.Add(AutoConfigEntry{
		Config: &TestConfigA{},
		Before: []string{"*boot.TestConfigB"},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigB{},
		Before: []string{"*boot.TestConfigA"},
		Order:  50, // B 的 Order 更小，应该排在前面
	})

	// 获取匹配的配置（应该回退到 Order 排序）
	matched := registry.GetMatching(nil)

	if len(matched) != 2 {
		t.Fatalf("Expected 2 matched configs, got %d", len(matched))
	}

	// 验证回退到 Order 排序
	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	// B 的 Order=50 应该排在 A 的 Order=100 之前
	if typeName(matched[0]) != "*boot.TestConfigB" {
		t.Errorf("Expected TestConfigB first (lower Order), got %s", typeName(matched[0]))
	}
}

// TestAutoConfigRegistry_WithBeforeAfter 测试 WithBefore/WithAfter 选项
func TestAutoConfigRegistry_WithBeforeAfter(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 使用选项函数注册
	registry.Add(AutoConfigEntry{
		Config: &TestConfigA{},
		Before: []string{"*boot.TestConfigB"},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigB{},
		After:  []string{"*boot.TestConfigA"},
		Order:  100,
	})

	// 获取匹配的配置
	matched := registry.GetMatching(nil)

	if len(matched) != 2 {
		t.Fatalf("Expected 2 matched configs, got %d", len(matched))
	}

	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	// 验证 A 在 B 之前
	if typeName(matched[0]) != "*boot.TestConfigA" {
		t.Errorf("Expected TestConfigA first, got %s", typeName(matched[0]))
	}
}

// TestConfigDB 测试用数据库配置
type TestConfigDB struct{}

func (c *TestConfigDB) Configure(ctx ApplicationContext) error { return nil }

// TestConfigRedis 测试用 Redis 配置
type TestConfigRedis struct{}

func (c *TestConfigRedis) Configure(ctx ApplicationContext) error { return nil }

// TestConfigWeb 测试用 Web 配置
type TestConfigWeb struct{}

func (c *TestConfigWeb) Configure(ctx ApplicationContext) error { return nil }

// TestConfigCache 测试用缓存配置
type TestConfigCache struct{}

func (c *TestConfigCache) Configure(ctx ApplicationContext) error { return nil }

// TestAutoConfigRegistry_ComplexDependencies 测试复杂依赖关系
func TestAutoConfigRegistry_ComplexDependencies(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 复杂依赖：
	// DB 和 Redis 无依赖
	// Web 依赖 DB 和 Redis
	// Cache 依赖 Redis
	registry.Add(AutoConfigEntry{
		Config: &TestConfigDB{},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigRedis{},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigWeb{},
		After:  []string{"*boot.TestConfigDB", "*boot.TestConfigRedis"},
		Order:  100,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigCache{},
		After:  []string{"*boot.TestConfigRedis"},
		Order:  100,
	})

	// 获取匹配的配置
	matched := registry.GetMatching(nil)

	if len(matched) != 4 {
		t.Fatalf("Expected 4 matched configs, got %d", len(matched))
	}

	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	// 验证 DB 和 Redis 在 Web 和 Cache 之前
	dbIndex := -1
	redisIndex := -1
	webIndex := -1
	cacheIndex := -1

	for i, entry := range matched {
		name := typeName(entry)
		switch name {
		case "*boot.TestConfigDB":
			dbIndex = i
		case "*boot.TestConfigRedis":
			redisIndex = i
		case "*boot.TestConfigWeb":
			webIndex = i
		case "*boot.TestConfigCache":
			cacheIndex = i
		}
	}

	if dbIndex > webIndex {
		t.Error("DB should be before Web")
	}
	if redisIndex > webIndex {
		t.Error("Redis should be before Web")
	}
	if redisIndex > cacheIndex {
		t.Error("Redis should be before Cache")
	}
}

// TestWithBefore 测试 WithBefore 选项函数
func TestWithBefore(t *testing.T) {
	t.Parallel()
	entry := AutoConfigEntry{}
	opt := WithBefore("configA", "configB")
	opt(&entry)

	if len(entry.Before) != 2 {
		t.Errorf("Expected 2 before configs, got %d", len(entry.Before))
	}
	if entry.Before[0] != "configA" {
		t.Errorf("Expected 'configA', got %s", entry.Before[0])
	}
	if entry.Before[1] != "configB" {
		t.Errorf("Expected 'configB', got %s", entry.Before[1])
	}
}

// TestWithAfter 测试 WithAfter 选项函数
func TestWithAfter(t *testing.T) {
	t.Parallel()
	entry := AutoConfigEntry{}
	opt := WithAfter("configA", "configB")
	opt(&entry)

	if len(entry.After) != 2 {
		t.Errorf("Expected 2 after configs, got %d", len(entry.After))
	}
	if entry.After[0] != "configA" {
		t.Errorf("Expected 'configA', got %s", entry.After[0])
	}
	if entry.After[1] != "configB" {
		t.Errorf("Expected 'configB', got %s", entry.After[1])
	}
}

// TestDefaultConfig 测试用默认配置
type TestDefaultConfig struct{}

func (c *TestDefaultConfig) Configure(ctx ApplicationContext) error { return nil }

// TestCustomConfig 测试用自定义配置
type TestCustomConfig struct{}

func (c *TestCustomConfig) Configure(ctx ApplicationContext) error { return nil }

// TestOtherConfig 测试用其他配置
type TestOtherConfig struct{}

func (c *TestOtherConfig) Configure(ctx ApplicationContext) error { return nil }

// TestAutoConfigRegistry_OverrideWithBeforeAfter 测试 Override 与 Before/After 结合
func TestAutoConfigRegistry_OverrideWithBeforeAfter(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 注册默认配置
	registry.Add(AutoConfigEntry{
		Config: &TestDefaultConfig{},
		Order:  100,
	})

	// 注册覆盖配置（覆盖 TestDefaultConfig）
	registry.Add(AutoConfigEntry{
		Config:         &TestCustomConfig{},
		Override:       true,
		OverrideTarget: "*boot.TestDefaultConfig",
		Before:         []string{"*boot.TestOtherConfig"},
		Order:          -100,
	})

	// 注册其他配置
	registry.Add(AutoConfigEntry{
		Config: &TestOtherConfig{},
		Order:  200,
	})

	// 获取匹配的配置
	matched := registry.GetMatching(nil)

	if len(matched) != 2 {
		t.Fatalf("Expected 2 matched configs (TestDefaultConfig should be overridden), got %d", len(matched))
	}

	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	// 验证 TestDefaultConfig 被覆盖
	for _, entry := range matched {
		if typeName(entry) == "*boot.TestDefaultConfig" {
			t.Error("TestDefaultConfig should be overridden")
		}
	}

	// 验证 TestCustomConfig 在 TestOtherConfig 之前
	if typeName(matched[0]) != "*boot.TestCustomConfig" {
		t.Errorf("Expected TestCustomConfig first, got %s", typeName(matched[0]))
	}
}

// TestAutoConfigRegistry_EmptyBeforeAfter 测试空的 Before/After
func TestAutoConfigRegistry_EmptyBeforeAfter(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	registry.Add(AutoConfigEntry{
		Config: &TestConfigA{},
		Order:  200,
	})
	registry.Add(AutoConfigEntry{
		Config: &TestConfigB{},
		Order:  100,
	})

	// 获取匹配的配置（应该按 Order 排序）
	matched := registry.GetMatching(nil)

	if len(matched) != 2 {
		t.Fatalf("Expected 2 matched configs, got %d", len(matched))
	}

	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	// B 的 Order=100 应该排在 A 的 Order=200 之前
	if typeName(matched[0]) != "*boot.TestConfigB" {
		t.Errorf("Expected TestConfigB first (Order=100), got %s", typeName(matched[0]))
	}
}

// TestEnabledConfig 测试用启用配置
type TestEnabledConfig struct{}

func (c *TestEnabledConfig) Configure(ctx ApplicationContext) error { return nil }

// TestDisabledConfig 测试用禁用配置
type TestDisabledConfig struct{}

func (c *TestDisabledConfig) Configure(ctx ApplicationContext) error { return nil }

// TestAutoConfigRegistry_ConditionWithBeforeAfter 测试条件与 Before/After 结合
func TestAutoConfigRegistry_ConditionWithBeforeAfter(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()

	// 注册带条件的配置
	registry.Add(AutoConfigEntry{
		Config:     &TestEnabledConfig{},
		Conditions: []condition.Condition{condition.OnProperty("feature.enabled", "true")},
		Before:     []string{"*boot.TestDisabledConfig"},
		Order:      100,
	})
	registry.Add(AutoConfigEntry{
		Config:     &TestDisabledConfig{},
		Conditions: []condition.Condition{condition.OnProperty("feature.disabled", "true")},
		Order:      100,
	})

	// 创建 mock 条件上下文（满足 TestEnabledConfig 的条件）
	ctx := &mockConditionContextWithProps{
		properties: map[string]any{
			"feature.enabled": "true",
		},
	}

	// 获取匹配的配置
	matched := registry.GetMatching(ctx)

	// 只应该匹配 TestEnabledConfig
	if len(matched) != 1 {
		t.Fatalf("Expected 1 matched config, got %d", len(matched))
	}

	typeName := func(entry AutoConfigEntry) string {
		return reflect.TypeOf(entry.Config).String()
	}

	if typeName(matched[0]) != "*boot.TestEnabledConfig" {
		t.Errorf("Expected TestEnabledConfig, got %s", typeName(matched[0]))
	}
}

// mockConditionContextWithProps 带属性的 mock 条件上下文
type mockConditionContextWithProps struct {
	properties map[string]any
}

func (m *mockConditionContextWithProps) Environment() condition.EnvironmentAccessor {
	return &mockEnvAccessor{props: m.properties}
}

func (m *mockConditionContextWithProps) Container() condition.ContainerAccessor {
	return nil
}

func (m *mockConditionContextWithProps) GetBeanByType(t reflect.Type) (any, bool) {
	return nil, false
}

func (m *mockConditionContextWithProps) HasProperty(key string) bool {
	_, ok := m.properties[key]
	return ok
}

func (m *mockConditionContextWithProps) GetProperty(key string) (any, bool) {
	val, ok := m.properties[key]
	return val, ok
}

// mockEnvAccessor mock 环境访问器
type mockEnvAccessor struct {
	props map[string]any
}

func (m *mockEnvAccessor) GetProperty(key string) (any, bool) {
	val, ok := m.props[key]
	return val, ok
}
