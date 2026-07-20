# Asynq Starter

Asynq 异步任务队列自动配置模块，提供基于 Redis 的任务调度支持。

## 功能特性

- ✅ 自动配置 Asynq 客户端和调度器
- ✅ 支持任务重试
- ✅ 支持定时任务
- ✅ 支持任务优先级
- ✅ 支持任务进度追踪

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/asynq"
)
```

### 2. 配置文件

在 `application.json` 中添加 Asynq 配置：

```json
{
  "asynq": {
    "enabled": true,
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0,
    "pool-size": 10,
    "enable-scheduler": false
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/hibiken/asynq"
)

// 定义任务类型
const (
    TypeEmailDelivery = "email:deliver"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("asynq-demo"),
    )
    defer app.Stop()
    
    // 获取 Asynq Client
    client := core.MustGetBean[*asynq.Client](app.Container())
    
    // 创建任务
    payload, _ := json.Marshal(map[string]interface{}{
        "to":      "user@example.com",
        "subject": "Welcome",
        "body":    "Welcome to our platform!",
    })
    
    task := asynq.NewTask(TypeEmailDelivery, payload)
    
    // 添加任务到队列
    info, err := client.Enqueue(task,
        asynq.MaxRetry(3),
        asynq.Timeout(5*time.Minute),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Enqueued task: %s\n", info.ID)
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `asynq.enabled` | bool | false | 是否启用 Asynq |
| `asynq.host` | string | localhost | Redis 主机地址 |
| `asynq.port` | int | 6379 | Redis 端口 |
| `asynq.password` | string | "" | Redis 密码 |
| `asynq.db` | int | 0 | Redis 数据库 |
| `asynq.pool-size` | int | 10 | 连接池大小 |
| `asynq.enable-scheduler` | bool | false | 是否启用调度器 |

## 高级用法

### 任务处理器

```go
// 创建任务处理器
mux := asynq.NewServeMux()

// 注册任务处理器
mux.HandleFunc(TypeEmailDelivery, func(ctx context.Context, t *asynq.Task) error {
    var payload map[string]interface{}
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return err
    }
    
    fmt.Printf("Sending email to %s\n", payload["to"])
    // 发送邮件逻辑
    
    return nil
})

// 启动任务服务器
srv := asynq.NewServer(
    asynq.RedisClientOpt{Addr: "localhost:6379"},
    asynq.Config{
        Concurrency: 10,
        Queues: map[string]int{
            "critical": 6,
            "default":  3,
            "low":      1,
        },
    },
)

srv.Run(mux)
```

### 定时任务

```go
scheduler := core.MustGetBean[*asynq.Scheduler](app.Container())

// 添加定时任务
scheduler.Register(
    &asynq.PeriodicTask{
        Period: 24 * time.Hour,
        Task:   asynq.NewTask("report:generate", nil),
    },
)

// 启动调度器
scheduler.Run()
```

## 启动顺序

- **优先级**: `OrderPriorityTaskLayer` (-500)
- **触发条件**: `asynq.enabled=true`

## 依赖

- `github.com/hibiken/asynq`