# example-cache-lru

Demonstrates the enhance LRU cache features.

## Features Demonstrated

- **LRU cache creation** — with different configurations
- **Cache operations** — Get, Set, Delete, Exists
- **TTL expiration** — automatic key expiration
- **Eviction callback** — notified when items are evicted
- **Cache statistics** — Len for tracking cache size
- **MemoryCacheBuilder** — builder pattern for cache creation
- **CacheHelper** — GetOrSet pattern for cache-aside
- **CacheTemplate** — key-prefix based cache operations
- **Concurrent access safety** — goroutine-safe cache operations
- **LRU eviction order** — least recently used item evicted first

## Run

```bash
go run .
```

## Expected Output

```
=== enhance LRU Cache Example ===

--- 1. Basic LRU Cache ---
  name=enhance, version=1.0.0

--- 2. Cache with Eviction Callback ---
  [evict] key=a, value=1
  Evicted keys: [a]

--- 3. TTL Expiration ---
  Before expiry: temp=value, err=<nil>
  After expiry: temp=<nil>, err=cache: key not found
  TTL remaining for 'ttl-key': 5m0s

--- 4. Exists and Delete ---
  key1 exists: true
  key1 after delete: false

--- 5. Cache Statistics ---
  Cache length: 50

--- 6. MemoryCacheBuilder ---
  Builder cache: builder-key=builder-value

--- 7. CacheHelper (GetOrSet) ---
  [loader] Computing expensive data...
  GetOrSet result: computed-value
  GetOrSet cached: computed-value

--- 8. CacheTemplate ---
  Template get user:1 = Alice

--- 9. Concurrent Access Test ---
  200 concurrent operations completed, errors: 0
  Final cache length: 50

--- 10. LRU Eviction Order ---
  x exists: true (accessed before eviction)
  y exists: false (should be evicted)
  z exists: true
  w exists: true (newest)

=== Example completed successfully ===
```
