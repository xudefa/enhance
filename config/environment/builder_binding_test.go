package environment

import (
	"testing"
)

func TestBindConfig(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"name":    "test-app",
		"version": "1.0.0",
	}))

	type Config struct {
		Name    string
		Version string
	}

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "test-app" {
		t.Errorf("expected name 'test-app', got %s", cfg.Name)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", cfg.Version)
	}
}

func TestBindConfigPrefix(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name":    "test-app",
		"app.version": "1.0.0",
	}))

	type Config struct {
		Name    string
		Version string
	}

	cfg, err := BindConfigPrefix[Config](env, "app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "test-app" {
		t.Errorf("expected name 'test-app', got %s", cfg.Name)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", cfg.Version)
	}
}

func TestMustBindConfigPrefix(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"db.host": "localhost",
		"db.port": "5432",
	}))

	type DBConfig struct {
		Host string
		Port string
	}

	cfg := MustBindConfigPrefix[DBConfig](env, "db")
	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != "5432" {
		t.Errorf("expected port '5432', got %s", cfg.Port)
	}
}
