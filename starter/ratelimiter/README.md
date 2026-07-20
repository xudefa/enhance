# RateLimiter Starter

RateLimiter 限流器自动配置模块，提供基于令牌桶算法的请求限流支持。

## 功能特性

- ✅ 自动配置限流器
- ✅ 支持令牌桶算法
- ✅ 支持可配置速率和突发
- ✅ 支持阻塞和非阻塞模式
- ✅ 支持上下文超时控制

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/ratelimiter"
)
```

### 2. 配置文件

在 `application.json` 中添加 RateLimiter 配置：

```json
{
  "ratelimiter": {
    "enabled": true,
    "rate": 10.0,
    "burst": 20
  }
}
```

### 3. 使用示例

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "golang.org/x/time/rate"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("ratelimiter-demo"),
    )
    defer app.Stop()
    
    // 获取限流器实例
    limiter := core.MustGetBean[*rate.Limiter](app.Container())
    
    // 在 HTTP 中间件中使用
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "请求过于频繁，请稍后重试", http.StatusTooManyRequests)
            return
        }
        
        fmt.Fprintf(w, "Hello, World!")
    })
    
    app.Start()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ratelimiter.enabled` | bool | false | 是否启用限流器 |
| `ratelimiter.rate` | float64 | 10.0 | 每秒允许的请求数 |
| `ratelimiter.burst` | int | 20 | 最大突发请求数 |

## 限流算法

### 令牌桶算法

限流器使用令牌桶算法：
- 令牌以固定速率生成
- 每个请求消耗一个令牌
- 当令牌桶满时，新生成的令牌被丢弃
- 当令牌桶空时，请求被拒绝

## 高级用法

### 阻塞模式

```go
limiter := core.MustGetBean[*rate.Limiter](app.Container())

// 等待直到获得令牌（阻塞）
err := limiter.Wait(ctx)
if err != nil {
    log.Fatal(err)
}

// 执行请求
doRequest()
```

### 动态调整速率

```go
limiter := core.MustGetBean[*rate.Limiter](app.Container())

// 调整速率
limiter.SetLimit(50.0)
limiter.SetBurst(100)
```

### 中间件集成

```go
// Gin 中间件
func RateLimiterMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "请求过于频繁",
            })
            return
        }
        c.Next()
    }
}

// 使用中间件
r := gin.Default()
r.Use(RateLimiterMiddleware(limiter))
```

## 启动顺序

- **优先级**: `OrderPriorityMiddleware` (-2000)
- **触发条件**: `ratelimiter.enabled=true`

## 依赖

- `golang.org/x/time`