# enhance 代码风格指南

> 本文档定义 enhance 项目的代码风格规范。所有代码编写、审查、重构都应遵循本指南。
> 
> **相关文档**：[开发规范（AI 必读）](AGENTS.md) | [贡献指南](CONTRIBUTING.md) | [架构设计](ARCHITECTURE.md)

---

## 目录

- [1. 核心原则](#1-核心原则)
- [2. 命名规范](#2-命名规范)
- [3. 代码组织](#3-代码组织)
- [4. 注释与文档](#4-注释与文档)
- [5. 控制流](#5-控制流)
- [6. 函数设计](#6-函数设计)
- [7. 数据结构](#7-数据结构)
- [8. 错误处理](#8-错误处理)
- [9. 并发编程](#9-并发编程)
- [10. 接口设计](#10-接口设计)
- [11. 泛型使用](#11-泛型使用)
- [12. 测试风格](#12-测试风格)
- [13. 性能考量](#13-性能考量)
- [14. 代码审查清单](#14-代码审查清单)

---

## 1. 核心原则

### 1.1 设计哲学

| 原则 | 说明 | 示例 |
|------|------|------|
| 清晰优于巧妙 | 代码应该易于理解和维护 | 避免一行写完复杂逻辑 |
| 简单优于复杂 | 优先选择简单直接的实现 | 用 `if` 而非反射实现简单判断 |
| 可读性第一 | 代码首先是给人阅读的 | 变量名用 `userCount` 而非 `uc` |
| 零外部依赖 | 核心框架仅使用 Go 标准库 | 不引入第三方库 |

### 1.2 Go 惯用法

- 遵循 [Effective Go](https://go.dev/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 参考 Spring 的设计哲学，但**不照搬 Java 语法**
- 使用组合而非继承，使用接口而非具体类型

---

## 2. 命名规范

### 2.1 包命名

| 规则 | 说明 | 示例 |
|------|------|------|
| 全部小写 | 不使用大写或混合 | `container` ✅, `Container` ❌ |
| 无下划线 | 不使用下划线分隔 | `userservice` ✅, `user_service` ❌ |
| 简洁语义 | 包名应简短且有意义 | `core`, `aop`, `boot` |
| 目录一致 | 包名与最内层目录名一致 | `user-service/` → `package userservice` |

### 2.2 标识符命名

| 类型 | 规则 | 正确示例 | 错误示例 |
|------|------|----------|----------|
| 导出标识符 | 大写驼峰 | `UserID`, `GetUser` | `Get_User`, `userId` |
| 非导出标识符 | 小写驼峰 | `userID`, `getUser` | `get_user`, `Getuser` |
| 常量 | 大写驼峰 | `MaxConnections` | `MAX_CONNECTIONS` |
| 错误变量 | `Err` 前缀 | `ErrNotFound` | `NotFoundErr`, `errNotFound` |
| 接口 | `er` 后缀或功能描述 | `Reader`, `Logger` | `IReader`, `ReaderInterface` |
| 测试函数 | `Test功能_条件_期望` | `TestContainer_Get_NotFound` | `TestContainer1` |
| 布尔变量 | `is/has/can` 前缀 | `isValid`, `hasPermission` | `valid`, `permission` |

### 2.3 缩写词规范

| 缩写 | 全拼 | 使用场景 |
|------|------|----------|
| `ID` | Identifier | `UserID`, `getID()` |
| `URL` | Uniform Resource Locator | `ServerURL`, `parseURL()` |
| `HTTP` | Hypertext Transfer Protocol | `HTTPClient`, `httpServer` |
| `JSON` | JavaScript Object Notation | `JSONParser`, `marshalJSON()` |
| `SQL` | Structured Query Language | `SQLBuilder`, `execSQL()` |
| `API` | Application Programming Interface | `APIGateway`, `apiEndpoint` |
| `DB` | Database | `DBConnection`, `dbPool` |
| `Config` | Configuration | `AppConfig`, `loadConfig()` |

**导出时保持缩写全大写**：`HTTPClient` 而非 `HttpClient`
**非导出时小写**：`httpClient` 而非 `httpclient`

### 2.4 文件命名

- Go 源文件使用小写蛇形命名：`container.go`, `bean_factory.go`
- 测试文件以 `_test.go` 结尾：`container_test.go`
- 接口定义文件可单独为 `doc.go`
- 错误定义文件可单独为 `errors.go`

---

## 3. 代码组织

### 3.1 文件内顺序

```go
// 1. 包文档注释
// Package core 提供 IoC 容器实现...
package core

// 2. import 声明
import (
    "context"
    "fmt"

    "github.com/xudefa/enhance/log"
)

// 3. 常量定义
const DefaultTimeout = 30 * time.Second

// 4. 变量声明（含哨兵错误）
var ErrNotFound = errors.New("not found")

// 5. 类型定义（interface → struct → type alias）
type Container interface { ... }
type DefaultContainer struct { ... }
type BeanDefinition = map[string]any

// 6. 构造函数（返回接口类型）
func NewContainer() Container { ... }

// 7. 公共方法（按字母或逻辑分组）
func (c *DefaultContainer) Get(name string) (any, error) { ... }
func (c *DefaultContainer) Register(name string, def Definition) error { ... }

// 8. 私有方法
func (c *DefaultContainer) resolve(name string) (any, error) { ... }

// 9. 辅助函数
func validateName(name string) error { ... }
```

### 3.2 文件长度限制

| 限制 | 值 | 说明 |
|------|-----|------|
| 单个文件 | ≤ 500 行 | 超过时拆分到多个文件 |
| 单个函数 | ≤ 50 行 | 超过时提取子函数 |
| 单个类型方法 | ≤ 200 行 | 超过时考虑拆分类型 |

### 3.3 doc.go 文件规范

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

- `doc.go` ≤ 500 行：所有类型定义放在 `doc.go`
- `doc.go` > 500 行：创建 `types.go` 文件分担部分类型定义

### 3.4 导入规范

```go
import (
    // 第一组：标准库（按字母排序）
    "context"
    "fmt"
    "sync"
    "time"

    // 第二组：项目内部包（按字母排序）
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/log"
)
```

**禁止事项**：
- ❌ 禁止相对导入（如 `../foo`）
- ❌ 禁止未使用的导入
- ❌ 禁止核心框架引入外部依赖
- ❌ 禁止使用 `_` 导入（除非明确需要 `init()` 副作用）

---

## 4. 注释与文档

### 4.1 必须注释的场景

| 场景 | 要求 | 示例 |
|------|------|------|
| 导出类型 | 必须有 godoc | `// Container 是 IoC 容器接口...` |
| 导出函数 | 必须有 godoc | `// Register 注册一个 Bean 定义...` |
| 复杂逻辑 | 处理步骤 ≥ 3 | 说明执行流程 |
| 技术决策 | 非显而易见的选择 | 说明"为什么这样做" |
| 魔法数字 | 非常量的数值 | `// 3 是重试次数的经验值...` |
| 已知限制 | 当前实现的不足 | `// 注意：此实现不是线程安全的...` |

### 4.2 包注释

```go
// Package core 提供了一个类型安全的依赖注入（DI）容器实现，灵感来自 Spring Framework 的 IoC 容器。
//
// # 核心功能
//
//   - 编译期类型安全：通过泛型函数 Register[T]/Get[T] 在编译期检查 Bean 类型
//   - 零反射注册：用户 API 层完全避免反射，仅在容器内部使用 reflect.Type 存储类型信息
//   - 函数式依赖：通过工厂函数显式声明依赖关系
//   - 作用域管理：支持单例（Singleton）和原型（Prototype）作用域
//
// # 快速开始
//
//	container := core.NewContainer()
//	core.Register[*UserService](container,
//	    core.WithFactory[*UserService](func(c ...any) (any, error) {
//	        return &UserService{}, nil
//	    }),
//	)
//	svc := core.MustGet[*UserService](container, "")
//
// # 设计原则
//
// 核心框架零外部依赖，仅使用 Go 标准库。
package core
```

### 4.3 函数注释模板

```go
// CalculateDiscount 计算应用分级折扣后的最终价格。
// 折扣根据订单数量逐步应用：每个等级解锁额外的百分比减免。
// 如果数量无效或基础价格在应用折扣后会导致负值，则返回错误。
//
// 参数:
//   - basePrice: 任何折扣前的原始价格（必须为非负数）
//   - quantity: 订单的数量（必须为正数）
//   - tiers: 按最小数量阈值排序的折扣等级切片
//
// 返回最终折扣价格，四舍五入到小数点后两位。
// 如果 basePrice 为负数，返回 ErrInvalidPrice。
// 如果 quantity 为零或负数，返回 ErrInvalidQuantity。
//
// 示例:
//
//	tiers := []DiscountTier{
//	    {MinQuantity: 10, PercentOff: 5},
//	    {MinQuantity: 50, PercentOff: 15},
//	}
//	finalPrice, err := CalculateDiscount(100.00, 75, tiers)
func CalculateDiscount(basePrice float64, quantity int, tiers []DiscountTier) (float64, error) {
    // 实现代码
}
```

### 4.4 注释语言

- **使用中文注释**
- 技术术语保留英文（IoC, AOP, Bean, DI）
- 注释应说明"为什么这样做"而非"做了什么"

### 4.5 特殊标记

| 标记 | 说明 | 示例 |
|------|------|------|
| `TODO` | 待完成 | `// TODO: 实现缓存淘汰策略` |
| `FIXME` | 已知问题 | `// FIXME: 并发情况下可能丢失更新` |
| `NOTE` | 重要说明 | `// NOTE: 此方法不是线程安全的` |
| `DEPRECATED` | 已废弃 | `// DEPRECATED: 使用 NewContainer 替代` |

---

## 5. 控制流

### 5.1 早期返回（优先处理错误和边界条件）

```go
// ✅ 正确：早期返回
func process(data []byte) (*Result, error) {
    if len(data) == 0 {
        return nil, errors.New("empty data")
    }

    if !isValid(data) {
        return nil, errors.New("invalid data format")
    }

    parsed, err := parse(data)
    if err != nil {
        return nil, fmt.Errorf("parse failed: %w", err)
    }

    return transform(parsed), nil
}

// ❌ 错误：深层嵌套
func process(data []byte) (*Result, error) {
    if len(data) > 0 {
        if isValid(data) {
            parsed, err := parse(data)
            if err == nil {
                return transform(parsed), nil
            }
            return nil, err
        }
        return nil, errors.New("invalid")
    }
    return nil, errors.New("empty")
}
```

### 5.2 消除不必要的 `else`

```go
// ✅ 正确：无 else
func getStatus(code int) string {
    if code == 200 {
        return "OK"
    }
    if code == 404 {
        return "Not Found"
    }
    if code == 500 {
        return "Internal Server Error"
    }
    return "Unknown"
}

// ❌ 错误：多余 else
func getStatus(code int) string {
    if code == 200 {
        return "OK"
    } else if code == 404 {
        return "Not Found"
    } else {
        return "Unknown"
    }
}
```

### 5.3 命名布尔变量

```go
// ✅ 正确：命名布尔变量
isAdmin := user.Role == RoleAdmin
isOwner := resource.OwnerID == user.ID
isPublicVerified := resource.IsPublic && user.IsVerified

if isAdmin || isOwner || isPublicVerified {
    allowAccess()
}

// ❌ 错误：复杂条件直接判断
if user.Role == RoleAdmin || resource.OwnerID == user.ID || (resource.IsPublic && user.IsVerified) {
    allowAccess()
}
```

### 5.4 `switch` 优于 `if-else` 链

```go
// ✅ 正确：switch
switch status {
case StatusPending:
    handlePending()
case StatusApproved:
    handleApproved()
case StatusRejected:
    handleRejected()
default:
    handleUnknown()
}

// ❌ 错误：if-else 链
if status == StatusPending {
    handlePending()
} else if status == StatusApproved {
    handleApproved()
} else if status == StatusRejected {
    handleRejected()
} else {
    handleUnknown()
}
```

---

## 6. 函数设计

### 6.1 函数长度与职责

| 指标 | 限制 | 说明 |
|------|------|------|
| 函数行数 | ≤ 50 行 | 超过时提取子函数 |
| 参数数量 | ≤ 4 个 | 超过时使用选项模式 |
| 返回值数量 | ≤ 3 个 | 通常为 `(result, error)` |
| 圈复杂度 | ≤ 10 | 超过时拆分逻辑 |

### 6.2 参数顺序

```go
// 参数顺序：context → 主要参数 → 选项
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    // ...
}
```

| 位置 | 类型 | 说明 |
|------|------|------|
| 第 1 位 | `context.Context` | 请求上下文 |
| 第 2 位 | 主要参数 | 业务必需参数 |
| 第 3 位 | 可选参数 | 选项模式或默认值参数 |

### 6.3 接收器选择

| 接收器类型 | 使用场景 | 示例 |
|-----------|---------|------|
| 值接收器 | 小型、不可变类型 | `func (s Status) String() string` |
| 指针接收器 | 需要修改状态、大型结构 | `func (c *Container) Register(...)` |
| 一致性 | 同一类型的所有方法使用相同接收器 | 不要混用 |

### 6.4 函数式选项模式

```go
// 定义选项类型
type Option func(*options)

type options struct {
    timeout time.Duration
    retries int
    logger  log.Logger
}

// 默认值
func defaultOptions() *options {
    return &options{
        timeout: 30 * time.Second,
        retries: 3,
        logger:  log.Nop(),
    }
}

// 选项函数
func WithTimeout(d time.Duration) Option {
    return func(o *options) {
        o.timeout = d
    }
}

func WithRetries(n int) Option {
    return func(o *options) {
        o.retries = n
    }
}

// 使用
func NewClient(opts ...Option) *Client {
    o := defaultOptions()
    for _, opt := range opts {
        opt(o)
    }
    // ...
}

client := NewClient(
    WithTimeout(60 * time.Second),
    WithRetries(5),
)
```

---

## 7. 数据结构

### 7.1 结构体字段顺序

```go
type Container struct {
    // 1. 互斥体（放最前面）
    mu sync.RWMutex

    // 2. 核心状态
    beans     map[string]*BeanDefinition
    singletons map[string]any

    // 3. 配置
    enableFieldTag bool
    logger         log.Logger

    // 4. 依赖
    parent Container
}
```

### 7.2 切片和映射

```go
// ✅ 正确：预分配容量（已知大小时）
users := make([]User, 0, len(ids))
cache := make(map[string]int, 100)

// ✅ 正确：初始化为空（非 nil）
items := []string{}
counts := map[string]int{}

// ❌ 错误：nil 切片/映射（可能导致 panic）
var items []string      // nil，append 可用但 len 为 0
var counts map[string]int // nil，直接赋值会 panic
```

### 7.3 字符串处理

| 场景 | 推荐方法 | 说明 |
|------|---------|------|
| 简单转换 | `strconv.Itoa()`, `strconv.ParseInt()` | 性能更好 |
| 复杂格式化 | `fmt.Sprintf()` | 可读性更好 |
| 循环拼接 | `strings.Builder` | 避免内存分配 |
| 错误消息 | 使用 `%q` 显示字符串边界 | `fmt.Sprintf("invalid name: %q", name)` |

```go
// ✅ 正确：strings.Builder
var buf strings.Builder
for _, s := range parts {
    buf.WriteString(s)
    buf.WriteString(",")
}
result := buf.String()

// ❌ 错误：循环中 + 拼接
result := ""
for _, s := range parts {
    result += s + ","
}
```

---

## 8. 错误处理

### 8.1 错误处理原则

| 原则 | 说明 | 示例 |
|------|------|------|
| 不忽略错误 | 所有错误必须处理 | `if err != nil { ... }` |
| 包装错误 | 使用 `%w` 保留错误链 | `fmt.Errorf("context: %w", err)` |
| 哨兵错误 | 框架级错误用 `errors.New()` | `var ErrNotFound = errors.New("not found")` |
| 自定义错误 | 复杂错误用结构体 | `type ValidationError struct { ... }` |

### 8.2 错误包装

```go
// ✅ 正确：包装错误，添加上下文
if err := db.QueryRow("SELECT ...").Scan(&user); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("user %s not found: %w", id, ErrNotFound)
    }
    return nil, fmt.Errorf("query user %s failed: %w", id, err)
}

// ❌ 错误：直接返回原始错误
if err := db.QueryRow("SELECT ...").Scan(&user); err != nil {
    return nil, err
}
```

### 8.3 错误判断

```go
// ✅ 正确：使用 errors.Is/As
if errors.Is(err, ErrNotFound) {
    // 处理特定错误
}

var validationErr *ValidationError
if errors.As(err, &validationErr) {
    // 处理自定义错误类型
}

// ❌ 错误：直接比较
if err == ErrNotFound { // 包装后无法匹配
    // ...
}
```

### 8.4 自定义错误类型

```go
// ValidationError 验证错误
type ValidationError struct {
    Field   string
    Message string
    Code    string
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}

// Unwrap 支持 errors.Is/As
func (e *ValidationError) Unwrap() error {
    return nil
}
```

---

## 9. 并发编程

### 9.1 Goroutine 管理

```go
// ✅ 正确：使用 errgroup 管理 goroutine
var g errgroup.Group
for _, task := range tasks {
    t := task // 捕获循环变量
    g.Go(func() error {
        return process(t)
    })
}
if err := g.Wait(); err != nil {
    return fmt.Errorf("task execution failed: %w", err)
}

// ❌ 错误：裸 goroutine，无法等待和捕获错误
for _, task := range tasks {
    go process(task)
}
```

### 9.2 通道使用

```go
// ✅ 正确：明确通道方向和缓冲
func producer(ctx context.Context, out chan<- Item) {
    defer close(out)
    for _, item := range items {
        select {
        case out <- item:
        case <-ctx.Done():
            return
        }
    }
}

func consumer(ctx context.Context, in <-chan Item) error {
    for item := range in {
        if err := process(item); err != nil {
            return err
        }
    }
    return nil
}
```

### 9.3 同步原语选择

| 场景 | 推荐 | 说明 |
|------|------|------|
| 读写分离 | `sync.RWMutex` | 读多写少场景 |
| 互斥访问 | `sync.Mutex` | 简单互斥 |
| 一次性初始化 | `sync.Once` | 单例模式 |
| 等待完成 | `sync.WaitGroup` | 等待多个 goroutine |
| 条件等待 | `sync.Cond` | 复杂条件同步 |
| 值存储 | `sync.Map` | 并发安全的 map |

### 9.4 Context 使用

```go
// ✅ 正确：context 作为第一个参数
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    // ...
}

// 超时控制
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := service.DoSomething(ctx)
```

---

## 10. 接口设计

### 10.1 接口定义原则

| 原则 | 说明 | 示例 |
|------|------|------|
| 接口隔离 | 小接口优于大接口 | `io.Reader`, `io.Writer` |
| 通过组合形成大接口 | 小接口组合形成大接口 | `Container` 嵌入 `BeanGet` + `BeanRegister` |
| 避免预先定义 | 需要时再定义接口 | 不要一开始就定义大接口 |
| 构造函数返回接口 | 隐藏实现细节 | `func NewContainer() Container` |

### 10.2 实现隐藏规范

```go
// ✅ 正确：构造函数返回接口类型
func NewContainer() Container {
    return &defaultContainer{...}
}

// ❌ 错误：构造函数返回具体结构体指针
func NewContainer() *defaultContainer {
    return &defaultContainer{...}
}
```

### 10.3 接口示例

```go
// ✅ 正确：小接口
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// ✅ 正确：组合接口
type ReadWriter interface {
    Reader
    Writer
}

// ❌ 错误：大接口（上帝接口）
type DataProcessor interface {
    Read() ([]byte, error)
    Write([]byte) error
    Parse() (*Result, error)
    Validate() error
    Transform(*Result) (*Result, error)
    Save(*Result) error
    // ... 太多方法
}
```

---

## 11. 泛型使用

### 11.1 泛型最佳实践

| 场景 | 推荐 | 说明 |
|------|------|------|
| 类型安全容器 | `Repository[T any]` | 避免类型断言 |
| 工具函数 | `ZeroOf[T any]() T` | 替代反射 |
| 避免过度泛型 | 不要 `Container[T, U, V]` | 清晰优先于抽象 |

### 11.2 泛型示例

```go
// ✅ 正确：泛型仓储
type Repository[T any] interface {
    Create(entity *T) error
    FindByID(id any) (*T, error)
    FindAll() ([]T, error)
    Delete(id any) error
}

// ✅ 正确：泛型工具函数
func ZeroOf[T any]() T {
    var zero T
    return zero
}

func Clone[T any](v T) T {
    return v
}

// ❌ 错误：过度泛型
type Container[T any, U comparable, V ~int] struct {
    // 过于复杂，难以使用
}
```

---

## 12. 测试风格

### 12.1 表驱动测试

```go
func TestContainer_Get_NotFound(t *testing.T) {
    t.Parallel()
    
    container := core.NewContainer()
    
    _, err := core.GetByName[*UserService](container, "nonexistent")
    
    if err == nil {
        t.Error("Expected error for nonexistent bean")
    }
    if err != core.ErrBeanNotFound {
        t.Errorf("Expected ErrBeanNotFound, got %v", err)
    }
}

func TestContainer_Register_WithFactory(t *testing.T) {
    t.Parallel()
    
    container := core.NewContainer()
    
    err := core.Register[*UserService](container,
        core.WithName[*UserService]("testService"),
        core.WithFactory[*UserService](func(c ...any) (any, error) {
            return &UserService{Name: "test"}, nil
        }),
    )
    if err != nil {
        t.Fatalf("Register failed: %v", err)
    }
    
    svc, err := core.GetByName[*UserService](container, "testService")
    if err != nil {
        t.Fatalf("GetByName failed: %v", err)
    }
    if svc.Name != "test" {
        t.Errorf("Expected name 'test', got %q", svc.Name)
    }
}
```

### 12.2 测试命名

| 类型 | 格式 | 示例 |
|------|------|------|
| 单元测试 | `Test功能_条件_期望` | `TestContainer_Get_NotFound` |
| 边界测试 | `Test功能_Boundary_期望` | `TestParse_BoundaryZero_Success` |
| 集成测试 | `TestIntegration_场景_期望` | `TestIntegration_FullLifecycle_Success` |
| 基准测试 | `Benchmark功能` | `BenchmarkContainer_Register` |

### 12.3 覆盖率目标

| 模块 | 目标覆盖率 | 说明 |
|------|-----------|------|
| 核心框架 | 90%+ | IoC, AOP, Config 等 |
| 工具函数 | 95%+ | 字符串、类型转换等 |
| 集成测试 | 覆盖主要流程 | 端到端场景 |

---

## 13. 性能考量

### 13.1 内存优化

| 技巧 | 说明 | 示例 |
|------|------|------|
| 预分配 | 已知容量时预分配 | `make([]T, 0, cap)` |
| 对象池 | 频繁分配的对象 | `sync.Pool` |
| 避免反射 | 反射性能开销大 | 使用泛型替代 |
| 字符串拼接 | 循环中使用 Builder | `strings.Builder` |

```go
// ✅ 正确：预分配切片
func processItems(items []Item) []Result {
    results := make([]Result, 0, len(items)) // 预分配容量
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// ✅ 正确：使用对象池
var bufferPool = sync.Pool{
    New: func() any {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

func getBuffer() *bytes.Buffer {
    return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer) {
    buf.Reset()
    bufferPool.Put(buf)
}
```

### 13.2 反射优化

反射是性能瓶颈，应尽量减少使用：

```go
// ❌ 错误：频繁使用反射
func getValue(obj any, fieldName string) any {
    v := reflect.ValueOf(obj)
    f := v.FieldByName(fieldName)
    return f.Interface()
}

// ✅ 正确：使用泛型替代
func GetValue[T any](obj T) T {
    return obj
}

// ✅ 正确：缓存反射结果
var typeCache = sync.Map{}

func getCachedType(t reflect.Type) reflect.Type {
    if cached, ok := typeCache.Load(t); ok {
        return cached.(reflect.Type)
    }
    typeCache.Store(t, t)
    return t
}
```

### 13.3 基准测试

```go
func BenchmarkContainer_Register(b *testing.B) {
    c := core.NewContainer()
    b.ResetTimer() // 重置计时器，排除初始化时间
    for i := 0; i < b.N; i++ {
        name := fmt.Sprintf("bean-%d", i)
        _ = core.Register[*MockService](c,
            core.WithName[*MockService](name),
            core.WithFactory[*MockService](func(c ...any) (any, error) {
                return &MockService{}, nil
            }),
        )
    }
}

func BenchmarkContainer_Get(b *testing.B) {
    c := core.NewContainer()
    _ = core.Register[*MockService](c,
        core.WithName[*MockService]("test"),
        core.WithFactory[*MockService](func(c ...any) (any, error) {
            return &MockService{}, nil
        }),
    )
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = core.GetByName[*MockService](c, "test")
    }
}

// 运行基准测试
// go test -bench=. -benchmem ./core
```

**基准测试输出解读**：
- `ns/op`: 每次操作的纳秒数（越小越好）
- `B/op`: 每次操作的字节数（越小越好）
- `allocs/op`: 每次操作的分配次数（越小越好）

### 13.4 常见性能陷阱

| 陷阱 | 问题 | 解决方案 |
|------|------|----------|
| 循环中 `+` 拼接字符串 | 每次分配新内存 | 使用 `strings.Builder` |
| 未预分配切片 | 多次扩容复制 | `make([]T, 0, cap)` |
| 闭包捕获循环变量 | 数据竞争 | 在循环内创建副本 |
| 过度使用反射 | 性能开销大 | 使用泛型或代码生成 |
| 频繁 `interface{}` 转换 | 类型断言开销 | 使用泛型 |
| 大对象值传递 | 内存复制 | 使用指针传递 |

```go
// ❌ 错误：循环中 + 拼接字符串
func joinStrings(items []string) string {
    result := ""
    for _, s := range items {
        result += s + ","
    }
    return result
}

// ✅ 正确：使用 strings.Builder
func joinStrings(items []string) string {
    var buf strings.Builder
    for _, s := range items {
        buf.WriteString(s)
        buf.WriteString(",")
    }
    return buf.String()
}

// ❌ 错误：大对象值传递
type LargeStruct struct {
    Data [1024]byte
}

func process(s LargeStruct) { // 值传递，复制 1024 字节
    // ...
}

// ✅ 正确：使用指针传递
func process(s *LargeStruct) { // 指针传递，仅复制 8 字节
    // ...
}
```

### 13.5 性能优化优先级

1. **算法优化**：选择更优的算法（O(n²) → O(n log n)）
2. **减少分配**：减少内存分配次数
3. **缓存优化**：缓存计算结果
4. **并发优化**：利用多核并行处理
5. **底层优化**：使用更底层的数据结构

---

## 14. 代码审查清单

### 14.1 功能正确性

- [ ] 逻辑正确，边界条件处理得当
- [ ] 错误处理完善，没有忽略错误
- [ ] 并发安全，正确使用同步原语
- [ ] 资源正确释放（defer close）

### 14.2 代码质量

- [ ] 代码清晰易懂，符合命名规范
- [ ] 无冗余代码，遵循 DRY 原则
- [ ] 注释恰当，解释复杂逻辑
- [ ] 文件不超过 500 行，函数不超过 50 行

### 14.3 性能考虑

- [ ] 无明显性能瓶颈
- [ ] 内存分配合理（预分配、对象池）
- [ ] 循环和递归使用得当
- [ ] 避免不必要的反射

### 14.4 安全性

- [ ] 输入验证充分
- [ ] 无 SQL 注入风险
- [ ] 敏感信息妥善处理
- [ ] 无 goroutine 泄漏

### 14.5 测试覆盖

- [ ] 使用表驱动测试
- [ ] 覆盖正常路径和错误路径
- [ ] 测试覆盖率 >= 80%
- [ ] 基准测试覆盖性能敏感代码
- [ ] 测试函数使用 `t.Parallel()` 支持并发

---

## 参考资源

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Blog - Go 代码审查意见](https://go.dev/blog/comments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [enhance 开发规范（AI 必读）](AGENTS.md)
- [enhance 架构设计](ARCHITECTURE.md)