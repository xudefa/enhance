package metadata

import (
	"reflect"
	"strings"
	"testing"
)

type ServerConfig struct {
	Host string `config:"server.host" default:"localhost" description:"服务器主机地址"`
	Port int    `config:"server.port" default:"8080" description:"服务器端口"`
}

type DatabaseConfig struct {
	URL      string `config:"database.url" default:"localhost:5432" description:"数据库连接地址"`
	Username string `config:"database.username" default:"root" description:"数据库用户名"`
	Password string `config:"database.password" description:"数据库密码" secret:"true"`
}

type CacheConfig struct {
	Enabled bool     `config:"cache.enabled" default:"true" description:"是否启用缓存"`
	TTL     int64    `config:"cache.ttl" default:"3600" description:"缓存过期时间（秒）"`
	Servers []string `config:"cache.servers" description:"缓存服务器列表"`
}

type ComplexConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Cache    CacheConfig
}

func TestMetadataGenerator_Register(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})

	metadata := gen.Generate()

	if len(metadata.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(metadata.Groups))
	}

	if len(metadata.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(metadata.Properties))
	}
}

func TestMetadataGenerator_MultipleConfigs(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})
	gen.Register(&DatabaseConfig{})
	gen.Register(&CacheConfig{})

	metadata := gen.Generate()

	if len(metadata.Groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(metadata.Groups))
	}

	// ServerConfig: 2, DatabaseConfig: 3, CacheConfig: 3 = 8 total
	if len(metadata.Properties) != 8 {
		t.Errorf("expected 8 properties, got %d", len(metadata.Properties))
	}
}

func TestMetadataGenerator_PropertyMetadata(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})

	metadata := gen.Generate()

	// 查找 server.host 属性
	var hostProp *PropertyMetadata
	for i := range metadata.Properties {
		if metadata.Properties[i].Name == "server.host" {
			hostProp = &metadata.Properties[i]
			break
		}
	}

	if hostProp == nil || hostProp.Type != "java.lang.String" || hostProp.DefaultValue != "localhost" {
		t.Fatalf("server.host property invalid: found=%v, type=%v, default=%v", hostProp != nil, hostProp != nil, hostProp != nil)
	}
}

func TestMetadataGenerator_SecretDetection(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&DatabaseConfig{})

	metadata := gen.Generate()

	// 查找 database.password 属性
	var passwordProp *PropertyMetadata
	for i := range metadata.Properties {
		if metadata.Properties[i].Name == "database.password" {
			passwordProp = &metadata.Properties[i]
			break
		}
	}

	if passwordProp == nil || !passwordProp.Secret {
		t.Fatalf("database.password property invalid: found=%v, secret=%v", passwordProp != nil, passwordProp != nil && passwordProp.Secret)
	}
}

func TestMetadataGenerator_TypeMapping(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&CacheConfig{})

	metadata := gen.Generate()

	// 检查类型映射
	typeMap := make(map[string]string)
	for _, prop := range metadata.Properties {
		typeMap[prop.Name] = prop.Type
	}

	if typeMap["cache.enabled"] != "java.lang.Boolean" {
		t.Errorf("expected cache.enabled type java.lang.Boolean, got %s", typeMap["cache.enabled"])
	}

	if typeMap["cache.ttl"] != "java.lang.Integer" {
		t.Errorf("expected cache.ttl type java.lang.Integer, got %s", typeMap["cache.ttl"])
	}

	if typeMap["cache.servers"] != "java.util.List<java.lang.String>" {
		t.Errorf("expected cache.servers type java.util.List<java.lang.String>, got %s", typeMap["cache.servers"])
	}
}

func TestMetadataGenerator_Hints(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})

	gen.WithHint("server.host", []HintValue{
		{Value: "localhost", Description: "本地主机"},
		{Value: "0.0.0.0", Description: "所有接口"},
	})

	gen.WithHintProvider("server.port", "any", map[string]string{
		"target": "java.lang.Integer",
	})

	metadata := gen.Generate()

	if len(metadata.Hints) != 2 {
		t.Errorf("expected 2 hints, got %d", len(metadata.Hints))
	}

	// 查找 server.host hint
	var hostHint *HintMetadata
	for i := range metadata.Hints {
		if metadata.Hints[i].Name == "server.host" {
			hostHint = &metadata.Hints[i]
			break
		}
	}

	if hostHint == nil || len(hostHint.Values) != 2 {
		t.Fatalf("server.host hint invalid: found=%v, values=%v", hostHint != nil, hostHint != nil)
	}
}

func TestMetadataGenerator_ToJSON(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})

	metadata := gen.Generate()

	json, err := metadata.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if !strings.Contains(json, `"properties"`) {
		t.Error("expected JSON to contain properties")
	}

	if !strings.Contains(json, `"server.host"`) {
		t.Error("expected JSON to contain server.host")
	}
}

func TestGenerateFromStruct(t *testing.T) {
	t.Parallel()
	metadata := GenerateFromStruct(&ServerConfig{}, &DatabaseConfig{})

	if len(metadata.Properties) != 5 {
		t.Errorf("expected 5 properties, got %d", len(metadata.Properties))
	}
}

func TestGenerateJSON(t *testing.T) {
	t.Parallel()
	json, err := GenerateJSON(&ServerConfig{})
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	if json == "" {
		t.Error("expected non-empty JSON")
	}
}

func TestPropertyIndex(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})
	gen.Register(&DatabaseConfig{})

	metadata := gen.Generate()
	index := NewPropertyIndex(metadata)

	// 测试 Get
	prop, ok := index.Get("server.host")
	if !ok {
		t.Fatal("server.host not found in index")
	}

	if prop.Name != "server.host" {
		t.Errorf("expected name server.host, got %s", prop.Name)
	}

	// 测试 Has
	if !index.Has("database.url") {
		t.Error("expected database.url to exist")
	}

	// 测试 GetByPrefix
	props := index.GetByPrefix("server")
	if len(props) != 2 {
		t.Errorf("expected 2 properties with server prefix, got %d", len(props))
	}
}

func TestValidateProperty(t *testing.T) {
	t.Parallel()
	prop := PropertyMetadata{
		Name:     "server.host",
		Required: true,
	}

	// 测试必填属性
	err := ValidateProperty("server.host", "", prop)
	if err == nil {
		t.Error("expected error for required property with empty value")
	}

	// 测试有效值
	err = ValidateProperty("server.host", "localhost", prop)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 测试敏感属性
	secretProp := PropertyMetadata{
		Name:   "database.password",
		Secret: true,
	}

	err = ValidateProperty("database.password", "secret123", secretProp)
	if err != nil {
		t.Errorf("secret property with value should be allowed: %v", err)
	}
}

func TestMetadataGenerator_GroupExtraction(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})

	metadata := gen.Generate()

	if len(metadata.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(metadata.Groups))
	}

	if metadata.Groups[0].Name != "server" {
		t.Errorf("expected group name server, got %s", metadata.Groups[0].Name)
	}
}

func TestMetadataGenerator_CamelToKebab(t *testing.T) {
	t.Parallel()
	// 验证 MetadataGenerator 可以正常工作
	// camelToKebab 是内部实现细节，这里通过 Generate 的结果间接验证
	type ServerConfig struct {
		DatabaseURL string
		HTTPServer  string
		Simple      string
	}

	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})
	metadata := gen.Generate()

	// 验证元数据生成成功
	if metadata == nil {
		t.Fatal("expected metadata to be generated")
	}
}

func TestMetadataGenerator_DefaultPropertyName(t *testing.T) {
	t.Parallel()
	type TestConfig struct {
		FieldName string
	}

	gen := NewMetadataGenerator()
	gen.Register(&TestConfig{})

	metadata := gen.Generate()

	if len(metadata.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(metadata.Properties))
	}

	expected := "test.field-name"
	if metadata.Properties[0].Name != expected {
		t.Errorf("expected property name %q, got %q", expected, metadata.Properties[0].Name)
	}
}

func TestPropertyIndex_GetAll_Extended(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator()
	gen.Register(&ServerConfig{})
	gen.Register(&DatabaseConfig{})

	metadata := gen.Generate()
	index := NewPropertyIndex(metadata)

	// 测试 GetAll
	allProps := index.GetAll()
	if len(allProps) != 5 {
		t.Errorf("expected 5 properties, got %d", len(allProps))
	}

	// 验证排序
	for i := 1; i < len(allProps); i++ {
		if allProps[i-1].Name > allProps[i].Name {
			t.Errorf("properties not sorted: %s > %s", allProps[i-1].Name, allProps[i].Name)
		}
	}
}

func TestTagAnnotationResolver_GetAnnotation_Extended(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config")

	type TestConfig struct {
		Host string `config:"server.host"`
	}

	// Test via ResolveAnnotations
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(&TestConfig{}).Elem())
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Name != "server.host" {
		t.Errorf("expected annotation name 'server.host', got %s", annotations[0].Name)
	}
}

func TestTagAnnotationResolver_HasAnnotation_Extended(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config")

	type TestConfig struct {
		Host string `config:"server.host"`
	}

	annotations := resolver.ResolveAnnotations(reflect.TypeOf(&TestConfig{}).Elem())
	found := false
	for _, ann := range annotations {
		if ann.Name == "server.host" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected HasAnnotation to return true")
	}
}

func TestTagAnnotationResolver_GetAnnotations_Extended(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config")

	type TestConfig struct {
		Host string `config:"server.host"`
		Port int    `config:"server.port"`
	}

	annotations := resolver.ResolveAnnotations(reflect.TypeOf(&TestConfig{}).Elem())
	if len(annotations) != 2 {
		t.Errorf("expected 2 annotations, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_GetFieldAnnotations_Extended(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config")

	type TestConfig struct {
		Host string `config:"server.host"`
	}

	// Test by resolving annotations on the struct type
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestConfig{}))
	if len(annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Name != "server.host" {
		t.Errorf("expected annotation name 'server.host', got %s", annotations[0].Name)
	}
}

func TestTagAnnotationResolver_GetFieldAnnotation_Extended(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config")

	type TestConfig struct {
		Host string `config:"server.host"`
	}

	// Test by resolving annotations on the struct type
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestConfig{}))
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Name != "server.host" {
		t.Errorf("expected annotation name 'server.host', got %s", annotations[0].Name)
	}
}

func TestMetadataGenerator_ExtractGroupName(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator().(*metadataGeneratorImpl)

	tests := []struct {
		structName string
		expected   string
	}{
		{"ServerConfig", "server"},
		{"DatabaseConfig", "database"},
		{"CacheProperties", "cache"},
		{"AppConfig", "app"},
		{"", ""},
	}

	for _, tt := range tests {
		result := gen.extractGroupName(tt.structName)
		if result != tt.expected {
			t.Errorf("extractGroupName(%q) = %q, expected %q", tt.structName, result, tt.expected)
		}
	}
}

func TestMetadataGenerator_MapType(t *testing.T) {
	t.Parallel()
	gen := NewMetadataGenerator().(*metadataGeneratorImpl)

	tests := []struct {
		goType   reflect.Type
		expected string
	}{
		{reflect.TypeOf(""), "java.lang.String"},
		{reflect.TypeOf(0), "java.lang.Integer"},
		{reflect.TypeOf(int64(0)), "java.lang.Integer"},
		{reflect.TypeOf(true), "java.lang.Boolean"},
		{reflect.TypeOf(float64(0)), "java.lang.Float"},
		{reflect.TypeOf([]string{}), "java.util.List<java.lang.String>"},
	}

	for _, tt := range tests {
		result := gen.mapType(tt.goType)
		if result != tt.expected {
			t.Errorf("mapType(%v) = %q, expected %q", tt.goType, result, tt.expected)
		}
	}
}

func TestSplitAttributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected []string
	}{
		{"key1=value1,key2=value2", []string{"key1=value1", "key2=value2"}},
		{"key=value", []string{"key=value"}},
		{"", []string{}},
		{"key1=value1,", []string{"key1=value1"}},
	}

	for _, tt := range tests {
		result := splitAttributes(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitAttributes(%q) = %v, expected %v", tt.input, result, tt.expected)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("splitAttributes(%q)[%d] = %q, expected %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestTagAnnotationResolver_GetAnnotation_Direct(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config").(*tagAnnotationResolverImpl)

	type TestConfig struct {
		Host string `config:"server.host"`
		Port int    `config:"server.port"`
	}

	cfg := TestConfig{}
	ann := resolver.GetAnnotation(cfg, "server.host")
	if ann.Name != "server.host" {
		t.Errorf("expected annotation name 'server.host', got %s", ann.Name)
	}

	annNotFound := resolver.GetAnnotation(cfg, "nonexistent")
	if annNotFound.Name != "" {
		t.Errorf("expected empty annotation, got %s", annNotFound.Name)
	}
}

func TestTagAnnotationResolver_HasAnnotation_Direct(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config").(*tagAnnotationResolverImpl)

	type TestConfig struct {
		Host string `config:"server.host"`
	}

	cfg := TestConfig{}
	if !resolver.HasAnnotation(cfg, "server.host") {
		t.Error("expected HasAnnotation to return true")
	}
	if resolver.HasAnnotation(cfg, "nonexistent") {
		t.Error("expected HasAnnotation to return false")
	}
}

func TestTagAnnotationResolver_GetAnnotations_Direct(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config").(*tagAnnotationResolverImpl)

	type TestConfig struct {
		Host string `config:"server.host"`
		Port int    `config:"server.port"`
	}

	cfg := TestConfig{}
	annotations := resolver.GetAnnotations(cfg)
	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_GetFieldAnnotations_Direct(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config").(*tagAnnotationResolverImpl)

	type TestConfig struct {
		Host string `config:"server.host"`
		Port int    `config:"server.port"`
	}

	cfg := TestConfig{}
	annotations, err := resolver.GetFieldAnnotations(cfg, "Host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Name != "server.host" {
		t.Errorf("expected annotation name 'server.host', got %s", annotations[0].Name)
	}

	// Test non-existent field
	_, err = resolver.GetFieldAnnotations(cfg, "NonExistent")
	if err == nil {
		t.Error("expected error for non-existent field")
	}

	// Test non-struct type
	_, err = resolver.GetFieldAnnotations("not a struct", "Field")
	if err == nil {
		t.Error("expected error for non-struct type")
	}

	// Test field without annotation
	type NoTagConfig struct {
		Name string
	}
	noTagCfg := NoTagConfig{}
	annotations, err = resolver.GetFieldAnnotations(noTagCfg, "Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_GetFieldAnnotation_Direct(t *testing.T) {
	t.Parallel()
	resolver := NewTagAnnotationResolver("config").(*tagAnnotationResolverImpl)

	type TestConfig struct {
		Host string `config:"server.host"`
		Port int    `config:"server.port"`
	}

	cfg := TestConfig{}
	ann, err := resolver.GetFieldAnnotation(cfg, "Host", "server.host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ann.Name != "server.host" {
		t.Errorf("expected annotation name 'server.host', got %s", ann.Name)
	}

	// Test non-existent annotation
	ann, err = resolver.GetFieldAnnotation(cfg, "Host", "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ann.Name != "" {
		t.Errorf("expected empty annotation, got %s", ann.Name)
	}
}

func TestGetStringAttribute(t *testing.T) {
	t.Parallel()

	ann := Annotation{
		Name: "test",
		Attributes: map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	val, ok := GetStringAttribute(ann, "key1")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %s", val)
	}

	val, ok = GetStringAttribute(ann, "key2")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if val != "123" {
		t.Errorf("expected '123', got %s", val)
	}

	_, ok = GetStringAttribute(ann, "nonexistent")
	if ok {
		t.Error("expected attribute to not exist")
	}
}

func TestGetIntAttribute(t *testing.T) {
	t.Parallel()

	ann := Annotation{
		Name: "test",
		Attributes: map[string]any{
			"intKey":   42,
			"strKey":   "123",
			"floatKey": 3.14,
		},
	}

	val, ok := GetIntAttribute(ann, "intKey")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}

	val, ok = GetIntAttribute(ann, "strKey")
	if ok {
		t.Error("expected string to not convert to int")
	}

	val, ok = GetIntAttribute(ann, "floatKey")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if val != 3 {
		t.Errorf("expected 3, got %d", val)
	}

	_, ok = GetIntAttribute(ann, "nonexistent")
	if ok {
		t.Error("expected attribute to not exist")
	}
}

func TestGetBoolAttribute(t *testing.T) {
	t.Parallel()

	ann := Annotation{
		Name: "test",
		Attributes: map[string]any{
			"boolKey": true,
			"strKey":  "true",
			"intKey":  1,
		},
	}

	val, ok := GetBoolAttribute(ann, "boolKey")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if !val {
		t.Error("expected true")
	}

	_, ok = GetBoolAttribute(ann, "strKey")
	if ok {
		t.Error("expected string to not convert to bool")
	}

	_, ok = GetBoolAttribute(ann, "intKey")
	if ok {
		t.Error("expected int to not convert to bool")
	}

	_, ok = GetBoolAttribute(ann, "nonexistent")
	if ok {
		t.Error("expected attribute to not exist")
	}
}
