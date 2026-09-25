// Package main demonstrates the enhance event-driven architecture:
// EventBus creation, event publishing, sync/async handling,
// dead letter queue for failed events, and event ordering.
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xudefa/enhance/event"
)

// ==================== Custom Events ====================

// OrderCreatedEvent is published when an order is created.
type OrderCreatedEvent struct {
	event.BaseEvent
	OrderID string
	Amount  float64
}

// PaymentProcessedEvent is published when payment completes.
type PaymentProcessedEvent struct {
	event.BaseEvent
	OrderID string
	Status  string
}

// NotificationEvent is published to send user notifications.
type NotificationEvent struct {
	event.BaseEvent
	Message string
}

// eventTracker 汇总事件处理统计与日志，供演示验证与输出。
type eventTracker struct {
	syncHandled    atomic.Int32
	orderedHandled atomic.Int32
	failedHandled  atomic.Int32
	mu             sync.Mutex
	eventLog       []string
}

// newEventTracker 创建事件跟踪器。
func newEventTracker() *eventTracker {
	return &eventTracker{eventLog: make([]string, 0)}
}

func main() {
	fmt.Println("=== enhance Event-Driven Architecture Example ===")
	fmt.Println()

	// ---- 1. Create EventBus and Dead Letter Queue ----
	bus := event.NewEventBus()
	dlq := event.NewDeadLetterQueue()

	tr := newEventTracker()
	registerSyncHandlers(bus, dlq, tr)

	busWithOrder := event.NewEventBusWithOrdering()
	registerOrderedHandlers(busWithOrder, tr)

	publishEvents(bus)
	publishOrderedEvent(busWithOrder)

	printResults(tr)
	printDeadLetterStats(dlq)
	printEventLog(tr)

	demoAsyncPublisher()

	fmt.Println()
	fmt.Println("=== Example completed successfully ===")
}

// registerSyncHandlers 注册同步处理器，其中 Notification 处理器会失败并写入死信队列。
func registerSyncHandlers(bus event.EventBus, dlq *event.DeadLetterQueue, tr *eventTracker) {
	// ---- 2. Register sync handlers ----
	bus.Subscribe("OrderCreated", func(e event.ApplicationEvent) {
		evt := e.(*OrderCreatedEvent)
		tr.syncHandled.Add(1)
		tr.mu.Lock()
		tr.eventLog = append(tr.eventLog, fmt.Sprintf("sync:OrderCreated(%s, %.0f)", evt.OrderID, evt.Amount))
		tr.mu.Unlock()
	})

	bus.Subscribe("PaymentProcessed", func(e event.ApplicationEvent) {
		evt := e.(*PaymentProcessedEvent)
		tr.syncHandled.Add(1)
		tr.mu.Lock()
		tr.eventLog = append(tr.eventLog, fmt.Sprintf("sync:PaymentProcessed(%s, %s)", evt.OrderID, evt.Status))
		tr.mu.Unlock()
	})

	// ---- 3. Register a handler that always fails (for DLQ demo) ----
	bus.Subscribe("Notification", func(e event.ApplicationEvent) {
		evt := e.(*NotificationEvent)
		tr.failedHandled.Add(1)
		// Simulate a processing failure
		fe := event.FailedEvent{
			Event:      evt,
			Err:        fmt.Errorf("notification delivery failed for: %s", evt.Message),
			RetryCount: 3,
			MaxRetries: 3,
		}
		dlq.Add(fe)
		tr.mu.Lock()
		tr.eventLog = append(tr.eventLog, fmt.Sprintf("fail:Notification(%s)", evt.Message))
		tr.mu.Unlock()
	})
}

// registerOrderedHandlers 注册带执行优先级的处理器。
func registerOrderedHandlers(bus *event.EventBusWithOrdering, tr *eventTracker) {
	// ---- 4. Register a handler with ordering (via EventBusWithOrdering) ----
	bus.SubscribeWithConfig("PaymentProcessed", event.ListenerConfig{
		Handler: func(e event.ApplicationEvent) {
			evt := e.(*PaymentProcessedEvent)
			tr.orderedHandled.Add(1)
			tr.mu.Lock()
			tr.eventLog = append(tr.eventLog, fmt.Sprintf("ordered:PaymentProcessed(%s)", evt.OrderID))
			tr.mu.Unlock()
		},
		Order: 10,
	})
	bus.SubscribeWithConfig("PaymentProcessed", event.ListenerConfig{
		Handler: func(e event.ApplicationEvent) {
			tr.mu.Lock()
			tr.eventLog = append(tr.eventLog, fmt.Sprintf("ordered:PaymentAudit(%s)", e.(*PaymentProcessedEvent).OrderID))
			tr.mu.Unlock()
		},
		Order: 20,
	})
}

// publishEvents 发布订单、支付与通知事件。
func publishEvents(bus event.EventBus) {
	// ---- 5. Publish events ----
	fmt.Println("--- Publishing events ---")
	bus.Publish(&OrderCreatedEvent{
		BaseEvent: event.BaseEvent{EventType: "OrderCreated"},
		OrderID:   "ORD-001",
		Amount:    99.99,
	})

	bus.Publish(&PaymentProcessedEvent{
		BaseEvent: event.BaseEvent{EventType: "PaymentProcessed"},
		OrderID:   "ORD-001",
		Status:    "SUCCESS",
	})

	// This will fail and go to DLQ
	bus.Publish(&NotificationEvent{
		BaseEvent: event.BaseEvent{EventType: "Notification"},
		Message:   "Your order has been placed",
	})
}

// publishOrderedEvent 通过有序事件总线发布支付完成事件。
func publishOrderedEvent(bus *event.EventBusWithOrdering) {
	// ---- 6. Publish via EventBusWithOrdering (ordered handlers) ----
	fmt.Println("--- Publishing via ordered bus ---")
	bus.Publish(&PaymentProcessedEvent{
		BaseEvent: event.BaseEvent{EventType: "PaymentProcessed"},
		OrderID:   "ORD-002",
		Status:    "SUCCESS",
	})
}

// printResults 输出各处理器被调用次数。
func printResults(tr *eventTracker) {
	// ---- 7. Verify results ----
	fmt.Println()
	fmt.Println("--- Results ---")
	fmt.Printf("  Sync handlers invoked: %d\n", tr.syncHandled.Load())
	fmt.Printf("  Ordered handlers invoked: %d\n", tr.orderedHandled.Load())
	fmt.Printf("  Failed handlers invoked: %d\n", tr.failedHandled.Load())
}

// printDeadLetterStats 输出死信队列统计信息。
func printDeadLetterStats(dlq *event.DeadLetterQueue) {
	// ---- 8. Check dead letter queue ----
	fmt.Println()
	fmt.Println("--- Dead Letter Queue ---")
	dlqStats := dlq.Stats()
	fmt.Printf("  DLQ size: %d\n", dlqStats.Total)
	fmt.Printf("  Exhausted (max retries): %d\n", dlqStats.Exhausted)
	for eventType, count := range dlqStats.EventTypeCount {
		fmt.Printf("  - %s: %d event(s)\n", eventType, count)
	}
}

// printEventLog 输出事件处理日志。
func printEventLog(tr *eventTracker) {
	// ---- 9. Print event log ----
	fmt.Println()
	fmt.Println("--- Event Processing Log ---")
	for i, entry := range tr.eventLog {
		fmt.Printf("  %d. %s\n", i+1, entry)
	}
}

// demoAsyncPublisher 演示异步事件发布器。
func demoAsyncPublisher() {
	// ---- 10. Demonstrate async publisher ----
	fmt.Println()
	fmt.Println("--- Async Publisher ---")
	asyncBus := event.NewEventBus()
	var asyncCount atomic.Int32
	asyncBus.Subscribe("AsyncEvent", func(e event.ApplicationEvent) {
		asyncCount.Add(1)
		time.Sleep(10 * time.Millisecond)
	})

	publisher := event.NewAsyncPublisher(asyncBus,
		event.WithWorkerCount(3),
	)

	for i := 0; i < 5; i++ {
		publisher.Publish(context.Background(), &event.BaseEvent{
			EventType: "AsyncEvent",
		})
	}
	time.Sleep(200 * time.Millisecond)
	publisher.Close()
	fmt.Printf("  Async events processed: %d\n", asyncCount.Load())
}
