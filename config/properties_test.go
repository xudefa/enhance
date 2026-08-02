package config

import (
	"testing"
	"time"

	"github.com/xudefa/enhance/config/environment"
)

func TestBindProperties_BasicTypes(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Name    string  `enhance:"app.name"`
		Port    int     `enhance:"server.port" default:"8080"`
		Enabled bool    `enhance:"server.enabled" default:"true"`
		Rate    float64 `enhance:"server.rate" default:"0.5"`
	}

	env := environment.NewEnvironment()
	cfg := &TestConfig{}

	err := BindProperties(cfg, env)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	// app.name may come from environment config files
	if cfg.Name == "" {
		t.Errorf("expected Name to be set, got empty")
	}
	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", cfg.Port)
	}
	if cfg.Enabled != true {
		t.Errorf("expected Enabled=true, got %v", cfg.Enabled)
	}
	if cfg.Rate != 0.5 {
		t.Errorf("expected Rate=0.5, got %f", cfg.Rate)
	}
}

func TestBindProperties_NestedStruct(t *testing.T) {
	t.Parallel()
	type SSLConfig struct {
		Enabled bool   `enhance:"server.ssl.enabled" default:"false"`
		Cert    string `enhance:"server.ssl.cert"`
	}

	type ServerConfig struct {
		Port int `enhance:"server.port" default:"8080"`
		SSL  SSLConfig
	}

	env := environment.NewEnvironment()
	cfg := &ServerConfig{}

	err := BindProperties(cfg, env)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", cfg.Port)
	}
	if cfg.SSL.Enabled != false {
		t.Errorf("expected SSL.Enabled=false, got %v", cfg.SSL.Enabled)
	}
}

func TestBindProperties_WithPropertyValues(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Name string `enhance:"app.name"`
		Port int    `enhance:"server.port"`
	}

	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"app.name":    "my-app",
		"server.port": 9090,
	}))

	cfg := &TestConfig{}
	err := BindProperties(cfg, env)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if cfg.Name != "my-app" {
		t.Errorf("expected Name='my-app', got '%s'", cfg.Name)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", cfg.Port)
	}
}

func TestBindProperties_Slice(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Hosts []string `enhance:"server.hosts" default:"localhost,127.0.0.1"`
	}

	env := environment.NewEnvironment()
	cfg := &TestConfig{}

	err := BindProperties(cfg, env)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if len(cfg.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(cfg.Hosts))
	}
	if cfg.Hosts[0] != "localhost" {
		t.Errorf("expected Hosts[0]='localhost', got '%s'", cfg.Hosts[0])
	}
}

func TestBindProperties_Map(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Labels map[string]string `enhance:"app.labels"`
	}

	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"app.labels": map[string]any{
			"env":  "prod",
			"team": "backend",
		},
	}))

	cfg := &TestConfig{}
	err := BindProperties(cfg, env)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if cfg.Labels["env"] != "prod" {
		t.Errorf("expected Labels['env']='prod', got '%s'", cfg.Labels["env"])
	}
	if cfg.Labels["team"] != "backend" {
		t.Errorf("expected Labels['team']='backend', got '%s'", cfg.Labels["team"])
	}
}

func TestBindProperties_Duration(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Timeout time.Duration `enhance:"test.timeout" default:"30s"`
	}

	env := environment.NewEnvironment()
	cfg := &TestConfig{}

	err := BindProperties(cfg, env)
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected Timeout=30s, got %v", cfg.Timeout)
	}
}

func TestBindProperties_WithPrefix(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		Name string `enhance:"name" default:"test"`
		Port int    `enhance:"port" default:"8080"`
	}

	env := environment.NewEnvironment()
	cfg := &TestConfig{}

	err := BindProperties(cfg, env, WithBindPrefix("server"))
	if err != nil {
		t.Fatalf("BindProperties failed: %v", err)
	}

	// 应该查找 server.name 和 server.port
	if cfg.Name != "test" || cfg.Port != 8080 {
		t.Errorf("expected default values with prefix")
	}
}

func TestBindProperties_IntOverflow(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"server.small": 300,
	}))

	type TestConfig struct {
		Small int8 `enhance:"server.small"`
	}

	cfg := &TestConfig{}
	err := BindProperties(cfg, env)
	if err == nil {
		t.Fatal("expected error for int8 overflow (300)")
	}
}

func TestBindProperties_UintNegative(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"server.count": -1.0,
	}))

	type TestConfig struct {
		Count uint `enhance:"server.count"`
	}

	cfg := &TestConfig{}
	err := BindProperties(cfg, env)
	if err == nil {
		t.Fatal("expected error for negative float64 bound to uint")
	}
}

func TestBindProperties_UintOverflow(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		"server.count": 70000.0,
	}))

	type TestConfig struct {
		Count uint16 `enhance:"server.count"`
	}

	cfg := &TestConfig{}
	err := BindProperties(cfg, env)
	if err == nil {
		t.Fatal("expected error for uint16 overflow (70000)")
	}
}

func TestBindProperties_InvalidTarget(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()

	// 测试非指针
	err := BindProperties("not a pointer", env)
	if err == nil {
		t.Error("expected error for non-pointer target")
	}

	// 测试 nil 指针
	var nilCfg *struct{}
	err = BindProperties(nilCfg, env)
	if err == nil {
		t.Error("expected error for nil pointer")
	}

	// 测试非结构体指针
	str := "test"
	err = BindProperties(&str, env)
	if err == nil {
		t.Error("expected error for non-struct pointer")
	}
}
