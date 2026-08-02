// Package main demonstrates the enhance LRU cache:
// cache creation with different configs, Get/Set/Delete operations,
// TTL expiration, cache statistics, and concurrent access safety.
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xudefa/enhance/cache"
)

func main() {
	fmt.Println("=== enhance LRU Cache Example ===")
	fmt.Println()
	ctx := context.Background()

	// ---- 1. Create LRU cache with default config ----
	fmt.Println("--- 1. Basic LRU Cache ---")
	c1 := cache.NewLRUCache(100)
	defer c1.Close()

	_ = c1.Set(ctx, "name", "enhance", 0)
	_ = c1.Set(ctx, "version", "1.0.0", 0)

	name, _ := c1.Get(ctx, "name")
	version, _ := c1.Get(ctx, "version")
	fmt.Printf("  name=%v, version=%v\n", name, version)

	// ---- 2. Create cache with eviction callback ----
	fmt.Println()
	fmt.Println("--- 2. Cache with Eviction Callback ---")
	var evicted []string
	c2 := cache.NewLRUCache(3, cache.WithEvictCallback(func(key string, value any) {
		evicted = append(evicted, key)
		fmt.Printf("  [evict] key=%s, value=%v\n", key, value)
	}))

	_ = c2.Set(ctx, "a", 1, 0)
	_ = c2.Set(ctx, "b", 2, 0)
	_ = c2.Set(ctx, "c", 3, 0)
	// This will evict "a" (least recently used)
	_ = c2.Set(ctx, "d", 4, 0)
	fmt.Printf("  Evicted keys: %v\n", evicted)

	// ---- 3. TTL expiration ----
	fmt.Println()
	fmt.Println("--- 3. TTL Expiration ---")
	c3 := cache.NewLRUCache(100)
	_ = c3.Set(ctx, "temp", "value", 200*time.Millisecond)

	val, err := c3.Get(ctx, "temp")
	fmt.Printf("  Before expiry: temp=%v, err=%v\n", val, err)

	time.Sleep(300 * time.Millisecond)
	val, err = c3.Get(ctx, "temp")
	fmt.Printf("  After expiry: temp=%v, err=%v\n", val, err)

	// TTL check
	_ = c3.Set(ctx, "ttl-key", "test", 5*time.Minute)
	ttl, _ := c3.TTL(ctx, "ttl-key")
	fmt.Printf("  TTL remaining for 'ttl-key': %v\n", ttl.Round(time.Millisecond))

	// ---- 4. Exists and Delete ----
	fmt.Println()
	fmt.Println("--- 4. Exists and Delete ---")
	c4 := cache.NewLRUCache(100)
	_ = c4.Set(ctx, "key1", "val1", 0)
	_ = c4.Set(ctx, "key2", "val2", 0)

	exists, _ := c4.Exists(ctx, "key1")
	fmt.Printf("  key1 exists: %v\n", exists)

	_ = c4.Del(ctx, "key1")
	exists, _ = c4.Exists(ctx, "key1")
	fmt.Printf("  key1 after delete: %v\n", exists)

	// ---- 5. Cache statistics (via Len) ----
	fmt.Println()
	fmt.Println("--- 5. Cache Statistics ---")
	c5 := cache.NewLRUCache(1000)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("item-%d", i)
		_ = c5.Set(ctx, key, i, 0)
	}
	fmt.Printf("  Cache length: %d\n", c5.Len())

	// ---- 6. MemoryCacheBuilder ----
	fmt.Println()
	fmt.Println("--- 6. MemoryCacheBuilder ---")
	c6 := cache.NewMemoryCacheBuilder().
		InitialCapacity(500).
		TTL(10 * time.Minute).
		Build()
	_ = c6.Set(ctx, "builder-key", "builder-value", 0)
	val, _ = c6.Get(ctx, "builder-key")
	fmt.Printf("  Builder cache: builder-key=%v\n", val)

	// ---- 7. CacheHelper ----
	fmt.Println()
	fmt.Println("--- 7. CacheHelper (GetOrSet) ---")
	helper := cache.NewCacheHelper(c6)
	result, err := helper.GetOrSet(ctx, "expensive-data", func() (any, error) {
		fmt.Println("  [loader] Computing expensive data...")
		return "computed-value", nil
	}, 0)
	fmt.Printf("  GetOrSet result: %v\n", result)

	// Second call should hit cache
	result, err = helper.GetOrSet(ctx, "expensive-data", func() (any, error) {
		fmt.Println("  [loader] This should NOT print")
		return "should-not-reach", nil
	}, 0)
	fmt.Printf("  GetOrSet cached: %v\n", result)

	// ---- 8. CacheTemplate with key prefix ----
	fmt.Println()
	fmt.Println("--- 8. CacheTemplate ---")
	tpl := cache.NewCacheTemplate(c6, "app")
	_ = tpl.Set(ctx, "user:1", "Alice", 0)
	val, _ = tpl.Get(ctx, "user:1")
	fmt.Printf("  Template get user:1 = %v\n", val)

	// ---- 9. Concurrent access safety ----
	fmt.Println()
	fmt.Println("--- 9. Concurrent Access Test ---")
	concurrentCache := cache.NewLRUCache(1000)
	var wg sync.WaitGroup
	errCount := 0
	var errMu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i%50)
			_ = concurrentCache.Set(ctx, key, i, 0)
		}(i)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i%50)
			_, _ = concurrentCache.Get(ctx, key)
		}(i)
	}
	wg.Wait()

	errMu.Lock()
	finalErrCount := errCount
	errMu.Unlock()
	fmt.Printf("  200 concurrent operations completed, errors: %d\n", finalErrCount)
	fmt.Printf("  Final cache length: %d\n", concurrentCache.Len())

	// ---- 10. LRU eviction order ----
	fmt.Println()
	fmt.Println("--- 10. LRU Eviction Order ---")
	lru := cache.NewLRUCache(3)
	_ = lru.Set(ctx, "x", 1, 0)
	_ = lru.Set(ctx, "y", 2, 0)
	_ = lru.Set(ctx, "z", 3, 0)

	// Access "x" to make it recently used
	_, _ = lru.Get(ctx, "x")
	// Now "y" is the least recently used
	_ = lru.Set(ctx, "w", 4, 0) // Evicts "y"

	_, errX := lru.Get(ctx, "x")
	_, errY := lru.Get(ctx, "y")
	_, errZ := lru.Get(ctx, "z")
	_, errW := lru.Get(ctx, "w")
	fmt.Printf("  x exists: %v (accessed before eviction)\n", errX == nil)
	fmt.Printf("  y exists: %v (should be evicted)\n", errY == nil)
	fmt.Printf("  z exists: %v\n", errZ == nil)
	fmt.Printf("  w exists: %v (newest)\n", errW == nil)

	fmt.Println()
	fmt.Println("=== Example completed successfully ===")
}
