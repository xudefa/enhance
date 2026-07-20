package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCacheBuilder_Defaults(t *testing.T) {
	t.Parallel()
	builder := NewMemoryCacheBuilder()

	if builder.initialCapacity != 1024 {
		t.Errorf("expected default initialCapacity 1024, got %d", builder.initialCapacity)
	}
}

func TestMemoryCacheBuilder_ChainConfig(t *testing.T) {
	t.Parallel()
	cache := NewMemoryCacheBuilder().
		InitialCapacity(2048).
		Build()

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}

	// 验证缓存可以正常工作
	ctx := context.Background()
	err := cache.Set(ctx, "test", "value", 0)
	if err != nil {
		t.Errorf("failed to set cache: %v", err)
	}

	val, err := cache.Get(ctx, "test")
	if err != nil || val != "value" {
		t.Errorf("expected 'value', got %v, err=%v", val, err)
	}
}

func TestMemoryCacheBuilder_MustBuild(t *testing.T) {
	t.Parallel()
	cache := NewMemoryCacheBuilder().MustBuild()

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestCacheHelper_Get(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// 设置一个值
	err := helper.Set(ctx, "test-key", "test-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// 获取值
	val, err := helper.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	strVal, ok := val.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", val)
	}

	if strVal != "test-value" {
		t.Errorf("expected 'test-value', got %s", strVal)
	}
}

func TestCacheHelper_GetTypeMismatch(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// 设置字符串值
	err := helper.Set(ctx, "test-key", "string-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// 尝试作为 int 获取 - 应该失败类型断言
	val, err := helper.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	_, ok := val.(int)
	if ok {
		t.Error("expected type assertion to fail")
	}
}

func TestCacheHelper_GetOrSet(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	callCount := 0

	// 第一次调用应该执行函数
	val, err := helper.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "generated-value", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get or set: %v", err)
	}

	strVal, ok := val.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", val)
	}

	if strVal != "generated-value" {
		t.Errorf("expected 'generated-value', got %s", strVal)
	}

	if callCount != 1 {
		t.Errorf("expected callCount 1, got %d", callCount)
	}

	// 第二次调用应该使用缓存
	val, err = helper.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "should-not-be-called", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get from cache: %v", err)
	}

	strVal, ok = val.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", val)
	}

	if strVal != "generated-value" {
		t.Errorf("expected 'generated-value' from cache, got %s", strVal)
	}

	if callCount != 1 {
		t.Errorf("expected callCount still 1 (cached), got %d", callCount)
	}
}

func TestCacheHelper_GetOrSet_FunctionError(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	_, err := helper.GetOrSet(ctx, "test-key", func() (any, error) {
		return nil, nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCacheHelper_Invalidate(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// 设置一个值
	err := helper.Set(ctx, "test-key", "test-value", 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// 清除
	err = helper.Invalidate(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to invalidate: %v", err)
	}

	// 应该不存在
	exists, err := helper.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to check exists: %v", err)
	}

	if exists {
		t.Error("expected key to not exist after invalidation")
	}
}

func TestCacheHelper_InvalidateAll(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// 设置多个值
	_ = helper.Set(ctx, "key1", "value1", 5*time.Minute)
	_ = helper.Set(ctx, "key2", "value2", 5*time.Minute)
	_ = helper.Set(ctx, "key3", "value3", 5*time.Minute)

	// 清除所有
	err := helper.InvalidateAll(ctx, "key1", "key2")
	if err != nil {
		t.Fatalf("failed to invalidate all: %v", err)
	}

	// key1 和 key2 应该不存在
	exists1, _ := helper.Exists(ctx, "key1")
	exists2, _ := helper.Exists(ctx, "key2")
	exists3, _ := helper.Exists(ctx, "key3")

	if exists1 {
		t.Error("expected key1 to not exist")
	}

	if exists2 {
		t.Error("expected key2 to not exist")
	}

	if !exists3 {
		t.Error("expected key3 to still exist")
	}
}

func TestCacheHelper_Clear(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// 设置多个值
	_ = helper.Set(ctx, "key1", "value1", 5*time.Minute)
	_ = helper.Set(ctx, "key2", "value2", 5*time.Minute)

	// Clear
	err := helper.Clear(ctx)
	if err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	// 所有键应该不存在
	exists1, _ := helper.Exists(ctx, "key1")
	exists2, _ := helper.Exists(ctx, "key2")

	if exists1 {
		t.Error("expected key1 to not exist after clear")
	}

	if exists2 {
		t.Error("expected key2 to not exist after clear")
	}
}

func TestCacheHelper_TTL(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	ctx := context.Background()

	// 设置带 TTL 的值
	err := helper.Set(ctx, "test-key", "test-value", 10*time.Minute)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Get TTL
	ttl, err := helper.TTL(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get TTL: %v", err)
	}

	if ttl <= 0 || ttl > 10*time.Minute {
		t.Errorf("expected TTL between 0 and 10m, got %v", ttl)
	}
}

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
	val, err := template.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if val != "test-value" {
		t.Errorf("expected 'test-value', got %v", val)
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
	val, err := template.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "generated", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get or set: %v", err)
	}

	if val != "generated" {
		t.Errorf("expected 'generated', got %v", val)
	}

	if callCount != 1 {
		t.Errorf("expected callCount 1, got %d", callCount)
	}

	// 第二次调用应该使用缓存
	val, err = template.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "should-not-be-called", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get from cache: %v", err)
	}

	if val != "generated" {
		t.Errorf("expected 'generated' from cache, got %v", val)
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
	val, err := helper.GetOrSet(ctx, "nil-key", func() (any, error) {
		return nil, nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != nil {
		t.Errorf("expected nil value, got %v", val)
	}
}
