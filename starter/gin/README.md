# Gin Starter

Gin Web 框架自动配置模块，提供 HTTP 服务支持。

## 功能特性

- ✅ 自动配置 Gin Web 服务器
- ✅ 支持 debug/release/test 模式
- ✅ 内置 Recover 和 Logger 中间件
- ✅ 优雅关闭支持
- ✅ 自定义中间件支持
- ✅ 自动集成链路追踪中间件
- ✅ 自动挂载 Actuator 监控端点

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/gin"
)
```

### 2. 配置文件

在 `application.json` 中添加 Gin 配置：

```json
{
  "gin": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 8080,
    "mode": "debug",
    "enable_recover": true,
    "enable_logger": true
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/gin-gonic/gin"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("gin-demo"),
    )
    defer app.Stop()
    
    // 获取 Gin Engine 实例
    engine := core.MustGetBean[*gin.Engine](app.Container())
    
    // 定义路由
    engine.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    
    // 用户组
    v1 := engine.Group("/api/v1")
    {
        v1.GET("/users", func(c *gin.Context) {
            c.JSON(200, gin.H{"users": []string{}})
        })
    }
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `gin.enabled` | bool | false | 是否启用 Gin |
| `gin.host` | string | 0.0.0.0 | 服务器监听地址 |
| `gin.port` | int | 8080 | 服务器端口 |
| `gin.mode` | string | debug | 运行模式（debug/release/test） |
| `gin.enable_recover` | bool | true | 是否启用 Recover 中间件 |
| `gin.enable_logger` | bool | true | 是否启用 Logger 中间件 |

## 高级用法

### 自定义中间件

```go
engine := core.MustGetBean[*gin.Engine](app.Container())

// 添加自定义中间件
engine.Use(func(c *gin.Context) {
    start := time.Now()
    c.Next()
    duration := time.Since(start)
    log.Printf("Request took %s", duration)
})
```

### 路由分组

```go
// API 路由组
api := engine.Group("/api")
{
    api.GET("/health", healthCheck)
    api.POST("/users", createUser)
    api.GET("/users/:id", getUser)
}

// 管理路由组
admin := engine.Group("/admin")
{
    admin.GET("/stats", getStats)
}
```

### 静态文件服务

```go
// 服务静态文件
engine.Static("/assets", "./assets")
engine.StaticFile("/favicon.ico", "./favicon.ico")
```

### 参数绑定

```go
type User struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"gte=0,lte=130"`
}

engine.POST("/users", func(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, user)
})
```

### 链路追踪集成

当启用 `tracing.enabled=true` 时，Gin Starter 会自动注册链路追踪中间件：

```json
{
  "gin": {
    "enabled": true
  },
  "tracing": {
    "enabled": true,
    "service_name": "my-service",
    "sampling_rate": 1.0
  }
}
```

所有请求（包括业务接口和 Actuator 监控端点）都会自动记录链路追踪日志：

```go
// 获取 Tracer 实例
tracer := core.MustGetBean[*tracing.Tracer](app.Container())

// 查看链路数据
spans := tracer.GetSpans()
for _, span := range spans {
    fmt.Printf("TraceID: %s, SpanID: %s, Name: %s\n", 
        span.TraceID, span.SpanID, span.Name)
}
```

### Actuator 监控端点集成

当启用 `actuator.enabled=true` 时，Gin Starter 会自动挂载监控端点：

```json
{
  "gin": {
    "enabled": true
  },
  "actuator": {
    "enabled": true,
    "path": "/actuator"
  }
}
```

可用的监控端点：

| 端点 | 路径 | 说明 |
|------|------|------|
| Health | `/actuator/health` | 健康检查 |
| Metrics | `/actuator/metrics` | 应用指标 |
| Env | `/actuator/env` | 环境信息 |
| Beans | `/actuator/beans` | Bean 列表 |
| Info | `/actuator/info` | 应用信息 |

监控端点同样支持链路追踪，每次调用都会生成 TraceID 和 SpanID，并通过响应头返回。

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `gin.enabled=true`

## 依赖

- `github.com/gin-gonic/gin`