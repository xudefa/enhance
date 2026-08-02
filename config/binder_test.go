package config

import (
	"reflect"
	"testing"
	"time"
)

func TestBind_BasicTypes(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("server.host", "localhost")
	cfg.Set("server.port", 8080)
	cfg.Set("server.debug", true)

	type Config struct {
		Host  string `env:"server.host"`
		Port  int    `env:"server.port"`
		Debug bool   `env:"server.debug"`
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if c.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %q", c.Host)
	}
	if c.Port != 8080 {
		t.Errorf("expected port 8080, got %d", c.Port)
	}
	if !c.Debug {
		t.Error("expected debug true")
	}
}

func TestBind_Duration(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("timeout", "30s")

	type Config struct {
		Timeout time.Duration `env:"timeout"`
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if c.Timeout != 30*time.Second {
		t.Errorf("expected 30s, got %v", c.Timeout)
	}
}

func TestBind_StringList(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("hosts", "host1,host2,host3")

	type Config struct {
		Hosts []string `env:"hosts"`
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if len(c.Hosts) != 3 {
		t.Fatalf("expected 3 hosts, got %d", len(c.Hosts))
	}
	if c.Hosts[0] != "host1" || c.Hosts[1] != "host2" || c.Hosts[2] != "host3" {
		t.Errorf("unexpected hosts: %v", c.Hosts)
	}
}

func TestBind_StringMap(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("labels", "env=prod,region=us-east")

	type Config struct {
		Labels map[string]string `env:"labels"`
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if c.Labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %v", c.Labels)
	}
	if c.Labels["region"] != "us-east" {
		t.Errorf("expected region=us-east, got %v", c.Labels)
	}
}

func TestBind_DefaultValue(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()

	type Config struct {
		Host string `env:"server.host" default:"localhost"`
		Port int    `env:"server.port" default:"3000"`
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if c.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %q", c.Host)
	}
	if c.Port != 3000 {
		t.Errorf("expected port 3000, got %d", c.Port)
	}
}

func TestBind_NestedStruct(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("server.http.port", 8080)
	cfg.Set("server.tls.enabled", true)

	type ServerConfig struct {
		HTTP struct {
			Port int `env:"server.http.port"`
		}
		TLS struct {
			Enabled bool `env:"server.tls.enabled"`
		}
	}

	var c ServerConfig
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if c.HTTP.Port != 8080 {
		t.Errorf("expected http port 8080, got %d", c.HTTP.Port)
	}
	if !c.TLS.Enabled {
		t.Error("expected tls enabled")
	}
}

func TestBind_FlatNestedStruct(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("host", "localhost")
	cfg.Set("port", 9090)

	type Config struct {
		Server struct {
			Host string `env:"host"`
			Port int    `env:"port"`
		}
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if c.Server.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %q", c.Server.Host)
	}
	if c.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", c.Server.Port)
	}
}

func TestBind_NonStructTarget(t *testing.T) {
	t.Parallel()
	var s string
	err := Bind(NewConfig(), &s)
	if err == nil {
		t.Error("expected error for non-struct target")
	}
}

func TestBind_IntOverflow(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("small", 300)

	type Config struct {
		Small int8 `env:"small"`
	}

	var c Config
	err := Bind(cfg, &c)
	if err == nil {
		t.Fatal("expected error for int8 overflow (300)")
	}
}

func TestBind_Int16Overflow(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("port", 70000)

	type Config struct {
		Port int16 `env:"port"`
	}

	var c Config
	err := Bind(cfg, &c)
	if err == nil {
		t.Fatal("expected error for int16 overflow (70000)")
	}
}

func TestBind_UintOverflow(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("count", 500)

	type Config struct {
		Count uint8 `env:"count"`
	}

	var c Config
	err := Bind(cfg, &c)
	if err == nil {
		t.Fatal("expected error for uint8 overflow (500)")
	}
}

func TestBind_UintNegative(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("count", -1)

	type Config struct {
		Count uint `env:"count"`
	}

	var c Config
	err := Bind(cfg, &c)
	if err == nil {
		t.Fatal("expected error for negative value bound to uint")
	}
}

func TestBind_InRangeStillWorks(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("small", 100)

	type Config struct {
		Small int8 `env:"small"`
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}
	if c.Small != 100 {
		t.Errorf("expected 100, got %d", c.Small)
	}
}

func TestValidate_Required(t *testing.T) {
	t.Parallel()
	type Config struct {
		Host string `validate:"required"`
		Port int    `validate:"required"`
	}

	c := Config{Host: "localhost", Port: 8080}
	if err := Validate(&c); err != nil {
		t.Fatal(err)
	}

	c2 := Config{Host: "", Port: 0}
	err := Validate(&c2)
	if err == nil {
		t.Error("expected validation error for empty required fields")
	}
}

func TestValidate_MinMax(t *testing.T) {
	t.Parallel()
	type Config struct {
		Port int    `validate:"min=1,max=65535"`
		Name string `validate:"min=1,max=10"`
	}

	c := Config{Port: 8080, Name: "server"}
	if err := Validate(&c); err != nil {
		t.Fatal(err)
	}

	c2 := Config{Port: 0, Name: "server"}
	err := Validate(&c2)
	if err == nil {
		t.Error("expected validation error for port below min")
	}

	c3 := Config{Port: 8080, Name: ""}
	err = Validate(&c3)
	if err == nil {
		t.Error("expected validation error for name below min length")
	}

	c4 := Config{Port: 70000, Name: "server"}
	err = Validate(&c4)
	if err == nil {
		t.Error("expected validation error for port above max")
	}
}

func TestValidate_Enum(t *testing.T) {
	t.Parallel()
	type Config struct {
		Env string `validate:"enum=dev|staging|prod"`
	}

	c := Config{Env: "prod"}
	if err := Validate(&c); err != nil {
		t.Fatal(err)
	}

	c2 := Config{Env: "test"}
	err := Validate(&c2)
	if err == nil {
		t.Error("expected validation error for invalid enum value")
	}
}

func TestValidate_NestedStruct(t *testing.T) {
	t.Parallel()
	type Config struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"min=1,max=65535"`
		}
	}

	c := Config{}
	c.Server.Host = "localhost"
	c.Server.Port = 8080

	if err := Validate(&c); err != nil {
		t.Fatal(err)
	}

	c2 := Config{}
	c2.Server.Port = 0
	err := Validate(&c2)
	if err == nil {
		t.Error("expected validation error for empty required host")
	}
}

func TestBindAndValidate_Integration(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	cfg.Set("server.host", "localhost")
	cfg.Set("server.port", 8080)
	cfg.Set("server.timeout", "30s")
	cfg.Set("server.allowed_hosts", "localhost,127.0.0.1")

	type ServerConfig struct {
		Host         string        `env:"server.host" validate:"required"`
		Port         int           `env:"server.port" validate:"min=1,max=65535"`
		Timeout      time.Duration `env:"server.timeout"`
		AllowedHosts []string      `env:"server.allowed_hosts"`
	}

	var c ServerConfig
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if err := Validate(&c); err != nil {
		t.Fatal(err)
	}

	if c.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %q", c.Host)
	}
	if c.Port != 8080 {
		t.Errorf("expected port 8080, got %d", c.Port)
	}
	if c.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", c.Timeout)
	}
	if len(c.AllowedHosts) != 2 {
		t.Errorf("expected 2 allowed hosts, got %d", len(c.AllowedHosts))
	}
}

// CustomTestType 用于测试的自定义类型
type CustomTestType struct {
	Value string
}

func TestRegisterConverter_Custom(t *testing.T) {
	t.Parallel()
	RegisterConverter(CustomTestType{}, func(s string) (any, error) {
		return CustomTestType{Value: "custom:" + s}, nil
	})

	// 验证转换器已注册
	targetType := reflect.TypeOf(CustomTestType{})
	converter, ok := GetConverter(targetType)
	if !ok {
		t.Fatal("converter not registered")
	}
	_ = converter

	cfg := NewConfig()
	cfg.Set("custom", "test")

	type Config struct {
		Custom CustomTestType `env:"custom"`
	}

	var c Config
	if err := Bind(cfg, &c); err != nil {
		t.Fatal(err)
	}

	if c.Custom.Value != "custom:test" {
		t.Errorf("expected 'custom:test', got %q", c.Custom.Value)
	}
}
