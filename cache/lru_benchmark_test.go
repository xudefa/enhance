package cache

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		_ = cache.Set(ctx, key, i, time.Minute)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000)
	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = cache.Set(ctx, key, i, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		_, _ = cache.Get(ctx, key)
	}
}

func BenchmarkLRUCache_Concurrent(b *testing.B) {
	cache := NewLRUCache(1000)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", seq%100)
			_ = cache.Set(ctx, key, seq, time.Minute)
			_, _ = cache.Get(ctx, key)
			seq++
		}
	})
}

func BenchmarkShardedLRUCache_Set(b *testing.B) {
	cache := NewShardedLRUCache(1000, 16)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		_ = cache.Set(ctx, key, i, time.Minute)
	}
}

func BenchmarkShardedLRUCache_Get(b *testing.B) {
	cache := NewShardedLRUCache(1000, 16)
	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = cache.Set(ctx, key, i, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		_, _ = cache.Get(ctx, key)
	}
}

func BenchmarkShardedLRUCache_Concurrent(b *testing.B) {
	cache := NewShardedLRUCache(1000, 16)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", seq%100)
			_ = cache.Set(ctx, key, seq, time.Minute)
			_, _ = cache.Get(ctx, key)
			seq++
		}
	})
}

func benchLRUCacheComparisonRun(b *testing.B, shards int) {
	var cache Cache
	if shards == 1 {
		cache = NewLRUCache(1000)
	} else {
		cache = NewShardedLRUCache(1000, shards)
	}
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", seq%100)
			_ = cache.Set(ctx, key, seq, time.Minute)
			_, _ = cache.Get(ctx, key)
			seq++
		}
	})
}

func BenchmarkLRUCache_Comparison(b *testing.B) {
	b.Run("Single-Shard", func(b *testing.B) { benchLRUCacheComparisonRun(b, 1) })
	b.Run("Sharded-16", func(b *testing.B) { benchLRUCacheComparisonRun(b, 16) })
	b.Run("Sharded-32", func(b *testing.B) { benchLRUCacheComparisonRun(b, 32) })
	b.Run("Sharded-64", func(b *testing.B) { benchLRUCacheComparisonRun(b, 64) })
}
