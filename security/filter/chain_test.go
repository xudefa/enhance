package filter

import (
	"context"
	"testing"
)

func TestNewDefaultFilterChain(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChain()
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}
	if len(chain.GetFilters()) != 0 {
		t.Errorf("expected 0 filters, got %d", len(chain.GetFilters()))
	}
}

func TestNewDefaultFilterChainWithFilters(t *testing.T) {
	t.Parallel()

	f1 := newMockFilter(2)
	f2 := newMockFilter(1)
	chain := NewDefaultFilterChainWithFilters(f1, f2)

	if len(chain.GetFilters()) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(chain.GetFilters()))
	}

	filters := chain.GetFilters()
	if filters[0].Order() != 1 || filters[1].Order() != 2 {
		t.Error("expected filters to be sorted by order")
	}
}

func TestDefaultFilterChain_DoFilter_ExecutesAll(t *testing.T) {
	t.Parallel()

	f1 := newMockFilter(1)
	f2 := newMockFilter(2)
	f3 := newMockFilter(3)
	chain := NewDefaultFilterChainWithFilters(f1, f2, f3)

	err := chain.DoFilter(context.Background(), "req", "resp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f1.executed || !f2.executed || !f3.executed {
		t.Error("expected all filters to be executed")
	}
}

func TestDefaultFilterChain_AddFilter_ThenGet(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChain()
	f := newMockFilter(5)
	chain.AddFilter(f)

	filters := chain.GetFilters()
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	if filters[0].Order() != 5 {
		t.Errorf("expected order 5, got %d", filters[0].Order())
	}
}

func TestDefaultFilterChain_GetFilters_ReturnsCopy(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChainWithFilters(newMockFilter(1))
	filters := chain.GetFilters()
	filters = append(filters, newMockFilter(2))

	if len(chain.GetFilters()) != 1 {
		t.Error("expected original chain to not be modified")
	}
}

func TestDefaultFilterChain_SortFilters_ByOrder(t *testing.T) {
	t.Parallel()

	f1 := newMockFilter(10)
	f2 := newMockFilter(1)
	f3 := newMockFilter(5)
	chain := NewDefaultFilterChainWithFilters(f1, f2, f3)

	chain.(*defaultFilterChain).SortFilters()

	filters := chain.GetFilters()
	if filters[0].Order() != 1 || filters[1].Order() != 5 || filters[2].Order() != 10 {
		t.Error("expected filters sorted by order")
	}
}

func TestDefaultFilterChain_StringRep(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChainWithFilters(newMockFilter(1), newMockFilter(2))
	str := chain.(*defaultFilterChain).String()
	if str != "DefaultFilterChain{filters=2}" {
		t.Errorf("expected 'DefaultFilterChain{filters=2}', got %s", str)
	}
}

func TestVirtualFilterChain_DoFilter(t *testing.T) {
	t.Parallel()

	f1 := newMockFilter(1)
	chain := NewDefaultFilterChainWithFilters(f1)

	err := chain.DoFilter(context.Background(), "req", "resp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f1.executed {
		t.Error("expected filter to be executed via virtual chain")
	}
}

func TestVirtualFilterChain_AddFilter_NoOp(t *testing.T) {
	t.Parallel()

	vfc := &virtualFilterChain{
		chain: &defaultFilterChain{},
		index: 0,
	}
	vfc.AddFilter(newMockFilter(1))

	if len(vfc.GetFilters()) != 0 {
		t.Error("expected GetFilters to return nil for virtual chain")
	}
}

func TestVirtualFilterChain_GetFilters_Nil(t *testing.T) {
	t.Parallel()

	vfc := &virtualFilterChain{
		chain: &defaultFilterChain{},
		index: 0,
	}
	if vfc.GetFilters() != nil {
		t.Error("expected nil from virtual chain GetFilters")
	}
}

func TestDefaultFilterChain_ContextPropagation(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), "key", "value")
	f := newMockFilter(1)
	chain := NewDefaultFilterChainWithFilters(f)

	chain.DoFilter(ctx, nil, nil)

	if f.ctx == nil || f.ctx.Value("key") != "value" {
		t.Error("expected context to be propagated")
	}
}

func TestDefaultFilterChain_EmptyChain_DoFilter(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChain()
	err := chain.DoFilter(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error on empty chain: %v", err)
	}
}
