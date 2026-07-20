# Kafka Starter

Kafka 消息队列自动配置模块，提供分布式消息队列支持。

## 功能特性

- ✅ 自动配置 Kafka 连接
- ✅ 消息发布和订阅
- ✅ 消费者组支持
- ✅ 优雅退出机制
- ✅ 连接健康检查

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/kafka"
)
```

### 2. 配置文件

在 `application.json` 中添加 Kafka 配置：

```json
{
  "kafka": {
    "enabled": true,
    "brokers": ["localhost:9092"],
    "topic": "enhance-events",
    "group_id": "enhance-consumer"
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
    "github.com/xudefa/enhance/starter/kafka"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("kafka-demo"),
    )
    defer app.Stop()
    
    // 获取 Kafka 队列
    queue := core.MustGetBean[*kafka.KafkaQueue](app.Container())
    ctx := context.Background()
    
    // 发送消息
    err := queue.Publish(ctx, []byte("Hello, Kafka!"))
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
| `kafka.enabled` | bool | false | 是否启用 Kafka |
| `kafka.brokers` | []string | [] | Kafka Broker 地址列表 |
| `kafka.topic` | string | enhance-events | 主题名称 |
| `kafka.group_id` | string | enhance-consumer | 消费者组 ID |

## 高级用法

### 创建多个主题

```go
// 创建不同主题的队列
queue1 := kafka.NewKafkaQueue(brokers, "topic-1", "group-1")
queue2 := kafka.NewKafkaQueue(brokers, "topic-2", "group-2")
```

### 优雅关闭

```go
ctx, cancel := context.WithCancel(context.Background())
// 取消 context 以停止订阅
cancel()
queue.Close()
```

## 启动顺序

- **优先级**: `OrderPriorityBusinessLayer` (1000)
- **触发条件**: `kafka.enabled=true`

## 依赖

- `github.com/segmentio/kafka-go`