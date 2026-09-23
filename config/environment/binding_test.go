package environment

import (
	"reflect"
	"testing"
	"time"
)

func TestEnvironment_Bind(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"name": "test-app",
		"port": "8080",
	})

	type Config struct {
		Name string
		Port int
	}

	cfg := &Config{}
	err := env.Bind(cfg)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
}

func TestEnvironment_Bind_NilPointer(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"name": "test-app",
	})

	err := env.Bind(nil)
	if err == nil {
		t.Error("expected error for nil pointer")
	}
}

func TestEnvironment_BindKey(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})

	var name string
	err := env.BindKey("app.name", &name)
	if err != nil {
		t.Fatalf("BindKey failed: %v", err)
	}
	if name != "test-app" {
		t.Errorf("expected 'test-app', got %s", name)
	}
}

func TestEnvironment_BindKey_NotFound(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{})

	var name string
	err := env.BindKey("nonexistent", &name)
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestEnvironment_BindPrefix(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
		"app.port": "8080",
	})

	type AppConfig struct {
		Name string
		Port int
	}

	cfg := &AppConfig{}
	err := env.BindPrefix("app", cfg)
	if err != nil {
		t.Fatalf("BindPrefix failed: %v", err)
	}
}

func TestEnvironment_Validate(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"name": "test",
	})

	errs := env.Validate()
	if len(errs) > 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestEnvironment_ResolvePlaceholders(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"app.name": "myapp",
		"app.url":  "http://${app.name}.example.com",
	})

	got := env.ResolvePlaceholders("${app.url}")
	if got == "" {
		t.Error("expected resolved placeholder")
	}
}

func TestEnvironment_ResolvePlaceholders_WithDefault(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{})

	got := env.ResolvePlaceholders("${missing:default-value}")
	if got != "default-value" {
		t.Errorf("expected 'default-value', got %s", got)
	}
}

func TestEnvironment_ResolvePlaceholders_Nested(t *testing.T) {
	t.Parallel()
	env := NewMapEnvironment(map[string]string{
		"base": "hello",
		"msg":  "${base} world",
	})

	got := env.ResolvePlaceholders("${msg}")
	if got != "hello world" {
		t.Errorf("expected 'hello world', got %s", got)
	}
}

func TestBindProperties_ValueTag(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port":    "9090",
		"server.host":    "localhost",
		"server.timeout": "30",
	}))

	type ServerConfig struct {
		Port    int    `value:"${server.port:8080}"`
		Host    string `value:"${server.host:0.0.0.0}"`
		Timeout int    `value:"${server.timeout:30}"`
	}

	var config ServerConfig
	err := env.BindProperties(&config)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if config.Port != 9090 {
		t.Errorf("expected port 9090, got %d", config.Port)
	}
	if config.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", config.Host)
	}
	if config.Timeout != 30 {
		t.Errorf("expected timeout 30, got %d", config.Timeout)
	}
}

func TestBindProperties_DefaultValue(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	type ServerConfig struct {
		Port    int           `value:"${server.port:8080}"`
		Host    string        `value:"${server.host:0.0.0.0}"`
		Timeout time.Duration `value:"${server.timeout:30s}"`
	}

	var config ServerConfig
	err := env.BindProperties(&config)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if config.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Port)
	}
	if config.Host != "0.0.0.0" {
		t.Errorf("expected host '0.0.0.0', got '%s'", config.Host)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", config.Timeout)
	}
}

func TestBindProperties_NestedStruct(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port":        "9090",
		"server.ssl.enabled": "true",
	}))

	type SSLConfig struct {
		Enabled bool `value:"${server.ssl.enabled:false}"`
	}

	type ServerConfig struct {
		Port int `value:"${server.port:8080}"`
		SSL  SSLConfig
	}

	var config ServerConfig
	err := env.BindProperties(&config)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if config.Port != 9090 {
		t.Errorf("expected port 9090, got %d", config.Port)
	}
	if !config.SSL.Enabled {
		t.Error("expected SSL enabled to be true")
	}
}

func TestBindProperties_MapStructure(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.url":  "localhost:5432",
		"database.name": "mydb",
	}))

	type DatabaseConfig struct {
		URL  string `mapstructure:"database.url"`
		Name string `mapstructure:"database.name"`
	}

	var config DatabaseConfig
	err := env.BindProperties(&config)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if config.URL != "localhost:5432" {
		t.Errorf("expected URL 'localhost:5432', got '%s'", config.URL)
	}
	if config.Name != "mydb" {
		t.Errorf("expected name 'mydb', got '%s'", config.Name)
	}
}

func TestBindProperties_InvalidType(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	var config string
	err := env.BindProperties(config)
	if err == nil {
		t.Error("expected error for non-pointer target")
	}
}

func TestBindConfigPrefix_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string
		Port int
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
		"app.port": "8080",
	})

	cfg, err := BindConfigPrefix[Config](env, "app")
	if err != nil {
		t.Fatalf("BindConfigPrefix failed: %v", err)
	}

	if cfg.Name != "test-app" {
		t.Errorf("Expected Name 'test-app', got %s", cfg.Name)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port 8080, got %d", cfg.Port)
	}
}

func TestBindConfig_NestedStruct_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		AppName string `config:"app.name"`
		DBHost  string `config:"db.host"`
		DBPort  int    `config:"db.port"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "my-app",
		"db.host":  "localhost",
		"db.port":  "5432",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.AppName != "my-app" {
		t.Errorf("Expected AppName 'my-app', got %s", cfg.AppName)
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("Expected DBHost 'localhost', got %s", cfg.DBHost)
	}
	if cfg.DBPort != 5432 {
		t.Errorf("Expected DBPort 5432, got %d", cfg.DBPort)
	}
}

func TestBindConfig_WithTime_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Timeout time.Duration `config:"app.timeout"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.timeout": "5s",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Expected Timeout 5s, got %v", cfg.Timeout)
	}
}

func TestBindConfig_WithSlice_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Ports []int `config:"app.ports"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.ports": "8080,8081,8082",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if len(cfg.Ports) != 3 {
		t.Fatalf("Expected 3 ports, got %d", len(cfg.Ports))
	}
}

func TestBindConfig_Generic(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port":  9090,
		"server.host":  "localhost",
		"server.debug": true,
	}))

	type ServerConfig struct {
		Port  int    `config:"server.port"`
		Host  string `config:"server.host"`
		Debug bool   `config:"server.debug"`
	}

	cfg, err := BindConfig[ServerConfig](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", cfg.Host)
	}
	if !cfg.Debug {
		t.Error("expected debug to be true")
	}
}

func TestBindConfigPrefix_Generic(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port": 8080,
		"server.host": "0.0.0.0",
		"db.url":      "localhost:5432",
	}))

	type ServerConfig struct {
		Port int    `config:"port"`
		Host string `config:"host"`
	}

	cfg, err := BindConfigPrefix[ServerConfig](env, "server")
	if err != nil {
		t.Fatalf("BindConfigPrefix failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host '0.0.0.0', got '%s'", cfg.Host)
	}
}

func TestBindConfigRequired_Generic(t *testing.T) {
	t.Parallel()
	t.Run("all required fields present", func(t *testing.T) {
		env := NewEnvironment()
		env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
			"db.host": "localhost",
			"db.port": 5432,
		}))

		type DBConfig struct {
			Host string `config:"db.host" required:"true"`
			Port int    `config:"db.port" required:"true"`
		}

		cfg, err := BindConfigRequired[DBConfig](env)
		if err != nil {
			t.Fatalf("BindConfigRequired failed: %v", err)
		}
		if cfg.Host != "localhost" || cfg.Port != 5432 {
			t.Errorf("unexpected config values: %+v", cfg)
		}
	})

	t.Run("missing required field returns error", func(t *testing.T) {
		env := NewEnvironment()
		env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
			"db.host": "localhost",
		}))

		type DBConfig struct {
			Host string `config:"db.host" required:"true"`
			Port int    `config:"db.port" required:"true"`
		}

		_, err := BindConfigRequired[DBConfig](env)
		if err == nil {
			t.Fatal("expected error for missing required field")
		}
	})
}

func TestMustBindConfig_Generic(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		env := NewEnvironment()
		env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
			"app.name": "test-app",
		}))

		type AppConfig struct {
			Name string `config:"app.name"`
		}

		cfg := MustBindConfig[AppConfig](env)
		if cfg.Name != "test-app" {
			t.Errorf("expected name 'test-app', got '%s'", cfg.Name)
		}
	})
}

func TestBindConfig_ValueTag(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port": 9090,
	}))

	type ServerConfig struct {
		Port int `value:"${server.port:8080}"`
	}

	cfg, err := BindConfig[ServerConfig](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
}

func TestBindConfig_DefaultValue(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()

	type ServerConfig struct {
		Port int `value:"${server.port:8080}"`
	}

	cfg, err := BindConfig[ServerConfig](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080 (default), got %d", cfg.Port)
	}
}

func TestBindConfig_DefaultTag(t *testing.T) {
	t.Parallel()
	t.Run("default applied when key missing", func(t *testing.T) {
		t.Parallel()
		env := NewEnvironment()
		env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
			"server.host": "localhost",
		}))

		type ServerConfig struct {
			Port int    `config:"server.port" default:"8080"`
			Host string `config:"server.host"`
		}

		cfg, err := BindConfig[ServerConfig](env)
		if err != nil {
			t.Fatalf("BindConfig failed: %v", err)
		}
		if cfg.Port != 8080 {
			t.Errorf("expected default port 8080, got %d", cfg.Port)
		}
		if cfg.Host != "localhost" {
			t.Errorf("expected host 'localhost', got '%s'", cfg.Host)
		}
	})

	t.Run("explicit value overrides default", func(t *testing.T) {
		t.Parallel()
		env := NewEnvironment()
		env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
			"server.port": 9090,
		}))

		type ServerConfig struct {
			Port int `config:"server.port" default:"8080"`
		}

		cfg, err := BindConfig[ServerConfig](env)
		if err != nil {
			t.Fatalf("BindConfig failed: %v", err)
		}
		if cfg.Port != 9090 {
			t.Errorf("expected port 9090 from config, got %d", cfg.Port)
		}
	})
}

func TestBindConfig_NestedStruct_Generic(t *testing.T) {
	t.Parallel()
	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port":        9090,
		"server.ssl.enabled": true,
	}))

	type SSLConfig struct {
		Enabled bool `config:"enabled"`
	}

	type ServerConfig struct {
		Port int `config:"port"`
		SSL  SSLConfig
	}

	cfg, err := BindConfigPrefix[ServerConfig](env, "server")
	if err != nil {
		t.Fatalf("BindConfigPrefix failed: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if !cfg.SSL.Enabled {
		t.Error("expected SSL enabled")
	}
}

func TestIsEmptyValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "test", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"nil pointer", (*int)(nil), true},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1}, false},
		{"empty map", map[string]int{}, true},
		{"non-empty map", map[string]int{"a": 1}, false},
		{"zero float", 0.0, true},
		{"non-zero float", 1.5, false},
		{"empty struct", struct{}{}, true},
		{"zero time", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			got := isEmptyValue(v)
			if got != tt.expected {
				t.Errorf("isEmptyValue(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestMustBindConfig_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Logf("panic occurred (expected in some cases): %v", recovered)
		}
	}()

	env := NewEnvironment()

	type Config struct {
		Name string `config:"app.name"`
	}

	cfg := MustBindConfig[Config](env)
	if cfg.Name != "" {
		t.Errorf("expected empty name, got '%s'", cfg.Name)
	}
}

func TestMustBindConfigPrefix_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Logf("panic occurred (expected in some cases): %v", recovered)
		}
	}()

	env := NewEnvironment()

	type Config struct {
		Port int `config:"port"`
	}

	cfg := MustBindConfigPrefix[Config](env, "server")
	if cfg.Port != 0 {
		t.Errorf("expected 0 port, got %d", cfg.Port)
	}
}

func TestBindConfig_WithTimeDuration(t *testing.T) {
	t.Parallel()

	type Config struct {
		Timeout time.Duration `config:"app.timeout"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.timeout": "5s",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected Timeout 5s, got %v", cfg.Timeout)
	}
}

func TestBindConfig_WithSlice(t *testing.T) {
	t.Parallel()

	type Config struct {
		Ports []int `config:"app.ports"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.ports": "8080,8081,8082",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if len(cfg.Ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(cfg.Ports))
	}
}

func TestBindConfig_WithMapstructureTag(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `mapstructure:"app.name"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "mapstructure-test",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Name != "mapstructure-test" {
		t.Errorf("expected Name 'mapstructure-test', got %s", cfg.Name)
	}
}

func TestBindConfig_WithEnvTag(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `env:"APP_NAME"`
	}

	env := NewMapEnvironment(map[string]string{
		"APP_NAME": "env-test",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Name != "env-test" {
		t.Errorf("expected Name 'env-test', got %s", cfg.Name)
	}
}
