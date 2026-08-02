// Package main demonstrates the Asynq starter usage.
//
// This example shows how to use the Asynq starter to:
// 1. Auto-configure Asynq client and server
// 2. Create and enqueue tasks
// 3. Process tasks
//
// Prerequisites:
// - Redis server running on localhost:6379
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/asynq"
)

const (
	TaskEmailDelivery = "email:delivery"
	TaskImageResize   = "image:resize"
)

func main() {
	fmt.Println("=== Asynq Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("asynq-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Asynq connection failed: %v\n", err)
		fmt.Println("This example requires a running Redis server.")
		fmt.Println("Please start Redis and try again.")
		return
	}

	// Get the Asynq client from container
	client, err := core.GetByName[*asynq.Client](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get asynq client: %v\n", err)
		return
	}
	defer client.Close()

	// Demo 1: Enqueue a simple task
	fmt.Println("--- Demo 1: Enqueue Simple Task ---")
	task := asynq.NewTask(TaskEmailDelivery, []byte(`{"user_id": 123, "subject": "Welcome!"}`))
	info, err := client.Enqueue(task,
		asynq.Queue("email"),
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
	)
	if err != nil {
		fmt.Printf("Failed to enqueue task: %v\n", err)
		return
	}
	fmt.Printf("Task enqueued: ID=%s, Queue=%s\n", info.ID, info.Queue)

	// Demo 2: Enqueue a task with delay
	fmt.Println("\n--- Demo 2: Enqueue Delayed Task ---")
	delayedTask := asynq.NewTask(TaskEmailDelivery, []byte(`{"user_id": 456, "subject": "Reminder"}`))
	delayedInfo, err := client.Enqueue(delayedTask,
		asynq.Queue("email"),
		asynq.ProcessIn(10*time.Second),
	)
	if err != nil {
		fmt.Printf("Failed to enqueue delayed task: %v\n", err)
		return
	}
	fmt.Printf("Delayed task enqueued: ID=%s, Delay=10s\n", delayedInfo.ID)

	// Demo 3: Enqueue a high-priority task
	fmt.Println("\n--- Demo 3: Enqueue High-Priority Task ---")
	highPriorityTask := asynq.NewTask(TaskEmailDelivery, []byte(`{"user_id": 789, "subject": "Urgent!"}`))
	highPriorityInfo, err := client.Enqueue(highPriorityTask,
		asynq.Queue("critical"),
	)
	if err != nil {
		fmt.Printf("Failed to enqueue high-priority task: %v\n", err)
		return
	}
	fmt.Printf("High-priority task enqueued: ID=%s, Priority=8\n", highPriorityInfo.ID)

	// Demo 4: Enqueue multiple tasks
	fmt.Println("\n--- Demo 4: Enqueue Multiple Tasks ---")
	for i := 0; i < 5; i++ {
		task := asynq.NewTask(TaskImageResize, []byte(fmt.Sprintf(`{"image_id": %d, "width": 800, "height": 600}`, i)))
		info, err := client.Enqueue(task,
			asynq.Queue("images"),
		)
		if err != nil {
			fmt.Printf("Failed to enqueue task %d: %v\n", i, err)
			continue
		}
		fmt.Printf("Task %d enqueued: ID=%s\n", i, info.ID)
	}

	fmt.Println("\n=== Example completed successfully ===")
	fmt.Println("\nNote: Tasks are enqueued but not processed in this example.")
	fmt.Println("To process tasks, you would need to start an Asynq worker.")
}
