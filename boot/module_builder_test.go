package boot

import (
	"context"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
)

func TestModuleBuilder_Starter(t *testing.T) {
	t.Parallel()

	b := NewModule()
	s := newMockStarter("test")
	b.Starter(s)

	if len(b.starters) != 1 {
		t.Errorf("expected 1 starter, got %d", len(b.starters))
	}
}

func TestModuleBuilder_Build(t *testing.T) {
	t.Parallel()

	b := NewModule().
		Name("test").
		Bean(Provide(func(c core.Container) (string, error) {
			return "test", nil
		})).
		Starter(newMockStarter("test")).
		Condition(condition.OnProperty("app.name", "test"))

	mod := b.Build()

	if mod.moduleName != "test" {
		t.Errorf("expected module name 'test', got %v", mod.moduleName)
	}
	if len(mod.beans) != 1 {
		t.Errorf("expected 1 bean, got %d", len(mod.beans))
	}
	if len(mod.starters) != 1 {
		t.Errorf("expected 1 starter, got %d", len(mod.starters))
	}
	if len(mod.conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(mod.conditions))
	}
}

func TestModuleBuilder_Install(t *testing.T) {
	t.Parallel()

	c := core.NewContainer()
	b := NewModule().
		Bean(Provide(func(c core.Container) (string, error) {
			return "installed", nil
		}))

	if err := b.Install(c); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	beans, err := c.Get(reflect.TypeOf(""))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(beans) != 1 {
		t.Errorf("expected 1 bean after install, got %d", len(beans))
	}
}

func TestModuleBuilder_Invoke_NilFunction(t *testing.T) {
	t.Parallel()

	c := core.NewContainer()
	b := NewModule().Invoke(nil)

	err := b.Install(c)
	if err == nil {
		t.Fatal("expected error for nil invoke function")
	}
}

func TestNewModule_WithStringArg(t *testing.T) {
	t.Parallel()

	b := NewModule("test-module")
	if b.name != "test-module" {
		t.Errorf("expected name 'test-module', got %v", b.name)
	}
}

func TestNewModule_WithMultipleArgs(t *testing.T) {
	t.Parallel()

	s := newMockStarter("test")
	b := NewModule("test", s)

	if b.name != "test" {
		t.Errorf("expected name 'test', got %v", b.name)
	}
	if len(b.starters) != 1 {
		t.Errorf("expected 1 starter, got %d", len(b.starters))
	}
}

func TestModuleBuilder_ChainCall(t *testing.T) {
	t.Parallel()

	b := NewModule().
		Name("chain").
		Bean(Provide(func(c core.Container) (int, error) {
			return 42, nil
		})).
		Starter(newMockStarter("starter")).
		Condition(condition.OnProperty("app.name", "chain")).
		Hook(&testHookForModuleBuilder{name: "hook"})

	mod := b.Build()

	if mod.moduleName != "chain" {
		t.Errorf("expected 'chain', got %v", mod.moduleName)
	}
	if len(mod.beans) != 1 {
		t.Errorf("expected 1 bean, got %d", len(mod.beans))
	}
	if len(mod.starters) != 1 {
		t.Errorf("expected 1 starter, got %d", len(mod.starters))
	}
	if len(mod.conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(mod.conditions))
	}
	if len(mod.hooks) != 1 {
		t.Errorf("expected 1 hook, got %d", len(mod.hooks))
	}
}

// testHookForModuleBuilder 用于测试的简单 Hook 实现
type testHookForModuleBuilder struct {
	name string
}

func (h *testHookForModuleBuilder) Name() string {
	return h.name
}

func (h *testHookForModuleBuilder) OnInit(ctx context.Context) error {
	return nil
}

func (h *testHookForModuleBuilder) OnStart(ctx context.Context) error {
	return nil
}

func (h *testHookForModuleBuilder) OnStop(ctx context.Context) error {
	return nil
}
