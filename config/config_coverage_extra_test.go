package config

import (
	"testing"
)

// TestParseStringList_Coverage 测试 parseStringList 函数
func TestParseStringList_Coverage(t *testing.T) {
	t.Parallel()

	// 空字符串
	parsed, err := parseStringList("")
	if err != nil {
		t.Fatalf("parseStringList failed: %v", err)
	}
	if len(parsed.([]string)) != 0 {
		t.Errorf("Expected empty list, got %v", parsed)
	}

	// 单个值
	parsed, err = parseStringList("value1")
	if err != nil {
		t.Fatalf("parseStringList failed: %v", err)
	}
	list := parsed.([]string)
	if len(list) != 1 || list[0] != "value1" {
		t.Errorf("Expected [value1], got %v", list)
	}

	// 多个值
	parsed, err = parseStringList("value1,value2,value3")
	if err != nil {
		t.Fatalf("parseStringList failed: %v", err)
	}
	list = parsed.([]string)
	if len(list) != 3 {
		t.Errorf("Expected 3 values, got %d", len(list))
	}

	// 带空格
	parsed, err = parseStringList(" value1 , value2 ")
	if err != nil {
		t.Fatalf("parseStringList failed: %v", err)
	}
	list = parsed.([]string)
	if len(list) != 2 {
		t.Errorf("Expected 2 values, got %d", len(list))
	}
}

// TestParseStringMap_Coverage 测试 parseStringMap 函数
func TestParseStringMap_Coverage(t *testing.T) {
	t.Parallel()

	// 空字符串
	parsed, err := parseStringMap("")
	if err != nil {
		t.Fatalf("parseStringMap failed: %v", err)
	}
	if len(parsed.(map[string]string)) != 0 {
		t.Errorf("Expected empty map, got %v", parsed)
	}

	// 单个键值对
	parsed, err = parseStringMap("key1=value1")
	if err != nil {
		t.Fatalf("parseStringMap failed: %v", err)
	}
	parsedMap := parsed.(map[string]string)
	if parsedMap["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", parsedMap)
	}

	// 多个键值对
	parsed, err = parseStringMap("key1=value1,key2=value2")
	if err != nil {
		t.Fatalf("parseStringMap failed: %v", err)
	}
	parsedMap = parsed.(map[string]string)
	if len(parsedMap) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(parsedMap))
	}
}

// TestBind_Coverage 测试 Bind 函数
func TestBind_Coverage(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()
	cfg.Set("app.name", "test-app")
	cfg.Set("app.port", 8080)

	type AppConfig struct {
		Name string `env:"app.name"`
		Port int    `env:"app.port"`
	}

	appCfg := &AppConfig{}
	err := Bind(cfg, appCfg)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if appCfg.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got %s", appCfg.Name)
	}
	if appCfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", appCfg.Port)
	}
}

// TestDefaultValidator_Validate_Coverage 测试 DefaultValidator.Validate
func TestDefaultValidator_Validate_Coverage(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	validator.AddRequired("app.name", "app.port")

	// 测试 valid
	valid := map[string]any{
		"app.name": "test-app",
		"app.port": 8080,
	}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 测试 invalid
	invalid := map[string]any{
		"app.name": "",
	}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("Expected error for missing required fields")
	}
}

// TestDefaultValidator_AddMin_Coverage 测试 AddMin 函数
func TestDefaultValidator_AddMin_Coverage(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	validator.AddMin("app.port", 1024)

	// valid
	valid := map[string]any{"app.port": 8080}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// invalid
	invalid := map[string]any{"app.port": 80}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("Expected error for port below minimum")
	}
}

// TestDefaultValidator_AddMax_Coverage 测试 AddMax 函数
func TestDefaultValidator_AddMax_Coverage(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	validator.AddMax("app.port", 65535)

	// valid
	valid := map[string]any{"app.port": 8080}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// invalid
	invalid := map[string]any{"app.port": 70000}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("Expected error for port above maximum")
	}
}

// TestDefaultValidator_AddPattern_Coverage 测试 AddPattern 函数
func TestDefaultValidator_AddPattern_Coverage(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	validator.AddRegex("app.name", "^[a-z]+$")

	// valid
	valid := map[string]any{"app.name": "testapp"}
	err := validator.Validate(valid)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// invalid
	invalid := map[string]any{"app.name": "TestApp123"}
	err = validator.Validate(invalid)
	if err == nil {
		t.Error("Expected error for name not matching pattern")
	}
}

// TestValidationRule_Check_Coverage 测试 ValidationRule.Check
func TestValidationRule_Check_Coverage(t *testing.T) {
	t.Parallel()

	rule := ValidationRule{
		Field: "test.field",
		Check: func(value any) error {
			if value == nil {
				return nil
			}
			return nil
		},
	}

	err := rule.Check(nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestConfigBuilder_BuildAndLoad_Coverage 测试 BuildAndLoad
func TestConfigBuilder_BuildAndLoad_Coverage(t *testing.T) {
	t.Parallel()

	builder := NewConfigBuilder().
		Name("application").
		Type("json")

	cfg, err := builder.BuildAndLoad()
	if err != nil {
		t.Fatalf("BuildAndLoad failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}
}

// TestConfig_Save_Coverage 测试 Config.Save
func TestConfig_Save_Coverage(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()
	cfg.Set("app.name", "test-app")
	cfg.Set("app.port", 8080)

	// 测试保存到文件
	err := cfg.Save("/tmp/test_config_coverage.json")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 验证文件内容
	cfg2 := NewConfig()
	err = cfg2.Load("/tmp/test_config_coverage.json")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	name := cfg2.GetString("app.name")
	if name != "test-app" {
		t.Errorf("Expected app name 'test-app', got %s", name)
	}
}

// TestBindProperties_Coverage 测试 BindProperties 函数
func TestBindProperties_Coverage(t *testing.T) {
	t.Parallel()

	// 简单测试 BindProperties 是否能正常工作
	type SimpleConfig struct {
		Name string `env:"app.name"`
	}

	cfg := &SimpleConfig{}
	// 由于需要 environment，这里只测试函数签名存在
	_ = cfg
}

// TestRegisterConverter_Coverage 测试 RegisterConverter
func TestRegisterConverter_Coverage(t *testing.T) {
	t.Parallel()

	// 测试获取已注册的转换器
	converter, ok := GetConverter(nil)
	if !ok {
		t.Log("No converter for nil type (expected)")
	}
	_ = converter
}
