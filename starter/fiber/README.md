# Fiber Starter

Fiber 高性能 Web 框架自动配置模块，基于 fasthttp 构建。

## 功能特性

- ✅ 自动配置 Fiber Web 服务器
- ✅ 支持 Prefork 模式
- ✅ 高性能路由和中间件
- ✅ 兼容 Express.js API
- ✅ 优雅关闭支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/fiber"
)
```

### 2. 配置文件

在 `application.json` 中添加 Fiber 配置：

```json
{
  "fiber": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 3000,
    "prefork": false,
    "body-limit": 4194304
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("fiber-demo"),
    )
    defer app.Stop()
    
    // 获取 Fiber App 实例
    app := core.MustGetBean[*fiber.App](app.Container())
    
    // 定义路由
    app.Get("/ping", func(c *fiber.Ctx) error {
        return c.SendString("pong")
    })
    
    // 参数路由
    app.Get("/users/:id", func(c *fiber.Ctx) error {
        id := c.Params("id")
        return c.JSON(fiber.Map{
            "id": id,
        })
    })
    
    // POST 请求
    app.Post("/users", func(c *fiber.Ctx) error {
        user := new(User)
        if err := c.BodyParser(user); err != nil {
            return c.Status(400).SendString(err.Error())
        }
        return c.JSON(user)
    })
}

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `fiber.enabled` | bool | false | 是否启用 Fiber |
| `fiber.host` | string | 0.0.0.0 | 服务器监听地址 |
| `fiber.port` | int | 3000 | 服务器端口 |
| `fiber.prefork` | bool | false | 是否启用 Prefork 模式 |
| `fiber.body-limit` | int | 4194304 | 请求体大小限制（字节） |
| `fiber.concurrency` | int | 262144 | 最大并发连接数 |
| `fiber.read-timeout` | int | 0 | 读取超时（秒） |
| `fiber.write-timeout` | int | 0 | 写入超时（秒） |
| `fiber.idle-timeout` | int | 0 | 空闲超时（秒） |

## 高级用法

### 中间件

```go
app := core.MustGetBean[*fiber.App](app.Container())

// CORS 中间件
app.Use(cors.New())

// Logger 中间件
app.Use(logger.New())

// Recover 中间件
app.Use(recover.New())
```

### 路由分组

```go
// API 路由组
api := app.Group("/api")
{
    api.Get("/users", getUsers)
    api.Post("/users", createUser)
}

// 管理路由组
admin := app.Group("/admin")
{
    admin.Get("/stats", getStats)
}
```

### 静态文件

```go
// 服务静态文件
app.Static("/public", "./public")
app.Static("/images", "./images")
```

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `fiber.enabled=true`

## 依赖

- `github.com/gofiber/fiber/v2`