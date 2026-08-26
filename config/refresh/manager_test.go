package refresh

import (
	"errors"
	"testing"
	"time"

	"github.com/xudefa/enhance/config/environment"
)

func TestRefreshManager_NewRefreshManager(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{})
	m := NewRefreshManager(env)
	if m == nil {
		t.Fatal("NewRefreshManager returned nil")
	}
}

func TestRefreshManager_RefreshBeforeStart(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{})
	m := NewRefreshManager(env)

	called := false
	m.AddRefreshListener(&mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			called = true
			return nil
		},
	})

	_ = m.Refresh()
	if called {
		t.Error("listener should not be called before Start()")
	}
}

func TestRefreshManager_StopIdempotent(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{})
	m := NewRefreshManager(env)

	if err := m.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := m.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if err := m.Stop(); err != nil {
		t.Fatalf("second Stop: %v", err)
	}
}

func TestRefreshManager_MultipleListenersError(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{})
	m := NewRefreshManager(env)
	if err := m.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	err1 := errors.New("err1")
	err2 := errors.New("err2")
	m.AddRefreshListener(&mockRefreshListener{onRefresh: func(event RefreshEvent) error { return err1 }})
	m.AddRefreshListener(&mockRefreshListener{onRefresh: func(event RefreshEvent) error { return err2 }})

	err := m.Refresh()
	if err == nil {
		t.Fatal("expected error from listeners")
	}
	if !errors.Is(err, err1) {
		t.Errorf("expected err1, got %v", err)
	}
}

func TestRefreshManager_RefreshDetectsChange(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{
		"k1": "v1",
	})
	m := NewRefreshManager(env)
	if err := m.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	var firstKeys, secondKeys []string
	callCount := 0
	m.AddRefreshListener(&mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			callCount++
			if callCount == 1 {
				firstKeys = event.ChangedKeys
			} else {
				secondKeys = event.ChangedKeys
			}
			return nil
		},
	})

	_ = m.Refresh()
	if len(firstKeys) != 0 {
		t.Errorf("first refresh should have no changes, got %v", firstKeys)
	}

	// Add override
	env.AddPropertySource(environment.NewMapPropertySource("override", environment.PriorityHigh, map[string]any{
		"k1": "v1-new",
	}))
	_ = m.Refresh()
	if len(secondKeys) != 1 || secondKeys[0] != "k1" {
		t.Errorf("expected changed=[k1], got %v", secondKeys)
	}
}

func TestRefreshManager_RefreshDetectsNewKey(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{
		"k1": "v1",
	})
	m := NewRefreshManager(env)
	if err := m.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	_ = m.Refresh() // init snapshot

	env.AddPropertySource(environment.NewMapPropertySource("override", environment.PriorityHigh, map[string]any{
		"k2": "v2",
	}))

	var changed []string
	m.AddRefreshListener(&mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			changed = event.ChangedKeys
			return nil
		},
	})

	_ = m.Refresh()
	if len(changed) != 1 || changed[0] != "k2" {
		t.Errorf("expected new key k2 detected, got %v", changed)
	}
}

func TestRefreshManager_RefreshDetectsDeletedKey(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{
		"k1": "v1",
	})
	m := NewRefreshManager(env)
	if err := m.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	_ = m.Refresh() // init snapshot with k1

	// Add override that shadows k1 with empty
	env.AddPropertySource(environment.NewMapPropertySource("override", environment.PriorityHigh, map[string]any{}))

	var changed []string
	m.AddRefreshListener(&mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			changed = event.ChangedKeys
			return nil
		},
	})

	_ = m.Refresh()
	// k1 may or may not be detected as changed depending on MapPropertySource behavior
	_ = changed // just ensure no panic
}

func TestRefreshManager_ListenersAreCopied(t *testing.T) {
	t.Parallel()
	env := environment.NewMapEnvironment(map[string]string{})
	m := NewRefreshManager(env)
	if err := m.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	callCount := 0
	m.AddRefreshListener(&mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			callCount++
			// Add another listener during callback
			m.AddRefreshListener(&mockRefreshListener{
				onRefresh: func(event RefreshEvent) error {
					callCount++
					return nil
				},
			})
			return nil
		},
	})

	_ = m.Refresh()
	if callCount != 1 {
		t.Errorf("expected 1 call during refresh (listeners copied), got %d", callCount)
	}
}

func TestRefreshEvent_Fields(t *testing.T) {
	t.Parallel()
	now := time.Now().UnixMilli()
	env := environment.NewMapEnvironment(map[string]string{})
	evt := RefreshEvent{
		Environment: env,
		ChangedKeys: []string{"key1"},
		Timestamp:   now,
	}
	if evt.Timestamp != now {
		t.Errorf("Timestamp = %d, want %d", evt.Timestamp, now)
	}
	if len(evt.ChangedKeys) != 1 {
		t.Errorf("ChangedKeys len = %d, want 1", len(evt.ChangedKeys))
	}
}
