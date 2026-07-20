package context

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/lifecycle"
)

func TestNewApplicationContext(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	if ctx.Container() != container {
		t.Fatal("container mismatch")
	}
	if ctx.Environment() != env {
		t.Fatal("environment mismatch")
	}
	if ctx.Lifecycle().GetPhase() != lifecycle.PhaseInit {
		t.Fatal("expected init phase")
	}
	if ctx.IsRunning() {
		t.Fatal("should not be running initially")
	}
}

func TestApplicationContext_StartStop(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	if err := ctx.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if !ctx.IsRunning() {
		t.Fatal("should be running after start")
	}

	if err := ctx.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if ctx.IsRunning() {
		t.Fatal("should not be running after stop")
	}
}

func TestLifecycle_BasicTransitions(t *testing.T) {
	t.Parallel()
	lm := lifecycle.NewLifecycleManager()

	if lm.GetPhase() != lifecycle.PhaseInit {
		t.Fatal("expected init phase")
	}

	if err := lm.SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("set phase failed: %v", err)
	}
	if lm.GetPhase() != lifecycle.PhaseRunning {
		t.Fatal("expected running phase")
	}
}

func TestLifecycle_PhaseListener(t *testing.T) {
	t.Parallel()
	lm := lifecycle.NewLifecycleManager()
	var oldPhase, newPhase lifecycle.ApplicationPhase

	lm.AddListener(&testListener{
		fn: func(old, new lifecycle.ApplicationPhase) error {
			oldPhase = old
			newPhase = new
			return nil
		},
	})

	if err := lm.SetPhase(lifecycle.PhaseRunning); err != nil {
		t.Fatalf("set phase failed: %v", err)
	}

	if oldPhase != lifecycle.PhaseInit {
		t.Fatal("expected old phase to be init")
	}
	if newPhase != lifecycle.PhaseRunning {
		t.Fatal("expected new phase to be running")
	}
}

type testListener struct {
	fn func(lifecycle.ApplicationPhase, lifecycle.ApplicationPhase) error
}

func (l *testListener) OnPhaseChange(old, new lifecycle.ApplicationPhase) error {
	return l.fn(old, new)
}

func TestApplicationContext_RefreshScopeManager(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	mgr := ctx.RefreshScopeManager()
	if mgr == nil {
		t.Error("expected RefreshScopeManager to be initialized")
	}

	if mgr.Metrics() == nil {
		t.Error("expected RefreshScopeManager to have metrics")
	}
}
