// Package main demonstrates the OpenTelemetry starter usage.
//
// This example shows how to use the OTel starter to:
// 1. Auto-configure OpenTelemetry
// 2. Create spans
// 3. Add attributes and events to spans
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/otel"
)

func main() {
	fmt.Println("=== OpenTelemetry Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("otel-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	// Get the tracer from container
	tracerProvider, err := core.GetByName[trace.TracerProvider](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get tracer provider: %v\n", err)
		return
	}

	// Get a tracer
	tracer := tracerProvider.Tracer("example-tracer")

	// Demo 1: Simple span
	fmt.Println("--- Demo 1: Simple Span ---")
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "example-operation")
	defer span.End()

	fmt.Println("Span created: example-operation")
	span.AddEvent("Operation started")
	time.Sleep(100 * time.Millisecond)
	span.AddEvent("Operation completed")
	fmt.Println("Span ended")

	// Demo 2: Span with attributes
	fmt.Println("\n--- Demo 2: Span with Attributes ---")
	_, span2 := tracer.Start(ctx, "user-operation")
	defer span2.End()

	span2.SetAttributes(
		attribute.String("user.id", "user-123"),
		attribute.String("user.name", "John Doe"),
		attribute.Int("user.age", 30),
		attribute.Bool("user.active", true),
	)

	fmt.Println("Span with attributes created")
	fmt.Println("Attributes:")
	fmt.Println("  - user.id: user-123")
	fmt.Println("  - user.name: John Doe")
	fmt.Println("  - user.age: 30")
	fmt.Println("  - user.active: true")

	time.Sleep(100 * time.Millisecond)

	// Demo 3: Span with events
	fmt.Println("\n--- Demo 3: Span with Events ---")
	_, span3 := tracer.Start(ctx, "database-operation")
	defer span3.End()

	span3.AddEvent("Query started", trace.WithAttributes(
		attribute.String("db.statement", "SELECT * FROM users WHERE id = ?"),
		attribute.String("db.system", "postgresql"),
	))

	time.Sleep(50 * time.Millisecond)

	span3.AddEvent("Query completed", trace.WithAttributes(
		attribute.Int("db.rows_affected", 1),
		attribute.Float64("db.duration_ms", 50.5),
	))

	fmt.Println("Span with events created")
	fmt.Println("Events:")
	fmt.Println("  - Query started")
	fmt.Println("  - Query completed")

	// Demo 4: Span with status
	fmt.Println("\n--- Demo 4: Span with Status ---")
	_, span4 := tracer.Start(ctx, "error-operation")
	defer span4.End()

	span4.SetStatus(codes.Error, "Operation failed")
	span4.RecordError(fmt.Errorf("database connection timeout"))

	fmt.Println("Span with error status created")
	fmt.Println("Status: ERROR - Operation failed")

	// Demo 5: Nested spans
	fmt.Println("\n--- Demo 5: Nested Spans ---")
	ctx5, parentSpan := tracer.Start(ctx, "parent-operation")
	defer parentSpan.End()

	// Child span 1
	_, childSpan1 := tracer.Start(ctx5, "child-operation-1")
	time.Sleep(50 * time.Millisecond)
	childSpan1.End()

	// Child span 2
	_, childSpan2 := tracer.Start(ctx5, "child-operation-2")
	time.Sleep(50 * time.Millisecond)
	childSpan2.End()

	fmt.Println("Nested spans created:")
	fmt.Println("  - parent-operation")
	fmt.Println("    - child-operation-1")
	fmt.Println("    - child-operation-2")

	fmt.Println("\n=== Example completed successfully ===")
}
