# web/server 包 — HTTP 服务器与客户端

> **所属层级**: Infrastructure Layer  
> **设计理念**: 高性能 HTTP，零外部依赖

## 概述

`web/server` 包位于 Infrastructure Layer，提供 HTTP 服务器和客户端的完整实现，使用 Go 原生 `net/http` 包。合并了原 `net` 包的服务器、路由器、中间件和 HTTP 客户端功能。

### 核心组件

| 组件 | 说明 |
|------|------|
| **HTTPServer** | 高性能 HTTP 服务器，支持超时配置 |
| **DefaultRouter** | RESTful 路由器，支持路径参数 |
| **DefaultContext** | HTTP 请求上下文实现 |
| **中间件** | 日志、恢复、请求作用域等中间件 |
| **NetClient** | 默认 HTTP 客户端 |
| **RetryableClient** | 支持重试的 HTTP 客户端 |
| **CircuitBreakerClient** | 带断路器的 HTTP 客户端 |
| **TLS 支持** | TLS 配置和 HTTPS 客户端 |

## HTTP 服务器

### 创建服务器

```go
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
```

### TLS 支持

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

## 路由器

### 基本路由

```go
router := server.NewRouter()

// RESTful 路由
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.GET("/users/{id}", getUser)
router.PUT("/users/{id}", updateUser)
router.DELETE("/users/{id}", deleteUser)
router.PATCH("/users/{id}", patchUser)

// 路由组
api := router.Group("/api")
api.GET("/users", listUsers)

v1 := router.Group("/api/v1")
v1.GET("/users", listUsersV1)
```

### 路径参数

支持 `{name}` 格式的路径参数：

```go
router.GET("/users/{id}/posts/{postId}", handler)

func handler(ctx mvc.Context) {
    id := ctx.PathParam("id")
    postId := ctx.PathParam("postId")
}
```

## 中间件

### 内置中间件

```go
router := server.NewRouter()

// 日志中间件
router.Use(server.LoggingMiddleware())

// 恢复中间件（捕获 panic）
router.Use(server.RecoveryMiddleware())

// 请求作用域中间件
router.Use(server.RequestScopeMiddleware())
```

### 自定义中间件

```go
func AuthMiddleware(next mvc.HandlerFunc) mvc.HandlerFunc {
    return func(ctx mvc.Context) {
        token := ctx.Header("Authorization")
        if token == "" {
            ctx.AbortWithStatus(http.StatusUnauthorized)
            return
        }
        next(ctx)
    }
}

router.Use(AuthMiddleware)
```

## HTTP 客户端

### 基本使用

```go
client := server.NewClient("https://api.example.com",
    server.WithClientTimeout(10*time.Second),
)

// GET 请求
resp, err := client.Get(ctx, "/users")
if err != nil {
    log.Fatal(err)
}

// POST 请求
data := map[string]any{"name": "张三", "email": "zhangsan@example.com"}
resp, err = client.Post(ctx, "/users", data)
```

### 重试客户端

```go
client := server.NewClient("https://api.example.com")

retryableClient := server.NewRetryableClient(client,
    server.WithMaxAttempts(3),
    server.WithRetryStrategy(server.NewExponentialBackoff(
        100*time.Millisecond,
        10*time.Second,
        500, 502, 503, 504,
    )),
    server.WithOnRetry(func(attempt int, resp *server.HttpResponse, err error) {
        log.Printf("retry attempt %d", attempt)
    }),
)

resp, err := retryableClient.Get(ctx, "/users")
```

### 断路器客户端

```go
client := server.NewClient("https://api.example.com")

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

resp, err := circuitClient.Get(ctx, "/users")
```

### 请求作用域

```go
// 在请求作用域中存储和获取数据
func middleware(ctx mvc.Context) {
    scope := server.GetRequestScope(ctx)
    scope.Set("userID", "123")
    
    // 在后续处理中获取
    userID := scope.Get("userID").(string)
}
```

## 设计原则

- **参考 Spring Boot**：借鉴 Spring Boot 的设计理念
- **接口实现分离**：web/mvc 定义接口，web/server 提供实现
- **零外部依赖**：仅使用 Go 标准库
- **可扩展**：支持自定义中间件、客户端配置