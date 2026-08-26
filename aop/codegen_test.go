package aop

import (
	"reflect"
	"testing"
)

func TestGeneratedProxyRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()

	type TestProxy struct{}
	proxyType := reflect.TypeOf(TestProxy{})

	registry.Register("test-bean", proxyType)

	result, ok := registry.Get("test-bean")
	if !ok {
		t.Fatal("expected to find registered proxy")
	}
	if result != proxyType {
		t.Errorf("expected proxy type %v, got %v", proxyType, result)
	}
}

func TestGeneratedProxyRegistry_Get_NotFound(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()

	_, ok := registry.Get("non-existent")
	if ok {
		t.Error("expected not to find non-existent proxy")
	}
}

func TestGeneratedProxyRegistry_Has(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()

	type TestProxy struct{}
	registry.Register("test-bean", reflect.TypeOf(TestProxy{}))

	if !registry.Has("test-bean") {
		t.Error("expected Has to return true for registered bean")
	}

	if registry.Has("non-existent") {
		t.Error("expected Has to return false for non-existent bean")
	}
}

func TestGeneratedProxyRegistry_List(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()

	type TestProxy1 struct{}
	type TestProxy2 struct{}
	registry.Register("bean1", reflect.TypeOf(TestProxy1{}))
	registry.Register("bean2", reflect.TypeOf(TestProxy2{}))

	ids := registry.List()
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}

	// Check that both beans are in the list
	found1 := false
	found2 := false
	for _, id := range ids {
		if id == "bean1" {
			found1 = true
		}
		if id == "bean2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Error("expected both bean1 and bean2 in list")
	}
}

func TestGeneratedProxyRegistry_Clear(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()

	type TestProxy struct{}
	registry.Register("test-bean", reflect.TypeOf(TestProxy{}))

	if !registry.Has("test-bean") {
		t.Fatal("expected bean to be registered before clear")
	}

	registry.Clear()

	if registry.Has("test-bean") {
		t.Error("expected bean to be removed after clear")
	}
}

func TestRegisterGeneratedProxy(t *testing.T) {
	t.Parallel()

	type TestProxy struct{}
	proxyType := reflect.TypeOf(TestProxy{})

	RegisterGeneratedProxy("global-test-bean", proxyType)

	result, ok := GetGeneratedProxy("global-test-bean")
	if !ok {
		t.Fatal("expected to find registered proxy")
	}
	if result != proxyType {
		t.Errorf("expected proxy type %v, got %v", proxyType, result)
	}
}

func TestHasGeneratedProxy(t *testing.T) {
	t.Parallel()

	type TestProxy struct{}
	RegisterGeneratedProxy("has-test-bean", reflect.TypeOf(TestProxy{}))

	if !HasGeneratedProxy("has-test-bean") {
		t.Error("expected HasGeneratedProxy to return true")
	}

	if HasGeneratedProxy("non-existent-bean") {
		t.Error("expected HasGeneratedProxy to return false for non-existent")
	}
}

func TestGeneratedProxyFactory_Create(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()
	factory := &GeneratedProxyFactory{registry: registry}

	// Create a proxy struct with Target field - must be pointer type
	type TestProxy struct {
		Target any
	}
	proxyType := reflect.TypeOf(&TestProxy{}) // Pointer type
	registry.Register("test-bean", proxyType)

	target := &struct{}{}
	proxy, err := factory.Create("test-bean", target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}

	// Verify it's a pointer
	proxyValue := reflect.ValueOf(proxy)
	if proxyValue.Kind() != reflect.Ptr {
		t.Error("expected proxy to be a pointer")
	}

	// Verify Target field is set
	targetField := proxyValue.Elem().FieldByName("Target")
	if !targetField.IsValid() {
		t.Fatal("expected Target field to exist")
	}
	if targetField.Interface() != target {
		t.Errorf("expected Target to be set to target object")
	}
}

func TestGeneratedProxyFactory_Create_NotFound(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()
	factory := &GeneratedProxyFactory{registry: registry}

	_, err := factory.Create("non-existent", &struct{}{})
	if err == nil {
		t.Error("expected error for non-existent bean")
	}
}

func TestGeneratedProxyFactory_CreateOrFallback_WithFallback(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()
	factory := &GeneratedProxyFactory{registry: registry}

	target := &struct{}{}
	fallback := &mockWeaver{result: target}

	// No proxy registered, should fallback
	result := factory.CreateOrFallback("non-existent", target, fallback)
	if result != target {
		t.Error("expected fallback result")
	}
	if !fallback.called {
		t.Error("expected fallback weaver to be called")
	}
}

func TestGeneratedProxyFactory_CreateOrFallback_WithoutFallback(t *testing.T) {
	t.Parallel()

	registry := NewGeneratedProxyRegistry()
	factory := &GeneratedProxyFactory{registry: registry}

	target := &struct{}{}

	// No proxy registered, no fallback, should return target
	result := factory.CreateOrFallback("non-existent", target, nil)
	if result != target {
		t.Error("expected target to be returned when no proxy and no fallback")
	}
}

func TestAspectMetadataExtractor_Extract(t *testing.T) {
	t.Parallel()

	extractor := NewAspectMetadataExtractor()

	// Test with non-struct type
	nonStruct := reflect.TypeOf("")
	result := extractor.Extract(nonStruct)
	if result != nil {
		t.Error("expected nil for non-struct type")
	}

	// Test with struct without aspects field
	type NoAspects struct{}
	result = extractor.Extract(reflect.TypeOf(NoAspects{}))
	if result != nil {
		t.Error("expected nil for struct without aspects field")
	}
}

func TestAspectMetadataExtractor_ExtractFromBeanID(t *testing.T) {
	t.Parallel()

	extractor := NewAspectMetadataExtractor()

	// Test with non-existent bean
	result := extractor.ExtractFromBeanID("non-existent")
	if result != nil {
		t.Error("expected nil for non-existent bean")
	}
}

type mockWeaver struct {
	called bool
	result any
}

func (m *mockWeaver) Weave(target any) any {
	m.called = true
	return m.result
}

func (m *mockWeaver) AddAspects(aspects ...*AspectMeta) {
	// No-op for mock
}
