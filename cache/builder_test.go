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

	got, err := cache.Get(ctx, "test")
	if err != nil || got != "value" {
		t.Errorf("expected 'value', got %v, err=%v", got, err)
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
	value, err := helper.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	strVal, ok := value.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", value)
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
	value, err := helper.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	_, ok := value.(int)
	if ok {
		t.Error("expected type assertion to fail")
	}
}

func testCacheHelperGetOrSetFirstCall(t *testing.T, helper *CacheHelper) {
	t.Helper()
	var callCount int
	ctx := context.Background()

	// 第一次调用应该执行函数
	value, err := helper.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "generated-value", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get or set: %v", err)
	}

	strVal, ok := value.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", value)
	}

	if strVal != "generated-value" {
		t.Errorf("expected 'generated-value', got %s", strVal)
	}

	if callCount != 1 {
		t.Errorf("expected callCount 1, got %d", callCount)
	}
}

func testCacheHelperGetOrSetUsesCache(t *testing.T, helper *CacheHelper) {
	t.Helper()
	var callCount int
	ctx := context.Background()

	// 第二次调用应该使用缓存
	value, err := helper.GetOrSet(ctx, "test-key", func() (any, error) {
		callCount++
		return "should-not-be-called", nil
	}, 5*time.Minute)

	if err != nil {
		t.Fatalf("failed to get from cache: %v", err)
	}

	strVal, ok := value.(string)
	if !ok {
		t.Fatalf("expected string type, got %T", value)
	}

	if strVal != "generated-value" {
		t.Errorf("expected 'generated-value' from cache, got %s", strVal)
	}

	if callCount != 0 {
		t.Errorf("expected callCount still 0 (cached), got %d", callCount)
	}
}

func TestCacheHelper_GetOrSet(t *testing.T) {
	t.Parallel()
	memCache := NewMemoryCacheBuilder().Build()
	helper := NewCacheHelper(memCache)

	testCacheHelperGetOrSetFirstCall(t, helper)
	// 对已被第一次调用缓存的 key 再次获取
	testCacheHelperGetOrSetUsesCache(t, helper)
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
