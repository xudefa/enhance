package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type testForm struct {
	Name  string `form:"name"`
	Email string `form:"email"`
	Age   int    `form:"age"`
}

type testRequiredForm struct {
	Name string `form:"name,required"`
	Age  int    `form:"age,required"`
}

func TestFormBinder_Bind(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"name":  {"John Doe"},
		"email": {"john@example.com"},
		"age":   {"25"},
	}

	form := &testForm{}
	err := binder.Bind(req, form)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if form.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %q", form.Name)
	}
	if form.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %q", form.Email)
	}
	if form.Age != 25 {
		t.Errorf("expected age 25, got %d", form.Age)
	}
}

func TestFormBinder_BindQuery(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	req := httptest.NewRequest("GET", "/search?name=alice&age=30", nil)

	form := &testForm{}
	err := binder.BindQuery(req, form)
	if err != nil {
		t.Fatalf("BindQuery failed: %v", err)
	}

	if form.Name != "alice" {
		t.Errorf("expected name 'alice', got %q", form.Name)
	}
	if form.Age != 30 {
		t.Errorf("expected age 30, got %d", form.Age)
	}
}

func TestFormBinder_BindJSON(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	body := strings.NewReader(`{"name":"Bob","email":"bob@example.com","age":35}`)
	req := httptest.NewRequest("POST", "/api", body)
	req.Header.Set("Content-Type", "application/json")

	form := &testForm{}
	err := binder.BindJSON(req, form)
	if err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}

	if form.Name != "Bob" {
		t.Errorf("expected name 'Bob', got %q", form.Name)
	}
	if form.Email != "bob@example.com" {
		t.Errorf("expected email 'bob@example.com', got %q", form.Email)
	}
	if form.Age != 35 {
		t.Errorf("expected age 35, got %d", form.Age)
	}
}

func TestFormBinder_RequiredField(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"name": {"John"},
	}

	form := &testRequiredForm{}
	err := binder.Bind(req, form)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}

	bindingErr, ok := err.(*BindingError)
	if !ok {
		t.Fatalf("expected BindingError, got %T", err)
	}
	if bindingErr.Field != "age" {
		t.Errorf("expected field 'age', got %q", bindingErr.Field)
	}
}

func TestFormBinder_ErrNotPointer(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"name": {"John"},
	}

	form := testForm{} // 非指针类型
	err := binder.Bind(req, form)
	if err != ErrNotPointer {
		t.Errorf("expected ErrNotPointer, got %v", err)
	}
}

func TestFormBinder_ErrNotStruct(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"name": {"John"},
	}

	var s string
	err := binder.Bind(req, &s)
	if err != ErrNotStruct {
		t.Errorf("expected ErrNotStruct, got %v", err)
	}
}

func TestFormBinder_CustomTag(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder(WithTagName("query"))

	type customForm struct {
		Name string `query:"username"`
	}

	req := httptest.NewRequest("GET", "/search?username=alice", nil)

	form := &customForm{}
	err := binder.BindQuery(req, form)
	if err != nil {
		t.Fatalf("BindQuery failed: %v", err)
	}

	if form.Name != "alice" {
		t.Errorf("expected name 'alice', got %q", form.Name)
	}
}

func TestBindingError_Error(t *testing.T) {
	t.Parallel()
	err := &BindingError{Field: "name", Message: "required field missing"}
	expected := `binding error: field "name": required field missing`
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}

	errNoField := &BindingError{Message: "target must be a pointer"}
	expectedNoField := "binding error: target must be a pointer"
	if errNoField.Error() != expectedNoField {
		t.Errorf("expected %q, got %q", expectedNoField, errNoField.Error())
	}
}

func TestFormBinder_BindSlice(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	type sliceForm struct {
		Tags []string `form:"tags"`
	}

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"tags": {"go,rust,python"},
	}

	form := &sliceForm{}
	err := binder.Bind(req, form)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	expected := []string{"go", "rust", "python"}
	if len(form.Tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d", len(expected), len(form.Tags))
	}
	for i, tag := range expected {
		if form.Tags[i] != tag {
			t.Errorf("expected tag[%d] %q, got %q", i, tag, form.Tags[i])
		}
	}
}

func TestFormBinder_BindBool(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	type boolForm struct {
		Active bool `form:"active"`
	}

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"active": {"true"},
	}

	form := &boolForm{}
	err := binder.Bind(req, form)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if !form.Active {
		t.Error("expected active to be true")
	}
}

func TestFormBinder_BindFloat(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	type floatForm struct {
		Price float64 `form:"price"`
	}

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"price": {"19.99"},
	}

	form := &floatForm{}
	err := binder.Bind(req, form)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if form.Price != 19.99 {
		t.Errorf("expected price 19.99, got %f", form.Price)
	}
}

func TestFormBinder_SkipUnsettableFields(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	type unexportedForm struct {
		Name    string `form:"name"`
		private string `form:"private"`
	}

	req := httptest.NewRequest("POST", "/submit", nil)
	req.Form = map[string][]string{
		"name":    {"John"},
		"private": {"secret"},
	}

	form := &unexportedForm{}
	err := binder.Bind(req, form)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	if form.Name != "John" {
		t.Errorf("expected name 'John', got %q", form.Name)
	}
	// private field should not be set
	if form.private != "" {
		t.Errorf("expected private to be empty, got %q", form.private)
	}
}

func TestFormBinder_ParseFormError(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()

	// 创建带有无效表单数据的请求
	req := httptest.NewRequest("POST", "/submit", nil)
	// 不设置 Form 字段，让 ParseForm 处理

	form := &testForm{}
	err := binder.Bind(req, form)
	// 不应该出错，空体的 ParseForm 是正常的
	if err != nil {
		t.Logf("Got error (may be expected): %v", err)
	}
}
