package openapi

import (
	"reflect"
	"testing"
)

func TestSchemaObject_Struct(t *testing.T) {
	t.Parallel()
	min := 0.0
	max := 100.0
	minLen := 1
	maxLen := 100
	schema := SchemaObject{
		Type:                 "string",
		Format:               "uuid",
		Description:          "User ID",
		Properties:           map[string]SchemaObject{"name": {Type: "string"}},
		Required:             []string{"name"},
		Items:                &SchemaObject{Type: "string"},
		AdditionalProperties: &SchemaObject{Type: "string"},
		Enum:                 []string{"active", "inactive"},
		Default:              "default",
		Example:              "example",
		Minimum:              &min,
		Maximum:              &max,
		MinLength:            &minLen,
		MaxLength:            &maxLen,
		Pattern:              `^[a-z]+$`,
		Nullable:             true,
		ReadOnly:             true,
		WriteOnly:            true,
	}

	if schema.Type != "string" {
		t.Errorf("expected type 'string', got %s", schema.Type)
	}
	if !schema.Nullable {
		t.Error("expected nullable to be true")
	}
}

func TestComponentsObject_Struct(t *testing.T) {
	t.Parallel()
	comp := ComponentsObject{
		Schemas:         map[string]SchemaObject{"User": {Type: "object"}},
		Responses:       map[string]ResponseObject{"200": {Description: "OK"}},
		Parameters:      map[string]ParameterObject{"id": {Name: "id", In: "path"}},
		RequestBodies:   map[string]RequestBodyObject{"User": {Required: true}},
		SecuritySchemes: map[string]SecuritySchemeObject{"bearer": {Type: "http"}},
	}

	if len(comp.Schemas) != 1 {
		t.Errorf("expected 1 schema, got %d", len(comp.Schemas))
	}
	if len(comp.SecuritySchemes) != 1 {
		t.Errorf("expected 1 security scheme, got %d", len(comp.SecuritySchemes))
	}
}

func TestSecuritySchemeObject_Struct(t *testing.T) {
	t.Parallel()
	scheme := SecuritySchemeObject{
		Type:         "oauth2",
		Description:  "OAuth2 authentication",
		Name:         "Authorization",
		In:           "header",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Flows: &OAuthFlowsObject{
			AuthorizationCode: &OAuthFlowObject{
				AuthorizationURL: "https://auth.com/authorize",
				TokenURL:         "https://auth.com/token",
				Scopes:           map[string]string{"read": "Read access"},
			},
		},
	}

	if scheme.Type != "oauth2" {
		t.Errorf("expected type 'oauth2', got %s", scheme.Type)
	}
	if scheme.Flows == nil {
		t.Error("expected flows to be set")
	}
}

func TestOAuthFlowsObject_Struct(t *testing.T) {
	t.Parallel()
	flows := OAuthFlowsObject{
		Implicit:          &OAuthFlowObject{AuthorizationURL: "https://auth.com/implicit"},
		Password:          &OAuthFlowObject{TokenURL: "https://auth.com/password"},
		ClientCredentials: &OAuthFlowObject{TokenURL: "https://auth.com/client"},
		AuthorizationCode: &OAuthFlowObject{AuthorizationURL: "https://auth.com/code"},
	}

	if flows.Implicit == nil {
		t.Error("expected implicit flow to be set")
	}
	if flows.Password == nil {
		t.Error("expected password flow to be set")
	}
}

func TestTagObject_Struct(t *testing.T) {
	t.Parallel()
	tag := TagObject{
		Name:        "users",
		Description: "User operations",
	}

	if tag.Name != "users" {
		t.Errorf("expected name 'users', got %s", tag.Name)
	}
	if tag.Description != "User operations" {
		t.Errorf("expected description 'User operations', got %s", tag.Description)
	}
}

func TestDocumentBuilder_RegisterSchema_Schema(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.RegisterSchema("User", reflect.TypeOf(struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{}))

	if b.doc.Components == nil {
		t.Fatal("expected components to be initialized")
	}
	if _, ok := b.doc.Components.Schemas["User"]; !ok {
		t.Error("expected schema 'User' to be registered")
	}
}

func TestDocumentBuilder_RegisterSchema_Duplicate(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	type TestUser struct {
		Name string `json:"name"`
	}

	b.RegisterSchema("User", reflect.TypeOf(TestUser{}))
	b.RegisterSchema("User", reflect.TypeOf(TestUser{}))

	if len(b.doc.Components.Schemas) != 1 {
		t.Errorf("expected 1 schema (duplicate should be ignored), got %d", len(b.doc.Components.Schemas))
	}
}

func TestDocumentBuilder_GenerateSchema(t *testing.T) {
	t.Parallel()
	b := NewDocument()

	type Address struct {
		City string `json:"city"`
	}

	type TestUser struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Email   string  `json:"email,omitempty"`
		Address Address `json:"address"`
		private string
	}

	schema := b.generateSchema(reflect.TypeOf(TestUser{}))

	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %s", schema.Type)
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("expected property 'name'")
	}
	if _, ok := schema.Properties["age"]; !ok {
		t.Error("expected property 'age'")
	}
	if _, ok := schema.Properties["address"]; !ok {
		t.Error("expected property 'address'")
	}
	if _, ok := schema.Properties["private"]; ok {
		t.Error("expected private field to be excluded")
	}
	if _, ok := schema.Properties["email"]; !ok {
		t.Error("expected property 'email' to exist (omitempty only affects Required)")
	}
	// Verify email is not in Required list
	for _, req := range schema.Required {
		if req == "email" {
			t.Error("expected 'email' to not be in Required list due to omitempty")
		}
	}
}

func TestDocumentBuilder_GenerateFieldSchema(t *testing.T) {
	t.Parallel()
	b := NewDocument()

	tests := []struct {
		name         string
		value        any
		expectedType string
	}{
		{"bool", true, "boolean"},
		{"int", 42, "integer"},
		{"int8", int8(8), "integer"},
		{"int16", int16(16), "integer"},
		{"int32", int32(32), "integer"},
		{"int64", int64(64), "integer"},
		{"uint", uint(42), "integer"},
		{"uint8", uint8(8), "integer"},
		{"uint16", uint16(16), "integer"},
		{"uint32", uint32(32), "integer"},
		{"uint64", uint64(64), "integer"},
		{"float32", float32(3.14), "number"},
		{"float64", 3.14, "number"},
		{"string", "hello", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema := b.generateFieldSchema(reflect.TypeOf(tt.value))
			if schema.Type != tt.expectedType {
				t.Errorf("expected type '%s', got '%s'", tt.expectedType, schema.Type)
			}
		})
	}
}

func TestDocumentBuilder_GenerateFieldSchema_Slice(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	schema := b.generateFieldSchema(reflect.TypeOf([]string{}))
	if schema.Type != "array" {
		t.Errorf("expected type 'array', got '%s'", schema.Type)
	}
	if schema.Items == nil || schema.Items.Type != "string" {
		t.Error("expected items type to be 'string'")
	}
}

func TestDocumentBuilder_GenerateFieldSchema_Map(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	schema := b.generateFieldSchema(reflect.TypeOf(map[string]int{}))
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", schema.Type)
	}
	if schema.AdditionalProperties == nil || schema.AdditionalProperties.Type != "integer" {
		t.Error("expected additionalProperties type to be 'integer'")
	}
}

func TestDocumentBuilder_GenerateFieldSchema_Struct(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	type TestStruct struct {
		Name string
	}
	schema := b.generateFieldSchema(reflect.TypeOf(TestStruct{}))
	if schema.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", schema.Type)
	}
}

func TestDocumentBuilder_MapTypeToString(t *testing.T) {
	t.Parallel()
	b := NewDocument()

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"bool", true, "boolean"},
		{"int", 42, "integer"},
		{"int64", int64(64), "integer"},
		{"uint", uint(42), "integer"},
		{"float32", float32(3.14), "number"},
		{"float64", 3.14, "number"},
		{"string", "hello", "string"},
		{"struct", struct{}{}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := b.mapTypeToString(reflect.TypeOf(tt.value))
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
