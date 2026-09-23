package environment

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// ==================== Builder Core Tests ====================

func TestEnvironmentBuilder_WithJSONConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test-config.json")
	content := `{"test.key": "test-value", "app.name": "test-app"}`
	err := os.WriteFile(jsonFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	builder := NewEnvironmentBuilder().WithJSONConfig(jsonFile)
	if builder == nil {
		t.Fatal("expected builder to be created")
	}

	if len(builder.propertySources) == 0 {
		t.Error("expected builder to have property sources from JSON config")
	}

	env := builder.Build()
	if env == nil {
		t.Fatal("expected environment to be built")
	}
}

func TestEnvironmentBuilder_WithJSONConfig_NonExistent(t *testing.T) {
	t.Parallel()

	builder := NewEnvironmentBuilder().WithJSONConfig("/non/existent/file.json")
	env := builder.Build()
	if env == nil {
		t.Fatal("expected environment to be built even with non-existent file")
	}
}

func TestEnvironmentBuilder_MustBuild(t *testing.T) {
	t.Parallel()

	env := NewEnvironmentBuilder().MustBuild()
	if env == nil {
		t.Fatal("expected environment to be built")
	}
}

func TestEnvironmentBuilder_ChainedMethods(t *testing.T) {
	t.Parallel()

	env := NewEnvironmentBuilder().
		WithProfile("dev").
		WithProfiles("test", "staging").
		WithEnvPrefix("TEST").
		WithArgs("--app.name=test", "--app.port=8080").
		Build()

	if env == nil {
		t.Fatal("expected environment to be built")
	}

	if !env.AcceptsProfile("dev") {
		t.Error("expected 'dev' profile to be active")
	}
	if !env.AcceptsProfile("test") {
		t.Error("expected 'test' profile to be active")
	}
}

func TestEnvironmentBuilder_WithPropertySource(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key": "value",
	})

	builder := NewEnvironmentBuilder().WithPropertySource(source)
	if len(builder.propertySources) != 1 {
		t.Errorf("expected 1 property source, got %d", len(builder.propertySources))
	}

	env := builder.Build()
	value, ok := env.GetProperty("key")
	if !ok || value != "value" {
		t.Errorf("expected 'value', got %v", value)
	}
}

func TestEnvironmentBuilder_WithPropertySourceFirst(t *testing.T) {
	t.Parallel()

	source1 := NewMapPropertySource("first", PriorityNormal, map[string]any{
		"key": "first-value",
	})
	source2 := NewMapPropertySource("second", PriorityNormal, map[string]any{
		"key": "second-value",
	})

	builder := NewEnvironmentBuilder().
		WithPropertySource(source1).
		WithPropertySourceFirst(source2)

	if len(builder.propertySources) != 2 {
		t.Fatalf("expected 2 property sources, got %d", len(builder.propertySources))
	}
	if builder.propertySources[0].Name() != "second" {
		t.Errorf("expected first source to be 'second', got %s", builder.propertySources[0].Name())
	}
}

// ==================== Builder Config Option Tests ====================

func TestWithProfiles(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opt := WithProfiles("dev", "test")
	opt(config)

	if len(config.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(config.Profiles))
	}
}

func TestWithDefaultProfile(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opt := WithDefaultProfile("prod")
	opt(config)

	if config.DefaultProfile != "prod" {
		t.Errorf("expected 'prod', got %s", config.DefaultProfile)
	}
}

func TestWithAutoDetectProfiles(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opt := WithAutoDetectProfiles(true)
	opt(config)

	if !config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be true")
	}
}

func TestWithPropertySources(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	source := NewMapPropertySource("test", PriorityNormal, map[string]any{})
	opt := WithPropertySources(source)
	opt(config)

	if len(config.PropertySources) != 1 {
		t.Errorf("expected 1 property source, got %d", len(config.PropertySources))
	}
}

func TestDefaultEnvironmentConfig(t *testing.T) {
	t.Parallel()

	config := DefaultEnvironmentConfig()

	if config.DefaultProfile != "default" {
		t.Errorf("expected 'default', got %s", config.DefaultProfile)
	}
	if !config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be true")
	}
}

func TestEnvironmentConfig_ApplyOptions(t *testing.T) {
	t.Parallel()

	config := &EnvironmentConfig{}
	opts := []EnvironmentOption{
		WithProfiles("dev"),
		WithDefaultProfile("default"),
		WithAutoDetectProfiles(false),
	}

	config.ApplyOptions(opts)

	if len(config.Profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(config.Profiles))
	}
	if config.DefaultProfile != "default" {
		t.Errorf("expected 'default', got %s", config.DefaultProfile)
	}
	if config.AutoDetectProfiles {
		t.Error("expected AutoDetectProfiles to be false")
	}
}

func TestCreateEnvironment(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key": "value",
	})

	env := CreateEnvironment(
		WithProfiles("dev"),
		WithDefaultProfile("default"),
		WithPropertySources(source),
	)

	if env == nil {
		t.Fatal("expected environment to be created")
	}

	if !env.AcceptsProfile("dev") {
		t.Error("expected 'dev' profile to be active")
	}

	value, ok := env.GetProperty("key")
	if !ok || value != "value" {
		t.Errorf("expected 'value', got %v", value)
	}
}

// ==================== Environment Operations Tests ====================

func TestEnvironment_AddPropertySourceFirst(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("first", PriorityNormal, map[string]any{
		"key": "first-value",
	}))
	env.AddPropertySourceFirst(NewMapPropertySource("second", PriorityNormal, map[string]any{
		"key": "second-value",
	}))

	value, ok := env.GetProperty("key")
	if !ok {
		t.Fatal("expected 'key' to exist")
	}
	if value != "second-value" {
		t.Errorf("expected 'second-value' (higher priority), got %v", value)
	}
}

func TestEnvironment_RemovePropertySource(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key": "value",
	}))

	_, ok := env.GetProperty("key")
	if !ok {
		t.Fatal("expected 'key' to exist before removal")
	}

	env.RemovePropertySource("test")

	_, ok = env.GetProperty("key")
	if ok {
		t.Error("expected 'key' to not exist after removal")
	}
}

func TestEnvironment_RemoveProfile(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("dev")
	env.AddActiveProfile("test")

	if !env.AcceptsProfile("dev") {
		t.Fatal("expected 'dev' profile to exist before removal")
	}

	env.RemoveProfile("dev")

	if env.AcceptsProfile("dev") {
		t.Error("expected 'dev' profile to not exist after removal")
	}

	if !env.AcceptsProfile("test") {
		t.Error("expected 'test' profile to still exist")
	}
}

func TestEnvironment_GetPropertySources(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	initialCount := len(env.GetPropertySources())

	source1 := NewMapPropertySource("source1", PriorityNormal, map[string]any{})
	source2 := NewMapPropertySource("source2", PriorityNormal, map[string]any{})
	env.AddPropertySource(source1)
	env.AddPropertySource(source2)

	sources := env.GetPropertySources()
	if len(sources) != initialCount+2 {
		t.Errorf("expected %d property sources, got %d", initialCount+2, len(sources))
	}
}

// ==================== Property Source Tests ====================

func TestPropertySource_Contains(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key1": "value1",
		"key2": "value2",
	})

	if !source.Contains("key1") {
		t.Error("expected source to contain 'key1'")
	}

	if source.Contains("nonexistent") {
		t.Error("expected source to not contain 'nonexistent'")
	}
}

func TestPropertySource_Keys(t *testing.T) {
	t.Parallel()

	source := NewMapPropertySource("test", PriorityNormal, map[string]any{
		"key1": "value1",
		"key2": "value2",
	})

	keys := source.Keys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestEnvPropertySource_Name(t *testing.T) {
	t.Parallel()

	source := NewEnvPropertySource("test-env", "TEST")
	if source.Name() != "test-env" {
		t.Errorf("expected name 'test-env', got %s", source.Name())
	}
}

func TestEnvPropertySource_Contains(t *testing.T) {
	t.Setenv("TEST_KEY", "test-value")

	source := NewEnvPropertySource("test-env", "TEST")
	if !source.Contains("key") {
		t.Error("expected env source to contain 'key' (mapped to 'TEST_KEY')")
	}
}

func TestArgsPropertySource_Name(t *testing.T) {
	t.Parallel()

	source := NewArgsPropertySource("test-args", []string{"--key=value"})
	if source.Name() != "test-args" {
		t.Errorf("expected name 'test-args', got %s", source.Name())
	}
}

func TestArgsPropertySource_Contains(t *testing.T) {
	t.Parallel()

	source := NewArgsPropertySource("test-args", []string{"--key=value"})
	if !source.Contains("key") {
		t.Error("expected args source to contain 'key'")
	}
}

// ==================== Environment Helper Tests ====================

func TestEnvironmentHelper_WithPrefix(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
		"app.port": "8080",
	}))

	helper := NewEnvironmentHelper(env)
	prefixedHelper := helper.WithPrefix("app")

	if prefixedHelper == nil {
		t.Fatal("expected prefixed helper to be created")
	}
}

func TestEnvironmentHelper_GetString(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetString("app.name", "default")

	if value != "test-app" {
		t.Errorf("expected 'test-app', got %v", value)
	}
}

func TestEnvironmentHelper_GetString_Default(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	helper := NewEnvironmentHelper(env)

	value := helper.GetString("non.existent", "default-value")
	if value != "default-value" {
		t.Errorf("expected 'default-value', got %v", value)
	}
}

func TestEnvironmentHelper_ContainsProperty(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)

	if !helper.ContainsProperty("app.name") {
		t.Error("expected 'app.name' to exist")
	}

	if helper.ContainsProperty("non.existent") {
		t.Error("expected 'non.existent' to not exist")
	}
}

func TestEnvironmentHelper_GetInt(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.port": 8080,
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetInt("app.port", 3000)

	if value != 8080 {
		t.Errorf("expected 8080, got %d", value)
	}
}

func TestEnvironmentHelper_GetBool(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.debug": true,
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetBool("app.debug", false)

	if !value {
		t.Error("expected true, got false")
	}
}

func TestEnvironmentHelper_GetFloat64(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.ratio": 1.5,
	}))

	helper := NewEnvironmentHelper(env)
	value := helper.GetFloat64("app.ratio", 1.0)

	if value != 1.5 {
		t.Errorf("expected 1.5, got %f", value)
	}
}

func TestEnvironmentHelper_IsDev(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("dev")

	helper := NewEnvironmentHelper(env)
	if !helper.IsDev() {
		t.Error("expected IsDev to return true")
	}
}

func TestEnvironmentHelper_IsProd(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("prod")

	helper := NewEnvironmentHelper(env)
	if !helper.IsProd() {
		t.Error("expected IsProd to return true")
	}
}

func TestEnvironmentHelper_IsTest(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("test")

	helper := NewEnvironmentHelper(env)
	if !helper.IsTest() {
		t.Error("expected IsTest to return true")
	}
}

func TestEnvironmentHelper_GetActiveProfile(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddActiveProfile("staging")

	helper := NewEnvironmentHelper(env)
	profile := helper.GetActiveProfile()

	if profile != "staging" {
		t.Errorf("expected 'staging', got %s", profile)
	}
}

func TestEnvironmentHelper_GetRequiredProperty(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)

	prop, err := helper.GetRequiredProperty("app.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prop != "test-app" {
		t.Errorf("expected 'test-app', got %v", prop)
	}

	_, err = helper.GetRequiredProperty("non.existent")
	if err == nil {
		t.Error("expected error for non-existent property")
	}
}

func TestEnvironmentHelper_WithPrefix_GetString(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
		"app.port": "8080",
	}))

	helper := NewEnvironmentHelper(env)
	prefixedHelper := helper.WithPrefix("app")

	if prefixedHelper == nil {
		t.Fatal("expected prefixed helper to be created")
	}

	value := prefixedHelper.GetString("name", "default")
	if value != "test-app" {
		t.Errorf("expected 'test-app', got %v", value)
	}
}

// ==================== Environment Template Tests ====================

func TestEnvironmentTemplate(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.url":  "postgres://localhost:5432/test",
		"database.host": "localhost",
		"database.port": 5432,
	}))

	template := NewEnvironmentTemplate(env)

	url := template.GetDatabaseURL("default")
	if url != "postgres://localhost:5432/test" {
		t.Errorf("expected database URL 'postgres://localhost:5432/test', got %s", url)
	}

	host := template.GetDatabaseHost("default")
	if host != "localhost" {
		t.Errorf("expected database host 'localhost', got %s", host)
	}

	port := template.GetDatabasePort(0)
	if port != 5432 {
		t.Errorf("expected database port 5432, got %d", port)
	}
}

func TestEnvironmentTemplate_GetDatabaseName(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"database.name": "mydb",
	}))

	template := NewEnvironmentTemplate(env)
	name := template.GetDatabaseName("default")

	if name != "mydb" {
		t.Errorf("expected 'mydb', got %s", name)
	}
}

func TestEnvironmentTemplate_GetServerHost(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.host": "0.0.0.0",
	}))

	template := NewEnvironmentTemplate(env)
	host := template.GetServerHost("default")

	if host != "0.0.0.0" {
		t.Errorf("expected '0.0.0.0', got %s", host)
	}
}

func TestEnvironmentTemplate_GetServerPort(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"server.port": 8080,
	}))

	template := NewEnvironmentTemplate(env)
	port := template.GetServerPort(3000)

	if port != 8080 {
		t.Errorf("expected 8080, got %d", port)
	}
}

func TestEnvironmentTemplate_GetLogLevel(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"log.level": "debug",
	}))

	template := NewEnvironmentTemplate(env)
	level := template.GetLogLevel("info")

	if level != "debug" {
		t.Errorf("expected 'debug', got %s", level)
	}
}

func TestEnvironmentTemplate_GetRedisHost(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.host": "127.0.0.1",
	}))

	template := NewEnvironmentTemplate(env)
	host := template.GetRedisHost("localhost")

	if host != "127.0.0.1" {
		t.Errorf("expected '127.0.0.1', got %s", host)
	}
}

func TestEnvironmentTemplate_GetRedisPort(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.port": 6379,
	}))

	template := NewEnvironmentTemplate(env)
	port := template.GetRedisPort(6380)

	if port != 6379 {
		t.Errorf("expected 6379, got %d", port)
	}
}

func TestEnvironmentTemplate_GetRedisPassword(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"redis.password": "secret",
	}))

	template := NewEnvironmentTemplate(env)
	password := template.GetRedisPassword("")

	if password != "secret" {
		t.Errorf("expected 'secret', got %s", password)
	}
}

func TestEnvironmentTemplate_IsDebugMode(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"debug": true,
	}))

	template := NewEnvironmentTemplate(env)
	if !template.IsDebugMode() {
		t.Error("expected debug mode to be true")
	}
}

func TestEnvironmentTemplate_IsVerbose(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"verbose": true,
	}))

	template := NewEnvironmentTemplate(env)
	if !template.IsVerbose() {
		t.Error("expected verbose mode to be true")
	}
}

// ==================== Type Converter Tests ====================

func TestTypeConverter_ToSlice(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	converted, err := converter.ConvertTo("a,b,c", reflect.TypeOf([]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slice, ok := converted.Interface().([]string)
	if !ok {
		t.Fatal("expected result to be []string")
	}

	if len(slice) != 3 {
		t.Errorf("expected 3 elements, got %d", len(slice))
	}
}

func TestTypeConverter_ToUint(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	converted, err := converter.ConvertTo("42", reflect.TypeOf(uint(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := converted.Interface().(uint)
	if !ok {
		t.Fatal("expected result to be uint")
	}

	if typed != 42 {
		t.Errorf("expected 42, got %d", typed)
	}
}

func TestTypeConverter_ToFloat(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	converted, err := converter.ConvertTo("3.14", reflect.TypeOf(float64(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := converted.Interface().(float64)
	if !ok {
		t.Fatal("expected result to be float64")
	}

	if typed != 3.14 {
		t.Errorf("expected 3.14, got %f", typed)
	}
}

func TestTypeConverter_ConvertNumeric(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	converted, err := converter.ConvertTo("127", reflect.TypeOf(int8(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := converted.Interface().(int8)
	if !ok {
		t.Fatal("expected result to be int8")
	}

	if typed != 127 {
		t.Errorf("expected 127, got %d", typed)
	}
}

func TestTypeConverter_SpecialConvert(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	converted, err := converter.ConvertTo("5s", reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := converted.Interface().(time.Duration)
	if !ok {
		t.Fatal("expected result to be time.Duration")
	}

	if typed != 5*time.Second {
		t.Errorf("expected 5s, got %v", typed)
	}
}

// ==================== Utility Function Tests ====================

func TestIsTimeType(t *testing.T) {
	t.Parallel()

	if !isTimeType(reflect.TypeOf(time.Time{})) {
		t.Error("expected time.Time to be recognized as time type")
	}

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

	got := hasNestedExplicitKeys(rv)
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

	file := FindDefaultConfigFile()
	_ = file
}

func TestGetConfigFileExtension(t *testing.T) {
	t.Parallel()

	ext := GetConfigFileExtension(ConfigTypeJSON)
	if ext != "json" {
		t.Errorf("expected extension 'json', got %s", ext)
	}
}

func TestParseConfigType(t *testing.T) {
	t.Parallel()

	configType := ParseConfigType("/path/to/config.json")
	if configType != ConfigTypeJSON {
		t.Errorf("expected ConfigTypeJSON, got %s", configType)
	}
}
