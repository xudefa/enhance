// Package main demonstrates the Prometheus starter usage.
//
// This example shows how to use the Prometheus starter to:
// 1. Auto-configure Prometheus metrics
// 2. Register and use metrics
// 3. Expose metrics endpoint
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl http://localhost:9090/metrics
package main

import (
	"fmt"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/metrics"

	_ "github.com/xudefa/enhance/starter/prometheus"
)

func main() {
	fmt.Println("=== Prometheus Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("prometheus-example"),
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

	// Get the metrics registry from container
	registry, err := core.GetByName[metrics.MeterRegistry](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get metrics registry: %v\n", err)
		return
	}

	// Demo 1: Counter
	fmt.Println("--- Demo 1: Counter ---")
	requestCounter := registry.Counter("http_requests_total", "method", "GET", "status", "200")
	requestCounter.Inc()
	requestCounter.Add(10)
	fmt.Printf("Counter incremented: %v\n", requestCounter.Value())

	// Demo 2: Gauge
	fmt.Println("\n--- Demo 2: Gauge ---")
	memoryGauge := registry.Gauge("memory_usage_bytes")
	memoryGauge.Set(1024 * 1024 * 100) // 100 MB
	fmt.Printf("Gauge set: %v\n", memoryGauge.Value())

	// Demo 3: Histogram
	fmt.Println("\n--- Demo 3: Histogram ---")
	requestDuration := registry.Histogram("http_request_duration_seconds")
	requestDuration.Record(0.1) // 100ms
	requestDuration.Record(0.2) // 200ms
	requestDuration.Record(0.15) // 150ms
	fmt.Printf("Histogram observations: %v\n", requestDuration.Count())

	// Demo 4: Collect all metrics
	fmt.Println("\n--- Demo 4: Collect All Metrics ---")
	allMetrics := registry.Collect()
	fmt.Printf("Total metrics: %d\n", len(allMetrics))
	for _, m := range allMetrics {
		fmt.Printf("  - %s: %v (tags: %v)\n", m.Name, m.Value, m.Tags)
	}

	// Simulate some activity
	fmt.Println("\n--- Simulating Activity ---")
	for i := 0; i < 5; i++ {
		requestCounter.Inc()
		memoryGauge.Set(float64(1024*1024*100 + i*1024*1024))
		time.Sleep(100 * time.Millisecond)
	}

	// Final metrics collection
	fmt.Println("\n--- Final Metrics ---")
	finalMetrics := registry.Collect()
	for _, m := range finalMetrics {
		fmt.Printf("  - %s: %v\n", m.Name, m.Value)
	}

	fmt.Println("\n=== Example completed successfully ===")
	fmt.Println("Metrics endpoint: http://localhost:9090/metrics")
}
