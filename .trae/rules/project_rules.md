# enhance 项目规则（Trae IDE AI 助手必读）

> 本文件是 Trae IDE AI 助手参与 enhance 项目开发时的核心上下文。所有代码生成、修改、审查必须严格遵守。

---

## 项目概述

enhance 是一个 Go 企业级框架（参考 Spring Framework/Spring Boot 设计），526+ Go 文件、30+ 顶层包。三层架构：Boot Layer → Core Layer → Infrastructure Layer。

- **模块路径**：`github.com/xudefa/enhance`
- **Go 版本**：1.25+
- **核心原则**：零外部依赖（核心包）、doc.go 门面、依赖方向单向、Go 惯用法优先

---

## 硬性规范（违反即拒收）

### 零外部依赖
- 核心框架包（core/boot/context/condition/lifecycle/config/event/web/security/cache/schedule/log/metrics/actuator/tracing/mq/resilience/validation/exception/tenant/email/async/audit/retry/i18n/spel/testing/webtest/openapi/metadata/devtools）**禁止引入任何第三方依赖**
- 第三方集成必须在 `starter/<name>/` 中实现，各自独立 go.mod
- starter 包之间**禁止互相依赖**
- 使用第三方依赖的示例放 `examples/`，核心示例放 `cmd/demo/`

### doc.go 门面
- 包内所有对外暴露的接口、类型别名、公共结构体定义统一放在 `doc.go`
- `doc.go` **禁止包含任何实现逻辑**
- 构造函数**必须返回接口类型**，不返回具体结构体指针
- 接口命名 `er` 后缀（Reader/Logger/Cache），**禁止 `I` 前缀**
- 接口方法 ≤5 个，超过时拆分为小接口组合
- `doc.go` 超过 500 行时创建 `types.go` 分担类型定义

### 依赖方向
- Boot Layer → Core Layer → Infrastructure Layer（单向，禁止循环依赖）
- `core/` 零外部依赖（仅标准库 + 自身子包）
- starter 包之间禁止互相依赖

### Go 惯用法
- 早期返回，避免深层嵌套
- 函数式选项模式（WithXxx）用于配置
- 组合优于继承，泛型优先于反射
- `context.Context` 作为第一个参数
- 错误包装 `fmt.Errorf("...: %w", err)`，判断用 `errors.Is/As`，**禁止 `==` 比较错误**
- **禁止裸 goroutine**，使用 errgroup 或 WaitGroup
- 文件：生产 ≤500 行，测试 ≤1000 行
- 函数：≤80 行（不含注释和空行）
- 算法嵌套深度：≤17 层
- 算法复杂度：时间与空间复杂度不能同时 ≥ O(n²)

### 测试规范
- 表驱动测试 + `t.Parallel()`（测试函数和子测试都必须加）
- 覆盖正常路径和错误路径，覆盖率 ≥80%
- 测试文件与被测文件一一对应：`a.go` → `a_test.go`
- 禁止将同一文件的单测分散到多个测试文件

---

## 命名规范

| 类型 | 规则 | 正确 | 错误 |
|------|------|------|------|
| 包名 | 全小写无下划线 | `cache` | `Cache`, `user_service` |
| 导出标识符 | 大写驼峰 | `UserID`, `GetUser` | `Get_User`, `userId` |
| 非导出标识符 | 小写驼峰 | `userID`, `getUser` | `get_user` |
| 常量 | 大写驼峰 | `MaxConnections` | `MAX_CONNECTIONS` |
| 错误变量 | `Err` 前缀 | `ErrNotFound` | `NotFoundErr` |
| 接口 | `er` 后缀 | `Reader`, `Logger` | `IReader`, `ReaderInterface` |
| 测试函数 | `Test功能_条件_期望` | `TestContainer_Get_NotFound` | `TestContainer1` |
| 布尔变量 | `is/has/can` 前缀 | `isValid`, `hasPermission` | `valid`, `permission` |
| 缩写词 | 导出全大写/非导出小写 | `HTTPClient`, `httpClient` | `HttpClient`, `httpclient` |

---

## 包结构与职责速查

| 包 | 职责 | 核心接口 | 依赖 |
|---|------|----------|------|
| `core/` | IoC 容器（DI、泛型API） | `Container` | 仅标准库+子包 |
| `boot/` | 应用启动、自动配置、失败分析 | `Application`, `AutoConfiguration` | core, config, event, condition, lifecycle |
| `context/` | 应用上下文门面 | `ApplicationContext` | core, config, event, lifecycle |
| `condition/` | 条件化注册 | `Condition` | config/environment, spel |
| `config/` | 配置管理 | `Config`, `Loader` | config/environment |
| `event/` | 事件驱动 | `EventBus` | retry |
| `cache/` | 缓存抽象 | `Cache` | 无 |
| `web/` | HTTP 服务器/路由/中间件 | `Server`, `Router` | web子包 |
| `security/` | 安全框架 | `Authentication`, `HttpSecurity` | boot, condition, core, log |
| `log/` | 日志抽象 | `Logger` | 无 |
| `metrics/` | 指标收集 | `MeterRegistry` | boot, condition |
| `actuator/` | 运维端点 | `Endpoint` | boot, core, metrics |
| `schedule/` | 定时任务 | `Scheduler` | boot, condition, core, log |
| `resilience/` | 弹性容错 | `Breaker`, `Balancer` | 无 |
| `validation/` | 数据验证 | `Validator` | 无 |
| `exception/` | 异常处理 | `ExceptionHandler` | 无 |
| `lifecycle/` | 生命周期管理 | `LifecycleManager` | 无 |

---

## 高频避坑

1. **接口定义写进实现文件** → 必须在 doc.go（ADR-005）
2. **构造函数返回具体类型** → 必须返回接口（ADR-006）
3. **核心包 import 第三方库** → 必须放 starter/（ADR-004）
4. **用 `==` 比较错误** → 必须用 errors.Is/As
5. **改接口直接删方法/改签名** → 先做影响分析（读 AI_CHANGE_IMPACT.md）
6. **新 starter 依赖其他 starter** → starter 互不依赖
7. **测试未加 t.Parallel()** → 测试函数和子测试都必须加
8. **事件类型字符串拼错** → 事件通过 Type() string 路由，字符串需全局一致（ADR-009）
9. **Cache.Get 未命中判断** → 返回 ErrNotFound，不要用零值判断
10. **Cron 表达式缺秒位** → 6 字段 Spring 风格：`秒 分 时 日 月 周`
11. **函数超过 80 行（不含注释）** → 提取子函数拆分
12. **嵌套深度超过 17 层** → 早期返回/卫语句消除嵌套
13. **时间+空间同时 ≥ O(n²)** → 优化算法或使用空间换时间/时间换空间策略

---

## 修改影响面（改前必读）

| 修改的包 | 受影响下游（需回归测试） |
|----------|------------------------|
| `core` | actuator, boot, config, context, schedule, security, testing, tracing, web |
| `boot` | actuator, metrics, schedule, security, testing, tracing, web, cmd |
| `config` | actuator, boot, condition, context, schedule, security, testing, tracing |
| `condition` | actuator, boot, metrics, schedule, security, tracing, web, cmd |
| `log` | schedule, security, tracing, web |
| `event` | boot, config, context, cmd |

---

## 自动配置优先级

| 优先级常量 | 值 | 典型组件 |
|-----------|-----|---------|
| `OrderPriorityInfrastructure` | -3000 | log, config, condition |
| `OrderPriorityDataLayer` | -2000 | cache, gorm, redis |
| `OrderPriorityServiceDiscovery` | -1800 | consul, nacos |
| `OrderPriorityAuthentication` | -1500 | jwt, oauth2 |
| `OrderPriorityAuthorization` | -1200 | casbin, rbac |
| `OrderPriorityTaskLayer` | -500 | cron, asynq |
| `OrderPrioritySecurityCore` | -100 | security |
| `OrderPriorityWebLayer` | 0 | web |
| `OrderPriorityBusinessLayer` | 1000 | event, schedule |
| `OrderPriorityMonitoring` | 2000 | metrics, actuator |

---

## AI 任务 SOP

### 新增包
1. 建目录 → 创建 doc.go（只放定义） → 创建实现文件（构造函数返回接口） → 表驱动测试 → 验证

### 修改接口
1. 读 AI_CHANGE_IMPACT.md → 确认无破坏性变更 → 改后更新 API_CHANGELOG.md → 检查依赖方向

### 新增 auto-config/starter
1. 参考 boot 包既有 auto-config → starter 互不依赖 → 示例放 examples/

### 编写测试
1. 表驱动 + t.Parallel() → 覆盖正常+错误路径 → 参考 docs/TESTING_GUIDE.md

---

## 验证命令（任务完成前必做）

```bash
# 完整验证
make ai-verify

# 快速检查
make ai-quick

# 质量评分（≥80）
make quality-score

# godoc 覆盖率
make godoc-coverage

# 核心检查
go build ./...
go test ./...
go test -race <affected-packages>
go fmt ./...
go vet ./...
```

---

## 文档导航

| 文档 | 用途 |
|------|------|
| `docs/AI_INDEX.md` | AI 维护导航总入口 |
| `docs/AI_NAVIGATION_MAP.md` | 逐包代码导航 |
| `docs/AI_CHANGE_IMPACT.md` | 修改影响面分析 |
| `docs/AI_TASK_HANDBOOK.md` | 任务 SOP 手册 |
| `docs/AI_PITFALLS.md` | 避坑指南 |
| `docs/AI_QUICK_REF.md` | 快速参考卡 |
| `docs/ADR.md` | 架构决策记录 |
| `docs/TESTING_GUIDE.md` | 测试规范 |
| `AGENTS.md` | 全局开发规范 |
| `ARCHITECTURE.md` | 架构设计 |
| `CODING_STYLE.md` | 代码风格 |