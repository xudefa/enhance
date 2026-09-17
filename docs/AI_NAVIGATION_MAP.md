# enhance 代码导航地图（逐包）

> AI 定位代码用。每个包给出：核心接口（在 doc.go）、文件职责、关键实现、依赖、注意事项。
> 与源码一致性以 grep 抽查为准。数据更新日期见文件尾部。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md) · 变更影响：[AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md)

## 模板说明

每个包统一使用以下六段式：

```markdown
### `包名`（一句话职责）
- **核心接口**：`doc.go` 中的接口
- **文件职责**：表格（文件 → 职责）
- **关键实现文件**：核心逻辑所在文件 + 函数
- **依赖**：import 的 enhance 内部包
- **注意事项**：该包的隐蔽约定
```

本文档覆盖六个分组的全部顶层包：

| 分组 | 包 |
|------|----|
| A. 核心框架与配置 | core、boot、context、condition、lifecycle、config |
| B. 数据与网络 | cache、web、mq、resilience、email、metadata |
| C. 可观测性 | log、metrics、actuator、tracing、observability（非代码包） |
| D. 安全与验证 | security、validation、exception、tenant |
| E. 调度与事件 | event、schedule、async、retry、audit |
| F. 工具与扩展 | i18n、spel、testing、webtest、openapi、devtools |

---

## A. 核心框架与配置

### `core`（类型安全 DI 容器，泛型 API 层 + reflect.Type 内部存储）
- **核心接口**：`BeanGet`、`BeanExistenceChecker`、`BeanLister`、`BeanRegister`、`BeanIDGenerator`、`BeanCreator`、`Container`（组合上述小接口）、`ContainerExt`；类型 `BeanOption`（core/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 核心接口门面 + 包文档（设计原则、快速开始） |
| `container.go` | 默认容器实现，`NewContainer() Container` |
| `container_beanid.go` | Bean ID 生成/解析（id = 包路径.类型名#自定义名） |
| `container_validate.go` | 依赖校验、循环依赖检测 |
| `generic_api.go` | 泛型包装 API（`Register[T]`/`Get[T]`/`MustGet` 等，用户层入口） |
| `options.go` | 函数式选项（`WithType`/`WithScope`/`WithLazy`/`WithInit`/`WithDestroy`） |
| `errors.go` | 错误定义 |

- **关键实现文件**：`container.go`（`NewContainer`，container.go:487）、`generic_api.go`（泛型 API）、`container_validate.go`（`validateDependencies`）
- **依赖**：`core/lifecycle`、`core/registry`、`core/scope`（子包）
- **注意事项**：Go 语法限制接口方法不能带类型参数 → `Container` 接口用 `reflect.Type` 存储类型，用户层必须走泛型 API（ADR-001，见 core/doc.go 设计说明）；接口按职责拆小再组合（`Container` = 6 个小接口 + Initialize/Destroy）；注册表/作用域用 `sync.Map` 优化读多写少（ADR-007）。

### `boot`（应用启动器、自动配置、Starter、失败分析）
- **核心接口**：`Application`、`ApplicationContext`、`AutoConfiguration`、`Starter`、`StarterRegistry`、`BootError`、`FailureAnalyzer`、`EventBusResult`；类型别名 `Banner`/`BannerMode`/`TextBanner`/`ASCIIArtBanner`/`LegacyBanner`、`BeanProvider`（boot/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口门面 + 向后兼容别名（`BootErrorStruct`/`StarterRegistryStruct`）+ OrderPriority 执行顺序说明 |
| `application.go` | `NewApplication(opts ...BootOption) (*Boot, error)`，应用入口 |
| `boot.go` | `Boot` 应用启动器，实现 `Application` 接口 |
| `config.go` | `BootConfig` 启动配置结构体 |
| `starter.go` | `starterRegistryImpl`（StarterRegistry 实现，Kahn 拓扑排序的 `GetOrdered`） |
| `boot_starter.go` | `startStarters` 启动所有匹配的 Starter，部分失败逆序停止 |
| `boot_error.go` | `bootError` 实现（错误码常量 + BootError 接口） |
| `autoconfig.go` | `AutoConfigEntry` / `AutoConfigurationOption` / `RegisterAutoConfig` |
| `failure_analyzer.go` | `FailureReport` / `FailureAnalyzerRegistry` / `SimpleFailureAnalyzer` |
| `failure_analyzers.go` + `*_analyzer.go`（4 个） | 具体分析器：BeanNotFound/CircularDependency/ConfigLoad/DuplicateBean/PortInUse |
| `module.go` / `module_builder.go` / `module_compose.go` | `Module` 可组合配置单元（Bean 注册 + Starter） |
| `banner.go` | `DefaultBanner` 默认横幅实例 |
| `configcenter.go` | `ConfigCenterProvider`/`ConfigCenterAdapter` + 工厂注册 |
| `report.go` / `startup_report.go` | 启动报告 |
| `plugin.go` / `plugin_context.go` | 插件扩展点 |
| `boot_start.go` | 配置加载、配置中心、环境注入 |
| `adapter.go` / `boot_helpers.go` | 适配与辅助 |

- **关键实现文件**：`application.go`（`NewApplication`）、`starter.go`（`GetOrdered` 拓扑排序）、`failure_analyzers.go` 系列
- **依赖**：`boot/banner`、`condition`、`config`、`config/environment`、`context`、`core`、`core/registry`、`event`、`lifecycle`
- **注意事项**：`BootError` 由结构体改为接口，保留 `BootErrorStruct` 类型别名兼容（doc.go §向后兼容注释区）；Starter 拓扑排序失败时回退到注册顺序（`GetOrdered` 注释）；自动配置顺序由 `OrderPriority` 枚举控制（-3000 基础设施 → +2000 监控）。

### `context`（应用上下文门面，聚合容器/环境/生命周期/事件）
- **核心接口**：`EventPublisher`、`AsyncEventPublisher`、`EventBusAccess`、`ApplicationContext`（context/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 四接口门面 + Facade 设计模式说明 |
| `context.go` | `DefaultApplicationContext` 默认实现 |
| `builder.go` | `NewApplicationContextBuilder` 构建器 |

- **关键实现文件**：`context.go`（`DefaultApplicationContext`）、`builder.go`
- **依赖**：`config/environment`、`config/refresh`、`core`、`core/registry`、`event`、`lifecycle`
- **注意事项**：`ApplicationContext` 是门面（Facade），聚合各子系统但不重复已有方法——Bean 操作通过 `Container()` 拿 `core.Container`，生命周期通过 `Lifecycle()` 拿 `lifecycle.LifecycleManager`；`AsyncEventPublisher` 由 adapter 适配 `event.AsyncPublisher`（context/doc.go 设计模式节）。

### `condition`（条件化注册，OnProperty/OnBean/OnProfile 等内置条件）
- **核心接口**：`Condition`、`EnvironmentAccessor`、`ContainerAccessor`、`ConditionContext`（condition/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口门面 + `keyLister` |
| `property_condition.go` | `OnProperty`（属性存在且匹配指定值） |
| `missing_property_condition.go` | `OnMissingProperty` |
| `property_or_default_condition.go` | `OnPropertyOrDefault`（带默认值） |
| `property_prefix_condition.go` | `OnPropertyPrefix`（前缀存在） |
| `bean_condition.go` | `OnBean`（Bean 存在） |
| `module_condition.go` | `OnModuleLoaded` / `OnMissingModule` |
| `profile_condition.go` | `OnProfile` |
| `custom_condition.go` | `Custom`（自定义条件） |
| `all_condition.go` / `any_condition.go` / `not_condition.go` | `And` / `Or` / `Not` 逻辑组合 |
| `composite.go` / `described_condition.go` / `func_condition.go` / `extended.go` / `helpers.go` / `builder.go` | 复合/描述/函数式/扩展/辅助 |

- **关键实现文件**：`property_condition.go`（`OnProperty`）、`bean_condition.go`（`OnBean`）、`all_condition.go`（`And`）
- **依赖**：`config/environment`（读属性）、`spel`（见 extended.go）
- **注意事项**：条件通过 `ConditionContext` 统一访问 environment/container，本身不直接依赖容器；`extended.go` 额外依赖 `spel`，其余文件零依赖；`OnModuleLoaded` 是 Go 替代 Java `OnClass` 的方案（doc.go 架构设计）。

### `lifecycle`（生命周期三阶段管理 INIT→RUNNING→STOPPED）
- **核心接口/类型**：`ApplicationPhase`（枚举）、`PhaseListener`、`Hook`、`HookFunc`、`LifecycleManager`、`LifecycleBuilder`、`BeanInitFunc`、`BeanDestroyFunc`、`PhaseTransition`（lifecycle/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口/类型门面（含 `LifecycleManager`/`LifecycleBuilder` 结构体定义） |
| `lifecycle.go` | `NewLifecycleManager` + 阶段转换（SetPhase/AddListener） |
| `hooks.go` | Hook 钩子实现（OnInit/OnStart/OnStop） |

- **关键实现文件**：`lifecycle.go`（`NewLifecycleManager`，lifecycle.go:16）
- **依赖**：无内部依赖（仅标准库）
- **注意事项**：阶段转换必须是正向的 `INIT → RUNNING → STOPPED`，反向转换返回错误；`BeanInitFunc`/`BeanDestroyFunc` 用函数类型而非接口（doc.go 注释，更符合 Go 惯用法）。

### `config`（配置管理：加载/绑定/验证/热更新/配置中心）
- **核心接口**：`Config`、`Validator`、`Loader`、`ConfigCenter`；类型 `LoaderOption`/`LoaderModel`/`ConfigData`/`ConfigCenterConfig`/`WatchEvent`/`WatchCallback`/`WatchManager`/`ValidationError(s)`/`ValidationRule`（config/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口与类型门面（含 `WatchManager`/`memoryConfig` 结构体定义） |
| `config.go` | `NewConfig`（内存 Config 实现） |
| `builder.go` | `NewConfigBuilder` 配置构建器 |
| `loader.go` | Loader 实现 |
| `binder.go` / `binder_validate.go` | 结构体绑定 + 绑定校验 |
| `validator.go` | 配置验证 |
| `center.go` | 配置中心适配（Nacos 风格 Group 等） |
| `watch.go` | 热重载管理 |
| `model.go` / `properties.go` | 数据模型 |

- **关键实现文件**：`config.go`（`NewConfig`，config.go:21）、`builder.go`、`loader.go`
- **依赖**：`config/environment`；相关子包 `config/environment`（分层 PropertySource + Profile）、`config/refresh`（刷新作用域）
- **注意事项**：配置加载优先级 命令行 > 环境变量 > 配置文件 > 默认值（doc.go）；热重载事件三类型 `modify`/`delete`/`create`（`EventModify`/`EventDelete`/`EventCreate` 常量）；`Loader` 的 `Priority()` 越大越优先。

---

## B. 数据与网络

### `cache`（缓存抽象：LRU / 分片 LRU / TTL）
- **核心接口**：`Cache`、`CacheInspector`、`Clearable`；函数类型 `Getter`（cache/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口门面 + Getter 定义 |
| `lru.go` | `NewLRUCache` LRU 实现 |
| `lru_sharded.go` | 分片 LRU（高并发场景） |
| `ttl_cache.go` | 基于 LRU + TTL 的缓存实现 |
| `builder.go` | `NewMemoryCacheBuilder` 链式构建 |
| `errors.go` | 错误定义（`ErrNotFound` 等） |

- **关键实现文件**：`lru.go`（`NewLRUCache`，lru.go:61）、`ttl_cache.go`、`builder.go`
- **依赖**：无内部依赖
- **注意事项**：`Get` 未命中返回 `ErrNotFound`（doc.go Cache 注释）；`Set` 的 `ttl<=0` 表示永不过期；`Getter` 支持缓存穿透保护（返回 nil 值不会被缓存）。

### `web`（Web 分层面板，doc.go 全部为类型别名）
- **核心接口**：全部为类型别名——`Context`/`Router`/`Server`/`Controller`/`HandlerFunc`/`MiddlewareFunc` = `web/core.*`；`EngineFactory`/`EngineType`/`EngineRegistry`/`ServerOption`/`ServerConfig` = `web/engine.*`；`WebStarter`/`WebConfig` = `web/mvc.*`（web/doc.go）；`types.go` 提供 `EngineStdLib` 常量与 `GlobalEngineRegistry` 等辅助
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 全部类型别名（门面） |
| `types.go` | 引擎常量、全局引擎注册表、辅助函数（`NewEngineRegistry`/`DefaultServerConfig`/`WithHost`） |
| `web/core/` | 核心接口定义（doc.go）+ `route_registry.go` 路由注册 |
| `web/engine/` | 引擎注册表/工厂/适配器（`registry.go`/`adapter.go`）+ `stdlib/` 标准库 net/http 实现 |
| `web/mvc/` | MVC 层：`starter.go`（WebStarter）、`controller.go`、`config.go`、`websocket.go` |
| `web/server/` | net/http 服务端实现：`http_server.go`/`router.go`/`context.go`/`response.go`/`binder.go`/`client.go`/`retry.go`/`circuit_breaker.go`/`tls.go` 等 |
| `web/middleware/` | 中间件（RequestID/AccessLog/Error/CORS，middleware.go） |
| `web/binding/` | 参数绑定（binder.go） |
| `web/tls/` | TLS 支持 |
| `web/annotation/`、`web/filter/`、`web/interceptor/`、`web/response/` | 空目录（截至核对日期无 Go 文件） |

- **关键实现文件**：`web/engine/stdlib/engine.go`（默认引擎）、`web/server/http_server.go`、`web/mvc/starter.go`
- **依赖**：`web/core`、`web/engine`、`web/mvc`、`web/server`（子包）
- **注意事项**：doc.go 全部为类型别名、无实现——改接口先改 `web/core`，再同步 engine/mvc/server 实现；默认引擎 `web/engine.StdLib`；扩展新引擎按 doc.go 扩展指南实现 `engine.Factory` + 子接口并注册到 `engine.GlobalRegistry`。

### `mq`（消息队列抽象，内存队列 + 对象池）
- **核心接口**：`Queue`；类型 `Message`（Ack/Nack 并发安全）、`MessageHandler`、`QueueOption`、`BaseQueue`（mq/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | Queue 接口 + Message/BaseQueue 类型门面 |
| `mq.go` | 内存队列 `InMemoryQueue`（`NewInMemoryQueue`）+ `Message` 对象池（`AcquireMessage`/`ReleaseMessage`） |
| `mq_helpers.go` | 辅助函数 |

- **关键实现文件**：`mq.go`（`NewInMemoryQueue`，mq.go:151；`AcquireMessage`，mq.go:40）
- **依赖**：无内部依赖
- **注意事项**：`Message` 用 `sync.Pool` 复用减少 GC 压力；`acknowledged` 用 `atomic.Int32` 保证 Ack/Nack 并发安全；集成后端在 `starter/kafka`、`starter/rabbitmq`。

### `resilience`（弹性容错：熔断 + 负载均衡 + 注册中心）
- **核心接口**：`Breaker`、`Selector`、`Registry`、`Balancer`；类型 `State`（Closed/Open/HalfOpen）、`InstanceInfo`/`ServiceInstance`/`HealthStatus`、`BreakerOption`（resilience/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口与类型门面 + 默认常量 + 错误哨兵 |
| `breaker.go` | `NewBreaker` 熔断器实现 |
| `selector.go` | `roundRobinSelectorImpl`（Selector 的轮询实现） |
| `roundrobin.go` / `random.go` / `leastconn.go` / `responsetime.go` / `hash.go` / `sticky.go` | 各负载均衡策略（轮询/随机/最少连接/响应时间/一致性哈希/会话保持） |
| `center.go` | `InMemoryRegistry` 内存服务注册中心 |
| `health.go` | `NewHealthAware` 健康感知 |
| `adaptive.go` | 自适应限流 |
| `builder.go` | `NewRegistryBuilder` |
| `types.go` | 状态字符串实现 |

- **关键实现文件**：`breaker.go`（`NewBreaker`）、`selector.go`、`roundrobin.go`
- **依赖**：无内部依赖
- **注意事项**：熔断三状态 CLOSED/OPEN/HALF_OPEN + 默认常量（`DefaultErrorThreshold`=0.5、`DefaultWaitDuration`=30s）；两套负载均衡并存——`Selector.Select([]InstanceInfo)` 与 `Balancer.Next([]*ServiceInstance)`，入参结构不同，注意区分；错误用哨兵变量 `ErrCircuitOpen`/`ErrNoInstances` 等。

### `email`（SMTP 邮件发送）
- **核心接口**：`Sender`；类型 `Message`/`Attachment`/`SenderOption`（email/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | Sender 接口 + Message/Attachment + 配置键 |
| `sender.go` | `NewSender` SMTP 实现（基于 net/smtp） |

- **关键实现文件**：`sender.go`（`NewSender`，sender.go:63）
- **依赖**：无内部依赖（net/smtp 标准库）
- **注意事项**：默认 SMTP `localhost:25`；密码从环境变量 `ENHANCE_EMAIL_PASSWORD` 读取；发件人地址必须显式设置（`DefaultFrom` 为空、不自动推断）。

### `metadata`（配置元数据生成 + struct tag 注解解析）
- **核心接口**：`MetadataGenerator`、`PropertyIndex`、`TagAnnotationResolver`；类型 `Annotation`、`PropertyMetadata`/`GroupMetadata`/`HintMetadata`、`ConfigurationMetadata`（metadata/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 元数据类型门面 |
| `metadata.go` | `NewMetadataGenerator` / `NewPropertyIndex` / `GenerateFromStruct` 元数据生成与索引 |
| `resolver.go` | `NewTagAnnotationResolver` 解析 `metadata:"name:attr=val"` tag |

- **关键实现文件**：`metadata.go`（`GenerateFromStruct`，metadata.go:280；`NewPropertyIndex`，metadata.go:223）、`resolver.go`
- **依赖**：无内部依赖
- **注意事项**：`GenerateFromStruct` 产出 Spring Boot 风格 `ConfigurationMetadata`（Groups/Properties/Hints），可 `ToJSON()` 序列化；`TagAnnotationResolver` 支持多属性类型自动转换。

---

## C. 可观测性

### `log`（日志抽象 + slog 实现，多小接口组合）
- **核心接口**：`Logger`、`LoggerWithSync`、`LoggerWithFields`、`LoggerFatal`、`LoggerWithLevel`、`LoggerLevelChecker`、`LoggerWithName`、`LoggerWithCaller`、`LoggerWithTimeout`、`Sampler`；类型 `Level`（7 级）/`KeyValue`/`RandomSampler`/`ThresholdSampler`/`SampledLogger`/`LoggerOption`（log/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 采样器门面 |
| `logger.go` | `Build()` 默认 Logger + Level 字符串 |
| `slog.go` | slog 后端实现（`WithLevel` 等） |
| `logger_builder.go` | `LoggerBuilder` 链式构建 |
| `builder.go` | 采样器实现 |
| `context.go` | 上下文日志（`TraceContextKey`、`FromContext`） |

- **关键实现文件**：`logger.go`（`Build`，logger.go:35）、`slog.go`、`logger_builder.go`
- **依赖**：无内部依赖
- **注意事项**：采用小接口组合而非一个大接口——`Logger` 基座仅 Debug/Info/Warn/Error，其余能力（Sync/Fields/Fatal/Level/Caller/Timeout/Name）各自独立接口，实现方可按需提供；zap/zerolog 集成在 `starter/` 子包；采样器 `Sampler` + `SampledLogger` 配合使用。

### `metrics`（指标收集：Counter/Gauge/Histogram + 导出器）
- **核心接口**：`Counter`、`Gauge`、`Histogram`、`Exporter`、`MeterRegistry`；快照结构 `Metric`（metrics/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + Metric 快照门面 |
| `registry.go` | `NewSimpleRegistry` 简单内存注册表 |
| `registry_helpers.go` | 直方图快照收集辅助 |
| `prometheus_exporter.go` | Prometheus 导出器 |
| `autoconfig.go` | 自动配置（注册 SimpleRegistry 为单例 Bean） |
| `builder.go` | 构建器 |

- **关键实现文件**：`registry.go`（`NewSimpleRegistry`，registry.go:220）、`prometheus_exporter.go`
- **依赖**：`boot`、`condition`、`config/environment`、`core`
- **注意事项**：`Metric` 是采集用的快照结构（Name/Value/Tags/Type/Timestamp/Count/Sum）；开关配置键 `MetricsEnabled = "metrics.enabled"`；Prometheus/OpenTelemetry 集成在 `starter/` 子包。

### `actuator`（运维端点：health/info/metrics/env/beans/admin）
- **核心接口**：`AppContext`、`RouteRegistrar`、`SanitizeStrategy`；常量（配置键/默认值）（actuator/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 三接口 + 配置键常量 |
| `actuator.go` | `New(ctx) *Actuator` 管理器 + `NewDatabaseHealthIndicator`/`NewRedisHealthIndicator` |
| `autoconfig.go` | 自动配置 |
| `health.go` | 健康检查 |
| `disk_space_health_indicator.go` / `memory_health_indicator.go` / `process_health_indicator.go` | 各健康指标实现 |
| `router.go` | 路由注册 |
| `endpoint_registry.go` / `http_endpoint_registry_adapter.go` / `std_http_handler_registry.go` | 端点注册与适配 |
| `sanitize.go` | 敏感信息过滤（SanitizeStrategy） |
| `http_starter.go` | HTTP Starter |
| `path_normalizer.go` | 路径规范化 |

- **关键实现文件**：`actuator.go`（`New`，actuator.go:30）、`autoconfig.go`
- **依赖**：`actuator/admin`、`actuator/health`、`boot`、`condition`、`config/environment`、`core`、`event`、`metrics`
- **注意事项**：通过空导入 `import _ "github.com/xudefa/enhance/actuator"` 自动启用；端点前缀默认 `/actuator`；`SanitizeStrategy` 防止 /env 等端点泄露敏感配置。

### `tracing`（分布式追踪：Span/Tracer/采样/导出/上下文传播）
- **核心接口**：`Sampler`、`Exporter`；类型 `Tracer`（struct）/`Span`/`SpanContext`/`TraceID`/`SpanID`/`SpanStatus`/`SpanEvent`/`SpanOption`/`TracerOption`（tracing/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 类型门面 + HTTP 头常量 |
| `tracer.go` | `NewTracer` 追踪器实现 |
| `span.go` | Span 实现 |
| `trace_helper.go` | `TraceHelper` 助手 |
| `autoconfig.go` | `TracingAutoConfiguration` 自动配置 |
| `errors.go` | 错误定义 |

- **关键实现文件**：`tracer.go`（`NewTracer`，tracer.go:52）、`span.go`
- **依赖**：`boot`、`condition`、`config/environment`、`core`、`log`
- **注意事项**：HTTP 头使用 Go `http.CanonicalHeaderKey` 格式——`X-Trace-Id` 而非 `X-Trace-ID`（doc.go 显式标注）；内置采样器 `AlwaysOnSampler`/`AlwaysOffSampler`/`ProbabilitySampler`；Span 数量上限防内存溢出（`DefaultMaxSpans`=10000）；ID 用 `crypto/rand` 生成。

### `observability`（非代码包）
- **核心接口**：无（非代码包，无顶层 Go 文件）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `README.md` | 包说明：可观测性规划设计（Actuator + Micrometer 理念、metrics/logging 子模块） |
| `metrics/` | 观测指标落点（doc.go + metrics.go） |

- **关键实现文件**：无
- **依赖**：无
- **注意事项**：**非代码包**——只含 README + `metrics/` 子目录，无顶层 Go 源文件；实际日志/指标能力在 `log`、`metrics` 包。

---

## D. 安全与验证

### `security`（安全框架，接口数量最多的包）
- **核心接口**：`SecurityContext`、`AccessDeniedHandler`、`AuthenticationEntryPoint`、`SecurityMetadataSource`、`SecurityRequest`、`SecurityResponse`、`GrantedAuthority`、`Configurer`、`Builder`、`HttpSecurity`（= Configurer + Builder）、`AuthorizeRequests`、`ExpressionInterceptUrlRegistry`、`SecurityConfigurer`（已废弃）、`LogoutHandler`、`LogoutSuccessHandler`、`RateLimiter`、`RateLimitStrategy`、`CsrfTokenRepository` + 子包类型别名（Authentication/AuthenticationManager/UserDetails/PasswordEncoder/AccessDecisionManager/SecurityFilter/SecurityFilterChain 等）（security/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口门面（约 20 个）+ 子包类型别名 + 错误哨兵 |
| `security.go` | 安全框架设计说明与辅助（authenticate/hasPermission 风格） |
| `security_builder.go` | `NewSecurityBuilder` 构建器 |
| `config.go` | `NewSecurityConfig` 函数式选项模式（推荐用法） |
| `http_security.go` | `HttpSecurity` 链式实现（`authorizeRule` 元数据源） |
| `authentication.go` | 认证令牌（`NewUsernamePasswordAuthenticationToken`） |
| `user_details.go` | `UserDetails` 用户详情 |
| `password_encoder.go` | 密码编码器（明文仅限开发/测试） |
| `authorization.go` / `casbin_voter.go` | 授权与 Casbin 投票器 |
| `filter_chain.go` / `filter.go` / `config_filters.go` / `http_adapter.go` | 过滤器链 + 规则 + HTTP 适配 |
| `auth_filters.go` / `logout.go` / `cors_filter.go` / `csrf.go` / `rate_limit*.go` | 各安全过滤器 |
| `autoconfig.go` | 自动配置（security.enabled=true 时装配） |

- **关键实现文件**：`config.go`（`NewSecurityConfig`）、`http_security.go`（HttpSecurity 实现）、`rate_limit.go`
- **依赖**：`boot`、`condition`、`config/environment`、`core`、`log`、`security/authentication`、`security/authorization`、`security/filter`
- **注意事项**：接口数量全库最多（20+），配置类接口按 **authentication / authorization / filter** 三个子包分组，根包用类型别名保持包级访问；`HttpSecurity` 是 `Configurer`（13 个设置方法）+ `Builder`（`AuthorizeRequests`+`Build`）组合的 16 方法链式大接口；`SecurityConfigurer` 已废弃，改用 `NewSecurityConfig` 函数式选项；4 个错误哨兵 `ErrAuthenticationFailed`/`ErrAccessDenied`/`ErrUserNotFound`/`ErrBadCredentials`。

### `validation`（参数校验：tag 驱动 + 跨字段 + HTTP 中间件）
- **核心接口**：`Validator`、`CustomValidator`、`MiddlewareValidator`、`ResponseWriter`、`Binder`；类型 `ValidationError(s)`/`ValidatorRegistry`/`RuleBuilder`/`TagValidator`/`MiddlewareConfig`/`ErrorResponse`（validation/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 类型门面 |
| `validation.go` | `NewTagValidator` / `NewTagValidatorWithRegistry` + 错误描述 |
| `validate.go` | `Validate` 单值规则验证 |
| `rules.go` | 内置规则（Required/Min/Max/Email/URL/Regex/In/NotIn…） |
| `crossfield.go` | 跨字段校验（如密码确认） |
| `groups.go` | 验证组 `GroupedTagValidator` |
| `registry.go` | `NewValidatorRegistry` 自定义验证器注册表（sync.Map） |
| `builder.go` | `NewRuleBuilder` / `NewValidatorChain` / `NewRegexCache` |
| `middleware.go` | `NewValidateMiddleware` HTTP 集成 |
| `request_validation.go` | `NewRequestValidator` |
| `binding.go` | `NewDefaultBinder`/`NewJSONBinder`/`NewFormBinder`/`NewQueryBinder` |
| `unsafe.go` / `pool.go` | 辅助 |

- **关键实现文件**：`validation.go`（`NewTagValidator`）、`rules.go`、`registry.go`（`NewValidatorRegistry`）
- **依赖**：无内部依赖
- **注意事项**：struct tag `validate:"required,email"` 驱动字段级校验；跨字段规则在 `crossfield.go`；`ValidatorRegistry` 用 `sync.Map` 优化读多写少；一次收集全部校验错误后返回（非短路）。

### `exception`（全局异常处理 + 统一错误码）
- **核心接口**：`Logger`、`MetricsRecorder`、`ResponseWriter`、`ExceptionHandler`、`ExceptionResolver`；类型 `ErrorResponse`/`ErrorCode`/`ErrorCodeRegistry`/`ExceptionHandlerConfig`/`KeyValue`（exception/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 类型门面 |
| `handler.go` | `NewDefaultExceptionHandler` 核心处理器（opts 选项） |
| `resolver.go` | `NewResolverChain` 解析器链（按 Order 排序尝试） |
| `builtin.go` | `NewDefaultExceptionResolver` / `NewBuiltinExceptionResolver` 兜底解析器 |
| `errorcode.go` | `ErrorCodeExceptionResolver` + `NewErrorCodeRegistry` + `BusinessError` |
| `response.go` | `NewErrorResponse` 统一错误响应 |
| `middleware.go` | `ErrorMiddleware` HTTP 中间件 |
| `builder.go` | `NewExceptionHandlerBuilder` / `NewErrorResponseBuilder` |
| `metrics_recorder.go` | `NewDefaultMetricsRecorder` |
| `security_adapter.go` | 适配 security（AccessDeniedHandler/AuthenticationEntryPoint） |

- **关键实现文件**：`handler.go`（`NewDefaultExceptionHandler`，handler.go:45）、`resolver.go`（`NewResolverChain`）
- **依赖**：无内部依赖
- **注意事项**：统一错误响应格式 `{code, message, requestId, traceId, details, timestamp}`；多个解析器按 `Order()` 从小到大依次尝试、`Supports()` 命中即用；`ErrorCode{HTTP状态码, 用户消息, 调试信息}` 三分量；安全过滤敏感信息防止泄露。

### `tenant`（多租户：解析/上下文/隔离）
- **核心接口**：`TenantResolver`、`TenantManager`、`TenantMiddleware`、`TenantIsolation`、`TenantRegistry`、`TenantProvider`；类型 `Tenant`（tenant/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + Tenant 结构门面 |
| `tenant.go` | `NewHeaderResolver` / `NewTenantManager` / `NewTenantMiddleware` 实现 |
| `tenant_helpers.go` | 辅助函数（`splitPath`） |

- **关键实现文件**：`tenant.go`（`NewTenantManager`，tenant.go:105）
- **依赖**：无内部依赖
- **注意事项**：`SetCurrentTenant`/`GetCurrentTenant` 是进程级共享状态，并发请求会互相覆盖——多请求场景必须用 context 传递（`SetTenantToContext` / `TenantFromContext`，doc.go 有显式警告）；解析策略支持请求头/子域名/JWT/路径。

---

## E. 调度与事件

### `event`（事件驱动：发布/订阅/死信/事务事件）
- **核心接口**：`ApplicationEvent`、`EventBus`、`AsyncPublisherBus`；类型 `EventListener`/`ListenerConfig`/`BaseEvent`/`EventBusWithOrdering`（struct）/`LegacyEventBusAdapter` + 内置生命周期事件常量（event/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 类型门面 + 5 个内置事件常量（EnvironmentPrepared/ContextRefreshed/ApplicationStarted/ApplicationReady/ApplicationStopped） |
| `event.go` | `eventBus` 实现（sync.Map 无锁读 + CAS 订阅）+ `NewEventBus` |
| `listener.go` | `NewEventBusWithOrdering`（优先级/过滤/异步）+ `LegacyEventBusAdapter` |
| `builder.go` | `NewEventBusBuilder` 构建器 |
| `deadletter.go` | 死信配置 + `RetryPolicy`（委托 retry 包，向后兼容） |
| `deadletter_bus.go` | `NewEventBusWithDeadLetter` 死信队列总线 |
| `transactional.go` | `TransactionalEvent` 事务事件包装器 |
| `async.go` | 异步发布器 |

- **关键实现文件**：`event.go`（`NewEventBus`，event.go:10）、`listener.go`、`deadletter_bus.go`
- **依赖**：`retry`（deadletter.go 委托）
- **注意事项**：事件用**字符串 `Type()` 路由**而非反射类型（ADR-009，doc.go 显式设计原则）；`EventBus` 现为接口（doc.go:187），默认实现是未导出的 `eventBus` 结构体，`EventBusWithOrdering` 是导出结构体；监听器存储用 `sync.Map` + CAS 无锁更新（ADR-007）；`listenerSlice` 指针包装使 slice 可用于 atomic 的 CompareAndSwap。

### `schedule`（定时任务调度：Cron / 固定延迟 / 固定频率）
- **核心接口**：`Task`、`Scheduler`；类型 `SchedulerOption`（schedule/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 配置键常量 |
| `task.go` | `NewTask` 函数式任务（`functionTask` 实现） |
| `scheduler.go` | 默认 Scheduler 实现 + `SchedulerOption` |
| `builder.go` | `SchedulerBuilder` 链式构建 |
| `autoconfig.go` | `ScheduleAutoConfiguration` 自动配置 |

- **关键实现文件**：`task.go`（`NewTask`，task.go:28）、`scheduler.go`、`builder.go`
- **依赖**：`boot`、`condition`、`config/environment`、`core`、`log`
- **注意事项**：Cron 为 **6 字段 Spring 风格（秒起）**；`Register` 任务名唯一，重复注册返回 error；配置键 `ScheduleEnabled`/`SchedulePoolSize`/`ScheduleScanAnnotations`。

### `async`（异步执行器：goroutine 池 + Future）
- **核心接口/类型**：`Future`（struct）、`ExecutorOption`、`RejectHandler`（async/doc.go）；核心执行器 `AsyncExecutor` 结构体定义在 `executor.go`
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | Future/ExecutorOption/RejectHandler 类型门面 |
| `executor.go` | `AsyncExecutor` 结构体 + `NewAsyncExecutor`/`NewFuture`/`Submit`/`SubmitVoid`/`Shutdown`/`ShutdownWithTimeout` + Future 的 `Get`/`GetWithContext`/`GetWithTimeout`/`IsDone` |

- **关键实现文件**：`executor.go`（`NewAsyncExecutor`，executor.go:93；`Future.Get` 系列）
- **依赖**：无内部依赖
- **注意事项**：核心是 `AsyncExecutor` **结构体非接口**（doc.go 只放辅助类型）；懒启动模式——首次 `Submit` 才启动 worker；任务 panic 会被 `recover` 并写入 Future 错误；`Shutdown` 用 `sync.Once` 保证幂等。

### `retry`（独立重试机制：策略 + 退避 + jitter 防惊群）
- **核心接口/类型**：`doc.go` 无 `^type` 定义（仅包注释）；核心类型在 `retry.go`：`RetryPolicy`（struct）、`BackoffStrategy`（None/Fixed/Linear/Exponential）、`Executor`（struct）、`RetryableFunc[T]`、`RetryInfo`、`OnRetryFunc`、`ExecutorOption`
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 包文档 + 使用示例 |
| `retry.go` | 策略构造器（`NoRetry`/`FixedDelay`/`LinearBackoff`/`ExponentialBackoff`）+ `NewExecutor`/`MustNewExecutor` + 泛型 `Execute[T]` + `CalculateDelay`（jitter + 上限） |

- **关键实现文件**：`retry.go`（`NewExecutor`，retry.go:179；`Execute[T]`，retry.go:225）
- **依赖**：无内部依赖
- **注意事项**：从 `event/deadletter.go` 和 `web/server/retry.go` 抽取的独立包（包注释）；`CalculateDelay` 内置 jitter（Fixed 默认 10%、Exponential 默认 20%），延迟先加抖动再截断 `MaxDelay`（含 24h 硬顶）；幂等重试通过泛型 `Execute[T]`（`RetryableFunc[T]` 签名），另有 `ExecuteVoid` 方法。

### `audit`（审计日志：事件驱动记录 + 同步/异步双模式）
- **核心接口**：`EventWriter`、`Auditor`、`AuditInterceptor`、`AuditLogger`；类型 `Event`、`EventType`/`EventSeverity`（枚举）、`AuditorOption`（audit/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + Event/枚举门面 + 内置事件类型（CREATE/UPDATE/DELETE/READ/LOGIN/SECURITY…）+ 错误哨兵 |
| `auditor.go` | `NewAuditor` 审计器实现（`WithWriter`/`WithAsync`/`WithBufferSize`，同步/异步 + 后台消费） |
| `writer.go` | `NewConsoleWriter` / `NewFileWriter` 两类 EventWriter（JSON 序列化 + 缓冲写入） |
| `logger.go` | `NewAuditLogger` 审计日志助手（预设 actor/source，Create/Update/Delete/Login/Severity 便捷方法） |
| `interceptor.go` | `NewAuditInterceptor` 方法调用拦截（出错标记 failure 事件） |

- **关键实现文件**：`auditor.go`（`NewAuditor`，auditor.go:36）、`writer.go`（`NewFileWriter`，writer.go:56）
- **依赖**：无内部依赖（仅标准库）
- **注意事项**：`Auditor` 默认控制台写入器 + 同步模式，缓冲区默认 1000（auditor.go 注释）；异步模式下事件先入缓冲通道，`Close` 时会 `drainEvents` 尽量消费剩余事件；`WithWriter` 等选项通过类型断言 `a.(*auditorImpl)` 下发，未导出实现；两个错误哨兵 `ErrWriterClosed`/`ErrChannelFull`，通道满时回退同步直写。

---

## F. 工具与扩展

### `i18n`（国际化：资源包消息源 + 回退）
- **核心接口**：`MessageSource`；类型 `Locale`/`DefaultLocale`（i18n/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | MessageSource 接口 + Locale + `formatMessage`/`escapePercent` 格式化逻辑 |
| `message.go` | `NewResourceBundleMessageSource` 资源包实现 |

- **关键实现文件**：`message.go`（`NewResourceBundleMessageSource`，message.go:36）
- **依赖**：无内部依赖
- **注意事项**：消息查找回退顺序 精确区域（zh_CN）→ 语言（zh）→ fallback 父源 → 返回 code 本身；`escapePercent` 转义非格式动词的 `%`，避免 "折扣 50%" 这类字面量被 `fmt.Sprintf` 破坏（doc.go 显式说明）。

### `spel`（SpEL 表达式语言）
- **核心接口**：`Expression`、`ExpressionParser`、`EvaluationContext`、`PropertyAccessor`、`MethodInterceptor`、`MethodInvocation`（spel/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 6 个接口门面 |
| `impl.go` | 构造函数门面：`NewSpelParser`（impl.go:64）/`NewStandardEvaluationContext`（impl.go:50）/`NewReflectPropertyAccessor`/`NewInterceptorChain`/`NewSimpleMethodInvocation`/`NewLoggingInterceptor` |
| `context.go` | 求值上下文与属性访问器实现（`standardEvaluationContextImpl`/`reflectPropertyAccessorImpl`） |
| `evaluator.go` / `evaluator_expressions.go` | 解析器与表达式求值实现（`spelParserImpl.ParseExpression`、`propertyExpressionImpl`） |
| `interceptor.go` | 拦截器链实现（`interceptorChainImpl.Proceed` 等） |
| `type_utils.go` | 类型工具 |

- **关键实现文件**：`impl.go`（`NewSpelParser`，impl.go:64）、`evaluator.go`、`context.go`
- **依赖**：无内部依赖
- **注意事项**：`PropertyAccessor` 基于反射提供属性读写；`EvaluationContext` 支持命名变量（`#role` 语法）；`MethodInterceptor` 通过 `Proceed()` 支持拦截器链式调用。

### `testing`（测试工具：断言 + Mock + TestRunner）
- **核心接口**：`TestingT`、`TestContext`、`Mock`；类型 `TestRunner`（testing.go）、`ExpectationRequest`（testing/doc.go + testing.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口门面 |
| `assertions.go` | 断言函数：`Assert`/`AssertEqual`/`AssertNil`/`AssertTrue`/`AssertError`/`AssertExpectations` 等 |
| `mock.go` | Mock 实现（`Expect`/`Call`/`Verify`/`Reset` + `AssertExpectations`） |
| `testing.go` | `NewTestRunner` 测试运行器（`Run(fn func(TestContext))`，基于 boot 启动应用） |

- **关键实现文件**：`assertions.go`（`AssertEqual`）、`mock.go`、`testing.go`（`NewTestRunner`）
- **依赖**：`boot`、`config/environment`、`core`、`core/registry`
- **注意事项**：`TestingT` 兼容标准库 `testing.T`（Errorf/Fatalf/Helper 三方法）；`Mock.Expect` 默认期望调用 1 次，`ExpectTimes` 指定次数；包名与标准库 `testing` 相同，导入时注意区分；`TestRunner` 内部创建真实 boot 应用。

### `webtest`（Web 测试：测试服务器 + 请求构建器 + 响应断言）
- **核心接口**：`TestServer`、`RequestBuilder`、`ResponseVerifier`；doc.go 包注释另列 `TestClient`（webtest/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 三接口门面 |
| `client.go` | `NewWebTestClient(handler http.Handler) *WebTestClient` 实现 |

- **关键实现文件**：`client.go`（`NewWebTestClient`，client.go:20）
- **依赖**：无内部依赖
- **注意事项**：基于 `http.Handler` 构造，无需真实端口；响应断言支持 `AssertStatus`/`AssertHeader`/`AssertBody`/`AssertBodyContains`（`AssertJSONPath` 见 doc.go 使用示例，实际能力以 `client.go` 为准）。

### `openapi`（OpenAPI 3.0 文档生成 + Swagger UI）
- **核心接口/类型**：`doc.go` 无 `^type` 定义（仅包文档）；核心类型在 `openapi.go`（`OpenAPIDocument`）、`builder.go`、`schemas.go`（`NewDocument()` → `DocumentBuilder`）、`paths.go`、`swagger.go`
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 包文档（注解约定、配置属性） |
| `openapi.go` | `OpenAPIDocument` 文档结构 + `InfoObject` |
| `builder.go` | 文档构建（`AddPath`） |
| `schemas.go` | `NewDocument` → `DocumentBuilder` + `ComponentsObject` |
| `paths.go` | `OperationObject` 路径/操作定义 |
| `swagger.go` | Swagger UI HTML 集成（swaggerUIHTML） |

- **关键实现文件**：`schemas.go`（`NewDocument`，schemas.go:78）、`openapi.go`
- **依赖**：无内部依赖
- **注意事项**：从控制器注解（`@Operation`/`@Parameter`/`@Response`）自动生成 OpenAPI 3.0，重开 `/swagger/index.html`；空导入 `import _ "github.com/xudefa/enhance/openapi"` 启用；配置键 `openapi.enabled` 等。

### `devtools`（开发工具：热重载 + 文件监控 + 开发模式检测）
- **核心接口**：`HotReloader`、`HotReloaderInfo`（组合 HotReloader + IsRunning/GetWatchedFiles/GetWatchDirs/Restart）、`FileWatcher`、`DevModeDetector`；类型 `ReloadEvent`、`ReloadType`（CREATED/MODIFIED/DELETED）、`ReloadCallback`、`HotReloaderOption`（devtools/doc.go）
- **文件职责**：

| 文件 | 职责 |
|------|------|
| `doc.go` | 接口 + 类型门面 + 环境变量启用说明（`ENHANCE_DEV_MODE=true`） |
| `devtools.go` | `NewHotReloader` 热重载实现（MD5 哈希轮询 + `WithWatchDirs`/`WithExtensions`/`WithInterval`/`WithIgnoreDirs`） |
| `devtools_handlers.go` | `NewDevModeDetector` / `NewFileWatcher` 实现 + `NewLiveReloadServer` 实时重载服务器 |

- **关键实现文件**：`devtools.go`（`NewHotReloader`，devtools.go:90）、`devtools_handlers.go`（`NewFileWatcher`，devtools_handlers.go:55）
- **依赖**：无内部依赖（仅标准库）
- **注意事项**：`NewHotReloader` 返回 `HotReloaderInfo`（非 `HotReloader`），内置默认忽略 `.git`/`node_modules`/`vendor`，默认轮询间隔 2s；文件变更通过 MD5 哈希快照对比而非事件监听（`computeFileHash`）；`NewLiveReloadServer` 要求传入的 reloader 必须是 `*hotReloaderImpl`（类型断言，否则报错）；开发模式检测读取 `DEV_MODE`/`DEVELOPMENT`/`GO_ENV` 环境变量。

---

## 一致性抽查记录

抽查日期 2026-09-16，覆盖 core/cache/boot/event/security/web/openapi/retry 八个包（接口名 grep + 文件名 test -f 断言），全部通过：

```bash
grep -n "type Container interface" core/doc.go          # core.doc.go:271 ✓
grep -n "type Cache interface" cache/doc.go             # cache.doc.go:62 ✓
grep -n "type Starter interface" boot/doc.go            # boot.doc.go:231 ✓
grep -n "type ApplicationEvent interface" event/doc.go  # event.doc.go:86 ✓
grep -n "type HttpSecurity interface" security/doc.go   # security.doc.go:284 ✓
grep -n "type Context = core.Context" web/doc.go        # web.doc.go:74 ✓
grep -n "type OpenAPIDocument struct" openapi/openapi.go # openapi.go:4 ✓
grep -n "type RetryPolicy struct" retry/retry.go        # retry.go:30 ✓
test -f core/generic_api.go && test -f cache/lru.go && test -f boot/failure_analyzers.go \
  && test -f event/deadletter.go && test -f security/rate_limit.go && test -f web/types.go \
  && test -f openapi/builder.go && test -f retry/retry.go   # 文件名断言全通过 ✓
```

---

> 数据核对日期：2026-09-16。源码更新后请更新本文件对应条目。
