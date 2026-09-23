# enhance AI 快速参考卡

> 一页速查：AI 在 enhance 上执行任务时最常需要的规则、模式、命令。
> 完整规范见 [AGENTS.md](../AGENTS.md)、[CODING_STYLE.md](../CODING_STYLE.md)。

---

## 代码模板

### 新接口定义（doc.go）

```go
// Cache 提供键值缓存能力
type Cache interface {
    Get(ctx context.Context, key string) (any, error)
    Set(ctx context.Context, key string, value any, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

### 构造函数（返回接口）

```go
func NewCache(opts ...CacheOption) Cache {
    c := &cache{maxSize: 256}
    for _, opt := range opts {
        opt(c)
    }
    return c
}
```

### 函数式选项

```go
type CacheOption func(*cache)

func WithMaxSize(size int) CacheOption {
    return func(c *cache) { c.maxSize = size }
}

func WithTTL(ttl time.Duration) CacheOption {
    return func(c *cache) { c.defaultTTL = ttl }
}
```

### 表驱动测试

```go
func TestCache_Get(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        key     string
        want    string
        wantErr bool
    }{
        {"hit", "k1", "v1", false},
        {"miss", "k2", "", true},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got, err := cache.Get(ctx, tt.key)
            if (err != nil) != tt.wantErr {
                t.Errorf("Get(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
            }
        })
    }
}
```

### 错误定义与处理

```go
var ErrNotFound = errors.New("cache: key not found")

func (c *cache) Get(ctx context.Context, key string) (any, error) {
    v, ok := c.data.Load(key)
    if !ok {
        return nil, fmt.Errorf("get key %q: %w", key, ErrNotFound)
    }
    return v, nil
}

if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

### 自动配置注册

```go
func init() {
    boot.RegisterAutoConfig(
        boot.AutoConfigEntry{
            Name:  "enhance.cache",
            Order: boot.OrderPriorityDataLayer,
            Condition: condition.All(
                condition.OnProperty("enhance.cache.enabled", "true"),
                condition.OnMissingBean[cache.Cache](),
            ),
            Configure: configureCache,
        },
    )
}
```

---

## 硬性规则速查

| 规则 | 正确 | 错误 |
|------|------|------|
| 构造函数返回类型 | `func NewX() X` | `func NewX() *x` |
| 接口命名 | `Reader`, `Cache` | `IReader`, `ReaderInterface` |
| 错误比较 | `errors.Is(err, ErrX)` | `err == ErrX` |
| 错误包装 | `fmt.Errorf("...: %w", err)` | `fmt.Errorf("...: %v", err)` |
| goroutine | `errgroup.Go()` / `WaitGroup` | `go func(){}()` |
| 接口位置 | `doc.go` | 实现文件 |
| 第三方依赖 | `starter/<name>/` | 核心包 |
| 测试并发 | `t.Parallel()` | 省略 |
| 配置模式 | `WithXxx()` 函数式选项 | 构造函数参数 |
| 依赖方向 | Boot→Core→Infra | 反向/循环 |
| 函数行数 | ≤80 行（不含注释） | >80 行 |
| 嵌套深度 | ≤17 层 | >17 层 |
| 算法复杂度 | 时间+空间不同时 ≥ O(n²) | 双 O(n²) |

---

## 包依赖速查

```
独立包（零内部依赖）: async, audit, cache, devtools, email, exception,
  i18n, log, lifecycle, metadata, mq, openapi, resilience, retry,
  spel, tenant, validation, webtest

低影响包（少量下游）: event→boot,config,context; retry→event;
  spel→condition; lifecycle→boot,context

高影响包（大量下游）: core, boot, config, condition, log
  → 修改前必须读 AI_CHANGE_IMPACT.md
```

---

## 验证命令速查

```bash
make ai-verify          # 完整验证（编译+测试+格式+静态分析+审计+文档同步+依赖方向）
make ai-quick           # 快速检查（编译+测试+格式）
make quality-score      # 质量评分（≥80）
make ai-check           # AI 可维护性检查
make ai-docs-sync       # 文档同步检查
make ai-deps-check      # 依赖方向检查
make skill-check        # 平台规则一致性检查
go test -race ./...     # 竞争检测
go vet ./...            # 静态分析
```

---

## 文件组织速查

```
<包>/
  doc.go              # 接口门面（只放定义，≤500行）
  types.go            # 类型定义（doc.go 超限时分担）
  <实现>.go           # 实现逻辑
  errors.go           # 错误定义
  builder.go          # 构建器
  autoconfig.go       # 自动配置（如需）
  <实现>_test.go      # 测试（与被测文件一一对应）
  README.md           # 包说明
```

---

## 更多代码模式

### Must 函数（允许 panic）

```go
func MustGet[T any](container Container, name string) T {
    instance, err := GetByName[T](container, name)
    if err != nil {
        panic(err)
    }
    return instance
}
```

> **注意**：仅 `Must*` 函数和 `main()` 中允许 `panic(err)`，其他地方禁止。

### 接口组合（拆分大接口）

```go
type BeanReader interface {
    GetByName(name string) (any, error)
    GetByType(typ reflect.Type) (any, error)
}

type BeanWriter interface {
    Register(name string, instance any) error
}

type Container interface {
    BeanReader
    BeanWriter
    Contains(name string) bool
}
```

### 事件定义与发布

```go
type UserCreatedEvent struct {
    UserID string
    Name   string
}

func (e UserCreatedEvent) Type() string {
    return "user.created"
}

bus.Publish(ctx, UserCreatedEvent{UserID: "123", Name: "Alice"})
```

### 条件化注册

```go
condition.All(
    condition.OnProperty("enhance.cache.enabled", "true"),
    condition.OnMissingBean[cache.Cache](),
)

condition.Any(
    condition.OnBean[redis.Client](),
    condition.OnBean[gorm.DB](),
)
```

### 错误定义完整模式

```go
var (
    ErrNotFound      = errors.New("cache: key not found")
    ErrExpired       = errors.New("cache: key expired")
    ErrMaxSizeExceed = errors.New("cache: max size exceeded")
)
```

---

## 自动配置优先级速查

| 优先级 | 值 | 组件 |
|--------|-----|------|
| Infrastructure | -3000 | log, config, condition |
| DataLayer | -2000 | cache, gorm, redis |
| ServiceDiscovery | -1800 | consul, nacos |
| Authentication | -1500 | jwt, oauth2 |
| Authorization | -1200 | casbin, rbac |
| TaskLayer | -500 | cron, asynq |
| SecurityCore | -100 | security |
| WebLayer | 0 | web |
| BusinessLayer | 1000 | event, schedule |
| Monitoring | 2000 | metrics, actuator |

---

## 常见修改影响面

| 修改的包 | 受影响下游 |
|----------|-----------|
| `core` | actuator, boot, config, context, schedule, security, testing, tracing, web |
| `boot` | actuator, metrics, schedule, security, testing, tracing, web, cmd |
| `config` | actuator, boot, condition, context, schedule, security, testing, tracing |
| `condition` | actuator, boot, metrics, schedule, security, tracing, web, cmd |