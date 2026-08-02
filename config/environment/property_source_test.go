package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMapPropertySource_GetProperty(t *testing.T) {
	t.Parallel()
	src := NewMapPropertySource("test", 0, map[string]any{
		"server.port": 8080,
		"server.host": "localhost",
	})

	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected property to exist")
	}
	v, ok := val.(int)
	if !ok {
		t.Fatalf("expected int, got %T", val)
	}
	if v != 8080 {
		t.Fatalf("expected 8080, got %v", val)
	}

	_, ok = src.GetProperty("nonexistent")
	if ok {
		t.Fatal("expected property to not exist")
	}
}

func TestMapPropertySource_Priority(t *testing.T) {
	t.Parallel()
	src1 := NewMapPropertySource("low", 100, nil)
	src2 := NewMapPropertySource("high", 200, nil)

	if src1.Priority() >= src2.Priority() {
		t.Fatal("expected src1 to have lower priority than src2")
	}
}

func TestArgsPropertySource(t *testing.T) {
	t.Parallel()
	src := NewArgsPropertySource("args", []string{
		"--server.port=9090",
		"--server.host=example.com",
		"--some-flag",
	})

	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if v != "9090" {
		t.Fatalf("expected 9090, got %v", val)
	}

	val, ok = src.GetProperty("server.host")
	if !ok {
		t.Fatal("expected server.host to exist")
	}
	s, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if s != "example.com" {
		t.Fatalf("expected example.com, got %v", val)
	}
}

func TestToEnvKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"server.port", "SERVER_PORT"},
		{"server.host", "SERVER_HOST"},
		{"app.name", "APP_NAME"},
		{"simple", "SIMPLE"},
		{"nested.key.path", "NESTED_KEY_PATH"},
	}
	for _, tt := range tests {
		result := toEnvKey(tt.input)
		if result != tt.expected {
			t.Errorf("toEnvKey(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestEnvPropertySource_GetProperty(t *testing.T) {
	origLookupEnv := lookupEnv
	defer func() { lookupEnv = origLookupEnv }()

	lookupEnv = func(key string) (string, bool) {
		if key == "GO_BOOT_SERVER_PORT" {
			return "9090", true
		}
		return "", false
	}

	src := NewEnvPropertySource("env", "GO_BOOT")
	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if v != "9090" {
		t.Fatalf("expected 9090, got %v", val)
	}
}

func TestNewDefaultPropertySource(t *testing.T) {
	t.Parallel()
	src := NewDefaultPropertySource("defaults", map[string]any{
		"server.port": 8080,
		"server.host": "localhost",
	})

	if src.Name() != "defaults" {
		t.Fatalf("Name() = %s, want defaults", src.Name())
	}
	if src.Priority() != PriorityFallback {
		t.Fatalf("Priority() = %d, want %d", src.Priority(), PriorityFallback)
	}

	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := val.(int)
	if !ok || v != 8080 {
		t.Fatalf("server.port = %v, want 8080", val)
	}
}

func TestDefaultPropertySource_OverriddenByOtherSource(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	defaults := NewDefaultPropertySource("defaults", map[string]any{
		"server.port": 8080,
		"server.host": "default.com",
	})
	normal := NewMapPropertySource("normal", PriorityNormal, map[string]any{
		"server.port": 9090,
	})

	env.AddPropertySource(defaults)
	env.AddPropertySource(normal)

	port := env.GetInt("server.port", 0)
	if port != 9090 {
		t.Fatalf("expected 9090 (normal priority), got %d", port)
	}
	host := env.GetString("server.host", "")
	if host != "default.com" {
		t.Fatalf("expected default.com (fallback from defaults), got %s", host)
	}
}

func TestDefaultPropertySource_DoesNotOverrideFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "application.json")
	if err := os.WriteFile(configFile, []byte(`{"app":{"name":"from-file"}}`), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	env := NewEnvironment()
	fileSource, err := NewJSONPropertySource("application-config", configFile)
	if err != nil {
		t.Fatalf("failed to create JSON source: %v", err)
	}

	// 模拟常见流程：先加载 application.json，再添加默认值源。
	// 无论添加顺序如何，文件配置源都必须优先于默认值源。
	env.AddPropertySource(fileSource)
	env.AddPropertySource(NewDefaultPropertySource("defaults", map[string]any{
		"app.name": "from-default",
	}))

	if got := env.GetString("app.name", ""); got != "from-file" {
		t.Fatalf("GetString(app.name) = %q, want from-file (file must beat defaults)", got)
	}
}

func TestEnvPropertySource_EmptyPrefix(t *testing.T) {
	origLookupEnv := lookupEnv
	defer func() { lookupEnv = origLookupEnv }()

	lookupEnv = func(key string) (string, bool) {
		if key == "SERVER_PORT" {
			return "8080", true
		}
		return "", false
	}

	src := NewEnvPropertySource("env", "")
	val, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist with empty prefix")
	}
	v, ok := val.(string)
	if !ok {
		t.Fatalf("expected string, got %T", val)
	}
	if v != "8080" {
		t.Fatalf("expected 8080, got %v", val)
	}
}
