// Package main demonstrates the Cron starter usage.
//
// This example shows how to use the Cron starter to:
// 1. Auto-configure Cron scheduler
// 2. Register cron jobs
// 3. Start and stop the scheduler
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/cron"
)

func main() {
	fmt.Println("=== Cron Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("cron-example"),
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

	// Get the Cron scheduler from container
	cronScheduler, err := core.GetByName[*cron.Cron](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get cron scheduler: %v\n", err)
		return
	}

	// Register cron jobs
	fmt.Println("--- Registering Cron Jobs ---")

	// Job 1: Run every 5 seconds
	job1ID, err := cronScheduler.AddFunc("*/5 * * * * *", func() {
		fmt.Printf("[%s] Job 1 executed (every 5 seconds)\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		fmt.Printf("Failed to add job 1: %v\n", err)
		return
	}
	fmt.Printf("Job 1 registered with ID: %d\n", job1ID)

	// Job 2: Run every 10 seconds
	job2ID, err := cronScheduler.AddFunc("*/10 * * * * *", func() {
		fmt.Printf("[%s] Job 2 executed (every 10 seconds)\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		fmt.Printf("Failed to add job 2: %v\n", err)
		return
	}
	fmt.Printf("Job 2 registered with ID: %d\n", job2ID)

	// Job 3: Run at specific time (every minute at :00)
	job3ID, err := cronScheduler.AddFunc("0 * * * * *", func() {
		fmt.Printf("[%s] Job 3 executed (every minute at :00)\n", time.Now().Format("15:04:05"))
	})
	if err != nil {
		fmt.Printf("Failed to add job 3: %v\n", err)
		return
	}
	fmt.Printf("Job 3 registered with ID: %d\n", job3ID)

	// Start the cron scheduler
	fmt.Println("\n--- Starting Cron Scheduler ---")
	cronScheduler.Start()
	fmt.Println("Cron scheduler started")

	// List registered jobs
	fmt.Println("\n--- Registered Jobs ---")
	for _, entry := range cronScheduler.Entries() {
		fmt.Printf("Job ID: %d, Next: %v\n", entry.ID, entry.Next)
	}

	// Wait for some jobs to execute
	fmt.Println("\n--- Waiting for Jobs to Execute (30 seconds) ---")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for signal
	app.WaitForSignal()
}
