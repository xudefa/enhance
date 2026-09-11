package environment

import (
	"testing"
)

func TestEnvironmentHelper_WithPrefix(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
		"app.port": "8080",
	}))

	helper := NewEnvironmentHelper(env)
	prefixedHelper := helper.WithPrefix("app")

	if prefixedHelper == nil {
		t.Fatal("expected prefixed helper to be created")
	}
}

func TestEnvironmentHelper_GetString(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetString("app.name", "default")

	if value != "test-app" {
		t.Errorf("expected 'test-app', got %v", value)
	}
}

func TestEnvironmentHelper_GetString_Default(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	helper := NewEnvironmentHelper(env)

	value := helper.GetString("non.existent", "default-value")
	if value != "default-value" {
		t.Errorf("expected 'default-value', got %v", value)
	}
}

func TestEnvironmentHelper_ContainsProperty(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)

	if !helper.ContainsProperty("app.name") {
		t.Error("expected 'app.name' to exist")
	}

	if helper.ContainsProperty("non.existent") {
		t.Error("expected 'non.existent' to not exist")
	}
}

func TestEnvironmentHelper_GetInt(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.port": 8080,
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetInt("app.port", 3000)

	if value != 8080 {
		t.Errorf("expected 8080, got %d", value)
	}
}

func TestEnvironmentHelper_GetBool(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.debug": true,
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetBool("app.debug", false)

	if !value {
		t.Error("expected true, got false")
	}
}

func TestEnvironmentHelper_GetFloat64(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.ratio": 1.5,
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetFloat64("app.ratio", 1.0)

	if value != 1.5 {
		t.Errorf("expected 1.5, got %f", value)
	}
}

func TestEnvironmentHelper_IsDev(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("dev")

	helper := NewEnvironmentHelper(env)
	if !helper.IsDev() {
		t.Error("expected IsDev to return true")
	}
}

func TestEnvironmentHelper_IsProd(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("prod")

	helper := NewEnvironmentHelper(env)
	if !helper.IsProd() {
		t.Error("expected IsProd to return true")
	}
}

func TestEnvironmentHelper_IsTest(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("test")

	helper := NewEnvironmentHelper(env)
	if !helper.IsTest() {
		t.Error("expected IsTest to return true")
	}
}

func TestEnvironmentHelper_GetActiveProfile(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("staging")

	helper := NewEnvironmentHelper(env)
	profile := helper.GetActiveProfile()

	if profile != "staging" {
		t.Errorf("expected 'staging', got %s", profile)
	}
}

func TestEnvironmentHelper_GetRequiredProperty(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)

	// 测试获取存在的属性
	val, err := helper.GetRequiredProperty("app.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "test-app" {
		t.Errorf("expected 'test-app', got %v", val)
	}

	// 测试获取不存在的属性
	_, err = helper.GetRequiredProperty("non.existent")
	if err == nil {
		t.Error("expected error for non-existent property")
	}
}

func TestEnvironmentHelper_WithPrefix_GetString(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
		"app.port": "8080",
	}))

	helper := NewEnvironmentHelper(env)
	prefixedHelper := helper.WithPrefix("app")

	if prefixedHelper == nil {
		t.Fatal("expected prefixed helper to be created")
	}

	// 测试带前缀的键
	value := prefixedHelper.GetString("name", "default")
	if value != "test-app" {
		t.Errorf("expected 'test-app', got %v", value)
	}
}
