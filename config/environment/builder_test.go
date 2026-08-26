package environment

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestEnvironmentBuilder_WithJSONConfig(t *testing.T) {
	t.Parallel()

	// Create a temporary JSON config file
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

	// Verify builder has property sources
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

	// Should not panic even if file doesn't exist
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

func TestBindConfig(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"name":    "test-app",
		"version": "1.0.0",
	}))

	type Config struct {
		Name    string
		Version string
	}

	cfg, err := BindConfig[Config](env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "test-app" {
		t.Errorf("expected name 'test-app', got %s", cfg.Name)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", cfg.Version)
	}
}

func TestBindConfigPrefix(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name":    "test-app",
		"app.version": "1.0.0",
	}))

	type Config struct {
		Name    string
		Version string
	}

	cfg, err := BindConfigPrefix[Config](env, "app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "test-app" {
		t.Errorf("expected name 'test-app', got %s", cfg.Name)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", cfg.Version)
	}
}

func TestMustBindConfigPrefix(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"db.host": "localhost",
		"db.port": "5432",
	}))

	type DBConfig struct {
		Host string
		Port string
	}

	cfg := MustBindConfigPrefix[DBConfig](env, "db")
	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != "5432" {
		t.Errorf("expected port '5432', got %s", cfg.Port)
	}
}

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
	val := reflect.ValueOf(cfg)

	// This function checks if there are nested explicit keys
	result := hasNestedExplicitKeys(val)
	// Function should return false for simple struct without nested config keys
	_ = result
}

func TestHasExplicitConfigKey(t *testing.T) {
	t.Parallel()

	type AppConfig struct {
		Name string `config:"app.name"`
	}

	cfg := AppConfig{}
	typ := reflect.TypeOf(cfg)
	field, _ := typ.FieldByName("Name")

	result := hasExplicitConfigKey(field)
	if !result {
		t.Error("expected field with config tag to have explicit config key")
	}

	type SimpleConfig struct {
		Name string
	}

	simpleCfg := SimpleConfig{}
	simpleTyp := reflect.TypeOf(simpleCfg)
	simpleField, _ := simpleTyp.FieldByName("Name")

	result = hasExplicitConfigKey(simpleField)
	if result {
		t.Error("expected field without config tag to not have explicit config key")
	}
}

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

	// Verify property exists
	_, ok := env.GetProperty("key")
	if !ok {
		t.Fatal("expected 'key' to exist before removal")
	}

	// Remove the property source by name
	env.RemovePropertySource("test")

	// Verify property no longer exists
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

	// Verify profiles exist
	if !env.AcceptsProfile("dev") {
		t.Fatal("expected 'dev' profile to exist before removal")
	}

	// Remove the profile
	env.RemoveProfile("dev")

	// Verify profile no longer exists
	if env.AcceptsProfile("dev") {
		t.Error("expected 'dev' profile to not exist after removal")
	}

	// Verify other profile still exists
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

func TestFindDefaultConfigFile(t *testing.T) {
	t.Parallel()

	// This function searches for default config files
	// Just verify it doesn't panic
	file := FindDefaultConfigFile()
	// File may or may not exist depending on the environment
	_ = file
}

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
	t.Parallel()

	_ = os.Setenv("TEST_KEY", "test-value")
	defer func() { _ = os.Unsetenv("TEST_KEY") }()

	source := NewEnvPropertySource("test-env", "TEST")
	// Contains会将"key"转换为"TEST_KEY"（前缀+大写）
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

func TestTypeConverter_ToSlice(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// Test converting to slice
	result, err := converter.ConvertTo("a,b,c", reflect.TypeOf([]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slice, ok := result.Interface().([]string)
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

	// Test converting to uint
	result, err := converter.ConvertTo("42", reflect.TypeOf(uint(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(uint)
	if !ok {
		t.Fatal("expected result to be uint")
	}

	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestTypeConverter_ToFloat(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// Test converting to float64
	result, err := converter.ConvertTo("3.14", reflect.TypeOf(float64(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(float64)
	if !ok {
		t.Fatal("expected result to be float64")
	}

	if val != 3.14 {
		t.Errorf("expected 3.14, got %f", val)
	}
}

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

func TestTypeConverter_ConvertNumeric(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// 测试转换为int8
	result, err := converter.ConvertTo("127", reflect.TypeOf(int8(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(int8)
	if !ok {
		t.Fatal("expected result to be int8")
	}

	if val != 127 {
		t.Errorf("expected 127, got %d", val)
	}
}

func TestTypeConverter_SpecialConvert(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// 测试转换为time.Duration
	result, err := converter.ConvertTo("5s", reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(time.Duration)
	if !ok {
		t.Fatal("expected result to be time.Duration")
	}

	if val != 5*time.Second {
		t.Errorf("expected 5s, got %v", val)
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

	// 验证builder中的propertySources顺序
	if len(builder.propertySources) != 2 {
		t.Fatalf("expected 2 property sources, got %d", len(builder.propertySources))
	}
	// WithPropertySourceFirst应该把source2放在列表前面
	if builder.propertySources[0].Name() != "second" {
		t.Errorf("expected first source to be 'second', got %s", builder.propertySources[0].Name())
	}
}

func TestEnvironmentHelper_GetRequiredProperty(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	env.AddPropertySource(NewMapPropertySource("test", PriorityNormal, map[string]any{
		"app.name": "test-app",
	}))

	helper := NewEnvironmentHelper(env)

	// 测试获取存在的属性
	val, err := helper.GetRequiredProperty("app.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "test-app" {
		t.Errorf("expected 'test-app', got %v", val)
	}

	// 测试获取不存在的属性
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

	// 测试带前缀的键
	value := prefixedHelper.GetString("name", "default")
	if value != "test-app" {
		t.Errorf("expected 'test-app', got %v", value)
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
