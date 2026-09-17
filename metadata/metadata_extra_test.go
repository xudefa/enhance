package metadata

import (
	"reflect"
	"testing"
)

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
		groupName := gen.extractGroupName(tt.structName)
		if groupName != tt.expected {
			t.Errorf("extractGroupName(%q) = %q, expected %q", tt.structName, groupName, tt.expected)
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
		javaType := gen.mapType(tt.goType)
		if javaType != tt.expected {
			t.Errorf("mapType(%v) = %q, expected %q", tt.goType, javaType, tt.expected)
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
		parts := splitAttributes(tt.input)
		if len(parts) != len(tt.expected) {
			t.Errorf("splitAttributes(%q) = %v, expected %v", tt.input, parts, tt.expected)
			continue
		}
		for i, v := range parts {
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

	value, ok := GetStringAttribute(ann, "key1")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got %s", value)
	}

	value, ok = GetStringAttribute(ann, "key2")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if value != "123" {
		t.Errorf("expected '123', got %s", value)
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

	value, ok := GetIntAttribute(ann, "intKey")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if value != 42 {
		t.Errorf("expected 42, got %d", value)
	}

	value, ok = GetIntAttribute(ann, "strKey")
	if ok {
		t.Error("expected string to not convert to int")
	}

	value, ok = GetIntAttribute(ann, "floatKey")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if value != 3 {
		t.Errorf("expected 3, got %d", value)
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

	value, ok := GetBoolAttribute(ann, "boolKey")
	if !ok {
		t.Error("expected attribute to exist")
	}
	if !value {
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
