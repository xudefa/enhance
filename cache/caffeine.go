// Package cache 提供类似 Caffeine 的本地缓存实现，采用 LRU 淘汰策略。
//
// CaffeineCache 是一个实现 Cache 接口的高性能本地缓存。
// 当缓存达到最大容量时，它提供 LRU（最近最少使用）淘汰策略，
// 并为单个缓存条目提供 TTL（生存时间）支持。
//
// 用法：
//
//	cache := cache.NewCaffeineCache(
//		cache.WithCaffeineMaxSize(1000),
//		cache.WithCaffeineDefaultTTL(5*time.Minute),
//	)
//
//	cache.Set(ctx, "key", "value", 10*time.Minute)
//	value, err := cache.Get(ctx, "key")
package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// caffeineCache implements a high-performance local cache with LRU eviction.
type caffeineCache struct {
	maxSize    int
	defaultTTL time.Duration
	items      map[string]*caffeineItem
	lru        *list.List
	mu         sync.RWMutex
}

// caffeineItem represents a cached item with LRU tracking.
type caffeineItem struct {
	key        string
	value      any
	expireAt   time.Time
	lruElement *list.Element
}

// CaffeineOption configures the Caffeine cache.
type CaffeineOption func(*caffeineCache)

// WithCaffeineMaxSize 设置缓存中的最大条目数。
// 当缓存达到此大小时，最近最少使用的条目将被淘汰。
func WithCaffeineMaxSize(maxSize int) CaffeineOption {
	return func(c *caffeineCache) {
		c.maxSize = maxSize
	}
}

// WithCaffeineDefaultTTL 设置缓存条目的默认 TTL。
// 当调用 Set() 且 ttl <= 0 时使用此 TTL。
func WithCaffeineDefaultTTL(ttl time.Duration) CaffeineOption {
	return func(c *caffeineCache) {
		c.defaultTTL = ttl
	}
}

// NewCaffeineCache 创建一个新的类 Caffeine 本地缓存，采用 LRU 淘汰策略。
//
// 该缓存使用 map 实现 O(1) 查找，使用双向链表实现 O(1) LRU 跟踪。默认配置：
//   - 最大容量：1000 个条目
//   - 默认 TTL：5 分钟
func NewCaffeineCache(opts ...CaffeineOption) *caffeineCache {
	c := &caffeineCache{
		maxSize:    1000,
		defaultTTL: 5 * time.Minute,
		items:      make(map[string]*caffeineItem),
		lru:        list.New(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Get 通过键检索缓存值。
// 如果键不存在或条目已过期，则返回 ErrNotFound。
// 访问条目会更新其 LRU 位置（移到末尾）。
func (c *caffeineCache) Get(ctx context.Context, key string) (any, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, ErrNotFound
	}

	// 检查是否过期
	if time.Now().After(item.expireAt) {
		c.mu.Lock()
		// 双重检查
		if item, exists := c.items[key]; exists && time.Now().After(item.expireAt) {
			c.deleteItem(item)
			c.mu.Unlock()
			return nil, ErrNotFound
		}
		c.mu.Unlock()
		return nil, ErrNotFound
	}

	// 更新 LRU 位置（移到尾部表示最近使用）
	c.mu.Lock()
	c.lru.MoveToBack(item.lruElement)
	c.mu.Unlock()

	return item.value, nil
}

// Set 在缓存中存储一个带 TTL 的值。
// 如果 ttl <= 0，则使用默认 TTL。
// 如果键已存在，则更新值并刷新 LRU 位置。
// 如果缓存达到最大容量，则淘汰最近最少使用的条目。
func (c *caffeineCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果 ttl <= 0，使用默认 TTL
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	// 如果键已存在，更新值并刷新 LRU 位置
	if existing, exists := c.items[key]; exists {
		existing.value = value
		existing.expireAt = time.Now().Add(ttl)
		c.lru.MoveToBack(existing.lruElement)
		return nil
	}

	// 如果达到最大容量，淘汰最久未使用的项
	if len(c.items) >= c.maxSize {
		c.evictLRU()
	}

	// 添加新项
	item := &caffeineItem{
		key:      key,
		value:    value,
		expireAt: time.Now().Add(ttl),
	}
	item.lruElement = c.lru.PushBack(key)
	c.items[key] = item

	return nil
}

// Del 通过键删除缓存条目。
// Non-existent keys are silently ignored.
func (c *caffeineCache) Del(ctx context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		if item, exists := c.items[key]; exists {
			c.deleteItem(item)
		}
	}

	return nil
}

// Exists 检查键是否存在且未过期。
// 检查期间会清理过期条目。
func (c *caffeineCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if time.Now().After(item.expireAt) {
		c.mu.Lock()
		// 双重检查
		if item, exists := c.items[key]; exists && time.Now().After(item.expireAt) {
			c.deleteItem(item)
		}
		c.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// TTL gets the remaining TTL for a key.
// Returns ErrNotFound if the key doesn't exist.
// Returns 0 if the item has no expiration (永不过期).
func (c *caffeineCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return 0, ErrNotFound
	}

	if time.Now().After(item.expireAt) {
		return 0, nil
	}

	return time.Until(item.expireAt), nil
}

// Close 关闭缓存并释放所有资源。
func (c *caffeineCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*caffeineItem)
	c.lru = list.New()

	return nil
}

// deleteItem 从映射和 LRU 列表中删除条目。
// 调用者必须持有写锁。
func (c *caffeineCache) deleteItem(item *caffeineItem) {
	if item.lruElement != nil {
		c.lru.Remove(item.lruElement)
	}
	delete(c.items, item.key)
}

// evictLRU 从缓存中删除最近最少使用的条目。
// LRU 条目位于列表前端。
// 调用者必须持有写锁。
func (c *caffeineCache) evictLRU() {
	front := c.lru.Front()
	if front == nil {
		return
	}

	key := front.Value.(string)
	if item, exists := c.items[key]; exists {
		c.deleteItem(item)
	}
}

// Size 返回缓存中当前的条目数量。
func (c *caffeineCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Clear 从缓存中删除所有条目。
func (c *caffeineCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*caffeineItem)
	c.lru = list.New()

	return nil
}

// Stats 返回缓存统计信息用于监控。
func (c *caffeineCache) Stats() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]any{
		"size":        len(c.items),
		"max_size":    c.maxSize,
		"default_ttl": c.defaultTTL.String(),
	}
}
