package scope

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestSingletonScopeGet(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()

	callCount := int32(0)
	factory := func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return "instance", nil
	}

	// First call should create instance
	instance1, err := s.Get("bean1", factory)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Second call should return cached instance
	instance2, err := s.Get("bean1", factory)
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

func TestSingletonScopeDifferentBeans(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()

	factory1 := func(c ...any) (any, error) { return "bean1", nil }
	factory2 := func(c ...any) (any, error) { return "bean2", nil }

	instance1, _ := s.Get("bean1", factory1)
	instance2, _ := s.Get("bean2", factory2)

	if instance1 == instance2 {
		t.Error("Expected different instances for different bean IDs")
	}
}

func TestSingletonScopeRemove(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()

	factory := func(c ...any) (any, error) { return "instance", nil }

	_, _ = s.Get("bean1", factory)
	s.Remove("bean1")

	// After remove, should create new instance
	newInstance, _ := s.Get("bean1", factory)
	if newInstance == nil {
		t.Error("Expected new instance after remove")
	}
}

func TestSingletonScopeClear(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()

	factory := func(c ...any) (any, error) { return "instance", nil }

	_, _ = s.Get("bean1", factory)
	_, _ = s.Get("bean2", factory)
	s.Clear()

	// After clear, should create new instances
	newInstance1, _ := s.Get("bean1", factory)
	newInstance2, _ := s.Get("bean2", factory)

	if newInstance1 == nil || newInstance2 == nil {
		t.Error("Expected new instances after clear")
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

func TestPrototypeScopeGet(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()

	callCount := int32(0)
	factory := func(c ...any) (any, error) {
		count := atomic.AddInt32(&callCount, 1)
		// Use struct to avoid string interning
		return struct{ ID int32 }{ID: count}, nil
	}

	// Each call should create new instance
	instance1, err := s.Get("bean1", factory)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	instance2, err := s.Get("bean1", factory)
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

func TestPrototypeScopeRemove(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()

	// Prototype scope remove should be a no-op
	s.Remove("bean1")
}

func TestPrototypeScopeClear(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()

	// Prototype scope clear should be a no-op
	s.Clear()
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
