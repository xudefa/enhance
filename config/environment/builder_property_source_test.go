package environment

import (
	"os"
	"testing"
)

func TestPropertySource_Contains(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key1": "value1",
		"key2": "value2",
	})

	if !source.Contains("key1") {
		t.Error("expected source to contain 'key1'")
	}

	if source.Contains("nonexistent") {
		t.Error("expected source to not contain 'nonexistent'")
	}
}

func TestPropertySource_Keys(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key1": "value1",
		"key2": "value2",
	})

	keys := source.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestEnvPropertySource_Name(t *testing.T) {
	t.Parallel()

	source := NewEnvPropertySource("test-env", "TEST")
	if source.Name() != "test-env" {
		t.Errorf("expected name 'test-env', got %s", source.Name())
	}
}

func TestEnvPropertySource_Contains(t *testing.T) {
	t.Parallel()

	_ = os.Setenv("TEST_KEY", "test-value")
	defer func() { _ = os.Unsetenv("TEST_KEY") }()

	source := NewEnvPropertySource("test-env", "TEST")
	// Contains会将"key"转换为"TEST_KEY"（前缀+大写）
	if !source.Contains("key") {
		t.Error("expected env source to contain 'key' (mapped to 'TEST_KEY')")
	}
}

func TestArgsPropertySource_Name(t *testing.T) {
	t.Parallel()

	source := NewArgsPropertySource("test-args", []string{"--key=value"})
	if source.Name() != "test-args" {
		t.Errorf("expected name 'test-args', got %s", source.Name())
	}
}

func TestArgsPropertySource_Contains(t *testing.T) {
	t.Parallel()

	source := NewArgsPropertySource("test-args", []string{"--key=value"})
	if !source.Contains("key") {
		t.Error("expected args source to contain 'key'")
	}
}
