# Core - 类型安全的 IoC 容器

> 参考 Spring IoC 设计思想，遵循 Go 语言哲学的依赖注入容器

## 设计哲学

| 原则 | 说明 |
|------|------|
| **编译期类型安全** | 通过泛型在编译期检查 Bean 类型，避免运行时错误 |
| **构造器注入优先** | 通过工厂函数显式声明依赖，依赖关系清晰 |
| **减少反射** | 用户 API 零反射，反射仅用于字段注入场景 |
| **显式优于隐式** | 依赖关系清晰可见，避免魔法行为 |
| **小接口** | 接口隔离，每个接口职责单一 |
| **零外部依赖** | 仅使用 Go 标准库 |

## 架构总览

```
core/
├── doc.go                    # 核心接口定义（Container, BeanDef, BeanOption）
├── container.go              # Scope 类型定义
├── container_impl.go         # 默认容器实现
├── generic_api.go            # 泛型 API（Register[T], Get[T], MustGet[T], Has[T]）
├── errors.go                 # 错误定义
├── container_test.go         # 容器测试（17 个用例）
│
├── scope/                    # 作用域管理
│   ├── doc.go                # Scope, ScopeRegistry 接口
│   └── scope_impl.go         # Singleton/Prototype 实现（sync.Map）
│
├── lifecycle/                # 生命周期管理
│   ├── doc.go                # LifecycleManager 接口
│   └── lifecycle_impl.go     # 生命周期管理器实现
│
├── binding/                  # 数据绑定
│   ├── doc.go                # Binder, ValueResolver, TypeConverter + Inject[T]
│   ├── binding_impl.go       # 字段注入和配置绑定实现
│   ├── inject_impl.go        # 泛型注入实现
│   └── binding_test.go       # 绑定测试（8 个用例）
│
└── registry/                 # Bean 注册表（内部）
    ├── doc.go                # BeanRegistry, BeanIDGenerator 接口
    └── registry_impl.go      # 注册表实现（sync.Map）
```

## 核心接口

### Container 接口体系

```
┌─────────────────────────────────────────────────┐
│              Container 接口                      │
│  ┌───────────────────────────────────────────┐  │
│  │  BeanGet (获取接口)                        │  │
│  │    - Get(typ) ([]any, error)              │  │
│  │    - GetByTypeAndName(name, typ)          │  │
│  │    - Has(name, typ) bool                  │  │
│  │    - HasType(typ) bool                    │  │
│  ├───────────────────────────────────────────┤  │
│  │  BeanRegister (注册接口)                   │  │
│  │    - RegisterBean(def) error              │  │
│  │    - RegisterWithType(typ) error          │  │
│  │    - RegisterBeanNotExistType(name, typ)  │  │
│  ├───────────────────────────────────────────┤  │
│  │  生命周期管理                               │  │
│  │    - Initialize() error                   │  │
│  │    - Destroy() error                      │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
         ▲
         │ 扩展
┌─────────────────────────────────────────────────┐
│              ContainerExt 接口                   │
│  ┌───────────────────────────────────────────┐  │
│  │    - SetParent(parent)                    │  │
│  │    - GetParent() Container                │  │
│  │    - Types() []reflect.Type               │  │
│  │    - BeanCount() int                      │  │
│  │    - BeanCountType(typ) int               │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

### Bean ID 格式

```
包路径.类型名#自定义名称
```

示例：
- `github.com/example/app/services.UserService` (无自定义名称)
- `github.com/example/app/services.UserService#primary` (有自定义名称)

**注意**：Bean ID 由容器内部自动生成，用户 API 不直接暴露通过 beanId 注册或获取 Bean 的操作。

## 子包说明

### 1. scope - 作用域管理

**核心接口**：

| 接口 | 说明 |
|------|------|
| `Scope` | 作用域策略接口，定义 Bean 实例的获取和存储策略 |
| `ScopeRegistry` | 作用域注册表，管理和扩展作用域 |

**内置作用域**：

| 作用域 | 说明 | 生命周期 |
|--------|------|---------|
| Singleton | 单例，全局唯一（默认） | 应用生命周期 |
| Prototype | 原型，每次获取创建新实例 | 获取时创建 |

**并发优化**：使用 `sync.Map` 替代 `sync.RWMutex`，优化读多写少的注册场景。

**扩展点**：实现 `Scope` 接口可自定义作用域（如 Request、Session）。

### 2. lifecycle - 生命周期管理

**核心接口**：

| 接口/类型 | 说明 |
|-----------|------|
| `LifecycleBean` | Bean 生命周期接口（Init/Destroy） |
| `LifecycleListener` | 生命周期监听器接口 |
| `LifecycleManager` | 生命周期管理器接口 |
| `Phase` | 生命周期阶段枚举 |

**生命周期阶段**：

```
注册 → 实例化 → 依赖注入 → 初始化 → 使用中 → 销毁
```

**使用方式**：

方式一：实现 `LifecycleBean` 接口（推荐，零反射）

```go
type MyService struct{}

func (s *MyService) Init() error {
    // 初始化逻辑
    return nil
}

func (s *MyService) Destroy() error {
    // 清理逻辑
    return nil
}
```

方式二：使用函数式选项

```go
core.Register(container, "myService",
    func(c core.Container) *MyService { return &MyService{} },
    core.WithInit(func(s *MyService) error { ... }),
    core.WithDestroy(func(s *MyService) error { ... }),
)
```

### 3. binding - 数据绑定

**核心接口**：

| 接口 | 说明 |
|------|------|
| `Binder` | 数据绑定器接口 |
| `ValueResolver` | 配置值解析器接口 |
| `TypeConverter` | 类型转换器接口 |

**注入方式**：

| 方式 | 说明 | 推荐度 | 反射使用 |
|------|------|--------|---------|
| 构造器注入 | 通过工厂函数参数注入 | ★★★★★ | 无 |
| 泛型注入 | 通过 `Inject[T]` / `MustInject[T]` 函数 | ★★★★ | 无 |
| 字段注入 | 通过 `inject` 标签注入 | ★★★ | 是 |
| 配置绑定 | 通过 `value` 标签绑定配置值 | ★★★★ | 是 |

**泛型注入函数**：

```go
// 安全获取（返回错误）
func Inject[T any](container core.BeanGet, beanName string) (T, error)

// 必须获取（失败 panic）
func MustInject[T any](container core.BeanGet, beanName string) T
```

### 4. registry - Bean 注册表（内部）

**核心接口**：

| 接口 | 说明 |
|------|------|
| `BeanRegistry` | Bean 注册表接口（内部使用） |
| `BeanIDGenerator` | Bean ID 生成器 |

**注册表结构**：
- `beanID → BeanDef`：Bean 定义映射
- `beanID → any`：Singleton Bean 实例缓存
- `reflect.Type → []beanID`：类型到 Bean ID 列表的映射

**注意**：此包仅供容器内部使用，不对外暴露。

## 快速开始

### 1. 创建容器

```go
container := core.NewContainer()
```

### 2. 注册 Bean（编译期类型检查）

```go
// 工厂注册（推荐）
core.Register(container, "db", func(c core.Container) *Database {
    return &Database{DSN: "localhost:3306"}
})

// 带依赖的注册
core.Register(container, "userService", func(c core.Container) *UserService {
    db := core.MustGet[*Database](c, "db") // 编译期检查类型
    return &UserService{DB: db}
})
```

### 3. 获取 Bean（编译期类型检查）

```go
// 安全获取（返回错误）
userService, err := core.Get[*UserService](container, "userService")
if err != nil {
    // 处理错误
}

// 必须获取（失败 panic）
userService := core.MustGet[*UserService](container, "userService")
```

### 4. 初始化容器

```go
// 创建所有非延迟初始化的 Singleton Bean
container.Initialize()
```

### 5. 销毁容器

```go
// 调用所有 Singleton Bean 的 Destroy 回调
container.Destroy()
```

## 高级特性

### Bean 生命周期回调

```go
// 方式一：使用选项函数
core.Register(container, "myService", 
    func(c core.Container) *MyService {
        return &MyService{}
    },
    core.WithInit(func(s *MyService) error {
        // 初始化逻辑
        return nil
    }),
    core.WithDestroy(func(s *MyService) error {
        // 清理逻辑
        return nil
    }),
)

// 方式二：实现 LifecycleBean 接口
type MyService struct{}

func (s *MyService) Init() error {
    // 初始化逻辑
    return nil
}

func (s *MyService) Destroy() error {
    // 清理逻辑
    return nil
}
```

### 延迟初始化

```go
// 延迟初始化，首次获取时创建
core.Register(container, "expensive", func(c core.Container) *Expensive {
    return &Expensive{
        // 耗时初始化
    }
}, core.WithLazy[*Expensive](true))
```

### 原型作用域

```go
// 每次获取都创建新实例
core.Register(container, "requestHandler", func(c core.Container) *RequestHandler {
    return &RequestHandler{}
}, core.WithScope[*RequestHandler](core.Prototype))
```

### 子容器

```go
// 创建父容器
parent := core.NewContainer()
core.Register(parent, "config", func(c core.Container) *Config {
    return &Config{Timeout: 30}
})

// 创建子容器
child := core.NewContainer()
child.(core.ContainerExt).SetParent(parent)

// 子容器可以获取父容器的 Bean
config := core.MustGet[*Config](child, "config") // 从 parent 获取
```

### 条件注册

```go
// 仅当 Bean 不存在时注册
ext := container.(core.ContainerExt)
ext.RegisterBeanNotExistType("defaultCache", reflect.TypeOf(&Cache{}))
```

## 编译期类型安全

### 传统方式（运行时错误）

```go
// ❌ 错误：运行时才能发现类型不匹配
container.Register("db", &Database{})
db := container.Get("db").(*WrongType) // 运行时 panic
```

### 类型安全方式（编译期错误）

```go
// ✅ 正确：编译期就能发现类型不匹配
core.Register(container, "db", func(c core.Container) *Database {
    return &Database{}
})
db := core.MustGet[*Database](container, "db") // 编译期检查类型
```

### 编译期检查的优势

| 场景 | 传统方式 | 类型安全方式 |
|------|---------|-------------|
| 类型不匹配 | 运行时 panic | 编译期错误 |
| 依赖缺失 | 运行时 nil | 运行时错误（可处理） |
| 工厂返回错误类型 | 运行时 panic | 编译期错误 |
| 重构后类型变更 | 运行时 panic | 编译期错误 |

## 性能考量

| 操作 | 复杂度 | 说明 |
|------|--------|------|
| Register | O(1) | 直接存储 |
| Get(Singleton) | O(1) | 首次创建后缓存 |
| Get(Prototype) | O(1) | 每次调用工厂 |
| Inject | O(1) | 同 Get |

## 与 Spring IoC 对比

| 特性 | Spring | enhance/core |
|------|--------|--------------|
| 配置方式 | XML/注解/Java Config | 泛型函数式 API |
| 类型安全 | 运行时检查 | 编译期泛型检查 |
| 反射使用 | 大量反射 | 最小化反射 |
| Bean 定义 | BeanDefinition | BeanDef[T] |
| 依赖注入 | @Autowired | 工厂函数 + Inject[T] |
| 作用域 | 5 种作用域 | Singleton/Prototype（可扩展） |
| 生命周期 | @PostConstruct/@PreDestroy | 接口 + 函数式选项 |
| 自动扫描 | Classpath 扫描 | 不适用（Go 无运行时类路径扫描） |

## 最佳实践

### 1. 优先使用构造器注入

```go
// ✅ 推荐：依赖关系明确
core.Register(container, "userService", func(c core.Container) *UserService {
    db := core.MustGet[*Database](c, "db")
    return &UserService{DB: db}
})
```

### 2. 使用 MustGet 确定存在的 Bean

```go
// ✅ 确定 Bean 一定存在
db := core.MustGet[*Database](container, "db")

// ⚠️ 可能不存在时使用 Get
cache, err := core.Get[*Cache](container, "cache")
if err != nil {
    // 处理缺失情况
}
```

### 3. 避免循环依赖

```go
// ❌ 错误：循环依赖
core.Register(container, "a", func(c core.Container) *A {
    b := core.MustGet[*B](c, "b")
    return &A{B: b}
})

core.Register(container, "b", func(c core.Container) *B {
    a := core.MustGet[*A](c, "a") // 循环依赖
    return &B{A: a}
})
```

### 4. 使用延迟初始化优化启动时间

```go
// ✅ 延迟初始化不常用的 Bean
core.Register(container, "reportService", func(c core.Container) *ReportService {
    return &ReportService{
        // 耗时初始化
    }
}, core.WithLazy[*ReportService](true))
```

### 5. 实现 LifecycleBean 接口

```go
// ✅ 推荐：接口方式，零反射
type MyService struct{}

func (s *MyService) Init() error { ... }
func (s *MyService) Destroy() error { ... }
```

## 测试覆盖

| 包 | 测试数 | 状态 |
|----|--------|------|
| core | 17 | ✅ 全部通过 |
| binding | 8 | ✅ 全部通过 |
| **总计** | **25** | ✅ **全部通过** |

运行测试：

```bash
go test ./core/... -count=1
```