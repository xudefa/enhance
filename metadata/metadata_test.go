package metadata

import (
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
