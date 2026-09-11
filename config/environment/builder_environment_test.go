package environment

import (
	"testing"
)

func TestEnvironment_AddPropertySourceFirst(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("first", PriorityNormal, map[string]any{
		"key": "first-value",
	}))
	env.AddPropertySourceFirst(NewMapPropertySource("second", PriorityNormal, map[string]any{
		"key": "second-value",
	}))

	value, ok := env.GetProperty("key")
	if !ok {
		t.Fatal("expected 'key' to exist")
	}
	if value != "second-value" {
		t.Errorf("expected 'second-value' (higher priority), got %v", value)
	}
}

func TestEnvironment_RemovePropertySource(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key": "value",
	}))

	// Verify property exists
	_, ok := env.GetProperty("key")
	if !ok {
		t.Fatal("expected 'key' to exist before removal")
	}

	// Remove the property source by name
	env.RemovePropertySource("test")

	// Verify property no longer exists
	_, ok = env.GetProperty("key")
	if ok {
		t.Error("expected 'key' to not exist after removal")
	}
}

func TestEnvironment_RemoveProfile(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("dev")
	env.AddActiveProfile("test")

	// Verify profiles exist
	if !env.AcceptsProfile("dev") {
		t.Fatal("expected 'dev' profile to exist before removal")
	}

	// Remove the profile
	env.RemoveProfile("dev")

	// Verify profile no longer exists
	if env.AcceptsProfile("dev") {
		t.Error("expected 'dev' profile to not exist after removal")
	}

	// Verify other profile still exists
	if !env.AcceptsProfile("test") {
		t.Error("expected 'test' profile to still exist")
	}
}

func TestEnvironment_GetPropertySources(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	initialCount := len(env.GetPropertySources())

	source1 := NewMapPropertySource("source1", PriorityNormal, map[string]any{})
	source2 := NewMapPropertySource("source2", PriorityNormal, map[string]any{})
	env.AddPropertySource(source1)
	env.AddPropertySource(source2)

	sources := env.GetPropertySources()
	if len(sources) != initialCount+2 {
		t.Errorf("expected %d property sources, got %d", initialCount+2, len(sources))
	}
}
