package main

import (
	"fmt"
	"time"

	"github.com/xudefa/enhance/resilience"
)

func main() {
	fmt.Println("=== Resilience Circuit Breaker Demo ===")

	breaker := resilience.NewBreaker(
		resilience.WithMaxRequests(3),
		resilience.WithErrorThreshold(0.5),
		resilience.WithWaitDuration(2*time.Second),
	)

	fmt.Println("\n1. Initial state:")
	fmt.Printf("   State: %s\n", breaker.State())

	fmt.Println("\n2. Sending requests (circuit CLOSED):")
	for i := range 5 {
		if err := breaker.Allow(); err != nil {
			fmt.Printf("   Request %d: BLOCKED (%v)\n", i+1, err)
			continue
		}
		fmt.Printf("   Request %d: allowed (CLOSED)\n", i+1)
		breaker.RecordSuccess()
	}

	fmt.Println("\n3. Simulating failures:")
	for i := range 5 {
		if err := breaker.Allow(); err != nil {
			fmt.Printf("   Request %d: BLOCKED (%v)\n", i+1, err)
			break
		}
		fmt.Printf("   Request %d: allowed, recording failure\n", i+1)
		breaker.RecordFailure()
	}

	fmt.Println("\n4. State after failures:")
	fmt.Printf("   State: %s\n", breaker.State())

	fmt.Println("\n5. Requests while OPEN:")
	for i := range 3 {
		if err := breaker.Allow(); err != nil {
			fmt.Printf("   Request %d: BLOCKED (%v)\n", i+1, err)
		}
	}

	fmt.Println("\n6. Waiting for half-open...")
	time.Sleep(2100 * time.Millisecond)

	fmt.Println("\n7. Sending probe request (HALF-OPEN):")
	if err := breaker.Allow(); err != nil {
		fmt.Printf("   Probe: BLOCKED (%v)\n", err)
	} else {
		fmt.Println("   Probe: allowed (HALF-OPEN)")
		breaker.RecordSuccess()
	}
	fmt.Printf("   State after success: %s\n", breaker.State())

	fmt.Println("\n8. Resume normal traffic:")
	for i := range 3 {
		if err := breaker.Allow(); err != nil {
			fmt.Printf("   Request %d: BLOCKED\n", i+1)
			continue
		}
		fmt.Printf("   Request %d: allowed (CLOSED)\n", i+1)
		breaker.RecordSuccess()
	}

	fmt.Println("\n=== Circuit Breaker Demo Complete ===")
}
