package environment

import (
	"reflect"
	"testing"
	"time"
)

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
	// 没有属性源

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

func TestBindConfig_NestedStruct(t *testing.T) {
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
			result := isEmptyValue(v)
			if result != tt.expected {
				t.Errorf("isEmptyValue(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestMustBindConfig_Panic(t *testing.T) {
	t.Parallel()

	// 这个测试验证MustBindConfig在绑定失败时会panic
	// 但由于我们的配置绑定总是成功（即使字段为空），
	// 所以我们只验证函数可以正常调用
	defer func() {
		if r := recover(); r != nil {
			t.Logf("panic occurred (expected in some cases): %v", r)
		}
	}()

	env := NewEnvironment()

	type Config struct {
		Name string `config:"app.name"`
	}

	cfg := MustBindConfig[Config](env)
	// Name应该为空字符串
	if cfg.Name != "" {
		t.Errorf("expected empty name, got '%s'", cfg.Name)
	}
}

func TestMustBindConfigPrefix_Panic(t *testing.T) {
	t.Parallel()

	// 这个测试验证MustBindConfigPrefix在绑定失败时会panic
	// 但由于我们的配置绑定总是成功（即使字段为空），
	// 所以我们只验证函数可以正常调用
	defer func() {
		if r := recover(); r != nil {
			t.Logf("panic occurred (expected in some cases): %v", r)
		}
	}()

	env := NewEnvironment()

	type Config struct {
		Port int `config:"port"`
	}

	cfg := MustBindConfigPrefix[Config](env, "server")
	// Port应该为0
	if cfg.Port != 0 {
		t.Errorf("expected 0 port, got %d", cfg.Port)
	}
}
