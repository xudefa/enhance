// Package main demonstrates the Redis starter usage.
//
// This example shows how to use the Redis starter to:
// 1. Auto-configure Redis connection
// 2. Use Redis as a cache
// 3. Perform basic cache operations
//
// Prerequisites:
// - Redis server running on localhost:6379
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/cache"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/redis"
)

func main() {
	fmt.Println("=== Redis Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("redis-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Redis connection failed: %v\n", err)
		fmt.Println("This example requires a running Redis server.")
		fmt.Println("Please start Redis and try again.")
		return
	}

	// Get the Redis cache from container
	ctx := context.Background()
	redisCache, err := core.GetByName[cache.Cache](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get cache: %v\n", err)
		return
	}

	// Demonstrate cache operations
	fmt.Println("--- Cache Operations ---")

	// Set a value
	key := "example:key"
	value := "Hello from Redis!"
	ttl := 5 * time.Minute

	fmt.Printf("Setting key: %s = %s (TTL: %v)\n", key, value, ttl)
	if err := redisCache.Set(ctx, key, value, ttl); err != nil {
		fmt.Printf("Set failed: %v\n", err)
		return
	}

	// Get the value
	retrieved, err := redisCache.Get(ctx, key)
	if err != nil {
		fmt.Printf("Get failed: %v\n", err)
		return
	}
	fmt.Printf("Got value: %v\n", retrieved)

	// Check if key exists
	exists, err := redisCache.Exists(ctx, key)
	if err != nil {
		fmt.Printf("Exists failed: %v\n", err)
		return
	}
	fmt.Printf("Key exists: %v\n", exists)

	// Set multiple values
	fmt.Println("\n--- Multiple Values ---")
	users := map[string]string{
		"user:1": "Alice",
		"user:2": "Bob",
		"user:3": "Charlie",
	}

	for k, v := range users {
		if err := redisCache.Set(ctx, k, v, 10*time.Minute); err != nil {
			fmt.Printf("Failed to set %s: %v\n", k, err)
			continue
		}
		fmt.Printf("Set: %s = %s\n", k, v)
	}

	// Get all values
	for k := range users {
		val, err := redisCache.Get(ctx, k)
		if err != nil {
			fmt.Printf("Get %s failed: %v\n", k, err)
			continue
		}
		fmt.Printf("Get: %s = %v\n", k, val)
	}

	// Delete a key
	fmt.Println("\n--- Delete ---")
	if err := redisCache.Del(ctx, "user:1"); err != nil {
		fmt.Printf("Del failed: %v\n", err)
		return
	}
	fmt.Println("Deleted: user:1")

	// Verify deletion
	exists, err = redisCache.Exists(ctx, "user:1")
	if err != nil {
		fmt.Printf("Exists check failed: %v\n", err)
		return
	}
	fmt.Printf("user:1 exists after deletion: %v\n", exists)

	fmt.Println("\n=== Example completed successfully ===")
}
