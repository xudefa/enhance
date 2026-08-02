package core

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core/registry"
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

	if err != ErrBeanNotFound {
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

func TestHas(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container, WithName[*TestGenericBean]("hasBean"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !Has[*TestGenericBean](container, "hasBean") {
		t.Error("Expected Has to return true for registered bean")
	}
}

func TestHasNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	if Has[*TestGenericBean](container, "nonexistent") {
		t.Error("Expected Has to return false for nonexistent bean")
	}
}

func TestHasWrongType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container, WithName[*TestGenericBean]("hasBean2"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if Has[*TestGenericBean2](container, "hasBean2") {
		t.Error("Expected Has to return false for wrong type")
	}
}

func TestRegisterMultipleTypes(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container, WithName[*TestGenericBean]("bean1"))
	if err != nil {
		t.Fatalf("Register bean1 failed: %v", err)
	}

	err = Register[*TestGenericBean2](container, WithName[*TestGenericBean2]("bean2"))
	if err != nil {
		t.Fatalf("Register bean2 failed: %v", err)
	}

	if !Has[*TestGenericBean](container, "bean1") {
		t.Error("Expected bean1 to exist")
	}

	if !Has[*TestGenericBean2](container, "bean2") {
		t.Error("Expected bean2 to exist")
	}

	if Has[*TestGenericBean](container, "bean2") {
		t.Error("Expected bean2 to not match TestGenericBean type")
	}
}

func TestRegisterAndGetMultipleBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container, WithName[*TestGenericBean]("bean1"))
	if err != nil {
		t.Fatalf("Register bean1 failed: %v", err)
	}

	err = Register[*TestGenericBean](container, WithName[*TestGenericBean]("bean2"))
	if err != nil {
		t.Fatalf("Register bean2 failed: %v", err)
	}

	beans, err := container.Get(reflect.TypeOf((*TestGenericBean)(nil)))
	if err != nil {
		t.Fatalf("Get by type failed: %v", err)
	}

	if len(beans) != 2 {
		t.Errorf("Expected 2 beans, got %d", len(beans))
	}
}

func TestRegisterWithNilFactoryInDef(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	def := registry.BeanDef{
		Type:    reflect.TypeOf((*TestGenericBean)(nil)),
		Name:    "nilFactoryBean",
		Factory: nil,
	}

	err := container.RegisterBean(def)
	if err == nil {
		t.Fatal("Expected RegisterBean to fail with nil factory")
	}
}

func TestGetByNameWithDestroyedContainer(t *testing.T) {
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

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	_, err = GetByName[*TestGenericBean](container, "testBean")
	if err == nil {
		t.Error("Expected error for destroyed container")
	}
}

func TestMustGetWithDestroyedContainer(t *testing.T) {
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

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustGet to panic for destroyed container")
		}
	}()

	MustGet[*TestGenericBean](container, "testBean")
}

func TestRegisterWithPointerAndNonPointer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container, WithName[*TestGenericBean]("pointerBean"))
	if err != nil {
		t.Fatalf("Register pointer failed: %v", err)
	}

	bean, err := GetByName[*TestGenericBean](container, "pointerBean")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean == nil {
		t.Error("Expected bean to be non-nil")
	}
}

func TestRegisterWithAllOptions(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	destroyCalled := false

	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("allOptions"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			return &TestGenericBean{Name: "all-options"}, nil
		}),
		WithInit[*TestGenericBean](func(bean any) error {
			initCalled = true
			return nil
		}),
		WithDestroy[*TestGenericBean](func(bean any) error {
			destroyCalled = true
			return nil
		}),
		WithLazy[*TestGenericBean](false),
		WithPrimary[*TestGenericBean](true),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !initCalled {
		t.Error("Expected Init callback to be called")
	}

	bean := MustGet[*TestGenericBean](container, "allOptions")
	if bean.Name != "all-options" {
		t.Errorf("Expected name 'all-options', got %q", bean.Name)
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	if !destroyCalled {
		t.Error("Expected Destroy callback to be called")
	}
}

func TestRegisterWithPrototypeScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := 0
	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("prototype"),
		WithScope[*TestGenericBean]("prototype"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			callCount++
			return &TestGenericBean{Name: "prototype"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean1 := MustGet[*TestGenericBean](container, "prototype")
	bean2 := MustGet[*TestGenericBean](container, "prototype")

	if bean1 == bean2 {
		t.Error("Expected different instances for prototype scope")
	}

	if callCount != 2 {
		t.Errorf("Expected factory to be called 2 times, got %d", callCount)
	}
}

func TestRegisterDuplicateBeanGeneric(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container, WithName[*TestGenericBean]("duplicate"))
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	err = Register[*TestGenericBean](container, WithName[*TestGenericBean]("duplicate"))
	if err != nil {
		t.Errorf("Expected no error for duplicate registration, got: %v", err)
	}
}

func TestGetByNameWithGetByTypeAndName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container,
		WithName[*TestGenericBean]("typedBean"),
		WithFactory[*TestGenericBean](func(c ...any) (any, error) {
			return &TestGenericBean{Name: "typed"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := GetByName[*TestGenericBean](container, "typedBean")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean.Name != "typed" {
		t.Errorf("Expected name 'typed', got %q", bean.Name)
	}
}

func TestRegisterWithEmptyNameGeneric(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestGenericBean](container)
	if err != nil {
		t.Fatalf("Register with empty name failed: %v", err)
	}

	bean, err := GetByName[*TestGenericBean](container, "")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean == nil {
		t.Error("Expected bean to be non-nil")
	}
}
