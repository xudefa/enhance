# Web 模块架构文档

## 概述

Web 模块采用面向接口编程的设计原则，支持灵活更换网络库和 Web 框架。通过抽象层和适配器模式，实现了核心业务逻辑与底层框架的解耦。

## 核心设计原则

### 1. 面向接口编程

所有关键组件都定义了清晰的接口：

- **EngineFactory**: 网络引擎工厂接口
- **Server**: HTTP 服务器接口
- **Router**: 路由器接口
- **Context**: HTTP 请求上下文接口
- **Controller**: 控制器接口

### 2. 适配器模式

提供通用适配器，简化第三方框架的集成：

- **ContextAdapter**: 将 http.ResponseWriter 和 *http.Request 适配为 Context 接口
- **ServerAdapter**: 将不同框架的服务器适配为统一的 Server 接口
- **RouterAdapter**: 将不同框架的路由器适配为统一的 Router 接口

### 3. 工厂模式

通过 EngineFactory 创建特定框架的组件实例，支持运行时切换。

## 架构层次

```
┌─────────────────────────────────────────────────────────┐
│                    应用层 (Application)                   │
│                  WebStarter, Controllers                  │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                   接口层 (Interfaces)                     │
│         EngineFactory, Server, Router, Context           │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                  适配器层 (Adapters)                       │
│      ContextAdapter, ServerAdapter, RouterAdapter        │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                 引擎实现层 (Engines)                       │
│    StdLib, Gin, Echo, Fiber, Fasthttp, Gnet, Evio       │
└─────────────────────────────────────────────────────────┘
```

## 核心接口

### EngineFactory

网络引擎工厂接口，负责创建特定框架的组件：

```go
type EngineFactory interface {
    Type() EngineType
    CreateRouter() (Router, error)
    CreateServer(opts ...ServerOption) (Server, error)
}
```

### Server

HTTP 服务器接口，定义服务器生命周期：

```go
type Server interface {
    Start() error
    Stop(ctx context.Context) error
    SetHandler(handler any)
    Use(m any)
}
```

### Router

路由器接口，负责路由注册：

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

### Context

HTTP 请求上下文接口，统一不同框架的上下文：

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

## 支持的引擎

### 默认引擎（已注册）

- **EngineStdLib**: 标准库 net/http

### 可扩展引擎（需导入对应包后注册）

- **EngineGin**: Gin 框架
- **EngineEcho**: Echo 框架
- **EngineFiber**: Fiber 框架
- **EngineFasthttp**: Fasthttp 框架
- **EngineGnet**: Gnet 框架
- **EngineEvio**: Evio 框架

## 使用示例

### 1. 使用默认引擎（标准库）

```go
// 创建路由器
router, _ := web.GlobalEngineRegistry.CreateRouter()

// 注册路由
router.GET("/hello", func(ctx web.Context) {
    ctx.String(200, "Hello, World!")
})

// 创建服务器
server, _ := web.GlobalEngineRegistry.CreateServer(
    web.WithHost("0.0.0.0"),
    web.WithPort(8080),
)

// 设置处理器并启动
server.SetHandler(router)
server.Start()
```

### 2. 切换到 Gin 框架

```go
// 假设已导入 Gin 适配器包
// import _ "github.com/xudefa/enhance/web/integration/gin"

// 切换到 Gin 引擎
web.GlobalEngineRegistry.SetDefault(web.EngineGin)

// 后续代码完全相同
router, _ := web.GlobalEngineRegistry.CreateRouter()
router.GET("/hello", func(ctx web.Context) {
    ctx.String(200, "Hello from Gin!")
})

server, _ := web.GlobalEngineRegistry.CreateServer(
    web.WithHost("0.0.0.0"),
    web.WithPort(8080),
)
server.SetHandler(router)
server.Start()
```

### 3. 注册自定义引擎

```go
// 实现 EngineFactory 接口
type CustomEngineFactory struct{}

func (f *CustomEngineFactory) Type() web.EngineType {
    return web.EngineType("custom")
}

func (f *CustomEngineFactory) CreateRouter() (web.Router, error) {
    // 使用 RouterAdapter 简化实现
    return web.NewRouterAdapter(
        getFunc, postFunc, putFunc, deleteFunc, patchFunc, groupFunc, useFunc,
    ), nil
}

func (f *CustomEngineFactory) CreateServer(opts ...web.ServerOption) (web.Server, error) {
    // 使用 ServerAdapter 简化实现
    return web.NewServerAdapter(
        startFunc, stopFunc, setHandlerFunc, useFunc,
    ), nil
}

// 注册引擎
web.GlobalEngineRegistry.Register(&CustomEngineFactory{})
```

## 扩展指南

### 添加新的网络引擎

1. **实现 EngineFactory 接口**
   - 定义引擎类型常量
   - 实现 Type() 方法返回引擎类型
   - 实现 CreateRouter() 方法创建路由器
   - 实现 CreateServer() 方法创建服务器

2. **实现 Router 接口**（或使用 RouterAdapter）
   - 实现所有 HTTP 方法的路由注册
   - 实现 Group() 方法支持路由组
   - 实现 Use() 方法支持中间件

3. **实现 Server 接口**（或使用 ServerAdapter）
   - 实现 Start() 方法启动服务器
   - 实现 Stop() 方法优雅关闭
   - 实现 SetHandler() 方法设置处理器
   - 实现 Use() 方法注册中间件

4. **实现 Context 接口**（或使用 ContextAdapter）
   - 实现所有请求和响应相关方法
   - 确保中间件链正确执行

5. **在 init() 中注册**
   ```go
   func init() {
       web.GlobalEngineRegistry.Register(&MyEngineFactory{})
   }
   ```

### 适配器使用场景

- **ContextAdapter**: 当框架的上下文可以转换为 http.ResponseWriter 和 *http.Request 时
- **ServerAdapter**: 当框架的服务器有自定义的启动/停止逻辑时
- **RouterAdapter**: 当框架的路由器有自定义的路由注册逻辑时

## 性能优化

1. **路由缓存**: 预编译路由模式，避免每次请求都解析
2. **中间件链**: 使用切片预分配，减少内存分配
3. **并发安全**: 使用读写锁保护共享状态
4. **零拷贝**: 尽可能避免不必要的数据拷贝

## 测试策略

- 单元测试覆盖所有核心接口
- 集成测试验证引擎切换功能
- 基准测试确保性能不退化
- 并发测试验证线程安全性

## 向后兼容性

- 默认使用标准库 net/http，保持向后兼容
- 所有现有代码无需修改即可运行
- 新功能是可选的，不影响现有行为

## 未来规划

- [ ] 添加 Gin 框架适配器
- [ ] 添加 Echo 框架适配器
- [ ] 添加 Fiber 框架适配器
- [ ] 添加 Fasthttp 引擎支持
- [ ] 添加 Gnet 高性能引擎支持
- [ ] 添加 HTTP/3 支持
- [ ] 添加 WebSocket 统一接口
- [ ] 添加 gRPC 网关支持