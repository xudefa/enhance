package environment

import (
	"testing"
	"time"
)

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
