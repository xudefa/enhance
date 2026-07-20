# Cron Starter

Cron 定时任务自动配置模块，提供灵活的定时任务调度支持。

## 功能特性

- ✅ 自动配置 Cron 调度器
- ✅ 支持秒级定时任务
- ✅ 支持 Cron 表达式
- ✅ 任务恢复机制
- ✅ 任务管理接口

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/cron"
)
```

### 2. 配置文件

在 `application.json` 中添加 Cron 配置：

```json
{
  "cron": {
    "enabled": true,
    "with-logger": false
  }
}
```

### 3. 使用示例

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/robfig/cron/v3"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("cron-demo"),
    )
    defer app.Stop()
    
    // 获取 Cron 实例
    c := core.MustGetBean[*cron.Cron](app.Container())
    
    // 添加定时任务（每分钟执行）
    c.AddFunc("0 * * * * *", func() {
        fmt.Println("Every minute")
    })
    
    // 添加定时任务（每小时 30 分执行）
    c.AddFunc("0 30 * * * *", func() {
        fmt.Println("Every hour on the half hour")
    })
    
    // 添加定时任务（每天凌晨 2 点执行）
    c.AddFunc("0 0 2 * * *", func() {
        fmt.Println("Every day at 2 AM")
    })
    
    app.Start()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `cron.enabled` | bool | false | 是否启用 Cron |
| `cron.with-logger` | bool | false | 是否启用日志 |

## Cron 表达式

Cron 表达式格式：`秒 分 时 日 月 周`

| 字段 | 必填 | 允许值 | 特殊字符 |
|------|------|--------|----------|
| 秒 | 是 | 0-59 | * / , - |
| 分 | 是 | 0-59 | * / , - |
| 时 | 是 | 0-23 | * / , - |
| 日 | 是 | 1-31 | * / , - ? |
| 月 | 是 | 1-12 或 JAN-DEC | * / , - |
| 周 | 是 | 0-6 或 SUN-SAT | * / , - ? |

### 常用表达式

| 表达式 | 说明 |
|--------|------|
| `0 * * * * *` | 每分钟执行 |
| `0 0 * * * *` | 每小时执行 |
| `0 0 0 * * *` | 每天午夜执行 |
| `0 0 12 * * *` | 每天中午 12 点执行 |
| `0 0 0 1 * *` | 每月 1 号执行 |
| `0 0 0 * * 0` | 每周日执行 |
| `*/10 * * * * *` | 每 10 秒执行 |

## 高级用法

### 任务管理

```go
c := core.MustGetBean[*cron.Cron](app.Container())

// 添加任务并获取 ID
id, _ := c.AddFunc("0 * * * * *", func() {
    fmt.Println("Task running")
})

// 移除任务
c.Remove(id)

// 获取所有任务
entries := c.Entries()
for _, entry := range entries {
    fmt.Printf("ID: %d, Next: %v\n", entry.ID, entry.Next)
}
```

### 任务链

```go
// 使用任务链包装器
c.AddJob("0 * * * * *", cron.NewChain(
    cron.SkipIfStillRunning(cron.DefaultLogger),
).Then(cron.FuncJob(func() {
    fmt.Println("Running job")
})))
```

## 启动顺序

- **优先级**: `OrderPriorityTaskLayer` (-500)
- **触发条件**: `cron.enabled=true`

## 依赖

- `github.com/robfig/cron/v3`