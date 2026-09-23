package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// UserCreatedEvent 用户创建事件。
type UserCreatedEvent struct {
	BaseEvent
	UserID   int
	UserName string
}

// TestEventBus_SubscribePublish 测试事件总线的基本订阅和发布功能
func TestEventBus_SubscribePublish(t *testing.T) {
	t.Parallel()
	// 创建事件总线实例
	bus := NewEventBus()
	var called atomic.Int32

	// 订阅 "test.event" 事件
	bus.Subscribe("test.event", func(e ApplicationEvent) {
		called.Add(1)
	})

	// 发布一个测试事件
	bus.Publish(&BaseEvent{
		EventType: "test.event",
		EventTime: time.Now(),
	})

	// 验证监听器被调用了一次
	if called.Load() != 1 {
		t.Fatal("期望监听器被调用一次，但实际未调用")
	}
}

// TestEventBus_MultipleListeners 测试多个监听器订阅同一事件
func TestEventBus_MultipleListeners(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var count int32

	// 注册 3 个监听器到 "multi" 事件
	for range 3 {
		bus.Subscribe("multi", func(e ApplicationEvent) {
			atomic.AddInt32(&count, 1)
		})
	}

	// 发布一个 "multi" 事件
	bus.Publish(&BaseEvent{EventType: "multi"})

	// 验证所有 3 个监听器都被调用
	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("期望调用 3 次，实际调用 %d 次", count)
	}
}

// TestEventBus_Unsubscribe 测试取消订阅功能
func TestEventBus_Unsubscribe(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	var called int32

	// 定义监听器函数
	listener := func(e ApplicationEvent) {
		atomic.AddInt32(&called, 1)
	}

	// 订阅事件并发布
	bus.Subscribe("test", listener)
	bus.Publish(&BaseEvent{EventType: "test"})

	// 验证取消订阅前监听器被调用
	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("期望取消订阅前监听器被调用一次")
	}

	// 取消订阅并再次发布事件
	bus.Unsubscribe("test", listener)
	bus.Publish(&BaseEvent{EventType: "test"})

	// 验证取消订阅后监听器不再被调用
	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("期望取消订阅后监听器不再被调用")
	}
}

// TestEventBus_UnsubscribeNil 测试取消订阅 nil 监听器的安全性
func TestEventBus_UnsubscribeNil(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	bus.Subscribe("test", func(e ApplicationEvent) {})

	// 取消订阅 nil 监听器不应该引发 panic
	bus.Unsubscribe("test", nil)
	bus.Unsubscribe("nonexistent", nil)
}

// TestEventBus_NoPanicOnUnsubscribedEvent 测试发布没有监听器的事件不会引发 panic
func TestEventBus_NoPanicOnUnsubscribedEvent(t *testing.T) {
	t.Parallel()
	bus := NewEventBus()
	// 发布一个没有监听器的事件，不应该触发 panic
	bus.Publish(&BaseEvent{EventType: "nonexistent"})
}

func TestBaseEvent_Timestamp(t *testing.T) {
	t.Parallel()
	now := time.Now()
	event := &BaseEvent{EventType: "test", EventTime: now}

	if event.Type() != "test" {
		t.Fatalf("expected type test, got %s", event.Type())
	}
	if !event.Timestamp().Equal(now) {
		t.Fatal("timestamp mismatch")
	}

	// 零值时间戳应该自动填充
	e2 := &BaseEvent{EventType: "test2"}
	if e2.Timestamp().IsZero() {
		t.Fatal("expected auto-populated timestamp")
	}
}

// TestEventBus_ConcurrentSubscribePublish 测试并发订阅和发布的正确性
func TestEventBus_ConcurrentSubscribePublish(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	var callCount int32
	const goroutines = 100

	// 并发订阅
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			bus.Subscribe("concurrent", func(e ApplicationEvent) {
				atomic.AddInt32(&callCount, 1)
			})
		}()
	}
	wg.Wait()

	// 发布事件
	bus.Publish(&BaseEvent{EventType: "concurrent"})

	// 验证所有监听器都被调用
	if atomic.LoadInt32(&callCount) != goroutines {
		t.Fatalf("期望 %d 次调用，实际 %d 次", goroutines, callCount)
	}
}

// TestEventBus_ConcurrentUnsubscribe 测试并发取消订阅的安全性
func TestEventBus_ConcurrentUnsubscribe(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	const goroutines = 50

	var listeners [goroutines]EventListener
	var wg sync.WaitGroup

	// 注册所有监听器
	for i := 0; i < goroutines; i++ {
		listeners[i] = func(e ApplicationEvent) {}
		bus.Subscribe("test", listeners[i])
	}

	// 并发取消订阅
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			bus.Unsubscribe("test", listeners[idx])
		}(i)
	}
	wg.Wait()

	// 验证所有监听器都已取消
	var callCount int32
	bus.Subscribe("test", func(e ApplicationEvent) {
		atomic.AddInt32(&callCount, 1)
	})
	bus.Publish(&BaseEvent{EventType: "test"})

	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatal("取消订阅后仍有残留监听器")
	}
}

// TestEventBus_MultipleEventTypes 测试多种事件类型的隔离性
func TestEventBus_MultipleEventTypes(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	var countA, countB, countC int32

	bus.Subscribe("a", func(e ApplicationEvent) {
		atomic.AddInt32(&countA, 1)
	})
	bus.Subscribe("b", func(e ApplicationEvent) {
		atomic.AddInt32(&countB, 1)
	})
	bus.Subscribe("c", func(e ApplicationEvent) {
		atomic.AddInt32(&countC, 1)
	})

	// 发布不同类型的事件
	bus.Publish(&BaseEvent{EventType: "a"})
	bus.Publish(&BaseEvent{EventType: "b"})
	bus.Publish(&BaseEvent{EventType: "a"})
	bus.Publish(&BaseEvent{EventType: "c"})

	if atomic.LoadInt32(&countA) != 2 {
		t.Fatalf("a 期望 2 次，实际 %d 次", countA)
	}
	if atomic.LoadInt32(&countB) != 1 {
		t.Fatalf("b 期望 1 次，实际 %d 次", countB)
	}
	if atomic.LoadInt32(&countC) != 1 {
		t.Fatalf("c 期望 1 次，实际 %d 次", countC)
	}
}

// TestEventBus_UnsubscribeNonExistent 测试取消不存在的监听器
func TestEventBus_UnsubscribeNonExistent(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	bus.Subscribe("test", func(e ApplicationEvent) {})

	// 取消不存在的监听器应该是空操作
	nonExistent := func(e ApplicationEvent) {}
	bus.Unsubscribe("test", nonExistent)
	bus.Unsubscribe("nonexistent", func(e ApplicationEvent) {})
}

// TestEventBus_PublishOrder 测试监听器调用顺序
func TestEventBus_PublishOrder(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	var order []int
	var mu sync.Mutex

	bus.Subscribe("ordered", func(e ApplicationEvent) {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
	})
	bus.Subscribe("ordered", func(e ApplicationEvent) {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
	})
	bus.Subscribe("ordered", func(e ApplicationEvent) {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
	})

	bus.Publish(&BaseEvent{EventType: "ordered"})

	expected := []int{1, 2, 3}
	if len(order) != len(expected) {
		t.Fatalf("期望 %d 次调用，实际 %d 次", len(expected), len(order))
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("调用顺序错误：期望 %v，实际 %v", expected, order)
		}
	}
}

// TestEventBus_EmbeddedEvent 测试嵌入 BaseEvent 的自定义事件
func TestEventBus_EmbeddedEvent(t *testing.T) {
	t.Parallel()

	type UserCreatedEvent struct {
		BaseEvent
		UserID int
		Email  string
	}

	bus := NewEventBus()
	var receivedEmail string
	var receivedUserID int

	bus.Subscribe("user.created", func(e ApplicationEvent) {
		if evt, ok := e.(*UserCreatedEvent); ok {
			receivedEmail = evt.Email
			receivedUserID = evt.UserID
		}
	})

	evt := &UserCreatedEvent{
		BaseEvent: BaseEvent{EventType: "user.created"},
		UserID:    123,
		Email:     "test@example.com",
	}
	bus.Publish(evt)

	if receivedEmail != "test@example.com" {
		t.Fatalf("期望邮箱 'test@example.com'，实际 '%s'", receivedEmail)
	}
	if receivedUserID != 123 {
		t.Fatalf("期望用户ID 123，实际 %d", receivedUserID)
	}
}

func NewUserCreatedEvent(userID int, userName string) *UserCreatedEvent {
	return &UserCreatedEvent{
		UserID:   userID,
		UserName: userName,
	}
}

func (e *UserCreatedEvent) Type() string {
	return "user.created"
}

func (e *UserCreatedEvent) Timestamp() time.Time {
	return e.EventTime
}

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
