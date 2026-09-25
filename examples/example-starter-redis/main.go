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

	app := newApp()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Redis connection failed: %v\n", err)
		fmt.Println("This example requires a running Redis server.")
		fmt.Println("Please start Redis and try again.")
		return
	}

	ctx := context.Background()
	redisCache, err := getCache(app)
	if err != nil {
		return
	}

	// 尝试获取 CacheInspector 接口（支持 Exists 操作）
	cacheInspector, ok := redisCache.(cache.CacheInspector)
	if !ok {
		fmt.Println("Warning: cache does not support Exists operation")
	}

	if err := demoBasicOps(ctx, redisCache, cacheInspector); err != nil {
		return
	}
	if err := demoMultipleValues(ctx, redisCache, cacheInspector); err != nil {
		return
	}

	fmt.Println("\n=== Example completed successfully ===")
}

// newApp 创建应用实例，配置名称与激活的 Profile。
func newApp() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("redis-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// getCache 从容器的 Bean 中获取 Redis 缓存实例。
func getCache(app *boot.Boot) (cache.Cache, error) {
	redisCache, err := core.GetByName[cache.Cache](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get cache: %v\n", err)
		return nil, fmt.Errorf("Failed to get cache: %w", err)
	}
	return redisCache, nil
}

// demoBasicOps 演示单个键的写入、读取与存在性检查。
func demoBasicOps(ctx context.Context, redisCache cache.Cache, inspector cache.CacheInspector) error {
	// Demonstrate cache operations
	fmt.Println("--- Cache Operations ---")

	// Set a value
	key := "example:key"
	value := "Hello from Redis!"
	ttl := 5 * time.Minute

	fmt.Printf("Setting key: %s = %s (TTL: %v)\n", key, value, ttl)
	if err := redisCache.Set(ctx, key, value, ttl); err != nil {
		fmt.Printf("Set failed: %v\n", err)
		return fmt.Errorf("Set failed: %w", err)
	}

	// Get the value
	retrieved, err := redisCache.Get(ctx, key)
	if err != nil {
		fmt.Printf("Get failed: %v\n", err)
		return fmt.Errorf("Get failed: %w", err)
	}
	fmt.Printf("Got value: %v\n", retrieved)

	// Check if key exists
	if inspector != nil {
		exists, err := inspector.Exists(ctx, key)
		if err != nil {
			fmt.Printf("Exists failed: %v\n", err)
			return fmt.Errorf("Exists failed: %w", err)
		}
		fmt.Printf("Key exists: %v\n", exists)
	}
	return nil
}

// demoMultipleValues 演示多个键的写入、读取、删除与删除验证。
func demoMultipleValues(ctx context.Context, redisCache cache.Cache, inspector cache.CacheInspector) error {
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
		cacheVal, err := redisCache.Get(ctx, k)
		if err != nil {
			fmt.Printf("Get %s failed: %v\n", k, err)
			continue
		}
		fmt.Printf("Get: %s = %v\n", k, cacheVal)
	}

	// Delete a key
	fmt.Println("\n--- Delete ---")
	if err := redisCache.Del(ctx, "user:1"); err != nil {
		fmt.Printf("Del failed: %v\n", err)
		return fmt.Errorf("Del failed: %w", err)
	}
	fmt.Println("Deleted: user:1")

	// Verify deletion
	if inspector != nil {
		exists, err := inspector.Exists(ctx, "user:1")
		if err != nil {
			fmt.Printf("Exists check failed: %v\n", err)
			return fmt.Errorf("Exists check failed: %w", err)
		}
		fmt.Printf("user:1 exists after deletion: %v\n", exists)
	}
	return nil
}
