# enhance 包间依赖与变更影响分析

> AI 在**修改任何包之前**阅读本文，确认影响面与回归范围。
> 依赖数据来自源码 import（grep 校验，非测试文件优先），与 DEPENDENCIES.md 不一致处以源码为准（差异见 [§6 与 DEPENDENCIES.md 的差异](#6-与-dependenciesmd-的差异)）。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md) · 同级：[AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md) · 修改接口 SOP 流程见 [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md) SOP 3。

## 1. 依赖方向总览

Boot Layer → Core Layer → Infrastructure Layer（详见 [DEPENDENCIES.md](../DEPENDENCIES.md)）。
`core/` 零外部依赖（仅标准库）；各层无条件只能单向依赖，禁止循环依赖（AGENTS.md §0.4）。

```
Boot Layer（boot/context/condition）          ── 应用启动、自动配置、条件装配
    ↓
Core Layer（core/config/event/lifecycle）     ── IoC 容器、配置、事件、生命周期
    ↓
Infrastructure Layer（web/cache/log/metrics…） ── 基础设施能力
```
> 包归属见 [AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md)（AGENTS.md §1.1 三层架构）。

**变更提示**：越靠近下层的包（尤其 `core/`、`boot/`、`config/`、`condition/`、`log/`）影响面越大；改它们先看 §3 反向影响表。

## 2. 包直接依赖表（正向：A 依赖 B 才能编译）

> 数据来源：各包非测试文件（`*.go` 非 `*_test.go`）中的真实 import 语句。

| 包 | 依赖的 enhance 包 | 说明 |
|----|------------------|------|
| actuator | actuator/admin, actuator/health, boot, condition, config/environment, core, metrics | 应用层聚合 |
| boot | boot/banner, condition, config, config/environment, context, core, core/registry, event, lifecycle | 启动核心 |
| condition | config/environment, spel | 条件判断（spel 见 `condition/extended.go`） |
| config | config/environment | 配置管理 |
| context | config/environment, config/refresh, core, core/registry, event, lifecycle | 应用上下文门面 |
| core | core/lifecycle, core/registry, core/scope | IoC 容器（仅标准库 + 自身子包） |
| event | retry | 事件驱动（deadletter 委托 retry） |
| metrics | boot, condition | 指标收集 |
| schedule | boot, condition, config/environment, core, log | 定时任务 |
| security | boot, condition, config/environment, core, log, security/authentication, security/authorization, security/filter | 安全框架 |
| testing | boot, config/environment, core, core/registry | 测试工具 |
| tracing | boot, condition, config/environment, core, log | 分布式追踪 |
| web | web/core, web/engine, web/mvc | Web 分层面板（类型别名） |
| async / audit / cache / devtools / email / exception / i18n / metadata / mq / openapi / resilience / retry / spel / tenant / validation / webtest | 无 | 独立包（不依赖任何 enhance 内部包） |
| log / lifecycle | 无 | 独立包（仅标准库） |
| observability | 无 | 非代码包（仅 README + 子目录） |

> 注 1：`web` **根包**仅依赖 `web/core、web/engine、web/mvc` 三个子包；`webtest` 仅出现在 `web/*_test.go` 中，`web/server` 不进 web 根包的依赖（但非测试子包如 `web/engine/stdlib`、`web/tls` 会使用它）。
> 注 2：`metrics` 的 `config/environment`、`core` 依赖仅出现在 `metrics/*_test.go`（`autoconfig_test.go`），非测试代码只依赖 `boot`、`condition`。
> 注 3：无内部依赖 ≠ 无下游，反向关系见 §3 反向影响表（如 `cache` 被 `starter/redis` 引用、`retry` 被 `event` 引用）。

## 3. 反向影响表（修改 X 会影响谁）

> 推导：遍历主模块全部非测试文件，凡 import 路径以 `enhance/X` 或 `enhance/X/…` 开头的包，都是 X 的下游。
> 表内仅列**框架内部**下游（含 `cmd/demo`）；`starter/`（第三方集成）、`examples/`（独立示例模块）与根包 `enhance.go` 在下表"备注"列单独标注。
> 回归测试范围 = 受影响的下游包 + X 自身。

| 修改的包 | 框架内受影响下游（需回归测试） | 备注（模块外部消费者） |
|----------|------------------------------|------------------------|
| core | actuator, boot, config, context, schedule, security, testing, tracing, web | starter/*（第三方集成适配，30 个导入 core）、examples/*、cmd/demo |
| boot | actuator, metrics, schedule, security, testing, tracing, web, cmd | 根包 enhance.go、starter/*、examples/* |
| config | actuator, boot, condition, context, schedule, security, testing, tracing | starter/*、examples/* |
| condition | actuator, boot, metrics, schedule, security, tracing, web, cmd | starter/*、examples/* |
| log | schedule, security, tracing, web（web/middleware、web/mvc、web/server） | starter/*、examples/* |
| event | boot, config（config/refresh/router.go）, context, cmd | examples |
| metrics | actuator | examples |
| context | boot | — |
| lifecycle | boot, context | — |
| cache | 框架内无（`cache` 零依赖零下游） | starter/redis、cmd/demo/cache-local |
| retry | event | — |
| spel | condition | — |
| resilience | 框架内无 | cmd/demo/resilience |
| validation | 框架内无 | cmd/demo（validation） |
| security | 框架内无（security 子包属自身） | starter/jwt、starter/casbin*、examples/* |
| tracing | 框架内无 | starter/chi、echo、fiber、gin、examples/* |
| actuator | 框架内无 | examples/* |
| web | 框架内无（web 各子包仅 web 自身消费） | cmd/demo/web-rest-api（web/mvc、web/server)、examples/*（web/core、web/engine） |
| schedule | 框架内无 | cmd/demo（shedule*） |
| testing | 无 | — |
| async / audit / devtools / email / exception / i18n / metadata / mq / openapi / tenant / webtest | 无下游，仅该包自身 | — |

**重要**：
- 修改 `core`、`boot`、`config`、`condition`、`log` 这类被多包引用的包时，**必须以 `go build ./...` + 受影响包 `go test` 全量回归**（见 §5），并优先走接口兼容（新增接口而非修改签名）。
- `web` 根包是纯类型别名（`web/doc.go`），改接口实际落在 `web/core`，改完需同步 `web/engine`、`web/mvc`、`web/server` 实现并跑 `cmd/demo/web-rest-api`（见 [AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md)）。
- 例外的"入口"文件（`enhance.go` 根包 import boot）改动 boot 时一并回归。

## 4. 自动配置执行顺序

> 值越小越先执行。新增 auto-config/starter 时必须从下表选择优先级（见 [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md) §新增自动配置）。

| 优先级常量（源码） | 值 | 层级 | 代表模块 |
|--------|------|------|----------|
| OrderPriorityInfrastructure | -3000 | 基础设施 | log、config、condition |
| OrderPriorityDataLayer | -2000 | 数据层 | cache、gorm、redis |
| OrderPriorityServiceDiscovery | -1800 | 服务发现 | consul、nacos、eureka |
| OrderPriorityAuthentication | -1500 | 认证层 | jwt、oauth2、ldap |
| OrderPriorityAuthorizationGorm | -1300 | 授权-GORM 适配 | casbin-gorm |
| OrderPriorityAuthorization | -1200 | 授权层 | casbin、rbac、abac |
| OrderPriorityTaskLayer | -500 | 任务层 | cron、asynq |
| OrderPrioritySecurityCore | -100 | 安全核心 | security（过滤器链、访问控制） |
| OrderPriorityWebLayer | 0 | Web 层 | web（HTTP 服务器、路由） |
| OrderPriorityMiddleware | 100 | 中间件层 | ratelimiter、cors、压缩 |
| OrderPriorityBusinessLayer | 1000 | 业务层 | event、schedule、async、mq |
| OrderPriorityMonitoringLayer | 2000 | 监控层 | metrics、actuator、tracing |

> 注：`boot/doc.go` 的包注释只列了其中 9 个（含 AuthorizationGorm，真正遗漏的是 ServiceDiscovery、Middleware、TaskLayer 三个），本文以 `boot/autoconfig.go` 的 12 个常量全集为准。

## 5. 修改任意包后的必做验证

1. `go build ./...` —— 全量编译（含受影响包）
2. 受影响包定向测试：`go test <受影响包路径列表>`（范围见 §3 反向影响表）
3. `go test -race ./...` 或至少受影响包 `-race`
4. `gofmt -l .` 无输出；修改导出符号时检查 godoc 注释
5. 若改了接口：按 [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md) §修改接口 做影响面排查与兼容处理
6. 修改了 `web/core`、`boot/autoconfig.go` 的 Order 常量等"关键枢纽"时，额外回归相关示例：`go build ./cmd/demo/...`、`go test ./starter/...`

## 6. 与 DEPENDENCIES.md 的差异

本文数据以源码为准，与 DEPENDENCIES.md（根目录）不一致处如下：

| 项 | DEPENDENCIES.md（旧） | 源码事实（本文采用） |
|----|----------------------|---------------------|
| 核心层优先级 | 有"核心层 -2500"（IoC/生命周期） | `boot/autoconfig.go` **无** `OrderPriorityCore` 常量；-2500 无对应枚举 |
| 执行顺序 | 9 个优先级层级 | 源码 12 个常量（多 ServiceDiscovery -1800、Middleware 100、TaskLayer -500） |
| condition 依赖 | 依赖 `boot` | 源码不 import boot；依赖 `config/environment` + `spel`（额外） |
| cache 依赖 | 依赖 `core` | 源码零内部依赖（依赖仅存在于文档/ioc 用法，代码层无） |
| starter/gorm 依赖 | 依赖 `log, config` | 实际依赖 boot、condition、config/environment、core、log |
| web 依赖 | 依赖 `boot, log` | web 根包非测试源码只依赖子包 `web/core、web/engine、web/mvc`；`web/server` 不进根包依赖（但被子包 `web/engine/stdlib`、`web/tls` 非测试使用） |
| actuator 依赖 | 无明细 | 依赖 actuator/admin、actuator/health、boot、condition、config/environment、core、metrics（event 仅测试文件） |

> 注：DEPENDENCIES.md 是 2026-07-29 的文档快照，描述偏设计意图（如"web 依赖 boot/log"反映 MVC Starter 约定）；实际编译依赖以本文档/源码为准。发现进一步偏差时，以源码为准并更新本文。

## 7. 一致性抽查记录

抽查日期 2026-09-16，覆盖前向依赖与反向影响，全部通过：

```bash
grep -rn 'enhance/core"' boot/*.go | head -2        # boot 依赖 core 属实（boot/doc.go, application.go…）✓
grep -rn 'enhance/config/environment"' condition/*.go | head -1  # condition 依赖 config/environment ✓
grep -n 'enhance/spel"' condition/extended.go       # condition 额外依赖 spel ✓
grep -hE '^\s+"github.com/xudefa/enhance/[^"]+"' metrics/*.go  # metrics 非测试仅 boot/condition ✓
grep -rn 'enhance/retry"' event/*.go                # event 依赖 retry 属实（deadletter.go）✓
grep -rn 'enhance/cache"' starter/redis/*.go        # cache 存在下游 starter/redis ✓
grep -rn 'github.com/xudefa/enhance/boot"' enhance.go  # 根包 enhance.go 依赖 boot ✓
```

> 数据核对日期：2026-09-16。源码 import 变更后请更新本文（正反向表用 §2/§3 描述的命令重跑）。