# enhance 常见 AI 任务操作手册（SOP）

> 本文给出 AI 在 enhance 上执行各类开发任务的**标准步骤**。每个 SOP 可直接照做。
> 规范依据：[AGENTS.md](../AGENTS.md) + [CODING_STYLE.md](../CODING_STYLE.md)。提交规范：每完成一个逻辑单元提交一次。
>
> 上一级文档：[AI_INDEX.md](AI_INDEX.md) · 影响面分析：[AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) · 代码导航：[AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md)

---

## 0. 通用前置检查（所有任务）

每次动手前必须逐项确认，未通过则停下来解决后再继续。

- [ ] 已读 [AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md) 确认目标包位置、文件职责、核心接口
- [ ] 已读 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) 确认目标包的影响面与回归范围（§3 反向影响表）
- [ ] 已读 [AGENTS.md](../AGENTS.md) 确认硬性规范：零外部依赖、doc.go 门面、依赖方向单向、接口命名 `er` 后缀
- [ ] 包级使用完整模块路径导入（如 `github.com/xudefa/enhance/core`），禁止相对导入
- [ ] 不引入任何第三方依赖（先确认有没有现成的接口/实现；第三方集成必须放 `starter/<name>/`，各自独立 go.mod）
- [ ] 依赖方向确认：Boot Layer → Core Layer → Infrastructure Layer，单向无环（AGENTS.md §0.4）
- [ ] 当前工作目录在 `go.work` 管理的多模块工作区中，必要时执行 `go work sync`

---

## SOP 1：新增一个包

> 适用场景：需要在 enhance 中添加全新的功能模块（如 `cache`、`resilience`）。

### 步骤

**1. 建目录**
创建包目录，包名小写无下划线，语义明确（如 `cache`、`retry`、`validation`）。

```bash
mkdir -p <pkg>
```

> 依据：AGENTS.md §2.1 包名小写。

**2. 创建 `doc.go` 门面文件**

`doc.go` 是包的唯一对外"门面"，**只含接口定义、类型定义和包文档，不含任何实现逻辑**（AGENTS.md §1.4）。

内容清单：
- 包级 godoc 注释（职责、设计思路、使用示例）
- 所有对外暴露的接口定义
- 公共结构体定义（不含方法实现）
- 类型别名、枚举常量、错误变量

文件长度限制：`doc.go` ≤ 500 行，超出时创建 `types.go` 分担类型定义。

> 依据：AGENTS.md §0.4、§1.4 doc.go 门面规范。

**3. 创建实现文件**

构造函数**返回接口类型**，不返回具体结构体指针。实现细节（结构体字段、私有方法）全部隐藏在实现文件中。

```go
// ✅ 正确
func NewCache() Cache {
    return &cache{...}
}

// ❌ 错误
func NewCache() *cache {
    return &cache{...}
}
```

> 依据：AGENTS.md §0.4 实现隐藏规范。

**4. 自动配置/Starter（如需）**

若新包需要条件装配，新建 `autoconfig.go`：
- 在 `init()` 中调用 `boot.RegisterAutoConfig`，传入条件如 `condition.OnProperty("xxx.enabled", "true")`
- 使用 `OnMissingBean` 避免重复注册
- 按 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) §4 选择合适的 OrderPriority
- 检查无循环依赖

> 依据：AGENTS.md §3.3 自动配置机制。

**5. 创建 `README.md`**

使用说明文档，包含：功能简介、快速上手代码示例、配置项说明。

**6. 创建测试文件**

表驱动测试 + `t.Parallel()`，覆盖正常路径和错误路径，覆盖率尽量 ≥ 80%。

> 详见 SOP 5（测试规范）。
> 依据：AGENTS.md §2.3 测试规范。

**7. 验证**

```bash
go build ./... && go test ./<pkg> && go test -race ./<pkg>
```

**8. 提交**

```bash
git add <pkg>/
git commit -m "feat: add <pkg> package"
```

---

## SOP 2：新增接口/类型

> 适用场景：需要在现有包中添加新的接口定义或公共类型。

### 步骤

**1. 确认归属包，定位 `doc.go`**

接口定义**必须放在 `doc.go`（或 `types.go`）中**，不放实现文件。

> 依据：AGENTS.md §1.4、§0.4。

**2. 命名规范**

- 接口使用 `er` 后缀或功能描述命名（如 `Reader`、`Logger`、`Cache`）
- **禁止** `I` 前缀（如 `IReader`、`ICache`）
- 接口方法数量 ≤ 5 个，超过则拆分为多个小接口通过组合形成大接口

> 依据：AGENTS.md §0.4、§2.1。

**3. 编写 godoc 注释**

每个方法必须有完整的 godoc 注释，包含：方法名、参数说明、返回值说明、错误语义。

```go
// Cache 提供键值缓存操作能力。
type Cache interface {
    // Get 根据 key 获取缓存值，未命中返回 nil 和 ErrCacheMiss。
    Get(ctx context.Context, key string) (any, error)

    // Set 将值写入缓存，ttl 为 0 表示使用默认过期时间。
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
}
```

> 依据：AGENTS.md §0.4 接口设计原则。

**4. 补齐实现**

现有实现类需新增方法以满足接口契约；新增方法的实现放对应的实现文件中。构造函数返回接口类型。

**5. 类型别名重新导出（如需跨包复用）**

若接口需要被多个包使用，采用类型别名重新导出模式：

```go
// web/doc.go —— 重新导出子包接口
type Router = mvc.Router
type Context = mvc.Context
```

> 依据：AGENTS.md §0.4 包交互模式一。

**6. 验证 + 提交**

```bash
go build ./...
go test ./<pkg>
go test -race ./<pkg>
git add <pkg>/
git commit -m "feat: add Xxx interface to <pkg>"
```

---

## SOP 3：修改现有接口

> 适用场景：需要修改已有接口的方法签名、新增方法、或移除方法。
> ⚠️ 此操作影响面大，务必先读 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md)。

### 步骤

**1. 影响面排查**

用 `grep` 找出全部实现方与调用方：

```bash
grep -rn '<接口名>' --include='*.go' .
```

重点区分：实现方（`implements` 该接口的结构体方法）和调用方（使用该接口作为参数/返回值的函数）。

> 依据：AI_CHANGE_IMPACT.md §3 反向影响表。

**2. 确定回归范围**

参考 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) §3 反向影响表，确定受影响的下游包。回归范围 = 所有受影响的下游包 + 当前包自身。

**3. 兼容性策略**

优先采用以下兼容方案，**避免直接删除/修改签名导致破坏性变更**：

- **加新方法**：在接口中新增方法，所有已有实现补上新方法的实现
- **加新接口**：将新方法放到新接口，通过类型断言判断是否实现
- **类型别名保留旧 API**：必要时用类型别名保留旧接口（参考 `boot.BootErrorStruct` 兼容做法）
- **新增接口而非修改签名**：尤其是 `core`、`boot` 等被广泛引用的包

> 依据：AGENTS.md §0.4 依赖方向单向规则。

**4. 更新所有实现方与调用方**

新增方法需在所有已有实现中补齐实现；修改签名需同步更新所有调用方。

**5. 验证**

```bash
go build ./...
go test <受影响包路径列表>
go test -race <受影响包路径列表>
```

范围示例（修改 `core` 时）：

```bash
go test ./core/... ./boot/... ./context/... ./config/... ./schedule/... ./security/... ./testing/... ./tracing/... ./web/... ./actuator/...
```

> 依据：AI_CHANGE_IMPACT.md §5 修改任意包后的必做验证。

**6. 提交**

```bash
git add <相关文件>
git commit -m "refactor: update <接口名> in <pkg> with backward compatibility"
```

---

## SOP 4：新增 auto-config / starter

> 适用场景：需要为框架添加自动配置类或第三方 Starter 集成。

### 步骤

**1. 实现配置类**

- 实现 `AutoConfiguration` 接口的 `Configure(ctx ApplicationContext)` 方法
- 或实现 `boot.Starter` 接口的 `Configure/Start/Stop` + `Dependencies()` 方法

> 依据：AGENTS.md §3.3 自动配置机制、§7.4 Starter 机制。

**2. 条件装配**

在 `init()` 中调用 `boot.RegisterAutoConfig`，配置条件：

- `condition.OnProperty("xxx.enabled", "true")`：配置属性存在且匹配时启用
- `condition.OnMissingBean`：当对应 Bean 不存在时才注册，避免重复注册

```go
func init() {
    boot.RegisterAutoConfig(
        &MyAutoConfiguration{},
        condition.OnProperty("my.enabled", "true"),
    )
}
```

> 依据：AGENTS.md §3.3 条件注解表。

**3. 选择 OrderPriority**

按 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) §4 执行顺序选择合适的优先级。值越小越先执行；确保新 starter 的执行顺序在依赖它的包之前。

- 基础设施层（log、config）：OrderPriorityInfrastructure
- 数据层（cache、redis）：OrderPriorityDataLayer
- Web 层：OrderPriorityWebLayer
- 业务层（event、schedule）：OrderPriorityBusinessLayer
- 监控层（metrics、actuator）：OrderPriorityMonitoringLayer

> 优先级常量值见 `boot/autoconfig.go` 源码，以源码为准。

**4. Starter 独立性**

- 第三方 Starter 放 `starter/<name>/` 目录
- 每个 Starter **自带独立 `go.mod`**，声明自身依赖
- Starter 间**禁止互相依赖**（每个 Starter 完全独立）
- Starter **不得依赖其他 Starter 包**

> 依据：AGENTS.md §0.1 零外部依赖、第三方集成放 starter。

**5. 示例放 `examples/<name>/`**

使用该 Starter 的示例放 `examples/<name>/`，独立 `go.mod`，展示完整使用流程。

**6. 验证 + 提交**

```bash
go build ./...
go test ./...
go test -race <受影响包>
go build ./cmd/demo/...
go test ./starter/...
git add starter/<name>/ examples/<name>/
git commit -m "feat: add <name> starter with auto-configuration"
```

---

## SOP 5：新增/修改测试

> 适用场景：编写新测试用例或修改已有测试。
> 规范依据：AGENTS.md §2.3 单元测试并发规范。
> 详细指南：[TESTING_GUIDE.md](TESTING_GUIDE.md)

### 步骤

**1. 表驱动模板**

所有测试必须使用表驱动风格，标准模板：

```go
func TestService_DoSomething(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "hello", "expected", false},
        {"empty input", "", "", true},
        {"special characters", "!@#$%", "expected", false},
        {"long input", "this is a very long input string", "expected", false},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            // Arrange
            instance := NewService()
            ctx := context.Background()

            // Act
            got, err := instance.DoSomething(ctx, tt.input)

            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("DoSomething(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("DoSomething(%q) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}
```

**2. 并发安全测试模板**

```go
func TestService_Concurrent(t *testing.T) {
    t.Parallel()

    instance := NewService()
    ctx := context.Background()

    const goroutines = 100
    errs := make(chan error, goroutines)

    for i := 0; i < goroutines; i++ {
        go func() {
            _, err := instance.DoSomething(ctx, "test")
            errs <- err
        }()
    }

    for i := 0; i < goroutines; i++ {
        if err := <-errs; err != nil {
            t.Errorf("Concurrent DoSomething() error = %v", err)
        }
    }
}
```

**3. Context 测试模板**

```go
func TestService_ContextCancellation(t *testing.T) {
    t.Parallel()

    instance := NewService()
    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    _, err := instance.DoSomething(ctx, "test")
    if err == nil {
        t.Error("DoSomething() expected error on cancelled context")
    }
}
```

**4. 覆盖范围**

每个测试必须覆盖：
- **正常路径**：典型输入，期望输出
- **错误路径**：无效输入、边界条件、nil/空值
- **边界**：零值、最大值、并发场景（如适用）
- **上下文**：context 取消、超时

**5. 运行测试**

```bash
go test -v ./<pkg>          # 查看详细输出，确认 PASS
go test -race ./<pkg>       # 检测数据竞争
```

> 依据：AGENTS.md §5.1 必须通过（阻塞项）。

**6. 提交**

```bash
git add <pkg>/*_test.go
git commit -m "test: add unit tests for <功能描述>"
```

---

## SOP 6：任务验收（提交前最后一关）

> 适用场景：任何代码变更完成后，提交前必须通过本验收流程。
> 所有命令必须**依次执行且全部通过**。

### 验收命令序列

```bash
# 1. 全量编译检查
go build ./...

# 2. 全量测试
go test ./...

# 3. 竞态检测（受影响包）
go test -race <受影响包路径列表>

# 4. 格式检查（无输出表示通过）
gofmt -l .

# 5. 静态分析
go vet ./...
```

> 依据：AGENTS.md §5.1 必须通过（阻塞项）、AI_CHANGE_IMPACT.md §5 修改任意包后的必做验证。

### 自查清单

验收命令全部通过后，再逐项自查：

- [ ] 导出符号均有 godoc 注释（接口、类型、函数、常量）
- [ ] 无多余 `else` 分支（if 已返回时不需要 else）
- [ ] 错误使用 `%w` 包装（`fmt.Errorf("描述: %w", err)`）
- [ ] 错误判断用 `errors.Is` / `errors.As`，禁止直接 `==` 比较
- [ ] 无裸 goroutine（必须用 errgroup 或 WaitGroup 管理）
- [ ] 文件 ≤ 500 行、函数 ≤ 50 行、接口方法 ≤ 5 个
- [ ] 不引入任何第三方依赖（核心包）
- [ ] 构造函数返回接口，不返回具体结构体指针
- [ ] 接口命名 `er` 后缀，禁止 `I` 前缀

> 依据：AGENTS.md §5.2 规范检查、§5.3 测试检查、§6.1 绝对禁止。

### 提交

```bash
git add <相关文件>
git commit -m "<类型>: <简洁描述>"
```

提交类型：`feat`（新功能）、`fix`（修复）、`refactor`（重构）、`test`（测试）、`docs`（文档）。
