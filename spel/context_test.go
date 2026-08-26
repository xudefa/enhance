package spel

import (
	"reflect"
	"testing"
)

func TestStandardEvaluationContext_GetRootObject(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext("root")
	if ctx.GetRootObject() != "root" {
		t.Errorf("GetRootObject() = %v, want 'root'", ctx.GetRootObject())
	}
}

func TestStandardEvaluationContext_SetRootObject(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext("initial")
	ctx.SetRootObject("updated")
	if ctx.GetRootObject() != "updated" {
		t.Errorf("after SetRootObject: got %v, want 'updated'", ctx.GetRootObject())
	}
}

func TestStandardEvaluationContext_GetVariable_Found(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext(nil)
	ctx.SetVariable("key", "value")

	val, ok := ctx.GetVariable("key")
	if !ok {
		t.Fatal("GetVariable should find existing variable")
	}
	if val != "value" {
		t.Errorf("GetVariable() = %v, want 'value'", val)
	}
}

func TestStandardEvaluationContext_GetVariable_NotFound(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext(nil)
	_, ok := ctx.GetVariable("nonexistent")
	if ok {
		t.Error("GetVariable should return false for missing variable")
	}
}

func TestStandardEvaluationContext_SetVariable_Overwrite(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext(nil)
	ctx.SetVariable("x", 1)
	ctx.SetVariable("x", 2)

	val, ok := ctx.GetVariable("x")
	if !ok || val != 2 {
		t.Errorf("SetVariable should overwrite: got %v, want 2", val)
	}
}

func TestStandardEvaluationContext_GetPropertyAccessor(t *testing.T) {
	t.Parallel()

	ctx := NewStandardEvaluationContext(nil)
	pa := ctx.GetPropertyAccessor()
	if pa == nil {
		t.Fatal("GetPropertyAccessor() should not return nil")
	}
}

func TestReflectPropertyAccessor_GetProperty_Basic(t *testing.T) {
	t.Parallel()

	type User struct {
		Name string
		Age  int
	}

	accessor := NewReflectPropertyAccessor()
	user := User{Name: "Alice", Age: 30}

	name, err := accessor.GetProperty(user, "Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Alice" {
		t.Errorf("Name = %v, want 'Alice'", name)
	}

	age, err := accessor.GetProperty(user, "Age")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if age != 30 {
		t.Errorf("Age = %v, want 30", age)
	}
}

func TestReflectPropertyAccessor_GetProperty_NilTarget(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	_, err := accessor.GetProperty(nil, "Name")
	if err == nil {
		t.Error("expected error for nil target")
	}
}

func TestReflectPropertyAccessor_GetProperty_NonStruct(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	_, err := accessor.GetProperty("not a struct", "Name")
	if err == nil {
		t.Error("expected error for non-struct target")
	}
}

func TestReflectPropertyAccessor_GetProperty_JSONTag(t *testing.T) {
	t.Parallel()

	type Tagged struct {
		FullName string `json:"full_name"`
	}

	accessor := NewReflectPropertyAccessor()
	tagged := Tagged{FullName: "Bob"}

	val, err := accessor.GetProperty(tagged, "full_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Bob" {
		t.Errorf("FullName via json tag = %v, want 'Bob'", val)
	}
}

func TestReflectPropertyAccessor_GetProperty_SpelTag(t *testing.T) {
	t.Parallel()

	type Tagged struct {
		DisplayName string `spel:"display_name"`
	}

	accessor := NewReflectPropertyAccessor()
	tagged := Tagged{DisplayName: "Charlie"}

	val, err := accessor.GetProperty(tagged, "display_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Charlie" {
		t.Errorf("DisplayName via spel tag = %v, want 'Charlie'", val)
	}
}

func TestReflectPropertyAccessor_SetProperty_Basic(t *testing.T) {
	t.Parallel()

	type User struct {
		Name string
	}

	accessor := NewReflectPropertyAccessor()
	user := &User{Name: "Old"}

	err := accessor.SetProperty(user, "Name", "New")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "New" {
		t.Errorf("Name = %v, want 'New'", user.Name)
	}
}

func TestReflectPropertyAccessor_SetProperty_NilTarget(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	err := accessor.SetProperty(nil, "Name", "value")
	if err == nil {
		t.Error("expected error for nil target")
	}
}

func TestReflectPropertyAccessor_SetProperty_NonStruct(t *testing.T) {
	t.Parallel()

	accessor := NewReflectPropertyAccessor()
	err := accessor.SetProperty("not a struct", "Name", "value")
	if err == nil {
		t.Error("expected error for non-struct target")
	}
}

func TestReflectPropertyAccessor_SetProperty_NotFound(t *testing.T) {
	t.Parallel()

	type Simple struct{ Name string }
	accessor := NewReflectPropertyAccessor()

	err := accessor.SetProperty(&Simple{}, "NonExistent", "value")
	if err == nil {
		t.Error("expected error for non-existent property")
	}
}

func TestReflectPropertyAccessor_SetProperty_NonNilableNilValue(t *testing.T) {
	t.Parallel()

	type Strict struct{ Count int }
	accessor := NewReflectPropertyAccessor()

	err := accessor.SetProperty(&Strict{}, "Count", nil)
	if err == nil {
		t.Error("expected error for nil to non-nilable type")
	}
}

func TestReflectPropertyAccessor_SetProperty_NilToNilable(t *testing.T) {
	t.Parallel()

	type Loose struct{ Items []string }
	accessor := NewReflectPropertyAccessor()
	loose := &Loose{Items: []string{"a"}}

	err := accessor.SetProperty(loose, "Items", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loose.Items != nil {
		t.Errorf("Items should be nil, got %v", loose.Items)
	}
}

func TestReflectPropertyAccessor_SetProperty_TypeConversion(t *testing.T) {
	t.Parallel()

	type Count struct{ N int64 }
	accessor := NewReflectPropertyAccessor()
	c := &Count{}

	err := accessor.SetProperty(c, "N", int(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.N != 42 {
		t.Errorf("N = %d, want 42", c.N)
	}
}

func TestReflectPropertyAccessor_SetProperty_InconvertibleType(t *testing.T) {
	t.Parallel()

	type Name struct{ V string }
	accessor := NewReflectPropertyAccessor()

	type Bad struct{ X int }
	err := accessor.SetProperty(&Name{}, "V", Bad{X: 1})
	if err == nil {
		t.Error("expected error for inconvertible type")
	}
}

func TestIsNilable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  any
		want bool
	}{
		{"non-nil slice", []int{1}, true},
		{"non-nil map", map[string]int{"a": 1}, true},
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", new(int), true},
		{"nil channel", (chan int)(nil), true},
		{"nil func", (func())(nil), true},
		{"int", 0, false},
		{"string", "", false},
		{"bool", false, false},
		{"struct", struct{}{}, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isNilable(reflect.TypeOf(tt.val))
			if got != tt.want {
				t.Errorf("isNilable(%T) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}
