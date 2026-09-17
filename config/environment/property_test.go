package environment

import (
	"testing"
)

func TestEnvironment_GetString_Property(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "myapp",
	})

	prop := env.GetString("app.name", "default")
	if prop != "myapp" {
		t.Errorf("expected 'myapp', got %s", prop)
	}

	prop = env.GetString("missing.key", "default")
	if prop != "default" {
		t.Errorf("expected 'default', got %s", prop)
	}
}

func TestEnvironment_GetInt_Property(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.port": "8080",
	})

	prop := env.GetInt("app.port", 3000)
	if prop != 8080 {
		t.Errorf("expected 8080, got %d", prop)
	}

	prop = env.GetInt("missing.port", 3000)
	if prop != 3000 {
		t.Errorf("expected 3000, got %d", prop)
	}
}

func TestEnvironment_GetBool_Property(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.debug": "true",
	})

	prop := env.GetBool("app.debug", false)
	if !prop {
		t.Error("expected true, got false")
	}

	prop = env.GetBool("missing.key", false)
	if prop {
		t.Error("expected false, got true")
	}
}

func TestEnvironment_ContainsProperty_Property(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "myapp",
	})

	if !env.ContainsProperty("app.name") {
		t.Error("expected property to exist")
	}
	if env.ContainsProperty("missing.key") {
		t.Error("expected property to not exist")
	}
}

func TestEnvironment_GetRequiredProperty_Property(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "myapp",
	})

	prop, err := env.GetRequiredProperty("app.name")
	if err != nil {
		t.Fatalf("GetRequiredProperty failed: %v", err)
	}
	if prop != "myapp" {
		t.Errorf("expected 'myapp', got %v", prop)
	}

	_, err = env.GetRequiredProperty("missing.key")
	if err == nil {
		t.Error("expected error for missing property")
	}
}

func TestEnvironment_GetFloat64_Property(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.ratio": "1.5",
	})

	prop := env.GetFloat64("app.ratio", 1.0)
	if prop != 1.5 {
		t.Errorf("expected 1.5, got %f", prop)
	}

	prop = env.GetFloat64("missing.key", 1.0)
	if prop != 1.0 {
		t.Errorf("expected 1.0, got %f", prop)
	}
}

func TestEnvironment_IsPropertyEmpty_Property(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name":  "myapp",
		"app.empty": "",
	})

	if env.IsPropertyEmpty("app.name") {
		t.Error("expected property to not be empty")
	}
	if !env.IsPropertyEmpty("app.empty") {
		t.Error("expected property to be empty")
	}
	if !env.IsPropertyEmpty("missing.key") {
		t.Error("expected missing property to be empty")
	}
}
