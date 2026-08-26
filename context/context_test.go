package context

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/event"
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

func TestApplicationContext_HasProperty(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	if ctx.HasProperty("nonexistent") {
		t.Error("expected property to not exist")
	}
}

func TestApplicationContext_GetProperty(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	_, ok := ctx.GetProperty("nonexistent")
	if ok {
		t.Error("expected property to not exist")
	}
}

func TestApplicationContext_EventBus(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	bus := ctx.EventBus()
	if bus == nil {
		t.Fatal("expected EventBus to be initialized")
	}

	publisher := ctx.EventPublisher()
	if publisher == nil {
		t.Fatal("expected EventPublisher to be initialized")
	}
}

func TestApplicationContext_AsyncEventPublisher(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	asyncPublisher := ctx.AsyncEventPublisher()
	if asyncPublisher == nil {
		t.Fatal("expected AsyncEventPublisher to be initialized")
	}
}

func TestApplicationContext_Register(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	type TestService struct {
		Name string
	}

	err := ctx.Register(reflect.TypeOf(&TestService{}),
		core.WithName[*TestService]("testService"),
		core.WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
}

func TestApplicationContext_GetByType(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	type TestService struct {
		Name string
	}

	err := ctx.Register(reflect.TypeOf(&TestService{}),
		core.WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	instance, err := ctx.GetByType(reflect.TypeOf(&TestService{}))
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}

	svc, ok := instance.(*TestService)
	if !ok {
		t.Fatal("expected *TestService")
	}
	if svc.Name != "test" {
		t.Errorf("expected name 'test', got %s", svc.Name)
	}
}

func TestApplicationContext_GetByType_NotFound(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	type MissingService struct{}

	_, err := ctx.GetByType(reflect.TypeOf(&MissingService{}))
	if err == nil {
		t.Error("expected error for missing bean")
	}
}

func TestApplicationContext_Invoke_NoArgs(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	called := false
	err := ctx.Invoke(func() {
		called = true
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if !called {
		t.Error("function should have been called")
	}
}

func TestApplicationContext_Invoke_WithArgs(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	type TestService struct {
		Name string
	}

	err := ctx.Register(reflect.TypeOf(&TestService{}),
		core.WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "injected"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	var receivedName string
	err = ctx.Invoke(func(svc *TestService) {
		receivedName = svc.Name
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if receivedName != "injected" {
		t.Errorf("expected 'injected', got %s", receivedName)
	}
}

func TestApplicationContext_Invoke_InvalidFunction(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	err := ctx.Invoke("not a function")
	if err == nil {
		t.Error("expected error for non-function")
	}
}

func TestApplicationContext_Invoke_WithReturn(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	result := 0
	err := ctx.Invoke(func() int {
		result = 42
		return 0
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestApplicationContext_Invoke_WithError(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	err := ctx.Invoke(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestApplicationContext_PublishEvents(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()
	ctx := NewApplicationContext(container, env)

	var receivedCount int
	ctx.EventBus().Subscribe(event.EventApplicationStarted, func(evt event.ApplicationEvent) {
		receivedCount++
	})
	ctx.EventBus().Subscribe(event.EventApplicationReady, func(evt event.ApplicationEvent) {
		receivedCount++
	})

	if err := ctx.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if receivedCount < 2 {
		t.Errorf("expected at least 2 events, got %d", receivedCount)
	}
}
