package event

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMakeEventKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType string
		ts        int64
	}{
		{"basic", "test.event", time.Now().UnixNano()},
		{"empty type", "", 0},
		{"future timestamp", "event", time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key := makeEventKey(tt.eventType, tt.ts)
			if !strings.HasPrefix(key, tt.eventType) {
				t.Errorf("expected key to start with %q, got %q", tt.eventType, key)
			}
		})
	}
}

func TestMakeEventKey_Unique(t *testing.T) {
	t.Parallel()

	ts := time.Now().UnixNano()
	keys := make(map[string]bool)
	for range 100 {
		key := makeEventKey("test", ts)
		if keys[key] {
			t.Fatalf("duplicate key: %s", key)
		}
		keys[key] = true
	}
}

func TestEventBusWithDeadLetter_NilEvent(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithDeadLetter(context.Background())
	bus.PublishWithRecovery(nil)

	if bus.DeadLetterQueue().Size() != 0 {
		t.Error("nil event should not be added to DLQ")
	}
}

func TestEventBusWithDeadLetter_Close(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithDeadLetter(context.Background(),
		WithMaxRetries(0),
	)

	var called bool
	bus.Subscribe("close.test", func(e ApplicationEvent) {
		called = true
	})

	bus.PublishWithRecovery(&BaseEvent{EventType: "close.test"})
	time.Sleep(100 * time.Millisecond)

	bus.Close()

	if !called {
		t.Error("expected handler to be called before close")
	}
}
