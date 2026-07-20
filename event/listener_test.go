package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEventBusWithOrdering_BasicSubscribePublish(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	var called int32

	bus.Subscribe("test.event", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	})

	bus.Publish(&BaseEvent{EventType: "test.event"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener to be called once")
	}
}

func TestEventBusWithOrdering_OrderedExecution(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	var order []int
	var mu sync.Mutex

	bus.SubscribeWithConfig("test", ListenerConfig{
		Handler: func(e ApplicationEvent) {
			mu.Lock()
			order = append(order, 3)
			mu.Unlock()
		},
		Order: 30,
	})

	bus.SubscribeWithConfig("test", ListenerConfig{
		Handler: func(e ApplicationEvent) {
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
		},
		Order: 10,
	})

	bus.SubscribeWithConfig("test", ListenerConfig{
		Handler: func(e ApplicationEvent) {
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
		},
		Order: 20,
	})

	bus.Publish(&BaseEvent{EventType: "test"})

	mu.Lock()
	defer mu.Unlock()

	if len(order) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(order))
	}

	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("expected execution order [1,2,3], got %v", order)
	}
}

func TestEventBusWithOrdering_ConditionFilter(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	var called int32

	bus.SubscribeWithConfig("test", ListenerConfig{
		Handler: func(e ApplicationEvent) {
			atomic.AddInt32(&called, 1)
		},
		Condition: func(e ApplicationEvent) bool {
			return e.Type() == "test"
		},
	})

	bus.Publish(&BaseEvent{EventType: "test"})
	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener to be called")
	}

	bus.Publish(&BaseEvent{EventType: "other"})
	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener NOT to be called for other event")
	}
}

func TestEventBusWithOrdering_SubscribeOnce(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	var called int32

	bus.SubscribeOnce("test", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	})

	bus.Publish(&BaseEvent{EventType: "test"})
	bus.Publish(&BaseEvent{EventType: "test"})
	bus.Publish(&BaseEvent{EventType: "test"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected listener to be called only once, got %d", called)
	}
}

func TestEventBusWithOrdering_AsyncExecution(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	var called int32
	var wg sync.WaitGroup
	wg.Add(1)

	bus.SubscribeWithConfig("test", ListenerConfig{
		Handler: func(e ApplicationEvent) {
			atomic.AddInt32(&called, 1)
			wg.Done()
		},
		Async: true,
	})

	bus.Publish(&BaseEvent{EventType: "test"})

	wg.Wait()
	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected async listener to be called")
	}
}

func TestEventBusWithOrdering_Unsubscribe(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	var called int32

	listener := func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	}

	bus.Subscribe("test", listener)
	bus.Publish(&BaseEvent{EventType: "test"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener to be called before unsubscribe")
	}

	bus.Unsubscribe("test", listener)
	bus.Publish(&BaseEvent{EventType: "test"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected listener NOT to be called after unsubscribe")
	}
}

func TestEventBusWithOrdering_Listeners(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()

	bus.Subscribe("test", func(e ApplicationEvent) {})
	bus.Subscribe("test", func(e ApplicationEvent) {})
	bus.Subscribe("other", func(e ApplicationEvent) {})

	if bus.Listeners("test") != 2 {
		t.Errorf("expected 2 listeners for 'test', got %d", bus.Listeners("test"))
	}

	if bus.Listeners("other") != 1 {
		t.Errorf("expected 1 listener for 'other', got %d", bus.Listeners("other"))
	}

	if bus.Listeners("nonexistent") != 0 {
		t.Errorf("expected 0 listeners for 'nonexistent', got %d", bus.Listeners("nonexistent"))
	}
}

func TestEventBusWithOrdering_Clear(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()

	bus.Subscribe("test", func(e ApplicationEvent) {})
	bus.Subscribe("test", func(e ApplicationEvent) {})
	bus.Subscribe("other", func(e ApplicationEvent) {})

	bus.Clear("test")

	if bus.Listeners("test") != 0 {
		t.Errorf("expected 0 listeners after clear, got %d", bus.Listeners("test"))
	}

	if bus.Listeners("other") != 1 {
		t.Errorf("expected 'other' listeners unaffected, got %d", bus.Listeners("other"))
	}
}

func TestEventBusWithOrdering_ClearAll(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()

	bus.Subscribe("test", func(e ApplicationEvent) {})
	bus.Subscribe("other", func(e ApplicationEvent) {})

	bus.ClearAll()

	if bus.Listeners("test") != 0 {
		t.Errorf("expected 0 listeners after clear all, got %d", bus.Listeners("test"))
	}

	if bus.Listeners("other") != 0 {
		t.Errorf("expected 0 listeners after clear all, got %d", bus.Listeners("other"))
	}
}

func TestListenerConfig_ChainMethods(t *testing.T) {
	t.Parallel()
	config := NewListenerConfig(func(e ApplicationEvent) {}).
		WithOrder(10).
		WithCondition(ConditionAlways()).
		WithAsync(true)

	if config.Order != 10 {
		t.Errorf("expected order 10, got %d", config.Order)
	}

	if config.Condition == nil {
		t.Error("expected non-nil condition")
	}

	if !config.Condition(&BaseEvent{EventType: "test"}) {
		t.Error("expected ConditionAlways to return true")
	}

	if !config.Async {
		t.Error("expected async to be true")
	}
}

func TestConditionFunctions(t *testing.T) {
	t.Parallel()
	// 测试 ConditionAlways
	always := ConditionAlways()
	if !always(&BaseEvent{EventType: "test"}) {
		t.Error("ConditionAlways should return true")
	}

	// 测试 ConditionType
	typeCond := ConditionType("a", "b")
	if !typeCond(&BaseEvent{EventType: "a"}) {
		t.Error("ConditionType should match 'a'")
	}
	if !typeCond(&BaseEvent{EventType: "b"}) {
		t.Error("ConditionType should match 'b'")
	}
	if typeCond(&BaseEvent{EventType: "c"}) {
		t.Error("ConditionType should not match 'c'")
	}

	// 测试 ConditionAfter
	now := time.Now()
	afterCond := ConditionAfter(now.Add(-time.Hour))
	if !afterCond(&BaseEvent{EventType: "test", EventTime: now}) {
		t.Error("ConditionAfter should match recent event")
	}

	// 测试 ConditionBefore
	beforeCond := ConditionBefore(now.Add(time.Hour))
	if !beforeCond(&BaseEvent{EventType: "test", EventTime: now}) {
		t.Error("ConditionBefore should match past event")
	}
}

func TestLegacyEventBusAdapter(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	adapter := NewLegacyEventBusAdapter(bus)

	var called int32
	adapter.Subscribe("test", func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	})

	adapter.Publish(&BaseEvent{EventType: "test"})

	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected adapter to forward calls")
	}
}

func TestEventBusWithOrdering_NoPanicOnUnsubscribedEvent(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	bus.Publish(&BaseEvent{EventType: "nonexistent"})
}

func TestEventBusWithOrdering_UnsubscribeNil(t *testing.T) {
	t.Parallel()
	bus := NewEventBusWithOrdering()
	bus.Subscribe("test", func(e ApplicationEvent) {})
	bus.Unsubscribe("test", nil)
	bus.Unsubscribe("nonexistent", nil)
}
