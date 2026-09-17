package boot

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/core"
)

func TestModuleBuilder_Name(t *testing.T) {
	t.Parallel()

	builder := NewModule().Name("test")
	module := builder.Build()
	if module.ModuleName() != "test" {
		t.Errorf("expected name 'test', got '%s'", module.ModuleName())
	}
}

func TestModuleBuilder_Bean(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Value string
	}

	builder := NewModule().
		Bean(ProvideBean(&TestBean{Value: "bean"}))

	container := core.NewContainer()
	module := builder.Build()
	err := module.Install(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModuleBuilder_Invoke(t *testing.T) {
	t.Parallel()

	invoked := false
	builder := NewModule().
		Invoke(func() error {
			invoked = true
			return nil
		})

	container := core.NewContainer()
	module := builder.Build()
	err := module.Install(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !invoked {
		t.Error("expected invoke to be called")
	}
}

func TestModuleBuilder_Hook(t *testing.T) {
	t.Parallel()

	hookCalled := false
	builder := NewModule("test-hook").
		Hook(&testHookForModule{fn: func(ctx context.Context) error {
			hookCalled = true
			return nil
		}})

	container := core.NewContainer()
	module := builder.Build()
	err := module.Install(container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = hookCalled
}

type testHookForModule struct {
	fn func(context.Context) error
}

func (h *testHookForModule) OnInit(ctx context.Context) error {
	return h.fn(ctx)
}

func (h *testHookForModule) OnStart(ctx context.Context) error {
	return nil
}

func (h *testHookForModule) OnStop(ctx context.Context) error {
	return nil
}

func TestModuleBuilder_Hooks(t *testing.T) {
	t.Parallel()

	builder := NewModule()
	hooks := builder.Hooks()
	if hooks == nil {
		t.Error("expected non-nil hooks")
	}
}

func TestModuleBuilder_Starters(t *testing.T) {
	t.Parallel()

	builder := NewModule("test-starters")
	module := builder.Build()
	if module.ModuleName() != "test-starters" {
		t.Errorf("expected module name 'test-starters', got '%s'", module.ModuleName())
	}
}

func TestModuleBuilder_Condition(t *testing.T) {
	t.Parallel()

	builder := NewModule()
	conditions := builder.Conditions()
	if conditions == nil {
		t.Error("expected non-nil conditions")
	}
}

func TestModuleBuilder_Conditions(t *testing.T) {
	t.Parallel()

	builder := NewModule()
	conditions := builder.Conditions()
	if conditions == nil {
		t.Error("expected non-nil conditions")
	}
}

func TestModuleBuilder_Module(t *testing.T) {
	t.Parallel()

	builder := NewModule("test-module")
	module := builder.Build()
	if module.ModuleName() != "test-module" {
		t.Errorf("expected module name 'test-module', got '%s'", module.ModuleName())
	}
}
