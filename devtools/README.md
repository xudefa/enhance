# devtools 包 — 开发工具

> **所属层级**: Infrastructure Layer  
> **设计理念**: 热重载支持，开发效率提升  
> **设计灵感**: Spring Boot DevTools

## 概述

`devtools` 包提供开发环境下的热重载功能支持，参考 Spring Boot DevTools 设计。通过监控文件变化并自动触发回调，实现配置热加载、模板刷新等开发辅助功能。

### 核心功能

| 功能 | 说明 |
|------|------|
| **文件监控** | 支持多目录监控，自动检测文件变化 |
| **热重载** | 文件变化时触发回调，实现热加载 |
| **开发模式检测** | 自动检测是否为开发环境 |
| **实时通知** | 支持 LiveReload 服务器通知客户端 |

---

## 核心接口

### HotReloader 热重载管理器

```go
type HotReloader struct {
    // ...
}
```

#### 创建

```go
reloader := devtools.NewHotReloader(
    devtools.WithWatchDirs("config", "templates"),
    devtools.WithExtensions(".json", ".yaml", ".yml"),
    devtools.WithInterval(2*time.Second),
    devtools.WithIgnoreDirs(".git", "node_modules"),
)
```

#### 选项函数

| 函数 | 说明 | 默认值 |
|------|------|--------|
| `WithWatchDirs(dirs...)` | 设置监控目录 | 无 |
| `WithExtensions(exts...)` | 设置监控的文件扩展名 | 所有文件 |
| `WithInterval(duration)` | 设置轮询间隔 | `2s` |
| `WithIgnoreDirs(dirs...)` | 设置忽略的目录 | `.git`, `node_modules`, `vendor` |

#### 注册回调

```go
reloader.OnReload(func(event devtools.ReloadEvent) {
    fmt.Printf("File changed: %s (%s)\n", event.File, event.Type)
})
```

#### 启动与停止

```go
// 启动文件监控
if err := reloader.Start(); err != nil {
    // 处理错误
}

// 停止文件监控
reloader.Stop()

// 重启热重载
if err := reloader.Restart(); err != nil {
    // 处理错误
}
```

#### 状态查询

```go
// 检查是否正在运行
if reloader.IsRunning() {
    // 热重载正在运行
}

// 获取所有被监控的文件
files := reloader.GetWatchedFiles()
```

### ReloadEvent 重载事件

```go
type ReloadEvent struct {
    File      string
    Type      ReloadType
    Timestamp time.Time
    OldHash   string
    NewHash   string
}
```

#### ReloadType 重载类型

| 常量 | 值 | 说明 |
|------|----|------|
| `ReloadTypeCreated` | `"CREATED"` | 文件创建 |
| `ReloadTypeModified` | `"MODIFIED"` | 文件修改 |
| `ReloadTypeDeleted` | `"DELETED"` | 文件删除 |

### DevModeDetector 开发模式检测器

```go
type DevModeDetector struct {
    // ...
}
```

#### 创建和检测

```go
detector := devtools.NewDevModeDetector()

if detector.IsDevMode() {
    // 开发模式,启用热重载
    reloader.Start()
}
```

检测逻辑:
- 检查环境变量 `DEV_MODE`、`DEVELOPMENT`、`GO_ENV`
- 值为 `true`、`development` 或 `dev` 时返回 `true`

### LiveReloadServer 实时重载服务器

```go
type LiveReloadServer struct {
    // ...
}
```

#### 创建和使用

```go
server := devtools.NewLiveReloadServer(35729, reloader)

// 启动服务器
if err := server.Start(); err != nil {
    // 处理错误
}

// 停止服务器
server.Stop()
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/devtools"
)

func main() {
    // 创建热重载管理器
    reloader := devtools.NewHotReloader(
        devtools.WithWatchDirs("config"),
        devtools.WithExtensions(".json", ".yaml"),
    )

    // 注册重载回调
    reloader.OnReload(func(event devtools.ReloadEvent) {
        fmt.Printf("File %s: %s\n", event.File, event.Type)
        // 重新加载配置
        reloadConfig(event.File)
    })

    // 启动监控
    if err := reloader.Start(); err != nil {
        panic(err)
    }
    defer reloader.Stop()

    // 阻塞主进程
    select {}
}
```

---

## API 参考

### 开发模式检测

```go
detector := devtools.NewDevModeDetector()

if detector.IsDevMode() {
    // 开发模式: 启用热重载
    reloader := devtools.NewHotReloader(
        devtools.WithWatchDirs("config", "templates"),
    )
    reloader.OnReload(handleReload)
    reloader.Start()
    defer reloader.Stop()
} else {
    // 生产模式: 一次性加载配置
    loadConfig()
}
```

### 多目录监控

```go
reloader := devtools.NewHotReloader(
    devtools.WithWatchDirs(
        "config",
        "templates",
        "static",
    ),
    devtools.WithExtensions(
        ".json",
        ".yaml",
        ".html",
        ".css",
        ".js",
    ),
    devtools.WithInterval(1*time.Second),
)

// 不同类型的文件使用不同的处理逻辑
reloader.OnReload(func(event devtools.ReloadEvent) {
    ext := filepath.Ext(event.File)
    switch ext {
    case ".json", ".yaml":
        reloadConfig(event.File)
    case ".html":
        reloadTemplate(event.File)
    case ".css", ".js":
        notifyBrowserReload()
    }
})
```

---

## 使用示例

### 配置热加载

```go
type ConfigManager struct {
    config *Config
    mu     sync.RWMutex
}

func (m *ConfigManager) ReloadConfig(file string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    newConfig, err := loadConfigFromFile(file)
    if err != nil {
        return err
    }
    
    m.config = newConfig
    return nil
}

func main() {
    configManager := &ConfigManager{}
    
    reloader := devtools.NewHotReloader(
        devtools.WithWatchDirs("config"),
        devtools.WithExtensions(".yaml"),
    )
    
    reloader.OnReload(func(event devtools.ReloadEvent) {
        if err := configManager.ReloadConfig(event.File); err != nil {
            log.Printf("Failed to reload config: %v", err)
        } else {
            log.Printf("Config reloaded: %s", event.File)
        }
    })
    
    reloader.Start()
    defer reloader.Stop()
}
```

### 与 Web 框架集成

```go
func main() {
    // 检测开发模式
    detector := devtools.NewDevModeDetector()
    
    if detector.IsDevMode() {
        // 启用 LiveReload
        reloader := devtools.NewHotReloader(
            devtools.WithWatchDirs("templates", "static"),
        )
        
        liveReloadServer := devtools.NewLiveReloadServer(35729, reloader)
        liveReloadServer.Start()
        defer liveReloadServer.Stop()
        
        reloader.Start()
        defer reloader.Stop()
    }
    
    // 启动 Web 服务器
    http.ListenAndServe(":8080", mux)
}
```

---

## 最佳实践

### 1. 根据环境启用热重载

```go
// ✅ 推荐：检测开发模式
detector := devtools.NewDevModeDetector()
if detector.IsDevMode() {
    reloader := devtools.NewHotReloader(
        devtools.WithWatchDirs("config", "templates"),
    )
    reloader.Start()
    defer reloader.Stop()
}

// ⚠️ 不推荐：生产环境也启用热重载
reloader := devtools.NewHotReloader(
    devtools.WithWatchDirs("config"),
)
reloader.Start()
```

### 2. 合理设置监控目录和扩展名

```go
// ✅ 推荐：只监控必要的目录和文件类型
reloader := devtools.NewHotReloader(
    devtools.WithWatchDirs("config", "templates"),
    devtools.WithExtensions(".yaml", ".html"),
    devtools.WithIgnoreDirs(".git", "node_modules", "vendor"),
)

// ⚠️ 不推荐：监控整个项目目录
reloader := devtools.NewHotReloader(
    devtools.WithWatchDirs("."),
)
```

### 3. 处理重载错误

```go
// ✅ 推荐：处理重载错误
reloader.OnReload(func(event devtools.ReloadEvent) {
    if err := reloadConfig(event.File); err != nil {
        log.Printf("Failed to reload %s: %v", event.File, err)
        // 保留旧配置，不中断服务
    } else {
        log.Printf("Successfully reloaded %s", event.File)
    }
})

// ⚠️ 不推荐：忽略错误
reloader.OnReload(func(event devtools.ReloadEvent) {
    reloadConfig(event.File) // 错误被忽略
})
```

### 4. 使用 LiveReload 提升开发体验

```go
// ✅ 推荐：启用 LiveReload
liveReloadServer := devtools.NewLiveReloadServer(35729, reloader)
liveReloadServer.Start()
defer liveReloadServer.Stop()

// 在 HTML 模板中注入 LiveReload 脚本
func injectLiveReloadScript(html string) string {
    script := `<script src="http://localhost:35729/livereload.js"></script>`
    return strings.Replace(html, "</body>", script+"</body>", 1)
}

// ⚠️ 不推荐：手动刷新浏览器
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将 HotReloader 注册为 Bean
container.Register(
    reflect.TypeOf(&devtools.HotReloader{}),
    core.Bean(createHotReloader()),
    core.Singleton(),
)

// 注入使用
type ConfigService struct {
    Reloader *devtools.HotReloader `inject:"hotReloader"`
}

func (s *ConfigService) Start() error {
    return s.Reloader.Start()
}
```