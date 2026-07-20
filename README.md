# enhance

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/enhance)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/enhance)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/enhance/test.yml?branch=master)](https://github.com/xudefa/enhance/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/enhance.svg)](https://pkg.go.dev/github.com/xudefa/enhance) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/enhance)](https://goreportcard.com/report/github.com/xudefa/enhance)

**Go 语言企业级应用开发增强框架** — 参考 Spring Framework 和 Spring Boot 设计，提供 IoC 容器、依赖注入、AOP、自动配置、Actuator 等企业级特性，帮助开发者快速构建可测试、松耦合、高扩展的 Go 应用程序。

> **设计理念**：零外部依赖的工程化框架，借鉴 Spring Boot 的设计思想，为 Go 开发者提供熟悉的企业级开发体验。
> 
> **Go 版本要求**：1.21+ | **当前版本**：v0.0.0

---

## ✨ 核心特性

- 🎯 **零外部依赖** — 核心框架仅使用 Go 标准库，Integration 层通过接口抽象隔离第三方依赖
- 🔄 **IoC 容器** — 完整的依赖注入支持，构造器/字段/方法注入，泛型 API
- 🔀 **AOP 框架** — 5 种通知类型，切点匹配，动态代理，代码生成
- ⚡ **自动配置** — 条件装配，Starter 机制，即插即用
- 📊 **可观测性** — 日志、指标、健康检查、分布式追踪
- 🛡️ **安全框架** — 认证、授权、过滤器链、JWT、Casbin
- 💾 **数据访问** — 泛型 Repository，事务管理，多 ORM 支持
- 🌐 **Web 框架** — HTTP 服务器，MVC 控制器，中间件，WebSocket

---

## 目录

- [快速开始](#快速开始)
- [核心示例](#核心示例)
- [功能特性](#功能特性)
- [架构设计](#架构设计)
- [文档导航](#文档导航)
- [开发指南](#开发指南)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

---

## 快速开始

### 安装

```bash
go get github.com/xudefa/enhance
```

### 5 分钟上手

```go
package main

import (
    "fmt"
    "reflect"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
)

type HelloService struct {
    Message string `inject:"message"`
}

func (s *HelloService) Say() string {
    return s.Message
}

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
    )
    defer app.Stop()

    app.Container().Register(
        reflect.TypeOf(&HelloService{}),
        core.Bean(&HelloService{Message: "Hello, enhance!"}),
    )

    svc := core.MustGetBean[*HelloService](app.Container())
    fmt.Println(svc.Say()) // Output: Hello, enhance!

    app.Start()
    app.WaitForSignal()
}
```

---

## 功能特性

### 核心模块

| 模块 | 说明 | 文档 |
|------|------|------|
| [IoC 容器](core/README.md) | 依赖注入、构造器自动推导、泛型 API | [→](core/README.md) |
| [AOP 框架](aop/README.md) | 5 种通知类型、切点匹配、代码生成 | [→](aop/README.md) |
| [应用启动器](boot/README.md) | 自动配置、Starter 机制 | [→](boot/README.md) |
| [应用上下文](context/README.md) | 聚合容器、环境配置、生命周期 | [→](context/README.md) |
| [条件判断](condition/README.md) | OnProperty / OnBean / OnClass 条件装配 | [→](condition/README.md) |
| [事件驱动](event/README.md) | 发布/订阅、异步事件、事务绑定 | [→](event/README.md) |
| [生命周期](lifecycle/README.md) | 应用生命周期阶段管理 | [→](lifecycle/README.md) |

### 数据与缓存

| 模块 | 说明 | 文档 |
|------|------|------|
| [缓存抽象](cache/README.md) | 统一 Cache 接口、LRU 缓存 | [→](cache/README.md) |
| [邮件发送](email/README.md) | 基于 net/smtp 的邮件发送 | [→](email/README.md) |

### 网络与通信

| 模块 | 说明 | 文档 |
|------|------|------|
| [Web 框架](web/README.md) | HTTP 服务器、路由器、中间件 | [→](web/README.md) |
| [弹性与容错](resilience/README.md) | 服务注册发现、负载均衡、熔断器 | [→](resilience/README.md) |

### 可观测性

| 模块 | 说明 | 文档 |
|------|------|------|
| [日志抽象](log/README.md) | Logger 接口、slog 默认实现 | [→](log/README.md) |
| [运维端点](actuator/README.md) | /health、/metrics、/env 端点 | [→](actuator/README.md) |
| [指标收集](metrics/README.md) | Counter、Gauge、MeterRegistry | [→](metrics/README.md) |
| [可观测性](observability/README.md) | 日志、指标可观测性 | [→](observability/README.md) |

### 安全与验证

| 模块 | 说明 | 文档 |
|------|------|------|
| [安全框架](security/README.md) | 认证、授权、过滤器链 | [→](security/README.md) |
| [数据验证](validation/README.md) | HTTP 请求验证、结构体验证 | [→](validation/README.md) |

### 可靠性与调度

| 模块 | 说明 | 文档 |
|------|------|------|
| [定时任务](schedule/README.md) | Cron 解析、最小堆调度器 | [→](schedule/README.md) |
| [异步执行](async/README.md) | 异步执行器 | [→](async/README.md) |
| [重试机制](retry/README.md) | 重试策略 | [→](retry/README.md) |
| [审计日志](audit/README.md) | 审计日志记录 | [→](audit/README.md) |

### 工具与扩展

| 模块 | 说明 | 文档 |
|------|------|------|
| [配置管理](config/README.md) | Config 接口、Loader 链、Validator | [→](config/README.md) |
| [国际化](i18n/README.md) | MessageSource 消息源、多区域支持 | [→](i18n/README.md) |
| [异常处理](exception/README.md) | 全局异常处理、错误码体系 | [→](exception/README.md) |
| [表达式语言](spel/README.md) | SpEL 风格表达式解析 | [→](spel/README.md) |
| [测试框架](testing/README.md) | TestRunner、Mock、断言 | [→](testing/README.md) |

---

## 核心示例

### IoC 容器

```go
package main

import (
    "fmt"
    "reflect"
    "github.com/xudefa/enhance/core"
)

type UserService struct {
    DB *Database `inject:"database"`
}

func (s *UserService) GetUser(id int) string {
    return fmt.Sprintf("User %d", id)
}

type Database struct{}

func main() {
    c := core.New()
    c.Register(reflect.TypeOf(&Database{}), core.Bean(&Database{}))
    c.Register(reflect.TypeOf(&UserService{}), core.Bean(&UserService{}))
    
    svc := core.MustGetBean[*UserService](c)
    fmt.Println(svc.GetUser(1)) // Output: User 1
}
```

### AOP 切面编程

```go
package main

import (
    "fmt"
    "reflect"
    "github.com/xudefa/enhance/aop"
    "github.com/xudefa/enhance/core"
)

type LoggingAspect struct{}

func (l *LoggingAspect) Before(ctx aop.JoinPoint) {
    fmt.Printf("Before: %s\n", ctx.Method().Name)
}

func (l *LoggingAspect) After(ctx aop.JoinPoint, result any, err error) {
    fmt.Printf("After: %s, result: %v\n", ctx.Method().Name, result)
}

type UserService struct{}

func (u *UserService) GetUser(id int) string {
    return fmt.Sprintf("User %d", id)
}

func main() {
    c := core.New()
    c.Register(reflect.TypeOf(&UserService{}), core.Bean(&UserService{}))
    aop.RegisterAspect(&LoggingAspect{}, aop.WithPointcut("*UserService.*"))
    
    svc := core.MustGetBean[*UserService](c)
    svc.GetUser(1)
    // Output:
    // Before: GetUser
    // After: GetUser, result: User 1
}
```

### 事件驱动

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/event"
)

type UserCreatedEvent struct {
    event.BaseEvent
    UserID int
}

func (e *UserCreatedEvent) EventName() string {
    return "user.created"
}

func main() {
    bus := event.NewEventBus()
    bus.Subscribe("user.created", func(e event.ApplicationEvent) {
        evt := e.(*UserCreatedEvent)
        fmt.Printf("User created: %d\n", evt.UserID)
    })
    bus.Publish(&UserCreatedEvent{UserID: 123})
    // Output: User created: 123
}
```

### 定时任务

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/schedule"
)

func main() {
    scheduler := schedule.NewScheduler()
    scheduler.Schedule("*/5 * * * * ?", func() {
        fmt.Println("Task executed at:", schedule.Now())
    })
    scheduler.Start()
    defer scheduler.Stop()
    select {}
}
```

---

## 架构设计

### 三层架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Boot Layer                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐   │
│  │ Application  │  │ AutoConfig   │  │ Starter Packages         │   │
│  │ Runner       │  │ Registry     │  │ (web, data, security...) │   │
│  └──────────────┘  └──────────────┘  └──────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────┐
│                         Core Layer                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ IoC      │ │ AOP      │ │ Event    │ │ Lifecycle│ │ Config   │   │
│  │ Container│ │ Engine   │ │ Bus      │ │ Manager  │ │ Manager  │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────┐
│                         Infrastructure Layer                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ Data     │ │ Web      │ │ Security │ │ Actuator │ │ Schedule │   │
│  │ Access   │ │ Server   │ │ Framework│ │ Endpoints│ │ Engine   │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### 架构分层说明

| 层级 | 职责 | 核心模块 |
|------|------|----------|
| **Boot Layer** | 应用启动、自动配置、Starter 管理 | `boot`, `condition`, `context` |
| **Core Layer** | IoC 容器、AOP、事件驱动、配置管理 | `core`, `aop`, `event`, `config`, `lifecycle` |
| **Infrastructure Layer** | Web、安全、监控、缓存、数据访问等基础设施 | `web`, `security`, `actuator`, `cache`, `schedule`, `log`, `metrics` |

### 核心接口速查

| 接口 | 所属层 | 说明 |
|------|--------|------|
| `core.Container` | Core | IoC 容器：Register、Get、Resolve |
| `aop.Advice` | Core | AOP 通知：Before、After、Around 等 |
| `boot.Application` | Boot | 应用实例：Start、Stop、Container |
| `cache.Cache` | Infrastructure | 缓存操作：Get、Set、Del |
| `config.Config` | Core | 配置访问：Get、Unmarshal、Watch |
| `log.Logger` | Infrastructure | 日志记录：Debug、Info、Error |
| `metrics.MeterRegistry` | Infrastructure | 指标注册：Counter、Gauge、Timer |

详细接口定义请参阅 [架构设计文档](ARCHITECTURE.md)。

---

## Go 风格优化

enhance 框架经过六大 Go 风格优化，从 Java/Spring 风格转向符合 Go 惯用法的设计：

| 优化项 | Java 风格 | Go 惯用法 | 收益 |
|--------|-----------|-----------|------|
| **泛型注入** | `Factory(..., reflect.TypeOf(...))` | `FactoryOf[T](...)` + `MustGetBean[T]` | 速度提升 ~3x |
| **条件装配** | 实现 `Condition` 接口 | `OnProperty()`, `OnBean()` 函数式条件 | 代码减少 60% |
| **模块组合** | `func init()` 隐式注册 | `NewModule().Bean().Condition()` 显式组合 | 依赖清晰 |
| **生命周期** | 7 阶段状态机 | `OnInit/OnStart/OnStop` 3 阶段钩子 | 简化 57% |
| **配置绑定** | `GetProperty("key")` | `BindConfig[T]()` 结构体绑定 | 类型安全 |
| **AOP 代理** | 运行时反射代理 | `//go:generate` 编译时代码生成 | 零反射开销 |

详细优化说明请参阅 [架构设计文档](ARCHITECTURE.md)。

---

## 文档导航

### 核心文档

| 文档 | 说明 | 读者 |
|------|------|------|
| [架构设计](ARCHITECTURE.md) | 三层架构、核心模块、接口定义 | 开发者、架构师 |
| [模块依赖](DEPENDENCIES.md) | 模块依赖图、执行顺序、优先级 | 开发者、架构师 |
| [开发规范](AGENTS.md) | AI 智能体编码规范 | AI 智能体、开发者 |
| [代码风格](CODING_STYLE.md) | 命名、注释、组织规范 | 开发者 |
| [贡献指南](CONTRIBUTING.md) | 参与开发、提交 PR | 贡献者 |
| [集成总览](starter/INTEGRATION_OVERVIEW.md) | 第三方集成清单和指南 | 开发者、架构师 |

### 子模块文档

#### 核心框架

| 文档 | 说明 |
|------|------|
| [IoC 容器](core/README.md) | Bean 注册、依赖注入、泛型 API |
| [AOP 框架](aop/README.md) | 切点匹配、通知类型、动态代理 |
| [应用启动器](boot/README.md) | 自动配置、Starter 机制 |
| [应用上下文](context/README.md) | 聚合容器、环境配置、生命周期 |
| [条件判断](condition/README.md) | OnProperty / OnBean / OnClass 条件装配 |
| [生命周期](lifecycle/README.md) | 应用生命周期阶段管理 |

#### 数据与缓存

| 文档 | 说明 |
|------|------|
| [缓存抽象](cache/README.md) | 统一 Cache 接口、LRU 缓存 |
| [邮件发送](email/README.md) | 基于 net/smtp 的邮件发送 |

#### Starter 集成

| 文档 | 说明 |
|------|------|
| [Redis](starter/redis/README.md) | Redis 缓存支持 |
| [Kafka](starter/kafka/README.md) | Kafka 消息队列 |
| [OpenTelemetry](starter/otel/README.md) | OpenTelemetry 链路追踪 |
| [GORM](starter/gorm/README.md) | GORM 数据库集成 |
| [XORM](starter/xorm/README.md) | XORM 数据库集成 |
| [JWT](starter/jwt/README.md) | JWT 认证集成 |
| [Casbin](starter/casbin/README.md) | Casbin 授权集成 |

#### 网络与通信

| 文档 | 说明 |
|------|------|
| [Web 框架](web/README.md) | HTTP 服务器、路由器、中间件 |
| [弹性与容错](resilience/README.md) | 服务注册发现、负载均衡、熔断器 |

#### 可观测性

| 文档 | 说明 |
|------|------|
| [日志抽象](log/README.md) | Logger 接口、slog 默认实现 |
| [运维端点](actuator/README.md) | /health、/metrics、/env 端点 |
| [指标收集](metrics/README.md) | Counter、Gauge、MeterRegistry |
| [可观测性](observability/README.md) | 日志、指标可观测性 |

#### 安全与验证

| 文档 | 说明 |
|------|------|
| [安全框架](security/README.md) | 认证、授权、过滤器链 |
| [数据验证](validation/README.md) | HTTP 请求验证、结构体验证 |

#### 可靠性与调度

| 文档 | 说明 |
|------|------|
| [事件驱动](event/README.md) | 发布/订阅、异步事件、事务绑定 |
| [定时任务](schedule/README.md) | Cron 解析、最小堆调度器 |
| [异步执行](async/README.md) | 异步执行器 |
| [重试机制](retry/README.md) | 重试策略 |
| [审计日志](audit/README.md) | 审计日志记录 |

#### 工具与扩展

| 文档 | 说明 |
|------|------|
| [国际化](i18n/README.md) | MessageSource 消息源、多区域支持 |
| [异常处理](exception/README.md) | 全局异常处理、错误码体系 |
| [表达式语言](spel/README.md) | SpEL 风格表达式解析 |
| [元数据](metadata/README.md) | 元数据管理、解析器 |
| [测试框架](testing/README.md) | TestRunner、Mock、断言 |
| [开发工具](devtools/README.md) | 开发辅助工具 |

---

## 开发指南

### 构建与测试

```bash
# 构建
go build ./...

# 测试
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测

# 代码规范
go fmt ./...
golangci-lint run
```

### 代码规范速查

- **命名**：包名小写，导出标识符大写驼峰，错误变量 `Err` 前缀
- **导入**：标准库 → 项目内部包，分组空行分隔
- **选项模式**：优先使用函数式选项模式
- **错误处理**：不忽略错误，使用 `%w` 包装
- **注释**：中文注释，导出类型/函数必须有 godoc

详细规范请参阅 [开发规范](AGENTS.md) 和 [代码风格](CODING_STYLE.md)。

### 架构设计原则

| 原则 | 说明 |
|------|------|
| 零外部依赖 | 核心框架仅使用 Go 标准库 |
| 接口优先 | 优先定义接口，再提供默认实现 |
| 函数式选项 | 使用函数式选项模式提供灵活配置 |
| Go 语言优先 | 参考 Spring 哲学，遵循 Go 惯用法 |

---

## 贡献指南

欢迎提交 Issue 和 Pull Request！

- 📖 [贡献指南](CONTRIBUTING.md) — 如何参与项目开发
- 🤖 [开发规范](AGENTS.md) — AI 智能体开发规范
- 🏗️ [架构设计](ARCHITECTURE.md) — 详细架构设计文档

### 快速贡献

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 提交更改：`git commit -m 'feat: add my feature'`
4. 推送分支：`git push origin feature/my-feature`
5. 在 GitHub 上创建 Pull Request

详细贡献流程请参阅 [贡献指南](CONTRIBUTING.md)。

---

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。