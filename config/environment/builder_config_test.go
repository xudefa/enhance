package environment

import (
	"testing"
)

func TestWithProfiles(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opt := WithProfiles("dev", "test")
	opt(config)

	if len(config.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(config.Profiles))
	}
}

func TestWithDefaultProfile(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opt := WithDefaultProfile("prod")
	opt(config)

	if config.DefaultProfile != "prod" {
		t.Errorf("expected 'prod', got %s", config.DefaultProfile)
	}
}

func TestWithAutoDetectProfiles(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opt := WithAutoDetectProfiles(true)
	opt(config)

	if !config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be true")
	}
}

func TestWithPropertySources(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	source := NewMapPropertySource("test", PriorityNormal, map[string]any{})
	opt := WithPropertySources(source)
	opt(config)

	if len(config.PropertySources) != 1 {
		t.Errorf("expected 1 property source, got %d", len(config.PropertySources))
	}
}

func TestDefaultEnvironmentConfig(t *testing.T) {
	t.Parallel()

	config := DefaultEnvironmentConfig()

	if config.DefaultProfile != "default" {
		t.Errorf("expected 'default', got %s", config.DefaultProfile)
	}
	if !config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be true")
	}
}

func TestEnvironmentConfig_ApplyOptions(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opts := []EnvironmentOption{
		WithProfiles("dev"),
		WithDefaultProfile("default"),
		WithAutoDetectProfiles(false),
	}

	config.ApplyOptions(opts)

	if len(config.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(config.Profiles))
	}
	if config.DefaultProfile != "default" {
		t.Errorf("expected 'default', got %s", config.DefaultProfile)
	}
	if config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be false")
	}
}

func TestCreateEnvironment(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key": "value",
	})

	env := CreateEnvironment(
		WithProfiles("dev"),
		WithDefaultProfile("default"),
		WithPropertySources(source),
	)

	if env == nil {
		t.Fatal("expected environment to be created")
	}

	if !env.AcceptsProfile("dev") {
		t.Error("expected 'dev' profile to be active")
	}

	value, ok := env.GetProperty("key")
	if !ok || value != "value" {
		t.Errorf("expected 'value', got %v", value)
	}
}
