# AI 维护文档体系实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `docs/` 下建立 5 份导航索引级 AI 维护文档（总入口、代码导航地图、变更影响、任务 SOP、坑点），打通 `AGENTS.md` 入口，并顺带修复文档编写过程中发现的 doc.go 注释缺失问题。

**Architecture:** 分主题多文件，每份文档职责单一、互相索引。AI 可按需加载导航地图/影响分析/SOP/坑点，避免全量读取。数据来源为各包 `doc.go`、源码 import 与现有根文档（`AGENTS.md`、`ARCHITECTURE.md`、`DEPENDENCIES.md`、`ADR.md`）。

**Tech Stack:** Go 标准库、Markdown、git。

## Global Constraints

- 全部文档使用**中文**，与现有文档（AGENTS.md 等）语言一致
- 文档粒度：**导航索引级**，不逐功能深讲（深讲是各包 README 职责）
- 与现有根文档互补，**不重写** `AGENTS.md`、`ARCHITECTURE.md`、`DEPENDENCIES.md`、`CODING_STYLE.md` 内容，只做入口打通/引用
- `AI_CHANGE_IMPACT.md` 的依赖数据以**实际 import（grep 源码）为准**，与 `DEPENDENCIES.md` 不一致处以源码为准并注明差异
- 顺手修复**仅限**：给缺少 godoc 注释的导出符号补注释、给缺少包级文档的 doc.go 补包注释。**禁止**改动任何实现逻辑、命名、行为
- 每次对 go 源码的修改必须通过 `go build ./...` 和相关包 `go test`
- 每完成一个任务即提交一次 git

## 已收集的源码数据（供各任务直接使用）

### 各包核心接口（来自各包 doc.go，正则 `^type [A-Z]\w+ interface`）

| 包 | 核心接口 |
|----|----------|
| actuator | AppContext, RouteRegistrar, SanitizeStrategy |
| audit | EventWriter, Auditor, AuditInterceptor, AuditLogger |
| boot | Application, ApplicationContext, AutoConfiguration, Starter, StarterRegistry, BootError, FailureAnalyzer, EventBusResult |
| cache | Cache, CacheInspector, Clearable |
| condition | Condition, EnvironmentAccessor, ContainerAccessor, ConditionContext |
| config | Config, Validator, Loader, ConfigCenter |
| context | EventPublisher, AsyncEventPublisher, EventBusAccess, ApplicationContext |
| core | BeanGet, BeanExistenceChecker, BeanLister, BeanRegister, BeanIDGenerator, Container, ContainerExt, BeanCreator |
| email | Sender |
| event | ApplicationEvent, EventBus, AsyncPublisherBus |
| exception | Logger, MetricsRecorder, ResponseWriter, ExceptionHandler, ExceptionResolver |
| i18n | MessageSource |
| lifecycle | PhaseListener, Hook |
| log | Logger, LoggerWithSync, LoggerWithFields, LoggerFatal, LoggerWithLevel, LoggerLevelChecker, LoggerWithName, LoggerWithCaller, LoggerWithTimeout, Sampler |
| metadata | MetadataGenerator, PropertyIndex, TagAnnotationResolver |
| metrics | Counter, Gauge, Histogram, Exporter, MeterRegistry |
| mq | Queue |
| resilience | Breaker, Selector, Registry, Balancer |
| schedule | Task, Scheduler |
| security | SecurityContext, AccessDeniedHandler, AuthenticationEntryPoint, SecurityMetadataSource, SecurityRequest, SecurityResponse, GrantedAuthority, Configurer, Builder, HttpSecurity, AuthorizeRequests, ExpressionInterceptUrlRegistry, SecurityConfigurer, LogoutHandler, LogoutSuccessHandler, RateLimiter, RateLimitStrategy, CsrfTokenRepository |
| spel | Expression, ExpressionParser, EvaluationContext, PropertyAccessor, MethodInterceptor, MethodInvocation |
| tenant | TenantResolver, TenantManager, TenantMiddleware, TenantIsolation, TenantRegistry, TenantProvider |
| testing | TestingT, TestContext, Mock |
| tracing | Sampler, Exporter |
| validation | Validator, CustomValidator, MiddlewareValidator, ResponseWriter, Binder |
| webtest | TestServer, RequestBuilder, ResponseVerifier |

**特殊说明（doc.go 非标准 interface 定义或仅有类型别名）：**
- `web/doc.go`：全部是类型别名（`type Context = core.Context` 等），覆盖 core.Context/Router/Server/Controller/HandlerFunc/MiddlewareFunc 与 engine/mvc 子包。实际接口在 `web/core/`、`web/server/`、`web/mvc/`、`web/engine/` 子包
- `async/doc.go`：定义 `Future` struct、`ExecutorOption`、`RejectHandler`；核心 `Executor` 接口需在 `executor.go` 中确认
- `retry/doc.go`：无 `^type` 定义；核心接口在 `retry.go` 中确认
- `openapi/doc.go`：无 `^type` 定义；`Document` 等类型在 `openapi.go`/`builder.go` 中确认
- `metrics/doc.go`：额外有 `Metric` struct
- `observability/`：无顶层 Go 文件，仅 `README.md` 与 `metrics/` 子目录，导航地图中标注"非代码包"
- `core/doc.go` 中部分接口同时存在 `core/container.go` 等实现文件中

### 各包内部依赖（grep import，`observability` 无代码，`cmd/`、`docs/`、`starter/`、`examples/` 不在范围）

| 包 → 依赖的 enhance 包 |
|------------------------|
| actuator → actuator/admin, actuator/health, boot, condition, config/environment, core, event, metrics |
| boot → boot/banner, condition, config, config/environment, core, core/registry, event, lifecycle |
| condition → config/environment |
| config → config/environment |
| context → config/environment, config/refresh, core, core/registry, event, lifecycle |
| core → core/lifecycle, core/registry, core/scope |
| event → retry |
| metrics → boot, condition, config/environment, core |
| schedule → boot, condition, config/environment, core, log |
| security → boot, condition, config/environment, core, log, security/authentication, security/authorization, security/filter |
| testing → boot, config/environment, core, core/registry |
| tracing → boot, condition, config/environment, core, log |
| web → web/core, web/engine, web/mvc, web/server |
| async, audit, cache, email, exception, i18n, lifecycle, log, mq, openapi, resilience, retry, spel, tenant, validation, webtest → 无内部依赖（独立包） |

---
---

## Task 1: AGENTS.md 入口打通 + 创建 docs/AI_INDEX.md

**Files:**
- Modify: `/Users/xudefa/workspace/enhance/AGENTS.md`（顶部引用区）
- Create: `/Users/xudefa/workspace/enhance/docs/AI_INDEX.md`

**Interfaces:**
- Produces: `docs/AI_INDEX.md` —— 是后续 Task 2-5 文档的导航总入口（各文档互相链接到它）

- [ ] **Step 1: 修改 AGENTS.md 入口**

在 `AGENTS.md` 顶部"重要"引用块（当前为三个链接那行）追加一行。精确替换：

```markdown
> **相关文档**：[代码风格指南](CODING_STYLE.md) | [架构设计](ARCHITECTURE.md) | [贡献指南](CONTRIBUTING.md)
```

改为：

```markdown
> **相关文档**：[代码风格指南](CODING_STYLE.md) | [架构设计](ARCHITECTURE.md) | [贡献指南](CONTRIBUTING.md) | [AI 维护导航](docs/AI_INDEX.md)
>
> **AI 首次接入必读**：在动手改代码前，先阅读 [docs/AI_INDEX.md](docs/AI_INDEX.md) 完成代码库导航，并参考 [docs/AI_TASK_HANDBOOK.md](docs/AI_TASK_HANDBOOK.md) 中的任务 SOP。
```

- [ ] **Step 2: 创建 docs/AI_INDEX.md**

创建文件，内容骨架（中文，实测占位：标题 + 表格 + 列表）：

```markdown
# enhance 框架 AI 维护导航总入口

> 本文档是 AI/开发者维护 enhance 代码库的**第一入口**。所有变更任务开始前请先阅读本节，按需加载导航文档。

## 1. 这是什么

enhance 是一个 Go 企业级框架（526+ Go 文件、30+ 顶层包）。为降低 AI 维护成本，docs/ 下提供一组**导航索引级**文档，AI 可按需加载。

## 2. 文档导航

| 文档 | 用途 | 什么时候读 |
|------|------|-----------|
| [AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md) | 逐包代码导航：文件职责、核心接口、关键实现 | 需要定位某个包/文件/接口时 |
| [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) | 包间依赖、修改影响面、回归范围 | 动手改代码**之前** |
| [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md) | 新增包/接口/模块/测试/验收的 SOP | 执行具体开发任务时 |
| [AI_PITFALLS.md](AI_PITFALLS.md) | 各包设计决策、坑点、隐蔽约定 | 遇到诡异行为/历史决策时 |
| [ADR.md](ADR.md) | 架构决策记录归档（现有） | 想了解"为什么这样设计"时 |

现有根文档（补充，不重复内容）：
[AGENTS.md](../AGENTS.md)（全局规范） · [ARCHITECTURE.md](../ARCHITECTURE.md)（架构设计） · [DEPENDENCIES.md](../DEPENDENCIES.md)（模块依赖） · [CODING_STYLE.md](../CODING_STYLE.md)（代码风格） · [CONTRIBUTING.md](../CONTRIBUTING.md)（贡献指南）

## 3. 首次接入：按顺序完成

1. 读本文件（第 1-2 节）
2. 读 [AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md) 了解包结构
3. 读 [AGENTS.md](../AGENTS.md) 了解硬性规范（零外部依赖、doc.go 门面、依赖方向）
4. 读 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) 了解你准备修改的包的影响面
5. 读 [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md) 5.7 验收清单，再开始动手

## 4. 按任务类型读文档

| 任务 | 必读 |
|------|------|
| 新增一个包 | AI_TASK_HANDBOOK §新增包、AI_NAVIGATION_MAP |
| 修改现有包的接口 | AI_CHANGE_IMPACT、AI_TASK_HANDBOOK §修改接口 |
| 新增 auto-config/starter | AI_TASK_HANDBOOK §新增自动配置、AI_PITFALLS |
| 新增测试 | AI_TASK_HANDBOOK §测试、AGENTS.md §2.3 |
| 排查诡异行为 | AI_PITFALLS、ADR.md |

## 5. 硬性规范速查（违反会被拒收）

### 5.1 必须遵守
- ✅ 核心框架包（core/boot/...）**零外部依赖**，只 import 标准库 + enhance 内部包
- ✅ 包间只能通过接口交互，实现隐藏；`doc.go` 是门面（只放定义）
- ✅ 构造函数返回接口；接口命名 `er` 后缀，禁止 `I` 前缀；接口 ≤5 方法
- ✅ 依赖方向单向（Boot → Core → Infrastructure），禁止循环依赖
- ✅ Go 惯用法：早期返回、函数式选项、组合优于继承、泛型优先于反射
- ✅ 测试用表驱动，测试函数与子测试 `t.Parallel()`

### 5.2 绝对禁止
- ❌ 核心包引入第三方依赖（第三方集成必须放 `starter/`）
- ❌ 相对导入、忽略错误、裸 goroutine、直接用 `==` 比较错误
- ❌ 在 doc.go 写实现逻辑
- ❌ 大改现有接口而不做影响分析与兼容处理

## 6. 验收标准（任务完成前自检）

1. `go build ./...` 通过
2. 相关包 `go test` 通过
3. 导出符号均有 godoc 注释
4. 文档引用的文件名/接口名与源码一致（用 grep 抽查）
```

- [ ] **Step 3: 验证**

运行：
```bash
grep -n "AI_INDEX" /Users/xudefa/workspace/enhance/AGENTS.md
test -f /Users/xudefa/workspace/enhance/docs/AI_INDEX.md && wc -l /Users/xudefa/workspace/enhance/docs/AI_INDEX.md
```
预期：`AGENTS.md` 出现 AI_INDEX 链接，`AI_INDEX.md` 存在且 > 60 行。

- [ ] **Step 4: 提交**

```bash
cd /Users/xudefa/workspace/enhance && git add AGENTS.md docs/AI_INDEX.md && git commit -m "docs: AGENTS 入口打通并新增 AI 维护导航总入口"
```

---

## Task 2: docs/AI_NAVIGATION_MAP.md —— 逐包代码导航地图

**Files:**
- Create: `/Users/xudefa/workspace/enhance/docs/AI_NAVIGATION_MAP.md`

**Interfaces:**
- Consumes: Task 1 的 `docs/AI_INDEX.md`（用它做互引）
- Produces: 供 Task 3/5 复核的包结构底稿

- [ ] **Step 1: 写文档头 + 模板说明**

创建文件，开头内容：

```markdown
# enhance 代码导航地图（逐包）

> AI 定位代码用。每个包给出：核心接口（在 doc.go）、文件职责、关键实现、依赖、注意事项。
> 与源码一致性以 grep 抽查为准。数据更新日期见文件尾部。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md)

## 模板说明

每个包统一使用以下六段式：

markdown 结构：
### `包名`（一句话职责）
- **核心接口**：`doc.go` 中的接口
- **文件职责**：表格（文件 → 职责）
- **关键实现文件**：核心逻辑所在文件 + 函数
- **依赖**：import 的 enhance 内部包
- **注意事项**：该包的隐蔽约定
```

- [ ] **Step 2: 按下列分组写入全部包的六段式内容**

执行方式：对每个包先运行 `grep -E '^type |^func ' <pkg>/doc.go` 与 `ls <pkg>/*.go`（去 `_test.go`），结合下表"核心接口/依赖/特殊说明"逐条书写。**禁止臆造函数名**，只写 grep 或 `Read` 确认过的内容。

**A. 核心框架与配置（6 包）**：`core`、`boot`、`context`、`condition`、`lifecycle`、`config`
**B. 数据与网络（6 包）**：`cache`、`web`、`mq`、`resilience`、`email`、`metadata`
**C. 可观测性（5 包）**：`log`、`metrics`、`actuator`、`tracing`、`observability`（标注"非代码包"）
**D. 安全与验证（4 包）**：`security`、`validation`、`exception`、`tenant`
**E. 调度与事件（4 包）**：`event`、`schedule`、`async`、`retry`
**F. 工具与扩展（5 包）**：`i18n`、`spel`、`testing`、`webtest`、`openapi`

**六段式内容来源：**
- 核心接口：使用 Global Constraints 中的"各包核心接口"表 + `grep '^type [A-Z]\w* interface\|^type [A-Z]\w* struct' <pkg>/doc.go`
- 文件职责：`ls <pkg>/*.go | grep -v _test.go`，逐个用 `Read` 首部注释判断职责；无法判断时写"实现文件，见关键接口"
- 依赖：使用 Global Constraints 中的"各包内部依赖"表
- 注意事项：来自 `doc.go` 包注释中的设计说明、`docs/ADR.md` 对应条目（如 ADR-005 doctor 门面、ADR-007 sync.Map、ADR-009 事件字符串路由）

**每个包必写的注意事项（已有据可查）：**
- `core`：接口方法不能带类型参数 → Container 用 reflect.Type，用户层用泛型 API（见 core/doc.go 包注释）；接口拆分小接口组合（BeanGet/BeanRegister 等）
- `boot`：BootError 由结构体改接口，保留 BootErrorStruct 类型别名兼容（doctor 注释 §向后兼容）
- `event`：事件用字符串 Type() 路由而非反射类型（ADR-009）
- `cache`：Get 未命中返回 ErrNotFound；ttl<=0 表示永不过期（doc.go）
- `web`：doc.go 全部为类型别名，实现分布在 web/core、web/engine、web/mvc、web/server 子包
- `condition`：依赖 config/environment（读配置）
- `security`：接口数量最多（20+），按 authentication/authorization/filter 子包分组
- `observability`：无 Go 代码，仅 README + metrics 子目录

**每个包注意事项中的坑点同步提炼进 Task 5，不在此重复放大。**

- [ ] **Step 3: 一致性抽查（每个包至少 2 个断言）**

```bash
grep -c "^type [A-Z].*interface" /Users/xudefa/workspace/enhance/core/doc.go   # 断言 core 接口存在
grep -n "type Cache interface" /Users/xudefa/workspace/enhance/cache/doc.go    # 断言 cache 条目属实
```
对随机 5-8 个包的：接口名 `grep -n "type <接口名> interface" <pkg>/doc.go`、文件名 `test -f <pkg>/<文件名>`。抽查不通过则修正文档对应条目。

- [ ] **Step 4: 尾部标注数据日期 + 提交**

文件尾部追加：

```markdown
---
> 数据核对日期：2026-09-16。源码更新后请更新本文件对应条目。
```

提交：
```bash
cd /Users/xudefa/workspace/enhance && git add docs/AI_NAVIGATION_MAP.md && git commit -m "docs: 新增逐包代码导航地图"
```

---

## Task 3: docs/AI_CHANGE_IMPACT.md —— 包间依赖与变更影响

**Files:**
- Create: `/Users/xudefa/workspace/enhance/docs/AI_CHANGE_IMPACT.md`

**Interfaces:**
- Consumes: Task 2 的导航地图（引用其包结构说明）、Global Constraints 依赖表
- Produces: 供 Task 4 的 SOP 引用（修改接口/新增包的回归范围依据）

- [ ] **Step 1: 写文档头 + 依赖方向总览**

```markdown
# enhance 包间依赖与变更影响分析

> AI 在**修改任何包之前**阅读本文，确认影响面与回归范围。
> 依赖数据来自源码 import（grep 校验），与 DEPENDENCIES.md 不一致处以源码为准。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md) · 同级：[AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md)

## 1. 依赖方向总览

Boot Layer → Core Layer → Infrastructure Layer（详见 [DEPENDENCIES.md](../DEPENDENCIES.md)）。
`core/` 零外部依赖（仅标准库）；各层无条件只能单向依赖。

## 2. 包直接依赖表（正向：A 依赖 B 才能编译）

| 包 | 依赖的 enhance 包 | 说明 |
|----|------------------|------|
| actuator | actuator/admin, actuator/health, boot, condition, config/environment, core, event, metrics | 应用层聚合 |
| boot | boot/banner, condition, config, config/environment, core, core/registry, event, lifecycle | 启动核心 |
| ...（其余包从数据中照填） |
| async/cache/email/...（无依赖包） | 无 | 独立包 |
```

- [ ] **Step 2: 反向影响表（修改 X 会影响谁）**

反向推导：遍历所有包，凡是依赖了 X 的包都列入 X 的影响面。人工用 grep 复核一次：

```bash
grep -rl 'enhance/core' /Users/xudefa/workspace/enhance --include='*.go' | grep -v _test.go | grep -v /cmd/ | sed 's|/Users/xudefa/workspace/enhance/||' | cut -d/ -f1 | sort -u
```
输出示例（以实际为准）：`actuator boot context metrics schedule security testing tracing ...`

文档表格示例：

| 修改的包 | 受影响的下游包（需回归测试） |
|----------|------------------------------|
| core | boot, context, metrics, schedule, security, testing, tracing, actuator, ... |
| boot | actuator, metrics, schedule, security, testing, tracing, ... |
| config | boot, context, ...（凡 import config 者） |
| log | schedule, security, tracing, ... |
| 无依赖包（如 cache/email/mq） | 无下游，仅该包自身 |

- [ ] **Step 3: 执行顺序表**

从 `boot/doc.go` 的 `OrderPriority`（包注释）整理为表：

| 优先级 | 名称 | 层级 | 代表模块 |
|--------|------|------|----------|
| -3000 | OrderPriorityInfrastructure | 基础设施 | log, config, condition |
| -2000 | OrderPriorityDataLayer | 数据层 | cache, gorm |
| -1500 | OrderPriorityAuthentication | 认证层 | jwt |
| -1200 | OrderPriorityAuthorization | 授权层 | casbin |
| -100 | OrderPrioritySecurityCore | 安全核心 | security |
| 0 | OrderPriorityWebLayer | Web 层 | web |
| 1000 | OrderPriorityBusinessLayer | 业务层 | event, schedule, async |
| 2000 | OrderPriorityMonitoringLayer | 监控层 | actuator, metrics |

值越小先执行。新增 auto-config 时必须选对优先级（引 Task 4 SOP）。

- [ ] **Step 4: 修改后验证清单**

```markdown
## 4. 修改任意包后的必做验证

1. `go build ./...` —— 全量编译（含受影响包）
2. 受影响包定向测试：`go test <受影响包路径列表>`
3. `go test -race ./...` 或至少受影响包 `-race`
4. `gofmt -l .` 无输出；修改导出符号时检查 godoc 注释
5. 若改了接口：按 AI_TASK_HANDBOOK §修改接口 做影响面排查与兼容处理
```

- [ ] **Step 5: 一致性抽查 + 提交**

抽查命令：
```bash
grep -rn 'enhance/core"' /Users/xudefa/workspace/enhance/boot/*.go | head -3   # boot 依赖 core 属实
grep -rln 'No dependency' docs/AI_CHANGE_IMPACT.md > /dev/null 2>&1 || true
```
提交：
```bash
cd /Users/xudefa/workspace/enhance && git add docs/AI_CHANGE_IMPACT.md && git commit -m "docs: 新增包间依赖与变更影响分析"
```

---

## Task 4: docs/AI_TASK_HANDBOOK.md —— 常见任务 SOP

**Files:**
- Create: `/Users/xudefa/workspace/enhance/docs/AI_TASK_HANDBOOK.md`

**Interfaces:**
- Consumes: Task 3 的验证清单（作为"修改后验证"引用）、Global Constraints
- Produces: AI 直接执行的操作手册

- [ ] **Step 1: 写文档头 + 通用前置检查**

```markdown
# enhance 常见 AI 任务操作手册（SOP）

> 本文给出 AI 在 enhance 上执行各类开发任务的**标准步骤**。每个 SOP 可直接照做。
> 规范依据：AGENTS.md + CODING_STYLE.md。提交规范：每完成一个逻辑单元提交一次。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md)

## 0. 通用前置检查（所有任务）
- [ ] 已读 AI_NAVIGATION_MAP.md 确认目标包位置
- [ ] 已读 AI_CHANGE_IMPACT.md 确认影响面
- [ ] 包级使用完整模块路径导入，禁止相对导入
- [ ] 不引入任何第三方依赖（先确认有没有现成的接口/实现）
```

- [ ] **Step 2: 写 6 个 SOP，每个含可执行步骤与命令**

**SOP 1：新增一个包**
1. 建目录 `<pkg>/`，包名小写无下划线
2. 创建 `<pkg>/doc.go`：包级 godoc（说明职责/设计/使用示例）+ 所有对外接口/类型定义；≤500 行，超出建 `types.go`
3. 创建实现文件（如 `impl.go`），构造函数**返回接口**；实现细节私有
4. 若需自动配置/Starter：新建 `autoconfig.go`，用 `boot.RegisterAutoConfig` + `condition.OnProperty/OnBean` **条件装配**，按 AI_CHANGE_IMPACT §3 选 OrderPriority，检查无循环依赖
5. 创建 `<pkg>/README.md`（使用说明）
6. 创建 `<pkg>/*_test.go`：表驱动 + `t.Parallel()`，覆盖正常/错误路径，覆盖率尽量 ≥80%
7. 验证：`go build ./... && go test ./<pkg>` 
8. 提交：`git add <pkg>/ && git commit -m "feat: add <pkg> package"`

**SOP 2：新增接口/类型**
1. 确认归属包；接口定义放该包 `doc.go`（或 types.go），**不放实现文件**
2. 命名：接口 `er` 后缀或功能名，禁止 `I` 前缀；方法 ≤5 个
3. 每个方法写 godoc：方法名/参数/返回值/错误语义
4. 既有实现类补方法以满足接口；构造函数返回接口
5. 若接口要被多个包复用，用 `web` 的"类型别名重新导出"模式（见 AI_NAVIGATION_MAP web 条目）
6. 验证 + 提交（同 SOP 1 第 7-8 步）

**SOP 3：修改现有接口**
1. 用 `grep -rn '<接口名>' --include='*.go'` 找出全部实现方与调用方
2. 用 AI_CHANGE_IMPACT.md 受影响表确定回归范围
3. 优先加新方法/新接口（组合），避免删除/修改签名破坏实现方；必要时用类型别名保留旧 API（参考 boot.BootErrorStruct 兼容做法）
4. 更新所有实现方与调用方；新增方法补实现
5. 验证：受影响包 `go test XXX` + `go build ./...`
6. 提交

**SOP 4：新增 auto-config / starter**
1. 实现 `AutoConfiguration.Configure(ctx ApplicationContext)` 或在 `boot.Starter` 中实现 Configure/Start/Stop + Dependencies()
2. `init()` 中 `boot.RegisterAutoConfig`，条件用 `condition.OnProperty("xxx.enabled","true")` 等；`OnMissingBean` 避免重复注册
3. 选对 OrderPriority（AI_CHANGE_IMPACT §3）；starter 间禁止互相依赖（每个 starter 独立）
4. 第三方 starter 放 `starter/<name>/`，自带独立 go.mod，不得依赖其他 starter 包
5. 示例放 `examples/<name>/`（独立 go.mod）
6. 验证 + 提交

**SOP 5：新增/修改测试**
1. 表驱动：`tests := []struct{...}`，必须 `t.Parallel()`，子测试也 `t.Parallel()`
2. 覆盖正常路径 + 错误路径 + 边界
3. 禁止修改全局状态/共享资源（需并发的场景用 `sync` 保护）
4. 运行：`go test -v ./<pkg>` → PASS 后 `go test -race ./<pkg>`
5. 提交

**SOP 6：任务验收（提交前最后一关）**
依次运行且全部通过：
```bash
go build ./...
go test ./...
go test -race <受影响包>
gofmt -l .          # 无输出
go vet ./...
```
自查：导出符号有 godoc；无 `else` 多余分支；错误用 `%w` 包装；无裸 goroutine。然后提交。

- [ ] **Step 3: 验证 + 提交**

验证：`grep -c 'SOP' docs/AI_TASK_HANDBOOK.md` 应 ≥6；`grep -c 'go test' docs/AI_TASK_HANDBOOK.md` 应 ≥4。
提交：
```bash
cd /Users/xudefa/workspace/enhance && git add docs/AI_TASK_HANDBOOK.md && git commit -m "docs: 新增常见任务操作手册"
```

---

## Task 5: docs/AI_PITFALLS.md —— 设计决策与坑点

**Files:**
- Create: `/Users/xudefa/workspace/enhance/docs/AI_PITFALLS.md`

**Interfaces:**
- Consumes: Task 2 导航地图中各包注意事项、`docs/ADR.md`、AGENTS.md §6
- Produces: 供 Task 6 校验闭环

- [ ] **Step 1: 写文档头 + 全局坑点**

```markdown
# enhance 设计决策与坑点（AI 避坑指南）

> 遇到诡异行为、历史决策、隐蔽约定时查这里。每条注明出处（源码注释 / ADR-XXX）。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md) · 关联：[ADR.md](ADR.md)

## 1. 全局硬约束（违反即返工）
- 核心包零外部依赖（ADR-004）；第三方集成必须走 starter/ 且各自独立 go.mod
- 依赖方向单向，禁止循环依赖；低优先级包不能依赖高优先级包
- 构造函数返回接口（ADR-006）；接口 ≤5 方法
- 事件用字符串路由而非反射类型（ADR-009）
- 优先级用 OrderPriority 枚举，禁止魔法数字（ADR-010）
- 错误不丢上下文：`fmt.Errorf("...: %w", err)`，判断用 errors.Is/As
```

- [ ] **Step 2: 逐包坑点（从源码注释与 ADR 提炼）**

对每个包写 1-5 条，含出处。**禁止臆造**，必须能对应到源码读到的内容。已有据可查的：

| 包 | 坑点（出处） |
|----|-------------|
| core | Container 接口无泛型方法，用户必须走泛型 API（core/doc.go）；BeanID 格式 `包路径.类型名#自定义名`（core/doc.go BeanRegister） |
| boot | BootError 已改为接口，旧代码 `BootErrorStruct` 兼容（boot/doc.go）；starter 拓扑排序失败会回退注册顺序（boot/doc.go StarterRegistry） |
| event | 事件类型字符串需全局一致，拼错即不路由（ADR-009）；EventBus 是结构体非接口（ARCHITECTURE.md） |
| cache | Get 未命中返回 ErrNotFound（cache/doc.go）；ttl<=0 永不过期 |
| web | 顶层是别名门面，真正实现在 web/ 子包（web/doc.go）；改接口先改 web/core/ server/ engine/ mvc/ |
| condition | 依赖 config/environment 读属性，配置未就绪时条件可能求值失败 |
| security | 接口众多（20+），各子包分组；HttpSecurity 是链式 Large interface（16 方法），改它影响面大 |
| async | doc.go 仅 Future/ExecutorOption，Executor 在 executor.go；改 executor 注意 worker 池语义 |
| validation | 支持 unsafe 与 pool 优化，注意并发安全边界（validation/unsafe.go, pool.go） |
| resilience | 包含多种负载均衡算法实现（roundrobin/leastconn/random/sticky/adaptive...），共用 Builder |
| metrics | Counter/Gauge/Histogram + Exporter + Prometheus exporter 实现（prometheus_exporter.go） |
| tracing | Tracer/Span 结构体，Exporter/Sampler 接口 |
| lifecycle | PhaseListener/Hook 双机制；区分 hooks.go 与 lifecycle.go 职责 |
| exception | Handler + Resolver 体系，含 builtin 解析器与 metrics_recorder |
| spel | evaluator 拆分多文件（evaluator.go, evaluator_expressions.go...），改表达式解析注意类型工具 type_utils.go |
| schedule | Cron 为 6 字段（参考 schedule/README）；scheduler.go + builder.go + task.go 分工 |
| tenant | TenantResolver/Manager/Middleware 三套接口协同，改租户上下文注意 Context 传递 |
| openapi | 无接口（结构体实现），Document 在 openapi.go；配合 swagger.go 生成 UI |
| metadata | MetadataGenerator + PropertyIndex；注意与 tag 注解解析的关系 |
| mq | Queue 接口；mq_helpers.go 提供实用工具 |
| email | Sender 接口 + smtp 实现（sender.go）；doc.go 只有 Sender |
| i18n | MessageSource + Locale，支持百分号转义（percent_test.go） |
| testing | TestContext/TestingT/Mock；Mock.Expect 需要精确匹配参数 |
| webtest | TestServer/RequestBuilder/ResponseVerifier；client.go 实现 |
| retry | 核心在 retry.go（无 doc.go 类型）；Retryable 接口在 retry.go 确认 |
| log | Logger 接口族（10 个接口组合），builder 支持动态级别/sampler |
| config | Config + Loader + Validator + ConfigCenter 多概念；watch.go 支持热刷新 |
| context | ApplicationContext 聚合容器/环境/事件；builder.go 构建 |
| actuator | 健康检查 Indicators 体系（memory/disk/process）+ 端点注册（endpoint_registry.go） |
| observability | 非代码包（仅 README + metrics 子目录），勿在此加 Go 文件 |

- [ ] **Step 3: AI 易错模式**

```markdown
## 3. AI 高频错误
1. 把接口定义写进实现文件（违反 doc.go 门面规范，AGENTS.md §1.4）
2. 构造函数返回具体类型而非接口（违反 ADR-006）
3. 核心包误 import 第三方库（违反 ADR-004）
4. 用 `==` 比较错误而非 errors.Is/As（AGENTS.md §6.1）
5. 改接口直接删方法/改签名，未做影响面与兼容处理（SOP 3）
6. 新 starter 依赖其他 starter 包（AGENTS.md §0.1）
7. 测试未加 t.Parallel()（AGENTS.md §2.3）
```

- [ ] **Step 4: 验证 + 提交**

验证：`grep -c '|' docs/AI_PITFALLS.md` 应涵盖 ≥20 个包；每条坑点含出处关键词（doc.go/ADR/xxx.go）。
提交：
```bash
cd /Users/xudefa/workspace/enhance && git add docs/AI_PITFALLS.md && git commit -m "docs: 新增设计决策与坑点指南"
```

---

## Task 6: 顺手修复 doc.go 注释缺失 + 全量验证 + 收尾

**Files:**
- Modify: 需要补注释的 `<pkg>/doc.go`（若有，全流程动态发现）
- Modify: `/Users/xudefa/workspace/enhance/docs/AI_NAVIGATION_MAP.md`（如有修复记录补充）
- Test: `go build ./...` + 相关包 `go test`

**Interfaces:**
- Consumes: Task 2-5 全部文档；全流程中记录的注释缺失清单
- Produces: 干净的注释、验证通过的代码库

- [ ] **Step 1: 汇总全流程发现的注释缺失项**

执行者扫描各包 doc.go，找出**缺少包级 godoc 注释**或**导出符号缺注释**的包。仅当满足以下条件才修复：
- 只是缺注释，无实现改动
- 能准确判断导出符号语义（不确定就不改，记入文档而不是该代码）

候选查找命令：
```bash
for d in */; do d="${d%/}"; case "$d" in starter|examples|docs|cmd|devtools|.git|.vscode|.superpowers) continue;; esac; [ -f "$d/doc.go" ] || continue; first=$(grep -m1 '^// Package' "$d/doc.go"); [ -z "$first" ] && echo "缺包文档: $d"; done
```

预期候选（已知）：`async`、`retry`、`openapi`、`observability`（无 doc.go，排查是否应有）。**逐个 Read 确认后修复**，每个修复示例（补包级注释）：

```go
// Package retry 提供重试机制封装。
//
// 使用方式见 retry.go 中 Retryable 接口与 Do 函数。参考 AGENTS.md §1.2。
package retry
```

- [ ] **Step 2: 修复（若上一步存在候选）**

对每个确认的包，编辑 doc.go：
- 无包注释 → 补 `// Package xxx ...` 注释块（说明职责 + 一句话使用方式）
- 导出符号缺注释 → 补一行 godoc

**边界**：不得改动函数体、签名、命名、行为；不得移动代码。每个修改的包同步在 `docs/AI_NAVIGATION_MAP.md` 对应条目"注意事项"补一句"已补包注释"。

- [ ] **Step 3: 全量验证**

```bash
cd /Users/xudefa/workspace/enhance
go build ./...
go test ./core/... ./boot/... ./event/... ./cache/... ./web/... ./retry/... ./async/... ./openapi/...
gofmt -l . | grep -v starter | head
```
预期：build 通过；相关包 test 通过；gofmt 无输出（或仅有 starter/ 预置差异）。

- [ ] **Step 4: 文档互链收尾 + README 补充导航**

1. 检查五份文档互链完整：
```bash
cd /Users/xudefa/workspace/enhance && for f in docs/AI_INDEX.md docs/AI_NAVIGATION_MAP.md docs/AI_CHANGE_IMPACT.md docs/AI_TASK_HANDBOOK.md docs/AI_PITFALLS.md; do echo "== $f"; grep -o 'AI_[A-Z_]*\.md' "$f" | sort -u; done
```
每个文件缺失的互引（尤其 AI_INDEX 引用其余四份）补齐。

2. 在 `README.md` 文档导航"核心文档"表格追加一行：
```markdown
| [AI 维护文档](docs/AI_INDEX.md) | AI 维护代码库的导航入口（地图/影响/SOP/坑点） | AI 智能体、开发者 |
```

- [ ] **Step 5: 提交**

```bash
cd /Users/xudefa/workspace/enhance && git add -A docs/ README.md && git commit -m "docs: 顺手修复包文档注释并完善 AI 维护文档互链"
```
> 若 Step 2 修复了 go 源码，需将对应 `*.go` 一并 add 进本提交并在 commit message 中注明。

---

## Self-Review 结果

**Spec 覆盖检查：**
- ✔ docs 下 5 份文档 → Task 1(AI_INDEX)、Task 2(NAVIGATION)、Task 3(CHANGE_IMPACT)、Task 4(HANDBOOK)、Task 5(PITFALLS)
- ✔ AGENTS.md 入口打通 → Task 1 Step 1
- ✔ 导航索引级粒度 → Global Constraints + 六段式模板（Task 2）
- ✔ 变更影响分析（依赖矩阵/反向影响/执行顺序/验证清单）→ Task 3
- ✔ 任务 SOP（新增包/接口/改接口/autoconfig/测试/验收）→ Task 4 的 6 个 SOP
- ✔ 每包坑点 → Task 5 逐包表 + 出处
- ✔ 顺手修复小问题（仅注释、最小改动、验证）→ Task 6 + Global Constraints 边界
- ✔ 中文 → Global Constraints
- ✔ 验收标准 1-8 → Task 6 Step 3-4（build/test/互链/README）

**占位符扫描：** 无 TBD/TODO；所有自称"确认来源"的条目均给出确切 grep/Read 手段。Task 5 表中"在 retry.go 确认"等表述本身要求执行者验证，属数据采集步骤而非占位符。

**类型/命名一致性：** 各任务接口名、包头名、文档文件名与 Task 1-5 定义一致；SOP 编号（SOP 1-6）与引用一致；`AI_CHANGE_IMPACT §3` 与 `AI_TASK_HANDBOOK §0` 的相互引用对应。