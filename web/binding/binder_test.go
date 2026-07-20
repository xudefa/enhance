package binding

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type testForm struct {
	Name  string `form:"name,required"`
	Email string `form:"email"`
	Age   int    `form:"age"`
}

type testQuery struct {
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
	Search string  `form:"search"`
	Active bool    `form:"active"`
	Score  float64 `form:"score"`
}

type testJSON struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestNewBinder(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	if binder == nil {
		t.Fatal("expected binder to be created")
	}

	if binder.tagName != "form" {
		t.Errorf("expected default tag name 'form', got %s", binder.tagName)
	}
}

func TestWithTagName(t *testing.T) {
	t.Parallel()
	binder := NewBinder(WithTagName("json"))

	if binder.tagName != "json" {
		t.Errorf("expected tag name 'json', got %s", binder.tagName)
	}
}

func TestBinder_Bind(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	body := strings.NewReader("name=John&email=john@example.com&age=30")
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var form testForm
	err := binder.Bind(req, &form)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if form.Name != "John" {
		t.Errorf("expected name 'John', got %s", form.Name)
	}

	if form.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %s", form.Email)
	}

	if form.Age != 30 {
		t.Errorf("expected age 30, got %d", form.Age)
	}
}

func TestBinder_BindQuery(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	req := httptest.NewRequest("GET", "/test?page=1&limit=10&search=test&active=true&score=9.5", nil)

	var query testQuery
	err := binder.BindQuery(req, &query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if query.Page != 1 {
		t.Errorf("expected page 1, got %d", query.Page)
	}

	if query.Limit != 10 {
		t.Errorf("expected limit 10, got %d", query.Limit)
	}

	if query.Search != "test" {
		t.Errorf("expected search 'test', got %s", query.Search)
	}

	if !query.Active {
		t.Error("expected active to be true")
	}

	if query.Score != 9.5 {
		t.Errorf("expected score 9.5, got %f", query.Score)
	}
}

func TestBinder_BindJSON(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	body := strings.NewReader(`{"name":"John","email":"john@example.com","age":30}`)
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")

	var data testJSON
	err := binder.BindJSON(req, &data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Name != "John" {
		t.Errorf("expected name 'John', got %s", data.Name)
	}

	if data.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %s", data.Email)
	}

	if data.Age != 30 {
		t.Errorf("expected age 30, got %d", data.Age)
	}
}

func TestBinder_RequiredField(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	body := strings.NewReader("email=john@example.com")
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var form testForm
	err := binder.Bind(req, &form)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}

	bindErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if bindErr.Field != "name" {
		t.Errorf("expected field 'name', got %s", bindErr.Field)
	}
}

func TestBinder_ErrNotPointer(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	req := httptest.NewRequest("POST", "/test", nil)

	var form testForm
	err := binder.Bind(req, form)
	if err == nil {
		t.Fatal("expected error for non-pointer target")
	}

	if err != ErrNotPointer {
		t.Errorf("expected ErrNotPointer, got %v", err)
	}
}

func TestBinder_ErrNotStruct(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	req := httptest.NewRequest("POST", "/test", nil)

	var slice []string
	err := binder.Bind(req, &slice)
	if err == nil {
		t.Fatal("expected error for non-struct target")
	}

	if err != ErrNotStruct {
		t.Errorf("expected ErrNotStruct, got %v", err)
	}
}

func TestBinder_CustomTag(t *testing.T) {
	t.Parallel()
	type customForm struct {
		Name string `json:"name,required"`
	}

	binder := NewBinder(WithTagName("json"))

	body := strings.NewReader("name=John")
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var form customForm
	err := binder.Bind(req, &form)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if form.Name != "John" {
		t.Errorf("expected name 'John', got %s", form.Name)
	}
}

func TestBindingError_Error(t *testing.T) {
	t.Parallel()
	err := &Error{
		Message: "field is required",
		Field:   "name",
	}

	expected := "binding error: field \"name\": field is required"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestBinder_BindSlice(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	body := strings.NewReader("tags=go,web,test")
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	type sliceForm struct {
		Tags []string `form:"tags"`
	}

	var form sliceForm
	err := binder.Bind(req, &form)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(form.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(form.Tags))
	}

	if form.Tags[0] != "go" || form.Tags[1] != "web" || form.Tags[2] != "test" {
		t.Errorf("expected tags [go web test], got %v", form.Tags)
	}
}

func TestBinder_BindBool(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/test?active="+tt.value, nil)

		type boolForm struct {
			Active bool `form:"active"`
		}

		var form boolForm
		err := binder.BindQuery(req, &form)
		if err != nil {
			t.Fatalf("unexpected error for value '%s': %v", tt.value, err)
		}

		if form.Active != tt.expected {
			t.Errorf("for value '%s', expected %v, got %v", tt.value, tt.expected, form.Active)
		}
	}
}

func TestBinder_BindFloat(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	req := httptest.NewRequest("GET", "/test?score=9.5", nil)

	type floatForm struct {
		Score float64 `form:"score"`
	}

	var form floatForm
	err := binder.BindQuery(req, &form)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if form.Score != 9.5 {
		t.Errorf("expected score 9.5, got %f", form.Score)
	}
}

func TestBinder_SkipUnsettableFields(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	type unexportedForm struct {
		Name    string `form:"name"`
		private string `form:"private"`
	}

	body := strings.NewReader("name=John&private=secret")
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var form unexportedForm
	err := binder.Bind(req, &form)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if form.Name != "John" {
		t.Errorf("expected name 'John', got %s", form.Name)
	}

	if form.private != "" {
		t.Error("expected private field to remain empty")
	}
}

func TestBinder_ParseFormError(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	req := httptest.NewRequest("POST", "/test", nil)
	req.Body = &errorReader{}

	var form testForm
	err := binder.Bind(req, &form)
	if err == nil {
		t.Fatal("expected error for parse form failure")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (e *errorReader) Close() error {
	return nil
}

func BenchmarkBinder_Bind(b *testing.B) {
	binder := NewBinder()

	body := strings.NewReader("name=John&email=john@example.com&age=30")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var form testForm
		_ = binder.Bind(req, &form)
	}
}

func BenchmarkBinder_BindQuery(b *testing.B) {
	binder := NewBinder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test?page=1&limit=10&search=test", nil)

		var query testQuery
		_ = binder.BindQuery(req, &query)
	}
}

func BenchmarkBinder_BindJSON(b *testing.B) {
	binder := NewBinder()

	body := strings.NewReader(`{"name":"John","email":"john@example.com","age":30}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")

		var data testJSON
		_ = binder.BindJSON(req, &data)
	}
}
