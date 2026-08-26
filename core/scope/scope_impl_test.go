package scope

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSingletonScope_Get_Cached(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()
	var count atomic.Int32
	factory := func(c ...any) (any, error) {
		count.Add(1)
		return "bean", nil
	}

	v1, _ := s.Get("b1", factory)
	v2, _ := s.Get("b1", factory)
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

func TestSingletonScope_Remove(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()
	var seq atomic.Int32
	factory := func(c ...any) (any, error) {
		n := seq.Add(1)
		return fmt.Sprintf("v%d", n), nil
	}

	v1, _ := s.Get("b1", factory)
	s.Remove("b1")
	v2, _ := s.Get("b1", factory)
	if v1 == v2 {
		t.Errorf("expected different instance after remove, got both %v", v1)
	}
}

func TestSingletonScope_Clear(t *testing.T) {
	t.Parallel()
	s := NewSingletonScope()
	factory := func(c ...any) (any, error) { return "v", nil }

	_, _ = s.Get("b1", factory)
	_, _ = s.Get("b2", factory)
	s.Clear()

	var count atomic.Int32
	factory2 := func(c ...any) (any, error) {
		count.Add(1)
		return "new", nil
	}
	_, _ = s.Get("b1", factory2)
	_, _ = s.Get("b2", factory2)
	if count.Load() != 2 {
		t.Errorf("expected 2 factory calls after clear, got %d", count.Load())
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
			v, err := s.Get("b1", factory)
			if err != nil {
				t.Errorf("Get error: %v", err)
			}
			if v != "bean" {
				t.Errorf("expected 'bean', got %v", v)
			}
		}()
	}
	wg.Wait()
	if count.Load() != 1 {
		t.Errorf("expected factory called once, got %d", count.Load())
	}
}

func TestPrototypeScope_Get(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()
	var count atomic.Int32
	factory := func(c ...any) (any, error) {
		count.Add(1)
		return struct{ ID int32 }{ID: count.Load()}, nil
	}

	v1, _ := s.Get("b1", factory)
	v2, _ := s.Get("b1", factory)
	if v1 == v2 {
		t.Error("expected different instances for prototype")
	}
	if count.Load() != 2 {
		t.Errorf("expected 2 factory calls, got %d", count.Load())
	}
}

func TestPrototypeScope_Remove_NoOp(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()
	s.Remove("b1") // should not panic
}

func TestPrototypeScope_Clear_NoOp(t *testing.T) {
	t.Parallel()
	s := NewPrototypeScope()
	s.Clear() // should not panic
}

func TestScopeRegistry_GetBuiltin(t *testing.T) {
	t.Parallel()
	r := NewScopeRegistry()

	s := r.Get(SingletonScope)
	if s == nil {
		t.Error("expected singleton scope")
	}

	p := r.Get(PrototypeScope)
	if p == nil {
		t.Error("expected prototype scope")
	}

	n := r.Get("nonexistent")
	if n != nil {
		t.Error("expected nil for nonexistent scope")
	}
}

func TestScopeRegistry_Has(t *testing.T) {
	t.Parallel()
	r := NewScopeRegistry()

	if !r.Has(SingletonScope) {
		t.Error("expected Has true for singleton")
	}
	if !r.Has(PrototypeScope) {
		t.Error("expected Has true for prototype")
	}
	if r.Has("custom") {
		t.Error("expected Has false for custom")
	}
}

func TestScopeRegistry_RegisterCustom(t *testing.T) {
	t.Parallel()
	r := NewScopeRegistry()

	custom := NewSingletonScope()
	r.Register("custom", custom)

	if !r.Has("custom") {
		t.Error("expected Has true for custom scope")
	}
	got := r.Get("custom")
	if got != custom {
		t.Error("expected custom scope to match")
	}
}

func TestScopeRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	r := NewScopeRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Get(SingletonScope)
			r.Has(PrototypeScope)
		}()
	}
	wg.Wait()
}
