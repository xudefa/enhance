package config

import (
	"sync"
	"testing"
	"time"
)

func TestNewWatchManager(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()
	if m == nil {
		t.Fatal("NewWatchManager returned nil")
	}
}

func TestWatchManager_RegisterAndNotify(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()

	var received WatchEvent
	m.Register("test", func(event WatchEvent) {
		received = event
	})

	evt := WatchEvent{
		Type:      EventModify,
		Key:       "app.name",
		Value:     "new-value",
		Timestamp: time.Now(),
		Source:    "test",
	}
	m.Notify(evt)

	if received.Key != "app.name" {
		t.Errorf("Key = %q, want %q", received.Key, "app.name")
	}
	if received.Value != "new-value" {
		t.Errorf("Value = %v, want %q", received.Value, "new-value")
	}
}

func TestWatchManager_Unregister(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()

	called := false
	m.Register("test", func(event WatchEvent) {
		called = true
	})
	m.Unregister("test")
	m.Notify(WatchEvent{})

	if called {
		t.Error("callback should not be called after Unregister")
	}
}

func TestWatchManager_MultipleCallbacks(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()

	var mu sync.Mutex
	callCount := 0
	for i := 0; i < 3; i++ {
		key := "cb" + string(rune('0'+i))
		m.Register(key, func(event WatchEvent) {
			mu.Lock()
			callCount++
			mu.Unlock()
		})
	}

	m.Notify(WatchEvent{})

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestWatchManager_NilCallback(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()
	m.Register("nil-cb", nil)
	// Should not panic
	m.Notify(WatchEvent{})
}

func TestWatchManager_Close(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()
	m.Register("cb", func(event WatchEvent) {})
	m.AddSource("src", make(chan WatchEvent))
	m.Close()

	// After close, operations should be no-ops
	m.Register("cb2", func(event WatchEvent) {}) // no panic
	m.Unregister("cb")                           // no panic
	m.Notify(WatchEvent{})                        // no panic
	m.AddSource("src2", make(chan WatchEvent))     // no panic
}

func TestWatchManager_CloseIdempotent(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()
	m.Close()
	m.Close() // no panic
}

func TestWatchManager_NotifyAfterClose(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()

	called := false
	m.Register("cb", func(event WatchEvent) {
		called = true
	})
	m.Close()
	m.Notify(WatchEvent{})

	if called {
		t.Error("callback should not be called after Close")
	}
}

func TestWatchManager_GetSourceAfterClose(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()
	m.AddSource("src", make(chan WatchEvent))
	m.Close()

	_, ok := m.GetSource("src")
	if ok {
		t.Error("GetSource should return false after Close")
	}
}

func TestWatchEvent_Fields(t *testing.T) {
	t.Parallel()
	now := time.Now()
	evt := WatchEvent{
		Type:      EventCreate,
		Key:       "new.key",
		Value:     42,
		Timestamp: now,
		Source:    "test-source",
	}
	if evt.Type != EventCreate {
		t.Errorf("Type = %q", evt.Type)
	}
	if evt.Key != "new.key" {
		t.Errorf("Key = %q", evt.Key)
	}
	if evt.Value != 42 {
		t.Errorf("Value = %v", evt.Value)
	}
	if evt.Source != "test-source" {
		t.Errorf("Source = %q", evt.Source)
	}
}

func TestWatchManager_OverwriteCallback(t *testing.T) {
	t.Parallel()
	m := NewWatchManager()

	called1 := false
	called2 := false
	m.Register("key", func(event WatchEvent) { called1 = true })
	m.Register("key", func(event WatchEvent) { called2 = true })

	m.Notify(WatchEvent{})

	if called1 {
		t.Error("first callback should be overwritten")
	}
	if !called2 {
		t.Error("second callback should be called")
	}
}
