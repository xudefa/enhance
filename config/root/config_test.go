package config

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestConfig_GetProperty(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})

	cfg := NewConfig(env)

	result := cfg.GetProperty("app.name")
	if result != "test-app" {
		t.Errorf("GetProperty() = %v, want test-app", result)
	}
}

func TestConfig_GetPropertyWithDefault(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	cfg := NewConfig(env)

	result := cfg.GetPropertyWithDefault("missing", "default")
	if result != "default" {
		t.Errorf("GetPropertyWithDefault() = %v, want default", result)
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
			result := cfg.ContainsProperty(tt.key)
			if result != tt.expected {
				t.Errorf("ContainsProperty() = %v, want %v", result, tt.expected)
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
		result, err := cfg.GetRequiredProperty("app.name")
		if err != nil {
			t.Errorf("GetRequiredProperty() error = %v", err)
		}
		if result != "test-app" {
			t.Errorf("GetRequiredProperty() = %v, want test-app", result)
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
