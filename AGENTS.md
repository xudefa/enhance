# enhance 项目开发规范（AI 智能体必读）

> **重要**：本文档是 AI 智能体参与 enhance 项目开发时必须遵循的全局上下文和规范。所有代码生成、修改、审查都必须严格遵守本文档中的规则。
>
> **相关文档**：[代码风格指南](CODING_STYLE.md) | [架构设计](ARCHITECTURE.md) | [贡献指南](CONTRIBUTING.md)

---

## 0. 核心原则（必须遵守）

### 0.1 零外部依赖
- **核心框架仅使用 Go 标准库**，不引入任何第三方依赖
- **与第三方依赖解耦**：所有与第三方依赖的集成通过接口定义和适配器实现
- 第三方依赖必须在 `starter/` 包中实现，starter 包中，不允许依赖另外一个starter包，避免循环依赖
- **示例代码分类存放**：
  - 使用第三方依赖的示例放在 `examples/` 目录下，每个示例都有一个独立的go.mod文件，依赖不同的第三方依赖
  - 无第三方依赖的示例（核心包的使用示例）放在 `cmd/demo/` 目录下，无需go.mod文件，直接依赖核心包

### 0.2 Go 语言优先
- 参考 Spring Framework/Spring Boot 的设计哲学，但**不照搬 Java 语法**
- 遵循 Go 的惯用法（idiomatic Go），而非 Java 风格
- 使用 Go 泛型、接口、组合等特性，而非继承
- **禁止使用 Java 风格的动态类加载机制**：Go 没有 `Class.forName()`，所有类型必须在编译时确定

### 0.3 工程化框架
- enhance 是**工程化框架**，不是"轻量级框架"
- 提供完整的企业级特性：IoC、AOP、自动配置、Actuator 等
- 保持代码质量、可测试性、可维护性

### 0.4 依赖方向与接口隔离原则

> **核心思想**：包与包之间只能通过接口交互，实现细节必须隐藏在包内部。

#### 依赖方向规则
- **core 包零依赖**：核心 IoC 容器仅使用 Go 标准库
- **单向依赖**：依赖方向只能从上层向下层，禁止循环依赖
- **接口隔离**：各包通过接口定义与外部交互

#### 接口定义规范

**doc.go 集中管理**：
- 包内所有对外暴露的接口、类型别名、公共结构体定义统一放在 `doc.go` 文件中
- `doc.go` 文件不包含任何实现逻辑，仅包含接口定义、类型定义和包文档
- `doc.go` 超过 500 行时，创建 `types.go` 文件分担类型定义

**接口设计原则**：
- 接口命名使用 `er` 后缀或功能描述（如 `Reader`、`Logger`），禁止 `I` 前缀
- 接口方法数量遵循最小化原则，通常 1-5 个方法
- 接口方法必须有完整的 godoc 注释

#### 实现隐藏规范

**构造函数返回接口**：
```go
// ✅ 正确：构造函数返回接口类型
func NewEventBus() EventBus {
    return &eventBus{...}
}

// ❌ 错误：构造函数返回具体结构体指针
func NewEventBus() *eventBus {
    return &eventBus{...}
}
```

#### 包交互模式

**模式一：类型别名重新导出（推荐）**：
```go
// web/doc.go — 重新导出子包接口
type Router = mvc.Router
type Context = mvc.Context
```

**模式二：直接定义接口**：
```go
// cache/doc.go — 直接定义接口
type Cache interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
}
```

#### 禁止事项
- ❌ 禁止在 `doc.go` 中写实现逻辑
- ❌ 禁止调用方直接依赖实现结构体（应通过接口引用）
- ❌ 禁止核心包引入第三方依赖
- ❌ 禁止接口方法超过 10 个（应拆分）
- ❌ 禁止在接口命名中使用 `I` 前缀

---

## 1. 项目架构

### 1.1 三层架构

enhance 采用三层架构设计，职责清晰：

| 层级 | 职责 | 核心模块 |
|------|------|----------|
| **Boot Layer** | 应用启动、自动配置、Starter 管理 | `boot`, `condition`, `context` |
| **Core Layer** | IoC 容器、AOP、事件驱动、配置管理 | `core`, `aop`, `event`, `config`, `lifecycle` |
| **Infrastructure Layer** | Web、安全、监控、缓存、数据访问等基础设施 | `web`, `security`, `actuator`, `cache`, `schedule`, `log`, `metrics` |

### 1.2 包结构清单

> **说明**：每个包的核心接口都定义在 `doc.go` 文件中，实现细节隐藏在各自实现文件中。

#### 核心框架

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `core/` | IoC 容器（依赖注入、组件扫描、泛型 API） | `core.Container`, `core.Scope`, `core.BeanPostProcessor` |
| `aop/` | AOP 框架（5 种通知 + 切点匹配 + 代码生成） | `aop.Advice`, `aop.PointCut`, `aop.Advisor` |
| `boot/` | 应用启动器、自动配置、横幅、失败分析 | `boot.AutoConfiguration`, `boot.Starter` |
| `context/` | 应用上下文（聚合容器、环境、生命周期、事件） | `context.ApplicationContext` |
| `condition/` | 条件判断（OnProperty / OnBean / OnClass） | `condition.Condition` |
| `lifecycle/` | 生命周期阶段管理（7 个阶段） | `lifecycle.Lifecycle`, `lifecycle.Phase` |

#### 配置与环境

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `config/` | 配置管理（Config + Loader + Validator + Watch） | `config.Config`, `config.Loader` |
| `config/environment/` | 环境配置（分层 PropertySource + Profile） | `environment.Environment` |

#### 数据与缓存

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `cache/` | 缓存抽象（LRU + 内存缓存） | `cache.Cache` |

#### 网络与通信

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `web/` | HTTP 服务器/客户端、路由器、中间件 | `web.Router`, `web.Server` |
| `resilience/` | 弹性与容错（熔断器 + 负载均衡） | `resilience.Breaker`, `resilience.Selector` |

#### 可观测性

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `log/` | 日志抽象（Logger + slog 实现） | `log.Logger` |
| `metrics/` | 指标收集（Counter + Gauge + Registry） | `metrics.MeterRegistry` |
| `actuator/` | 运维端点（/health, /metrics, /env） | `actuator.Endpoint` |

#### 安全与验证

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `security/` | 安全框架（认证 + 授权 + 过滤器链） | `security.Authentication`, `security.HttpSecurity` |
| `validation/` | 数据验证（HTTP 请求验证） | `validation.RequestValidator` |
| `exception/` | 异常处理（全局异常 + 错误码） | `exception.Handler` |

#### 可靠性与调度

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `event/` | 事件驱动（发布/订阅 + 死信队列 + 事务绑定） | `event.EventBus` |
| `schedule/` | 定时任务调度（Cron + Scheduler） | `schedule.Scheduler` |
| `async/` | 异步执行器 | `async.Executor` |
| `retry/` | 重试机制 | `retry.Retryable` |
| `audit/` | 审计日志 | `audit.Auditor` |

#### 工具与扩展

| 包 | 说明 | 核心接口 |
|---|------|----------|
| `i18n/` | 国际化（MessageSource） | `i18n.MessageSource` |
| `spel/` | 表达式语言（SpEL 风格） | `spel.ExpressionParser` |
| `testing/` | 测试框架（TestRunner + Mock） | `testing.TestRunner` |

### 1.3 目录约定
- **核心包示例代码**：统一放在 `cmd/demo/` 目录，展示核心框架功能，无需第三方依赖
- **第三方依赖集成示例**：统一放在 `examples/` 目录，展示与第三方库的集成
- **文档**：每个模块包含 `README.md` 说明文档
- **测试文件**：与源码同目录，`*_test.go` 命名
- **命令工具**：`cmd/` 目录（如 `cmd/goaop/` 代码生成器）

### 1.4 doc.go 文件组织规范（必须遵守）

> **核心规则**：`doc.go` 是包的"门面文件"，只包含接口定义、类型定义和包文档，不包含任何实现逻辑。

#### 内容清单

**必须包含**：
- 包级 godoc 注释
- 所有对外暴露的接口定义
- 公共结构体定义（不含方法实现）
- 类型别名、枚举常量、错误变量

**禁止包含**：
- ❌ 任何函数实现（构造函数除外）
- ❌ 接口方法的具体实现
- ❌ 业务逻辑代码

#### 文件拆分规则

```
cache/
├── doc.go          # 接口定义 + 公共类型（≤ 500 行）
├── lru.go          # LRUCache 实现
├── builder.go      # Builder 实现
└── cache_test.go   # 测试文件
```

**拆分规则**：
- `doc.go` ≤ 500 行：所有类型定义放在 `doc.go`
- `doc.go` > 500 行：创建 `types.go` 分担部分类型定义
- `types.go` > 500 行：考虑拆分为多个子包

---

## 2. 编码规范

> **详细编码规范请参阅**：[代码风格指南](CODING_STYLE.md)

### 2.1 核心规则速查

| 规则 | 要求 | 详见 |
|------|------|------|
| 命名规范 | 包名小写、导出标识符大写驼峰、错误变量 `Err` 前缀、接口 `er` 后缀 | [CODING_STYLE.md §2](CODING_STYLE.md#2-命名规范) |
| 导入规范 | 标准库 → 项目内部包，禁止相对导入和未使用导入 | [CODING_STYLE.md §3.3](CODING_STYLE.md#33-导入规范) |
| 函数式选项 | 优先使用选项模式而非配置结构体 | [CODING_STYLE.md §6.4](CODING_STYLE.md#64-函数式选项模式) |
| 控制流 | 早期返回、消除多余 `else`、`switch` 优于 `if-else` 链 | [CODING_STYLE.md §5](CODING_STYLE.md#5-控制流) |
| 函数设计 | ≤ 50 行、≤ 4 参数、`context` 作为第一个参数 | [CODING_STYLE.md §6](CODING_STYLE.md#6-函数设计) |
| 错误处理 | 不忽略错误、使用 `%w` 包装、`errors.Is/As` 判断 | [CODING_STYLE.md §8](CODING_STYLE.md#8-错误处理) |
| 并发安全 | 使用 `errgroup` 管理 goroutine、明确通道方向 | [CODING_STYLE.md §9](CODING_STYLE.md#9-并发编程) |
| 接口设计 | 小接口（≤ 5 方法）、通过组合形成大接口 | [CODING_STYLE.md §10](CODING_STYLE.md#10-接口设计) |
| 测试风格 | 表驱动测试、覆盖率 ≥ 80%、`t.Parallel()` | [CODING_STYLE.md §12](CODING_STYLE.md#12-测试风格) |

### 2.2 文件与代码长度限制

| 限制 | 值 |
|------|-----|
| 单个文件 | ≤ 500 行 |
| 单个函数 | ≤ 50 行 |
| 单个类型方法 | ≤ 200 行 |
| 接口方法 | ≤ 5 个 |

### 2.3 单元测试并发规范（必须遵守）

> **目标**：通过并发执行测试用例，显著减少单测运行时间，提升开发效率。

#### 基本原则
- **所有独立测试函数必须使用 `t.Parallel()`**
- **表驱动测试的子测试也应支持并发**
- **避免全局状态竞争**

#### 标准写法

**独立测试函数**：
```go
func TestService_DoSomething(t *testing.T) {
	t.Parallel() // 必须在函数开头调用
	// 测试逻辑...
}
```

**表驱动测试**：
```go
func TestService_Validate(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "hello", false},
		{"invalid", "", true},
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

#### 禁止并发的场景
- 测试用例修改了全局变量或包级状态
- 测试用例操作了共享资源且未加锁保护
- 测试用例依赖特定的执行顺序

---

## 3. 领域特定规范

### 3.1 IoC 容器

```go
// 创建容器
c := core.New()
core.EnableFieldTag(true)

// Bean 注册（泛型 API）
c.Register(reflect.TypeOf(&MyService{}), core.FactoryOf[MyService](func(c core.Container) (*MyService, error) {
    return &MyService{}, nil
}))

// 获取 Bean
svc := core.MustGetBean[*MyService](c)
```

### 3.2 AOP 框架

```go
// 通知类型：Before, After, Around, AfterReturning, AfterThrowing
// 切点匹配：MatchByName, MatchByPrefix, MatchByRegex

// 织入流程
weaver := aop.NewWeaver()
weaver.AddAspects(aspects...)
proxy := weaver.Weave(target)
```

**关键规则**：
- Around 通知**必须**调用 `proceed` 使调用链继续
- 多个通知通过 `aop.WithOrder(n)` 排序，值越小优先级越高

### 3.3 自动配置机制

```go
func init() {
    boot.RegisterAutoConfig(
        &CircuitAutoConfiguration{},
        condition.OnProperty("circuit.enabled", "true"),
    )
}
```

**条件注解**：

| 条件 | 说明 |
|------|------|
| `OnProperty` | 配置属性存在且匹配 |
| `OnBean` | Bean 存在 |
| `OnMissingBean` | Bean 不存在 |
| `OnClass` | 类型存在 |
| `OnProfile` | Profile 匹配 |

### 3.4 应用启动

```go
func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProfiles("dev"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    app.Start()
    app.WaitForSignal()
}
```

**配置优先级**：命令行参数 > 环境变量 > 配置文件 > 默认值

---

## 4. AI 工作流指南

### 4.1 代码生成流程

1. **理解需求**：确认用户意图，明确功能边界
2. **查阅架构**：参考本文档第 1 节，确定功能所属模块
3. **遵循规范**：遵守 [CODING_STYLE.md](CODING_STYLE.md) 中的编码规范
4. **生成代码**：按照文件内顺序组织代码
5. **编写测试**：使用表驱动测试，覆盖正常路径和错误路径
6. **自检清单**：完成第 5 节的检查清单

### 4.2 代码修改流程

1. **定位代码**：找到需要修改的文件和函数
2. **理解上下文**：阅读相关代码，理解现有逻辑
3. **最小修改**：只修改必要的部分，保持原有风格
4. **更新测试**：同步更新相关测试用例
5. **验证编译**：确保代码编译通过

### 4.3 代码审查要点

- [ ] 是否遵循命名规范
- [ ] 是否使用函数式选项模式
- [ ] 是否早期返回，无多余 else
- [ ] 导出类型是否有文档注释
- [ ] 错误是否已正确处理
- [ ] 并发是否安全
- [ ] 接口是否小而精
- [ ] 是否使用表驱动测试
- [ ] 测试是否使用 `t.Parallel()` 支持并发

---

## 5. 代码质量检查清单

AI 在生成代码后必须自检以下项目：

### 5.1 必须通过
- [ ] 代码编译通过（`go build ./...`）
- [ ] 测试通过（`go test ./...`）
- [ ] 无数据竞争（`go test -race ./...`）
- [ ] 代码已格式化（`go fmt ./...`）
- [ ] 依赖已修复（`go mod tidy`）

### 5.2 规范检查
- [ ] 遵循命名规范
- [ ] 导入分组正确
- [ ] 使用函数式选项模式
- [ ] 早期返回，无多余 else
- [ ] 导出类型有文档注释
- [ ] 错误已正确处理
- [ ] 并发安全
- [ ] 接口小而精
- [ ] 文件不超过 500 行
- [ ] 函数不超过 50 行

### 5.3 测试检查
- [ ] 使用表驱动测试
- [ ] 覆盖正常路径和错误路径
- [ ] 测试覆盖率 >= 80%
- [ ] 测试函数使用 `t.Parallel()` 支持并发运行
- [ ] 表驱动子测试也使用 `t.Parallel()`
- [ ] 无全局状态竞争

---

## 6. 禁止事项

### 6.1 绝对禁止
- ❌ 框架引入外部依赖
- ❌ 使用相对导入
- ❌ 忽略错误返回值
- ❌ 使用全局变量（除非明确设计）
- ❌ 照搬 Java 语法（如 getter/setter 模式）
- ❌ 使用 `else` 分支（当 `if` 已返回时）
- ❌ 裸 goroutine（不使用 errgroup 或 WaitGroup）
- ❌ 直接比较错误（必须使用 `errors.Is/As`）

### 6.2 强烈不建议
- ⚠️ 过度使用反射
- ⚠️ 过深的继承层次（Go 使用组合）
- ⚠️ 超过 4 层的嵌套
- ⚠️ 魔法数字（使用常量）
- ⚠️ 过长的函数（> 50 行）
- ⚠️ 过度泛型（> 2 个类型参数）
- ⚠️ 大接口（> 5 个方法）

---

## 7. 参考文档

- [架构设计文档](ARCHITECTURE.md)
- [README.md](README.md)
- [贡献指南](CONTRIBUTING.md)
- [代码风格指南](CODING_STYLE.md)