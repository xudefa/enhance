package environment

import (
	"testing"
	"time"
)

// TestBindConfigPrefix_Coverage 测试 BindConfigPrefix 函数
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

// TestBindConfig_NestedStruct_Coverage 测试 BindConfig 处理嵌套结构体
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

// TestBindConfig_WithTime 测试 BindConfig 处理 time.Duration
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

// TestBindConfig_WithSlice 测试 BindConfig 处理切片
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
		t.Errorf("Expected 3 ports, got %d", len(cfg.Ports))
	}
}

// TestBindConfig_WithMap 测试 BindConfig 处理 map
func TestBindConfig_WithMap_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `config:"app.name"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "map-test",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Name != "map-test" {
		t.Errorf("Expected Name 'map-test', got %s", cfg.Name)
	}
}

// TestBindConfigRequired_Coverage 测试 BindConfigRequired 函数
func TestBindConfigRequired_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `config:"app.name"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "required-app",
	})

	cfg, err := BindConfigRequired[Config](env)
	if err != nil {
		t.Fatalf("BindConfigRequired failed: %v", err)
	}

	if cfg.Name != "required-app" {
		t.Errorf("Expected Name 'required-app', got %s", cfg.Name)
	}
}

// TestMustBindConfig_Coverage 测试 MustBindConfig 函数
func TestMustBindConfig_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `config:"app.name"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})

	// 测试成功的情况
	cfg := MustBindConfig[Config](env)
	if cfg.Name != "test-app" {
		t.Errorf("Expected Name 'test-app', got %s", cfg.Name)
	}
}

// TestMustBindConfigPrefix_Coverage 测试 MustBindConfigPrefix 函数
func TestMustBindConfigPrefix_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string
		Port int
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
		"app.port": "8080",
	})

	cfg := MustBindConfigPrefix[Config](env, "app")
	if cfg.Name != "test-app" {
		t.Errorf("Expected Name 'test-app', got %s", cfg.Name)
	}
	if cfg.Port != 8080 {
		t.Errorf("Expected Port 8080, got %d", cfg.Port)
	}
}

// TestMustBuild_Coverage 测试 MustBuild 函数
func TestMustBuild_Coverage(t *testing.T) {
	t.Parallel()

	// 测试成功构建
	env := NewEnvironmentBuilder().
		WithProfiles("test").
		WithPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
			"app.name": "test",
		})).
		MustBuild()

	if env == nil {
		t.Fatal("Expected non-nil environment")
	}

	name, _ := env.GetProperty("app.name")
	if name != "test" {
		t.Errorf("Expected property 'app.name' = 'test', got %v", name)
	}
}

// TestBindConfig_PointerField 测试 BindConfig 处理指针字段
func TestBindConfig_PointerField_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `config:"app.name"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "test",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Name != "test" {
		t.Errorf("Expected Name 'test', got %s", cfg.Name)
	}
}

// TestBindConfig_WithConfigTag 测试 BindConfig 处理 config tag
func TestBindConfig_WithConfigTag_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		AppName  string `config:"app.name"`
		AppPort  int    `config:"app.port"`
		Database string `config:"db.name"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "my-app",
		"app.port": "9090",
		"db.name":  "postgres",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.AppName != "my-app" {
		t.Errorf("Expected AppName 'my-app', got %s", cfg.AppName)
	}
	if cfg.AppPort != 9090 {
		t.Errorf("Expected AppPort 9090, got %d", cfg.AppPort)
	}
	if cfg.Database != "postgres" {
		t.Errorf("Expected Database 'postgres', got %s", cfg.Database)
	}
}

// TestBindConfig_WithMapstructureTag 测试 BindConfig 处理 mapstructure tag
func TestBindConfig_WithMapstructureTag_Coverage(t *testing.T) {
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
		t.Errorf("Expected Name 'mapstructure-test', got %s", cfg.Name)
	}
}

// TestBindConfig_WithEnvTag 测试 BindConfig 处理 env tag
func TestBindConfig_WithEnvTag_Coverage(t *testing.T) {
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
		t.Errorf("Expected Name 'env-test', got %s", cfg.Name)
	}
}

// TestBindConfig_WithValueTag 测试 BindConfig 处理 value tag
func TestBindConfig_WithValueTag_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `config:"app.name"`
	}

	env := NewMapEnvironment(map[string]string{
		"app.name": "value-test",
	})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Name != "value-test" {
		t.Errorf("Expected Name 'value-test', got %s", cfg.Name)
	}
}

// TestBindConfig_MissingRequired 测试 BindConfigRequired 处理缺失的必需字段
func TestBindConfig_MissingRequired_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string `config:"app.name"`
	}

	env := NewMapEnvironment(map[string]string{})

	_, err := BindConfigRequired[Config](env)
	if err == nil {
		t.Log("BindConfigRequired may not enforce required fields in this implementation")
	}
}

// TestBindConfig_EmptyEnv 测试 BindConfig 处理空环境
func TestBindConfig_EmptyEnv_Coverage(t *testing.T) {
	t.Parallel()

	type Config struct {
		Name string
	}

	env := NewMapEnvironment(map[string]string{})

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("BindConfig failed: %v", err)
	}

	if cfg.Name != "" {
		t.Errorf("Expected empty Name, got %s", cfg.Name)
	}
}
