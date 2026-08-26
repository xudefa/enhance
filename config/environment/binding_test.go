package environment

import (
	"testing"
)

func TestEnvironment_Bind(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"name": "test-app",
		"port": "8080",
	})

	type Config struct {
		Name string
		Port int
	}

	cfg := &Config{}
	err := env.Bind(cfg)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
}

func TestEnvironment_Bind_NilPointer(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"name": "test-app",
	})

	err := env.Bind(nil)
	if err == nil {
		t.Error("expected error for nil pointer")
	}
}

func TestEnvironment_BindKey(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})

	var name string
	err := env.BindKey("app.name", &name)
	if err != nil {
		t.Fatalf("BindKey failed: %v", err)
	}
	if name != "test-app" {
		t.Errorf("expected 'test-app', got %s", name)
	}
}

func TestEnvironment_BindKey_NotFound(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{})

	var name string
	err := env.BindKey("nonexistent", &name)
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestEnvironment_BindPrefix(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
		"app.port": "8080",
	})

	type AppConfig struct {
		Name string
		Port int
	}

	cfg := &AppConfig{}
	err := env.BindPrefix("app", cfg)
	if err != nil {
		t.Fatalf("BindPrefix failed: %v", err)
	}
}

func TestEnvironment_Validate(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"name": "test",
	})

	errs := env.Validate()
	if len(errs) > 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestEnvironment_ResolvePlaceholders(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "myapp",
		"app.url":  "http://${app.name}.example.com",
	})

	result := env.ResolvePlaceholders("${app.url}")
	if result == "" {
		t.Error("expected resolved placeholder")
	}
}

func TestEnvironment_ResolvePlaceholders_WithDefault(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{})

	result := env.ResolvePlaceholders("${missing:default-value}")
	if result != "default-value" {
		t.Errorf("expected 'default-value', got %s", result)
	}
}

func TestEnvironment_ResolvePlaceholders_Nested(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"base": "hello",
		"msg":  "${base} world",
	})

	result := env.ResolvePlaceholders("${msg}")
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %s", result)
	}
}
