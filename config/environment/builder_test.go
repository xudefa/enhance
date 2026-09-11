package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironmentBuilder_WithJSONConfig(t *testing.T) {
	t.Parallel()

	// Create a temporary JSON config file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test-config.json")
	content := `{"test.key": "test-value", "app.name": "test-app"}`
	err := os.WriteFile(jsonFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	builder := NewEnvironmentBuilder().WithJSONConfig(jsonFile)
	if builder == nil {
		t.Fatal("expected builder to be created")
	}

	// Verify builder has property sources
	if len(builder.propertySources) == 0 {
		t.Error("expected builder to have property sources from JSON config")
	}

	env := builder.Build()
	if env == nil {
		t.Fatal("expected environment to be built")
	}
}

func TestEnvironmentBuilder_WithJSONConfig_NonExistent(t *testing.T) {
	t.Parallel()

	// Should not panic even if file doesn't exist
	builder := NewEnvironmentBuilder().WithJSONConfig("/non/existent/file.json")
	env := builder.Build()
	if env == nil {
		t.Fatal("expected environment to be built even with non-existent file")
	}
}

func TestEnvironmentBuilder_MustBuild(t *testing.T) {
	t.Parallel()

	env := NewEnvironmentBuilder().MustBuild()
	if env == nil {
		t.Fatal("expected environment to be built")
	}
}

func TestEnvironmentBuilder_ChainedMethods(t *testing.T) {
	t.Parallel()

	env := NewEnvironmentBuilder().
		WithProfile("dev").
		WithProfiles("test", "staging").
		WithEnvPrefix("TEST").
		WithArgs("--app.name=test", "--app.port=8080").
		Build()

	if env == nil {
		t.Fatal("expected environment to be built")
	}

	if !env.AcceptsProfile("dev") {
		t.Error("expected 'dev' profile to be active")
	}
	if !env.AcceptsProfile("test") {
		t.Error("expected 'test' profile to be active")
	}
}

func TestEnvironmentBuilder_WithPropertySource(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key": "value",
	})

	builder := NewEnvironmentBuilder().WithPropertySource(source)
	if len(builder.propertySources) != 1 {
		t.Errorf("expected 1 property source, got %d", len(builder.propertySources))
	}

	env := builder.Build()
	value, ok := env.GetProperty("key")
	if !ok || value != "value" {
		t.Errorf("expected 'value', got %v", value)
	}
}

func TestEnvironmentBuilder_WithPropertySourceFirst(t *testing.T) {
	t.Parallel()

	source1 := NewMapPropertySource("first", PriorityNormal, map[string]any{
		"key": "first-value",
	})
	source2 := NewMapPropertySource("second", PriorityNormal, map[string]any{
		"key": "second-value",
	})

	builder := NewEnvironmentBuilder().
		WithPropertySource(source1).
		WithPropertySourceFirst(source2)

	// 验证builder中的propertySources顺序
	if len(builder.propertySources) != 2 {
		t.Fatalf("expected 2 property sources, got %d", len(builder.propertySources))
	}
	// WithPropertySourceFirst应该把source2放在列表前面
	if builder.propertySources[0].Name() != "second" {
		t.Errorf("expected first source to be 'second', got %s", builder.propertySources[0].Name())
	}
}
