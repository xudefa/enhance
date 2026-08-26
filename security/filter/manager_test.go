package filter

import (
	"testing"
)

func TestNewSecurityFilterChainManager(t *testing.T) {
	t.Parallel()

	m := NewSecurityFilterChainManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestSecurityFilterChainManager_AddAndRetrieve(t *testing.T) {
	t.Parallel()

	m := NewSecurityFilterChainManager()
	chain := newMockSecurityFilterChain(true)
	m.AddSecurityFilterChain(chain)

	result := m.GetSecurityFilterChain("test")
	if result == nil {
		t.Fatal("expected non-nil chain")
	}
	if result != chain {
		t.Error("expected the same chain to be returned")
	}
}

func TestSecurityFilterChainManager_NoMatchResult(t *testing.T) {
	t.Parallel()

	m := NewSecurityFilterChainManager()
	chain := newMockSecurityFilterChain(false)
	m.AddSecurityFilterChain(chain)

	result := m.GetSecurityFilterChain("test")
	if result != nil {
		t.Error("expected nil when no chain matches")
	}
}

func TestSecurityFilterChainManager_FirstMatchWins(t *testing.T) {
	t.Parallel()

	m := NewSecurityFilterChainManager()
	chain1 := newMockSecurityFilterChain(true)
	chain2 := newMockSecurityFilterChain(true)
	m.AddSecurityFilterChain(chain1)
	m.AddSecurityFilterChain(chain2)

	result := m.GetSecurityFilterChain("test")
	if result != chain1 {
		t.Error("expected first matching chain to be returned")
	}
}

func TestSecurityFilterChainManager_GetAll(t *testing.T) {
	t.Parallel()

	m := NewSecurityFilterChainManager()
	chain1 := newMockSecurityFilterChain(true)
	chain2 := newMockSecurityFilterChain(false)
	m.AddSecurityFilterChain(chain1)
	m.AddSecurityFilterChain(chain2)

	chains := m.GetSecurityFilterChains()
	if len(chains) != 2 {
		t.Errorf("expected 2 chains, got %d", len(chains))
	}
}

func TestSecurityFilterChainManager_Empty(t *testing.T) {
	t.Parallel()

	m := NewSecurityFilterChainManager()

	result := m.GetSecurityFilterChain("test")
	if result != nil {
		t.Error("expected nil for empty manager")
	}

	chains := m.GetSecurityFilterChains()
	if len(chains) != 0 {
		t.Errorf("expected 0 chains, got %d", len(chains))
	}
}

func TestNewSimpleSecurityFilterChain(t *testing.T) {
	t.Parallel()

	matcher := MatcherFunc(func(request interface{}) bool { return true })
	f := newMockFilter(1)
	chain := NewSimpleSecurityFilterChain(matcher, f)

	if !chain.Matches("test") {
		t.Error("expected Matches to return true")
	}
}

func TestSimpleSecurityFilterChain_NoMatch(t *testing.T) {
	t.Parallel()

	matcher := MatcherFunc(func(request interface{}) bool { return false })
	f := newMockFilter(1)
	chain := NewSimpleSecurityFilterChain(matcher, f)

	if chain.Matches("test") {
		t.Error("expected Matches to return false")
	}
}

func TestSimpleSecurityFilterChain_DoFilter(t *testing.T) {
	t.Parallel()

	f := newMockFilter(1)
	matcher := MatcherFunc(func(request interface{}) bool { return true })
	chain := NewSimpleSecurityFilterChain(matcher, f)

	err := chain.DoFilter(nil, "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.executed {
		t.Error("expected filter to be executed")
	}
}

func TestSimpleSecurityFilterChain_GetFilters(t *testing.T) {
	t.Parallel()

	f1 := newMockFilter(1)
	f2 := newMockFilter(2)
	matcher := MatcherFunc(func(request interface{}) bool { return true })
	chain := NewSimpleSecurityFilterChain(matcher, f1, f2)

	filters := chain.GetFilters()
	if len(filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(filters))
	}
}

func TestMatcherFunc_Matches_True(t *testing.T) {
	t.Parallel()

	matcher := MatcherFunc(func(request interface{}) bool { return true })
	if !matcher.Matches("anything") {
		t.Error("expected true")
	}
}

func TestMatcherFunc_Matches_False(t *testing.T) {
	t.Parallel()

	matcher := MatcherFunc(func(request interface{}) bool { return false })
	if matcher.Matches("anything") {
		t.Error("expected false")
	}
}
