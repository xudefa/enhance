# boot 包 — 应用启动器

> **所属层级**: Boot Layer  
> **设计理念**: 自动配置，快速启动  
> **设计灵感**: Spring Boot SpringApplication

## 概述

`boot` 包是 enhance 框架的应用启动核心，提供应用启动的全生命周期管理。支持自动配置、启动器管理、横幅打印、失败分析等特性。

### 核心功能

| 功能 | 说明 |
|------|------|
| **自动配置** | 各模块通过 `RegisterAutoConfig` 注册自动配置 |
| **启动器管理** | 管理组件的启动和停止生命周期 |
| **横幅打印** | 应用启动时显示 ASCII 艺术横幅 |
| **失败分析** | 启动失败时提供友好的错误提示 |
| **配置加载** | 灵活的配置加载机制，支持多环境配置 |
| **覆盖机制** | 用户自定义配置覆盖默认自动配置 |

---

## 核心接口

### Boot 结构体

```go
type Boot struct {
    ctx           *contextpkg.DefaultApplicationContext
    config        *BootConfig
    configLoader  *environment.ConfigLoader
    starters      []Starter
}
```

主要方法：

| 方法 | 说明 |
|------|------|
| `Start()` | 启动应用，执行完整的生命周期 |
| `Stop()` | 停止应用，优雅关闭 |
| `IsRunning()` | 检查应用是否运行中 |
| `WaitForSignal()` | 等待终止信号（SIGINT/SIGTERM），自动优雅关闭 |
| `Context()` | 返回应用上下文 |
| `Container()` | 返回 IoC 容器 |
| `Environment()` | 返回环境配置 |

### Starter 接口

```go
type Starter interface {
    Name() string
    Dependencies() []string
    Configure(ctx ApplicationContext) error
    Start(ctx ApplicationContext) error
    Stop(ctx ApplicationContext) error
    GetCondition() condition.Condition
}
```

### AutoConfiguration 接口

```go
type AutoConfiguration interface {
    Configure(ctx ApplicationContext) error
}
```

### Banner 接口

```go
type Banner interface {
    Print(ctx ApplicationContext)
}
```

### FailureAnalyzer 接口

```go
type FailureAnalyzer interface {
    CanAnalyze(err error) bool
    Analyze(err error) *FailureReport
}
```

---

## 快速开始

### 创建应用

```go
package main

import (
    "log"
    "github.com/xudefa/enhance/boot"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProfiles("dev"),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
    defer app.Stop()

    app.WaitForSignal()
}
```

### 函数式选项

| 选项 | 说明 |
|------|------|
| `WithAppName(name)` | 设置应用名称 |
| `WithVersion(version)` | 设置版本号 |
| `WithProfiles(profiles...)` | 设置激活的 Profile |
| `WithConfigLocation(location)` | 设置配置文件路径 |
| `WithConfigType(configType)` | 设置配置文件类型 (json/yaml) |
| `WithPropertySource(source)` | 添加自定义配置源 |
| `WithoutAutoConfig()` | 禁用自动配置执行 |
| `WithoutStarters()` | 禁用启动器自动管理 |

---

## API 参考

### 配置加载

#### 配置文件命名

- 基础配置：`application.json`
- 环境特定配置：`application-{profile}.json`

#### 配置文件搜索路径（按优先级）

1. `/etc/config`
2. 当前目录
3. `./config`
4. 可执行文件所在目录
5. 可执行文件所在目录的 `./config`

#### 配置加载顺序

1. 加载基础配置
2. 加载环境特定配置
3. 环境特定配置覆盖基础配置

### 自动配置

#### 注册自动配置

```go
func init() {
    boot.RegisterAutoConfig(&GinAutoConfiguration{},
        condition.OnProperty("gin.enabled", "true"),
    )
}
```

#### 带选项的注册

```go
func init() {
    boot.RegisterAutoConfigWith(&GinAutoConfiguration{},
        boot.WithOrder(int(boot.OrderPriorityWebLayer)),
        boot.WithDependsOn("dbConfig", "redisConfig"),
    )
}
```

#### 覆盖机制

```go
func init() {
    // 覆盖默认的 Gin 自动配置
    boot.RegisterAutoConfigWith(&MyGinConfig{},
        boot.WithOverride("*GinAutoConfiguration"),
        // Order 默认设置为 -100，确保在原始配置之前执行
    )
}
```

### 执行顺序优先级

框架定义了 9 个执行顺序层级，值越小优先级越高：

| 优先级常量 | 值 | 适用组件 |
|-----------|-----|---------|
| `OrderPriorityInfrastructure` | -3000 | 日志、配置中心等基础设施 |
| `OrderPriorityDataLayer` | -2000 | 数据库、缓存、对象存储 |
| `OrderPriorityAuthentication` | -1500 | JWT、OAuth2、LDAP 认证 |
| `OrderPriorityAuthorizationGorm` | -1300 | Casbin GORM 适配器等数据库授权适配器 |
| `OrderPriorityAuthorization` | -1200 | Casbin、RBAC、ABAC 授权 |
| `OrderPrioritySecurityCore` | -100 | 安全过滤器链、访问控制 |
| `OrderPriorityWebLayer` | 0 | HTTP 服务器、路由、中间件 |
| `OrderPriorityBusinessLayer` | 1000 | 定时任务、消息队列、事件总线 |
| `OrderPriorityMonitoringLayer` | 2000 | Actuator、健康检查、链路追踪 |

**依赖关系示例**：

```
日志(-3000) → 数据库(-2000) → 认证(-1500) → CasbinGorm(-1300) → Casbin(-1200) → 安全(-100) → Web(0) → 业务(1000) → 监控(2000)
```

**第三方插件集成指南**：
- Redis 缓存: `OrderPriorityDataLayer` (-2000)
- 消息队列: `OrderPriorityBusinessLayer` (1000)
- 对象存储: `OrderPriorityDataLayer` (-2000)
- 搜索引擎: `OrderPriorityDataLayer` (-2000)
- 链路追踪: `OrderPriorityMonitoringLayer` (2000)
- 指标收集: `OrderPriorityMonitoringLayer` (2000)

### 启动器

#### 注册启动器

```go
func init() {
    boot.RegisterStarter(MyStarter{})
}
```

#### 依赖拓扑排序

`GetOrdered()` 使用 **Kahn 算法**对启动器进行拓扑排序：

```go
A → B → C
A → C

排序结果: [A, B, C]
```

逆序停止：停止时按排序结果的逆序执行，保证依赖方先于被依赖方停止。

### 横幅

#### 内置实现

| 类型 | 说明 |
|------|------|
| `LegacyBanner` | 默认 ASCII 艺术横幅 |
| `TextBanner` | 文本横幅，支持属性模板 |
| `ASCIIArtBanner` | ASCII 艺术横幅 |
| `CustomTemplateBanner` | 自定义模板横幅 |

#### 默认横幅效果

```
  ________                  ____             __
 /  _____/  ____    ____   / __ )  ____     / /_
/   \  ___ /  _ \  /  _ \ / __  | / __ \   / __ \
\    \_\  (  <_> )(  <_> ) /_/ / / /_/ /  / /_/ /
 \______  /\____/  \____/_____/  \____/   /___/
        \/
:: my-app :: v1.0.0 :: profiles(dev)
```

### 失败分析

#### FailureReport

```go
type FailureReport struct {
    Headline          string
    Description       string
    Action            string
    Cause             string
    Details           map[string]any
    StackTrace        string
    PossibleSolutions []string
}
```

#### 注册失败分析器

```go
boot.RegisterFailureAnalyzer(MyAnalyzer{})
```

或使用 `SimpleFailureAnalyzer`：

```go
boot.RegisterFailureAnalyzer(
    boot.NewSimpleFailureAnalyzer(func(err error) *boot.FailureReport {
        if strings.Contains(err.Error(), "port") {
            return &boot.FailureReport{
                Description: "端口被占用",
                Action:      "请检查端口是否被其他进程占用",
                Cause:       err.Error(),
            }
        }
        return nil
    }),
)
```

#### 失败报告输出

```
====================
APPLICATION FAILED TO START
====================

描述: 端口被占用

动作: 请检查端口是否被其他进程占用

原因: listen tcp :8080: bind: address already in use

可能的解决方案:
  1. 检查端口占用：lsof -i :8080
  2. 修改配置文件中的端口号
  3. 停止占用端口的进程
```

### 结构化启动错误（BootError）

```go
type BootError struct {
    Phase       string   // 错误发生的阶段
    Original    error    // 原始错误
    Analyzed    string   // FailureAnalyzer 分析结果
    Suggestions []string // 修复建议
}
```

#### 错误阶段

| 阶段 | 说明 |
|------|------|
| `configuring` | 配置加载和自动配置阶段 |
| `context-refreshed` | 上下文刷新阶段 |
| `starting` | 启动器启动阶段 |
| `stopping` | 停止阶段 |

#### 使用示例

```go
app, err := boot.NewApplication()
if err != nil {
    if bootErr, ok := err.(*boot.BootError); ok {
        fmt.Printf("启动失败阶段: %s\n", bootErr.Phase)
        fmt.Printf("原始错误: %v\n", bootErr.Original)
        if bootErr.Analyzed != "" {
            fmt.Printf("分析: %s\n", bootErr.Analyzed)
        }
        for _, s := range bootErr.Suggestions {
            fmt.Printf("建议: %s\n", s)
        }
    }
}
```

---

## 启动流程

### 启动阶段

```
PhaseInitializing
    ↓
PhaseConfiguring
    ├── 执行 AutoConfiguration（按顺序）
    ├── 注册 Starter（拓扑排序）
    ├── 调用 Starter.Configure()
    └── 发布 EventEnvironmentPrepared
    ↓
PhaseContextRefreshed
    ├── 发布 EventContextRefreshed
    ↓
PhaseReady
    ├── 调用 Starter.Start()
    ├── 打印 Banner
    ↓
PhaseRunning
    ├── 发布 EventApplicationStarted
    └── 发布 EventApplicationReady
```

### 停止流程

```
PhaseRunning
    ↓
PhaseStopping
    ├── 逆序停止 Starter（反向依赖顺序）
    ↓
PhaseStopped
    └── 发布 EventApplicationStopped
```

---

## 使用示例

### 自定义启动器

```go
type MyStarter struct{}

func (s *MyStarter) Name() string { return "my-starter" }
func (s *MyStarter) Dependencies() []string { return nil }
func (s *MyStarter) Configure(ctx boot.ApplicationContext) error {
    // 配置阶段逻辑
    return nil
}
func (s *MyStarter) Start(ctx boot.ApplicationContext) error {
    // 启动阶段逻辑
    return nil
}
func (s *MyStarter) Stop(ctx boot.ApplicationContext) error {
    // 停止阶段逻辑
    return nil
}
func (s *MyStarter) GetCondition() condition.Condition {
    return condition.OnProperty("my.enabled", "true")
}

func init() {
    boot.RegisterStarter(&MyStarter{})
}
```

### 自定义横幅

```go
type MyBanner struct{}

func (b *MyBanner) Print(ctx boot.ApplicationContext) {
    fmt.Println("================================")
    fmt.Println("  My Application Starting...")
    fmt.Println("================================")
}

// 使用自定义横幅
app, _ := boot.NewApplication()
app.SetBanner(&MyBanner{})
```

---

## 最佳实践

### 1. 使用自动配置简化集成

```go
// ✅ 推荐：使用自动配置
func init() {
    boot.RegisterAutoConfig(&RedisAutoConfiguration{},
        condition.OnProperty("redis.enabled", "true"),
    )
}

// ⚠️ 不推荐：手动配置所有组件
func main() {
    // 手动创建和配置所有组件
}
```

### 2. 合理使用启动器依赖

```go
// ✅ 推荐：声明依赖关系
func (s *MyStarter) Dependencies() []string {
    return []string{"dbStarter", "redisStarter"}
}

// ⚠️ 不推荐：不声明依赖，可能导致初始化顺序错误
func (s *MyStarter) Dependencies() []string {
    return nil
}
```

### 3. 使用条件控制启动器启用

```go
// ✅ 推荐：使用条件控制
func (s *MyStarter) GetCondition() condition.Condition {
    return condition.All(
        condition.OnProperty("my.enabled", "true"),
        condition.OnClass("github.com/some/lib"),
    )
}
```

### 4. 优雅关闭应用

```go
// ✅ 推荐：使用 WaitForSignal 等待终止信号
func main() {
    app, _ := boot.NewApplication()
    app.Start()
    defer app.Stop()
    
    app.WaitForSignal() // 等待 SIGINT/SIGTERM
}
```

### 5. 自定义失败分析器

```go
// ✅ 推荐：提供友好的错误提示
boot.RegisterFailureAnalyzer(
    boot.NewSimpleFailureAnalyzer(func(err error) *boot.FailureReport {
        if strings.Contains(err.Error(), "connection refused") {
            return &boot.FailureReport{
                Description: "数据库连接失败",
                Action:      "请检查数据库服务是否启动，配置是否正确",
                Cause:       err.Error(),
                PossibleSolutions: []string{
                    "检查数据库服务状态",
                    "验证数据库连接配置",
                    "检查网络连接",
                },
            }
        }
        return nil
    }),
)
```