# Redis Starter

Redis 缓存自动配置模块，提供分布式缓存支持。

## 功能特性

- ✅ 自动配置 Redis 连接
- ✅ 实现 `cache.Cache` 接口
- ✅ 支持键前缀隔离
- ✅ 连接池管理
- ✅ 连接健康检查

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/redis"
)
```

### 2. 配置文件

在 `application.json` 中添加 Redis 配置：

```json
{
  "redis": {
    "enabled": true,
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0,
    "prefix": "enhance:",
    "pool_size": 10
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    "time"
    
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/cache"
    "github.com/xudefa/enhance/core"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("redis-demo"),
    )
    defer app.Stop()
    
    // 获取缓存实例
    cache := core.MustGetBean[cache.Cache](app.Container())
    ctx := context.Background()
    
    // 设置缓存
    cache.Set(ctx, "user:1", "John", 5*time.Minute)
    
    // 获取缓存
    val, err := cache.Get(ctx, "user:1")
    if err != nil {
        // 处理错误
    }
    
    // 删除缓存
    cache.Del(ctx, "user:1")
    
    // 检查键是否存在
    exists, _ := cache.Exists(ctx, "user:1")
    
    // 获取过期时间
    ttl, _ := cache.TTL(ctx, "user:1")
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `redis.enabled` | bool | false | 是否启用 Redis |
| `redis.host` | string | localhost | Redis 服务器地址 |
| `redis.port` | int | 6379 | Redis 端口 |
| `redis.password` | string | "" | Redis 密码 |
| `redis.db` | int | 0 | Redis 数据库编号 |
| `redis.prefix` | string | enhance: | 键前缀 |
| `redis.pool_size` | int | 10 | 连接池大小 |

## 高级用法

### 获取底层 Redis Client

```go
redisCache := core.MustGetBean[cache.Cache](app.Container())
if rc, ok := redisCache.(*redis.RedisCache); ok {
    client := rc.GetClient()
    // 使用原生 Redis 客户端
}
```

### 多 Redis 实例

如需支持多 Redis 实例，可以创建多个 `RedisCache` 实例并注册到容器。

## 启动顺序

- **优先级**: `OrderPriorityDataLayer` (-2000)
- **触发条件**: `redis.enabled=true`

## 依赖

- `github.com/redis/go-redis/v9`