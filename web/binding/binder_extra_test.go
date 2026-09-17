package binding

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBinder_SliceIntOverflowReturnsError(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	type sliceIntForm struct {
		Vals []int16 `form:"vals"`
	}

	req := httptest.NewRequest("GET", "/test?vals=1,9999999999", nil)
	var form sliceIntForm
	if err := binder.BindQuery(req, &form); err == nil {
		t.Fatal("expected overflow error for slice element, got nil")
	}
}

func TestBinder_BindRepeatedQueryValuesToSlice(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	type sliceForm struct {
		Tags []string `form:"tags"`
	}

	req := httptest.NewRequest("GET", "/test?tags=a&tags=b&tags=c", nil)

	var form sliceForm
	if err := binder.BindQuery(req, &form); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"a", "b", "c"}
	if len(form.Tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %v", len(expected), len(form.Tags), form.Tags)
	}
	for i, v := range expected {
		if form.Tags[i] != v {
			t.Errorf("tags[%d] = %q, want %q", i, form.Tags[i], v)
		}
	}
}

func TestBinder_BindRepeatedCommaAndQueryValues(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	type sliceForm struct {
		Tags []string `form:"tags"`
	}

	req := httptest.NewRequest("GET", "/test?tags=a,b&tags=c", nil)

	var form sliceForm
	if err := binder.BindQuery(req, &form); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"a", "b", "c"}
	if len(form.Tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %v", len(expected), len(form.Tags), form.Tags)
	}
	for i, v := range expected {
		if form.Tags[i] != v {
			t.Errorf("tags[%d] = %q, want %q", i, form.Tags[i], v)
		}
	}
}

func TestBinder_BindRepeatedIntValuesToSlice(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	type intSliceForm struct {
		Nums []int32 `form:"nums"`
	}

	req := httptest.NewRequest("GET", "/test?nums=1&nums=2&nums=3", nil)

	var form intSliceForm
	if err := binder.BindQuery(req, &form); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int32{1, 2, 3}
	if len(form.Nums) != len(expected) {
		t.Fatalf("expected %d nums, got %d: %v", len(expected), len(form.Nums), form.Nums)
	}
	for i, v := range expected {
		if form.Nums[i] != v {
			t.Errorf("nums[%d] = %d, want %d", i, form.Nums[i], v)
		}
	}
}

func TestBinder_ScalarFieldTakesFirstValue(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	type scalarForm struct {
		Name string `form:"name"`
	}

	req := httptest.NewRequest("GET", "/test?name=first&name=second", nil)

	var form scalarForm
	if err := binder.BindQuery(req, &form); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if form.Name != "first" {
		t.Errorf("expected first value, got %q", form.Name)
	}
}

func TestBinder_RequiredFieldRejectsEmptyValue(t *testing.T) {
	t.Parallel()
	binder := NewBinder()

	// testForm.Name 标记为 required，?name= 视为缺失
	req := httptest.NewRequest("GET", "/test?name=&email=x@y.com", nil)

	var form testForm
	err := binder.BindQuery(req, &form)
	if err == nil {
		t.Fatal("expected error for empty required field")
	}

	bindErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if bindErr.Field != "name" {
		t.Errorf("expected field 'name', got %q", bindErr.Field)
	}
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

		var target testJSON
		_ = binder.BindJSON(req, &target)
	}
}
