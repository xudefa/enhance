// Package cache 提供缓存抽象层，用于 enhance 框架。
//
// 该模块提供多种缓存策略实现，包括 LRU 缓存和 TTL 缓存。
// 参考 Spring Cache 的设计理念，提供统一的缓存抽象接口，
// 使得应用可以轻松切换不同的缓存实现（内存缓存、Redis 等）。
//
// # 设计原则
//
//   - 接口优先：定义统一的 Cache 接口，便于替换实现
//   - 零外部依赖：核心缓存实现仅使用 Go 标准库
//   - 并发安全：所有实现都支持并发访问
//   - 灵活配置：通过函数式选项模式提供灵活配置
//
// # 架构设计
//
//   - Cache: 缓存操作接口
//   - Getter: 缓存获取器，支持缓存穿透保护
//   - Builder: 缓存构建器，支持链式配置
//   - LRUCache: LRU（Least Recently Used）缓存实现
//   - ShardedLRUCache: 分片 LRU 缓存，适合高并发场景
//   - TTLCache: 基于 LRU + TTL 的高性能缓存实现
//
// # 支持的缓存策略
//
//   - LRU 缓存: 基于最近最少使用算法的缓存实现
//   - TTL 缓存: 基于 LRU + TTL 的高性能缓存实现
//   - 分片缓存: 支持并发安全的分片 LRU 缓存
//
// # 使用方式
//
// 使用 LRU 缓存：
//
//	cache := cache.NewLRUCache(1000) // 最大 1000 个条目
//	cache.Set(context.Background(), "key", "value", 5*time.Minute)
//	value, err := cache.Get(context.Background(), "key")
//
// 使用缓存构建器：
//
//	cache := cache.NewMemoryCacheBuilder().
//	    InitialCapacity(1000).
//	    TTL(5*time.Minute).
//	    Build()
//
// 使用缓存获取器（带穿透保护）：
//
//	getter := cache.NewGetter(func(key string) (any, error) {
//	    // 从数据库加载
//	    return loadFromDB(key)
//	})
//	value, err := getter.Get("key")
package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// Cache 缓存操作接口。
//
// 所有缓存实现都应该实现此接口，
// 以便与 enhance 的依赖注入系统集成。
type Cache interface {
	// Get 获取指定键的缓存值，不存在返回 ErrNotFound。
	Get(ctx context.Context, key string) (any, error)

	// Set 设置缓存键值，ttl<=0 表示永不过期。
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Del 删除指定的缓存键。
	Del(ctx context.Context, keys ...string) error

	// Close 关闭缓存连接并释放资源。
	Close() error
}

// CacheInspector 可选的缓存检查接口。
//
// 用于需要检查键是否存在或获取过期时间的场景。
type CacheInspector interface {
	Cache
	// Exists 检查键是否存在且未过期。
	Exists(ctx context.Context, key string) (bool, error)
	// TTL 获取键的剩余过期时间。
	TTL(ctx context.Context, key string) (time.Duration, error)
}

// Getter 缓存旁路模式的值加载函数。
//
// 在缓存未命中时调用，从数据源加载值。
// 返回 nil 值不会被缓存。
type Getter func(ctx context.Context, key string) (any, error)

// LRUCache LRU（最近最少使用）缓存实现。
//
// 基于双向链表和哈希表实现 O(1) 时间复杂度的缓存操作。
// 支持 TTL 过期淘汰和容量限制淘汰。
// LRUCache 是并发安全的，所有操作都通过互斥锁保护。
type LRUCache struct {
	mu         sync.Mutex
	capacity   int
	items      map[string]*list.Element
	evictList  *list.List
	onEvict    func(key string, value any)
	defaultTTL time.Duration
}

// LRUOption LRU 缓存选项函数。
type LRUOption func(*LRUCache)

// ShardedLRUCache 分片 LRU 缓存实现。
//
// 将全局锁拆分为多个分段锁，提升高并发场景下的性能。
// 每个分片独立维护自己的 LRU 链表和哈希表。
type ShardedLRUCache struct {
	shards     []*lruShard
	capacity   int
	shardCount int
	onEvict    func(key string, value any)
}

// lruEntry LRU 缓存项内部结构。
type lruEntry struct {
	key       string
	value     any
	expiresAt time.Time
}

// lruShard 单个 LRU 分片内部结构。
type lruShard struct {
	mu        sync.RWMutex
	capacity  int
	items     map[string]*list.Element
	evictList *list.List
}
