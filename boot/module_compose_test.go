package boot

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/lifecycle"
)

func TestModuleComposer_Basic(t *testing.T) {
	t.Parallel()
	mod := ModuleDef("test").
		DependsOn("config", "database").
		Beans(
			Provide(func(c core.Container) (string, error) { return "test", nil }),
		).
		Starters().
		Hooks(
			lifecycle.OnInitFunc(func(ctx context.Context) error { return nil }),
		).
		Build()

	if mod.ModuleName() != "test" {
		t.Errorf("ModuleName() = %q, want %q", mod.ModuleName(), "test")
	}
	if len(mod.beans) != 1 {
		t.Errorf("beans count = %d, want 1", len(mod.beans))
	}
	if len(mod.hooks) != 1 {
		t.Errorf("hooks count = %d, want 1", len(mod.hooks))
	}
}

func TestModuleComposer_Conditions(t *testing.T) {
	t.Parallel()
	mod := ModuleDef("conditional").
		OnProperty("feature.enabled").
		OnPropertyEquals("mode", "production").
		OnPropertyMissing("legacy.enabled").
		Build()

	if len(mod.conditions) != 3 {
		t.Errorf("conditions count = %d, want 3", len(mod.conditions))
	}
}

func TestModuleComposer_Invoke(t *testing.T) {
	t.Parallel()
	called := false
	mod := ModuleDef("invoke-test").
		Invoke(func(c core.Container) error {
			called = true
			return nil
		}).
		Build()

	if len(mod.invokes) != 1 {
		t.Errorf("invokes count = %d, want 1", len(mod.invokes))
	}

	// Test invoke execution
	err := mod.invokes[0](nil)
	if err != nil {
		t.Errorf("invoke returned error: %v", err)
	}
	if !called {
		t.Error("invoke function was not called")
	}
}

func TestModuleGroup_Basic(t *testing.T) {
	t.Parallel()
	mod1 := ModuleDef("module1").Build()
	mod2 := ModuleDef("module2").Build()

	group := ModuleGroup("backend").
		Include(mod1, mod2).
		Build()

	if len(group) != 2 {
		t.Errorf("group modules count = %d, want 2", len(group))
	}
}

func TestModuleList(t *testing.T) {
	t.Parallel()
	mod1 := ModuleDef("a").Build()
	mod2 := ModuleDef("b").Build()

	list := ModuleList(mod1, mod2)
	if len(list) != 2 {
		t.Errorf("list count = %d, want 2", len(list))
	}
}

func TestEmptyModule(t *testing.T) {
	t.Parallel()
	mod := EmptyModule()
	if mod.ModuleName() != "" {
		t.Errorf("EmptyModule should have empty name, got %q", mod.ModuleName())
	}
	if len(mod.beans) != 0 {
		t.Errorf("EmptyModule should have no beans, got %d", len(mod.beans))
	}
}

func TestModuleFromFunc(t *testing.T) {
	t.Parallel()
	called := false
	mod := ModuleFromFunc("func-module", func(c core.Container) error {
		called = true
		return nil
	})

	if mod.ModuleName() != "func-module" {
		t.Errorf("ModuleName() = %q, want %q", mod.ModuleName(), "func-module")
	}

	err := mod.invokes[0](nil)
	if err != nil {
		t.Errorf("ModuleFromFunc invoke returned error: %v", err)
	}
	if !called {
		t.Error("ModuleFromFunc function was not called")
	}
}

func TestModuleOption(t *testing.T) {
	t.Parallel()
	mod := ModuleFromFunc("option-module",
		func(c core.Container) error { return nil },
		WithModuleConditions(condition.OnProperty("test.enabled")),
		WithModuleHooks(lifecycle.OnInitFunc(func(ctx context.Context) error { return nil })),
		WithModuleBeans(Provide(func(c core.Container) (string, error) { return "test", nil })),
	)

	if len(mod.conditions) != 1 {
		t.Errorf("conditions count = %d, want 1", len(mod.conditions))
	}
	if len(mod.hooks) != 1 {
		t.Errorf("hooks count = %d, want 1", len(mod.hooks))
	}
	if len(mod.beans) != 1 { // 1 from WithModuleBeans
		t.Errorf("beans count = %d, want 1", len(mod.beans))
	}
}

func TestLifecycleHook(t *testing.T) {
	t.Parallel()
	initCalled := false
	startCalled := false
	stopCalled := false

	hook := LifecycleHook(
		func(ctx context.Context) error { initCalled = true; return nil },
		func(ctx context.Context) error { startCalled = true; return nil },
		func(ctx context.Context) error { stopCalled = true; return nil },
	)

	ctx := context.Background()
	if err := hook.OnInit(ctx); err != nil {
		t.Errorf("OnInit error: %v", err)
	}
	if err := hook.OnStart(ctx); err != nil {
		t.Errorf("OnStart error: %v", err)
	}
	if err := hook.OnStop(ctx); err != nil {
		t.Errorf("OnStop error: %v", err)
	}

	if !initCalled || !startCalled || !stopCalled {
		t.Error("LifecycleHook callbacks were not all called")
	}
}

func TestOnInitHook(t *testing.T) {
	t.Parallel()
	called := false
	hook := OnInitHook(func(ctx context.Context) error { called = true; return nil })

	ctx := context.Background()
	if err := hook.OnInit(ctx); err != nil {
		t.Errorf("OnInit error: %v", err)
	}
	if !called {
		t.Error("OnInitHook callback was not called")
	}
}

func TestOnStartHook(t *testing.T) {
	t.Parallel()
	called := false
	hook := OnStartHook(func(ctx context.Context) error { called = true; return nil })

	ctx := context.Background()
	if err := hook.OnStart(ctx); err != nil {
		t.Errorf("OnStart error: %v", err)
	}
	if !called {
		t.Error("OnStartHook callback was not called")
	}
}

func TestOnStopHook(t *testing.T) {
	t.Parallel()
	called := false
	hook := OnStopHook(func(ctx context.Context) error { called = true; return nil })

	ctx := context.Background()
	if err := hook.OnStop(ctx); err != nil {
		t.Errorf("OnStop error: %v", err)
	}
	if !called {
		t.Error("OnStopHook callback was not called")
	}
}
