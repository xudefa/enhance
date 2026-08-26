package validation

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBindFieldsFromValues_AllTypes(t *testing.T) {
	t.Parallel()

	type AllTypesRequest struct {
		Name   string  `form:"name"`
		Age    int     `form:"age"`
		Score  float64 `form:"score"`
		Active bool    `form:"active"`
		Count  uint    `form:"count"`
	}

	values := map[string][]string{
		"name":   {"John"},
		"age":    {"30"},
		"score":  {"95.5"},
		"active": {"true"},
		"count":  {"100"},
	}

	validator := NewTagValidator()
	var req AllTypesRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Name != "John" {
		t.Errorf("expected name 'John', got '%s'", req.Name)
	}
	if req.Age != 30 {
		t.Errorf("expected age 30, got %d", req.Age)
	}
	if req.Score != 95.5 {
		t.Errorf("expected score 95.5, got %f", req.Score)
	}
	if !req.Active {
		t.Error("expected active to be true")
	}
	if req.Count != 100 {
		t.Errorf("expected count 100, got %d", req.Count)
	}
}

func TestBindFieldsFromValues_InvalidInt(t *testing.T) {
	t.Parallel()

	type IntRequest struct {
		Age int `form:"age"`
	}

	values := map[string][]string{
		"age": {"not-a-number"},
	}

	validator := NewTagValidator()
	var req IntRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err == nil {
		t.Fatal("expected error for invalid int")
	}
}

func TestBindFieldsFromValues_InvalidFloat(t *testing.T) {
	t.Parallel()

	type FloatRequest struct {
		Score float64 `form:"score"`
	}

	values := map[string][]string{
		"score": {"not-a-float"},
	}

	validator := NewTagValidator()
	var req FloatRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err == nil {
		t.Fatal("expected error for invalid float")
	}
}

func TestBindFieldsFromValues_InvalidBool(t *testing.T) {
	t.Parallel()

	type BoolRequest struct {
		Active bool `form:"active"`
	}

	values := map[string][]string{
		"active": {"not-a-bool"},
	}

	validator := NewTagValidator()
	var req BoolRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err == nil {
		t.Fatal("expected error for invalid bool")
	}
}

func TestBindFieldsFromValues_UnsupportedType(t *testing.T) {
	t.Parallel()

	type UnsupportedRequest struct {
		Data []byte `form:"data"`
	}

	values := map[string][]string{
		"data": {"test"},
	}

	validator := NewTagValidator()
	var req UnsupportedRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestBindFieldsFromValues_WithValidator(t *testing.T) {
	t.Parallel()

	type ValidatedRequest struct {
		Email string `form:"email" validate:"email"`
	}

	values := map[string][]string{
		"email": {"invalid-email"},
	}

	validator := NewTagValidator()
	var req ValidatedRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err == nil {
		t.Fatal("expected validation error for invalid email")
	}
}

func TestBindFieldsFromValues_PointerField(t *testing.T) {
	t.Parallel()

	type PointerRequest struct {
		Name *string `form:"name"`
	}

	values := map[string][]string{
		"name": {"John"},
	}

	validator := NewTagValidator()
	var req PointerRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Name == nil || *req.Name != "John" {
		t.Errorf("expected name 'John', got %v", req.Name)
	}
}

func TestBindFieldsFromValues_NonSettableField(t *testing.T) {
	t.Parallel()

	type NonSettableRequest struct {
		name string // lowercase, unexported field
	}

	values := map[string][]string{
		"name": {"John"},
	}

	validator := NewTagValidator()
	var req NonSettableRequest
	err := bindFieldsFromValues(values, &req, validator)
	// Should not error, just skip unexported fields
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBindFieldsFromValues_EmptyValues(t *testing.T) {
	t.Parallel()

	type EmptyRequest struct {
		Name string `form:"name"`
	}

	values := map[string][]string{}

	validator := NewTagValidator()
	var req EmptyRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Name != "" {
		t.Errorf("expected empty name, got '%s'", req.Name)
	}
}

func TestNewDefaultBinder_NilValidator(t *testing.T) {
	t.Parallel()

	binder := NewDefaultBinder(nil)
	if binder == nil {
		t.Fatal("expected non-nil binder")
	}
	if binder.Validator == nil {
		t.Fatal("expected non-nil validator")
	}
}

func TestNewDefaultBinder_WithValidator(t *testing.T) {
	t.Parallel()

	validator := NewTagValidator()
	binder := NewDefaultBinder(validator)
	if binder == nil {
		t.Fatal("expected non-nil binder")
	}
	if binder.Validator != validator {
		t.Fatal("expected validator to be set")
	}
}

func TestBindQuery(t *testing.T) {
	t.Parallel()

	type QueryRequest struct {
		Page  int `form:"page"`
		Limit int `form:"limit"`
	}

	binder := NewDefaultBinder(nil)

	values := map[string][]string{
		"page":  {"1"},
		"limit": {"20"},
	}

	var req QueryRequest
	err := binder.bindQuery(values, &req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Page != 1 {
		t.Errorf("expected page 1, got %d", req.Page)
	}
	if req.Limit != 20 {
		t.Errorf("expected limit 20, got %d", req.Limit)
	}
}

func TestBindFieldsFromValues_Int8Int16(t *testing.T) {
	t.Parallel()

	type SmallIntRequest struct {
		SmallInt8  int8  `form:"small8"`
		SmallInt16 int16 `form:"small16"`
	}

	values := map[string][]string{
		"small8":  {"10"},
		"small16": {"200"},
	}

	validator := NewTagValidator()
	var req SmallIntRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.SmallInt8 != 10 {
		t.Errorf("expected small8 10, got %d", req.SmallInt8)
	}
	if req.SmallInt16 != 200 {
		t.Errorf("expected small16 200, got %d", req.SmallInt16)
	}
}

func TestBindFieldsFromValues_Uint8Uint16(t *testing.T) {
	t.Parallel()

	type SmallUintRequest struct {
		SmallUint8  uint8  `form:"small8"`
		SmallUint16 uint16 `form:"small16"`
	}

	values := map[string][]string{
		"small8":  {"10"},
		"small16": {"200"},
	}

	validator := NewTagValidator()
	var req SmallUintRequest
	err := bindFieldsFromValues(values, &req, validator)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.SmallUint8 != 10 {
		t.Errorf("expected small8 10, got %d", req.SmallUint8)
	}
	if req.SmallUint16 != 200 {
		t.Errorf("expected small16 200, got %d", req.SmallUint16)
	}
}

func TestDefaultBinder_BindForm(t *testing.T) {
	t.Parallel()

	type FormRequest struct {
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	binder := NewDefaultBinder(nil)

	req := createFormRequest(map[string]string{
		"name": "John",
		"age":  "30",
	})

	var formReq FormRequest
	err := binder.Bind(req, &formReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if formReq.Name != "John" {
		t.Errorf("expected name 'John', got '%s'", formReq.Name)
	}
	if formReq.Age != 30 {
		t.Errorf("expected age 30, got %d", formReq.Age)
	}
}

func TestDefaultBinder_BindJSON(t *testing.T) {
	t.Parallel()

	type JSONRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	binder := NewDefaultBinder(nil)

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var jsonReq JSONRequest
	err := binder.Bind(req, &jsonReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jsonReq.Name != "John" {
		t.Errorf("expected name 'John', got '%s'", jsonReq.Name)
	}
	if jsonReq.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got '%s'", jsonReq.Email)
	}
}

func TestBindAndValidate_Form(t *testing.T) {
	t.Parallel()

	type TestRequest struct {
		Name string `form:"name" validate:"required"`
	}

	req := createFormRequest(map[string]string{
		"name": "John",
	})

	var testReq TestRequest
	err := BindAndValidate(req, &testReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if testReq.Name != "John" {
		t.Errorf("expected name 'John', got '%s'", testReq.Name)
	}
}

func createFormRequest(values map[string]string) *http.Request {
	form := url.Values{}
	for k, v := range values {
		form.Add(k, v)
	}
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}
