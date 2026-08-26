package refresh

import (
	"testing"
	"time"

	"github.com/xudefa/enhance/event"
)

func TestEventRouter_FindAffectedBeans_NoMatch(t *testing.T) {
	t.Parallel()
	bus := event.NewEventBus()
	router := NewEventRouter(bus)

	affected := router.findAffectedBeans([]string{"unknown.key"})
	if len(affected) != 0 {
		t.Errorf("expected 0 affected beans, got %d", len(affected))
	}
}

func TestEventRouter_FindAffectedBeans_EmptyKeys(t *testing.T) {
	t.Parallel()
	bus := event.NewEventBus()
	router := NewEventRouter(bus)
	router.RegisterBean("svc", []string{"k1"})

	affected := router.findAffectedBeans(nil)
	if len(affected) != 0 {
		t.Errorf("expected 0 affected, got %d", len(affected))
	}
}

func TestEventRouter_OnConfigChange_PublishesEvents(t *testing.T) {
	t.Parallel()
	bus := event.NewEventBus()
	router := NewEventRouter(bus)

	var receivedEvents []*BeanRefreshEvent
	bus.Subscribe("BeanRefresh", func(e event.ApplicationEvent) {
		if bre, ok := e.(*BeanRefreshEvent); ok {
			receivedEvents = append(receivedEvents, bre)
		}
	})

	router.RegisterBean("svc1", []string{"key1"})
	router.RegisterBean("svc2", []string{"key2"})

	configEvt := NewConfigChangeEvent(
		"modify",
		[]string{"key1"},
		map[string]any{"key1": "old"},
		map[string]any{"key1": "new"},
		"test",
	)
	bus.Publish(&configEvt)

	if len(receivedEvents) != 1 {
		t.Fatalf("expected 1 BeanRefreshEvent, got %d", len(receivedEvents))
	}
	if receivedEvents[0].BeanID != "svc1" {
		t.Errorf("BeanID = %q, want %q", receivedEvents[0].BeanID, "svc1")
	}
}

func TestEventRouter_OnConfigChange_IgnoresNonConfigEvent(t *testing.T) {
	t.Parallel()
	bus := event.NewEventBus()
	router := NewEventRouter(bus)

	router.RegisterBean("svc1", []string{"key1"})

	var count int
	bus.Subscribe("BeanRefresh", func(e event.ApplicationEvent) {
		count++
	})

	bus.Publish(&dummyEvent{})

	if count != 0 {
		t.Errorf("expected 0 events, got %d", count)
	}
}

func TestBeanRefreshEvent_TypeAndTimestamp(t *testing.T) {
	t.Parallel()
	e := &BeanRefreshEvent{
		BeanID:     "myBean",
		ConfigKeys: []string{"k1"},
		OldValues:  map[string]any{"k1": "old"},
		NewValues:  map[string]any{"k1": "new"},
	}
	if e.Type() != "BeanRefresh" {
		t.Errorf("Type() = %q, want %q", e.Type(), "BeanRefresh")
	}
	if e.Timestamp().IsZero() {
		t.Error("Timestamp() should not be zero")
	}
}

func TestEventRouter_DuplicateKeys(t *testing.T) {
	t.Parallel()
	bus := event.NewEventBus()
	router := NewEventRouter(bus)

	router.RegisterBean("svc1", []string{"k1", "k1"})

	affected := router.findAffectedBeans([]string{"k1"})
	if len(affected) != 1 {
		t.Errorf("expected 1 affected bean (deduped), got %d", len(affected))
	}
}

type dummyEvent struct{}

func (d *dummyEvent) Type() string      { return "dummy" }
func (d *dummyEvent) Timestamp() time.Time { return time.Time{} }
