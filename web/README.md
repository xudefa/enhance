# web 包 — Web 框架

> **所属层级**: Infrastructure Layer  
> **设计理念**: HTTP 服务器，MVC 模式，插件化架构  
> **设计灵感**: Spring MVC + Spring Boot Web

## 概述

`web` 包提供完整的 Web 框架功能，支持 HTTP 服务器、MVC 控制器、注解路由、中间件、WebSocket、HTTP 客户端等特性。采用插件化架构，可灵活替换底层网络框架。

### 子包结构

| 子包 | 说明 |
|------|------|
| `web/` | 注解路由扫描器、引擎注册表、统一接口 |
| `web/server/` | HTTP 服务器、路由器、中间件、HTTP 客户端 |
| `web/mvc/` | MVC 控制器、WebSocket、Web 启动器 |
| `web/tls/` | TLS 证书管理、HTTPS 客户端、AES/RSA 加密工具 |

### 核心功能

| 功能 | 说明 |
|------|------|
| **HTTP 服务器** | 高性能 HTTP 服务器，支持超时配置、TLS |
| **路由器** | 支持 RESTful 路由、路径参数、路由组 |
| **中间件** | 日志、恢复、请求作用域、认证等中间件 |
| **MVC 控制器** | 支持控制器注册、路由组、内容协商 |
| **注解路由** | 通过结构体标签和方法注释自动注册路由 |
| **WebSocket** | WebSocket 连接支持 |
| **HTTP 客户端** | 内置重试、TLS、断路器的 HTTP 客户端 |
| **TLS/加密** | TLS 证书管理、AES/RSA 加密工具 |
| **插件化架构** | 支持替换底层网络框架（Gin、Hertz、Fasthttp 等） |

---

## 核心接口

### Server 接口

```go
type Server interface {
    Start() error
    Stop() error
    SetHandler(handler http.Handler)
    NewRouter() Router
}
```

### Router 接口

```go
type Router interface {
    GET(path string, handler HandlerFunc)
    POST(path string, handler HandlerFunc)
    PUT(path string, handler HandlerFunc)
    DELETE(path string, handler HandlerFunc)
    PATCH(path string, handler HandlerFunc)
    Use(middleware ...MiddlewareFunc)
    Group(path string) Router
}
```

### Context 接口

```go
type Context interface {
    RequestMethod() string
    PathParam(name string) string
    QueryParam(name string) string
    Header(name string) string
    JSON(code int, data any) error
    String(code int, s string) error
    Bind(target any) error
}
```

### EngineFactory 接口

```go
type EngineFactory interface {
    Type() EngineType
    CreateRouter() (Router, error)
    CreateServer(opts ...ServerOption) (Server, error)
}
```

---

## 快速开始

### 创建 HTTP 服务器

```go
package main

import (
    "net/http"
    "github.com/xudefa/enhance/web"
    "github.com/xudefa/enhance/web/server"
    "github.com/xudefa/enhance/web/mvc"
)

func main() {
    server := server.NewHTTPServer(
        server.WithHost(":8080"),
        server.WithReadTimeout(30*time.Second),
        server.WithWriteTimeout(30*time.Second),
    )

    router := server.NewRouter()
    router.GET("/hello", func(ctx mvc.Context) {
        ctx.JSON(http.StatusOK, map[string]string{"message": "Hello!"})
    })

    server.SetHandler(router)
    server.Start()
}
```

### 使用 MVC 控制器

```go
type UserController struct{}

func (c *UserController) Routes(router mvc.Router) {
    router.GET("/users", c.ListUsers)
    router.GET("/users/{id}", c.GetUser)
    router.POST("/users", c.CreateUser)
}

func (c *UserController) ListUsers(ctx mvc.Context) {
    ctx.JSON(http.StatusOK, []User{{ID: "1", Name: "Alice"}})
}

// 注册控制器
mvc.RegisterController(&UserController{})
```

### 使用中间件

```go
router := server.NewRouter()
router.Use(server.LoggingMiddleware())
router.Use(server.RecoveryMiddleware())
```

---

## API 参考

### 注解路由

#### 支持的注解

| 注解 | 说明 | 示例 |
|------|------|------|
| `@RestController` | 控制器结构体 | `@RestController(base-path=/api/users)` |
| `@GetMapping` | GET 请求映射 | `@GetMapping(path=/, produces=application/json)` |
| `@PostMapping` | POST 请求映射 | `@PostMapping(path=/, consumes=application/json)` |
| `@PutMapping` | PUT 请求映射 | `@PutMapping(path=/{id})` |
| `@DeleteMapping` | DELETE 请求映射 | `@DeleteMapping(path=/{id})` |
| `@PatchMapping` | PATCH 请求映射 | `@PatchMapping(path=/{id})` |

#### 注解属性

| 属性 | 说明 | 示例 |
|------|------|------|
| `base-path` | 控制器基础路径 | `base-path=/api/users` |
| `path` | 方法路由路径 | `path=/{id}` |
| `consumes` | 请求内容类型 | `consumes=application/json` |
| `produces` | 响应内容类型 | `produces=application/json` |

#### 使用示例

```go
type UserController struct {
    web.RestController `route:"base-path=/api/users"`
}

// @GetMapping(path=/, produces=application/json)
func (c *UserController) ListUsers(ctx context.Context) ([]*User, error) {
    return getUsers(), nil
}

// @GetMapping(path=/{id}, produces=application/json)
func (c *UserController) GetUser(ctx context.Context, id string) (*User, error) {
    return getUserByID(id), nil
}

// @PostMapping(path=/, consumes=application/json, produces=application/json)
func (c *UserController) CreateUser(ctx context.Context, user *User) (*User, error) {
    return createUser(user), nil
}
```

### 插件化架构

#### 支持的引擎

| 引擎类型 | 常量 | 说明 | 状态 |
|---------|------|------|------|
| StdLib | `EngineStdLib` | 标准库 net/http | ✅ 已实现 |
| Gin | `EngineGin` | Gin 框架 | 📝 示例代码 |
| Hertz | `EngineHertz` | Hertz 框架 | 📝 示例代码 |
| Fasthttp | `EngineFasthttp` | Fasthttp 框架 | 📝 示例代码 |

#### 切换引擎

```go
// 注册 Gin 引擎
web.GlobalEngineRegistry.Register(&web.GinEngineFactory{})

// 设置默认引擎
web.GlobalEngineRegistry.SetDefault(web.EngineGin)

// 创建路由器（使用 Gin 引擎）
router, _ := web.GlobalEngineRegistry.CreateRouter()
router.GET("/api/users", func(ctx web.Context) {
    ctx.JSON(200, map[string]string{"message": "Hello"})
})
```

### HTTP 客户端

#### 基本使用

```go
client := server.NewClient("https://api.example.com",
    server.WithClientTimeout(30*time.Second),
)

resp, err := client.Get(ctx, "/users")
```

#### 重试客户端

```go
retryableClient := server.NewRetryableClient(client,
    server.WithMaxAttempts(3),
    server.WithRetryStrategy(server.NewExponentialBackoff(
        100*time.Millisecond,
        10*time.Second,
        500, 502, 503, 504,
    )),
)
```

#### 断路器客户端

```go
circuitClient := server.NewCircuitBreakerClient(client,
    server.WithCircuitMaxFailures(5),
    server.WithCircuitResetTimeout(30*time.Second),
    server.WithFallback(func(ctx context.Context) (*server.HttpResponse, error) {
        return &server.HttpResponse{
            StatusCode: 503,
            Body:       []byte(`{"error": "service unavailable"}`),
        }, nil
    }),
)
```

### TLS 支持

#### HTTPS 服务器

```go
tlsConfig, err := server.LoadTLSConfig("cert.pem", "key.pem")
if err != nil {
    log.Fatal(err)
}

httpsServer := server.NewHTTPServer(
    server.WithHost(":8443"),
    server.WithTLS(tlsConfig),
)
```

#### 加密工具

```go
// AES 加密/解密
ciphertext, err := tls.AESEncrypt(plaintext, key, iv)
plaintext, err := tls.AESDecrypt(ciphertext, key, iv)

// RSA 加密/解密
ciphertext, err := tls.RSAEncrypt(publicKey, plaintext)
plaintext, err := tls.RSADecrypt(privateKey, ciphertext)
```

---

## 使用示例

### 完整 Web 应用示例

```go
package main

import (
    "net/http"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/web/mvc"
)

type UserController struct{}

func (c *UserController) Routes(router mvc.Router) {
    router.GET("/users", c.ListUsers)
    router.GET("/users/{id}", c.GetUser)
    router.POST("/users", c.CreateUser)
    router.PUT("/users/{id}", c.UpdateUser)
    router.DELETE("/users/{id}", c.DeleteUser)
}

func (c *UserController) ListUsers(ctx mvc.Context) {
    ctx.JSON(http.StatusOK, []User{
        {ID: "1", Name: "Alice"},
        {ID: "2", Name: "Bob"},
    })
}

func (c *UserController) GetUser(ctx mvc.Context) {
    id := ctx.PathParam("id")
    ctx.JSON(http.StatusOK, User{ID: id, Name: "User " + id})
}

func (c *UserController) CreateUser(ctx mvc.Context) {
    var user User
    if err := ctx.Bind(&user); err != nil {
        ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusCreated, user)
}

func (c *UserController) UpdateUser(ctx mvc.Context) {
    id := ctx.PathParam("id")
    ctx.JSON(http.StatusOK, map[string]string{"message": "Updated " + id})
}

func (c *UserController) DeleteUser(ctx mvc.Context) {
    id := ctx.PathParam("id")
    ctx.JSON(http.StatusOK, map[string]string{"message": "Deleted " + id})
}

func main() {
    mvc.RegisterController(&UserController{})
    
    app, _ := boot.NewApplication(
        boot.WithAppName("my-web-app"),
    )
    app.Start()
    defer app.Stop()
    
    app.WaitForSignal()
}
```

### WebSocket 示例

```go
type ChatController struct{}

func (c *ChatController) Routes(router mvc.Router) {
    router.GET("/ws/chat", c.HandleWebSocket)
}

func (c *ChatController) HandleWebSocket(ctx mvc.Context) {
    conn, err := ctx.Upgrade()
    if err != nil {
        return
    }
    defer conn.Close()

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        // 处理消息
        conn.WriteMessage(websocket.TextMessage, message)
    }
}
```

---

## 最佳实践

### 1. 使用 MVC 控制器组织路由

```go
// ✅ 推荐：使用控制器组织相关路由
type UserController struct{}

func (c *UserController) Routes(router mvc.Router) {
    router.GET("/users", c.ListUsers)
    router.POST("/users", c.CreateUser)
}

// ⚠️ 不推荐：散乱的路由注册
router.GET("/users", listUsersHandler)
router.POST("/users", createUserHandler)
```

### 2. 合理使用中间件

```go
// ✅ 推荐：按顺序添加中间件
router.Use(server.LoggingMiddleware())
router.Use(server.RecoveryMiddleware())
router.Use(server.AuthMiddleware())

// ⚠️ 不推荐：在每个路由上重复添加中间件
router.GET("/users", authMiddleware, listUsersHandler)
router.POST("/users", authMiddleware, createUserHandler)
```

### 3. 使用注解路由简化代码

```go
// ✅ 推荐：使用注解路由
type UserController struct {
    web.RestController `route:"base-path=/api/users"`
}

// @GetMapping(path=/, produces=application/json)
func (c *UserController) ListUsers(ctx context.Context) ([]*User, error) {
    return getUsers(), nil
}

// ⚠️ 不推荐：手动注册路由
func init() {
    mvc.RegisterController(&UserController{})
}
```

### 4. 使用 HTTP 客户端的重试和断路器

```go
// ✅ 推荐：使用重试和断路器
client := server.NewRetryableClient(baseClient,
    server.WithMaxAttempts(3),
    server.WithRetryStrategy(server.NewExponentialBackoff(...)),
)

// ⚠️ 不推荐：不使用重试机制
resp, err := client.Get(ctx, "/users")
if err != nil {
    // 直接失败
}
```

### 5. 使用插件化架构切换引擎

```go
// ✅ 推荐：通过接口编程，不依赖具体引擎
router, _ := web.GlobalEngineRegistry.CreateRouter()
router.GET("/api/users", handler)

// 随时可以切换引擎
web.GlobalEngineRegistry.SetDefault(web.EngineGin)
```