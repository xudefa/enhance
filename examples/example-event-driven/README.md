# example-event-driven

Demonstrates the enhance event-driven architecture.

## Features Demonstrated

- **EventBus creation** — creating a basic event bus
- **Event definition and publishing** — custom event types with BaseEvent embedding
- **Synchronous event handling** — multiple listeners per event type
- **Dead letter queue** — capturing failed events with retry metadata
- **Event ordering** — EventBusWithOrdering with priority-based listener execution
- **Async publisher** — worker pool for asynchronous event processing

## Run

```bash
go run .
```

## Expected Output

```
=== enhance Event-Driven Architecture Example ===

--- Publishing events ---
--- Publishing via ordered bus ---

--- Results ---
  Sync handlers invoked: 2
  Ordered handlers invoked: 1
  Failed handlers invoked: 1

--- Dead Letter Queue ---
  DLQ size: 1
  Exhausted (max retries): 1
  - Notification: 1 event(s)

--- Event Processing Log ---
  1. sync:OrderCreated(ORD-001, 100)
  2. sync:PaymentProcessed(ORD-001, SUCCESS)
  3. fail:Notification(Your order has been placed)
  4. ordered:PaymentProcessed(ORD-002)
  5. ordered:PaymentAudit(ORD-002)

--- Async Publisher ---
  Async events processed: 5

=== Example completed successfully ===
```
