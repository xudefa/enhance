package config

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

// ==================== extractSubMap 测试 ====================

func TestExtractSubMap(t *testing.T) {
	t.Parallel()
	src := map[string]any{
		"app.name":  "myapp",
		"app.port":  8080,
		"db.host":   "localhost",
		"db.port":   3306,
		"unrelated": true,
	}

	got := extractSubMap(src, "app.")

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got["name"] != "myapp" {
		t.Errorf("expected name=myapp, got %v", got["name"])
	}
	if got["port"] != 8080 {
		t.Errorf("expected port=8080, got %v", got["port"])
	}
}

func TestExtractSubMap_Empty(t *testing.T) {
	t.Parallel()
	src := map[string]any{"a": 1}
	got := extractSubMap(src, "nonexistent.")
	if len(got) != 0 {
		t.Errorf("expected empty got, got %v", got)
	}
}

func TestExtractSubMap_EmptyData(t *testing.T) {
	t.Parallel()
	src := map[string]any{}
	got := extractSubMap(src, "prefix.")
	if len(got) != 0 {
		t.Errorf("expected empty got, got %v", got)
	}
}

// ==================== AutoEnv / detectEnv / getEnv 测试 ====================

func TestAutoEnv(t *testing.T) {
	builder := NewConfigBuilder()

	t.Setenv("APP_ENV", "production")

	got := builder.AutoEnv()
	if got != builder {
		t.Error("AutoEnv should return builder for chaining")
	}
}

func TestDetectEnv_APP_ENV(t *testing.T) {
	t.Setenv("APP_ENV", "staging")

	got := detectEnv()
	if got != "staging" {
		t.Errorf("expected 'staging', got '%s'", got)
	}
}

func TestDetectEnv_GO_ENV(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("GO_ENV", "test")

	got := detectEnv()
	if got != "test" {
		t.Errorf("expected 'test', got '%s'", got)
	}
}

func TestDetectEnv_ENV(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("GO_ENV", "")
	t.Setenv("ENV", "uat")

	got := detectEnv()
	if got != "uat" {
		t.Errorf("expected 'uat', got '%s'", got)
	}
}

func TestDetectEnv_Default(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("GO_ENV", "")
	t.Setenv("ENV", "")

	got := detectEnv()
	if got != "dev" {
		t.Errorf("expected 'dev', got '%s'", got)
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_GET_ENV_KEY", "hello")

	got := getEnv("TEST_GET_ENV_KEY")
	if got != "hello" {
		t.Errorf("expected 'hello', got '%s'", got)
	}
}

func TestGetEnv_Empty(t *testing.T) {
	t.Parallel()
	got := getEnv("NONEXISTENT_KEY_12345")
	if got != "" {
		t.Errorf("expected empty string, got '%s'", got)
	}
}

// ==================== PropertyBinder WithPrefix / WithValidator 测试 ====================

func TestPropertyBinder_WithPrefix(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	binder := NewPropertyBinder(env)

	got := binder.WithPrefix("app")
	if got != binder {
		t.Error("WithPrefix should return binder for chaining")
	}
	if binder.prefix != "app" {
		t.Errorf("expected prefix 'app', got '%s'", binder.prefix)
	}
}

func TestPropertyBinder_WithValidator(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	binder := NewPropertyBinder(env)
	validator := NewValidator()

	got := binder.WithValidator(validator)
	if got != binder {
		t.Error("WithValidator should return binder for chaining")
	}
	if binder.validator != validator {
		t.Error("expected validator to be set")
	}
}

// ==================== BindOption WithBindValidator 测试 ====================

func TestWithBindValidator(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	binder := NewPropertyBinder(env)
	validator := NewValidator()

	opt := WithBindValidator(validator)
	opt(binder)

	if binder.validator != validator {
		t.Error("expected validator to be set via WithBindValidator")
	}
}

// ==================== ConfigBuilder Build 测试 ====================

func TestConfigBuilder_Build_WithAutoEnv(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	cfg, err := NewConfigBuilder().
		Name("test").
		AutoEnv().
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Env != "test" {
		t.Errorf("expected Env 'test', got '%s'", cfg.Env)
	}
}

func TestConfigBuilder_Build_WithEnvPrefix(t *testing.T) {
	t.Parallel()
	cfg, err := NewConfigBuilder().
		Name("test").
		EnvPrefix("MYAPP_").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OptionName != "MYAPP_" {
		t.Errorf("expected OptionName 'MYAPP_', got '%s'", cfg.OptionName)
	}
}
