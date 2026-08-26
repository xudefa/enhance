package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTTLCache_GetSet(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	if err := c.Set(ctx, "key1", "value1", time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestTTLCache_GetSet_Overwrite(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "original", time.Minute)
	_ = c.Set(ctx, "key1", "updated", time.Minute)

	val, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "updated" {
		t.Errorf("expected 'updated', got %v", val)
	}
	if c.Size() != 1 {
		t.Errorf("expected size 1 after overwrite, got %d", c.Size())
	}
}

func TestTTLCache_GetExpired(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	_, err := c.Get(ctx, "key1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for expired key, got %v", err)
	}
}

func TestTTLCache_GetNotFound(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_, err := c.Get(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTTLCache_Del_Existing(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)
	if err := c.Del(ctx, "key1"); err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	_, err := c.Get(ctx, "key1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestTTLCache_Del_NonExisting(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	if err := c.Del(ctx, "nonexistent"); err != nil {
		t.Fatalf("Del on nonexistent key should not error: %v", err)
	}
}

func TestTTLCache_Exists(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)

	exists, err := c.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected key to exist")
	}
}

func TestTTLCache_Exists_NonExisting(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	exists, err := c.Exists(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected key to not exist")
	}
}

func TestTTLCache_Exists_Expired(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	exists, err := c.Exists(ctx, "key1")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected expired key to not exist")
	}
}

func TestTTLCache_TTL_WithTTL(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Second)

	ttl, err := c.TTL(ctx, "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl <= 0 || ttl > time.Second {
		t.Errorf("expected TTL between 0 and 1s, got %v", ttl)
	}
}

func TestTTLCache_TTL_NoTTL(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", 0)

	ttl, err := c.TTL(ctx, "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl != -1 {
		t.Errorf("expected TTL -1 for no expiration, got %v", ttl)
	}
}

func TestTTLCache_TTL_NotFound(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_, err := c.TTL(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTTLCache_TTL_Expired(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	_, err := c.TTL(ctx, "key1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for expired key TTL, got %v", err)
	}
}

func TestTTLCache_Close(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)
	_ = c.Set(ctx, "key2", "value2", time.Minute)

	if err := c.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if c.Size() != 0 {
		t.Errorf("expected size 0 after close, got %d", c.Size())
	}

	_, err := c.Get(ctx, "key1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after close, got %v", err)
	}
}

func TestTTLCache_Size(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	if c.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", c.Size())
	}

	_ = c.Set(ctx, "key1", "value1", time.Minute)
	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}

	_ = c.Set(ctx, "key2", "value2", time.Minute)
	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}

	_ = c.Del(ctx, "key1")
	if c.Size() != 1 {
		t.Errorf("expected size 1 after delete, got %d", c.Size())
	}
}

func TestTTLCache_Clear(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)
	_ = c.Set(ctx, "key2", "value2", time.Minute)

	if err := c.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if c.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", c.Size())
	}
}

func TestTTLCache_Stats(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(
		WithTTLMaxSize(100),
		WithTTLDefaultTTL(5*time.Minute),
	)
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)

	stats := c.Stats()
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats["size"] != 1 {
		t.Errorf("expected stats size 1, got %v", stats["size"])
	}
	if stats["max_size"] != 100 {
		t.Errorf("expected stats max_size 100, got %v", stats["max_size"])
	}
	if stats["default_ttl"] != "5m0s" {
		t.Errorf("expected stats default_ttl '5m0s', got %v", stats["default_ttl"])
	}
}

func TestTTLCache_LRUEviction(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(2))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)
	_ = c.Set(ctx, "key2", "value2", time.Minute)
	_ = c.Set(ctx, "key3", "value3", time.Minute)

	_, err := c.Get(ctx, "key1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected key1 to be evicted, got err=%v", err)
	}

	val, err := c.Get(ctx, "key2")
	if err != nil {
		t.Fatalf("expected key2 to exist: %v", err)
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}

	val, err = c.Get(ctx, "key3")
	if err != nil {
		t.Fatalf("expected key3 to exist: %v", err)
	}
	if val != "value3" {
		t.Errorf("expected value3, got %v", val)
	}
}

func TestTTLCache_WithTTLMaxSize(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(5))
	if c.maxSize != 5 {
		t.Errorf("expected maxSize 5, got %d", c.maxSize)
	}
}

func TestTTLCache_WithTTLDefaultTTL(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLDefaultTTL(10 * time.Minute))
	if c.defaultTTL != 10*time.Minute {
		t.Errorf("expected defaultTTL 10m, got %v", c.defaultTTL)
	}
}

func TestTTLCache_DefaultConfig(t *testing.T) {
	t.Parallel()
	c := NewTTLCache()
	if c.maxSize != 1000 {
		t.Errorf("expected default maxSize 1000, got %d", c.maxSize)
	}
	if c.defaultTTL != 5*time.Minute {
		t.Errorf("expected default defaultTTL 5m, got %v", c.defaultTTL)
	}
}

func TestTTLCache_Set_NegativeTTL(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(
		WithTTLMaxSize(10),
		WithTTLDefaultTTL(50*time.Millisecond),
	)
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", -1)
	time.Sleep(100 * time.Millisecond)

	_, err := c.Get(ctx, "key1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for default TTL expiry, got %v", err)
	}
}

func TestTTLCache_Del_Multiple(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(10))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)
	_ = c.Set(ctx, "key2", "value2", time.Minute)
	_ = c.Set(ctx, "key3", "value3", time.Minute)

	if err := c.Del(ctx, "key1", "key3"); err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	if c.Size() != 1 {
		t.Errorf("expected size 1 after deleting 2 keys, got %d", c.Size())
	}

	_, err := c.Get(ctx, "key2")
	if err != nil {
		t.Errorf("expected key2 to exist: %v", err)
	}
}

func TestTTLCache_LRU_Access_Updates_Position(t *testing.T) {
	t.Parallel()
	c := NewTTLCache(WithTTLMaxSize(2))
	ctx := context.Background()

	_ = c.Set(ctx, "key1", "value1", time.Minute)
	_ = c.Set(ctx, "key2", "value2", time.Minute)

	// Access key1 to make it recently used
	_, _ = c.Get(ctx, "key1")

	// Add key3, should evict key2 (least recently used)
	_ = c.Set(ctx, "key3", "value3", time.Minute)

	_, err := c.Get(ctx, "key1")
	if err != nil {
		t.Errorf("expected key1 to exist after access, got %v", err)
	}

	_, err = c.Get(ctx, "key2")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected key2 to be evicted, got %v", err)
	}

	_, err = c.Get(ctx, "key3")
	if err != nil {
		t.Errorf("expected key3 to exist, got %v", err)
	}
}

func TestErrNotFound_IsComparable(t *testing.T) {
	t.Parallel()
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("ErrNotFound should be comparable")
	}
}

func TestErrCacheMiss_IsComparable(t *testing.T) {
	t.Parallel()
	if !errors.Is(ErrCacheMiss, ErrCacheMiss) {
		t.Error("ErrCacheMiss should be comparable")
	}
}

func TestErrNotFound_ErrorString(t *testing.T) {
	t.Parallel()
	expected := "cache: key not found"
	if ErrNotFound.Error() != expected {
		t.Errorf("ErrNotFound.Error() = %q, want %q", ErrNotFound.Error(), expected)
	}
}

func TestErrCacheMiss_ErrorString(t *testing.T) {
	t.Parallel()
	expected := "cache: key expired or not found"
	if ErrCacheMiss.Error() != expected {
		t.Errorf("ErrCacheMiss.Error() = %q, want %q", ErrCacheMiss.Error(), expected)
	}
}
