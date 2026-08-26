# cache 包 — 缓存抽象

> **所属层级**: Infrastructure Layer  
> **设计理念**: 统一接口，Cache-Aside 模式  
> **设计灵感**: Spring Cache + Go 惯用法

## 概述

`cache` 包提供统一的缓存操作接口抽象，支持不同的缓存实现（如 Redis、内存缓存、TTL 本地缓存）无缝替换。核心设计模式包括 Cache-Aside（缓存旁路）模式，通过 `Getter` 函数实现缓存未命中时的数据加载。

### 核心功能

| 功能 | 说明 |
|------|------|
| **统一接口** | Cache 接口定义标准缓存操作 |
| **Cache-Aside 模式** | 缓存未命中时自动从数据源加载 |
| **内存缓存** | MemoryCache 适用于测试和轻量场景 |
| **高性能本地缓存** | TTLCache 支持 LRU 淘汰 + TTL |
| **批量操作** | 支持批量获取、设置、删除 |
| **TTL 过期** | 支持为每个缓存项设置独立的过期时间 |

---

## 核心接口

### Cache 接口

```go
type Cache interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    Close() error
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Get` | 获取指定键的缓存值，键不存在返回 `ErrNotFound`，已过期返回 `ErrCacheMiss` |
| `Set` | 设置缓存键值，可指定 TTL 过期时间。`ttl <= 0` 表示永不过期 |
| `Del` | 删除一个或多个缓存键 |
| `Exists` | 检查键是否存在（已过期的键视为不存在） |
| `TTL` | 获取键的剩余过期时间。未设置过期时间返回 0，不存在返回 `ErrNotFound` |
| `Close` | 关闭缓存连接并释放资源 |

### Getter — 数据加载函数

```go
type Getter func(ctx context.Context, key string) (any, error)
```

用于 Cache-Aside（缓存旁路）模式：当缓存未命中时，通过 Getter 从数据源加载数据并回填缓存。

### 错误 sentinel

```go
var (
    ErrNotFound  = errors.New("cache: key not found")
    ErrCacheMiss = errors.New("cache: key expired or not found")
)
```

| 错误 | 说明 |
|------|------|
| `ErrNotFound` | 缓存键不存在（从未设置） |
| `ErrCacheMiss` | 缓存键已过期或未命中 |

---

## 快速开始

### 创建内存缓存

```go
package main

import (
    "context"
    "time"
    "github.com/xudefa/enhance/cache"
)

func main() {
    c := cache.NewMemoryCache()
    ctx := context.Background()

    // 设置缓存，5 分钟过期
    err := c.Set(ctx, "user:1", userData, 5*time.Minute)

    // 获取缓存
    val, err := c.Get(ctx, "user:1")
}
```

### Cache-Aside 模式

```go
val, err := c.GetWithGetter(ctx, "user:1", func(ctx context.Context, key string) (any, error) {
    // 从数据库加载
    var user User
    db.First(&user, 1)
    return user, nil
})
```

---

## API 参考

### MemoryCache — 内存缓存实现

基于 `sync.RWMutex` 和 `map[string]cacheItem` 的并发安全内存缓存实现，支持 TTL 过期和延迟清理。

#### 创建

```go
cache := cache.NewMemoryCache()
```

#### 使用场景

- 测试环境
- 轻量级应用
- 缓存原型开发

#### 扩展方法

| 方法 | 说明 |
|------|------|
| `GetWithGetter` | Cache-Aside 模式：缓存未命中时从数据源加载 |
| `GetMulti` | 批量获取多个键的缓存值 |
| `SetMulti` | 批量设置多个缓存键值 |
| `DeleteMulti` | 批量删除多个缓存键 |
| `Clear` | 清空所有缓存 |

### TinyLFUCache — 高性能本地缓存

基于 LRU（最近最少使用）淘汰策略的高性能本地缓存实现，适用于需要控制内存使用量的场景。

#### 创建

```go
// 使用默认配置（最大 1000 项，默认 TTL 5 分钟）
cache := cache.NewTinyLFUCache()

// 自定义配置
cache := cache.NewTinyLFUCache(
    cache.WithTinyLFUMaxSize(5000),
    cache.WithTinyLFUDefaultTTL(10*time.Minute),
)
```

#### 特性

| 特性 | 说明 |
|------|------|
| **LRU 淘汰** | 当缓存达到最大容量时，自动淘汰最久未使用的项 |
| **TTL 支持** | 支持为每个缓存项设置独立的过期时间 |
| **并发安全** | 使用 `sync.RWMutex` 保护并发访问 |
| **O(1) 操作** | 基于 map + 双向链表实现，Get/Set 都是 O(1) 时间复杂度 |

#### 使用示例

```go
c := cache.NewTinyLFUCache(cache.WithMaxSize(1000))
ctx := context.Background()

// 设置缓存
_ = c.Set(ctx, "user:1", userData, 5*time.Minute)

// 获取缓存（会更新 LRU 位置）
val, err := c.Get(ctx, "user:1")

// 查看缓存统计
stats := c.Stats()
```

---

## 使用示例

### 基础操作

```go
c := cache.NewMemoryCache()
ctx := context.Background()

// 设置缓存，5 分钟过期
err := c.Set(ctx, "user:1", userData, 5*time.Minute)

// 获取缓存
val, err := c.Get(ctx, "user:1")
if errors.Is(err, cache.ErrNotFound) {
    // 键不存在
}
if errors.Is(err, cache.ErrCacheMiss) {
    // 键已过期
}

// 检查是否存在
exists, _ := c.Exists(ctx, "user:1")

// 获取剩余 TTL
ttl, _ := c.TTL(ctx, "user:1")

// 批量操作
items := map[string]any{"a": 1, "b": 2}
_ = c.SetMulti(ctx, items, time.Minute)
result, _ := c.GetMulti(ctx, []string{"a", "b"})
_ = c.DeleteMulti(ctx, []string{"a", "b"})

// 清空
_ = c.Clear(ctx)
```

### Cache-Aside 模式

```go
val, err := c.GetWithGetter(ctx, "user:1", func(ctx context.Context, key string) (any, error) {
    // 从数据库加载
    var user User
    db.First(&user, 1)
    return user, nil
})
```

### 与依赖注入集成

```go
// 注册为 Bean
container.Register(
    reflect.TypeOf(&cache.MemoryCache{}),
    core.Bean(cache.NewMemoryCache()),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    Cache cache.Cache `inject:"cache"`
}
```

---

## 最佳实践

### 1. 使用 Cache-Aside 模式简化缓存逻辑

```go
// ✅ 推荐：使用 GetWithGetter 自动回填缓存
val, err := c.GetWithGetter(ctx, "user:1", func(ctx context.Context, key string) (any, error) {
    return db.GetUser(id)
})

// ⚠️ 不推荐：手动实现缓存逻辑
val, err := c.Get(ctx, key)
if err == cache.ErrNotFound {
    val, err = db.GetUser(id)
    if err == nil {
        c.Set(ctx, key, val, ttl)
    }
}
```

### 2. 合理设置 TTL

```go
// ✅ 推荐：根据数据更新频率设置合理的 TTL
c.Set(ctx, "user:1", userData, 5*time.Minute)    // 用户数据 5 分钟
c.Set(ctx, "config:app", config, 1*time.Hour)     // 配置数据 1 小时

// ⚠️ 不推荐：所有数据使用相同 TTL
c.Set(ctx, key, value, 10*time.Minute)
```

### 3. 使用 TinyLFUCache 控制内存使用

```go
// ✅ 推荐：设置最大缓存大小
c := cache.NewTinyLFUCache(
    cache.WithTinyLFUMaxSize(5000),
    cache.WithTinyLFUDefaultTTL(10*time.Minute),
)

// ⚠️ 不推荐：不限制缓存大小，可能导致内存溢出
c := cache.NewMemoryCache()
```

### 4. 批量操作提升性能

```go
// ✅ 推荐：使用批量操作
items := map[string]any{"user:1": u1, "user:2": u2}
c.SetMulti(ctx, items, time.Minute)

// ⚠️ 不推荐：循环单个设置
for key, value := range items {
    c.Set(ctx, key, value, time.Minute)
}
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将缓存注册为 Bean
container.Register(
    reflect.TypeOf(&cache.MemoryCache{}),
    core.Bean(cache.NewMemoryCache()),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    Cache cache.Cache `inject:"cache"`
}
```