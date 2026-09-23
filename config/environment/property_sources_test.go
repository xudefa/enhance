package environment

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ==================== Map Property Source Tests ====================

func TestMapPropertySource_GetProperty(t *testing.T) {
	t.Parallel()
	src := NewMapPropertySource("test", 0, map[string]any{
		"server.port": 8080,
		"server.host": "localhost",
	})

	prop, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected property to exist")
	}
	v, ok := prop.(int)
	if !ok {
		t.Fatalf("expected int, got %T", prop)
	}
	if v != 8080 {
		t.Fatalf("expected 8080, got %v", prop)
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
		parsed := toEnvKey(tt.input)
		if parsed != tt.expected {
			t.Errorf("toEnvKey(%q) = %q, want %q", tt.input, parsed, tt.expected)
		}
	}
}

// ==================== Args Property Source Tests ====================

func TestArgsPropertySource(t *testing.T) {
	t.Parallel()
	src := NewArgsPropertySource("args", []string{
		"--server.port=9090",
		"--server.host=example.com",
		"--some-flag",
	})

	prop, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := prop.(string)
	if !ok {
		t.Fatalf("expected string, got %T", prop)
	}
	if v != "9090" {
		t.Fatalf("expected 9090, got %v", prop)
	}

	prop, ok = src.GetProperty("server.host")
	if !ok {
		t.Fatal("expected server.host to exist")
	}
	s, ok := prop.(string)
	if !ok {
		t.Fatalf("expected string, got %T", prop)
	}
	if s != "example.com" {
		t.Fatalf("expected example.com, got %v", prop)
	}
}

// ==================== Env Property Source Tests ====================

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
	prop, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := prop.(string)
	if !ok {
		t.Fatalf("expected string, got %T", prop)
	}
	if v != "9090" {
		t.Fatalf("expected 9090, got %v", prop)
	}
}

func TestEnvPropertySource_EmptyPrefix(t *testing.T) {
	origLookupEnv := lookupEnv
	defer func() { lookupEnv = origLookupEnv }()

	lookupEnv = func(key string) (string, bool) {
		if key == "SERVER_PORT" {
			return "9090", true
		}
		return "", false
	}

	src := NewEnvPropertySource("env", "")
	prop, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	if prop != "9090" {
		t.Fatalf("expected 9090, got %v", prop)
	}
}

// ==================== Default Property Source Tests ====================

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

	prop, ok := src.GetProperty("server.port")
	if !ok {
		t.Fatal("expected server.port to exist")
	}
	v, ok := prop.(int)
	if !ok || v != 8080 {
		t.Fatalf("server.port = %v, want 8080", prop)
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

	env.AddPropertySource(fileSource)
	env.AddPropertySource(NewDefaultPropertySource("defaults", map[string]any{
		"app.name": "from-default",
	}))

	if got := env.GetString("app.name", ""); got != "from-file" {
		t.Fatalf("GetString(app.name) = %q, want from-file (file must beat defaults)", got)
	}
}

// ==================== JSON Property Source Tests ====================

func testAssertJSONPropertySource(t *testing.T, source *JSONPropertySource) {
	t.Helper()
	if source.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", source.Name())
	}

	if source.Priority() != PriorityLowest {
		t.Errorf("Expected priority PriorityLowest, got %v", source.Priority())
	}

	prop, ok := source.GetProperty("server.host")
	if !ok {
		t.Error("Expected to find 'server.host'")
	}
	if prop != "localhost" {
		t.Errorf("Expected 'localhost', got '%v'", prop)
	}

	prop, ok = source.GetProperty("server.port")
	if !ok {
		t.Error("Expected to find 'server.port'")
	}
	if prop != float64(8080) {
		t.Errorf("Expected 8080, got '%v'", prop)
	}

	_, ok = source.GetProperty("nonexistent.key")
	if ok {
		t.Error("Expected not to find 'nonexistent.key'")
	}

	if !source.Contains("app.name") {
		t.Error("Expected Contains('app.name') to return true")
	}
	if source.Contains("nonexistent") {
		t.Error("Expected Contains('nonexistent') to return false")
	}

	keys := source.Keys()
	if len(keys) != 4 {
		t.Errorf("Expected 4 keys, got %d", len(keys))
	}
}

func TestNewJSONPropertySource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	testData := `{
		"server": {
			"host": "localhost",
			"port": 8080
		},
		"app": {
			"name": "test-app",
			"version": "1.0.0"
		}
	}`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	source, err := NewJSONPropertySource("test", testFile)
	if err != nil {
		t.Fatalf("Failed to create JSONPropertySource: %v", err)
	}

	testAssertJSONPropertySource(t, source)
}

func TestNewJSONPropertySourceOrDefault(t *testing.T) {
	t.Parallel()
	source := NewJSONPropertySourceOrDefault("test", "/nonexistent/file.json")
	if source == nil {
		t.Error("Expected source to be created even for nonexistent file")
	}

	_, ok := source.GetProperty("any.key")
	if ok {
		t.Error("Expected no properties from nonexistent file")
	}
}

func TestJSONPropertySourceNestedKeys(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nested.json")

	testData := `{
		"level1": {
			"level2": {
				"level3": {
					"value": "deep"
				}
			}
		}
	}`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	source, err := NewJSONPropertySource("nested", testFile)
	if err != nil {
		t.Fatalf("Failed to create JSONPropertySource: %v", err)
	}

	prop, ok := source.GetProperty("level1.level2.level3.value")
	if !ok {
		t.Error("Expected to find deep nested key")
	}
	if prop != "deep" {
		t.Errorf("Expected 'deep', got '%v'", prop)
	}
}

func TestJSONPropertySourceInvalidJSON(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(testFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := NewJSONPropertySource("invalid", testFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestFindApplicationConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	applicationConfig := filepath.Join(tmpDir, "application.json")
	if err := os.WriteFile(applicationConfig, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create application config file: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	found := FindApplicationConfigFile()
	if found == "" {
		t.Error("Expected to find application config file")
	}

	if err := os.Remove(applicationConfig); err != nil {
		t.Fatalf("Failed to remove application config: %v", err)
	}

	configFile := filepath.Join(configDir, "application.json")
	if err := os.WriteFile(configFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	found = FindApplicationConfigFile()
	if found == "" {
		t.Error("Expected to find application config file in config directory")
	}
}

func TestFlattenKeys(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}

	flattened := flattenKeys(input, "")

	if len(flattened) != 2 {
		t.Errorf("Expected 2 flattened keys, got %d", len(flattened))
	}

	found := make(map[string]bool)
	for _, k := range flattened {
		found[k] = true
	}
	if !found["server.host"] {
		t.Error("Expected to find server.host")
	}
	if !found["server.port"] {
		t.Error("Expected to find server.port")
	}
}

// ==================== Property Accessor Tests ====================

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
}

// ==================== Config Change Event Tests ====================

func TestConfigChangeEvent_BasicFields(t *testing.T) {
	t.Parallel()

	event := NewConfigChangeEvent(
		"modify",
		WithEventKeys([]string{"server.port", "server.host"}),
		WithEventValues(map[string]any{"server.port": "8080"}, map[string]any{"server.port": "9090"}),
		WithEventSource("nacos"),
	)

	if event.EventType != "modify" {
		t.Errorf("EventType = %v, want modify", event.EventType)
	}
	if len(event.Keys) != 2 {
		t.Errorf("Keys length = %v, want 2", len(event.Keys))
	}
	if event.Keys[0] != "server.port" {
		t.Errorf("Keys[0] = %v, want server.port", event.Keys[0])
	}
	if event.Source != "nacos" {
		t.Errorf("Source = %v, want nacos", event.Source)
	}
}

func TestConfigChangeEvent_Type(t *testing.T) {
	t.Parallel()

	event := NewConfigChangeEvent("modify", WithEventSource("test"))

	if event.Type() != "ConfigChange" {
		t.Errorf("Type() = %v, want ConfigChange", event.Type())
	}
}

func TestConfigChangeEvent_Timestamp(t *testing.T) {
	t.Parallel()

	before := time.Now()
	event := NewConfigChangeEvent("modify", WithEventSource("test"))
	after := time.Now()

	ts := event.Timestamp()
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Timestamp() = %v, want between %v and %v", ts, before, after)
	}
}

func TestConfigChangeEvent_Values(t *testing.T) {
	t.Parallel()

	oldVals := map[string]any{"a": 1}
	newVals := map[string]any{"a": 2}
	event := NewConfigChangeEvent(
		"modify",
		WithEventKeys([]string{"a"}),
		WithEventValues(oldVals, newVals),
		WithEventSource("test"),
	)

	if event.OldValues["a"] != 1 {
		t.Errorf("OldValues[a] = %v, want 1", event.OldValues["a"])
	}
	if event.NewValues["a"] != 2 {
		t.Errorf("NewValues[a] = %v, want 2", event.NewValues["a"])
	}
}

func TestConfigChangeEvent_Metadata(t *testing.T) {
	t.Parallel()

	event := NewConfigChangeEvent("modify", WithEventSource("test"))

	if event.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}
	if len(event.Metadata) != 0 {
		t.Errorf("Metadata length = %v, want 0", len(event.Metadata))
	}
}
