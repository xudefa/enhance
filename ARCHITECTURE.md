# enhance 架构设计文档

> **文档版本**：v0.0 | **最后更新**：2026-07-29

本文档详细描述 enhance 框架的架构设计、核心模块、接口定义和扩展点。

---

## 目录

- [架构概述](#架构概述)
- [核心模块](#核心模块)
- [自动配置机制](#自动配置机制)
- [扩展点设计](#扩展点设计)
- [性能优化](#性能优化)
- [最佳实践](#最佳实践)

---

## 架构概述

enhance 是一个参考 Spring Framework 和 Spring Boot 设计的 Go 企业级项目开发增强框架，旨在帮助开发者快速搭建 Go 项目，并提供极高的兼容扩展能力。

### 设计目标

| 目标 | 说明 |
|------|------|
| **快速开发** | 提供类似 Spring Boot 的开发体验 |
| **零外部依赖** | 核心框架仅使用 Go 标准库 |
| **高扩展性** | 模块化设计，支持插件化扩展 |
| **企业级特性** | IoC、AOP、数据访问、安全、监控等完整生态 |

### 架构分层

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

### 包结构清单

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `core/` | IoC 容器（依赖注入、组件扫描、泛型 API） | `core.Container` |
| `aop/` | AOP 框架（5 种通知 + 切点匹配 + 代码生成） | `aop.Advice`, `aop.PointCut` |
| `boot/` | 应用启动器、自动配置、横幅、失败分析 | `boot.AutoConfiguration` |
| `context/` | 应用上下文（聚合容器、环境、生命周期） | `context.ApplicationContext` |
| `condition/` | 条件判断（OnProperty / OnBean / OnModuleLoaded） | `condition.Condition` |
| `event/` | 事件驱动（发布/订阅、死信队列、事务绑定） | `event.EventBus`, `event.EventBusWithOrdering` |
| `cache/` | 缓存抽象（LRU + 内存缓存） | `cache.Cache` |
| `web/` | HTTP 服务器/客户端、路由器、中间件 | `web.Server`, `web.Router` |
| `security/` | 安全框架（认证 + 授权 + 过滤器链） | `security.Authentication`, `security.HttpSecurity` |
| `actuator/` | 运维端点（/health, /metrics, /env） | `actuator.Endpoint` |
| `metrics/` | 指标收集（Counter + Gauge + Histogram） | `metrics.MeterRegistry`, `metrics.Exporter` |
| `schedule/` | 定时任务调度（Cron + 6 字段表达式） | `schedule.Scheduler`, `schedule.Task` |
| `log/` | 日志抽象（Logger 接口 + slog/zerolog 集成） | `log.Logger` |
| `async/` | 异步执行（goroutine 池 + Future） | `async.AsyncExecutor`, `async.Future` |
| `audit/` | 审计日志（事件记录 + 拦截器） | `audit.Auditor`, `audit.AuditLogger` |
| `i18n/` | 国际化（多区域消息源） | `i18n.MessageSource`, `i18n.Locale` |
| `validation/` | 数据验证（字段级 + 跨字段验证） | `validation.Validator`, `validation.MiddlewareValidator` |
| `exception/` | 异常处理（全局异常处理器） | `exception.ExceptionHandler`, `exception.ErrorCode` |
| `spel/` | 表达式语言（SpEL 风格） | `spel.ExpressionParser`, `spel.EvaluationContext` |
| `lifecycle/` | 生命周期管理（3 阶段状态机） | `lifecycle.LifecycleManager`, `lifecycle.PhaseListener` |
| `resilience/` | 弹性容错（熔断器 + 负载均衡） | `resilience.Breaker`, `resilience.Selector`, `resilience.Balancer` |
| `mq/` | 消息队列（统一抽象） | `mq.Queue`, `mq.Message` |
| `tenant/` | 多租户（租户隔离 + 解析器） | `tenant.TenantResolver`, `tenant.TenantManager` |
| `tracing/` | 分布式追踪（链路追踪） | `tracing.Tracer`, `tracing.Span` |
| `openapi/` | OpenAPI 文档（自动生成） | `openapi.Document` |
| `testing/` | 测试框架（TestRunner + Mock） | `testing.TestRunner` |
| `devtools/` | 开发工具（热重载） | `devtools.HotReloader`, `devtools.FileWatcher` |
| `email/` | 邮件发送（SMTP） | `email.Sender`, `email.Message` |

---

## 核心模块

> **说明**：本节详细描述各核心模块的职责、接口和使用方式。更多示例请参阅各模块的 README.md 文档。

### 1. IoC 容器

#### 模块职责

- Bean 注册与管理
- 依赖注入（构造器、字段、方法注入）
- Bean 生命周期管理
- 作用域管理（Singleton、Prototype）
- 泛型 API 支持

#### 核心接口

```go
// Container IoC 容器接口
type Container interface {
    // 查询
    Get(typ reflect.Type) ([]any, error)
    GetByTypeAndName(name string, typ reflect.Type) (any, error)
    GetAll() []any
    Has(name string, typ reflect.Type) bool
    HasType(typ reflect.Type) bool
    Types() []reflect.Type
    ListBeans() map[string]*registry.BeanDef
    
    // 注册
    RegisterBean(def registry.BeanDef) error
    RegisterInstance(instance any, typ reflect.Type) error
    
    // Bean ID 生成
    Generate(typ reflect.Type, customName ...string) string
    Parse(beanID string) (pkgPath, typeName, customName string)
    
    // Bean 创建
    CreateBean(beanID string) (any, error)
    
    // 生命周期
    Initialize() error
    Destroy() error
}
```

#### 依赖注入策略

| 注入方式 | 说明 | 示例 |
|---------|------|------|
| 构造器注入 | 通过构造函数注入依赖 | `NewService(repo)` |
| 字段注入 | 通过结构体标签注入 | `` `inject:"repository"` `` |
| 方法注入 | 通过 setter 方法注入 | `SetRepo(repo)` |

#### Bean 生命周期

```
实例化 → 属性填充 → 初始化 → 使用中 → 销毁
   │          │          │          │        │
   ▼          ▼          ▼          ▼        ▼
 创建实例  依赖注入   回调方法   正常服务   清理资源
```

生命周期回调（Go 风格函数类型）：

```go
type BeanInitFunc func(bean any) error
type BeanDestroyFunc func(bean any) error
```

---

### 2. AOP 框架

#### 模块职责

- 切点表达式匹配
- 通知执行
- 动态代理生成
- 拦截器链管理
- 代码生成支持

#### 核心接口

```go
// Pointcut 切点接口
type Pointcut interface {
    Matches(target any, methodName string) bool
    MatchClass(t reflect.Type) bool
    Expression() string
}

// Advice 通知接口
type Advice interface {
    Type() AdviceType
    Order() int
    Execute(ctx context.Context, joinPoint JoinPoint) (any, error)
}

// Advisor 通知器接口
type Advisor interface {
    Advice() Advice
    PointCut() Pointcut
    Order() int
}

// JoinPoint 连接点接口
type JoinPoint interface {
    Target() any
    Method() string
    Args() []any
    Proceed() (any, error)
    ProceedWithArgs(args []any) (any, error)
}
```

#### 通知类型

| 通知类型 | 执行时机 | 用途 |
|---------|---------|------|
| Before | 目标方法前 | 日志、权限检查 |
| After | 目标方法后 | 资源清理 |
| AfterReturning | 方法成功返回后 | 结果处理 |
| AfterThrowing | 方法抛出异常后 | 异常处理 |
| Around | 包裹目标方法 | 事务、性能监控 |

#### 切点表达式

```go
// 支持的匹配模式
aop.WithPointcut("*Service.*")           // 匹配所有 Service 的方法
aop.WithPointcut("UserService.GetUser")  // 精确匹配
aop.WithPointcut("prefix.*")             // 前缀匹配
aop.WithPointcut("regex:.*Service$")     // 正则匹配
```

---

### 3. 事件驱动

#### 模块职责

- 事件发布与订阅
- 异步事件处理
- 事件过滤器
- 事务事件支持
- 死信队列

#### 核心接口

```go
// EventBus 事件总线（结构体，非接口）
type EventBus struct {
    listeners sync.Map
}

// EventBusWithOrdering 支持优先级和过滤条件的事件总线
type EventBusWithOrdering struct {
    mu        sync.Mutex
    listeners sync.Map
    nextID    int
}

// ApplicationEvent 应用事件接口
type ApplicationEvent interface {
    Type() string
    Timestamp() time.Time
}

// EventListener 事件监听器函数类型
type EventListener func(event ApplicationEvent)
```

#### 内置事件

| 事件 | 触发时机 |
|------|---------|
| ApplicationStartingEvent | 应用启动开始 |
| ApplicationStartedEvent | 应用启动完成 |
| ApplicationReadyEvent | 应用准备就绪 |
| ApplicationStoppingEvent | 应用停止开始 |
| ApplicationStoppedEvent | 应用停止完成 |

---

### 4. 配置管理

#### 模块职责

- 多源配置加载
- 配置绑定与转换
- 配置热刷新
- 配置验证

#### 配置源优先级

```
命令行参数 > 环境变量 > 配置文件 > 默认值
```

#### 核心接口

```go
// Config 配置接口
type Config interface {
    Get(key string) any
    GetString(key string) string
    GetInt(key string) int
    GetBool(key string) bool
    GetAll() map[string]any
    Set(key string, value any)
    Load(path string) error
    Save(path string) error
}

// Loader 配置加载器
type Loader interface {
    Load(opts ...LoaderOption) (Config, error)
    Priority() int
    SupportsWatch() bool
}
```

#### 支持的配置格式

| 格式 | 扩展名 | 说明 |
|------|--------|------|
| JSON | .json | 默认格式 |
| YAML | .yml, .yaml | 推荐格式 |
| Properties | .properties | 键值对格式 |
| TOML | .toml | TOML 格式 |

---

### 5. 数据访问

> **注意**：数据访问模块通过接口抽象实现，核心框架不直接依赖任何 ORM。

#### 模块职责

- 泛型 Repository 接口
- 事务管理
- 连接池管理
- SQL 构建器

#### 核心接口

```go
// Repository 泛型仓储接口
type Repository[T any] interface {
    Create(entity *T) error
    CreateBatch(entities []T) error
    Delete(id any) error
    Update(entity *T) error
    FindByID(id any) (*T, error)
    FindOne(where string, args ...any) (*T, error)
    FindAll(where string, args ...any) ([]T, error)
    Count(where string, args ...any) (int64, error)
}

// Transactor 事务管理器
type Transactor interface {
    Begin(opts ...TransactionOption) (Transaction, error)
    Execute(ctx context.Context, fn func(tx Transaction) error) error
}
```

---

### 6. Web 框架

#### 模块职责

- HTTP 服务器抽象
- 路由管理
- 请求/响应处理
- 中间件链
- MVC 控制器支持

#### 核心接口

```go
// Server HTTP 服务器接口
type Server interface {
    Start() error
    Stop(ctx context.Context) error
    SetHandler(handler any)
    Use(middleware any)
}

// HandlerFunc HTTP 处理函数
type HandlerFunc func(ctx Context)

// Router 路由器接口
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

---

### 7. 安全框架

#### 模块职责

- 认证管理
- 授权控制
- 过滤器链
- JWT 支持
- Casbin 集成

#### 核心组件

```go
// SecurityFilter 安全过滤器（类型别名 → filter.Filter）
type SecurityFilter = filter.Filter

// AuthenticationManager 认证管理器（类型别名 → authentication.AuthenticationManager）
type AuthenticationManager = authentication.AuthenticationManager

// AccessDecisionManager 访问决策管理器（类型别名 → authorization.AccessDecisionManager）
type AccessDecisionManager = authorization.AccessDecisionManager

// HttpSecurity HTTP 安全配置接口（16 个方法，链式 API）
type HttpSecurity interface {
    AuthenticationManager(authManager AuthenticationManager) HttpSecurity
    UserDetailsService(userDetailsService UserDetailsService) HttpSecurity
    AuthorizeRequests(authorizer AuthorizeRequests) HttpSecurity
    AddFilter(filter SecurityFilter) HttpSecurity
    Csrf() HttpSecurity
    FormLogin(loginProcessingUrl string, defaultSuccessUrl ...string) HttpSecurity
    Logout(logoutUrl string, successHandler ...LogoutSuccessHandler) HttpSecurity
    Build() (SecurityFilterChain, error)
    // ... 更多方法
}
```

---

### 8. Actuator

#### 模块职责

- 健康检查
- 指标收集
- 审计日志
- 环境信息

#### 内置端点

| 端点 | 路径 | 说明 |
|------|------|------|
| health | /actuator/health | 健康检查 |
| metrics | /actuator/metrics | 指标信息 |
| env | /actuator/env | 环境配置 |
| info | /actuator/info | 应用信息 |
| beans | /actuator/beans | Bean 列表 |

#### 健康检查器

```go
// Indicator 健康指标
type Indicator interface {
    Name() string
    Health(ctx context.Context) Health
}

// Health 健康信息
type Health struct {
    Status    Status
    Details   map[string]any
    Error     error
    Timestamp time.Time
}

// Status 健康状态
type Status int

const (
    StatusUp      Status = iota // 正常
    StatusDown                   // 不可用
    StatusDegraded               // 降级
    StatusOutage                 // 停服
    StatusUnknown                // 未知
)
```

---

### 9. 缓存

#### 模块职责

- 统一缓存接口
- 缓存注解支持
- 缓存策略管理
- 多级缓存
- LRU 淘汰算法

#### 核心接口

```go
// Cache 缓存接口
type Cache interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Del(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    Close() error
}
```

---

### 10. 调度

#### 模块职责

- Cron 表达式解析
- 任务调度
- 任务管理
- 执行监控

#### 核心接口

```go
// Scheduler 调度器
type Scheduler interface {
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
    Register(task Task) error
    Unregister(name string) bool
    IsRunning() bool
    RegisteredTasks() []Task
}

// Task 定时任务接口
type Task interface {
    Name() string
    Cron() string
    Execute(ctx context.Context) error
}
```

---

### 11. 日志抽象

#### 模块职责

- 统一日志接口
- 结构化日志
- 多级别支持
- 上下文传播

#### 核心接口

```go
// Logger 日志接口
type Logger interface {
    Debug(ctx context.Context, msg string, keys ...KeyValue)
    Info(ctx context.Context, msg string, keys ...KeyValue)
    Warn(ctx context.Context, msg string, keys ...KeyValue)
    Error(ctx context.Context, msg string, keys ...KeyValue)
    Sync() error
    With(ctx context.Context, keys ...KeyValue) Logger
}

// KeyValue 日志键值对
type KeyValue struct {
    Key   string
    Value any
}
```

---

### 12. 指标收集

#### 模块职责

- Counter 计数器
- Gauge 仪表盘
- Histogram 直方图
- 指标导出

#### 核心接口

```go
// MeterRegistry 指标注册表
type MeterRegistry interface {
    Counter(name string, tags ...string) Counter
    Gauge(name string, tags ...string) Gauge
    Histogram(name string, tags ...string) Histogram
    Collect() []Metric
    RegisterExporter(exporter Exporter)
    Export() error
}

// Exporter 指标导出器
type Exporter interface {
    Export(metrics []Metric) error
}
```

---

### 13. 异步执行

#### 模块职责

- 线程池管理
- Future 模式
- 优雅关闭

#### 核心接口

```go
// AsyncExecutor 异步执行器（结构体，非接口）
type AsyncExecutor struct {
    workerCount int
    queueSize   int
    // ... 内部字段
}

// Future 异步结果（结构体，非接口）
type Future struct {
    done   chan struct{}
    result any
    err    error
    mu     sync.RWMutex
}
```

---

### 14. 审计日志

#### 模块职责

- 操作审计
- 事件记录
- 异步处理
- 多写入器支持

#### 核心接口

```go
// Auditor 审计日志器
type Auditor interface {
    Log(event Event)
    Close() error
    IsClosed() bool
}

// Event 审计事件
type Event struct {
    ID         string
    Timestamp  time.Time
    Actor      string
    Action     EventType
    Resource   string
    Target     string
    Details    map[string]any
    Severity   EventSeverity
    Source     string
    Result     string
    Duration   time.Duration
    Tags       []string
}

// AuditLogger 审计日志助手
type AuditLogger interface {
    Create(resource string, target string, details map[string]any)
    Update(resource string, target string, details map[string]any)
    Delete(resource string, target string)
    Login(target string, details map[string]any)
    Severity(resource string, target string, severity EventSeverity, details map[string]any)
}
```

---

### 15. 国际化

#### 模块职责

- 多区域消息源
- 消息格式化
- 回退机制

#### 核心接口

```go
// MessageSource 消息源
type MessageSource interface {
    GetMessage(code string, args ...any) string
    GetMessageWithLocale(code string, locale Locale, args ...any) string
}

// Locale 区域设置
type Locale struct {
    Language string
    Country  string
    Variant  string
}
```

---

### 16. 数据验证

#### 模块职责

- 字段级验证
- 跨字段验证
- 自定义验证器
- HTTP 中间件集成

#### 核心接口

```go
// Validator 验证器
type Validator interface {
    Validate(obj any) error
}

// ValidationError 验证错误
type ValidationError struct {
    Field   string
    Message string
    Value   any
}

// MiddlewareValidator 中间件验证器
type MiddlewareValidator interface {
    ValidateRequest(c any, obj any) error
    HandleValidationError(c any, err error)
}
```

---

### 17. 异常处理

#### 模块职责

- 全局异常处理
- 错误码体系
- HTTP 异常映射

#### 核心接口

```go
// ExceptionHandler 异常处理器
type ExceptionHandler interface {
    Handle(ctx context.Context, err error, response ResponseWriter) *ErrorResponse
    RegisterResolver(resolver ExceptionResolver)
    RegisterException(exceptionType reflect.Type, resolver ExceptionResolver)
}

// ErrorCode 错误码
type ErrorCode struct {
    Code    int
    Message string
    Detail  string
}
```

---

### 18. 表达式语言

#### 模块职责

- SpEL 风格表达式
- 属性访问
- 方法调用
- 变量管理

#### 核心接口

```go
// ExpressionParser 表达式解析器
type ExpressionParser interface {
    ParseExpression(expression string) (Expression, error)
}

// Expression 表达式
type Expression interface {
    GetValue(context EvaluationContext) (any, error)
    SetValue(context EvaluationContext, value any) error
    String() string
}

// EvaluationContext 表达式求值上下文
type EvaluationContext interface {
    GetRootObject() any
    SetRootObject(root any)
    GetVariable(name string) (any, bool)
    SetVariable(name string, value any)
}
```

---

### 19. 生命周期管理

#### 模块职责

- 3 阶段状态机（INIT → RUNNING → STOPPED）
- 生命周期回调
- 阶段监听器

#### 核心接口

```go
// LifecycleManager 生命周期管理器
type LifecycleManager struct {
    mu        sync.RWMutex
    phase     ApplicationPhase
    listeners []PhaseListener
}

// ApplicationPhase 应用阶段
type ApplicationPhase int

const (
    PhaseInit    ApplicationPhase = iota // 初始化阶段
    PhaseRunning                         // 运行阶段
    PhaseStopped                         // 已停止
)

// PhaseListener 生命周期阶段监听器
type PhaseListener interface {
    OnPhaseChange(oldPhase, newPhase ApplicationPhase) error
}
```

---

### 20. 弹性容错

#### 模块职责

- 熔断器
- 负载均衡
- 服务注册发现
- 限流

#### 核心接口

```go
// Breaker 熔断器
type Breaker interface {
    Allow() error
    RecordSuccess()
    RecordFailure()
    State() State
}

// Selector 负载均衡选择器
type Selector interface {
    Select(instances []InstanceInfo) (InstanceInfo, error)
}

// Registry 服务注册中心
type Registry interface {
    Register(ctx context.Context, info InstanceInfo) error
    Deregister(ctx context.Context, info InstanceInfo) error
    Discover(ctx context.Context, serviceName string) ([]InstanceInfo, error)
}

// Balancer 负载均衡
type Balancer interface {
    Next(backends []*ServiceInstance) (*ServiceInstance, error)
}
```

---

### 21. 消息队列

#### 模块职责

- 统一消息接口
- 多后端支持
- 消息确认（Ack/Nack）
- 死信队列
- 对象池优化

#### 核心接口

```go
// Queue 消息队列
type Queue interface {
    Send(msg *Message) error
    Receive() (*Message, error)
    ReceiveWithTimeout(timeout time.Duration) (*Message, error)
    Consume(handler MessageHandler) error
    StopConsuming()
    Purge() error
    Close() error
    Name() string
    Size() int
}

// Message 消息对象
type Message struct {
    ID         string
    Body       []byte
    Headers    map[string]string
    Timestamp  time.Time
    QueueName  string
    RetryCount int
    MaxRetries int
}
```

---

### 22. 多租户

#### 模块职责

- 租户解析
- 租户隔离
- 租户上下文
- 租户管理

#### 核心接口

```go
// TenantResolver 租户解析器
type TenantResolver interface {
    Resolve(req *http.Request) (string, error)
}

// TenantManager 租户管理器
type TenantManager interface {
    RegisterTenant(tenant *Tenant)
    GetTenant(tenantID string) (*Tenant, error)
    SetCurrentTenant(tenantID string) error
    GetCurrentTenant() *Tenant
}
```

---

### 23. 分布式追踪

#### 模块职责

- 链路追踪
- Span 管理
- 采样策略
- 上下文传播

#### 核心接口

```go
// Tracer 追踪器（结构体）
type Tracer struct {
    serviceName string
    sampler     Sampler
    exporter    Exporter
    maxSpans    int
}

// Sampler 采样器接口
type Sampler interface {
    ShouldSample() bool
}

// Exporter 导出器接口
type Exporter interface {
    ExportSpans(spans []*Span) error
}
```

---

### 24. OpenAPI 文档

#### 模块职责

- 自动生成文档
- OpenAPI 3.0 规范
- Swagger UI 支持

#### 核心接口

```go
// Document OpenAPI 文档
type Document interface {
    SetInfo(title, version, description string)
    AddPath(path string, method string, operation Operation)
    ToJSON() ([]byte, error)
    ToYAML() ([]byte, error)
}
```

---

### 25. 测试框架

#### 模块职责

- 集成测试支持
- Mock 对象
- 断言工具

#### 核心接口

```go
// TestRunner 测试运行器
type TestRunner interface {
    Run(fn func(ctx TestContext))
}

// TestContext 测试上下文
type TestContext interface {
    T() TestingT
    Container() any
    Register(name string, bean any)
    AddCleanup(fn func())
}

// Mock 模拟对象
type Mock interface {
    Expect(method string, args []any, result any, err error) Mock
    Call(method string, args ...any) (any, error)
    Verify() error
}
```

---

### 26. 开发工具

#### 模块职责

- 热重载支持
- 文件监控
- 开发模式检测

#### 核心接口

```go
// HotReloader 热重载管理器
type HotReloader interface {
    Start() error
    Stop()
    OnReload(callback ReloadCallback)
}

// FileWatcher 文件监控器
type FileWatcher interface {
    OnChange(callback ReloadCallback)
    Start() error
    Stop()
}
```

---

### 27. 邮件发送

#### 模块职责

- SMTP 邮件发送
- 多收件人支持
- HTML 格式邮件

#### 核心接口

```go
// Sender 邮件发送器
type Sender interface {
    Send(ctx context.Context, msg *Message) error
    Close() error
}

// Message 邮件消息
type Message struct {
    From        string
    To          []string
    Subject     string
    Body        string
    HTML        string
    Attachments []Attachment
}
```

---

## 自动配置机制

### 自动配置原理

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  ClassPath   │────▶│  Condition   │────▶│  Register    │
│  Scanner     │     │  Evaluator   │     │  Beans       │
└──────────────┘     └──────────────┘     └──────────────┘
```

### 条件注解

| 条件 | 说明 |
|------|------|
| `OnProperty` | 配置属性存在且匹配 |
| `OnBean` | Bean 存在 |
| `OnMissingBean` | Bean 不存在 |
| `OnModuleLoaded` | 模块已加载（Go 替代 OnClass） |
| `OnMissingModule` | 模块未加载 |

### Starter 包结构

```
starter-web/
├── go.mod               # 独立模块
├── autoconfig.go        # 自动配置实现
├── conditions.go        # 条件判断
└── README.md            # 使用说明
```

---

## 扩展点设计

### BeanPostProcessor

```go
// BeanPostProcessor Bean 后处理器
type BeanPostProcessor interface {
    PostProcessBeforeInitialization(bean any, name string) (any, error)
    PostProcessAfterInitialization(bean any, name string) (any, error)
}
```

### BeanFactoryPostProcessor

```go
// BeanFactoryPostProcessor Bean 工厂后处理器
type BeanFactoryPostProcessor interface {
    PostProcessBeanFactory(factory Container) error
}
```

### ApplicationListener

```go
// ApplicationListener 应用事件监听器
type ApplicationListener interface {
    OnEvent(event ApplicationEvent) error
    SupportsEvent(event ApplicationEvent) bool
}
```

---

## 性能优化

### 缓存策略

| 缓存类型 | 说明 | 效果 |
|---------|------|------|
| Bean 定义缓存 | 缓存 Bean 定义元数据 | 减少重复解析 |
| 切点匹配缓存 | 缓存切点匹配结果 | 提升 AOP 性能 |
| 配置解析缓存 | 缓存配置解析结果 | 加速配置访问 |
| 反射结果缓存 | 缓存反射操作结果 | 减少反射开销 |

### 并发优化

- 读写锁保护共享状态
- 无锁数据结构
- 对象池复用
- 连接池管理

### 内存优化

- 延迟初始化
- 及时释放资源
- 避免内存泄漏
- 预分配切片容量

---

## 最佳实践

### 依赖注入

| 实践 | 说明 |
|------|------|
| 优先构造器注入 | 依赖关系明确，易于测试 |
| 避免循环依赖 | 使用接口解耦 |
| 使用接口 | 优先依赖接口而非具体实现 |

### AOP 使用

| 实践 | 说明 |
|------|------|
| 职责单一 | 每个切面只关注一个横切关注点 |
| 避免过度使用 | AOP 会增加复杂性和性能开销 |
| 注意性能 | 避免在通知中执行耗时操作 |

### 事务管理

| 实践 | 说明 |
|------|------|
| 合理设置边界 | 事务范围应尽量小 |
| 避免长事务 | 长事务会占用数据库连接 |
| 声明式事务 | 优先使用声明式事务 |

### 缓存使用

| 实践 | 说明 |
|------|------|
| 合理设置 TTL | 避免缓存数据过期问题 |
| 避免缓存穿透 | 使用布隆过滤器或空值缓存 |
| 缓存预热 | 应用启动时预热热点数据 |

---

## 附录

### 核心接口速查

| 接口 | 说明 |
|------|------|
| `core.Container` | IoC 容器：Get、RegisterBean、Initialize |
| `aop.Advice` | AOP 通知：Before、After、Around 等 |
| `boot.Application` | 应用实例：Start、Stop、Container |
| `cache.Cache` | 缓存操作：Get、Set、Del、Exists、TTL |
| `config.Config` | 配置访问：Get、GetString、Set、Load、Save |
| `log.Logger` | 日志记录：Debug、Info、Warn、Error |
| `metrics.MeterRegistry` | 指标注册：Counter、Gauge、Histogram |

### 相关文档

- [README.md](README.md) — 项目介绍和快速开始
- [AGENTS.md](AGENTS.md) — AI 智能体开发规范
- [CONTRIBUTING.md](CONTRIBUTING.md) — 贡献指南
- [CODING_STYLE.md](CODING_STYLE.md) — 代码风格指南