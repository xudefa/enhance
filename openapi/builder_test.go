package openapi

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewDocument(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	if b == nil {
		t.Fatal("expected non-nil DocumentBuilder")
	}
	if b.doc.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI 3.0.3, got %s", b.doc.OpenAPI)
	}
}

func TestDocumentBuilder_SetInfo(t *testing.T) {
	t.Parallel()
	b := NewDocument().SetInfo("Test API", "2.0.0", "Test description")
	if b.doc.Info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got %s", b.doc.Info.Title)
	}
	if b.doc.Info.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %s", b.doc.Info.Version)
	}
	if b.doc.Info.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %s", b.doc.Info.Description)
	}
}

func TestDocumentBuilder_SetContact(t *testing.T) {
	t.Parallel()
	b := NewDocument().SetContact("Test", "http://test.com", "test@test.com")
	if b.doc.Info.Contact.Name != "Test" {
		t.Errorf("expected contact name 'Test', got %s", b.doc.Info.Contact.Name)
	}
	if b.doc.Info.Contact.URL != "http://test.com" {
		t.Errorf("expected contact URL 'http://test.com', got %s", b.doc.Info.Contact.URL)
	}
	if b.doc.Info.Contact.Email != "test@test.com" {
		t.Errorf("expected contact email 'test@test.com', got %s", b.doc.Info.Contact.Email)
	}
}

func TestDocumentBuilder_SetLicense(t *testing.T) {
	t.Parallel()
	b := NewDocument().SetLicense("Apache 2.0", "http://apache.org")
	if b.doc.Info.License.Name != "Apache 2.0" {
		t.Errorf("expected license name 'Apache 2.0', got %s", b.doc.Info.License.Name)
	}
	if b.doc.Info.License.URL != "http://apache.org" {
		t.Errorf("expected license URL 'http://apache.org', got %s", b.doc.Info.License.URL)
	}
}

func TestDocumentBuilder_SetTermsOfService(t *testing.T) {
	t.Parallel()
	b := NewDocument().SetTermsOfService("http://terms.com")
	if b.doc.Info.TermsOfService != "http://terms.com" {
		t.Errorf("expected terms 'http://terms.com', got %s", b.doc.Info.TermsOfService)
	}
}

func TestDocumentBuilder_AddServer(t *testing.T) {
	t.Parallel()
	b := NewDocument().AddServer("http://api.com", "Production")
	if len(b.doc.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(b.doc.Servers))
	}
	if b.doc.Servers[0].URL != "http://api.com" {
		t.Errorf("expected server URL 'http://api.com', got %s", b.doc.Servers[0].URL)
	}
	if b.doc.Servers[0].Description != "Production" {
		t.Errorf("expected server description 'Production', got %s", b.doc.Servers[0].Description)
	}
}

func TestDocumentBuilder_AddTag_Builder(t *testing.T) {
	t.Parallel()
	b := NewDocument().AddTag("users", "User operations")
	if len(b.doc.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(b.doc.Tags))
	}
	if b.doc.Tags[0].Name != "users" {
		t.Errorf("expected tag name 'users', got %s", b.doc.Tags[0].Name)
	}
	if b.doc.Tags[0].Description != "User operations" {
		t.Errorf("expected tag description 'User operations', got %s", b.doc.Tags[0].Description)
	}
}

func TestDocumentBuilder_AddSecurityScheme_Builder(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddSecurityScheme("bearer", SecuritySchemeObject{
		Type:   "http",
		Scheme: "bearer",
	})

	if b.doc.Components == nil {
		t.Fatal("expected components to be initialized")
	}
	if _, ok := b.doc.Components.SecuritySchemes["bearer"]; !ok {
		t.Error("expected security scheme to be added")
	}
}

func TestDocumentBuilder_AddSchema_Builder(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddSchema("User", SchemaObject{Type: "object"})

	if b.doc.Components == nil {
		t.Fatal("expected components to be initialized")
	}
	if _, ok := b.doc.Components.Schemas["User"]; !ok {
		t.Error("expected schema to be added")
	}
}

func TestDocumentBuilder_AddPath_Builder(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddPath("/users", "GET", OperationObject{
		Summary: "Get users",
	})

	if len(b.doc.Paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(b.doc.Paths))
	}
	if b.doc.Paths["/users"].Get == nil {
		t.Error("expected GET operation to be set")
	}
}

func TestDocumentBuilder_AddPath_POST(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddPath("/users", "POST", OperationObject{Summary: "Create user"})
	if b.doc.Paths["/users"].Post == nil {
		t.Error("expected POST operation to be set")
	}
}

func TestDocumentBuilder_AddPath_PUT(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddPath("/users", "PUT", OperationObject{Summary: "Update user"})
	if b.doc.Paths["/users"].Put == nil {
		t.Error("expected PUT operation to be set")
	}
}

func TestDocumentBuilder_AddPath_DELETE(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddPath("/users", "DELETE", OperationObject{Summary: "Delete user"})
	if b.doc.Paths["/users"].Delete == nil {
		t.Error("expected DELETE operation to be set")
	}
}

func TestDocumentBuilder_AddPath_PATCH(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddPath("/users", "PATCH", OperationObject{Summary: "Patch user"})
	if b.doc.Paths["/users"].Patch == nil {
		t.Error("expected PATCH operation to be set")
	}
}

func TestDocumentBuilder_RegisterController_Nil(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.RegisterController(nil)
	if len(b.doc.Paths) != 0 {
		t.Error("expected no paths to be registered for nil controller")
	}
}

type TestUserController struct{}

func (TestUserController) GetUsers()   {}
func (TestUserController) CreateUser() {}

func TestDocumentBuilder_RegisterController_Builder(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.RegisterController(&TestUserController{})

	if len(b.doc.Paths) == 0 {
		t.Error("expected paths to be registered")
	}
}

func TestDocumentBuilder_Build_Sorted(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	b.AddPath("/b", "GET", OperationObject{Summary: "B"})
	b.AddPath("/a", "GET", OperationObject{Summary: "A"})

	doc := b.Build()

	// Build方法应该对Paths进行排序
	// 验证路径存在
	if _, ok := doc.Paths["/a"]; !ok {
		t.Error("expected path /a to exist")
	}
	if _, ok := doc.Paths["/b"]; !ok {
		t.Error("expected path /b to exist")
	}

	// 验证路径数量
	if len(doc.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(doc.Paths))
	}
}

func TestDocumentBuilder_ToJSON_Builder(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	json, err := b.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if json == "" {
		t.Error("expected non-empty JSON")
	}
}

func TestDocumentBuilder_ToJSONBytes(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	data, err := b.ToJSONBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON bytes")
	}
}

func TestDocumentBuilder_SaveToFile_Builder(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	err := b.SaveToFile("/tmp/test-openapi.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentBuilder_ServeHTTP(t *testing.T) {
	t.Parallel()
	b := NewDocument()
	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rec := httptest.NewRecorder()

	b.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestDocumentBuilder_extractBasePath(t *testing.T) {
	t.Parallel()
	b := NewDocument()

	tests := []struct {
		name     string
		typeName string
		expected string
	}{
		{"UserController", "UserController", "/user"},
		{"UserHandler", "UserHandler", "/user"},
		{"User", "User", "/user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := b.extractBasePath(stringToType(tt.typeName))
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDocumentBuilder_extractHTTPMethod(t *testing.T) {
	t.Parallel()
	b := NewDocument()

	tests := []struct {
		name     string
		method   string
		expected string
	}{
		{"GetUsers", "GetUsers", "GET"},
		{"FindUser", "FindUser", "GET"},
		{"ListUsers", "ListUsers", "GET"},
		{"PostUser", "PostUser", "POST"},
		{"CreateUser", "CreateUser", "POST"},
		{"AddUser", "AddUser", "POST"},
		{"PutUser", "PutUser", "PUT"},
		{"UpdateUser", "UpdateUser", "PUT"},
		{"DeleteUser", "DeleteUser", "DELETE"},
		{"RemoveUser", "RemoveUser", "DELETE"},
		{"PatchUser", "PatchUser", "PATCH"},
		{"Unknown", "Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := b.extractHTTPMethod(stringToMethod(tt.method))
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	t.Parallel()
	if !contains([]string{"a", "b"}, "a") {
		t.Error("expected slice to contain 'a'")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Error("expected slice to not contain 'c'")
	}
}

func stringToType(name string) reflect.Type {
	type UserController struct{}
	type UserHandler struct{}
	type User struct{}
	type Empty struct{}

	switch name {
	case "UserController":
		return reflect.TypeOf(UserController{})
	case "UserHandler":
		return reflect.TypeOf(UserHandler{})
	case "User":
		return reflect.TypeOf(User{})
	default:
		return reflect.TypeOf(Empty{})
	}
}

func stringToMethod(name string) reflect.Method {
	type TestController struct{}

	m := map[string]func(){}
	m["GetUsers"] = func() {}
	m["FindUser"] = func() {}
	m["ListUsers"] = func() {}
	m["PostUser"] = func() {}
	m["CreateUser"] = func() {}
	m["AddUser"] = func() {}
	m["PutUser"] = func() {}
	m["UpdateUser"] = func() {}
	m["DeleteUser"] = func() {}
	m["RemoveUser"] = func() {}
	m["PatchUser"] = func() {}
	m["Unknown"] = func() {}

	t := reflect.TypeOf(TestController{})
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if method.Name == name {
			return method
		}
	}

	// Fallback: create a method with the given name
	if fn, ok := m[name]; ok {
		return reflect.Method{
			Name:    name,
			PkgPath: "",
			Type:    reflect.TypeOf(fn),
			Func:    reflect.ValueOf(fn),
			Index:   0,
		}
	}

	return reflect.Method{Name: name}
}
