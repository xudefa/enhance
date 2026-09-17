package cache

import (
	"context"
	"testing"
	"time"
)

func TestCacheTemplate_Key(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	key := template.Key("user:123")
	if key != "myapp:user:123" {
		t.Errorf("expected key 'myapp:user:123', got %s", key)
	}
}

func TestCacheTemplate_KeyNoPrefix(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "")

	key := template.Key("user:123")
	if key != "user:123" {
		t.Errorf("expected key 'user:123', got %s", key)
	}
}

func TestCacheTemplate_GetSet(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	// Set
	err := template.Set(ctx, "test-key", "test-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	// Get
	value, err := template.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if value != "test-value" {
		t.Errorf("expected 'test-value', got %v", value)
	}
}

func TestCacheTemplate_Del(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	// Set
	_ = template.Set(ctx, "test-key", "test-value", 5*time.Minute)

	// Del
	err := template.Del(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to del: %v", err)
	}

	// 应该不存在
	exists, err := template.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to check exists: %v", err)
	}

	if exists {
		t.Error("expected key to not exist after deletion")
	}
}

func TestCacheTemplate_GetOrSet(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	callCount := 0

	// 第一次调用
	value, err := template.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "generated", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get or set: %v", err)
	}

	if value != "generated" {
		t.Errorf("expected 'generated', got %v", value)
	}

	if callCount != 1 {
		t.Errorf("expected callCount 1, got %d", callCount)
	}

	// 第二次调用应该使用缓存
	value, err = template.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "should-not-be-called", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get from cache: %v", err)
	}

	if value != "generated" {
		t.Errorf("expected 'generated' from cache, got %v", value)
	}

	if callCount != 1 {
		t.Errorf("expected callCount still 1, got %d", callCount)
	}
}

func TestCacheConfig_Default(t *testing.T) {
	t.Parallel()
	config := DefaultCacheConfig()

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}

	if config.DefaultTTL != 30*time.Minute {
		t.Errorf("expected DefaultTTL 30m, got %v", config.DefaultTTL)
	}

	if config.MaxSize != 10000 {
		t.Errorf("expected MaxSize 10000, got %d", config.MaxSize)
	}

	if config.StatsEnabled {
		t.Error("expected StatsEnabled to be false")
	}
}

func TestCacheConfig_ApplyOptions(t *testing.T) {
	t.Parallel()
	config := DefaultCacheConfig()

	config.ApplyOptions([]CacheOption{
		WithCacheEnabled(false),
		WithDefaultTTL(1 * time.Hour),
		WithMaxSize(5000),
		WithKeyPrefix("test"),
		WithStatsEnabled(true),
	})

	if config.Enabled {
		t.Error("expected Enabled to be false")
	}

	if config.DefaultTTL != 1*time.Hour {
		t.Errorf("expected DefaultTTL 1h, got %v", config.DefaultTTL)
	}

	if config.MaxSize != 5000 {
		t.Errorf("expected MaxSize 5000, got %d", config.MaxSize)
	}

	if config.KeyPrefix != "test" {
		t.Errorf("expected KeyPrefix 'test', got %s", config.KeyPrefix)
	}

	if !config.StatsEnabled {
		t.Error("expected StatsEnabled to be true")
	}
}

func TestCacheHelper_GetNotFound(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	_, err := helper.Get(ctx, "non-existent-key")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}

func TestCacheTemplate_TTL(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	template := NewCacheTemplate(memCache, "myapp")

	ctx := context.Background()

	// 设置带 TTL 的值
	_ = template.Set(ctx, "test-key", "test-value", 10*time.Minute)

	// Get TTL
	ttl, err := template.TTL(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get TTL: %v", err)
	}

	if ttl <= 0 || ttl > 10*time.Minute {
		t.Errorf("expected TTL between 0 and 10m, got %v", ttl)
	}
}

func TestCacheHelper_GetOrSetWithNilResult(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// 测试 nil 结果
	value, err := helper.GetOrSet(ctx, "nil-key", func() (any, error) {
		return nil, nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != nil {
		t.Errorf("expected nil value, got %v", value)
	}
}
