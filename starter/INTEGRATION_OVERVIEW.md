# enhance 框架集成总览

> **文档版本**：v0.0 | **最后更新**：2026-07-29

本文档提供 enhance 框架所有第三方集成的完整清单和使用指南。

---

## 集成清单

### P0 - 核心集成（内置支持）

| 框架 | 模块 | 路径 | 功能 | 状态 |
|------|------|------|------|------|
| Gin | web | 内置 | HTTP Web 框架 | ✅ |
| GORM | data | starter/gorm | ORM 框架 | ✅ |
| XORM | data | starter/xorm | ORM 框架 | ✅ |
| slog | log | 内置 | 日志框架（标准库） | ✅ |

### P1 - 重要集成

| 框架 | 模块 | 路径 | 功能 | 状态 |
|------|------|------|------|------|
| Redis | starter/redis | `starter/redis/` | 分布式缓存 | ✅ |
| Kafka | starter/kafka | `starter/kafka/` | 消息队列 | ✅ |
| OpenTelemetry | starter/otel | `starter/otel/` | 链路追踪 | ✅ |
| Zap | starter/zerolog | `starter/zerolog/` | 高性能日志 | ✅ |

### P2 - 扩展集成

| 框架 | 模块 | 路径 | 功能 | 状态 |
|------|------|------|------|------|
| Echo | starter/echo | `starter/echo/` | Web 框架 | ✅ |
| gRPC | starter/grpc | `starter/grpc/` | RPC 框架 | ✅ |
| Nacos | starter/nacos | `starter/nacos/` | 配置中心 | ✅ |
| RabbitMQ | starter/rabbitmq | `starter/rabbitmq/` | 消息队列 | ✅ |

### P3 - 按需集成

| 框架 | 模块 | 路径 | 功能 | 状态 |
|------|------|------|------|------|
| ent | starter/ent | `starter/ent/` | ORM 框架 | ✅ |
| Chi | starter/chi | `starter/chi/` | HTTP 路由器 | ✅ |
| Apollo | starter/apollo | `starter/apollo/` | 配置中心 | ✅ |
| go-micro | starter/micro | `starter/micro/` | 微服务框架 | ✅ |

---

## 快速开始

### 1. Redis 集成

**配置文件** (`application.json`):
```json
{
  "redis": {
    "enabled": true,
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0,
    "prefix": "enhance:",
    "pool_size": 10
  }
}
```

**使用示例**:
```go
package main

import (
    "context"
    "time"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/cache"
    "github.com/xudefa/enhance/core"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("redis-demo"),
    )
    defer app.Stop()

    c := core.MustGetBean[cache.Cache](app.Container())
    ctx := context.Background()

    // 设置缓存
    c.Set(ctx, "user:1", "John", 5*time.Minute)

    // 获取缓存
    val, _ := c.Get(ctx, "user:1")
}
```

### 2. Kafka 集成

**配置文件**:
```json
{
  "kafka": {
    "enabled": true,
    "brokers": ["localhost:9092"],
    "topic": "enhance-events",
    "group_id": "enhance-consumer"
  }
}
```

**使用示例**:
```go
package main

import (
    "context"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/kafka"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("kafka-demo"),
    )
    defer app.Stop()

    queue := core.MustGetBean[*kafka.KafkaQueue](app.Container())
    ctx := context.Background()

    // 发送消息
    queue.Publish(ctx, []byte("Hello, Kafka!"))

    // 订阅消息
    queue.Subscribe(ctx, func(msg []byte) error {
        println("Received:", string(msg))
        return nil
    })
}
```

### 3. OpenTelemetry 集成

**配置文件**:
```json
{
  "otel": {
    "enabled": true,
    "endpoint": "localhost:4317",
    "service_name": "demo-service",
    "service_version": "1.0.0",
    "sampling_rate": 1.0
  }
}
```

**使用示例**:
```go
package main

import (
    "context"
    "go.opentelemetry.io/otel/attribute"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/otel"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("otel-demo"),
    )
    defer app.Stop()

    tracer := core.MustGetBean[*otel.OtelAutoConfiguration](app.Container())
    ctx := context.Background()

    // 创建 Span
    ctx, span := tracer.StartSpan(ctx, "my-operation")
    defer span.End()

    // 设置属性
    otel.SetAttributes(ctx, attribute.String("user.id", "123"))
}
```

### 4. Echo 集成

**配置文件**:
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

**使用示例**:
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

    e := core.MustGetBean[*echo.Echo](app.Container())

    e.GET("/", func(c echo.Context) error {
        return c.String(http.StatusOK, "Hello, Echo!")
    })

    app.Start()
    app.WaitForSignal()
}
```

### 5. gRPC 集成

**配置文件**:
```json
{
  "grpc": {
    "enabled": true,
    "port": 9090,
    "enable_reflection": true,
    "max_recv_msg_size": 4194304,
    "max_send_msg_size": 4194304
  }
}
```

**使用示例**:
```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "google.golang.org/grpc"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("grpc-demo"),
    )
    defer app.Stop()

    server := core.MustGetBean[*grpc.Server](app.Container())

    // 注册服务
    // pb.RegisterUserServiceServer(server, &UserService{})

    app.Start()
    app.WaitForSignal()
}
```

### 6. Nacos 集成

**配置文件**:
```json
{
  "nacos": {
    "enabled": true,
    "server_addr": "127.0.0.1",
    "port": 8848,
    "namespace_id": "public",
    "app_name": "demo-app",
    "username": "nacos",
    "password": "nacos",
    "timeout_ms": 10000,
    "log_dir": "/tmp/nacos/log",
    "cache_dir": "/tmp/nacos/cache",
    "log_level": "info"
  }
}
```

**使用示例**:
```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/nacos"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("nacos-demo"),
    )
    defer app.Stop()

    nacosConfig := core.MustGetBean[*nacos.NacosAutoConfiguration](app.Container())

    // 获取配置
    content, _ := nacosConfig.GetConfig("demo-config", "DEFAULT_GROUP")

    // 监听配置变更
    nacosConfig.ListenConfig("demo-config", "DEFAULT_GROUP", func(newContent string) {
        println("配置已更新:", newContent)
    })
}
```

### 7. RabbitMQ 集成

**配置文件**:
```json
{
  "rabbitmq": {
    "enabled": true,
    "host": "localhost",
    "port": 5672,
    "username": "guest",
    "password": "guest",
    "vhost": "/",
    "queue_name": "enhance-queue",
    "exchange": "",
    "routing_key": "enhance-key",
    "durable": true,
    "auto_delete": false
  }
}
```

**使用示例**:
```go
package main

import (
    "context"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/rabbitmq"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("rabbitmq-demo"),
    )
    defer app.Stop()

    queue := core.MustGetBean[*rabbitmq.RabbitMQQueue](app.Container())
    ctx := context.Background()

    // 发送消息
    queue.Publish(ctx, []byte("Hello, RabbitMQ!"))

    // 订阅消息
    queue.Subscribe(ctx, func(msg []byte) error {
        println("Received:", string(msg))
        return nil
    })
}
```

### 8. ent ORM 集成

**配置文件**:
```json
{
  "ent": {
    "enabled": true,
    "driver": "mysql",
    "dsn": "root:root@tcp(localhost:3306)/enhance?parseTime=True",
    "database": "enhance"
  }
}
```

**使用示例**:
```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/ent"
    entsql "entgo.io/ent/dialect/sql"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("ent-demo"),
    )
    defer app.Stop()

    driver := core.MustGetBean[*entsql.Driver](app.Container())

    // 创建 ent 客户端
    // client := ent.NewClient(ent.Driver(driver))
    // defer client.Close()
}
```

### 9. Chi 集成

**配置文件**:
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

**使用示例**:
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

    router := core.MustGetBean[*chi.Mux](app.Container())

    router.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, Chi!"))
    })

    app.Start()
    app.WaitForSignal()
}
```

### 10. Apollo 集成

**配置文件**:
```json
{
  "apollo": {
    "enabled": true,
    "app_id": "demo-app",
    "cluster": "default",
    "meta_addr": "http://localhost:8080",
    "namespace": "application",
    "is_backup_config": true,
    "secret": ""
  }
}
```

**使用示例**:
```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/apollo"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("apollo-demo"),
    )
    defer app.Stop()

    apolloConfig := core.MustGetBean[*apollo.ApolloAutoConfiguration](app.Container())

    // 获取配置
    value, _ := apolloConfig.GetConfig("demo.key", "application")
}
```

### 11. go-micro 集成

**配置文件**:
```json
{
  "micro": {
    "enabled": true,
    "service_name": "demo-service",
    "version": "1.0.0",
    "registry_addr": ""
  }
}
```

**使用示例**:
```go
package main

import (
    "github.com/go-micro/go-micro/v5"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/micro"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("micro-demo"),
    )
    defer app.Stop()

    service := core.MustGetBean[micro.Service](app.Container())

    // 注册处理器
    // micro.RegisterHandler(service, &Handler{})

    app.Start()
    app.WaitForSignal()
}
```

---

## 集成优先级

| 优先级 | 说明 | 启动顺序 |
|--------|------|----------|
| Infrastructure (-3000) | 基础设施层（日志、配置中心） | 最先执行 |
| DataLayer (-2000) | 数据层（数据库、缓存） | 第二优先 |
| WebLayer (0) | Web 层（HTTP 服务器、路由） | 第三优先 |
| BusinessLayer (1000) | 业务层（消息队列、微服务） | 第四优先 |
| MonitoringLayer (2000) | 监控层（链路追踪、指标） | 最后执行 |

---

## 集成原则

| 原则 | 说明 |
|------|------|
| **零依赖核心** | 核心框架不依赖任何第三方库 |
| **接口隔离** | Integration 层通过接口隔离第三方依赖 |
| **可选集成** | 通过 Starter 机制按需引入 |
| **统一抽象** | 提供统一接口，支持多实现切换 |
| **自动配置** | 引入即自动配置，零配置可用 |

---

## 下一步计划

- [ ] 完善集成测试
- [ ] 添加更多框架支持
- [ ] 优化文档和示例
- [ ] 提供集成最佳实践指南