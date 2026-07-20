# Zerolog Starter

Zerolog 高性能日志自动配置模块，提供结构化日志支持。

## 功能特性

- ✅ 自动配置 Zerolog 日志器
- ✅ 支持多种日志级别
- ✅ JSON 和控制台格式
- ✅ 文件输出支持
- ✅ 调用者信息记录

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/zerolog"
)
```

### 2. 配置文件

在 `application.json` 中添加 Zerolog 配置：

```json
{
  "log": {
    "zerolog": {
      "enabled": true,
      "level": "info",
      "format": "console",
      "time-format": "2006-01-02 15:04:05",
      "add-source": true,
      "output-path": ""
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
        boot.WithAppName("zerolog-demo"),
    )
    defer app.Stop()
    
    // 获取日志记录器
    logger := core.MustGetBean[log.Logger](app.Container())
    ctx := context.Background()
    
    // 记录不同级别的日志
    logger.Debug(ctx, "调试信息",
        log.KeyValue{Key: "user.id", Value: "123"},
    )
    
    logger.Info(ctx, "用户登录",
        log.KeyValue{Key: "user.name", Value: "John"},
        log.KeyValue{Key: "ip", Value: "192.168.1.100"},
    )
    
    logger.Warn(ctx, "磁盘空间不足",
        log.KeyValue{Key: "disk.usage", Value: "85%"},
    )
    
    logger.Error(ctx, "数据库连接失败",
        log.KeyValue{Key: "error", Value: err.Error()},
    )
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `log.zerolog.enabled` | bool | false | 是否启用 Zerolog |
| `log.zerolog.level` | string | info | 日志级别 (debug/info/warn/error/fatal/panic) |
| `log.zerolog.format` | string | console | 输出格式 (console/json) |
| `log.zerolog.time-format` | string | 2006-01-02 15:04:05 | 时间格式 |
| `log.zerolog.add-source` | bool | false | 是否添加调用者信息 |
| `log.zerolog.output-path` | string | "" | 输出文件路径（空则输出到 stdout） |

## 日志级别

| 级别 | 说明 |
|------|------|
| debug | 调试信息 |
| info | 一般信息 |
| warn | 警告信息 |
| error | 错误信息 |
| fatal | 致命错误（程序退出） |
| panic | 恐慌（程序 panic） |

## 高级用法

### 文件输出

```json
{
  "log": {
    "zerolog": {
      "enabled": true,
      "output-path": "/var/log/app.log"
    }
  }
}
```

### JSON 格式（生产环境推荐）

```json
{
  "log": {
    "zerolog": {
      "enabled": true,
      "format": "json"
    }
  }
}
```

### 添加调用者信息

```json
{
  "log": {
    "zerolog": {
      "enabled": true,
      "add-source": true
    }
  }
}
```

## 启动顺序

- **优先级**: `OrderPriorityInfrastructure` (-3000)
- **触发条件**: `log.zerolog.enabled=true`

## 依赖

- `github.com/rs/zerolog`