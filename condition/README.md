# condition 包 — 条件装配

> **所属层级**: Core Layer  
> **设计理念**: 按需加载，条件驱动  
> **设计灵感**: Spring Boot @Conditional

## 概述

`condition` 包提供条件装配能力，允许根据配置、类路径、自定义条件等动态决定是否注册 Bean 或执行自动配置。支持条件组合（AND/OR/NOT）、自定义条件、条件验证等特性。

### 核心功能

| 功能 | 说明 |
|------|------|
| **内置条件** | OnProperty、OnClass、OnBean、OnMissingBean、OnProfile 等 |
| **条件组合** | All（AND）、Any（OR）、Not（NOT）逻辑组合 |
| **自定义条件** | 实现 Condition 接口定义业务条件 |
| **表达式条件** | SpEL 风格表达式，支持占位符和运算 |
| **资源条件** | 检查类路径、文件、环境变量等资源存在性 |

---

## 核心接口

### Condition 接口

```go
type Condition interface {
    Matches(ctx ConditionContext) bool  // 判断条件是否匹配
    String() string                     // 条件描述（用于日志和调试）
}
```

### ConditionContext 接口

```go
type ConditionContext interface {
    Environment() interface{ GetProperty(key string) (any, bool) }
    Container() interface{ Has(id string) bool }
    ClassLoader() interface{ HasClass(name string) bool }
    GetBean(beanID string) (any, bool)
    HasProperty(key string) bool
    GetProperty(key string) (any, bool)
}
```

---

## 快速开始

### 在自动配置中使用条件

```go
func init() {
    boot.RegisterAutoConfig(
        &GinAutoConfiguration{},
        condition.OnProperty("gin.enabled", "true"),     // 属性条件
        condition.OnClass("github.com/gin-gonic/gin"),   // 类路径条件
    )
}
```

### 在启动器中使用条件

```go
type MyStarter struct{}

func (s *MyStarter) Name() string { return "my-starter" }
func (s *MyStarter) Dependencies() []string { return nil }
func (s *MyStarter) Configure(ctx boot.ApplicationContext) error { return nil }
func (s *MyStarter) Start(ctx boot.ApplicationContext) error { return nil }
func (s *MyStarter) Stop(ctx boot.ApplicationContext) error { return nil }

func (s *MyStarter) GetCondition() condition.Condition {
    return condition.All(
        condition.OnProperty("my.enabled", "true"),
        condition.OnClass("github.com/some/lib"),
    )
}

func init() {
    boot.RegisterStarter(&MyStarter{})
}
```

---

## API 参考

### 内置条件

| 条件 | 说明 | 示例 |
|------|------|------|
| `OnProperty(key, value?)` | 属性存在且匹配值 | `OnProperty("gin.enabled", "true")` |
| `OnMissingProperty(key)` | 属性不存在 | `OnMissingProperty("gin.enabled")` |
| `OnBean(beanID)` | Bean 存在 | `OnBean("redisClient")` |
| `OnMissingBean(beanID)` | Bean 不存在 | `OnMissingBean("customCache")` |
| `OnProfile(profile)` | Profile 匹配 | `OnProfile("dev")` 或 `OnProfile("!prod")` |
| `OnClass(className)` | 类路径存在 | `OnClass("github.com/gin-gonic/gin")` |
| `OnMissingClass(className)` | 类路径不存在 | `OnMissingClass("github.com/xx/xx")` |
| `OnPropertyPrefix(prefix)` | 配置前缀存在 | `OnPropertyPrefix("server")` |

### OnProperty 匹配逻辑

1. 仅传 key 时：存在且为有效字符串（非空）则匹配
2. 传 key+value 时：存在且值等于预期值则匹配

### OnProfile 匹配逻辑

1. 支持否定前缀 `!dev`（非 dev 环境时匹配）
2. 委托给 `Environment.AcceptsProfile()` 方法

### 组合条件

| 函数 | 逻辑 | 说明 |
|------|------|------|
| `All(conditions...)` | AND | 所有子条件都匹配时通过（短路） |
| `Any(conditions...)` | OR | 任一子条件匹配时通过（短路） |
| `Not(condition)` | NOT | 对子条件结果取反 |

```go
// AND 组合
cond := condition.All(
    condition.OnProperty("db.enabled", "true"),
    condition.OnBean("dataSource"),
    condition.OnProfile("!test"),
)

// OR 组合
cond := condition.Any(
    condition.OnProfile("dev"),
    condition.OnProfile("test"),
)

// NOT 组合
cond := condition.Not(condition.OnProfile("prod"))

// 嵌套组合
cond := condition.All(
    condition.OnProperty("app.enabled", "true"),
    condition.Any(
        condition.OnProfile("dev"),
        condition.OnProfile("staging"),
    ),
    condition.Not(
        condition.OnMissingBean("customHandler"),
    ),
)
```

### 自定义条件

```go
// 方式一：内联函数
cond := condition.Custom("hasCustomDB", func(ctx condition.ConditionContext) bool {
    val, ok := ctx.GetProperty("db.type")
    return ok && val == "postgres"
})

// 方式二：实现 Condition 接口
type MyCondition struct{}

func (m *MyCondition) Matches(ctx condition.ConditionContext) bool {
    bean, ok := ctx.GetBean("userService")
    if !ok {
        return false
    }
    return bean != nil
}

func (m *MyCondition) String() string {
    return "MyCondition"
}

// 使用
condition.All(
    condition.OnProperty("my.feature.enabled", "true"),
    &MyCondition{},
)
```

### 扩展条件

| 条件 | 说明 | 示例 |
|------|------|------|
| `OnExpression(expr)` | SpEL 表达式 | `OnExpression("${server.port} > 8080")` |
| `OnResourceExists(path)` | 资源存在 | `OnResourceExists("classpath:config.yml")` |
| `OnEnvVarExists(name)` | 环境变量存在 | `OnEnvVarExists("DATABASE_URL")` |

#### OnExpression 表达式

```go
// 简单比较
condition.OnExpression("${server.port} > 8080")

// 逻辑运算
condition.OnExpression("${app.env} == 'prod' && ${debug.enabled} == false")

// 三元运算
condition.OnExpression("${feature.flag} ? true : false")
```

#### OnResourceExists 资源路径

```go
// 类路径资源
condition.OnResourceExists("classpath:config.yml")

// 文件系统资源
condition.OnResourceExists("file:/etc/app/config.json")

// 默认类路径
condition.OnResourceExists("config.yml")
```

---

## 使用示例

### 条件组合示例

```go
// 生产环境且数据库配置完整时才注册
container.Register(
    reflect.TypeOf(&ProductionDB{}),
    core.Bean(&ProductionDB{}),
    core.Condition(condition.All(
        condition.OnProfile("prod"),
        condition.OnPropertyExists("db.host"),
        condition.OnPropertyExists("db.port"),
        condition.OnPropertyExists("db.name"),
    )),
)

// 开发环境或测试环境注册 Mock 服务
container.Register(
    reflect.TypeOf(&MockEmailService{}),
    core.Bean(&MockEmailService{}),
    core.Condition(condition.Any(
        condition.OnProfile("dev"),
        condition.OnProfile("test"),
    )),
)
```

### 自定义条件示例

```go
// 仅当应用版本 >= 2.0 时生效
var versionCondition = condition.Custom("version>=2.0", func(ctx condition.ConditionContext) bool {
    val, ok := ctx.GetProperty("app.version")
    if !ok {
        return false
    }
    return val == "2.0" || val == "2.1" || val == "3.0"
})

// 使用自定义条件
boot.RegisterAutoConfig(
    &NewFeatureAutoConfiguration{},
    versionCondition,
)
```

### 条件调试输出

```go
fmt.Println(condition.OnProperty("gin.enabled", "true").String())
// 输出: OnProperty(gin.enabled=true)

fmt.Println(condition.OnMissingBean("redis").String())
// 输出: OnMissingBean(redis)

fmt.Println(condition.All(
    condition.OnProperty("db.enabled", "true"),
    condition.Any(
        condition.OnProfile("dev"),
        condition.OnProfile("staging"),
    ),
).String())
// 输出: All(OnProperty(db.enabled=true), Any(OnProfile(dev), OnProfile(staging)))
```

---

## 与 Spring Boot @Conditional 对照

| Spring Boot | enhance | 说明 |
|-------------|---------|------|
| `@ConditionalOnProperty` | `condition.OnProperty()` | 配置属性条件 |
| `@ConditionalOnMissingProperty` | `condition.OnMissingProperty()` | 配置属性缺失 |
| `@ConditionalOnBean` | `condition.OnBean()` | Bean 存在条件 |
| `@ConditionalOnMissingBean` | `condition.OnMissingBean()` | Bean 缺失条件 |
| `@Profile` | `condition.OnProfile()` | Profile 条件 |
| `@ConditionalOnClass` | `condition.OnClass()` | 类路径条件 |
| `@ConditionalOnMissingClass` | `condition.OnMissingClass()` | 类路径缺失条件 |
| `@ConditionalOnExpression` | `condition.OnExpression()` | 表达式条件 |
| `@Conditional` (多个) | `condition.All()` / `condition.Any()` | 组合条件 |

---

## 最佳实践

### 1. 使用内置条件简化配置

```go
// ✅ 推荐：使用内置条件
container.Register(
    reflect.TypeOf(&MyService{}),
    core.Bean(&MyService{}),
    core.Condition(condition.OnProperty("my.enabled", "true")),
)

// ⚠️ 不推荐：手动检查条件
func init() {
    if os.Getenv("MY_ENABLED") == "true" {
        container.Register(...)
    }
}
```

### 2. 合理使用条件组合

```go
// ✅ 推荐：明确表达条件逻辑
condition.All(
    condition.OnProperty("feature.enabled", "true"),
    condition.OnMissingBean("customImpl"),
)

// ⚠️ 不推荐：嵌套过深
condition.All(
    condition.Any(
        condition.All(...),
        condition.Any(...),
    ),
    condition.None(...),
)
```

### 3. 使用 OnMissingBean 提供默认实现

```go
// ✅ 推荐：提供默认实现，允许用户覆盖
container.Register(
    reflect.TypeOf(&CacheService{}),
    core.Bean(&DefaultCacheService{}),
    core.Condition(condition.OnMissingBean("customCache")),
)
```

### 4. 自定义条件提供清晰的 String() 实现

```go
// ✅ 推荐：便于调试和日志
func (m *MyCondition) String() string {
    return "MyCondition(threshold=100)"
}
```

### 5. 避免过度使用条件装配

```go
// ✅ 推荐：仅在必要时使用条件
// - 可选依赖
// - 环境差异化配置
// - 提供默认实现

// ⚠️ 不推荐：核心业务逻辑使用条件装配
// 这会导致代码难以理解和调试
```