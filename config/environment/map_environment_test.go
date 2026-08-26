package environment

import (
	"testing"
)

func TestNewMapEnvironment(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"key1": "value1",
		"key2": "value2",
	})

	val, ok := env.GetProperty("key1")
	if !ok {
		t.Fatal("expected property to exist")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestNewMapEnvironmentWithProfiles(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironmentWithProfiles(
		map[string]string{
			"key1": "value1",
		},
		[]string{"dev", "test"},
	)

	val, ok := env.GetProperty("key1")
	if !ok {
		t.Fatal("expected property to exist")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}

	if !env.AcceptsProfiles("dev") {
		t.Error("expected 'dev' profile to be active")
	}
	if !env.AcceptsProfiles("test") {
		t.Error("expected 'test' profile to be active")
	}
	if env.AcceptsProfiles("prod") {
		t.Error("expected 'prod' profile to not be active")
	}
}

func TestEnvironment_AcceptsProfiles(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{})
	env.AddActiveProfile("dev")
	env.AddActiveProfile("test")

	if !env.AcceptsProfiles("dev", "prod") {
		t.Error("expected to accept at least one of the profiles")
	}
	if env.AcceptsProfiles("prod", "staging") {
		t.Error("expected to not accept any of the profiles")
	}
}

func TestEnvironment_GetPropertyWithDefault(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "myapp",
	})

	val := env.GetPropertyWithDefault("app.name", "default")
	if val != "myapp" {
		t.Errorf("expected 'myapp', got %s", val)
	}

	val = env.GetPropertyWithDefault("missing.key", "default")
	if val != "default" {
		t.Errorf("expected 'default', got %s", val)
	}
}
