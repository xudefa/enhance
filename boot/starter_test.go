package boot

import (
	"sync/atomic"
	"testing"

	"github.com/xudefa/enhance/condition"
)

type mockStarter struct {
	name         string
	dependencies []string
	configured   atomic.Bool
	started      atomic.Bool
	stopped      atomic.Bool
}

func newMockStarter(name string, deps ...string) *mockStarter {
	return &mockStarter{name: name, dependencies: deps}
}

func (m *mockStarter) Name() string                           { return m.name }
func (m *mockStarter) Dependencies() []string                 { return m.dependencies }
func (m *mockStarter) Configure(ctx ApplicationContext) error { m.configured.Store(true); return nil }
func (m *mockStarter) Start(ctx ApplicationContext) error     { m.started.Store(true); return nil }
func (m *mockStarter) Stop(ctx ApplicationContext) error      { m.stopped.Store(true); return nil }
func (m *mockStarter) GetCondition() condition.Condition      { return nil }

func TestStarterRegistry_Add(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	s := newMockStarter("test")

	registry.Register(s)
	if len(registry.GetAll()) != 1 {
		t.Fatal("expected 1 starter in registry")
	}
}

func TestStarterRegistry_GetAll(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("s1"))
	registry.Register(newMockStarter("s2"))

	all := registry.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 starters, got %d", len(all))
	}
}

func TestStarterRegistry_Get(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("s1"))
	registry.Register(newMockStarter("s2"))

	s := registry.Get("s1")
	if s == nil {
		t.Fatal("expected to find starter 's1'")
	}
	if s.Name() != "s1" {
		t.Fatalf("expected name 's1', got '%s'", s.Name())
	}

	missing := registry.Get("nonexistent")
	if missing != nil {
		t.Fatal("expected nil for nonexistent starter")
	}
}

func TestStarterRegistry_GetOrdered(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()
	registry.Register(newMockStarter("s1", "s2"))
	registry.Register(newMockStarter("s2"))

	ordered := registry.GetOrdered()
	if len(ordered) != 2 {
		t.Fatalf("expected 2 ordered starters, got %d", len(ordered))
	}
}

func TestStarterRegistry_Empty(t *testing.T) {
	t.Parallel()
	registry := NewStarterRegistry()

	if len(registry.GetAll()) != 0 {
		t.Error("expected empty registry")
	}
	if len(registry.GetOrdered()) != 0 {
		t.Error("expected empty ordered starters")
	}
}

func TestRegisterStarterGlobal(t *testing.T) {
	orig := globalStarterRegistry.Load()
	testReg := newStarterRegistryImpl()
	globalStarterRegistry.Store(testReg)
	defer func() { globalStarterRegistry.Store(orig) }()

	RegisterStarter(newMockStarter("global"))
	if len(testReg.GetAll()) != 1 {
		t.Fatal("expected 1 starter registered globally")
	}
}

func TestGlobalStarterRegistry(t *testing.T) {
	t.Parallel()
	registry := GlobalStarterRegistry()
	if registry == nil {
		t.Fatal("expected non-nil global registry")
	}
}
