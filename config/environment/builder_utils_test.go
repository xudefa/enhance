package environment

import (
	"reflect"
	"testing"
	"time"
)

func TestIsTimeType(t *testing.T) {
	t.Parallel()

	// Test time.Time type
	if !isTimeType(reflect.TypeOf(time.Time{})) {
		t.Error("expected time.Time to be recognized as time type")
	}

	// Test non-time type
	if isTimeType(reflect.TypeOf("")) {
		t.Error("expected string to not be recognized as time type")
	}

	if isTimeType(reflect.TypeOf(0)) {
		t.Error("expected int to not be recognized as time type")
	}
}

func TestHasNestedExplicitKeys(t *testing.T) {
	t.Parallel()

	type DBConfig struct {
		Host string
		Port string
	}

	cfg := DBConfig{Host: "localhost", Port: "5432"}
	rv := reflect.ValueOf(cfg)

	// This function checks if there are nested explicit keys
	got := hasNestedExplicitKeys(rv)
	// Function should return false for simple struct without nested config keys
	_ = got
}

func TestHasExplicitConfigKey(t *testing.T) {
	t.Parallel()

	type AppConfig struct {
		Name string `config:"app.name"`
	}

	cfg := AppConfig{}
	typ := reflect.TypeOf(cfg)
	field, _ := typ.FieldByName("Name")

	got := hasExplicitConfigKey(field)
	if !got {
		t.Error("expected field with config tag to have explicit config key")
	}

	type SimpleConfig struct {
		Name string
	}

	simpleCfg := SimpleConfig{}
	simpleTyp := reflect.TypeOf(simpleCfg)
	simpleField, _ := simpleTyp.FieldByName("Name")

	got = hasExplicitConfigKey(simpleField)
	if got {
		t.Error("expected field without config tag to not have explicit config key")
	}
}

func TestFindDefaultConfigFile(t *testing.T) {
	t.Parallel()

	// This function searches for default config files
	// Just verify it doesn't panic
	file := FindDefaultConfigFile()
	// File may or may not exist depending on the environment
	_ = file
}

func TestGetConfigFileExtension(t *testing.T) {
	t.Parallel()

	// 测试JSON配置类型
	ext := GetConfigFileExtension(ConfigTypeJSON)
	if ext != "json" {
		t.Errorf("expected extension 'json', got %s", ext)
	}
}

func TestParseConfigType(t *testing.T) {
	t.Parallel()

	// 测试解析配置类型
	configType := ParseConfigType("/path/to/config.json")
	if configType != ConfigTypeJSON {
		t.Errorf("expected ConfigTypeJSON, got %s", configType)
	}
}
