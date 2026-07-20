# lifecycle 包 — 生命周期管理

> **所属层级**: Core Layer  
> **设计理念**: 阶段管理，生命周期回调  
> **设计灵感**: Spring Lifecycle + Uber fx

## 概述

`lifecycle` 包提供完整的生命周期管理功能，包括应用生命周期阶段管理和 Bean 生命周期回调。

### 核心功能

| 功能 | 说明 |
|------|------|
| **应用生命周期阶段管理** | PhaseInitializing → PhaseConfiguring → PhaseContextRefreshed → PhaseReady → PhaseRunning → PhaseStopping → PhaseStopped |
| **Bean 生命周期回调** | BeanInitFunc（初始化）和 BeanDestroyFunc（销毁） |
| **生命周期钩子** | Hook 接口（OnInit/OnStart/OnStop） |
| **生命周期构建器** | 支持链式配置生命周期管理器 |

---

## 核心接口

### ApplicationPhase 应用生命周期阶段

```go
type ApplicationPhase int

const (
    PhaseInitializing     ApplicationPhase = iota // 初始化阶段：创建容器和基础组件
    PhaseConfiguring                              // 配置阶段：加载配置、注册 Bean
    PhaseContextRefreshed                         // 上下文刷新完成：所有 Bean 已注册
    PhaseReady                                    // 就绪阶段：应用准备就绪但尚未开始服务
    PhaseRunning                                  // 运行阶段：应用正常运行，处理请求
    PhaseStopping                                 // 停止阶段：应用正在停止，释放资源
    PhaseStopped                                  // 已停止：应用完全停止
)
```

#### 阶段转换规则

- 只允许正向转换(从早期阶段到后期阶段)
- 不允许回退或跳跃转换
- 转换时会通知所有注册的监听器

### PhaseListener 阶段监听器

```go
type PhaseListener interface {
    OnPhaseChange(oldPhase, newPhase ApplicationPhase) error
}
```

实现此接口可以在阶段变更时执行自定义逻辑。

### LifecycleManager 生命周期管理器

```go
type LifecycleManager struct {
    // ...
}
```

#### 创建

```go
manager := lifecycle.NewLifecycleManager()
```

#### 阶段操作

```go
// 获取当前阶段
phase := manager.GetPhase()

// 设置新阶段(会通知所有监听器)
err := manager.SetPhase(lifecycle.PhaseRunning)

// 添加阶段监听器
manager.AddListener(&myListener{})

// 设置错误处理回调
manager.SetErrorHandler(func(old, new lifecycle.ApplicationPhase, err error) {
    log.Printf("Phase transition error: %v", err)
})
```

### Bean 生命周期回调

#### BeanInitFunc 初始化钩子

```go
type BeanInitFunc func(bean any) error
```

在 Bean 创建后调用，替代 Spring 风格的 InitializingBean 接口：

```go
initHook := lifecycle.BeanInitFunc(func(bean any) error {
    if db, ok := bean.(*Database); ok {
        return db.Connect()
    }
    return nil
})
```

#### BeanDestroyFunc 销毁钩子

```go
type BeanDestroyFunc func(bean any) error
```

在 Bean 销毁前调用，替代 Spring 风格的 DisposableBean 接口：

```go
destroyHook := lifecycle.BeanDestroyFunc(func(bean any) error {
    if db, ok := bean.(*Database); ok {
        return db.Close()
    }
    return nil
})
```

### Hook 生命周期钩子接口

```go
type Hook interface {
    OnInit(ctx context.Context) error
    OnStart(ctx context.Context) error
    OnStop(ctx context.Context) error
}
```

#### 函数式钩子

```go
// 创建完整钩子
hook := lifecycle.NewHookFunc(onInit, onStart, onStop)

// 创建单一钩子
hook := lifecycle.OnInitFunc(func(ctx context.Context) error {
    // 初始化逻辑
    return nil
})
```

### HookRegistry 钩子注册表

```go
registry := lifecycle.NewHookRegistry()
registry.Register(hook)

// 按注册顺序执行 OnInit
registry.InitAll(ctx)

// 按注册顺序执行 OnStart
registry.StartAll(ctx)

// 按注册逆序执行 OnStop
registry.StopAll(ctx)
```

### LifecycleBuilder 生命周期构建器

```go
manager := lifecycle.NewLifecycleBuilder().
    InitialPhase(lifecycle.PhaseConfiguring).
    Listener(&myListener{}).
    Build()
```

---

## 快速开始

### 基本生命周期管理

```go
package main

import (
    "context"
    "fmt"
    "github.com/xudefa/enhance/lifecycle"
)

func main() {
    manager := lifecycle.NewLifecycleManager()

    // 添加阶段监听器
    manager.AddListener(&lifecycle.PhaseListenerFunc(func(old, new lifecycle.ApplicationPhase) error {
        fmt.Printf("Phase changed from %v to %v\n", old, new)
        return nil
    }))

    // 设置阶段
    manager.SetPhase(lifecycle.PhaseRunning)
    fmt.Println("Current phase:", manager.GetPhase())
}
```

---

## API 参考

### Bean 生命周期回调

```go
// 注册初始化钩子
container.RegisterBean(
    reflect.TypeOf(&Database{}),
    core.Bean(func() (*Database, error) {
        return &Database{}, nil
    }),
    core.InitFunc(func(bean any) error {
        db := bean.(*Database)
        return db.Connect()
    }),
    core.DestroyFunc(func(bean any) error {
        db := bean.(*Database)
        return db.Close()
    }),
)
```

### 生命周期钩子

```go
// 创建钩子
hook := lifecycle.NewHookFunc(
    func(ctx context.Context) error {
        fmt.Println("Initializing...")
        return nil
    },
    func(ctx context.Context) error {
        fmt.Println("Starting...")
        return nil
    },
    func(ctx context.Context) error {
        fmt.Println("Stopping...")
        return nil
    },
)

// 注册钩子
registry := lifecycle.NewHookRegistry()
registry.Register(hook)

// 执行生命周期
registry.InitAll(ctx)
registry.StartAll(ctx)
registry.StopAll(ctx)
```

---

## 使用示例

### 应用启动和关闭

```go
func main() {
    manager := lifecycle.NewLifecycleManager()

    // 注册钩子
    registry := lifecycle.NewHookRegistry()
    registry.Register(lifecycle.NewHookFunc(
        func(ctx context.Context) error {
            // 初始化数据库连接
            db.Connect()
            return nil
        },
        func(ctx context.Context) error {
            // 启动 HTTP 服务器
            server.Start()
            return nil
        },
        func(ctx context.Context) error {
            // 关闭数据库连接
            db.Close()
            return nil
        },
    ))

    // 设置阶段
    manager.SetPhase(lifecycle.PhaseInitializing)
    registry.InitAll(ctx)

    manager.SetPhase(lifecycle.PhaseRunning)
    registry.StartAll(ctx)

    // 等待信号
    <-signalChan

    // 关闭应用
    manager.SetPhase(lifecycle.PhaseStopping)
    registry.StopAll(ctx)
    manager.SetPhase(lifecycle.PhaseStopped)
}
```

### 自定义阶段监听器

```go
type LoggingPhaseListener struct{}

func (l *LoggingPhaseListener) OnPhaseChange(old, new lifecycle.ApplicationPhase) error {
    log.Printf("Application phase changed: %v -> %v", old, new)
    return nil
}

manager.AddListener(&LoggingPhaseListener{})
```

---

## 与 fx/go-spring/spring-boot 对比

| 特性 | fx | go-spring | spring-boot | enhance (Go 风格) |
|------|-----|-----------|-------------|-------------------|
| 生命周期钩子 | Invoke/Close | 7 阶段状态机 | BeanPostProcessor | Hook(OnInit/OnStart/OnStop) |
| Bean 初始化 | 构造函数 | InitializingBean | @PostConstruct | BeanInitFunc |
| Bean 销毁 | Close() | DisposableBean | @PreDestroy | BeanDestroyFunc |
| 类型安全 | 泛型 | 反射 | 反射 | 泛型 + 反射 |
| 代码风格 | Go 风格 | Java 风格 | Java 风格 | Go 风格 |

---

## 最佳实践

### 1. 使用生命周期管理器管理应用状态

```go
// ✅ 推荐：使用 LifecycleManager 管理阶段
manager := lifecycle.NewLifecycleManager()
manager.SetPhase(lifecycle.PhaseInitializing)
// 初始化逻辑
manager.SetPhase(lifecycle.PhaseRunning)

// ⚠️ 不推荐：手动管理状态
var phase string
phase = "initializing"
// 初始化逻辑
phase = "running"
```

### 2. 使用钩子注册表管理生命周期

```go
// ✅ 推荐：使用 HookRegistry 统一管理
registry := lifecycle.NewHookRegistry()
registry.Register(dbHook)
registry.Register(serverHook)
registry.InitAll(ctx)
registry.StartAll(ctx)
defer registry.StopAll(ctx)

// ⚠️ 不推荐：分散的生命周期逻辑
db.Connect()
server.Start()
defer db.Close()
defer server.Stop()
```

### 3. 使用 Builder 模式配置生命周期

```go
// ✅ 推荐：使用 Builder 链式配置
manager := lifecycle.NewLifecycleBuilder().
    InitialPhase(lifecycle.PhaseConfiguring).
    Listener(&myListener{}).
    ErrorHandler(func(old, new lifecycle.ApplicationPhase, err error) {
        log.Printf("Error: %v", err)
    }).
    Build()

// ⚠️ 不推荐：手动配置每个选项
manager := lifecycle.NewLifecycleManager()
manager.SetPhase(lifecycle.PhaseConfiguring)
manager.AddListener(&myListener{})
manager.SetErrorHandler(...)
```

### 4. 使用 Bean 生命周期回调

```go
// ✅ 推荐：使用 InitFunc 和 DestroyFunc
container.RegisterBean(
    reflect.TypeOf(&Database{}),
    core.Bean(newDatabase),
    core.InitFunc(func(bean any) error {
        return bean.(*Database).Connect()
    }),
    core.DestroyFunc(func(bean any) error {
        return bean.(*Database).Close()
    }),
)

// ⚠️ 不推荐：手动管理 Bean 生命周期
db := newDatabase()
db.Connect()
defer db.Close()
```

### 5. 实现自定义阶段监听器

```go
// ✅ 推荐：实现 PhaseListener 接口
type MetricsPhaseListener struct {
    metrics metrics.MeterRegistry
}

func (l *MetricsPhaseListener) OnPhaseChange(old, new lifecycle.ApplicationPhase) error {
    l.metrics.Counter("lifecycle.phase.changes", "from", old.String(), "to", new.String()).Inc()
    return nil
}

// ⚠️ 不推荐：硬编码阶段变更逻辑
func setPhase(phase lifecycle.ApplicationPhase) {
    currentPhase = phase
    metrics.Increment("phase_changes")
}
```