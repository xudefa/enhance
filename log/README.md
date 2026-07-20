# log 包 — 日志抽象层

> **所属层级**: Infrastructure Layer  
> **设计理念**: 统一接口，多实现支持  
> **设计灵感**: SLF4J + Go slog

## 概述

`log` 包提供统一的日志记录接口抽象，允许不同的日志库实现无缝替换。内置基于 Go 标准库 `log/slog` 的默认实现 `SlogLogger`，支持 JSON 和 text 两种输出格式。

### 核心功能

| 功能 | 说明 |
|------|------|
| **统一接口** | Logger 接口定义标准日志记录操作 |
| **多级别支持** | Debug、Info、Warn、Error、DPanic、Panic、Fatal |
| **结构化日志** | KeyValue 支持结构化数据记录 |
| **多格式输出** | JSON 和 text 两种输出格式 |
| **扩展接口** | 支持级别、名称、调用者、超时等扩展功能 |
| **依赖注入** | 可与 IoC 容器集成 |

---

## 核心接口

### Logger 接口

```go
type Logger interface {
    Debug(ctx context.Context, msg string, keys ...KeyValue)
    Info(ctx context.Context, msg string, keys ...KeyValue)
    Warn(ctx context.Context, msg string, keys ...KeyValue)
    Error(ctx context.Context, msg string, keys ...KeyValue)
    DPanic(ctx context.Context, msg string, keys ...KeyValue)
    Panic(ctx context.Context, msg string, keys ...KeyValue)
    Fatal(ctx context.Context, msg string, keys ...KeyValue)
    Sync() error
    With(ctx context.Context, keys ...KeyValue) Logger
}
```

### 方法说明

| 方法 | 说明 |
|------|------|
| `Debug` | 调试日志 |
| `Info` | 信息日志 |
| `Warn` | 警告日志 |
| `Error` | 错误日志 |
| `DPanic` | 致命错误日志，开发环境触发 panic |
| `Panic` | 记录日志并 panic |
| `Fatal` | 记录日志（注意：不主动调用 `os.Exit(1)`） |
| `Sync` | 同步日志缓冲区 |
| `With` | 返回带有额外固定字段的新 Logger |

### Level — 日志级别

```go
type Level int8

const (
    DebugLevel  Level = iota // 调试级别
    InfoLevel                // 信息级别
    WarnLevel                // 警告级别
    ErrorLevel               // 错误级别
    DPanicLevel              // 致命错误级别（开发环境 panic）
    PanicLevel               // panic 级别
    FatalLevel               // 致命级别（程序退出）
)
```

字符串表示：`"debug"`、`"info"`、`"warn"`、`"error"`、`"dpanic"`、`"panic"`、`"fatal"`。

### KeyValue — 结构化键值对

```go
type KeyValue struct {
    Key   string
    Value any
}
```

用于结构化日志记录：

```go
log.KeyValue{Key: "user_id", Value: 123}
log.KeyValue{Key: "duration", Value: time.Second}
```

### 扩展接口

| 接口 | 说明 |
|------|------|
| `LoggerWithLevel` | 支持在运行时使用任意级别记录日志 |
| `LoggerWithName` | 支持为 Logger 设置名称（如模块名） |
| `LoggerWithCaller` | 支持记录调用者源码位置信息 |
| `LoggerWithTimeout` | 支持超时自动刷盘 |

```go
type LoggerWithLevel interface {
    Logger
    Log(ctx context.Context, level Level, msg string, keys ...KeyValue)
}

type LoggerWithName interface {
    Logger
    WithName(name string) Logger
}

type LoggerWithCaller interface {
    Logger
    WithCaller(skip int) Logger
}

type LoggerWithTimeout interface {
    Logger
    WithTimeout(d time.Duration) Logger
}
```

---

## 快速开始

### 创建日志记录器

```go
package main

import (
    "context"
    "github.com/xudefa/enhance/log"
)

func main() {
    // 默认配置（JSON 格式，Info 级别，标准输出）
    logger := log.NewSlogLogger()
    
    ctx := context.Background()
    
    logger.Info(ctx, "服务启动", log.KeyValue{Key: "port", Value: 8080})
    logger.Debug(ctx, "SQL 查询", log.KeyValue{Key: "sql", Value: "SELECT * FROM users"})
    logger.Error(ctx, "请求失败",
        log.KeyValue{Key: "path", Value: "/api/users"},
        log.KeyValue{Key: "status", Value: 500},
    )
}
```

### 自定义配置

```go
logger := log.NewSlogLogger(
    log.WithLevel(log.DebugLevel),
    log.WithFormat("text"),
    log.WithTimeFormat("2006-01-02 15:04:05"),
    log.WithAddSource(true),
    log.WithOutput(os.Stderr),
    log.WithOutputPath("/var/log/app.log"),
)
```

---

## API 参考

### SlogLogger — 基于 slog 的实现

#### 创建

```go
// 默认配置
logger := log.NewSlogLogger()

// 自定义配置
logger := log.NewSlogLogger(
    log.WithLevel(log.DebugLevel),
    log.WithFormat("text"),
    log.WithTimeFormat("2006-01-02 15:04:05"),
    log.WithAddSource(true),
    log.WithOutput(os.Stderr),
    log.WithOutputPath("/var/log/app.log"),
)
```

#### Option

| 选项 | 说明 |
|------|------|
| `WithLevel(level Level)` | 设置日志级别（默认 Info） |
| `WithFormat(format string)` | 输出格式：`"json"`（默认）或 `"text"` |
| `WithTimeFormat(timeFormat string)` | 时间格式（默认 `"2006-01-02 15:04:05"`） |
| `WithAddSource(addSource bool)` | 是否添加源码位置（默认 false） |
| `WithOutput(w io.Writer)` | 设置输出 Writer（默认 os.Stdout） |
| `WithOutputPath(path string)` | 设置日志文件输出路径 |

#### Close

关闭日志文件句柄（仅在设置了文件输出时需要调用）：

```go
if closer, ok := logger.(*log.SlogLogger); ok {
    defer closer.Close()
}
```

### Build — 日志记录器构建

```go
func Build(opts ...LoggerOption) Logger

type LoggerOption func(*loggerConfig)

func WithLogger(logger Logger) LoggerOption
```

使用示例：

```go
logger := log.Build(log.WithLogger(log.NewSlogLogger(
    log.WithFormat("text"),
)))
```

### ToLevel — 级别转换

将字符串转换为日志级别：

```go
level := log.ToLevel("info") // log.InfoLevel
level := log.ToLevel("warn") // log.WarnLevel
```

---

## 使用示例

### 基础日志记录

```go
logger := log.NewSlogLogger(
    log.WithLevel(log.DebugLevel),
    log.WithFormat("json"),
)

ctx := context.Background()

logger.Info(ctx, "服务启动", log.KeyValue{Key: "port", Value: 8080})
logger.Debug(ctx, "SQL 查询", log.KeyValue{Key: "sql", Value: "SELECT * FROM users"})
logger.Error(ctx, "请求失败",
    log.KeyValue{Key: "path", Value: "/api/users"},
    log.KeyValue{Key: "status", Value: 500},
)
```

### 使用 With 添加固定字段

```go
// 使用 With 添加固定字段
logger2 := logger.With(ctx, log.KeyValue{Key: "service", Value: "user-svc"})
logger2.Info(ctx, "用户注册") // 自动携带 service=user-svc
```

### 与依赖注入集成

```go
container.Register(
    reflect.TypeOf(&log.SlogLogger{}),
    core.Bean(log.NewSlogLogger(
        log.WithLevel(log.InfoLevel),
        log.WithFormat("json"),
    )),
    core.Singleton(),
)

type UserService struct {
    Logger log.Logger `inject:"logger"`
}
```

---

## 最佳实践

### 1. 使用结构化日志

```go
// ✅ 推荐：使用 KeyValue 结构化记录
logger.Info(ctx, "用户登录",
    log.KeyValue{Key: "user_id", Value: 123},
    log.KeyValue{Key: "ip", Value: "192.168.1.1"},
)

// ⚠️ 不推荐：使用字符串拼接
logger.Info(ctx, fmt.Sprintf("用户登录 user_id=%d ip=%s", 123, "192.168.1.1"))
```

### 2. 合理设置日志级别

```go
// ✅ 推荐：生产环境使用 Info 级别
logger := log.NewSlogLogger(log.WithLevel(log.InfoLevel))

// ✅ 推荐：开发环境使用 Debug 级别
logger := log.NewSlogLogger(log.WithLevel(log.DebugLevel))

// ⚠️ 不推荐：生产环境使用 Debug 级别，影响性能
logger := log.NewSlogLogger(log.WithLevel(log.DebugLevel))
```

### 3. 使用 With 添加上下文信息

```go
// ✅ 推荐：为每个请求添加 trace_id
requestLogger := logger.With(ctx,
    log.KeyValue{Key: "trace_id", Value: traceID},
    log.KeyValue{Key: "user_id", Value: userID},
)
requestLogger.Info(ctx, "处理请求")

// ⚠️ 不推荐：每次手动添加
logger.Info(ctx, "处理请求",
    log.KeyValue{Key: "trace_id", Value: traceID},
    log.KeyValue{Key: "user_id", Value: userID},
)
```

### 4. 输出到文件

```go
// ✅ 推荐：生产环境输出到文件
logger := log.NewSlogLogger(
    log.WithOutputPath("/var/log/app.log"),
    log.WithFormat("json"),
)
defer logger.(*log.SlogLogger).Close()

// ⚠️ 不推荐：忘记关闭文件句柄
logger := log.NewSlogLogger(log.WithOutputPath("/var/log/app.log"))
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将 Logger 注册为 Bean
container.Register(
    reflect.TypeOf(&log.SlogLogger{}),
    core.Bean(log.NewSlogLogger(
        log.WithLevel(log.InfoLevel),
        log.WithFormat("json"),
    )),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    Logger log.Logger `inject:"logger"`
}
```