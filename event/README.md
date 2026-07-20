# event 包 — 事件驱动系统

> **所属层级**: Core Layer  
> **设计理念**: 观察者模式，解耦业务逻辑  
> **设计灵感**: Spring ApplicationEvent / ApplicationListener

## 概述

`event` 包提供应用事件驱动支持，实现观察者模式，解耦业务逻辑。支持同步/异步事件发布、事务事件、死信队列等高级特性。

### 核心功能

| 功能 | 说明 |
|------|------|
| **ApplicationEvent** | 应用事件接口，所有自定义事件需实现此接口 |
| **EventBus** | 事件总线，支持事件的发布和订阅 |
| **BaseEvent** | 基础事件实现，可直接使用或嵌入自定义事件结构体 |
| **AsyncPublisher** | 异步事件发布器，支持上下文超时和错误处理 |
| **TransactionalEvent** | 事务绑定事件，支持 BeforeCommit/AfterCommit/AfterRollback |
| **DeadLetterQueue** | 死信队列，支持重试机制和退避策略 |
| **EventBusWithOrdering** | 保证事件发布顺序的事件总线 |

---

## 核心接口

### ApplicationEvent 接口

```go
type ApplicationEvent interface {
    Type() string
    Timestamp() time.Time
}
```

### EventListener

```go
type EventListener func(event ApplicationEvent)
```

### EventBus

```go
type EventBus struct {
    mu        sync.RWMutex
    listeners map[string][]EventListener
}
```

### BaseEvent

```go
type BaseEvent struct {
    EventType string
    EventTime time.Time
}

func (e *BaseEvent) Type() string
func (e *BaseEvent) Timestamp() time.Time
```

### 内置事件类型

| 常量 | 值 | 说明 |
|------|----|------|
| `EventEnvironmentPrepared` | `"EnvironmentPrepared"` | 环境准备完成 |
| `EventContextRefreshed` | `"ContextRefreshed"` | 上下文刷新完成 |
| `EventApplicationStarted` | `"ApplicationStarted"` | 应用已启动 |
| `EventApplicationReady` | `"ApplicationReady"` | 应用已就绪 |
| `EventApplicationStopped` | `"ApplicationStopped"` | 应用已停止 |

---

## 快速开始

### 创建事件总线

```go
bus := event.NewEventBus()
```

### 订阅事件

```go
bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("应用已启动，时间:", e.Timestamp())
})
```

### 发布事件

```go
bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
```

### 取消订阅

```go
handler := func(e event.ApplicationEvent) {
    fmt.Println("收到事件:", e.Type())
}

bus.Subscribe(event.EventApplicationReady, handler)
bus.Unsubscribe(event.EventApplicationReady, handler)
```

---

## API 参考

### EventBus 方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `Subscribe(eventType, listener)` | 订阅指定类型的事件 | `bus.Subscribe("UserLogin", handler)` |
| `Unsubscribe(eventType, listener)` | 取消订阅 | `bus.Unsubscribe("UserLogin", handler)` |
| `Publish(event)` | 发布事件 | `bus.Publish(&event.BaseEvent{...})` |

### BaseEvent 使用

```go
// 创建基础事件
evt := &event.BaseEvent{EventType: "CustomEvent"}

// 嵌入到自定义事件中
type MyCustomEvent struct {
    event.BaseEvent
    UserID   string
    Action   string
}
```

---

## 使用示例

### 基本事件发布与订阅

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/event"
)

func main() {
    bus := event.NewEventBus()

    // 订阅多个事件
    bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
        fmt.Println("应用启动中...")
    })

    bus.Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
        fmt.Println("应用已就绪，可以提供服务")
    })

    bus.Subscribe(event.EventApplicationStopped, func(e event.ApplicationEvent) {
        fmt.Println("应用已停止")
    })

    // 发布事件
    bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
    bus.Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
    bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
}
```

### 自定义事件类型

```go
type UserRegisteredEvent struct {
    event.BaseEvent
    Username  string
    Email     string
}

func main() {
    bus := event.NewEventBus()

    bus.Subscribe("UserRegistered", func(e event.ApplicationEvent) {
        evt := e.(*UserRegisteredEvent)
        fmt.Printf("用户注册: %s (%s)\n", evt.Username, evt.Email)
    })

    bus.Publish(&UserRegisteredEvent{
        BaseEvent: event.BaseEvent{EventType: "UserRegistered"},
        Username:  "john",
        Email:     "john@example.com",
    })
}
```

### 多监听器

```go
bus := event.NewEventBus()

// 多个监听器订阅同一事件
bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("监听器 1: 记录启动日志")
})

bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("监听器 2: 发送启动通知")
})

bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("监听器 3: 初始化监控指标")
})

// 发布后三个监听器按订阅顺序依次调用
bus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
```

### 在 Boot 启动中的事件流

`Boot.Start()` 按顺序发布事件：

```go
// PhaseConfiguring 阶段
eventBus.Publish(&event.BaseEvent{EventType: event.EventEnvironmentPrepared})

// PhaseContextRefreshed 阶段
eventBus.Publish(&event.BaseEvent{EventType: event.EventContextRefreshed})

// PhaseRunning 阶段
eventBus.Publish(&event.BaseEvent{EventType: event.EventApplicationStarted})
eventBus.Publish(&event.BaseEvent{EventType: event.EventApplicationReady})
```

`Boot.Stop()` 发布：

```go
eventBus.Publish(&event.BaseEvent{EventType: event.EventApplicationStopped})
```

完整的事件时间线：

```
PhaseConfiguring    → EventEnvironmentPrepared
PhaseContextRefreshed → EventContextRefreshed
PhaseRunning        → EventApplicationStarted
PhaseRunning        → EventApplicationReady
PhaseStopped        → EventApplicationStopped
```

---

## 最佳实践

### 1. 使用自定义事件传递业务数据

```go
// ✅ 推荐：使用自定义事件
type OrderCreatedEvent struct {
    event.BaseEvent
    OrderID string
    Amount  float64
}

// ⚠️ 不推荐：使用 BaseEvent 传递复杂数据
bus.Publish(&event.BaseEvent{EventType: "OrderCreated"})
```

### 2. 及时取消订阅

避免内存泄漏，在对象销毁时取消订阅：

```go
type MyComponent struct {
    bus     *event.EventBus
    handler event.EventListener
}

func (c *MyComponent) Start() {
    c.handler = func(e event.ApplicationEvent) {
        // 处理事件
    }
    c.bus.Subscribe("MyEvent", c.handler)
}

func (c *MyComponent) Stop() {
    c.bus.Unsubscribe("MyEvent", c.handler)
}
```

### 3. 避免在事件处理中抛出异常

事件处理器应该捕获并处理异常，避免影响其他监听器：

```go
bus.Subscribe("UserLogin", func(e event.ApplicationEvent) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("事件处理异常: %v", r)
        }
    }()
    // 处理逻辑
})
```

### 4. 使用异步事件发布处理耗时操作

对于耗时的事件处理，使用异步发布避免阻塞主流程：

```go
// 异步发布事件
asyncPublisher := event.NewAsyncPublisher(bus)
asyncPublisher.Publish(context.Background(), &event.BaseEvent{
    EventType: "EmailNotification",
})
```

    bus.Publish(&UserRegisteredEvent{
        BaseEvent: event.BaseEvent{EventType: "UserRegistered"},
        Username:  "alice",
        Email:     "alice@example.com",
    })
}
```

### 使用事件进行模块解耦

```go
// order 模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("订单 %s 支付完成，更新订单状态\n", evt.OrderID)
})

// notification 模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("发送支付成功通知给用户 %s\n", evt.UserID)
})

// analytics 模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("统计: 订单金额 %.2f\n", evt.Amount)
})
```

## 与 context 包的关系

`DefaultApplicationContext` 内部持有 `EventBus`：

```go
ctx := context.NewApplicationContext(container, env)

ctx.EventBus().Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
    fmt.Println("上下文已就绪")
})

ctx.Start() // 触发 EventApplicationStarted 和 EventApplicationReady
ctx.Stop()  // 触发 EventApplicationStopped
```

---

## 使用场景

### 场景 1：应用生命周期管理

**描述**：监听应用启动、就绪、停止等生命周期事件，执行相应的初始化和清理操作。

```go
bus := event.NewEventBus()

bus.Subscribe(event.EventApplicationStarted, func(e event.ApplicationEvent) {
    fmt.Println("应用启动中，初始化资源...")
})

bus.Subscribe(event.EventApplicationReady, func(e event.ApplicationEvent) {
    fmt.Println("应用已就绪，开始处理请求...")
})

bus.Subscribe(event.EventApplicationStopped, func(e event.ApplicationEvent) {
    fmt.Println("应用停止中，清理资源...")
})
```

**最佳实践**：
- 在应用启动时订阅生命周期事件
- 使用事件进行资源初始化和清理
- 避免在监听器中执行耗时操作

### 场景 2：模块间解耦

**描述**：使用事件实现模块间松耦合通信，避免直接依赖。

```go
// 订单模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("订单 %s 支付完成，更新订单状态\n", evt.OrderID)
})

// 通知模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("发送支付成功通知给用户 %s\n", evt.UserID)
})

// 分析模块
bus.Subscribe("PaymentCompleted", func(e event.ApplicationEvent) {
    evt := e.(*PaymentCompletedEvent)
    fmt.Printf("统计: 订单金额 %.2f\n", evt.Amount)
})
```

**最佳实践**：
- 使用事件实现发布-订阅模式
- 每个模块独立订阅感兴趣的事件
- 避免在事件处理中引入循环依赖

### 场景 3：审计日志

**描述**：记录关键业务操作的事件日志，用于审计和追溯。

```go
bus.Subscribe("UserCreated", func(e event.ApplicationEvent) {
    evt := e.(*UserCreatedEvent)
    log.Printf("用户创建: ID=%s, Username=%s, Time=%s",
        evt.UserID, evt.Username, evt.Timestamp())
})

bus.Subscribe("OrderPlaced", func(e event.ApplicationEvent) {
    evt := e.(*OrderPlacedEvent)
    log.Printf("订单创建: OrderID=%s, UserID=%s, Amount=%.2f",
        evt.OrderID, evt.UserID, evt.Amount)
})
```

**最佳实践**：
- 使用事件记录关键业务操作
- 包含足够的上下文信息
- 异步处理审计日志，避免影响业务性能

### 场景 4：缓存失效

**描述**：当数据变更时，通过事件通知相关模块清理缓存。

```go
bus.Subscribe("DataUpdated", func(e event.ApplicationEvent) {
    evt := e.(*DataUpdatedEvent)
    cacheKey := fmt.Sprintf("%s:%s", evt.DataType, evt.DataID)
    cache.Delete(cacheKey)
    fmt.Printf("缓存已清理: %s\n", cacheKey)
})
```

**最佳实践**：
- 使用事件触发缓存清理
- 确保缓存清理的幂等性
- 考虑使用事件版本号避免重复处理