# Zap Starter

Zap 高性能日志框架自动配置模块，提供结构化日志支持。

## 功能特性

- ✅ 自动配置 Zap 日志器
- ✅ 支持 JSON/Console 格式
- ✅ 支持日志级别配置
- ✅ 支持文件输出和日志轮转
- ✅ 实现 log.Logger 接口

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/zap"
)
```

### 2. 配置文件

在 `application.json` 中添加 Zap 配置：

```json
{
  "log": {
    "zap": {
      "enabled": true,
      "level": "info",
      "format": "json",
      "output-path": "stdout"
    }
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/log"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("zap-demo"),
    )
    defer app.Stop()
    
    // 获取 Logger 实例
    logger := core.MustGetBean[log.Logger](app.Container())
    
    // 记录日志
    logger.Info(context.Background(), "Application started",
        log.KeyValue{Key: "version", Value: "1.0.0"},
        log.KeyValue{Key: "port", Value: 8080},
    )
    
    // 带字段的日志
    logger.Error(context.Background(), "Failed to connect",
        log.KeyValue{Key: "error", Value: "connection refused"},
    )
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `log.zap.enabled` | bool | false | 是否启用 Zap |
| `log.zap.level` | string | info | 日志级别（debug/info/warn/error/fatal） |
| `log.zap.format` | string | json | 日志格式（json/console） |
| `log.zap.output-path` | string | stdout | 输出路径（stdout 或文件路径） |

## 日志级别

| 级别 | 说明 |
|------|------|
| `debug` | 调试信息 |
| `info` | 一般信息 |
| `warn` | 警告信息 |
| `error` | 错误信息 |
| `fatal` | 致命错误 |

## 高级用法

### 文件输出

```json
{
  "log": {
    "zap": {
      "enabled": true,
      "level": "info",
      "format": "json",
      "output-path": "logs/app.log"
    }
  }
}
```

### 自定义字段

```go
logger := core.MustGetBean[log.Logger](app.Container())

// 创建带字段的日志记录器
childLogger := logger.With(context.Background(),
    log.KeyValue{Key: "service", Value: "user-service"},
    log.KeyValue{Key: "version", Value: "1.0.0"},
)

childLogger.Info(context.Background(), "Request processed")
```

## 启动顺序

- **优先级**: `OrderPriorityInfrastructure` (-4000)
- **触发条件**: `log.zap.enabled=true`

## 依赖

- `go.uber.org/zap`
- `gopkg.in/natefinch/lumberjack.v2`