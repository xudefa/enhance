# enhance Starter 模块

enhance 框架的第三方集成模块集合，提供自动配置支持。

## 文档导航

- [集成总览](INTEGRATION_OVERVIEW.md) - 所有第三方集成的完整清单和使用指南

## 模块清单

### P1 - 重要集成

| 模块 | 说明 | 文档 |
|------|------|------|
| [redis](redis/) | Redis 缓存支持 | [README](redis/README.md) |
| [kafka](kafka/) | Kafka 消息队列 | [README](kafka/README.md) |
| [otel](otel/) | OpenTelemetry 链路追踪 | [README](otel/README.md) |
| [zerolog](zerolog/) | Zerolog 高性能日志 | [README](zerolog/README.md) |
| [gin](gin/) | Gin Web 框架 | [README](gin/README.md) |
| [zap](zap/) | Zap 高性能日志 | [README](zap/README.md) |
| [prometheus](prometheus/) | Prometheus 监控 | [README](prometheus/README.md) |
| [mongodb](mongodb/) | MongoDB 数据库 | [README](mongodb/README.md) |
| [fiber](fiber/) | Fiber Web 框架 | [README](fiber/README.md) |
| [elasticsearch](elasticsearch/) | Elasticsearch 搜索引擎 | [README](elasticsearch/README.md) |
| [consul](consul/) | Consul 服务发现 | [README](consul/README.md) |
| [viper](viper/) | Viper 配置管理 | [README](viper/README.md) |
| [cobra](cobra/) | Cobra CLI 框架 | [README](cobra/README.md) |
| [swagger](swagger/) | Swagger API 文档 | [README](swagger/README.md) |
| [validator](validator/) | Validator 数据验证 | [README](validator/README.md) |
| [cron](cron/) | Cron 定时任务 | [README](cron/README.md) |
| [asynq](asynq/) | Asynq 异步任务队列 | [README](asynq/README.md) |
| [ratelimiter](ratelimiter/) | RateLimiter 限流器 | [README](ratelimiter/README.md) |

### P2 - 扩展集成

| 模块 | 说明 | 文档 |
|------|------|------|
| [echo](echo/) | Echo Web 框架 | [README](echo/README.md) |
| [grpc](grpc/) | gRPC 服务框架 | [README](grpc/README.md) |
| [nacos](nacos/) | Nacos 配置中心 | [README](nacos/README.md) |
| [rabbitmq](rabbitmq/) | RabbitMQ 消息队列 | [README](rabbitmq/README.md) |

### P3 - 按需集成

| 模块 | 说明 | 文档 |
|------|------|------|
| [ent](ent/) | ent ORM 框架 | [README](ent/README.md) |
| [chi](chi/) | Chi HTTP 路由器 | [README](chi/README.md) |
| [apollo](apollo/) | Apollo 配置中心 | [README](apollo/README.md) |
| [micro](micro/) | go-micro 微服务 | [README](micro/README.md) |

### 内置模块

| 模块 | 说明 | 文档 |
|------|------|------|
| [gorm](gorm/) | GORM ORM 框架 | [README](gorm/README.md) |
| [xorm](xorm/) | XORM ORM 框架 | [README](xorm/README.md) |
| [jwt](jwt/) | JWT 认证支持 | [README](jwt/README.md) |
| [casbin](casbin/) | Casbin 权限控制 | [README](casbin/README.md) |
| [casbin-gorm](casbin-gorm/) | Casbin + GORM 集成 | [README](casbin-gorm/README.md) |
| [casbin-xorm](casbin-xorm/) | Casbin + XORM 集成 | [README](casbin-xorm/README.md) |

## 快速开始

### 1. 引入依赖

在 `go.mod` 中添加需要的 starter 模块：

```go
require (
    github.com/xudefa/enhance/starter/redis v0.0.0
    github.com/xudefa/enhance/starter/kafka v0.0.0
)
```

### 2. 配置文件

在 `application.json` 中添加对应配置：

```json
{
  "redis": {
    "enabled": true,
    "host": "localhost",
    "port": 6379
  },
  "kafka": {
    "enabled": true,
    "brokers": ["localhost:9092"],
    "topic": "enhance-events"
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    _ "github.com/xudefa/enhance/starter/redis"
    _ "github.com/xudefa/enhance/starter/kafka"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("demo"),
    )
    defer app.Stop()
    
    app.Start()
    app.WaitForSignal()
}
```

## 启动顺序

| 优先级 | 层级 | 模块 |
|--------|------|------|
| -3000 | 基础设施层 | zerolog, nacos, apollo |
| -2000 | 数据层 | redis, gorm, xorm, ent |
| 0 | Web 层 | echo, grpc, chi |
| 1000 | 业务层 | kafka, rabbitmq, micro |
| 2000 | 监控层 | otel |

## 设计原则

1. **零依赖核心** - 核心框架不依赖任何第三方库
2. **接口隔离** - 通过接口隔离第三方依赖
3. **可选集成** - 通过 Starter 机制按需引入
4. **统一抽象** - 提供统一接口，支持多实现切换
5. **自动配置** - 引入即自动配置，零配置可用

## 开发指南

### 创建新的 Starter 模块

1. 创建 `starter/xxx/` 目录
2. 创建 `go.mod` 定义依赖
3. 创建 `autoconfig.go` 实现自动配置
4. 创建 `doc.go` 添加包文档
5. 创建 `README.md` 添加使用文档

### 自动配置模板

```go
package xxx

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/condition"
)

func init() {
    boot.RegisterAutoConfigWith(&XxxAutoConfiguration{},
        boot.WithConditions(
            condition.OnProperty(XxxEnabled, ConditionTrue),
        ),
        boot.WithOrder(int(boot.OrderPriorityXXX)),
    )
}

type XxxAutoConfiguration struct {
    // 配置字段
}

func (c *XxxAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
    // 配置逻辑
    return nil
}
```

## 贡献指南

欢迎提交新的 Starter 模块，请遵循以下规范：

1. 遵循现有模块的代码风格
2. 提供完整的测试用例
3. 提供详细的使用文档
4. 确保无外部依赖冲突

## 许可证

MIT License