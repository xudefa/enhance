// Package cache 提供缓存抽象层，用于 enhance 框架。
package cache

import (
	"container/list"
	"context"
	"hash/fnv"
	"time"
)

// NewShardedLRUCache 创建分片 LRU 缓存
//
// 参数:
//   - capacity: 总容量，将平均分配到各分片
//   - shardCount: 分片数量，建议为 2 的幂次（如 16, 32, 64）
//   - opts: 可选配置项
//
// 返回:
//   - *ShardedLRUCache: 分片 LRU 缓存实例
func NewShardedLRUCache(capacity int, shardCount int, opts ...LRUOption) *ShardedLRUCache {
	if capacity <= 0 {
		capacity = 1000
	}
	if shardCount <= 0 {
		shardCount = 16
	}

	perShard := capacity / shardCount
	if perShard < 1 {
		perShard = 1
	}

	shards := make([]*lruShard, shardCount)
	for i := range shards {
		shards[i] = &lruShard{
			capacity:  perShard,
			items:     make(map[string]*list.Element, perShard),
			evictList: list.New(),
		}
	}

	cache := &ShardedLRUCache{
		shards:     shards,
		capacity:   capacity,
		shardCount: shardCount,
	}

	// 应用选项（onEvict 回调会应用到所有分片）
	for _, opt := range opts {
		// 创建临时 LRUCache 来提取配置
		temp := &LRUCache{}
		opt(temp)
		if temp.onEvict != nil {
			cache.onEvict = temp.onEvict
		}
	}

	return cache
}

// getShardIndex 计算 key 对应的分片索引
func (c *ShardedLRUCache) getShardIndex(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % uint32(c.shardCount))
}

// fireEvicted 在锁外调用淘汰回调。
func (c *ShardedLRUCache) fireEvicted(evicted []evictedEntry) {
	if c.onEvict == nil {
		return
	}
	for _, e := range evicted {
		c.onEvict(e.key, e.value)
	}
}

// Get 获取缓存值
func (c *ShardedLRUCache) Get(ctx context.Context, key string) (any, error) {
	shard := c.shards[c.getShardIndex(key)]
	shard.mu.Lock()

	elem, ok := shard.items[key]
	if !ok {
		shard.mu.Unlock()
		return nil, ErrNotFound
	}

	entry, ok := elem.Value.(*lruEntry)
	if !ok {
		shard.mu.Unlock()
		return nil, ErrNotFound
	}

	// 检查是否过期
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		evicted := c.shardRemoveElement(shard, elem)
		shard.mu.Unlock()
		c.fireEvicted(evicted)
		return nil, ErrNotFound
	}

	// 移到最前
	shard.evictList.MoveToFront(elem)
	shard.mu.Unlock()
	return entry.value, nil
}

// Set 设置缓存值
func (c *ShardedLRUCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	shard := c.shards[c.getShardIndex(key)]
	shard.mu.Lock()

	// 如果已存在，更新值
	if elem, ok := shard.items[key]; ok {
		shard.evictList.MoveToFront(elem)
		entry, ok := elem.Value.(*lruEntry)
		if !ok {
			shard.mu.Unlock()
			return nil
		}
		entry.value = value
		if ttl <= 0 {
			entry.expiresAt = time.Time{}
			shard.mu.Unlock()
			return nil
		}
		entry.expiresAt = time.Now().Add(ttl)
		shard.mu.Unlock()
		return nil
	}

	// 如果超出容量，淘汰最久未使用的项
	var evicted []evictedEntry
	for shard.evictList.Len() >= shard.capacity {
		evicted = append(evicted, c.shardEvictOldest(shard)...)
	}

	// 添加新项
	entry := &lruEntry{
		key:   key,
		value: value,
	}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}
	elem := shard.evictList.PushFront(entry)
	shard.items[key] = elem

	shard.mu.Unlock()

	c.fireEvicted(evicted)

	return nil
}

// Del 删除缓存项
func (c *ShardedLRUCache) Del(ctx context.Context, keys ...string) error {
	// 按分片分组，减少锁竞争
	shardKeys := make(map[int][]string)
	for _, key := range keys {
		idx := c.getShardIndex(key)
		shardKeys[idx] = append(shardKeys[idx], key)
	}

	var evicted []evictedEntry
	for idx, keys := range shardKeys {
		shard := c.shards[idx]
		shard.mu.Lock()
		for _, key := range keys {
			if elem, ok := shard.items[key]; ok {
				evicted = append(evicted, c.shardRemoveElement(shard, elem)...)
			}
		}
		shard.mu.Unlock()
	}

	c.fireEvicted(evicted)
	return nil
}

// Exists 检查键是否存在且未过期
func (c *ShardedLRUCache) Exists(ctx context.Context, key string) (bool, error) {
	shard := c.shards[c.getShardIndex(key)]
	shard.mu.Lock()

	elem, ok := shard.items[key]
	if !ok {
		shard.mu.Unlock()
		return false, nil
	}

	entry, ok := elem.Value.(*lruEntry)
	if !ok {
		shard.mu.Unlock()
		return false, nil
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		evicted := c.shardRemoveElement(shard, elem)
		shard.mu.Unlock()
		c.fireEvicted(evicted)
		return false, nil
	}

	shard.mu.Unlock()
	return true, nil
}

// TTL 获取键的剩余过期时间
func (c *ShardedLRUCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	shard := c.shards[c.getShardIndex(key)]
	shard.mu.Lock()

	elem, ok := shard.items[key]
	if !ok {
		shard.mu.Unlock()
		return 0, ErrNotFound
	}

	entry, ok := elem.Value.(*lruEntry)
	if !ok {
		shard.mu.Unlock()
		return 0, ErrNotFound
	}

	// 如果没有设置 TTL，返回 -1 表示永不过期
	if entry.expiresAt.IsZero() {
		shard.mu.Unlock()
		return -1, nil
	}

	remaining := time.Until(entry.expiresAt)
	if remaining <= 0 {
		evicted := c.shardRemoveElement(shard, elem)
		shard.mu.Unlock()
		c.fireEvicted(evicted)
		return 0, ErrNotFound
	}

	shard.mu.Unlock()
	return remaining, nil
}

// Close 关闭缓存
func (c *ShardedLRUCache) Close() error {
	c.Clear()
	return nil
}

// Len 返回缓存项数量
func (c *ShardedLRUCache) Len() int {
	total := 0
	for _, shard := range c.shards {
		shard.mu.RLock()
		total += shard.evictList.Len()
		shard.mu.RUnlock()
	}
	return total
}

// Clear 清空缓存
func (c *ShardedLRUCache) Clear() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.items = make(map[string]*list.Element, shard.capacity)
		shard.evictList = list.New()
		shard.mu.Unlock()
	}
}

// shardEvictOldest 淘汰分片中最久未使用的项
func (c *ShardedLRUCache) shardEvictOldest(shard *lruShard) []evictedEntry {
	elem := shard.evictList.Back()
	if elem != nil {
		return c.shardRemoveElement(shard, elem)
	}
	return nil
}

// shardRemoveElement 移除分片中的元素
func (c *ShardedLRUCache) shardRemoveElement(shard *lruShard, elem *list.Element) []evictedEntry {
	if elem == nil || shard == nil {
		return nil
	}

	shard.evictList.Remove(elem)

	entry, ok := elem.Value.(*lruEntry)
	if !ok || entry == nil {
		return nil
	}

	delete(shard.items, entry.key)
	return []evictedEntry{{key: entry.key, value: entry.value}}
}
