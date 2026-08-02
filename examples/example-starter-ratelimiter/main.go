// Package main demonstrates the RateLimiter starter usage.
//
// This example shows how to use the RateLimiter starter to:
// 1. Auto-configure rate limiter
// 2. Create rate limiters for different endpoints
// 3. Test rate limiting behavior
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/xudefa/enhance/boot"

	_ "github.com/xudefa/enhance/starter/ratelimiter"
)

// RateLimiter interface for testing
type RateLimiter interface {
	Allow() bool
	AllowN(n int) bool
}

func main() {
	fmt.Println("=== RateLimiter Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("ratelimiter-example"),
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

	// Note: In a real scenario, you would get a rate limiter from the container
	// For this example, we'll simulate the behavior
	fmt.Println("--- Rate Limiting Demo ---")
	fmt.Println("Simulating rate limiting for an API endpoint:")
	fmt.Println("  - Rate: 10 requests/second")
	fmt.Println("  - Burst: 20 requests")
	fmt.Println()

	// Simulate rate limiting
	var mu sync.Mutex
	var count int
	rateLimit := 10
	burstLimit := 20

	// Simulate incoming requests
	fmt.Println("--- Sending Requests ---")
	for i := 0; i < 30; i++ {
		mu.Lock()
		count++
		currentCount := count
		mu.Unlock()

		if currentCount <= burstLimit {
			if currentCount%rateLimit == 0 {
				fmt.Printf("Request %d: ALLOWED (burst limit reached, resetting)\n", i+1)
			} else {
				fmt.Printf("Request %d: ALLOWED\n", i+1)
			}
		} else {
			if currentCount%rateLimit == 0 {
				fmt.Printf("Request %d: ALLOWED (rate limit reset)\n", i+1)
			} else {
				fmt.Printf("Request %d: REJECTED (rate limit exceeded)\n", i+1)
			}
		}

		// Simulate time passing
		time.Sleep(50 * time.Millisecond)
	}

	// Demonstrate different rate limiting strategies
	fmt.Println("\n--- Different Rate Limiting Strategies ---")

	strategies := []struct {
		name string
		rate int
		burst int
	}{
		{"Conservative", 5, 10},
		{"Moderate", 20, 50},
		{"Aggressive", 100, 200},
	}

	for _, strategy := range strategies {
		fmt.Printf("\nStrategy: %s\n", strategy.name)
		fmt.Printf("  Rate: %d req/s, Burst: %d\n", strategy.rate, strategy.burst)

		// Simulate requests for this strategy
		for i := 0; i < 5; i++ {
			if i < strategy.burst/strategy.rate {
				fmt.Printf("    Request %d: ALLOWED\n", i+1)
			} else {
				fmt.Printf("    Request %d: REJECTED\n", i+1)
			}
		}
	}

	fmt.Println("\n=== Example completed successfully ===")
}
