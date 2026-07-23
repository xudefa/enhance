package filter

import (
	"context"
	"testing"
)

// mockFilter 测试用过滤器实现。
type mockFilter struct {
	order    int
	executed bool
	ctx      context.Context
}

func newMockFilter(order int) *mockFilter {
	return &mockFilter{order: order}
}

func (f *mockFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain FilterChain) error {
	f.executed = true
	if ctx, ok := ctx.(context.Context); ok {
		f.ctx = ctx
	}
	return chain.DoFilter(ctx, request, response)
}

func (f *mockFilter) Order() int {
	return f.order
}

// mockSecurityFilterChain 测试用安全过滤器链实现。
type mockSecurityFilterChain struct {
	matcher MatcherFunc
	filters []Filter
	matched bool
}

func newMockSecurityFilterChain(match bool) *mockSecurityFilterChain {
	return &mockSecurityFilterChain{
		matcher: MatcherFunc(func(request interface{}) bool { return match }),
		filters: make([]Filter, 0),
	}
}

func (c *mockSecurityFilterChain) Matches(request interface{}) bool {
	c.matched = c.matcher.Matches(request)
	return c.matched
}

func (c *mockSecurityFilterChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	chain := NewDefaultFilterChainWithFilters(c.filters...)
	return chain.DoFilter(ctx, request, response)
}

func (c *mockSecurityFilterChain) GetFilters() []Filter {
	return c.filters
}

func TestDefaultFilterChain_DoFilter(t *testing.T) {
	t.Parallel()

	filter1 := newMockFilter(1)
	filter2 := newMockFilter(2)
	chain := NewDefaultFilterChainWithFilters(filter1, filter2)

	err := chain.DoFilter(context.Background(), nil, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !filter1.executed {
		t.Error("filter1 was not executed")
	}
	if !filter2.executed {
		t.Error("filter2 was not executed")
	}
}

func TestDefaultFilterChain_AddFilter(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChain()
	if len(chain.GetFilters()) != 0 {
		t.Errorf("expected 0 filters, got %d", len(chain.GetFilters()))
	}

	filter := newMockFilter(1)
	chain.AddFilter(filter)

	if len(chain.GetFilters()) != 1 {
		t.Errorf("expected 1 filter, got %d", len(chain.GetFilters()))
	}
}

func TestDefaultFilterChain_SortFilters(t *testing.T) {
	t.Parallel()

	filter1 := newMockFilter(3)
	filter2 := newMockFilter(1)
	filter3 := newMockFilter(2)
	chain := NewDefaultFilterChainWithFilters(filter1, filter2, filter3)

	chain.(*defaultFilterChain).SortFilters()

	filters := chain.GetFilters()
	if filters[0].Order() != 1 || filters[1].Order() != 2 || filters[2].Order() != 3 {
		t.Error("filters are not sorted correctly")
	}
}

func TestDefaultFilterChain_EmptyChain(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChain()
	err := chain.DoFilter(context.Background(), nil, nil)
	if err != nil {
		t.Errorf("unexpected error on empty chain: %v", err)
	}
}

func TestSecurityFilterChainManager_GetSecurityFilterChain(t *testing.T) {
	t.Parallel()

	chain1 := newMockSecurityFilterChain(true)
	chain2 := newMockSecurityFilterChain(false)

	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain1)
	manager.AddSecurityFilterChain(chain2)

	result := manager.GetSecurityFilterChain("test")
	if result == nil {
		t.Error("expected a security filter chain, got nil")
	}

	if result != chain1 {
		t.Error("expected chain1 to be returned")
	}
}

func TestSecurityFilterChainManager_NoMatch(t *testing.T) {
	t.Parallel()

	chain1 := newMockSecurityFilterChain(false)
	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain1)

	result := manager.GetSecurityFilterChain("test")
	if result != nil {
		t.Error("expected nil, got a security filter chain")
	}
}

func TestSecurityFilterChainManager_GetSecurityFilterChains(t *testing.T) {
	t.Parallel()

	chain1 := newMockSecurityFilterChain(true)
	chain2 := newMockSecurityFilterChain(false)

	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain1)
	manager.AddSecurityFilterChain(chain2)

	chains := manager.GetSecurityFilterChains()
	if len(chains) != 2 {
		t.Errorf("expected 2 chains, got %d", len(chains))
	}
}

func TestSecurityFilter_DoFilter_WithMatchingChain(t *testing.T) {
	t.Parallel()

	filter1 := newMockFilter(1)
	chain := newMockSecurityFilterChain(true)
	chain.filters = []Filter{filter1}

	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain)

	secFilter := NewSecurityFilter(manager)
	err := secFilter.DoFilter(context.Background(), "test", nil, NewDefaultFilterChain())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !filter1.executed {
		t.Error("filter1 was not executed")
	}
}

func TestSecurityFilter_DoFilter_WithoutMatchingChain(t *testing.T) {
	t.Parallel()

	filter1 := newMockFilter(1)
	chain := newMockSecurityFilterChain(false)
	chain.filters = []Filter{filter1}

	manager := NewSecurityFilterChainManager()
	manager.AddSecurityFilterChain(chain)

	nextFilter := newMockFilter(2)
	nextChain := NewDefaultFilterChainWithFilters(nextFilter)

	secFilter := NewSecurityFilter(manager)
	err := secFilter.DoFilter(context.Background(), "test", nil, nextChain)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if filter1.executed {
		t.Error("filter1 should not be executed when no chain matches")
	}
	if !nextFilter.executed {
		t.Error("next filter should be executed when no chain matches")
	}
}

func TestSimpleSecurityFilterChain(t *testing.T) {
	t.Parallel()

	filter1 := newMockFilter(1)
	matcher := MatcherFunc(func(request interface{}) bool { return true })
	chain := NewSimpleSecurityFilterChain(matcher, filter1)

	if !chain.Matches("test") {
		t.Error("expected matches to return true")
	}

	err := chain.DoFilter(context.Background(), "test", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !filter1.executed {
		t.Error("filter1 was not executed")
	}
}

func TestSecurityFilter_Order(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	secFilter := NewSecurityFilterWithOrder(manager, 10)

	if secFilter.Order() != 10 {
		t.Errorf("expected order 10, got %d", secFilter.Order())
	}

	secFilter.(*securityFilterImpl).SetOrder(20)
	if secFilter.Order() != 20 {
		t.Errorf("expected order 20, got %d", secFilter.Order())
	}
}

func TestSecurityFilter_Manager(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	secFilter := NewSecurityFilter(manager)

	if secFilter.(*securityFilterImpl).Manager() != manager {
		t.Error("expected manager to be the same")
	}
}

func TestDefaultFilterChain_String(t *testing.T) {
	t.Parallel()

	chain := NewDefaultFilterChainWithFilters(newMockFilter(1), newMockFilter(2))
	str := chain.(*defaultFilterChain).String()
	if str != "DefaultFilterChain{filters=2}" {
		t.Errorf("unexpected string: %s", str)
	}
}

func TestSecurityFilterChainManager_EmptyManager(t *testing.T) {
	t.Parallel()

	manager := NewSecurityFilterChainManager()
	result := manager.GetSecurityFilterChain("test")
	if result != nil {
		t.Error("expected nil for empty manager")
	}

	chains := manager.GetSecurityFilterChains()
	if len(chains) != 0 {
		t.Errorf("expected 0 chains, got %d", len(chains))
	}
}

func TestMatcherFunc_Matches(t *testing.T) {
	t.Parallel()

	matcher := MatcherFunc(func(request interface{}) bool {
		return request == "hello"
	})

	if !matcher.Matches("hello") {
		t.Error("expected Matches to return true for 'hello'")
	}
	if matcher.Matches("world") {
		t.Error("expected Matches to return false for 'world'")
	}
}
