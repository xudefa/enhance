package event

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestEvent_PublishAndSubscribe 测试事件发布与订阅。
func TestEvent_PublishAndSubscribe(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	var received atomic.Bool
	bus.Subscribe("user.created", func(e ApplicationEvent) {
		received.Store(true)
	})

	bus.Publish(NewUserCreatedEvent(1, "Alice"))

	if !received.Load() {
		t.Fatal("事件应该被接收")
	}
}

// TestEvent_MultipleListeners 测试多监听器。
func TestEvent_MultipleListeners(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	var count atomic.Int32

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	bus.Publish(NewUserCreatedEvent(1, "Alice"))

	if count.Load() != 3 {
		t.Fatalf("应该有 3 个监听器被调用，got: %d", count.Load())
	}
}

// TestEvent_Unsubscribe 测试取消订阅。
func TestEvent_Unsubscribe(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	var count atomic.Int32

	listener := func(e ApplicationEvent) {
		count.Add(1)
	}

	bus.Subscribe("user.created", listener)
	bus.Publish(NewUserCreatedEvent(1, "Alice"))

	if count.Load() != 1 {
		t.Fatalf("第一次发布应该有 1 个监听器被调用，got: %d", count.Load())
	}

	bus.Unsubscribe("user.created", listener)
	bus.Publish(NewUserCreatedEvent(2, "Bob"))

	if count.Load() != 1 {
		t.Fatalf("取消订阅后不应该再接收事件，got: %d", count.Load())
	}
}

// TestEvent_DifferentEventTypes 测试不同事件类型。
func TestEvent_DifferentEventTypes(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	var userCreated, orderCreated atomic.Bool

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		userCreated.Store(true)
	})

	bus.Subscribe("order.created", func(e ApplicationEvent) {
		orderCreated.Store(true)
	})

	bus.Publish(NewUserCreatedEvent(1, "Alice"))

	if !userCreated.Load() {
		t.Fatal("user.created 事件应该被接收")
	}
	if orderCreated.Load() {
		t.Fatal("order.created 事件不应该被接收")
	}
}

// TestEvent_EventData 测试事件数据传递。
func TestEvent_EventData(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	var capturedEvent *UserCreatedEvent
	var mu sync.Mutex

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		mu.Lock()
		defer mu.Unlock()
		capturedEvent = e.(*UserCreatedEvent)
	})

	bus.Publish(NewUserCreatedEvent(42, "Bob"))

	if capturedEvent == nil {
		t.Fatal("事件应该被捕获")
	}
	if capturedEvent.UserID != 42 {
		t.Fatalf("UserID 应该是 42，got: %d", capturedEvent.UserID)
	}
	if capturedEvent.UserName != "Bob" {
		t.Fatalf("UserName 应该是 Bob，got: %s", capturedEvent.UserName)
	}
}

// TestEvent_NoListeners 测试无监听器发布。
func TestEvent_NoListeners(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	// 不应该 panic
	bus.Publish(NewUserCreatedEvent(1, "Alice"))
}

// TestEvent_ConcurrentPublish 测试并发发布。
func TestEvent_ConcurrentPublish(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()

	var count atomic.Int32

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	// 并发发布 100 个事件
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(id int) {
			bus.Publish(NewUserCreatedEvent(id, "User"))
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	if count.Load() != 100 {
		t.Fatalf("应该有 100 个事件被处理，got: %d", count.Load())
	}
}

// TestEvent_Ordering 测试事件顺序。
func TestEvent_Ordering(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithOrdering()

	var order []int
	var mu sync.Mutex

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
	})

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
	})

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
	})

	bus.Publish(NewUserCreatedEvent(1, "Alice"))

	if len(order) != 3 {
		t.Fatalf("应该有 3 个监听器被调用，got: %d", len(order))
	}

	// 验证顺序
	for i, v := range order {
		if v != i+1 {
			t.Fatalf("顺序 %d 应该是 %d，got: %d", i, i+1, v)
		}
	}
}

// TestEvent_SubscribeOnce 测试一次性订阅。
func TestEvent_SubscribeOnce(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithOrdering()

	var count atomic.Int32

	bus.SubscribeOnce("user.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	bus.Publish(NewUserCreatedEvent(1, "Alice"))
	bus.Publish(NewUserCreatedEvent(2, "Bob"))

	if count.Load() != 1 {
		t.Fatalf("一次性订阅应该只被调用一次，got: %d", count.Load())
	}
}

// TestEvent_Clear 测试清除监听器。
func TestEvent_Clear(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithOrdering()

	var count atomic.Int32

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	bus.Publish(NewUserCreatedEvent(1, "Alice"))
	if count.Load() != 1 {
		t.Fatalf("第一次发布应该有 1 个监听器被调用，got: %d", count.Load())
	}

	bus.Clear("user.created")
	bus.Publish(NewUserCreatedEvent(2, "Bob"))

	if count.Load() != 1 {
		t.Fatalf("清除后不应该再接收事件，got: %d", count.Load())
	}
}

// TestEvent_ClearAll 测试清除所有监听器。
func TestEvent_ClearAll(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithOrdering()

	var count atomic.Int32

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	bus.Subscribe("order.created", func(e ApplicationEvent) {
		count.Add(1)
	})

	bus.ClearAll()

	bus.Publish(NewUserCreatedEvent(1, "Alice"))

	if count.Load() != 0 {
		t.Fatalf("清除所有后不应该再接收事件，got: %d", count.Load())
	}
}

// TestEvent_Listeners 测试监听器数量。
func TestEvent_Listeners(t *testing.T) {
	t.Parallel()

	bus := NewEventBusWithOrdering()

	bus.Subscribe("user.created", func(e ApplicationEvent) {})
	bus.Subscribe("user.created", func(e ApplicationEvent) {})
	bus.Subscribe("order.created", func(e ApplicationEvent) {})

	if bus.Listeners("user.created") != 2 {
		t.Fatalf("user.created 应该有 2 个监听器，got: %d", bus.Listeners("user.created"))
	}

	if bus.Listeners("order.created") != 1 {
		t.Fatalf("order.created 应该有 1 个监听器，got: %d", bus.Listeners("order.created"))
	}
}
