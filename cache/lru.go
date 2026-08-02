// Package cache 提供缓存抽象层，用于 enhance 框架。
package cache

import (
	"container/list"
	"context"
	"time"
)

// WithEvictCallback 设置淘汰回调函数。
//
// 当缓存项因容量限制被淘汰时调用此回调。
// 可用于记录日志、清理资源等场景。
func WithEvictCallback(fn func(key string, value any)) LRUOption {
	return func(c *LRUCache) {
		c.onEvict = fn
	}
}

// WithTTL 设置默认 TTL。
//
// 当 Set 传入 ttl <= 0 时使用该默认值。
func WithTTL(ttl time.Duration) LRUOption {
	return func(c *LRUCache) {
		c.defaultTTL = ttl
	}
}

// NewLRUCache 创建 LRU 缓存。
//
// 参数:
//   - capacity: 缓存容量，必须大于 0（<=0 时使用默认值 100）
//   - opts: 可选配置项
//
// 返回:
//   - *LRUCache: LRU 缓存实例
func NewLRUCache(capacity int, opts ...LRUOption) *LRUCache {
	if capacity <= 0 {
		capacity = 100
	}
	c := &LRUCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element, capacity),
		evictList: list.New(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// evictedEntry 被淘汰的缓存项。
type evictedEntry struct {
	key   string
	value any
}

// fireEvicted 在锁外调用淘汰回调。
func (c *LRUCache) fireEvicted(evicted []evictedEntry) {
	if c.onEvict == nil {
		return
	}
	for _, e := range evicted {
		c.onEvict(e.key, e.value)
	}
}

// Get 获取缓存值。
//
// 如果键不存在或已过期，返回 ErrNotFound。
// 命中缓存时会将该键移到链表前端（标记为最近使用）。
//
// 参数:
//   - ctx: 上下文（保留用于接口一致性，当前未使用）
//   - key: 缓存键
//
// 返回值:
//   - any: 缓存值
//   - error: 键不存在或已过期时返回 ErrNotFound
func (c *LRUCache) Get(ctx context.Context, key string) (any, error) {
	c.mu.Lock()
	elem, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return nil, ErrNotFound
	}

	entry, ok := elem.Value.(*lruEntry)
	if !ok {
		c.mu.Unlock()
		return nil, ErrNotFound
	}

	// 检查是否过期
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		evicted := c.removeElement(elem)
		c.mu.Unlock()
		c.fireEvicted(evicted)
		return nil, ErrNotFound
	}

	// 移到最前（标记为最近使用）
	c.evictList.MoveToFront(elem)
	c.mu.Unlock()
	return entry.value, nil
}

// Set 设置缓存值。
//
// 如果键已存在，更新值并刷新过期时间。
// 如果超出容量，淘汰最久未使用的项。
//
// 参数:
//   - ctx: 上下文（保留用于接口一致性，当前未使用）
//   - key: 缓存键
//   - value: 缓存值
//   - ttl: 过期时间（<=0 表示永不过期）
func (c *LRUCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.mu.Lock()

	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	// 如果已存在，更新值
	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		entry, ok := elem.Value.(*lruEntry)
		if !ok {
			c.mu.Unlock()
			return nil
		}
		entry.value = value
		if ttl <= 0 {
			entry.expiresAt = time.Time{}
			c.mu.Unlock()
			return nil
		}
		entry.expiresAt = time.Now().Add(ttl)
		c.mu.Unlock()
		return nil
	}

	// 容量淘汰
	var evicted []evictedEntry
	for c.evictList.Len() >= c.capacity {
		evicted = append(evicted, c.evictOldest()...)
	}

	// 添加新项
	entry := &lruEntry{
		key:   key,
		value: value,
	}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem

	c.mu.Unlock()

	c.fireEvicted(evicted)

	return nil
}

// Del 删除缓存项。
//
// 参数:
//   - ctx: 上下文（保留用于接口一致性，当前未使用）
//   - keys: 要删除的缓存键列表
func (c *LRUCache) Del(ctx context.Context, keys ...string) error {
	c.mu.Lock()
	var evicted []evictedEntry
	for _, key := range keys {
		if elem, ok := c.items[key]; ok {
			evicted = append(evicted, c.removeElement(elem)...)
		}
	}
	c.mu.Unlock()

	c.fireEvicted(evicted)
	return nil
}

// Exists 检查键是否存在且未过期。
//
// 参数:
//   - ctx: 上下文（保留用于接口一致性，当前未使用）
//   - key: 缓存键
//
// 返回值:
//   - bool: 键存在且未过期返回 true
//   - error: 始终返回 nil
func (c *LRUCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()

	elem, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return false, nil
	}

	entry, ok := elem.Value.(*lruEntry)
	if !ok {
		c.mu.Unlock()
		return false, nil
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		evicted := c.removeElement(elem)
		c.mu.Unlock()
		c.fireEvicted(evicted)
		return false, nil
	}

	c.mu.Unlock()
	return true, nil
}

// TTL 获取键的剩余过期时间。
//
// 参数:
//   - ctx: 上下文（保留用于接口一致性，当前未使用）
//   - key: 缓存键
//
// 返回值:
//   - time.Duration: 剩余过期时间（永不过期时返回 -1）
//   - error: 键不存在或已过期时返回 ErrNotFound
func (c *LRUCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.Lock()

	elem, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return 0, ErrNotFound
	}

	entry, ok := elem.Value.(*lruEntry)
	if !ok {
		c.mu.Unlock()
		return 0, ErrNotFound
	}

	// 永不过期
	if entry.expiresAt.IsZero() {
		c.mu.Unlock()
		return -1, nil
	}

	remaining := time.Until(entry.expiresAt)
	if remaining <= 0 {
		evicted := c.removeElement(elem)
		c.mu.Unlock()
		c.fireEvicted(evicted)
		return 0, ErrNotFound
	}

	c.mu.Unlock()
	return remaining, nil
}

// Close 关闭缓存并清空所有数据。
func (c *LRUCache) Close() error {
	c.Clear()
	return nil
}

// Len 返回缓存项数量。
func (c *LRUCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.evictList.Len()
}

// Clear 清空缓存。
func (c *LRUCache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]*list.Element, c.capacity)
	c.evictList = list.New()
	c.mu.Unlock()
}

// evictOldest 淘汰最久未使用的项。
//
// 注意：调用方必须持有写锁。
func (c *LRUCache) evictOldest() []evictedEntry {
	elem := c.evictList.Back()
	if elem != nil {
		return c.removeElement(elem)
	}
	return nil
}

// removeElement 移除元素并返回被淘汰的项。
//
// 注意：调用方必须持有写锁。
func (c *LRUCache) removeElement(elem *list.Element) []evictedEntry {
	if elem == nil {
		return nil
	}

	c.evictList.Remove(elem)

	entry, ok := elem.Value.(*lruEntry)
	if !ok || entry == nil {
		return nil
	}

	delete(c.items, entry.key)
	return []evictedEntry{{key: entry.key, value: entry.value}}
}
