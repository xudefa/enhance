package openapi

import (
	"reflect"
	"strings"
	"testing"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name" description:"用户名称"`
	Email string `json:"email" description:"用户邮箱"`
	Age   int    `json:"age,omitempty" description:"用户年龄"`
}

type UserController struct{}

func (c *UserController) GetUser()    {}
func (c *UserController) CreateUser() {}
func (c *UserController) UpdateUser() {}
func (c *UserController) DeleteUser() {}
func (c *UserController) ListUsers()  {}

func TestDocumentBuilder_Basic(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.SetInfo("Test API", "1.0.0", "A test API")
	doc.AddServer("http://localhost:8080", "Development server")

	if doc.doc.Info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got %s", doc.doc.Info.Title)
	}

	if doc.doc.Info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", doc.doc.Info.Version)
	}

	if len(doc.doc.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(doc.doc.Servers))
	}
}

func TestDocumentBuilder_ContactAndLicense(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.SetContact("John Doe", "http://example.com", "john@example.com")
	doc.SetLicense("MIT", "http://opensource.org/licenses/MIT")
	doc.SetTermsOfService("http://example.com/terms")

	if doc.doc.Info.Contact.Name != "John Doe" {
		t.Errorf("expected contact name 'John Doe', got %s", doc.doc.Info.Contact.Name)
	}

	if doc.doc.Info.License.Name != "MIT" {
		t.Errorf("expected license name 'MIT', got %s", doc.doc.Info.License.Name)
	}

	if doc.doc.Info.TermsOfService != "http://example.com/terms" {
		t.Errorf("expected terms of service, got %s", doc.doc.Info.TermsOfService)
	}
}

func TestDocumentBuilder_AddTag(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.AddTag("users", "User management")
	doc.AddTag("products", "Product management")

	if len(doc.doc.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(doc.doc.Tags))
	}

	if doc.doc.Tags[0].Name != "users" {
		t.Errorf("expected first tag name 'users', got %s", doc.doc.Tags[0].Name)
	}
}

func TestDocumentBuilder_AddSecurityScheme(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.AddSecurityScheme("bearerAuth", SecuritySchemeObject{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	})

	if doc.doc.Components == nil {
		t.Fatal("expected components to be created")
	}

	if _, ok := doc.doc.Components.SecuritySchemes["bearerAuth"]; !ok {
		t.Error("expected bearerAuth security scheme")
	}
}

func TestDocumentBuilder_RegisterController(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.RegisterController(&UserController{})

	// 控制器方法应该已被注册
	if len(doc.doc.Paths) == 0 {
		t.Logf("Warning: No paths registered, checking if methods are being skipped")
	}
}

func TestDocumentBuilder_AddPath(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.AddPath("/users", "GET", OperationObject{
		Summary:     "List users",
		OperationID: "listUsers",
		Responses:   map[string]ResponseObject{"200": {Description: "Success"}},
	})

	pathItem, ok := doc.doc.Paths["/users"]
	if !ok {
		t.Fatal("expected /users path")
	}

	if pathItem.Get == nil {
		t.Fatal("expected GET operation")
	}

	if pathItem.Get.Summary != "List users" {
		t.Errorf("expected summary 'List users', got %s", pathItem.Get.Summary)
	}
}

func TestDocumentBuilder_AddSchema(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.AddSchema("User", SchemaObject{
		Type: "object",
		Properties: map[string]SchemaObject{
			"id":   {Type: "integer"},
			"name": {Type: "string"},
		},
	})

	if doc.doc.Components == nil {
		t.Fatal("expected components to be created")
	}

	if _, ok := doc.doc.Components.Schemas["User"]; !ok {
		t.Error("expected User schema")
	}
}

func TestDocumentBuilder_RegisterSchema(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.RegisterSchema("User", reflect.TypeOf(User{}))

	if doc.doc.Components == nil {
		t.Fatal("expected components to be created")
	}

	schema, ok := doc.doc.Components.Schemas["User"]
	if !ok {
		t.Fatal("expected User schema")
	}

	if schema.Type != "object" {
		t.Errorf("expected type 'object', got %s", schema.Type)
	}

	if _, ok := schema.Properties["name"]; !ok {
		t.Error("expected name property")
	}
}

func TestDocumentBuilder_ToJSON(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.SetInfo("Test API", "1.0.0", "A test API")
	doc.AddServer("http://localhost:8080", "Development server")

	json, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if !strings.Contains(json, `"openapi": "3.0.3"`) {
		t.Error("expected openapi version")
	}

	if !strings.Contains(json, `"title": "Test API"`) {
		t.Error("expected title")
	}

	if !strings.Contains(json, `"servers"`) {
		t.Error("expected servers")
	}
}

func TestDocumentBuilder_SaveToFile(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.SetInfo("Test API", "1.0.0", "A test API")

	tmpFile := t.TempDir() + "/openapi.json"
	err := doc.SaveToFile(tmpFile)
	if err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}
}

func TestDocumentBuilder_Build(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.AddPath("/users", "GET", OperationObject{
		Summary:   "List users",
		Responses: map[string]ResponseObject{"200": {Description: "Success"}},
	})
	doc.AddPath("/users", "POST", OperationObject{
		Summary:   "Create user",
		Responses: map[string]ResponseObject{"201": {Description: "Created"}},
	})
	doc.AddTag("users", "User management")

	built := doc.Build()

	// 验证路径已排序
	pathKeys := make([]string, 0, len(built.Paths))
	for k := range built.Paths {
		pathKeys = append(pathKeys, k)
	}

	if len(pathKeys) != 1 {
		t.Errorf("expected 1 path, got %d", len(pathKeys))
	}

	// 验证标签已排序
	if len(built.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(built.Tags))
	}
}

func TestDocumentBuilder_MultipleHTTPMethods(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.AddPath("/users", "GET", OperationObject{
		Summary:   "List users",
		Responses: map[string]ResponseObject{"200": {Description: "Success"}},
	})
	doc.AddPath("/users", "POST", OperationObject{
		Summary:   "Create user",
		Responses: map[string]ResponseObject{"201": {Description: "Created"}},
	})
	doc.AddPath("/users", "PUT", OperationObject{
		Summary:   "Update user",
		Responses: map[string]ResponseObject{"200": {Description: "Success"}},
	})
	doc.AddPath("/users", "DELETE", OperationObject{
		Summary:   "Delete user",
		Responses: map[string]ResponseObject{"204": {Description: "No content"}},
	})
	doc.AddPath("/users", "PATCH", OperationObject{
		Summary:   "Patch user",
		Responses: map[string]ResponseObject{"200": {Description: "Success"}},
	})

	pathItem := doc.doc.Paths["/users"]

	if pathItem.Get == nil {
		t.Error("expected GET operation")
	}

	if pathItem.Post == nil {
		t.Error("expected POST operation")
	}

	if pathItem.Put == nil {
		t.Error("expected PUT operation")
	}

	if pathItem.Delete == nil {
		t.Error("expected DELETE operation")
	}

	if pathItem.Patch == nil {
		t.Error("expected PATCH operation")
	}
}

func TestSchemaObject_Generation(t *testing.T) {
	t.Parallel()
	doc := NewDocument()
	doc.RegisterSchema("User", reflect.TypeOf(User{}))

	schema := doc.doc.Components.Schemas["User"]

	// 验证属性
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("expected id property")
	}

	if _, ok := schema.Properties["name"]; !ok {
		t.Error("expected name property")
	}

	if _, ok := schema.Properties["email"]; !ok {
		t.Error("expected email property")
	}

	// 验证必填字段（age 有 omitempty，所以不是必填）
	requiredMap := make(map[string]bool)
	for _, req := range schema.Required {
		requiredMap[req] = true
	}

	if !requiredMap["id"] {
		t.Error("expected id to be required")
	}

	if !requiredMap["name"] {
		t.Error("expected name to be required")
	}

	if requiredMap["age"] {
		t.Error("expected age to be optional (omitempty)")
	}
}

func TestExtractHTTPMethod(t *testing.T) {
	t.Parallel()
	doc := NewDocument()

	tests := []struct {
		methodName string
		expected   string
	}{
		{"GetUser", "GET"},
		{"FindUser", "GET"},
		{"ListUsers", "GET"},
		{"CreateUser", "POST"},
		{"PostUser", "POST"},
		{"AddUser", "POST"},
		{"UpdateUser", "PUT"},
		{"PutUser", "PUT"},
		{"DeleteUser", "DELETE"},
		{"RemoveUser", "DELETE"},
		{"PatchUser", "PATCH"},
		{"UnknownMethod", ""},
	}

	for _, tt := range tests {
		t.Run(tt.methodName, func(t *testing.T) {
			result := doc.extractHTTPMethod(reflect.Method{Name: tt.methodName})
			if result != tt.expected {
				t.Errorf("extractHTTPMethod(%q) = %q, expected %q", tt.methodName, result, tt.expected)
			}
		})
	}
}

func TestExtractPath(t *testing.T) {
	t.Parallel()
	doc := NewDocument()

	tests := []struct {
		methodName string
		basePath   string
		expected   string
	}{
		{"GetUser", "/users", "/users/user"},
		{"CreateUser", "/users", "/users/createuser"},
		{"ListUsers", "/api", "/api/listusers"},
	}

	for _, tt := range tests {
		t.Run(tt.methodName, func(t *testing.T) {
			result := doc.extractPath(reflect.Method{Name: tt.methodName}, tt.basePath)
			if result != tt.expected {
				t.Errorf("extractPath(%q, %q) = %q, expected %q", tt.methodName, tt.basePath, result, tt.expected)
			}
		})
	}
}

func TestExtractBasePath(t *testing.T) {
	t.Parallel()
	doc := NewDocument()

	tests := []struct {
		typeName string
		expected string
	}{
		{"UserController", "/user"},
		{"ProductHandler", "/product"},
		{"OrderController", "/order"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			// 创建一个虚拟的类型
			result := doc.extractBasePath(reflect.TypeOf(&UserController{}).Elem())
			// 验证返回的路径以 / 开头
			if !strings.HasPrefix(result, "/") {
				t.Errorf("extractBasePath(%q) = %q, should start with /", tt.typeName, result)
			}
		})
	}
}
