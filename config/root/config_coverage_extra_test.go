package config

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

// TestConfig_GetProperty_NonString_Coverage 测试 GetProperty 处理非字符串值
func TestConfig_GetProperty_NonString_Coverage(t *testing.T) {
	t.Parallel()

	// 使用 Environment 直接设置非字符串值
	env := environment.NewEnvironment()
	source := environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"app.port": 8080,
	})
	env.AddPropertySource(source)

	cfg := NewConfig(env)

	// 非字符串值应该返回空字符串
	got := cfg.GetProperty("app.port")
	if got != "" {
		t.Errorf("GetProperty() for non-string = %v, want empty string", got)
	}
}

// TestConfig_GetProperty_Missing_Coverage 测试 GetProperty 处理缺失的属性
func TestConfig_GetProperty_Missing_Coverage(t *testing.T) {
	t.Parallel()

	env := environment.NewEnvironment()
	cfg := NewConfig(env)

	got := cfg.GetProperty("missing.key")
	if got != "" {
		t.Errorf("GetProperty() for missing key = %v, want empty string", got)
	}
}

// TestConfig_GetRequiredProperty_NonString_Coverage 测试 GetRequiredProperty 处理非字符串值
func TestConfig_GetRequiredProperty_NonString_Coverage(t *testing.T) {
	t.Parallel()

	env := environment.NewEnvironment()
	source := environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"app.port": 8080,
	})
	env.AddPropertySource(source)

	cfg := NewConfig(env)

	_, err := cfg.GetRequiredProperty("app.port")
	if err == nil {
		t.Error("GetRequiredProperty() should return error for non-string property")
	}
}

// TestConfig_GetPropertyWithDefault_Existing_Coverage 测试 GetPropertyWithDefault 处理已存在的属性
func TestConfig_GetPropertyWithDefault_Existing_Coverage(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name": "my-app",
	})
	cfg := NewConfig(env)

	got := cfg.GetPropertyWithDefault("app.name", "default")
	if got != "my-app" {
		t.Errorf("GetPropertyWithDefault() = %v, want my-app", got)
	}
}

// TestConfig_ContainsProperty_False_Coverage 测试 ContainsProperty 处理不存在的属性
func TestConfig_ContainsProperty_False_Coverage(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	cfg := NewConfig(env)

	if cfg.ContainsProperty("nonexistent") {
		t.Error("ContainsProperty() should return false for nonexistent property")
	}
}
