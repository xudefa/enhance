package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestShardedLRUCache_Basic(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestShardedLRUCache_NotFound(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestShardedLRUCache_TTL(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 立即获取应该成功
	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after TTL, got %v", err)
	}
}

func TestShardedLRUCache_Eviction(t *testing.T) {
	t.Parallel()
	// 使用小容量测试淘汰机制
	cache := NewShardedLRUCache(3, 1) // 1 个分片，容量 3
	ctx := context.Background()

	// 添加 3 个项
	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)
	_ = cache.Set(ctx, "key3", "value3", time.Minute)

	// 添加第 4 个项，应该淘汰最久未使用的 key1
	_ = cache.Set(ctx, "key4", "value4", time.Minute)

	_, err := cache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected key1 to be evicted")
	}

	// key2, key3, key4 应该还在
	for _, key := range []string{"key2", "key3", "key4"} {
		_, err := cache.Get(ctx, key)
		if err != nil {
			t.Errorf("expected %s to exist", key)
		}
	}
}

func TestShardedLRUCache_Del(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)

	err := cache.Del(ctx, "key1")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	_, err = cache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected key1 to be deleted")
	}

	// key2 应该还在
	val, err := cache.Get(ctx, "key2")
	if err != nil {
		t.Fatalf("Get key2 failed: %v", err)
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
}

func TestShardedLRUCache_Exists(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", 100*time.Millisecond)

	exists, err := cache.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Errorf("expected key1 to exist")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	exists, err = cache.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed after TTL: %v", err)
	}
	if exists {
		t.Errorf("expected key1 to be expired")
	}
}

func TestShardedLRUCache_TTLRemaining(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Second)

	remaining, err := cache.TTL(ctx, "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if remaining > time.Second || remaining <= 0 {
		t.Errorf("expected TTL around 1s, got %v", remaining)
	}
}

func TestShardedLRUCache_Clear(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("expected cache to be empty after clear")
	}
}

func TestShardedLRUCache_EvictCallback(t *testing.T) {
	t.Parallel()
	evicted := make(map[string]any)
	var mu sync.Mutex
	cache := NewShardedLRUCache(2, 1, WithEvictCallback(func(key string, value any) {
		mu.Lock()
		defer mu.Unlock()
		evicted[key] = value
	}))
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)
	_ = cache.Set(ctx, "key3", "value3", time.Minute) // 应该淘汰 key1

	mu.Lock()
	val, ok := evicted["key1"]
	mu.Unlock()
	if !ok || val != "value1" {
		t.Errorf("expected key1 to be evicted with value1")
	}
}

func TestShardedLRUCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(1000, 16)
	ctx := context.Background()

	// 并发写入
	done := make(chan bool)
	for i := range 100 {
		go func(n int) {
			for j := range 100 {
				key := string(rune('A'+(n%26))) + string(rune('0'+(j%10)))
				_ = cache.Set(ctx, key, j, time.Minute)
			}
			done <- true
		}(i)
	}

	for range 100 {
		<-done
	}

	// 验证没有 panic
	if cache.Len() == 0 {
		t.Error("expected cache to have items")
	}
}

func TestShardedLRUCache_NeverExpire(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	// 设置永不过期的项
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", -1)

	// 立即检查应该存在
	exists1, err := cache.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists1 {
		t.Error("key1 with ttl=0 should exist")
	}

	exists2, err := cache.Exists(ctx, "key2")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists2 {
		t.Error("key2 with ttl=-1 should exist")
	}

	// 等待一段时间后再次检查
	time.Sleep(100 * time.Millisecond)

	exists1, err = cache.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed after sleep: %v", err)
	}
	if !exists1 {
		t.Error("key1 with ttl=0 should still exist after sleep")
	}

	// Get 也应该能获取到
	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed for never-expire key: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestShardedLRUCache_DelMultipleKeys(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)
	_ = cache.Set(ctx, "key3", "value3", time.Minute)

	// 批量删除
	err := cache.Del(ctx, "key1", "key3")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	// key1 和 key3 应该被删除
	_, err = cache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected key1 to be deleted")
	}

	_, err = cache.Get(ctx, "key3")
	if err != ErrNotFound {
		t.Errorf("expected key3 to be deleted")
	}

	// key2 应该还在
	val, err := cache.Get(ctx, "key2")
	if err != nil {
		t.Fatalf("Get key2 failed: %v", err)
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
}

func TestShardedLRUCache_Len(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	if cache.Len() != 0 {
		t.Errorf("expected initial length 0, got %d", cache.Len())
	}

	_ = cache.Set(ctx, "key1", "value1", time.Minute)
	_ = cache.Set(ctx, "key2", "value2", time.Minute)

	if cache.Len() != 2 {
		t.Errorf("expected length 2, got %d", cache.Len())
	}
}

func TestShardedLRUCache_Close(t *testing.T) {
	t.Parallel()
	cache := NewShardedLRUCache(100, 16)
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", time.Minute)

	err := cache.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if cache.Len() != 0 {
		t.Errorf("expected cache to be empty after close")
	}
}

func TestShardedLRUCache_DefaultCapacity(t *testing.T) {
	t.Parallel()
	// 测试默认容量
	cache := NewShardedLRUCache(0, 0)
	if cache.capacity != 1000 {
		t.Errorf("expected default capacity 1000, got %d", cache.capacity)
	}
	if cache.shardCount != 16 {
		t.Errorf("expected default shard count 16, got %d", cache.shardCount)
	}
}
