package openapi

import (
	"testing"
)

func TestPathItem_Struct(t *testing.T) {
	t.Parallel()
	pi := PathItem{
		Summary:     "Test summary",
		Description: "Test description",
		Get:         &OperationObject{Summary: "Get"},
		Post:        &OperationObject{Summary: "Post"},
		Put:         &OperationObject{Summary: "Put"},
		Delete:      &OperationObject{Summary: "Delete"},
		Patch:       &OperationObject{Summary: "Patch"},
		Parameters:  []ParameterObject{{Name: "id", In: "path"}},
		Tags:        []string{"users"},
	}

	if pi.Summary != "Test summary" {
		t.Errorf("expected summary 'Test summary', got %s", pi.Summary)
	}
	if pi.Get.Summary != "Get" {
		t.Errorf("expected GET summary 'Get', got %s", pi.Get.Summary)
	}
	if len(pi.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(pi.Parameters))
	}
}

func TestOperationObject_Struct(t *testing.T) {
	t.Parallel()
	op := OperationObject{
		Summary:     "Test operation",
		Description: "Test description",
		OperationID: "testOp",
		Tags:        []string{"users"},
		Parameters:  []ParameterObject{{Name: "id", In: "path"}},
		RequestBody: &RequestBodyObject{Required: true},
		Responses:   map[string]ResponseObject{"200": {Description: "OK"}},
		Deprecated:  true,
		Security:    []map[string][]string{{"bearer": {}}},
	}

	if op.Summary != "Test operation" {
		t.Errorf("expected summary 'Test operation', got %s", op.Summary)
	}
	if op.OperationID != "testOp" {
		t.Errorf("expected operationId 'testOp', got %s", op.OperationID)
	}
	if !op.Deprecated {
		t.Error("expected deprecated to be true")
	}
}

func TestParameterObject_Struct(t *testing.T) {
	t.Parallel()
	param := ParameterObject{
		Name:        "id",
		In:          "path",
		Description: "User ID",
		Required:    true,
		Schema:      &SchemaObject{Type: "integer"},
		Example:     123,
	}

	if param.Name != "id" {
		t.Errorf("expected name 'id', got %s", param.Name)
	}
	if param.In != "path" {
		t.Errorf("expected in 'path', got %s", param.In)
	}
	if !param.Required {
		t.Error("expected required to be true")
	}
}

func TestRequestBodyObject_Struct(t *testing.T) {
	t.Parallel()
	body := RequestBodyObject{
		Description: "User data",
		Required:    true,
		Content:     map[string]MediaTypeObject{"application/json": {}},
	}

	if body.Description != "User data" {
		t.Errorf("expected description 'User data', got %s", body.Description)
	}
	if !body.Required {
		t.Error("expected required to be true")
	}
}

func TestResponseObject_Struct(t *testing.T) {
	t.Parallel()
	resp := ResponseObject{
		Description: "Success",
		Headers:     map[string]HeaderObject{"X-Request-ID": {Description: "Request ID"}},
		Content:     map[string]MediaTypeObject{"application/json": {}},
		Links:       map[string]LinkObject{"self": {OperationID: "getUser"}},
	}

	if resp.Description != "Success" {
		t.Errorf("expected description 'Success', got %s", resp.Description)
	}
}

func TestAPITag_Struct(t *testing.T) {
	t.Parallel()
	tag := APITag{
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

func TestAPIOperation_Struct(t *testing.T) {
	t.Parallel()
	op := APIOperation{
		Summary:     "Get user",
		Description: "Get user by ID",
		OperationID: "getUser",
		Tags:        []string{"users"},
		Deprecated:  true,
	}

	if op.Summary != "Get user" {
		t.Errorf("expected summary 'Get user', got %s", op.Summary)
	}
	if !op.Deprecated {
		t.Error("expected deprecated to be true")
	}
}

func TestAPIParam_Struct(t *testing.T) {
	t.Parallel()
	param := APIParam{
		Name:        "id",
		In:          "path",
		Description: "User ID",
		Required:    true,
		Example:     123,
	}

	if param.Name != "id" {
		t.Errorf("expected name 'id', got %s", param.Name)
	}
	if param.In != "path" {
		t.Errorf("expected in 'path', got %s", param.In)
	}
}

func TestAPIResponse_Struct(t *testing.T) {
	t.Parallel()
	resp := APIResponse{
		StatusCode:  200,
		Description: "Success",
		Type:        "User",
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}
	if resp.Description != "Success" {
		t.Errorf("expected description 'Success', got %s", resp.Description)
	}
}

func TestAPISecurity_Struct(t *testing.T) {
	t.Parallel()
	sec := APISecurity{
		Name:   "bearer",
		Scopes: []string{"read", "write"},
	}

	if sec.Name != "bearer" {
		t.Errorf("expected name 'bearer', got %s", sec.Name)
	}
	if len(sec.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(sec.Scopes))
	}
}
