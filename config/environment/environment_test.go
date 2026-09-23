package environment

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEnvironment_GetProperty(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	src1 := NewMapPropertySource("high", 200, map[string]any{"key": "value1"})
	src2 := NewMapPropertySource("low", 100, map[string]any{"key": "value2"})

	env.AddPropertySource(src2)
	env.AddPropertySource(src1)

	prop := env.GetString("key", "")
	if prop != "value1" {
		t.Fatalf("expected value1 (higher priority), got %s", prop)
	}
}

func TestEnvironment_Profile(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddActiveProfile("dev")

	if !env.AcceptsProfile("dev") {
		t.Fatal("expected to accept dev profile")
	}
	if env.AcceptsProfile("prod") {
		t.Fatal("expected to not accept prod profile")
	}
	if !env.AcceptsProfile("!prod") {
		t.Fatal("expected to accept !prod when prod is not active")
	}
}

func TestEnvironment_MultiSourceMerge(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	low := NewMapPropertySource("low", 100, map[string]any{
		"server.port": 8080,
		"server.host": "default.com",
	})
	high := NewMapPropertySource("high", 200, map[string]any{
		"server.port": 9090,
	})

	env.AddPropertySource(low)
	env.AddPropertySource(high)

	port := env.GetInt("server.port", 0)
	if port != 9090 {
		t.Fatalf("expected 9090 from high priority, got %d", port)
	}
	host := env.GetString("server.host", "")
	if host != "default.com" {
		t.Fatalf("expected default.com from low priority fallback, got %s", host)
	}
}

func TestEnvironment_GetBool(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", 0, map[string]any{
		"enabled":  true,
		"disabled": false,
	}))

	if !env.GetBool("enabled", false) {
		t.Fatal("expected enabled to be true")
	}
	if env.GetBool("disabled", true) {
		t.Fatal("expected disabled to be false")
	}
}

func TestEnvironment_ContainsProperty(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", 0, map[string]any{
		"exists": "yes",
	}))

	if !env.ContainsProperty("exists") {
		t.Fatal("expected ContainsProperty to be true")
	}
	if env.ContainsProperty("missing") {
		t.Fatal("expected ContainsProperty to be false")
	}
}

func TestResolvePlaceholders_Basic(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "myapp",
		"host":     "localhost",
		"port":     8080,
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "resolve existing key",
			input: "${app.name}",
			want:  "myapp",
		},
		{
			name:  "placeholder in text",
			input: "http://${host}:${port}",
			want:  "http://localhost:8080",
		},
		{
			name:  "no placeholder",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "nonexistent key without default",
			input: "${nonexistent}",
			want:  "${nonexistent}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_DefaultValue(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"host": "localhost",
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "existing key with default",
			input: "${host:fallback}",
			want:  "localhost",
		},
		{
			name:  "nonexistent key with default",
			input: "${port:3306}",
			want:  "3306",
		},
		{
			name:  "default with nested placeholder",
			input: "${nonexistent:${host}}",
			want:  "localhost",
		},
		{
			name:  "nested default fallback chain",
			input: "${a:${b:fallback}}",
			want:  "fallback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_Recursive(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"env":     "dev",
		"host":    "my-${env}.com",
		"url":     "http://${host}:${port}",
		"port":    "8080",
		"chained": "${url}",
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single level",
			input: "${host}",
			want:  "my-dev.com",
		},
		{
			name:  "multi level",
			input: "${url}",
			want:  "http://my-dev.com:8080",
		},
		{
			name:  "chained reference",
			input: "${chained}",
			want:  "http://my-dev.com:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_CircularReference(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"a": "${b}",
		"b": "${a}",
	}))

	got := env.ResolvePlaceholders("${a}")
	if got != "${a}" {
		t.Fatalf("expected '${a}' (preserved on circular ref), got %q", got)
	}
}

func TestGetProperty_AutoResolvePlaceholders(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"host": "localhost",
		"port": 8080,
		"url":  "http://${host}:${port}",
	}))

	prop, ok := env.GetProperty("url")
	if !ok {
		t.Fatal("expected url to exist")
	}
	got, ok := prop.(string)
	if !ok {
		t.Fatalf("expected string, got %T", prop)
	}
	if got != "http://localhost:8080" {
		t.Fatalf("url = %q, want http://localhost:8080", got)
	}
}

func TestResolvePlaceholders_Nested(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"db.host":    "localhost",
		"db.port":    "5432",
		"db.url":     "jdbc:postgresql://${db.host}:${db.port}/mydb",
		"app.db.url": "${db.url}?ssl=true",
	}))

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "nested placeholder in value",
			input: "${app.db.url}",
			want:  "jdbc:postgresql://localhost:5432/mydb?ssl=true",
		},
		{
			name:  "placeholder with colon in text",
			input: "prefix:${db.host}:suffix",
			want:  "prefix:localhost:suffix",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.ResolvePlaceholders(tt.input)
			if got != tt.want {
				t.Errorf("ResolvePlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolvePlaceholders_CircularDependencyDetection(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"a": "${b}",
		"b": "${c}",
		"c": "${a}",
	}))

	got := env.ResolvePlaceholders("${a}")
	if got != "${a}" {
		t.Fatalf("expected '${a}' (preserved on circular ref), got %q", got)
	}
}

func TestResolvePlaceholders_SelfReference(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"self": "${self}",
	}))

	got := env.ResolvePlaceholders("${self}")
	if got != "${self}" {
		t.Fatalf("expected '${self}' (preserved on self-circular ref), got %q", got)
	}
}

func TestParseProfiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected []string
	}{
		{"dev", []string{"dev"}},
		{"dev,prod", []string{"dev", "prod"}},
		{" dev , prod ", []string{"dev", "prod"}},
		{"", nil},
	}
	for _, tt := range tests {
		parsed := ParseProfiles(tt.input)
		if len(parsed) != len(tt.expected) {
			t.Fatalf("ParseProfiles(%q) = %v, want %v", tt.input, parsed, tt.expected)
		}
		for i := range parsed {
			if parsed[i] != tt.expected[i] {
				t.Fatalf("ParseProfiles(%q) = %v, want %v", tt.input, parsed, tt.expected)
			}
		}
	}
}

func TestEnvironment_NotifyAfterClose(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	done := make(chan struct{})
	env.AddConfigChangeListener(func(event ConfigChangeEvent) {
		close(done)
	})

	env.Close()

	env.notifyConfigChange(NewConfigChangeEvent("modify", WithEventKeys([]string{"k"}), WithEventSource("test")))

	select {
	case <-done:
		t.Fatal("config change listener called after Close()")
	case <-time.After(100 * time.Millisecond):
		// 预期：Close 后不再通知监听器
	}
}

func TestEnvironment_CloseConcurrentWithNotify(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			env.notifyConfigChange(NewConfigChangeEvent("modify", WithEventKeys([]string{"k"}), WithEventSource("test")))
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		env.Close()
	}()
	wg.Wait()
}

func TestEnvironment_ConfigChangeListener(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	var called int32
	done := make(chan struct{})

	env.AddConfigChangeListener(func(event ConfigChangeEvent) {
		atomic.StoreInt32(&called, 1)
		close(done)
	})

	event := NewConfigChangeEvent(
		"modify",
		WithEventKeys([]string{"test.key"}),
		WithEventValues(map[string]any{"test.key": "old"}, map[string]any{"test.key": "new"}),
		WithEventSource("test"),
	)

	env.notifyConfigChange(event)

	// 使用 channel 等待异步通知
	select {
	case <-done:
		// 监听器已被调用
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for config change listener")
	}

	if atomic.LoadInt32(&called) == 0 {
		t.Error("expected config change listener to be called")
	}
}

// ==================== Map Environment Tests ====================

func TestNewMapEnvironment(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"key1": "value1",
		"key2": "value2",
	})

	prop, ok := env.GetProperty("key1")
	if !ok {
		t.Fatal("expected property to exist")
	}
	if prop != "value1" {
		t.Errorf("expected 'value1', got %v", prop)
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

	prop, ok := env.GetProperty("key1")
	if !ok {
		t.Fatal("expected property to exist")
	}
	if prop != "value1" {
		t.Errorf("expected 'value1', got %v", prop)
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

	prop := env.GetPropertyWithDefault("app.name", "default")
	if prop != "myapp" {
		t.Errorf("expected 'myapp', got %s", prop)
	}

	prop = env.GetPropertyWithDefault("missing.key", "default")
	if prop != "default" {
		t.Errorf("expected 'default', got %s", prop)
	}
}

func TestEnvironment_GetProperty_Plan(t *testing.T) {
	t.Parallel()

	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
		"app.port": "8080",
	})

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing property", "app.name", "test-app"},
		{"missing property", "missing", ""},
		{"with default", "missing", "default"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := env.GetPropertyWithDefault(tt.key, tt.expected)
			if got != tt.expected {
				t.Errorf("GetPropertyWithDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEnvironment_GetActiveProfiles_Plan(t *testing.T) {
	t.Parallel()

	env := NewMapEnvironmentWithProfiles(
		map[string]string{},
		[]string{"dev", "test"},
	)

	profiles := env.GetActiveProfiles()
	if len(profiles) != 2 {
		t.Errorf("GetActiveProfiles() returned %d profiles, want 2", len(profiles))
	}
}

func TestEnvironment_AcceptsProfiles_Plan(t *testing.T) {
	t.Parallel()

	env := NewMapEnvironmentWithProfiles(
		map[string]string{},
		[]string{"dev", "test"},
	)

	tests := []struct {
		name     string
		profiles []string
		expected bool
	}{
		{"accepts dev", []string{"dev"}, true},
		{"accepts test", []string{"test"}, true},
		{"accepts prod", []string{"prod"}, false},
		{"accepts multiple", []string{"dev", "prod"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := env.AcceptsProfiles(tt.profiles...)
			if got != tt.expected {
				t.Errorf("AcceptsProfiles() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ==================== Profile Tests ====================

func TestGetProfileActive_FromArgs(t *testing.T) {
	args := []string{"--profile=dev", "--other=value"}
	profile := GetProfileActive(args)
	if profile != "dev" {
		t.Errorf("expected 'dev', got %s", profile)
	}
}

func TestGetProfileActive_FromEnv(t *testing.T) {
	t.Setenv("GO_BOOT_PROFILE", "test")

	args := []string{"--other=value"}
	profile := GetProfileActive(args)
	if profile != "test" {
		t.Errorf("expected 'test', got %s", profile)
	}
}

func TestGetProfileActive_ArgsPriority(t *testing.T) {
	t.Setenv("GO_BOOT_PROFILE", "env-profile")

	args := []string{"--profile=arg-profile"}
	profile := GetProfileActive(args)
	if profile != "arg-profile" {
		t.Errorf("expected 'arg-profile', got %s", profile)
	}
}

func TestGetProfileActive_None(t *testing.T) {
	t.Setenv("GO_BOOT_PROFILE", "")

	args := []string{"--other=value"}
	profile := GetProfileActive(args)
	if profile != "" {
		t.Errorf("expected empty string, got %s", profile)
	}
}

func TestParseProfiles_ProfileFunc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single profile",
			input:    "dev",
			expected: []string{"dev"},
		},
		{
			name:     "multiple profiles",
			input:    "dev,test",
			expected: []string{"dev", "test"},
		},
		{
			name:     "with spaces",
			input:    "dev, test , prod",
			expected: []string{"dev", "test", "prod"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "with empty items",
			input:    "dev,,test",
			expected: []string{"dev", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parsed := ParseProfiles(tt.input)
			if len(parsed) != len(tt.expected) {
				t.Fatalf("expected %d profiles, got %d", len(tt.expected), len(parsed))
			}
			for i, p := range tt.expected {
				if parsed[i] != p {
					t.Errorf("profile[%d]: expected %s, got %s", i, p, parsed[i])
				}
			}
		})
	}
}

// ==================== Environment Default Config Tests ====================

func testEnvAssertApplicationConfig(t *testing.T, env *Environment) {
	t.Helper()
	sources := env.GetPropertySources()
	if len(sources) < 3 {
		t.Errorf("Expected at least 3 sources, got %d", len(sources))
	}

	prop, ok := env.GetProperty("server.host")
	if !ok {
		t.Error("Expected to find 'server.host' from application config")
	}
	if prop != "0.0.0.0" {
		t.Errorf("Expected '0.0.0.0', got '%v'", prop)
	}

	prop, ok = env.GetProperty("server.port")
	if !ok {
		t.Error("Expected to find 'server.port' from application config")
	}
	if prop != float64(9090) {
		t.Errorf("Expected 9090, got '%v'", prop)
	}

	prop, ok = env.GetProperty("app.name")
	if !ok {
		t.Error("Expected to find 'app.name' from application config")
	}
	if prop != "test-app" {
		t.Errorf("Expected 'test-app', got '%v'", prop)
	}
}

func TestNewEnvironmentWithApplicationConfig(t *testing.T) {
	tmpDir := t.TempDir()

	applicationConfig := `{
		"server": {
			"host": "0.0.0.0",
			"port": 9090
		},
		"app": {
			"name": "test-app"
		}
	}`

	configFile := filepath.Join(tmpDir, "application.json")
	if err := os.WriteFile(configFile, []byte(applicationConfig), 0644); err != nil {
		t.Fatalf("Failed to create application config: %v", err)
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

	env := NewEnvironment()
	testEnvAssertApplicationConfig(t, env)
}

func TestNewEnvironmentWithoutApplicationConfig(t *testing.T) {
	tmpDir := t.TempDir()

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

	env := NewEnvironment()
	sources := env.GetPropertySources()
	if len(sources) < 2 {
		t.Errorf("Expected at least 2 sources (args + env), got %d", len(sources))
	}
}
