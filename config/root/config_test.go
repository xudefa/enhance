package config

import (
	"github.com/xudefa/enhance/config/environment"
	"testing"
)

func TestConfig_GetProperty(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})

	cfg := NewConfig(env)

	got := cfg.GetProperty("app.name")
	if got != "test-app" {
		t.Errorf("GetProperty() = %v, want test-app", got)
	}
}

func TestConfig_GetPropertyWithDefault(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	cfg := NewConfig(env)

	got := cfg.GetPropertyWithDefault("missing", "default")
	if got != "default" {
		t.Errorf("GetPropertyWithDefault() = %v, want default", got)
	}
}

func TestConfig_ContainsProperty(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})
	cfg := NewConfig(env)

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"existing property", "app.name", true},
		{"missing property", "missing", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := cfg.ContainsProperty(tt.key)
			if got != tt.expected {
				t.Errorf("ContainsProperty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_GetRequiredProperty(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})
	cfg := NewConfig(env)

	t.Run("existing property", func(t *testing.T) {
		t.Parallel()
		got, err := cfg.GetRequiredProperty("app.name")
		if err != nil {
			t.Errorf("GetRequiredProperty() error = %v", err)
		}
		if got != "test-app" {
			t.Errorf("GetRequiredProperty() = %v, want test-app", got)
		}
	})

	t.Run("missing property", func(t *testing.T) {
		t.Parallel()
		_, err := cfg.GetRequiredProperty("missing")
		if err == nil {
			t.Error("GetRequiredProperty() should return error for missing property")
		}
	})
}

func TestConfig_Environment(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	cfg := NewConfig(env)

	if cfg.Environment() != env {
		t.Error("Environment() should return the same environment instance")
	}
}

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
