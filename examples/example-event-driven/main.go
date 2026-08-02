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

func main() {
	fmt.Println("=== enhance Event-Driven Architecture Example ===")
	fmt.Println()

	// ---- 1. Create EventBus and Dead Letter Queue ----
	bus := event.NewEventBus()
	dlq := event.NewDeadLetterQueue()

	// Track events for verification
	var syncHandled atomic.Int32
	var orderedHandled atomic.Int32
	var failedHandled atomic.Int32
	var mu sync.Mutex
	eventLog := make([]string, 0)

	// ---- 2. Register sync handlers ----
	bus.Subscribe("OrderCreated", func(e event.ApplicationEvent) {
		evt := e.(*OrderCreatedEvent)
		syncHandled.Add(1)
		mu.Lock()
		eventLog = append(eventLog, fmt.Sprintf("sync:OrderCreated(%s, %.0f)", evt.OrderID, evt.Amount))
		mu.Unlock()
	})

	bus.Subscribe("PaymentProcessed", func(e event.ApplicationEvent) {
		evt := e.(*PaymentProcessedEvent)
		syncHandled.Add(1)
		mu.Lock()
		eventLog = append(eventLog, fmt.Sprintf("sync:PaymentProcessed(%s, %s)", evt.OrderID, evt.Status))
		mu.Unlock()
	})

	// ---- 3. Register a handler that always fails (for DLQ demo) ----
	bus.Subscribe("Notification", func(e event.ApplicationEvent) {
		evt := e.(*NotificationEvent)
		failedHandled.Add(1)
		// Simulate a processing failure
		fe := event.FailedEvent{
			Event:      evt,
			Err:        fmt.Errorf("notification delivery failed for: %s", evt.Message),
			RetryCount: 3,
			MaxRetries: 3,
		}
		dlq.Add(fe)
		mu.Lock()
		eventLog = append(eventLog, fmt.Sprintf("fail:Notification(%s)", evt.Message))
		mu.Unlock()
	})

	// ---- 4. Register a handler with ordering (via EventBusWithOrdering) ----
	busWithOrder := event.NewEventBusWithOrdering()
	busWithOrder.SubscribeWithConfig("PaymentProcessed", event.ListenerConfig{
		Handler: func(e event.ApplicationEvent) {
			evt := e.(*PaymentProcessedEvent)
			orderedHandled.Add(1)
			mu.Lock()
			eventLog = append(eventLog, fmt.Sprintf("ordered:PaymentProcessed(%s)", evt.OrderID))
			mu.Unlock()
		},
		Order: 10,
	})
	busWithOrder.SubscribeWithConfig("PaymentProcessed", event.ListenerConfig{
		Handler: func(e event.ApplicationEvent) {
			mu.Lock()
			eventLog = append(eventLog, fmt.Sprintf("ordered:PaymentAudit(%s)", e.(*PaymentProcessedEvent).OrderID))
			mu.Unlock()
		},
		Order: 20,
	})

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

	// ---- 6. Publish via EventBusWithOrdering (ordered handlers) ----
	fmt.Println("--- Publishing via ordered bus ---")
	busWithOrder.Publish(&PaymentProcessedEvent{
		BaseEvent: event.BaseEvent{EventType: "PaymentProcessed"},
		OrderID:   "ORD-002",
		Status:    "SUCCESS",
	})

	// ---- 7. Verify results ----
	fmt.Println()
	fmt.Println("--- Results ---")
	fmt.Printf("  Sync handlers invoked: %d\n", syncHandled.Load())
	fmt.Printf("  Ordered handlers invoked: %d\n", orderedHandled.Load())
	fmt.Printf("  Failed handlers invoked: %d\n", failedHandled.Load())

	// ---- 8. Check dead letter queue ----
	fmt.Println()
	fmt.Println("--- Dead Letter Queue ---")
	dlqStats := dlq.Stats()
	fmt.Printf("  DLQ size: %d\n", dlqStats.Total)
	fmt.Printf("  Exhausted (max retries): %d\n", dlqStats.Exhausted)
	for eventType, count := range dlqStats.EventTypeCount {
		fmt.Printf("  - %s: %d event(s)\n", eventType, count)
	}

	// ---- 9. Print event log ----
	fmt.Println()
	fmt.Println("--- Event Processing Log ---")
	for i, entry := range eventLog {
		fmt.Printf("  %d. %s\n", i+1, entry)
	}

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

	fmt.Println()
	fmt.Println("=== Example completed successfully ===")
}
