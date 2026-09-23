package scope

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// ==================== SingletonScope 测试 ====================

func TestSingletonScopeGet(t *testing.T) {
	t.Parallel()
	scope := NewSingletonScope()

	callCount := int32(0)
	factory := func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return "instance", nil
	}

	// First call should create instance
	instance1, err := scope.Get("bean1", factory)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Second call should return cached instance
	instance2, err := scope.Get("bean1", factory)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if instance1 != instance2 {
		t.Error("Expected same instance for singleton scope")
	}

	// Factory should only be called once
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected factory to be called once, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestSingletonScope_Get_Cached(t *testing.T) {
	t.Parallel()
	scope := NewSingletonScope()
	var count atomic.Int32
	factory := func(c ...any) (any, error) {
		count.Add(1)
		return "bean", nil
	}

	v1, _ := scope.Get("b1", factory)
	v2, _ := scope.Get("b1", factory)
	if v1 != v2 {
		t.Error("expected same cached instance")
	}
	if count.Load() != 1 {
		t.Errorf("expected factory called once, got %d", count.Load())
	}
}

func TestSingletonScope_Get_Error(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()
	expected := errors.New("factory error")
	factory := func(c ...any) (any, error) {
		return nil, expected
	}

	_, err := s.Get("b1", factory)
	if !errors.Is(err, expected) {
		t.Errorf("expected factory error, got %v", err)
	}
}

func TestSingletonScopeDifferentBeans(t *testing.T) {
	t.Parallel()
	scope := NewSingletonScope()

	factory1 := func(c ...any) (any, error) { return "bean1", nil }
	factory2 := func(c ...any) (any, error) { return "bean2", nil }

	instance1, _ := scope.Get("bean1", factory1)
	instance2, _ := scope.Get("bean2", factory2)

	if instance1 == instance2 {
		t.Error("Expected different instances for different bean IDs")
	}
}

func TestSingletonScopeRemove(t *testing.T) {
	t.Parallel()
	scope := NewSingletonScope()

	factory := func(c ...any) (any, error) { return "instance", nil }

	_, _ = scope.Get("bean1", factory)
	scope.Remove("bean1")

	// After remove, should create new instance
	newInstance, _ := scope.Get("bean1", factory)
	if newInstance == nil {
		t.Error("Expected new instance after remove")
	}
}

func TestSingletonScope_Remove(t *testing.T) {
	t.Parallel()
	scope := NewSingletonScope()
	var seq atomic.Int32
	factory := func(c ...any) (any, error) {
		n := seq.Add(1)
		return fmt.Sprintf("v%d", n), nil
	}

	v1, _ := scope.Get("b1", factory)
	scope.Remove("b1")
	v2, _ := scope.Get("b1", factory)
	if v1 == v2 {
		t.Errorf("expected different instance after remove, got both %v", v1)
	}
}

func TestSingletonScopeClear(t *testing.T) {
	t.Parallel()
	scope := NewSingletonScope()

	factory := func(c ...any) (any, error) { return "instance", nil }

	_, _ = scope.Get("bean1", factory)
	_, _ = scope.Get("bean2", factory)
	scope.Clear()

	// After clear, should create new instances
	newInstance1, _ := scope.Get("bean1", factory)
	newInstance2, _ := scope.Get("bean2", factory)

	if newInstance1 == nil || newInstance2 == nil {
		t.Error("Expected new instances after clear")
	}
}

func TestSingletonScope_Clear(t *testing.T) {
	t.Parallel()
	scope := NewSingletonScope()
	factory := func(c ...any) (any, error) { return "v", nil }

	_, _ = scope.Get("b1", factory)
	_, _ = scope.Get("b2", factory)
	scope.Clear()

	var count atomic.Int32
	factory2 := func(c ...any) (any, error) {
		count.Add(1)
		return "new", nil
	}
	_, _ = scope.Get("b1", factory2)
	_, _ = scope.Get("b2", factory2)
	if count.Load() != 2 {
		t.Errorf("expected 2 factory calls after clear, got %d", count.Load())
	}
}

func TestSingletonScopeConcurrent(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()

	callCount := int32(0)
	factory := func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return "instance", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			instance, err := s.Get("bean1", factory)
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if instance != "instance" {
				t.Errorf("Expected 'instance', got %v", instance)
			}
		}()
	}
	wg.Wait()

	// Factory should only be called once despite concurrent access
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected factory to be called once, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestSingletonScope_ConcurrentGet(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()
	var count atomic.Int32
	factory := func(c ...any) (any, error) {
		count.Add(1)
		return "bean", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := s.Get("b1", factory)
			if err != nil {
				t.Errorf("Get error: %v", err)
			}
			if got != "bean" {
				t.Errorf("expected 'bean', got %v", got)
			}
		}()
	}
	wg.Wait()
	if count.Load() != 1 {
		t.Errorf("expected factory called once, got %d", count.Load())
	}
}

// ==================== PrototypeScope 测试 ====================

func TestPrototypeScopeGet(t *testing.T) {
	t.Parallel()
	scope := NewPrototypeScope()

	callCount := int32(0)
	factory := func(c ...any) (any, error) {
		count := atomic.AddInt32(&callCount, 1)
		// Use struct to avoid string interning
		return struct{ ID int32 }{ID: count}, nil
	}

	// Each call should create new instance
	instance1, err := scope.Get("bean1", factory)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	instance2, err := scope.Get("bean1", factory)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if instance1 == instance2 {
		t.Error("Expected different instances for prototype scope")
	}

	// Factory should be called twice
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Expected factory to be called twice, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestPrototypeScope_Get(t *testing.T) {
	t.Parallel()
	scope := NewPrototypeScope()
	var count atomic.Int32
	factory := func(c ...any) (any, error) {
		count.Add(1)
		return struct{ ID int32 }{ID: count.Load()}, nil
	}

	v1, _ := scope.Get("b1", factory)
	v2, _ := scope.Get("b1", factory)
	if v1 == v2 {
		t.Error("expected different instances for prototype")
	}
	if count.Load() != 2 {
		t.Errorf("expected 2 factory calls, got %d", count.Load())
	}
}

func TestPrototypeScopeRemove(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()

	// Prototype scope remove should be a no-op
	s.Remove("bean1")
}

func TestPrototypeScope_Remove_NoOp(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()
	s.Remove("b1") // should not panic
}

func TestPrototypeScopeClear(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()

	// Prototype scope clear should be a no-op
	s.Clear()
}

func TestPrototypeScope_Clear_NoOp(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()
	s.Clear() // should not panic
}

func TestPrototypeScopeConcurrent(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()

	callCount := int32(0)
	factory := func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return "instance", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			instance, err := s.Get("bean1", factory)
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if instance != "instance" {
				t.Errorf("Expected 'instance', got %v", instance)
			}
		}()
	}
	wg.Wait()

	// Factory should be called 100 times for prototype scope
	if atomic.LoadInt32(&callCount) != 100 {
		t.Errorf("Expected factory to be called 100 times, got %d", atomic.LoadInt32(&callCount))
	}
}

// ==================== ScopeRegistry 测试 ====================

func TestScopeRegistryRegisterAndGet(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	// Get built-in scopes
	singleton := registry.Get(SingletonScope)
	if singleton == nil {
		t.Error("Expected singleton scope to exist")
	}

	prototype := registry.Get(PrototypeScope)
	if prototype == nil {
		t.Error("Expected prototype scope to exist")
	}

	// Get non-existent scope
	custom := registry.Get("custom")
	if custom != nil {
		t.Error("Expected custom scope to be nil")
	}
}

func TestScopeRegistry_GetBuiltin(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	s := registry.Get(SingletonScope)
	if s == nil {
		t.Error("expected singleton scope")
	}

	p := registry.Get(PrototypeScope)
	if p == nil {
		t.Error("expected prototype scope")
	}

	n := registry.Get("nonexistent")
	if n != nil {
		t.Error("expected nil for nonexistent scope")
	}
}

func TestScopeRegistryHas(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	if !registry.Has(SingletonScope) {
		t.Error("Expected Has to return true for singleton")
	}

	if !registry.Has(PrototypeScope) {
		t.Error("Expected Has to return true for prototype")
	}

	if registry.Has("custom") {
		t.Error("Expected Has to return false for custom")
	}
}

func TestScopeRegistry_Has(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	if !registry.Has(SingletonScope) {
		t.Error("expected Has true for singleton")
	}
	if !registry.Has(PrototypeScope) {
		t.Error("expected Has true for prototype")
	}
	if registry.Has("custom") {
		t.Error("expected Has false for custom")
	}
}

func TestScopeRegistryRegisterCustom(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	customScope := NewSingletonScope()
	registry.Register("custom", customScope)

	if !registry.Has("custom") {
		t.Error("Expected Has to return true for custom scope")
	}

	retrieved := registry.Get("custom")
	if retrieved == nil {
		t.Fatal("Expected custom scope to be retrieved")
	}

	if retrieved != customScope {
		t.Error("Expected retrieved scope to be the same as registered")
	}
}

func TestScopeRegistry_RegisterCustom(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	custom := NewSingletonScope()
	registry.Register("custom", custom)

	if !registry.Has("custom") {
		t.Error("expected Has true for custom scope")
	}
	got := registry.Get("custom")
	if got != custom {
		t.Error("expected custom scope to match")
	}
}

func TestScopeRegistryConcurrent(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			scope := registry.Get(SingletonScope)
			if scope == nil {
				t.Error("Expected singleton scope to exist")
			}
		}(i)
	}
	wg.Wait()
}

func TestScopeRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	registry := NewScopeRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.Get(SingletonScope)
			registry.Has(PrototypeScope)
		}()
	}
	wg.Wait()
}
