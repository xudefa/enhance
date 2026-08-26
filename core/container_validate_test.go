package core

import (
	"reflect"
	"testing"
)

func TestToPtrType_FromValueType(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(TestService{})
	ptrType := toPtrType(typ)
	if ptrType.Kind() != reflect.Ptr {
		t.Errorf("expected pointer kind, got %v", ptrType.Kind())
	}
}

func TestToPtrType_AlreadyPointer(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf((*TestService)(nil))
	ptrType := toPtrType(typ)
	if ptrType != typ {
		t.Errorf("expected same type for already-pointer, got %v != %v", ptrType, typ)
	}
}

func TestValidate_EmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	err := ext.Validate()
	if err != nil {
		t.Errorf("expected no error for empty container, got %v", err)
	}
}

func TestValidate_AllDepsRegistered(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type Dep struct{ Name string }
	type Svc struct{ D *Dep `inject:""` }

	_ = Register[*Dep](container)
	_ = Register[*Svc](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_MissingDep(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type Svc struct{ D *TestService `inject:""` }
	_ = Register[*Svc](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("expected error for missing dependency")
	}
}

func TestValidate_CircularDep(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_ = Register[*A](container)
	_ = Register[*B](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}

func TestValidate_AfterInit(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	_ = Register[*TestService](container)

	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ext := container.(ContainerExt)
	err = ext.Validate()
	if err == nil {
		t.Error("expected error for already-initialized container")
	}
}

func TestValidate_WithParentContainer(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	_ = Register[*TestService](parent)

	child := NewContainer()
	childExt := child.(ContainerExt)
	childExt.SetParent(parent)

	type Svc struct{ S *TestService `inject:""` }
	_ = Register[*Svc](child)

	err := childExt.Validate()
	if err != nil {
		t.Errorf("expected no error with parent providing dep, got %v", err)
	}
}

func TestValidate_NonStructType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	_ = Register[*TestService](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("expected no error for non-struct type, got %v", err)
	}
}

func TestValidate_ParentNotContainerExt(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type Svc struct{ D *TestService `inject:""` }
	_ = Register[*Svc](container)

	// parent without ContainerExt interface - no parent set, should still error
	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("expected error when parent cannot provide type info")
	}
}
