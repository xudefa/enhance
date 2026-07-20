# RabbitMQ Starter

RabbitMQ 消息队列自动配置模块，提供可靠的消息队列支持。

## 功能特性

- ✅ 自动配置 RabbitMQ 连接
- ✅ 消息发布和订阅
- ✅ 队列声明支持
- ✅ 持久化配置
- ✅ 优雅关闭支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/rabbitmq"
)
```

### 2. 配置文件

在 `application.json` 中添加 RabbitMQ 配置：

```json
{
  "rabbitmq": {
    "enabled": true,
    "host": "localhost",
    "port": 5672,
    "username": "guest",
    "password": "guest",
    "vhost": "/",
    "queue_name": "enhance-queue",
    "exchange": "",
    "routing_key": "enhance-key",
    "durable": true,
    "auto_delete": false
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
    "github.com/xudefa/enhance/starter/rabbitmq"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("rabbitmq-demo"),
    )
    defer app.Stop()
    
    // 获取 RabbitMQ 队列
    queue := core.MustGetBean[*rabbitmq.RabbitMQQueue](app.Container())
    ctx := context.Background()
    
    // 声明队列
    q, err := queue.DeclareQueue()
    if err != nil {
        // 处理错误
    }
    println("队列已声明:", q.Name)
    
    // 发送消息
    err = queue.Publish(ctx, []byte("Hello, RabbitMQ!"))
    if err != nil {
        // 处理错误
    }
    
    // 订阅消息
    go func() {
        err := queue.Subscribe(ctx, func(msg []byte) error {
            println("收到消息:", string(msg))
            return nil
        })
        if err != nil {
            println("订阅错误:", err.Error())
        }
    }()
    
    // 等待消息处理
    app.WaitForSignal()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `rabbitmq.enabled` | bool | false | 是否启用 RabbitMQ |
| `rabbitmq.host` | string | localhost | RabbitMQ 服务器地址 |
| `rabbitmq.port` | int | 5672 | RabbitMQ 端口 |
| `rabbitmq.username` | string | guest | 用户名 |
| `rabbitmq.password` | string | guest | 密码 |
| `rabbitmq.vhost` | string | / | 虚拟主机 |
| `rabbitmq.queue_name` | string | enhance-queue | 队列名称 |
| `rabbitmq.exchange` | string | "" | 交换机名称 |
| `rabbitmq.routing_key` | string | enhance-key | 路由键 |
| `rabbitmq.durable` | bool | true | 是否持久化 |
| `rabbitmq.auto_delete` | bool | false | 是否自动删除 |

## 高级用法

### 使用交换机

```go
// 配置使用交换机的队列
queue := core.MustGetBean[*rabbitmq.RabbitMQQueue](app.Container())

// 声明交换机
err := queue.channel.ExchangeDeclare(
    "my-exchange", // 名称
    "direct",      // 类型
    true,          // 持久化
    false,         // 自动删除
    false,         // 内部
    false,         // 不等待
    nil,           // 参数
)
```

### 消息确认

```go
// 手动确认消息
msgs, _ := queue.channel.Consume(
    queue.config.QueueName,
    "",
    false, // 不自动确认
    false,
    false,
    false,
    nil,
)

for msg := range msgs {
    // 处理消息
    msg.Ack(false) // 确认消息
}
```

### 优雅关闭

```go
rabbitmqConfig := core.MustGetBean[*rabbitmq.RabbitMQAutoConfiguration](app.Container())
rabbitmqConfig.Stop()
```

## 启动顺序

- **优先级**: `OrderPriorityBusinessLayer` (1000)
- **触发条件**: `rabbitmq.enabled=true`

## 依赖

- `github.com/rabbitmq/amqp091-go`