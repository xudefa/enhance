package boot

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

// TestContainerAccessorHas tests the Has method of containerAccessorAdapter.
func TestContainerAccessorHas(t *testing.T) {
	t.Parallel()

	t.Run("with real container", func(t *testing.T) {
		t.Parallel()
		container := core.NewContainer()
		err := container.RegisterInstance(&TestBean{Name: "bean1"}, reflect.TypeOf(&TestBean{}))
		if err != nil {
			t.Fatalf("RegisterInstance failed: %v", err)
		}

		adapter := &containerAccessorAdapter{container: container}
		beanType := reflect.TypeOf(&TestBean{})
		generatedID := container.Generate(beanType)
		if !adapter.Has(generatedID) {
			t.Errorf("Expected Has(%s) to return true", generatedID)
		}
		if adapter.Has("nonexistent-id") {
			t.Error("Expected Has to return false for nonexistent ID")
		}
	})

	t.Run("with mock container", func(t *testing.T) {
		t.Parallel()
		mockContainer := &mockContainerForHas{
			listBeansResult: map[string]*registry.BeanDef{
				"test-id": {Name: "test-id"},
			},
		}
		adapter := &containerAccessorAdapter{container: mockContainer}
		if !adapter.Has("test-id") {
			t.Error("Expected Has to return true for existing ID")
		}
		if adapter.Has("missing-id") {
			t.Error("Expected Has to return false for missing ID")
		}
	})
}

// TestContainerAccessorHasMultipleBeans tests Has with multiple beans.
func TestContainerAccessorHasMultipleBeans(t *testing.T) {
	t.Parallel()

	type TestBean1 struct{ Name string }
	type TestBean2 struct{ Value int }

	container := core.NewContainer()

	err := container.RegisterInstance(&TestBean1{Name: "bean1"}, reflect.TypeOf(&TestBean1{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}
	err = container.RegisterInstance(&TestBean2{Value: 42}, reflect.TypeOf(&TestBean2{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}
	beanType1 := reflect.TypeOf(&TestBean1{})
	generatedID1 := container.Generate(beanType1)

	if !adapter.Has(generatedID1) {
		t.Errorf("Expected Has(%s) to return true", generatedID1)
	}
	if adapter.Has("nonexistent-id") {
		t.Error("Expected Has to return false for nonexistent ID")
	}
}
