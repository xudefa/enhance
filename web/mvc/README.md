# web/mvc 包 — MVC 控制器与 Web 启动器

> **所属层级**: Infrastructure Layer  
> **设计理念**: 接口抽象，插件化架构

## 概述

`web/mvc` 包位于 Infrastructure Layer，提供 Web MVC 框架的抽象接口定义，允许用户切换不同的网络框架实现。合并了原 `mvc` 包的所有功能。

### 核心功能

| 功能 | 说明 |
|------|------|
| **MVC 控制器** | Controller 接口，路由注册 |
| **注解路由** | @RestController、@GetMapping 等注解，自动扫描注册 |
| **WebStarter** | 启动器集成，自动注册控制器 |
| **中间件** | 日志、恢复、CORS 等中间件支持 |
| **WebSocket** | WebSocket 服务器和连接接口 |

## 架构设计

```
web/mvc/          ← 抽象层（接口定义）
├── context.go     ← Context 接口
├── router.go      ← Router 接口
├── server.go      ← Server 接口
├── controller.go  ← Controller 接口
├── websocket.go   ← WebSocket 接口
└── starter.go     ← WebStarter 启动器

web/server/       ← 实现层（使用原生 net/http）
├── DefaultContext  ← 实现 mvc.Context
├── DefaultRouter   ← 实现 mvc.Router
├── httpServer      ← 实现 mvc.Server
└── 中间件实现
```

## 核心接口

### Context

HTTP 请求上下文接口，封装请求和响应操作：

```go
type Context interface {
    RequestMethod() string
    RequestURI() string
    PathParam(name string) string
    Query(name string) string
    QueryDefault(name, defaultVal string) string
    Header(key string) string
    BindJSON(target any) error
    SetStatusCode(code int)
    SetHeader(key, value string)
    JSON(code int, data any) error
    String(code int, format string, args ...any)
    AbortWithStatus(code int)
    AbortWithStatusJSON(code int, body any)
    Next()
    IsAborted() bool
    Context() context.Context
    SetContext(ctx context.Context)
}
```

### Router

路由器接口，提供路由注册功能：

```go
type Router interface {
    GET(path string, handler HandlerFunc)
    POST(path string, handler HandlerFunc)
    PUT(path string, handler HandlerFunc)
    DELETE(path string, handler HandlerFunc)
    PATCH(path string, handler HandlerFunc)
    Group(prefix string) Router
    Use(middleware MiddlewareFunc)
}
```

### Controller

控制器接口，通过实现此接口注册路由：

```go
type Controller interface {
    Routes(router Router)
}
```

### Server

HTTP 服务器接口，定义服务器生命周期：

```go
type Server interface {
    Start() error
    Stop(ctx context.Context) error
    SetHandler(handler any)
    Use(middleware any)
}
```

### WebSocket

WebSocket 服务器和连接接口：

```go
type WebSocketServer interface {
    Start() error
    Stop(ctx context.Context) error
    SetHandler(handler WebSocketHandler)
}

type WebSocketConnection interface {
    ReadMessage() (messageType int, message []byte, err error)
    WriteMessage(messageType int, message []byte) error
    Close() error
}
```

## 使用示例

### 定义控制器

```go
type UserController struct {
    Service *UserService
}

func (c *UserController) Routes(router mvc.Router) {
    router.GET("/users/{id}", c.GetUser)
    router.POST("/users", c.CreateUser)
}

func (c *UserController) GetUser(ctx mvc.Context) {
    id := ctx.PathParam("id")
    ctx.JSON(http.StatusOK, map[string]string{"id": id})
}

func init() {
    mvc.RegisterController(&UserController{})
}
```

### 创建 Web Starter

```go
// 创建路由器和服务器
router := server.NewRouter()
httpServer := server.NewHTTPServer(
    server.WithHost(":8080"),
    server.WithReadTimeout(30*time.Second),
)

// 创建 WebStarter
starter := mvc.NewWebStarter(
    mvc.WithConfig(mvc.DefaultConfig()),
    mvc.WithRouter(router),
    mvc.WithServer(httpServer),
    mvc.WithMiddlewares([]core.MiddlewareFunc{
        server.LoggingMiddleware(),
        server.RecoveryMiddleware(),
    }),
)

// 注册到全局注册表
boot.RegisterStarter(starter)
```

## 设计原则

- **接口抽象**：所有核心组件都定义为接口，易于替换实现
- **零外部依赖**：仅使用 Go 标准库
- **参考 Spring MVC**：借鉴 Spring MVC 的设计理念
- **可扩展**：支持自定义路由器、服务器和中间件

---

## 注解路由

注解路由参考 Spring Boot 的 `@RestController` / `@GetMapping` 等注解，支持自动扫描和路由注册。

### 注解类型

| 注解 | 说明 | 使用方式 |
|------|------|---------|
| `@RestController` | 声明 REST 控制器 | 嵌入结构体，设置 base-path |
| `@GetMapping` | GET 方法映射 | 方法注释中添加 `@GetMapping(path=/path)` |
| `@PostMapping` | POST 方法映射 | 方法注释中添加 `@PostMapping(path=/path)` |
| `@PutMapping` | PUT 方法映射 | 方法注释中添加 `@PutMapping(path=/path)` |
| `@DeleteMapping` | DELETE 方法映射 | 方法注释中添加 `@DeleteMapping(path=/path)` |
| `@PatchMapping` | PATCH 方法映射 | 方法注释中添加 `@PatchMapping(path=/path)` |
| `@RequestMapping` | 通用请求映射 | 方法注释中添加 `@RequestMapping(path=/path)` |

### 使用示例

```go
type UserController struct {
    web.RestController `route:"base-path=/api/users"`
}

// @GetMapping(path=/)
func (c *UserController) ListUsers(ctx context.Context) ([]*User, error) {
    return getUsers(), nil
}

// @GetMapping(path=/{id})
func (c *UserController) GetUser(ctx context.Context, id string) (*User, error) {
    return getUserByID(id), nil
}

// @PostMapping(path=/)
func (c *UserController) CreateUser(ctx context.Context, user *User) (*User, error) {
    return createUser(user), nil
}

// @DeleteMapping(path=/{id})
func (c *UserController) DeleteUser(ctx context.Context, id string) error {
    return deleteUser(id)
}
```

### 扫描和注册

使用 core.ComponentScanner 进行扫描：

```go
// 创建容器
container := core.New()

// 设置 web 包的容器引用
web.SetContainer(container)

// 扫描组件（自动识别 @RestController 并注册路由）
scanner := core.NewComponentScanner("./internal", core.WithAutoInject(true))
scanner.Scan(container)

// 注册到标准 mux
mux := http.NewServeMux()
web.GlobalRouteRegistry.RegisterToMux(mux)

// 启动服务器
http.ListenAndServe(":8080", mux)
```

### 内容类型配置

```go
type UserController struct {
    web.RestController `route:"base-path=/api/users"`
}

// @GetMapping(path=/, produces=application/json)
func (c *UserController) ListUsers(ctx context.Context) ([]*User, error) {
    return getUsers(), nil
}

// @PostMapping(path=/, consumes=application/json, produces=application/json)
func (c *UserController) CreateUser(ctx context.Context, user *User) (*User, error) {
    return createUser(user), nil
}
```

### 最佳实践

1. **使用注解路由** - 代码更简洁，路由声明更直观
2. **base-path 统一前缀** - 同一控制器使用相同 base-path
3. **RESTful 风格** - 遵循 REST API 设计规范
4. **内容类型声明** - 明确声明 consumes 和 produces