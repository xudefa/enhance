# enhance 设计决策与坑点（AI 避坑指南）

> 遇到诡异行为、历史决策、隐蔽约定时查这里。每条注明出处（源码注释 / ADR-XXX）。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md) · 代码导航：[AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md) · 关联：[ADR.md](ADR.md)

## 1. 全局硬约束（违反即返工）

- 核心包零外部依赖（ADR-004）；第三方集成必须走 `starter/` 且各自独立 go.mod（ADR-004 后果，AGENTS.md §0.1）
- 依赖方向单向，禁止循环依赖；低优先级包不能依赖高优先级包（AGENTS.md §0.4）；starter 包之间禁止互相依赖（AGENTS.md §0.1）
- 构造函数返回接口（ADR-006）；接口小而精（≤5 方法），大接口拆小接口组合（AGENTS.md §0.4）
- 事件用字符串路由而非反射类型，事件类型字符串全局一致（ADR-009）
- 优先级用 `OrderPriority` 枚举，禁止魔法数字（ADR-010）
- 条件化自动配置用函数而非注解（ADR-008）；放弃 AOP，用显式拦截器（ADR-003）
- 错误不丢上下文：`fmt.Errorf("...: %w", err)`，判断用 errors.Is/As，禁止 `==` 直接比较（AGENTS.md §6.1）
- 控制流用早期返回而非深层嵌套（ADR-011）
- 泛型 API 是用户接口，Container 内部 reflect.Type 是实现细节（ADR-001，core/doc.go "设计说明"）
- 接口定义集中在 `doc.go`，实现逻辑禁止入内（ADR-005）

## 2. 逐包坑点

> 表内「出处」即验证依据：`doc.go` 指该包 doc.go 注释/定义，`xxx.go` 指具体实现文件，`ADR-XX` 见 [ADR.md](ADR.md)。

| 包 | 坑点（出处） |
|----|-------------|
| core | Container 接口方法用 `reflect.Type`，无泛型方法（Go 语法限制）；用户必须走 `Register[T]`/`Get[T]` 泛型 API，不要直接调接口的反射方法（core/doc.go "设计说明"）；接口按职责拆为 BeanGet/BeanExistenceChecker/BeanLister/BeanRegister/BeanIDGenerator/BeanCreator 六个小接口组合成 Container（core/doc.go:271）；Bean ID 格式 `包路径.类型名#自定义名称`，自定义名为空时自动生成（core/doc.go BeanRegister、core/registry/doc.go:11）；注册表/实例缓存用 sync.Map，`sync.Map` 的值必须可比较（不能是 slice）（ADR-007） |
| boot | BootError 已改为接口，旧代码用类型别名 `BootErrorStruct = bootError` 兼容类型断言场景，新代码应面向 `BootError` 接口（boot/doc.go:133-137）；Starter/自动配置用 Kahn 拓扑排序处理依赖，存在循环依赖时回退到原始注册顺序（boot/doc.go:269、boot/starter.go:77、boot/autoconfig.go:375）；Application 优先直接使用 `*Boot`（`NewApplication` 返回 *Boot），需要接口类型（测试 mock）时再使用 `Application` 接口（boot/doc.go:147-176） |
| event | 事件通过 `Type() string` 字符串路由，字符串需全局一致，拼错/不一致即不路由（ADR-009，event/doc.go:79-88）；`EventBus` 现为接口（event/doc.go:187），默认实现是未导出 `eventBus`，`EventBusWithOrdering` 是导出结构体而非接口，需接口时用 `LegacyEventBusAdapter` 适配（event/doc.go:198-217、event/listener.go:15-17）；监听器存储用 sync.Map，由于 slice 不可比较，用 `listenerSlice` 指针包装才能支持 CompareAndSwap（event/doc.go:149-154） |
| cache | Get 未命中返回 `ErrNotFound`，不要用零值判断命中与否（cache/doc.go:63-64）；`Set` 的 `ttl<=0` 表示永不过期，传负数 TTL 不是错误而是不做过期（cache/doc.go:66-67）；Cache 接口 4 方法，仅需检查键/过期时间的场景用可选接口 `CacheInspector`（cache/doc.go:79-80） |
| web | 顶层 `web` 包是别名门面，`type Context = core.Context` 等全部重导出子包类型（web/doc.go:74-114）；真实实现在 `web/core`、`web/engine`、`web/mvc`、`web/server` 子包，改接口要同步改子包（web/doc.go "架构设计"）；新增网络引擎按扩展指南实现 engine.Factory/core.Router/core.Server/core.Context 并在 init() 注册（web/doc.go "扩展指南"） |
| condition | OnProperty 依赖 `config/environment` 读属性、OnBean 依赖容器，配置/容器未就绪时条件可能求值失败；条件可通过 And/Or/Not 组合（condition/doc.go:44-50）；Go 无 OnClass，用 `OnModuleLoaded`/`OnMissingModule` 替代（condition/doc.go:13-14, 21） |
| security | 顶层接口众多且多为别名（Authentication→authentication.Authentication 等），改抽象先看 security 子包分组（security/doc.go "架构设计"）；HttpSecurity 是链式大接口，由 Configurer（14 个设置方法）+ Builder（2 个构建方法）组合而成共 16 方法，改它影响面大（security/doc.go:230-287）；老式 `SecurityConfigurer` 已废弃，用 `NewSecurityConfig + WithXxx` 函数式选项替代（security/doc.go:319-325） |
| async | doc.go 仅定义 `Future` struct、`ExecutorOption`、`RejectHandler`，`AsyncExecutor` 接口在 executor.go（async/doc.go:54-71）；Submit 的 worker 池语义（池大小/队列/拒绝策略）由 WithPoolSize/WithQueueSize/WithRejectHandler 控制，任务队列满时走拒绝处理（async/doc.go:42-46） |
| validation | 提供 unsafe 类型转换（validation/unsafe.go）与对象池（validation/pool.go）优化，注意并发安全边界——不要越过其内置同步假设共享 unsafe 句柄；校验器注册表并发安全，自定义验证器经 ValidatorRegistry 注册（validation/doc.go:14） |
| resilience | 多负载均衡算法并存：roundrobin/leastconn/random/sticky/adaptive 等实现分散在各自文件，共用 Builder 构造（resilience/roundrobin.go、leastconn.go、random.go、sticky.go、adaptive.go、builder.go）；熔断器有 CLOSED/OPEN/HALF_OPEN 三态，半开状态有限流试探语义（resilience/doc.go:46-49） |
| metrics | Counter/Gauge/Histogram 接口+Exporter 接口在 metrics/doc.go，Prometheus 导出实现 `PrometheusExporter` 在 prometheus_exporter.go，控制台导出 ConsoleExporter 在 registry.go（metrics/doc.go:84、metrics/prometheus_exporter.go:14、metrics/registry.go:473）；具体后端集成在 starter/prometheus、starter/otel（metrics/doc.go:33-36） |
| tracing | Tracer/Span 是结构体（非纯接口），Sampler/Exporter 是接口（tracing/doc.go:8-12）；Span 数量限制、并发安全用读写锁+原子、ID 用 crypto/rand 生成（tracing/doc.go:22-23）；后端集成在 starter 下（tracing/doc.go "集成后端"） |
| lifecycle | 双机制并存：hooks.go 提供 `Hook`/`HookFunc`（OnInit/OnStart/OnStop），lifecycle.go 提供 `PhaseListener`（OnPhaseChange）——改生命周期逻辑先分清在加钩子还是阶段监听（lifecycle/doc.go:63-80）；阶段转换必须正向 INIT → RUNNING → STOPPED，反向转换返回错误（lifecycle/doc.go:42-43） |
| exception | Handler + Resolver 体系：核心入口 ExceptionHandler、解析器 ExceptionResolver、兜底 `DefaultExceptionResolver`（Order=1000 最后执行）在 exception/builtin.go（exception/doc.go:8-14、exception/builtin.go:23-29）；指标记录走 MetricsRecorder 在 metrics_recorder.go；错误码 ErrorCode/ErrorResponse 统一响应格式（exception/doc.go:12-14） |
| spel | expression 求值实现拆分多文件：evaluator.go / evaluator_expressions.go / type_utils.go，改表达式解析注意类型工具 `type_utils.go` 的影响（spel/evaluator.go、spel/evaluator_expressions.go、spel/type_utils.go）；Expression/ExpressionParser/EvaluationContext 接口在 spel/doc.go（spel/doc.go:49-69） |
| schedule | Cron 为 6 字段 Spring 风格：`秒 分 时 日 月 周`，缺秒位会解析错误或语义不符（schedule/doc.go:24-29、schedule/doc.go:48）；scheduler.go + builder.go + task.go 分工，Task.Name 唯一，重复 Register 返回 error（schedule/doc.go:60-61）；Register/Unregister 用任务名定位（schedule/doc.go:63-64） |
| tenant | TenantResolver/Manager/Middleware 三套接口协同，租户上下文经 Context 传递，改租户上下文注意 `TenantFromContext` 的传递链路（tenant/doc.go:9-12, 44-46）；中间件自动设置租户上下文（tenant/doc.go:11, 41-42） |
| openapi | 无典型 doc.go 接口门面，Document 结构体实现在 openapi.go，配合 swagger.go 生成 UI（openapi/openapi.go、openapi/swagger.go）；以空导入启用 `import _ "github.com/xudefa/enhance/openapi"`（openapi/doc.go:33） |
| metadata | MetadataGenerator（生成器）+ PropertyIndex（属性索引）双核心；注解解析基于 struct tag（TagAnnotationResolver），改注解解析注意与 tag 的关系（metadata/doc.go:11-13）；配置元数据含 Property/Group/Hint 三类结构（metadata/doc.go:9-10） |
| mq | 核心只有 `Queue` 接口 + Message/MessageHandler/QueueOption（mq/doc.go:8-11）；mq_helpers.go 提供 AcquireMessage 等实用工具，消息对象用 sync.Pool 复用（mq/doc.go:19, 29）；后端集成在 starter/kafka、starter/rabbitmq（mq/doc.go:43-46） |
| email | doc.go 只定义 `Sender` 接口 + Message/Attachment/SenderOption，SMTP 实现（net/smtp）在 sender.go（email/doc.go:8-11）；密码从环境变量 `ENHANCE_EMAIL_PASSWORD` 读取，勿硬编码（email/doc.go:39, 48） |
| i18n | MessageSource + Locale 双类型，消息回退顺序：精确区域 → 语言 → fallback → 消息代码本身（i18n/doc.go:39-43）；支持百分号转义，转义行为有专属测试 percent_test.go（i18n/percent_test.go） |
| testing | TestContext/TestingT/Mock 三件套（testing/doc.go:6-10）；`Mock.Expect(method, args, result, err)` 需要精确匹配参数（含参数列表），默认期望调用 1 次，`ExpectTimes` 可指定次数，最后必须 `Verify()`（testing/doc.go:93-107） |
| webtest | TestServer/RequestBuilder/ResponseVerifier/TestClient 等接口，HTTP 客户端实现在 client.go（webtest/doc.go:8-11）；断言含状态码/响应头/响应体/JSONPath（webtest/doc.go:20-37） |
| retry | 无 doc.go 类型门面，核心全在 retry.go：`RetryableFunc[T]`（泛型可重试函数签名）+ `Executor`（支持 jitter/上下文取消/回调） + BackoffStrategy（None/Fixed/Linear/Exponential）（retry/retry.go:157-158, retry/doc.go:6-9）；错误是否可重试经 `WithRetryable(fn)` 自定义（retry/retry.go:211-213） |
| log | Logger 接口族是多个小接口组合：Logger + LoggerWithSync/LoggerWithFields/LoggerFatal/LoggerWithLevel/LoggerLevelChecker/LoggerWithName/LoggerWithCaller/LoggerWithTimeout（log/doc.go:100-164），按能力组合而非一个大接口；builder 支持动态级别；后端集成在 starter/zap、starter/zerolog、log/slog（log/doc.go:38-42） |
| config | Config + Loader + Validator + ConfigCenter（配置中心）多概念并存，别混用（config/doc.go:8-16, 133）；watch.go 提供 WatchManager 配置热刷新/热重载（config/doc.go:11, config/watch.go）；配置优先级：命令行 > 环境变量 > 配置文件 > 默认值（config/doc.go:51-55） |
| context | ApplicationContext 是 Facade 聚合容器/环境/事件/生命周期，构建入口在 builder.go（context/doc.go:38-40, context/builder.go）；AsyncEventPublisher 由 asyncEventPublisherAdapter 适配 event.AsyncPublisher，改异步发布注意适配层（context/doc.go:40, 63-71）；EventBusAccess 支持 EventBus 与 EventBusWithOrdering（context/doc.go:73-76） |
| actuator | 健康检查是 Indicators 体系：Memory/DiskSpace/Process/System/Runtime 等 Indicator 实现 + IndicatorRegistry/aggregator（actuator/memory_health_indicator.go、disk_space_health_indicator.go、process_health_indicator.go、actuator/health/indicator_registry.go）；端点注册在 endpoint_registry.go，Sanitizer 过滤敏感信息（actuator/doc.go:8-10, actuator/endpoint_registry.go）；以空导入启用（actuator/doc.go:26） |
| observability | 非代码包：仅 README + metrics 子目录，**勿在此加 Go 文件**（observability/README.md、observability/metrics/） |

## 3. AI 高频错误

1. 把接口定义写进实现文件（违反 doc.go 门面规范，AGENTS.md §1.4 / ADR-005）
2. 构造函数返回具体类型而非接口（违反 ADR-006）
3. 核心包误 import 第三方库（违反 ADR-004）
4. 用 `==` 比较错误而非 errors.Is/As（违反 AGENTS.md §6.1）
5. 改接口直接删方法/改签名，未做影响面与兼容处理（违反 SOP 3，改前先读 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md)；先看 [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md)）
6. 新 starter 依赖其他 starter 包（违反 AGENTS.md §0.1）
7. 测试未加 t.Parallel()（违反 AGENTS.md §2.3，表驱动子测试也应加；优先表驱动测试，ADR-012）