package refresh

import (
	"errors"
	"testing"
	"time"

	"github.com/xudefa/enhance/config/environment"
)

func TestRefreshManager_Refresh(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})

	manager := NewRefreshManager(env)

	err := manager.Refresh()
	if err != nil {
		t.Errorf("Refresh() error = %v", err)
	}
}

func TestRefreshManager_AddRefreshListener(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	manager := NewRefreshManager(env)

	called := false
	listener := &mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			called = true
			return nil
		},
	}

	manager.AddRefreshListener(listener)

	_ = manager.Refresh()

	if !called {
		t.Error("RefreshListener.OnRefresh() not called")
	}
}

func TestRefreshManager_RefreshListenerError(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	manager := NewRefreshManager(env)

	expectedErr := errors.New("listener error")
	listener := &mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			return expectedErr
		},
	}

	manager.AddRefreshListener(listener)

	err := manager.Refresh()
	if !errors.Is(err, expectedErr) {
		t.Errorf("Refresh() error = %v, want %v", err, expectedErr)
	}
}

func TestRefreshManager_RefreshEventFields(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{
		"app.name": "test-app",
	})
	manager := NewRefreshManager(env)

	var receivedEvent RefreshEvent
	listener := &mockRefreshListener{
		onRefresh: func(event RefreshEvent) error {
			receivedEvent = event
			return nil
		},
	}

	manager.AddRefreshListener(listener)
	_ = manager.Refresh()

	if receivedEvent.Environment != env {
		t.Error("RefreshEvent.Environment should be the environment")
	}

	if receivedEvent.Timestamp == 0 {
		t.Error("RefreshEvent.Timestamp should be set")
	}
}

func TestRefreshManager_MultipleListeners(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	manager := NewRefreshManager(env)

	callCount := 0
	for i := 0; i < 3; i++ {
		listener := &mockRefreshListener{
			onRefresh: func(event RefreshEvent) error {
				callCount++
				return nil
			},
		}
		manager.AddRefreshListener(listener)
	}

	_ = manager.Refresh()

	if callCount != 3 {
		t.Errorf("expected 3 listener calls, got %d", callCount)
	}
}

func TestRefreshManager_StartStop(t *testing.T) {
	t.Parallel()

	env := environment.NewMapEnvironment(map[string]string{})
	manager := NewRefreshManager(env)

	if err := manager.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if err := manager.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestRefreshEvent_ChangedKeys(t *testing.T) {
	t.Parallel()

	event := RefreshEvent{
		Environment: environment.NewMapEnvironment(map[string]string{}),
		ChangedKeys: []string{"server.port", "server.host"},
		Timestamp:   time.Now().UnixMilli(),
	}

	if len(event.ChangedKeys) != 2 {
		t.Errorf("expected 2 ChangedKeys, got %d", len(event.ChangedKeys))
	}
}

type mockRefreshListener struct {
	onRefresh func(event RefreshEvent) error
}

func (l *mockRefreshListener) OnRefresh(event RefreshEvent) error {
	return l.onRefresh(event)
}
