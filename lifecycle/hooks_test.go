package lifecycle

import (
	"context"
	"sync"
	"testing"
)

func TestHookFunc(t *testing.T) {
	t.Parallel()
	t.Run("OnInit called", func(t *testing.T) {
		called := false
		hook := OnInitFunc(func(ctx context.Context) error {
			called = true
			return nil
		})
		_ = hook.OnInit(context.Background())
		if !called {
			t.Error("expected OnInit to be called")
		}
	})

	t.Run("OnStart called", func(t *testing.T) {
		called := false
		hook := OnStartFunc(func(ctx context.Context) error {
			called = true
			return nil
		})
		_ = hook.OnStart(context.Background())
		if !called {
			t.Error("expected OnStart to be called")
		}
	})

	t.Run("OnStop called", func(t *testing.T) {
		called := false
		hook := OnStopFunc(func(ctx context.Context) error {
			called = true
			return nil
		})
		_ = hook.OnStop(context.Background())
		if !called {
			t.Error("expected OnStop to be called")
		}
	})

	t.Run("nil functions do not panic", func(t *testing.T) {
		hook := HookFunc{}
		if err := hook.OnInit(context.Background()); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if err := hook.OnStart(context.Background()); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if err := hook.OnStop(context.Background()); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})
}

func TestHookRegistry(t *testing.T) {
	t.Parallel()
	t.Run("Register and execute hooks in order", func(t *testing.T) {
		registry := NewHookRegistry()
		var order []string

		registry.Register(OnInitFunc(func(ctx context.Context) error {
			order = append(order, "init1")
			return nil
		}))
		registry.Register(OnInitFunc(func(ctx context.Context) error {
			order = append(order, "init2")
			return nil
		}))
		registry.Register(OnStartFunc(func(ctx context.Context) error {
			order = append(order, "start1")
			return nil
		}))
		registry.Register(OnStopFunc(func(ctx context.Context) error {
			order = append(order, "stop1")
			return nil
		}))

		ctx := context.Background()
		if err := registry.InitAll(ctx); err != nil {
			t.Fatalf("InitAll error: %v", err)
		}
		if err := registry.StartAll(ctx); err != nil {
			t.Fatalf("StartAll error: %v", err)
		}
		if err := registry.StopAll(ctx); err != nil {
			t.Fatalf("StopAll error: %v", err)
		}

		expected := []string{"init1", "init2", "start1", "stop1"}
		if len(order) != len(expected) {
			t.Fatalf("expected %d calls, got %d", len(expected), len(order))
		}
		for i, exp := range expected {
			if order[i] != exp {
				t.Errorf("expected %q at position %d, got %q", exp, i, order[i])
			}
		}
	})

	t.Run("StopAll executes in reverse order", func(t *testing.T) {
		registry := NewHookRegistry()
		var order []string

		registry.Register(OnStopFunc(func(ctx context.Context) error {
			order = append(order, "stop1")
			return nil
		}))
		registry.Register(OnStopFunc(func(ctx context.Context) error {
			order = append(order, "stop2")
			return nil
		}))
		registry.Register(OnStopFunc(func(ctx context.Context) error {
			order = append(order, "stop3")
			return nil
		}))

		_ = registry.StopAll(context.Background())

		expected := []string{"stop3", "stop2", "stop1"}
		if len(order) != len(expected) {
			t.Fatalf("expected %d calls, got %d", len(expected), len(order))
		}
		for i, exp := range expected {
			if order[i] != exp {
				t.Errorf("expected %q at position %d, got %q", exp, i, order[i])
			}
		}
	})

	t.Run("InitAll stops on first error", func(t *testing.T) {
		registry := NewHookRegistry()
		registry.Register(OnInitFunc(func(ctx context.Context) error {
			return nil
		}))
		registry.Register(OnInitFunc(func(ctx context.Context) error {
			return context.Canceled
		}))
		registry.Register(OnInitFunc(func(ctx context.Context) error {
			t.Error("third hook should not be called after error")
			return nil
		}))

		err := registry.InitAll(context.Background())
		if err == nil {
			t.Fatal("expected error from InitAll")
		}
	})

	t.Run("RegisterFunc convenience method", func(t *testing.T) {
		registry := NewHookRegistry()
		var initCalled, startCalled, stopCalled bool

		registry.RegisterFunc(
			func(ctx context.Context) error { initCalled = true; return nil },
			func(ctx context.Context) error { startCalled = true; return nil },
			func(ctx context.Context) error { stopCalled = true; return nil },
		)

		ctx := context.Background()
		_ = registry.InitAll(ctx)
		_ = registry.StartAll(ctx)
		_ = registry.StopAll(ctx)

		if !initCalled || !startCalled || !stopCalled {
			t.Error("expected all hooks to be called")
		}
	})

	t.Run("Count returns correct number", func(t *testing.T) {
		registry := NewHookRegistry()
		if registry.Count() != 0 {
			t.Errorf("expected 0, got %d", registry.Count())
		}
		registry.Register(OnInitFunc(nil))
		if registry.Count() != 1 {
			t.Errorf("expected 1, got %d", registry.Count())
		}
		registry.Register(OnStartFunc(nil))
		if registry.Count() != 2 {
			t.Errorf("expected 2, got %d", registry.Count())
		}
	})

	t.Run("GetAll returns copy of hooks", func(t *testing.T) {
		registry := NewHookRegistry()
		registry.Register(OnInitFunc(nil))
		registry.Register(OnStartFunc(nil))

		hooks := registry.GetAll()
		if len(hooks) != 2 {
			t.Errorf("expected 2 hooks, got %d", len(hooks))
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		registry := NewHookRegistry()
		var wg sync.WaitGroup

		for range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				registry.Register(OnInitFunc(nil))
			}()
		}
		wg.Wait()

		if registry.Count() != 100 {
			t.Errorf("expected 100 hooks, got %d", registry.Count())
		}
	})
}

func TestGlobalHookRegistry(t *testing.T) {
	t.Parallel()
	t.Run("GlobalHookRegistry returns singleton", func(t *testing.T) {
		r1 := GlobalHookRegistry()
		r2 := GlobalHookRegistry()
		if r1 != r2 {
			t.Error("expected same registry instance")
		}
	})

	t.Run("RegisterHook adds to global registry", func(t *testing.T) {
		before := GlobalHookRegistry().Count()
		RegisterHook(OnInitFunc(nil))
		after := GlobalHookRegistry().Count()
		if after != before+1 {
			t.Errorf("expected count to increase by 1, got %d -> %d", before, after)
		}
	})

	t.Run("RegisterHookFunc adds to global registry", func(t *testing.T) {
		before := GlobalHookRegistry().Count()
		RegisterHookFunc(nil, nil, nil)
		after := GlobalHookRegistry().Count()
		if after != before+1 {
			t.Errorf("expected count to increase by 1, got %d -> %d", before, after)
		}
	})
}

func TestNewHookFunc(t *testing.T) {
	t.Parallel()
	t.Run("creates hook with all functions", func(t *testing.T) {
		var initCalled, startCalled, stopCalled bool
		hook := NewHookFunc(
			func(ctx context.Context) error { initCalled = true; return nil },
			func(ctx context.Context) error { startCalled = true; return nil },
			func(ctx context.Context) error { stopCalled = true; return nil },
		)

		ctx := context.Background()
		_ = hook.OnInit(ctx)
		_ = hook.OnStart(ctx)
		_ = hook.OnStop(ctx)

		if !initCalled || !startCalled || !stopCalled {
			t.Error("expected all functions to be called")
		}
	})

	t.Run("creates hook with nil functions", func(t *testing.T) {
		hook := NewHookFunc(nil, nil, nil)
		ctx := context.Background()

		if err := hook.OnInit(ctx); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if err := hook.OnStart(ctx); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if err := hook.OnStop(ctx); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})
}
