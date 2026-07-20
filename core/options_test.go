package core

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core/registry"
)

type TestOptionBean struct {
	Value string
}

func TestWithName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestOptionBean](container, WithName[*TestOptionBean]("myBean"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := GetByName[*TestOptionBean](container, "myBean")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean == nil {
		t.Error("Expected bean to be non-nil")
	}
}

func TestWithType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	customType := reflect.TypeOf((*TestOptionBean)(nil))
	err := Register[*TestOptionBean](container, WithType[*TestOptionBean](customType))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	beans, err := container.Get(customType)
	if err != nil {
		t.Fatalf("Get by type failed: %v", err)
	}

	if len(beans) != 1 {
		t.Errorf("Expected 1 bean, got %d", len(beans))
	}
}

func TestWithFactory(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	customValue := "custom-factory-value"
	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("factoryBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return &TestOptionBean{Value: customValue}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean, err := GetByName[*TestOptionBean](container, "factoryBean")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if bean.Value != customValue {
		t.Errorf("Expected value %q, got %q", customValue, bean.Value)
	}
}

func TestWithScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := 0
	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("prototypeBean"),
		WithScope[*TestOptionBean]("prototype"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			callCount++
			return &TestOptionBean{Value: "prototype"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	bean1, _ := GetByName[*TestOptionBean](container, "prototypeBean")
	bean2, _ := GetByName[*TestOptionBean](container, "prototypeBean")

	if bean1 == bean2 {
		t.Error("Expected different instances for prototype scope")
	}

	if callCount != 2 {
		t.Errorf("Expected factory to be called 2 times, got %d", callCount)
	}
}

func TestWithInit(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("initBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return &TestOptionBean{Value: "init"}, nil
		}),
		WithInit[*TestOptionBean](func(bean any) error {
			initCalled = true
			b := bean.(*TestOptionBean)
			b.Value = "initialized"
			return nil
		}),
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

	bean, _ := GetByName[*TestOptionBean](container, "initBean")
	if bean.Value != "initialized" {
		t.Errorf("Expected value 'initialized', got %q", bean.Value)
	}
}

func TestWithDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	destroyCalled := false
	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("destroyBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return &TestOptionBean{Value: "destroy"}, nil
		}),
		WithDestroy[*TestOptionBean](func(bean any) error {
			destroyCalled = true
			return nil
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

	if !destroyCalled {
		t.Error("Expected Destroy callback to be called")
	}
}

func TestWithLazy(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	factoryCalled := false
	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("lazyBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			factoryCalled = true
			return &TestOptionBean{Value: "lazy"}, nil
		}),
		WithInit[*TestOptionBean](func(bean any) error {
			return nil
		}),
		WithLazy[*TestOptionBean](true),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if factoryCalled {
		t.Error("Expected factory not to be called for lazy bean during Initialize")
	}

	_, err = GetByName[*TestOptionBean](container, "lazyBean")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if !factoryCalled {
		t.Error("Expected factory to be called when accessing lazy bean")
	}
}

func TestWithPrimary(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("primaryBean"),
		WithPrimary[*TestOptionBean](true),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return &TestOptionBean{Value: "primary"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	beans, err := container.Get(reflect.TypeOf((*TestOptionBean)(nil)))
	if err != nil {
		t.Fatalf("Get by type failed: %v", err)
	}

	if len(beans) != 1 {
		t.Errorf("Expected 1 bean, got %d", len(beans))
	}

	bean := beans[0].(*TestOptionBean)
	if bean.Value != "primary" {
		t.Errorf("Expected value 'primary', got %q", bean.Value)
	}
}

func TestMultipleOptions(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	destroyCalled := false

	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("multiOptionBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return &TestOptionBean{Value: "multi"}, nil
		}),
		WithInit[*TestOptionBean](func(bean any) error {
			initCalled = true
			return nil
		}),
		WithDestroy[*TestOptionBean](func(bean any) error {
			destroyCalled = true
			return nil
		}),
		WithLazy[*TestOptionBean](false),
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

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	if !destroyCalled {
		t.Error("Expected Destroy callback to be called")
	}
}

func TestWithFactoryReturningError(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("errorFactoryBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return nil, ErrInitFailed
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = GetByName[*TestOptionBean](container, "errorFactoryBean")
	if err == nil {
		t.Error("Expected error from factory")
	}
}

func TestWithInitReturningError(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("initErrorBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return &TestOptionBean{Value: "init-error"}, nil
		}),
		WithInit[*TestOptionBean](func(bean any) error {
			return ErrInitFailed
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err == nil {
		t.Error("Expected Initialize to fail")
	}
}

func TestWithDestroyReturningError(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestOptionBean](container,
		WithName[*TestOptionBean]("destroyErrorBean"),
		WithFactory[*TestOptionBean](func(c ...any) (any, error) {
			return &TestOptionBean{Value: "destroy-error"}, nil
		}),
		WithDestroy[*TestOptionBean](func(bean any) error {
			return ErrDestroyFailed
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
	if err == nil {
		t.Error("Expected Destroy to fail")
	}
}

func TestOptionBeanDefModification(t *testing.T) {
	t.Parallel()
	def := &registry.BeanDef{}

	opt := WithName[*TestOptionBean]("testName")
	opt(def)

	if def.Name != "testName" {
		t.Errorf("Expected name 'testName', got %q", def.Name)
	}

	opt2 := WithLazy[*TestOptionBean](true)
	opt2(def)

	if !def.Lazy {
		t.Error("Expected Lazy to be true")
	}

	opt3 := WithPrimary[*TestOptionBean](true)
	opt3(def)

	if !def.Primary {
		t.Error("Expected Primary to be true")
	}
}
