# Chi Starter

Chi HTTP 路由器自动配置模块，提供轻量级 HTTP 路由支持。

## 功能特性

- ✅ 自动配置 Chi 路由器
- ✅ 内置中间件支持（Recover、Logger、RequestID、RealIP）
- ✅ 优雅启动和关闭
- ✅ 路由分组支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/chi"
)
```

### 2. 配置文件

在 `application.json` 中添加 Chi 配置：

```json
{
  "chi": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 8080,
    "enable_recover": true,
    "enable_logger": true,
    "enable_request_id": true,
    "enable_real_ip": false
  }
}
```

### 3. 使用示例

```go
package main

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/chi"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("chi-demo"),
    )
    defer app.Stop()
    
    // 获取 Chi 路由器
    router := core.MustGetBean[*chi.Mux](app.Container())
    
    // 注册路由
    router.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, Chi!"))
    })
    
    router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
        id := chi.URLParam(r, "id")
        w.Write([]byte("User ID: " + id))
    })
    
    // 启动服务器
    app.Start()
    app.WaitForSignal()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `chi.enabled` | bool | false | 是否启用 Chi |
| `chi.host` | string | 0.0.0.0 | 服务器地址 |
| `chi.port` | int | 8080 | 服务器端口 |
| `chi.enable_recover` | bool | true | 是否启用 panic 恢复 |
| `chi.enable_logger` | bool | true | 是否启用请求日志 |
| `chi.enable_request_id` | bool | true | 是否启用请求 ID |
| `chi.enable_real_ip` | bool | false | 是否启用真实 IP |

## 高级用法

### 路由分组

```go
router := core.MustGetBean[*chi.Mux](app.Container())

// API 路由分组
router.Route("/api", func(r chi.Router) {
    r.Get("/users", getUsers)
    r.Post("/users", createUser)
    r.Put("/users/{id}", updateUser)
    r.Delete("/users/{id}", deleteUser)
})
```

### 自定义中间件

```go
router := core.MustGetBean[*chi.Mux](app.Container())

// 添加自定义中间件
router.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 前置处理
        println("请求开始")
        next.ServeHTTP(w, r)
        // 后置处理
        println("请求结束")
    })
})
```

### 优雅关闭

```go
chiConfig := core.MustGetBean[*chi.ChiAutoConfiguration](app.Container())

// 手动关闭
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
chiConfig.Stop(ctx)
```

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `chi.enabled=true`

## 依赖

- `github.com/go-chi/chi/v5`
- `github.com/go-chi/chi/v5/middleware`