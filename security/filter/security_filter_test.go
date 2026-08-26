package filter

import (
	"context"
	"testing"
)

func TestNewSecurityFilter(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	sf := NewSecurityFilter(manager)
	if sf == nil {
		t.Fatal("expected non-nil security filter")
	}
}

func TestNewSecurityFilterWithOrder(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	sf := NewSecurityFilterWithOrder(manager, 10)
	if sf == nil {
		t.Fatal("expected non-nil security filter")
	}
	if sf.Order() != 10 {
		t.Errorf("expected order 10, got %d", sf.Order())
	}
}

func TestSecurityFilter_DoFilter_WithMatchedChain(t *testing.T) {
	t.Parallel()

	f := newMockFilter(1)
	chain := newMockSecurityFilterChain(true)
	chain.filters = []Filter{f}

	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain)

	sf := NewSecurityFilter(manager)
	err := sf.DoFilter(context.Background(), "test", nil, NewDefaultFilterChain())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.executed {
		t.Error("expected filter from matched chain to be executed")
	}
}

func TestSecurityFilter_DoFilter_NoMatchingChain(t *testing.T) {
	t.Parallel()

	f := newMockFilter(1)
	chain := newMockSecurityFilterChain(false)
	chain.filters = []Filter{f}

	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain)

	nextFilter := newMockFilter(2)
	nextChain := NewDefaultFilterChainWithFilters(nextFilter)

	sf := NewSecurityFilter(manager)
	err := sf.DoFilter(context.Background(), "test", nil, nextChain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.executed {
		t.Error("expected matched chain filter not to be executed")
	}
	if !nextFilter.executed {
		t.Error("expected next chain filter to be executed")
	}
}

func TestSecurityFilter_DoFilter_EmptyManager(t *testing.T) {
	t.Parallel()

	nextFilter := newMockFilter(1)
	nextChain := NewDefaultFilterChainWithFilters(nextFilter)

	manager := NewSecurityFilterChainManager()
	sf := NewSecurityFilter(manager)

	err := sf.DoFilter(context.Background(), "test", nil, nextChain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nextFilter.executed {
		t.Error("expected next filter to be executed when no chain matches")
	}
}

func TestSecurityFilter_SetOrder(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	sf := NewSecurityFilter(manager)

	sf.(*securityFilterImpl).SetOrder(25)
	if sf.Order() != 25 {
		t.Errorf("expected order 25, got %d", sf.Order())
	}
}

func TestSecurityFilter_ManagerAccessor(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	sf := NewSecurityFilter(manager)

	if sf.(*securityFilterImpl).Manager() != manager {
		t.Error("expected manager to be the same")
	}
}

func TestSecurityFilter_Order_Default(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	sf := NewSecurityFilter(manager)

	if sf.Order() != 0 {
		t.Errorf("expected default order 0, got %d", sf.Order())
	}
}

func TestSecurityFilter_DoFilter_ContextPropagation(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), "testKey", "testVal")
	f := newMockFilter(1)
	chain := newMockSecurityFilterChain(true)
	chain.filters = []Filter{f}

	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain)

	sf := NewSecurityFilter(manager)
	sf.DoFilter(ctx, "test", nil, NewDefaultFilterChain())

	if f.ctx == nil || f.ctx.Value("testKey") != "testVal" {
		t.Error("expected context to be propagated to matched chain filter")
	}
}
