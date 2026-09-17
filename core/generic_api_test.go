package core

import (
	"errors"
	"testing"
)

type TestGenericBean struct {
	Name string
}

type TestGenericBean2 struct {
	Value int
}

func TestRegisterDefaultFactory(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container, WithName[*TestGenericBean]("defaultFactory"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := GetByName[*TestGenericBean](container, "defaultFactory")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean == nil {
		t.Error("Expected bean to be non-nil")
	}
}

func TestRegisterWithCustomFactory(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("customFactory"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			return &TestGenericBean{Name: "custom"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := GetByName[*TestGenericBean](container, "customFactory")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean.Name != "custom" {
		t.Errorf("Expected name 'custom', got %q", bean.Name)
	}
}

func TestRegisterDefaultScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := 0
	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("defaultScope"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			callCount++
			return &TestGenericBean{Name: "scope"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, _ = GetByName[*TestGenericBean](container, "defaultScope")
	_, _ = GetByName[*TestGenericBean](container, "defaultScope")

	if callCount != 1 {
		t.Errorf("Expected factory to be called once (singleton by default), got %d", callCount)
	}
}

func TestGetByName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("testBean"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			return &TestGenericBean{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := GetByName[*TestGenericBean](container, "testBean")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean.Name != "test" {
		t.Errorf("Expected name 'test', got %q", bean.Name)
	}
}

func TestGetByNameEmpty(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("testBean"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			return &TestGenericBean{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := GetByName[*TestGenericBean](container, "")
	if err != nil {
		t.Fatalf("GetByName with empty name failed: %v", err)
	}

	if bean.Name != "test" {
		t.Errorf("Expected name 'test', got %q", bean.Name)
	}
}

func TestGetByNameNotFoundGeneric(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_, err := GetByName[*TestGenericBean](container, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent bean")
	}

	if !errors.Is(err, ErrBeanNotFound) {
		t.Errorf("Expected ErrBeanNotFound, got %v", err)
	}
}

func TestGetByNameWithTypeNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_, err := GetByName[*TestGenericBean](container, "")
	if err == nil {
		t.Error("Expected error for bean not found by type")
	}
}

func TestMustGetGeneric(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("mustGetBean"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			return &TestGenericBean{Name: "must-get"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean := MustGet[*TestGenericBean](container, "mustGetBean")
	if bean.Name != "must-get" {
		t.Errorf("Expected name 'must-get', got %q", bean.Name)
	}
}

func TestMustGetPanic(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustGet to panic for nonexistent bean")
		}
	}()

	MustGet[*TestGenericBean](container, "nonexistent")
}

func TestMustGetEmptyName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("mustGetBean2"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			return &TestGenericBean{Name: "must-get-2"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean := MustGet[*TestGenericBean](container, "")
	if bean.Name != "must-get-2" {
		t.Errorf("Expected name 'must-get-2', got %q", bean.Name)
	}
}
