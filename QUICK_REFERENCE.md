# enhance 框架快速参考

本文档提供 enhance 框架优化后的 API 快速参考。

---

## P0: 约定优于配置

### 条件判断

```go
// 旧方式：必须显式配置 gin.enabled=true
condition.OnProperty("gin.enabled", "true")

// 新方式：未配置时默认启用
condition.OnPropertyOrDefault("gin.enabled", "true", "true")
```

### 已更新的 Starter

以下 Starter 已更新为默认启用（无需配置）：
- ✅ `gin` - Web 框架
- ✅ `gorm` - ORM
- ✅ `redis` - 缓存
- ✅ `zerolog` - 日志

---

## P0: 开发者体验优化

### 启动报告

默认启用，启动时自动打印：

```go
app := boot.New(
    boot.WithAppName("my-app"),
    boot.WithVersion("1.0.0"),
)
```

禁用启动报告：

```go
app := boot.New(
    boot.WithAppName("my-app"),
    boot.WithoutStartupReport(),
)
```

---

## P1: 模块化组合 API

### DSL 风格模块定义

```go
var DatabaseModule = boot.ModuleDef("database").
    DependsOn("config").                    // 声明依赖
    Beans(                                  // 注册 Bean
        boot.Provide(NewDatabase),
        boot.Provide(NewRepository),
    ).
    Starters(                               // 注册 Starter
        &MigrationStarter{},
    ).
    Hooks(                                  // 注册钩子
        lifecycle.OnInitFunc(initDB),
    ).
    OnProperty("database.enabled").         // 条件
    Build()
```

### 模块分组

```go
var BackendModules = boot.ModuleGroup("backend").
    Include(DatabaseModule, CacheModule, WebModule).
    Build()

app := boot.New(
    boot.WithModulesList(BackendModules...),
)
```

### 便捷函数

```go
// 从函数创建模块
var DBModule = boot.ModuleFromFunc("database", 
    func(c core.Container) error {
        // 安装逻辑
        return nil
    },
    boot.WithModuleConditions(
        condition.OnProperty("db.enabled"),
    ),
)

// 生命周期钩子
boot.LifecycleHook(onInit, onStart, onStop)
boot.OnInitHook(fn)
boot.OnStartHook(fn)
boot.OnStopHook(fn)
```

---

## P2: 插件系统

### 定义插件

```go
type MonitoringPlugin struct{}

func (p *MonitoringPlugin) Name() string { 
    return "monitoring" 
}
func (p *MonitoringPlugin) Version() string { 
    return "1.0.0" 
}
func (p *MonitoringPlugin) Dependencies() []string { 
    return []string{"database", "web"} 
}
func (p *MonitoringPlugin) Init(ctx boot.PluginContext) error {
    // 初始化逻辑
    return nil
}
func (p *MonitoringPlugin) Start(ctx boot.PluginContext) error {
    // 启动逻辑
    return nil
}
func (p *MonitoringPlugin) Stop(ctx boot.PluginContext) error {
    // 停止逻辑
    return nil
}
```

### 注册插件

```go
// 单个插件
app := boot.New(
    boot.WithAppName("my-app"),
    boot.WithPlugin(&MonitoringPlugin{}),
)

// 多个插件
app := boot.New(
    boot.WithAppName("my-app"),
    boot.WithPlugins(&PluginA{}, &PluginB{}, &PluginC{}),
)
```

### 插件生命周期

```
Registered → Initialized → Started → Stopped
                  ↓
               Error (if failed)
```

- **依赖顺序**：按依赖关系拓扑排序
- **初始化**：先初始化依赖插件
- **启动**：按依赖顺序启动
- **停止**：逆序停止

---

## 完整示例

```go
package main

import (
    "context"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
)

// 1. 定义模块
var DatabaseModule = boot.ModuleDef("database").
    Beans(
        boot.Provide(NewDatabase),
    ).
    Build()

// 2. 定义插件
type MonitoringPlugin struct{}

func (p *MonitoringPlugin) Name() string { return "monitoring" }
func (p *MonitoringPlugin) Version() string { return "1.0.0" }
func (p *MonitoringPlugin) Dependencies() []string { return []string{"database"} }
func (p *MonitoringPlugin) Init(ctx boot.PluginContext) error { return nil }
func (p *MonitoringPlugin) Start(ctx boot.PluginContext) error { return nil }
func (p *MonitoringPlugin) Stop(ctx boot.PluginContext) error { return nil }

func main() {
    // 3. 创建应用
    app := boot.New(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
        
        // 使用模块
        boot.WithModules(DatabaseModule),
        
        // 使用插件
        boot.WithPlugin(&MonitoringPlugin{}),
    )
    
    // 4. 启动（自动打印启动报告）
    if err := app.Start(); err != nil {
        panic(err)
    }
    
    // 5. 停止
    defer app.Stop(context.Background())
    
    select {} // 阻塞
}
```

---

## API 对比

| 功能 | 旧 API | 新 API |
|-----|--------|--------|
| 条件判断 | `OnProperty(key, value)` | `OnPropertyOrDefault(key, default, value)` |
| 模块定义 | `Module{Beans: []...}` | `ModuleDef("name").Beans(...).Build()` |
| 模块分组 | `MergeModules(...)` | `ModuleGroup("name").Include(...).Build()` |
| 生命周期 | 手动管理 | `Plugin` 接口自动管理 |
| 启动报告 | 无 | 自动打印 |

---

## 测试

```bash
# 运行所有测试
go test ./... -short

# 运行特定包测试
go test ./boot/ -v
go test ./condition/ -v

# 运行特定测试
go test ./boot/ -run "TestStartupReport" -v
go test ./boot/ -run "TestPluginManager" -v
go test ./boot/ -run "TestModuleComposer" -v
```

---

## 更多信息

- 详细文档：`OPTIMIZATION_SUMMARY.md`
- 示例代码：`examples/enhanced-demo/main.go`
- 测试用例：`boot/*_test.go`