# Echo Starter

Echo Web 框架自动配置模块，提供轻量级 HTTP 服务器支持。

## 功能特性

- ✅ 自动配置 Echo 服务器
- ✅ 内置中间件支持（Recover、Logger、CORS）
- ✅ 优雅启动和关闭
- ✅ 路由注册支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/echo"
)
```

### 2. 配置文件

在 `application.json` 中添加 Echo 配置：

```json
{
  "echo": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 8080,
    "hide_banner": false,
    "hide_port": false,
    "enable_recover": true,
    "enable_logger": true,
    "enable_cors": true
  }
}
```

### 3. 使用示例

```go
package main

import (
    "net/http"
    
    "github.com/labstack/echo/v4"
    
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/echo"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("echo-demo"),
    )
    defer app.Stop()
    
    // 获取 Echo 实例
    e := core.MustGetBean[*echo.Echo](app.Container())
    
    // 注册路由
    e.GET("/", func(c echo.Context) error {
        return c.String(http.StatusOK, "Hello, Echo!")
    })
    
    e.GET("/users/:id", func(c echo.Context) error {
        id := c.Param("id")
        return c.JSON(http.StatusOK, map[string]string{
            "user_id": id,
        })
    })
    
    // 启动服务器
    app.Start()
    app.WaitForSignal()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `echo.enabled` | bool | false | 是否启用 Echo |
| `echo.host` | string | 0.0.0.0 | 服务器地址 |
| `echo.port` | int | 8080 | 服务器端口 |
| `echo.hide_banner` | bool | false | 是否隐藏启动横幅 |
| `echo.hide_port` | bool | false | 是否隐藏端口日志 |
| `echo.enable_recover` | bool | true | 是否启用 panic 恢复 |
| `echo.enable_logger` | bool | true | 是否启用请求日志 |
| `echo.enable_cors` | bool | false | 是否启用 CORS |

## 高级用法

### 自定义中间件

```go
e := core.MustGetBean[*echo.Echo](app.Container())

// 添加自定义中间件
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // 前置处理
        println("请求开始")
        err := next(c)
        // 后置处理
        println("请求结束")
        return err
    }
})
```

### 路由分组

```go
api := e.Group("/api")
{
    api.GET("/users", getUsers)
    api.POST("/users", createUser)
    api.PUT("/users/:id", updateUser)
    api.DELETE("/users/:id", deleteUser)
}
```

### 优雅关闭

```go
echoConfig := core.MustGetBean[*echo.EchoAutoConfiguration](app.Container())

// 手动关闭
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
echoConfig.Stop(ctx)
```

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `echo.enabled=true`

## 依赖

- `github.com/labstack/echo/v4`