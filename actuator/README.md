# actuator 包 — 运维监控端点

> **所属层级**: Infrastructure Layer  
> **设计理念**: 生产就绪，可视化监控  
> **设计灵感**: Spring Boot Actuator

## 📖 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [可用端点](#可用端点)
- [HTTP 端点自动挂载](#http-端点自动挂载)
- [健康检查系统](#健康检查系统)
- [敏感信息保护](#敏感信息保护)
- [配置选项](#配置选项)
- [完整示例](#完整示例)
- [最佳实践](#最佳实践)
- [故障排查](#故障排查)

---

## 概述

`actuator` 包提供类似 Spring Boot Actuator 的运维端点支持，便于在生产环境中监控和管理应用。

### 核心功能

| 功能 | 说明 |
|------|------|
| 🔍 **健康检查** | 聚合健康指标，返回整体健康状态 |
| 📊 **指标收集** | 收集并返回所有注册的指标 |
| ⚙️ **环境信息** | 显示环境配置源及其属性 |
| 📦 **Bean 列表** | 列出 IoC 容器中注册的所有 Bean |
| 🔒 **敏感信息过滤** | 自动检测和掩盖敏感配置信息 |
| 🌐 **Admin 可视化** | Spring Boot Admin 风格的应用管理 |

### 子包结构

| 子包 | 说明 |
|------|------|
| `actuator/` | 运维端点管理器（健康检查、指标、环境、Bean 列表） |
| `actuator/health/` | 健康检查核心（Indicator 接口、Aggregator 聚合器） |
| `actuator/admin/` | Spring Boot Admin 风格可视化监控 |

---

## 快速开始

### 方式一：自动配置（推荐）

使用 boot 框架，Actuator 会自动配置和挂载：

```yaml
# application.yaml
actuator:
  enabled: true
```

```go
package main

import (
    "github.com/xudefa/enhance/boot"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("my-app"),
    )
    
    app.Start()
    defer app.Stop()
    
    // Actuator 端点已自动挂载到 HTTP 服务器
    // 访问 http://localhost:8080/actuator/health
}
```

### 方式二：手动配置

```go
package main

import (
    "net/http"
    "github.com/xudefa/enhance/actuator"
    "github.com/xudefa/enhance/context"
)

func main() {
    ctx := context.NewApplicationContext(container, env)
    
    // 创建 Actuator
    act := actuator.New(ctx)
    
    // 注册路由
    mux := http.NewServeMux()
    config := actuator.DefaultRouteConfig()
    registrar := &actuator.StdRouteRegistrar{Mux: mux}
    act.RegisterRoutes(registrar, config)
    
    http.ListenAndServe(":8080", mux)
}
```

---

## 可用端点

### Actuator 端点

| 端点 | 路径 | 说明 |
|------|------|------|
| Health | `/actuator/health` | 健康检查，聚合所有健康指标 |
| Metrics | `/actuator/metrics` | 应用指标 |
| Env | `/actuator/env` | 环境信息（自动过滤敏感信息） |
| Beans | `/actuator/beans` | IoC 容器中注册的所有 Bean |
| Info | `/actuator/info` | 应用信息 |
| Prometheus | `/metrics` | Prometheus 格式指标 |
| pprof | `/debug/pprof/*` | 调试端点（需启用 ExposeDebug） |

### Admin 可视化端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/admin/applications` | GET | 列出所有应用 |
| `/admin/instances` | GET | 列出所有实例 |
| `/admin/instances/{id}/health` | GET | 获取实例健康 |
| `/admin/instances/{id}/metrics` | GET | 获取实例指标 |
| `/admin/register` | POST | 注册实例 |
| `/admin/deregister` | POST | 注销实例 |

---

## HTTP 端点自动挂载

### 架构设计

Actuator 使用 `HttpEndpointRegistry` 接口实现**框架无关**的端点自动挂载，支持自动挂载到任意 HTTP 框架。

### 挂载策略（按优先级）

1. **HttpEndpointRegistry 接口**（推荐，框架无关）
2. **HttpHandlerRegistry 接口**（简化版）
3. **RouteRegistrar 接口**（向后兼容）
4. **独立 HTTP 服务器**（降级方案）

### 已集成的框架

| 框架 | 实现位置 | 状态 |
|------|----------|------|
| **Gin** | `starter/gin/endpoint_registry.go` | ✅ 已集成 |
| **Fiber** | `starter/fiber/endpoint_registry.go` | ✅ 已集成 |
| **Echo** | `starter/echo/endpoint_registry.go` | ✅ 已集成 |
| **Chi** | `starter/chi/endpoint_registry.go` | ✅ 已集成 |
| **默认 Router** | `web/mvc/starter.go` | ✅ 已集成 |

### 自动挂载流程

```
框架 AutoConfig.Configure()
  └─ 创建 XxxEndpointRegistry(engine)
  └─ 注册到容器: HttpEndpointRegistry 类型
  └─ ActuatorHttpStarter.Start()
     └─ 查找 HttpEndpointRegistry
     └─ 自动挂载所有端点
```

### 集成新框架

如果您需要将 Actuator 集成到其他 HTTP 框架，请参考以下步骤：

#### 1. 创建 EndpointRegistry 实现

```go
package yourframework

import (
    "net/http"
    "github.com/xudefa/enhance/actuator"
)

type YourFrameworkEndpointRegistry struct {
    engine    *YourEngine
    endpoints map[string]bool
}

func NewYourFrameworkEndpointRegistry(engine *YourEngine) *YourFrameworkEndpointRegistry {
    return &YourFrameworkEndpointRegistry{
        engine:    engine,
        endpoints: make(map[string]bool),
    }
}

func (r *YourFrameworkEndpointRegistry) RegisterEndpoint(method, path string, handler http.Handler) {
    if r.engine == nil || handler == nil {
        return
    }
    
    // 适配框架特定的路由注册
    r.engine.Add(method, path, func(c *Context) {
        handler.ServeHTTP(c.Response(), c.Request())
    })
    r.endpoints[path] = true
}

func (r *YourFrameworkEndpointRegistry) RegisterEndpoints(endpoints []actuator.EndpointConfig) {
    for _, ep := range endpoints {
        r.RegisterEndpoint(ep.Method, ep.Path, ep.Handler)
    }
}

func (r *YourFrameworkEndpointRegistry) HasEndpoint(path string) bool {
    _, exists := r.endpoints[path]
    return exists
}
```

#### 2. 在 AutoConfig 中注册

```go
func (c *YourFrameworkAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
    // ... 创建框架 engine ...
    
    // 注册 HttpEndpointRegistry
    endpointRegistry := NewYourFrameworkEndpointRegistry(c.engine)
    if err := ctx.Container().RegisterInstance(
        endpointRegistry, 
        reflect.TypeFor[actuator.HttpEndpointRegistry](),
    ); err != nil {
        c.logger.Warn(context.Background(), "注册 HttpEndpointRegistry 失败,Actuator 端点将无法自动挂载",
            log.KeyValue{Key: "error", Value: err.Error()},
        )
    }
    
    return nil
}
```

#### 3. 验证集成

```bash
curl http://localhost:8080/actuator/health
curl http://localhost:8080/actuator/metrics
curl http://localhost:8080/actuator/env
```

---

## 健康检查系统

### 内置健康指示器

| 指示器 | 说明 |
|--------|------|
| `FuncHealthIndicator` | 基于函数的通用健康指标 |
| `DatabaseHealthIndicator` | 数据库健康指标 |
| `RedisHealthIndicator` | Redis 健康指标 |

### 使用示例

#### 函数健康指标

```go
indicator := actuator.NewFuncHealthIndicator(
    "my-service",
    func(ctx context.Context) error {
        resp, err := http.Get("http://localhost:8080/health")
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        return nil
    },
)
```

#### 数据库健康指标

```go
indicator := actuator.NewDatabaseHealthIndicator(
    func(ctx context.Context) error {
        return db.PingContext(ctx)
    },
)
```

#### Redis 健康指标

```go
indicator := actuator.NewRedisHealthIndicator(
    func(ctx context.Context) error {
        return redisClient.Ping(ctx).Err()
    },
)
```

### Builder 模式

使用 Builder 模式简化健康指示器的创建：

```go
indicator := actuator.NewHealthIndicatorBuilder().
    Name("database").
    CheckFunc(db.Check).
    Timeout(5 * time.Second).
    Detail("type", "postgres").
    Build()
```

| 方法 | 说明 |
|------|------|
| `Name(name string)` | 设置指标名称 |
| `CheckFunc(fn)` | 设置检查函数 |
| `Timeout(d)` | 设置超时时间（默认 5s） |
| `Detail(key, value)` | 添加详细信息 |
| `Build()` | 构建健康指示器 |

### 自定义健康指示器

```go
type customIndicator struct{}

func (c *customIndicator) Name() string { return "custom" }

func (c *customIndicator) Health(ctx context.Context) health.Health {
    // 自定义健康检查逻辑
    return health.Health{
        Status: health.StatusUp,
        Details: map[string]any{
            "version": "1.0.0",
        },
    }
}

// 注册到容器
container.Register(
    reflect.TypeOf(&customIndicator{}),
    core.Bean(&customIndicator{}),
)
```

### 检查逻辑

所有内置健康指标统一使用 `checkHealth` 函数：

1. `checkFn` 为 `nil` → 返回 `StatusUnknown`
2. `checkFn` 返回错误 → 返回 `StatusDown`，详情中附带错误信息
3. 检查通过 → 返回 `StatusUp`

---

## 敏感信息保护

### 默认检测规则

**关键词检测**：password, secret, token, key, auth, credential, private, api_key, access_token, client_secret, oauth, bearer, jwt

**值格式检测**：私钥格式、JWT 令牌、长随机字符串

### 自定义检测策略

```go
type SanitizeStrategy interface {
    IsSensitive(key string, value any) bool
}

sanitizer := actuator.NewSanitizer()
sanitizer.AddStrategy(&myCustomStrategy{})

// 掩盖敏感值
value := sanitizer.Sanitize("db.password", "secret123")
// 返回: "***REDACTED***"
```

#### 示例：业务特定敏感信息检测

```go
type MyCustomStrategy struct{}

func (s *MyCustomStrategy) IsSensitive(key string, value any) bool {
    return strings.Contains(key, "api-key")
}

sanitizer := actuator.NewSanitizer()
sanitizer.AddStrategy(&MyCustomStrategy{})
```

---

## 配置选项

### 基础配置

```yaml
actuator:
  enabled: true              # 启用/禁用 Actuator（默认 true）
  path: /actuator            # 自定义端点路径（默认 /actuator）
```

### 端点暴露控制

```yaml
actuator:
  expose:
    health: true             # 健康检查端点（默认 true）
    metrics: true            # 指标端点（默认 true）
    env: true                # 环境信息端点（默认 true）
    beans: true              # Bean 列表端点（默认 true）
    info: true               # 应用信息端点（默认 true）
    prometheus: true         # Prometheus 端点（默认 true）
```

### 独立服务器配置（降级方案）

当无法找到 HTTP 服务器时，Actuator 会启动独立服务器：

```yaml
actuator:
  host: 0.0.0.0              # 默认 0.0.0.0
  port: 8081                 # 默认 8081
```

### 路由配置

```go
type RouteConfig struct {
    BasePath    string  // 基础路径，默认 "/actuator"
    ExposeDebug bool    // 是否暴露调试端点（pprof）
    Prefix      string  // 路径前缀
}
```

---

## 完整示例

```go
package main

import (
    "context"
    "net/http"
    "reflect"

    "github.com/xudefa/enhance/actuator"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("my-app"),
    )
    
    // 注册自定义健康指标
    dbIndicator := actuator.NewDatabaseHealthIndicator(
        func(ctx context.Context) error {
            return db.PingContext(ctx)
        },
    )
    app.Container().Register(
        reflect.TypeOf(dbIndicator),
        core.Bean(dbIndicator),
    )
    
    redisIndicator := actuator.NewRedisHealthIndicator(
        func(ctx context.Context) error {
            return redisClient.Ping(ctx).Err()
        },
    )
    app.Container().Register(
        reflect.TypeOf(redisIndicator),
        core.Bean(redisIndicator),
    )
    
    app.Start()
    defer app.Stop()
    
    // Actuator 端点已自动挂载
    // 访问 http://localhost:8080/actuator/health
}
```

---

## 最佳实践

### ✅ 1. 生产环境启用 Actuator

```yaml
actuator:
  enabled: true
```

### ✅ 2. 保护 Actuator 端点

```go
security.AuthorizeRequests(func(reg security.AuthorizeRequests) {
    reg.AntMatchers("/actuator/**").HasRole("ADMIN")
    reg.AnyRequest().Authenticated()
})
```

### ✅ 3. 监控关键依赖

```go
// 为所有关键依赖添加健康检查
indicator := actuator.NewDatabaseHealthIndicator(
    func(ctx context.Context) error {
        return db.PingContext(ctx)
    },
)
container.Register(
    reflect.TypeOf(indicator),
    core.Bean(indicator),
)
```

### ✅ 4. 生产环境禁用调试端点

```go
config := actuator.DefaultRouteConfig()
config.ExposeDebug = false  // 生产环境禁用 pprof
```

### ✅ 5. 始终注册 HttpEndpointRegistry

确保 Actuator 端点能自动挂载到 HTTP 框架：

```go
endpointRegistry := NewYourFrameworkEndpointRegistry(c.engine)
ctx.Container().RegisterInstance(
    endpointRegistry, 
    reflect.TypeFor[actuator.HttpEndpointRegistry](),
)
```

---

## 故障排查

### Actuator 端点未挂载

1. ✅ 检查 `actuator.enabled` 是否为 `true`
2. ✅ 检查框架是否注册了 `HttpEndpointRegistry`
3. ✅ 查看日志中是否有 "Started standalone server" 消息

### 使用独立服务器

如果日志显示 `"Started standalone server on :8081"`，说明 Actuator 未能找到 HTTP 服务器，使用了降级方案。

**常见原因**：
- 框架的 autoconfig 未注册 `HttpEndpointRegistry`
- 框架的 autoconfig 执行顺序在 actuator-http 之后

**解决方案**：
```go
// 在框架的 autoconfig 中手动注册
registry := NewYourFrameworkEndpointRegistry(engine)
ctx.Container().RegisterInstance(registry, reflect.TypeFor[actuator.HttpEndpointRegistry]())
```

### 敏感信息泄露

检查 `/actuator/env` 端点返回的数据，确保敏感信息已被掩盖：

```bash
curl http://localhost:8080/actuator/env | grep -i password
# 应该返回 "***REDACTED***" 而非真实密码
```

如需添加自定义敏感信息检测策略，参考 [敏感信息保护](#敏感信息保护) 章节。