# context 包 — 应用上下文

> **所属层级**: Core Layer  
> **设计理念**: 运行时核心入口，聚合所有子系统

## 概述

`context` 包提供应用上下文（ApplicationContext），聚合了 enhance 框架的四个核心子系统：Container（IoC）、Environment（配置）、Lifecycle（生命周期）、EventBus（事件）。

### 核心功能

| 功能 | 说明 |
|------|------|
| **Container** | IoC 依赖注入容器，管理 Bean 的注册和获取 |
| **Environment** | 分层配置源管理，支持多级配置源 |
| **Lifecycle** | 应用生命周期阶段管理 |
| **EventBus** | 事件发布与订阅 |
| **EventPublisher** | 事件发布器接口，解耦事件发布逻辑 |

---

## 核心接口

### ApplicationContext 接口

```go
type ApplicationContext interface {
    Container() core.Container
    Environment() *environment.Environment
    Lifecycle() *life.LifecycleManager
    EventBus() *event.EventBus
    EventPublisher() EventPublisher

    Register(name string, opts ...core.BuilderOption) error
    Get(name string) (any, error)
    Invoke(fn any) error

    Start() error
    Stop() error
    IsRunning() bool
}

// EventPublisher 事件发布器接口
type EventPublisher interface {
    Publish(event event.ApplicationEvent)
}
```

### DefaultApplicationContext

```go
type DefaultApplicationContext struct {
    container core.Container
    env       *environment.Environment
    lifecycle *life.LifecycleManager
    events    *event.EventBus
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Container()` | 返回 IoC 容器 |
| `Environment()` | 返回环境配置 |
| `Lifecycle()` | 返回生命周期管理器 |
| `EventBus()` | 返回事件总线 |
| `EventPublisher()` | 返回事件发布器接口 |
| `Register(name, opts...)` | 在容器中注册 Bean |
| `Get(name)` | 从容器中获取指定名称的 Bean |
| `Invoke(fn)` | 调用函数并自动注入依赖参数 |
| `Start()` | 启动应用，发布启动事件并切换至运行阶段 |
| `Stop()` | 停止应用，切换至停止阶段并发布停止事件 |
| `IsRunning()` | 检查应用是否处于运行状态 |

---

## 快速开始

### 创建应用上下文

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/context"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/config/environment"
)

func main() {
    container := core.New()
    env := environment.NewEnvironment()
    ctx := context.NewApplicationContext(container, env)

    // 注册 Bean
    _ = ctx.Register(
        reflect.TypeOf(&MyService{}),
        core.Bean(&MyService{Name: "hello"}),
        core.Singleton(),
    )

    // 获取 Bean
    svc := core.MustGetBean[*MyService](ctx.Container())
    fmt.Println(svc.Name)

    // 启动
    ctx.Start()
    defer ctx.Stop()
}

type MyService struct {
    Name string
}
```

### 启动与停止流程

#### Start() 流程

1. 发布 `EventApplicationStarted`
2. 设置生命周期阶段为 `PhaseRunning`
3. 发布 `EventApplicationReady`

#### Stop() 流程

1. 设置生命周期阶段为 `PhaseStopping`
2. 设置生命周期阶段为 `PhaseStopped`
3. 发布 `EventApplicationStopped`

---

## API 参考

### 辅助方法

```go
func (c *DefaultApplicationContext) GetBean(beanID string) (any, bool)
func (c *DefaultApplicationContext) HasProperty(key string) bool
func (c *DefaultApplicationContext) GetProperty(key string) (any, bool)
func (c *DefaultApplicationContext) ClassLoader() interface{ HasClass(name string) bool }
```

### ClassLoader 缓存优化

`buildInfoClassLoader` 使用 `runtime/debug.ReadBuildInfo()` 检查模块是否在编译依赖中。

#### 优化策略

- **sync.Once 延迟初始化**：首次调用 `HasClass` 时读取构建信息
- **依赖列表缓存**：将构建信息中的依赖模块列表缓存到内存
- **全局共享实例**：使用 `globalClassLoader` 全局共享实例

#### 性能提升

- **首次调用**：读取构建信息（约 1-5ms）
- **后续调用**：直接查询缓存（约 10-100ns）
- **内存开销**：每个依赖模块约 50 字节字符串，100 个依赖约 5KB

---

## 使用示例

### 事件订阅

```go
ctx.EventBus().Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("应用已启动，时间:", e.Timestamp())
})

ctx.Start()
```

### 生命周期监听

```go
ctx.Lifecycle().AddListener(&myPhaseListener{})

ctx.Start() // 触发 PhaseRunning 变更通知
```

### 方法注入

```go
_ = ctx.Invoke(func(svc *MyService) {
    fmt.Println("注入的 service:", svc.Name)
})
```

---

## 四子系统集成关系

```
┌──────────────────────────────────────────────────────┐
│              DefaultApplicationContext               │
│                                                      │
│  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │   Container  │  │        Environment           │  │
│  │  (core.Cont) │  │  (environment.Environment)   │  │
│  │              │  │                              │  │
│  │  Register()  │  │  GetProperty()               │  │
│  │  Get()       │  │  AddPropertySource()         │  │
│  │  Invoke()    │  │  GetActiveProfiles()         │  │
│  └──────────────┘  └──────────────────────────────┘  │
│                                                      │
│  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │  Lifecycle   │  │          EventBus            │  │
│  │  (life.Man)  │  │      (event.EventBus)        │  │
│  │              │  │                              │  │
│  │  SetPhase()  │  │  Publish()                   │  │
│  │  GetPhase()  │  │  Subscribe()                 │  │
│  │  AddListen() │  │  Unsubscribe()               │  │
│  └──────────────┘  └──────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

---

## 与 Boot 的关系

`boot.Boot` 在内部持有 `DefaultApplicationContext` 并进行更细粒度的生命周期控制：

- `Boot.Start()` 在 `PhaseConfiguring` → `PhaseReady` 之间插入自动配置执行、启动器配置等步骤
- `DefaultApplicationContext.Start()` 仅处理 `PhaseRunning` 阶段切换和事件发布

`DefaultApplicationContext` 同时实现了 `condition.ConditionContext` 所需的辅助方法（`GetBean`、`HasProperty`、`GetProperty`），通过 `conditionCtx` 适配器供条件系统使用。

---

## 最佳实践

### 1. 使用 ApplicationContext 作为统一入口

```go
// ✅ 推荐：通过上下文访问所有子系统
ctx := context.NewApplicationContext(container, env)
ctx.Container().Get("myService")
ctx.Environment().GetProperty("app.name")
ctx.EventBus().Publish(event)

// ⚠️ 不推荐：直接访问各个子系统
container.Get("myService")
env.GetProperty("app.name")
```

### 2. 合理使用事件订阅

```go
// ✅ 推荐：在启动前订阅事件
ctx.EventBus().Subscribe(event.EventApplicationStarted, handler)
ctx.Start()

// ⚠️ 不推荐：在启动后订阅，可能错过事件
ctx.Start()
ctx.EventBus().Subscribe(event.EventApplicationStarted, handler)
```

### 3. 使用 ClassLoader 缓存提升性能

```go
// ✅ 推荐：使用全局共享的 ClassLoader
classLoader := context.GlobalClassLoader()
classLoader.HasClass("github.com/some/lib")

// ⚠️ 不推荐：每次创建新的 ClassLoader
classLoader := context.NewBuildInfoClassLoader()
```