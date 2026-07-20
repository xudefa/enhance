# config 包 — 配置管理

> **所属层级**: Core Layer  
> **设计理念**: 分层配置，热重载支持  
> **设计灵感**: Spring Environment + Spring Cloud RefreshScope

## 概述

`config` 包提供完整的配置管理解决方案，合并了原 `config/`、`environment/` 和 `refresh/` 三个包的功能。支持分层配置源、Profile 机制、配置验证、热重载等特性。

### 核心功能

| 功能 | 说明 |
|------|------|
| **统一配置接口** | `Config` 接口定义与多种类型安全的获取方法 |
| **分层配置源** | `Environment` 提供分层配置源管理和 Profile 机制 |
| **配置加载器链** | 可组合的 `Loader` 接口，支持文件/环境变量/远程配置源 |
| **配置验证** | 必填、范围、正则、枚举、自定义规则验证 |
| **热重载机制** | `RefreshScope` 支持配置变更时自动刷新 Bean |
| **属性绑定** | 结构体标签绑定，支持默认值和类型转换 |
| **配置中心集成** | Nacos/Etcd/Apollo 等远程配置中心抽象 |

### 子包结构

```
config/
├── config.go              # Config 接口定义与内存实现
├── builder.go             # ConfigBuilder 配置模型构建器
├── loader.go              # Loader 接口与 LoaderChain 加载器链
├── validator.go           # Validator 接口与验证规则
├── watcher.go             # WatchManager 热重载管理器
├── properties.go          # PropertyBinder 属性绑定器
├── configcenter.go        # 配置中心接口（Nacos/Etcd/Apollo）
├── environment/           # 环境配置管理子包
│   ├── environment.go     # Environment 环境管理器
│   └── source.go          # PropertySource 配置源接口与实现
└── refresh/               # 热重载机制子包
    ├── scope.go           # RefreshScopeManager 刷新管理器
    ├── proxy.go           # RefreshProxy 代理模式
    ├── event.go           # 配置变更事件类型
    ├── builder.go         # RefreshProxyBuilder 代理构建器
    └── router.go          # 事件路由与分发
```

---

## 核心接口

### Config 接口

```go
type Config interface {
    // 值获取方法
    Get(key string) any
    GetAll() map[string]any
    GetString(key string) string
    GetStringMap(key string) map[string]any
    GetStringMapString(key string) map[string]string
    GetStringSlice(key string) []string
    GetInt(key string) int
    GetInt64(key string) int64
    GetIntSlice(key string) []int
    GetFloat64(key string) float64
    GetBool(key string) bool

    // 键操作方法
    HasKey(key string) bool

    // 结构体映射方法
    Unmarshal(target any) error
    UnmarshalKey(key string, target any) error

    // 热重载方法
    Watch(callback func(WatchEvent)) error
    StopWatch() error

    // 配置源信息
    GetSource() string
}
```

### WatchEvent 结构

```go
type WatchEvent struct {
    Type     string   // "modify" | "delete" | "create"
    Key      string   // 变更的键名
    OldValue any      // 旧值
    NewValue any      // 新值
}
```

---

## 快速开始

### 创建配置

```go
import "github.com/xudefa/enhance/config"

// 使用函数式选项创建配置模型
cfg, err := config.New(
    loadFn,                          // 加载函数（最后执行）
    config.WithConfigName("app"),
    config.WithConfigPath("./config", "/etc/app"),
    config.WithConfigType("yaml"),
    config.WithEnvironment("prod"),
    config.WithEnvVariable("APP"),
    config.WithConfigFile("/etc/app/config.yaml"),
    config.WithDefaultEnv(),
)
```

### 获取配置值

```go
// 基本类型获取
cfg.Get("server.port")                  // any — 原始值
cfg.GetString("app.name")               // string
cfg.GetInt("server.port")               // int
cfg.GetInt64("timeout")                 // int64
cfg.GetFloat64("threshold")             // float64
cfg.GetBool("server.enabled")           // bool
cfg.GetStringSlice("allowed.origins")   // []string
cfg.GetStringMap("database")            // map[string]any
cfg.GetStringMapString("database")      // map[string]string
cfg.GetIntSlice("ports")                // []int
cfg.HasKey("server.host")               // bool
cfg.GetAll()                            // map[string]any — 所有配置
```

### 结构体绑定

```go
type AppConfig struct {
    Name    string `mapstructure:"name"`
    Version string `mapstructure:"version"`
    Server  ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

var appCfg AppConfig

// 全量映射
if err := cfg.Unmarshal(&appCfg); err != nil {
    log.Fatal(err)
}

// 指定前缀映射
var srvCfg ServerConfig
if err := cfg.UnmarshalKey("server", &srvCfg); err != nil {
    log.Fatal(err)
}
```

---

## API 参考

### 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithConfigName(name)` | 配置文件名（不含扩展名） | `"config"` |
| `WithConfigPath(paths...)` | 搜索路径 | `["./", "./config"]` |
| `WithConfigType(type)` | 配置类型 | `"json"` |
| `WithEnvironment(env)` | 环境名称（附加到文件名） | `""` |
| `WithConfigFile(path)` | 完整路径（优先于名称+路径） | `""` |
| `WithEnvVariable(name)` | 环境变量前缀 | `""` |
| `WithDefaultEnv()` | 自动检测环境（APP_ENV → GO_ENV → ENV） | - |

### WithDefaultEnv 检测顺序

1. 检查 `APP_ENV` 环境变量
2. 检查 `GO_ENV` 环境变量
3. 检查 `ENV` 环境变量
4. 均未设置时使用空字符串（即无需环境后缀）

### 热重载

```go
// 注册变更监听
err := cfg.Watch(func(ev config.WatchEvent) {
    fmt.Printf("配置变更: %s %s = %v → %v\n",
        ev.Type, ev.Key, ev.OldValue, ev.NewValue)
})

// 停止监听
err := cfg.StopWatch()
```

---

## 子包：environment — 环境配置管理

提供分层配置源（PropertySource）管理和 Profile 机制。

### 配置源优先级（从高到低）

| 优先级 | 常量 | 配置源 | 说明 |
|--------|------|--------|------|
| 最高 (4) | `PriorityHighest` | `ArgsPropertySource` | 命令行参数 `--key=value` |
| 高 (3) | `PriorityHigh` | `EnvPropertySource` | 环境变量（默认 `GO_BOOT_` 前缀） |
| 中 (2) | `PriorityNormal` | `MapPropertySource` | 用户动态添加的配置源 |
| 中 (1) | `PriorityLow` | - | 保留 |
| 最低 (0) | `PriorityLowest` | 应用配置源 | `application.json` 或 `JSONPropertySource` |

### 快速开始

```go
import "github.com/xudefa/enhance/config/environment"

env := environment.NewEnvironment()

// 获取属性值
port, ok := env.GetProperty("server.port")
name, ok := env.GetRequiredProperty("app.name")

// 类型安全获取
portInt := env.GetIntegerProperty("server.port", 8080)
enabled := env.GetBoolProperty("server.enabled", true)

// Profile 机制
env.AddActiveProfile("prod")
if env.AcceptsProfiles("prod") {
    // 生产环境逻辑
}

// 结构体绑定
type ServerConfig struct {
    Host string `enhance:"server.host"`
    Port int    `enhance:"server.port"`
}
var cfg ServerConfig
env.BindProperties(&cfg)
```

---

## 子包：refresh — 热重载机制

参考 Spring Cloud 的 `@RefreshScope`，实现配置变更时 Bean 的自动刷新。

### 核心特性

- **代理模式** — 使用代理实现平滑切换，旧请求继续使用旧实例，新请求使用新实例
- **延迟刷新** — Bean 消费者持有代理引用，代理内部在刷新标记被设置后重新创建目标实例
- **事件驱动** — 统一的事件接口，支持多种配置源
- **指标收集** — 内置刷新指标，监控刷新性能

### 快速开始

```go
import (
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/config/refresh"
)

// 创建刷新管理器
manager := refresh.NewRefreshScopeManager(beanCreator, logger,
    refresh.WithRefreshEnabled(true),
    refresh.WithRefreshDelay(100*time.Millisecond),
    refresh.WithMaxRefreshAttempts(3),
)

// 注册可刷新 Bean
manager.RegisterRefreshableBean("myService", myService)

// 标记 Bean 需要刷新（配置变更时调用）
manager.MarkBeanForRefresh("myService")

// 获取刷新后的 Bean
bean, err := manager.GetRefreshedBean("myService")
```

### RefreshableBean 接口

```go
type RefreshableBean interface {
    OnConfigChange(event refresh.ConfigChangeEvent) error
}

type MyService struct {
    config *Config
}

func (s *MyService) OnConfigChange(event refresh.ConfigChangeEvent) error {
    // 自定义配置变更处理逻辑
    s.config.Reload()
    return nil
}
```

### 配置变更事件

```go
event := refresh.NewConfigChangeEvent(
    "modify",
    []string{"db.host", "db.port"},
    oldValues,
    newValues,
    "nacos",
)
```

---

## 使用示例

### 完整配置加载示例

```go
package main

import (
    "fmt"
    "log"
    "github.com/xudefa/enhance/config"
)

func main() {
    // 创建配置
    cfg, err := config.New(
        func() (map[string]any, error) {
            return map[string]any{
                "app.name":    "my-app",
                "app.version": "1.0.0",
                "server.host": "localhost",
                "server.port": 8080,
            }, nil
        },
        config.WithConfigName("app"),
        config.WithConfigPath("./config"),
        config.WithConfigType("yaml"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 获取配置值
    fmt.Println("App Name:", cfg.GetString("app.name"))
    fmt.Println("Server Port:", cfg.GetInt("server.port"))

    // 结构体绑定
    type AppConfig struct {
        Name   string `mapstructure:"app.name"`
        Server struct {
            Host string `mapstructure:"server.host"`
            Port int    `mapstructure:"server.port"`
        } `mapstructure:",squash"`
    }
    
    var appCfg AppConfig
    if err := cfg.Unmarshal(&appCfg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Config: %+v\n", appCfg)
}
```

### 热重载示例

```go
// 注册配置变更监听
err := cfg.Watch(func(ev config.WatchEvent) {
    fmt.Printf("配置变更: %s %s = %v → %v\n",
        ev.Type, ev.Key, ev.OldValue, ev.NewValue)
    
    // 更新应用配置
    updateAppConfig(ev.Key, ev.NewValue)
})
if err != nil {
    log.Fatal(err)
}

// 应用退出时停止监听
defer cfg.StopWatch()
```

---

## 最佳实践

### 1. 使用结构体绑定替代手动获取

```go
// ✅ 推荐：结构体绑定
type ServerConfig struct {
    Host string `mapstructure:"server.host"`
    Port int    `mapstructure:"server.port"`
}
var cfg ServerConfig
cfg.Unmarshal(&cfg)

// ⚠️ 不推荐：手动获取每个字段
host := cfg.GetString("server.host")
port := cfg.GetInt("server.port")
```

### 2. 合理使用配置源优先级

```go
// ✅ 推荐：命令行参数覆盖配置文件
// 启动命令：./app --server.port=9090
// 优先级：命令行 > 环境变量 > 配置文件 > 默认值
```

### 3. 使用 Profile 区分环境

```go
// ✅ 推荐：使用 Profile 区分环境
env := environment.NewEnvironment()
env.AddActiveProfile("prod")

// 加载 application-prod.json
// ⚠️ 不推荐：硬编码环境特定逻辑
if os.Getenv("ENV") == "prod" {
    // 生产环境逻辑
}
```

### 4. 热重载时注意线程安全

```go
// ✅ 推荐：使用原子操作或互斥锁保护配置更新
type ConfigManager struct {
    mu     sync.RWMutex
    config *AppConfig
}

func (m *ConfigManager) Update(cfg *AppConfig) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.config = cfg
}

func (m *ConfigManager) Get() *AppConfig {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.config
}
```