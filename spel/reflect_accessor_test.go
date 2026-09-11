package spel

import (
	"testing"
)

func TestReflectPropertyAccessor_SetProperty_PointerTarget(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := &testUser{Name: "Old"}

	err := accessor.SetProperty(user, "Name", "New")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "New" {
		t.Errorf("got %v, want 'New'", user.Name)
	}
}

func TestReflectPropertyAccessor_GetProperty_PointerTarget(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := &testUser{Name: "Alice"}

	val, err := accessor.GetProperty(user, "Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Alice" {
		t.Errorf("got %v, want 'Alice'", val)
	}
}

func TestReflectPropertyAccessor_SetProperty_ConvertibleTypes(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	target := &testSettable{Count: 0}

	err := accessor.SetProperty(target, "Count", int(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Count != 42 {
		t.Errorf("got %d, want 42", target.Count)
	}
}

func TestReflectPropertyAccessor_SetProperty_InconvertibleTypes(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	target := &testSettable{Name: ""}

	type incompatibleStruct struct{ X int }
	err := accessor.SetProperty(target, "Name", incompatibleStruct{X: 1})
	if err == nil {
		t.Error("expected error for inconvertible types")
	}
}

func TestReflectPropertyAccessor_GetProperty_NotFound(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := testUser{Name: "Alice"}

	_, err := accessor.GetProperty(user, "NonExistent")
	if err == nil {
		t.Error("expected error for non-existent property")
	}
}
