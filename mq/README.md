# mq 包 — 消息队列

> **所属层级**: Infrastructure Layer  
> **设计理念**: 统一抽象，多后端支持  
> **设计灵感**: Spring AMQP + Spring Kafka

## 概述

`mq` 包提供统一的消息队列抽象，参考 Spring AMQP 和 Spring Kafka 设计。支持多种消息队列后端（内存、Redis、Kafka 等），以及消息发送、接收、确认等功能。

### 核心功能

| 功能 | 说明 |
|------|------|
| **统一接口** | Queue 接口抽象，支持不同后端实现 |
| **消息确认** | 支持 Ack/Nack 消息确认机制 |
| **重试机制** | 支持消息重试和死信队列 |
| **消息工厂** | 提供消息队列工厂简化创建 |
| **消息模板** | 提供便捷的消息发送和接收方法 |

---

## 核心接口

### Message 消息对象

```go
type Message struct {
    ID         string
    Body       []byte
    Headers    map[string]string
    Timestamp  time.Time
    QueueName  string
    RetryCount int
    MaxRetries int
}
```

#### 消息方法

| 方法 | 说明 |
|------|------|
| `Ack()` | 确认消息 |
| `Nack(requeue bool)` | 拒绝消息，可选择重新入队 |
| `IsAcknowledged()` | 检查是否已确认 |
| `GetHeader(key string)` | 获取消息头 |
| `SetHeader(key, value string)` | 设置消息头 |

### Queue 消息队列接口

```go
type Queue interface {
    Send(msg Message) error
    Receive() (*Message, error)
    ReceiveWithTimeout(timeout time.Duration) (*Message, error)
    Consume(handler MessageHandler) error
    StopConsuming()
    Purge() error
    Close() error
    Name() string
    Size() int
}
```

#### 方法说明

| 方法 | 说明 |
|------|------|
| `Send` | 发送消息 |
| `Receive` | 接收消息（阻塞） |
| `ReceiveWithTimeout` | 带超时接收消息 |
| `Consume` | 消费消息（持续监听） |
| `StopConsuming` | 停止消费 |
| `Purge` | 清空队列 |
| `Close` | 关闭队列 |
| `Name` | 获取队列名称 |
| `Size` | 获取队列大小 |

### InMemoryQueue 内存消息队列

```go
type InMemoryQueue struct {
    // ...
}
```

#### 创建

```go
queue := mq.NewInMemoryQueue("my-queue")
```

#### 队列选项

| 函数 | 说明 | 默认值 |
|------|------|--------|
| `WithMaxRetries(n)` | 设置最大重试次数 | `3` |
| `WithDeadLetterQueue(dlq)` | 设置死信队列 | 无 |

#### 使用示例

```go
// 带死信队列
dlq := mq.NewInMemoryQueue("my-queue.dlq")
queue := mq.NewInMemoryQueue(
    "my-queue",
    mq.WithMaxRetries(5),
    mq.WithDeadLetterQueue(dlq),
)
```

### MessageQueueFactory 消息队列工厂

```go
type MessageQueueFactory struct {
    // ...
}
```

#### 创建和管理

```go
factory := mq.NewMessageQueueFactory()

// 创建内存消息队列
queue := factory.CreateInMemoryQueue("my-queue")

// 获取队列
queue, err := factory.GetQueue("my-queue")

// 删除队列
err := factory.DeleteQueue("my-queue")

// 列出所有队列
names := factory.ListQueues()
```

### MessagePublisher 消息发布者

```go
type MessagePublisher struct {
    // ...
}
```

#### 创建和使用

```go
publisher := mq.NewMessagePublisher(queue)

// 发布消息
err := publisher.Publish([]byte("hello world"), map[string]string{
    "content-type": "text/plain",
})

// 发布 JSON 消息
err := publisher.PublishJSON([]byte(`{"name": "John", "age": 30}`))
```

### MessageConsumer 消息消费者

```go
type MessageConsumer struct {
    // ...
}
```

#### 创建和控制

```go
consumer := mq.NewMessageConsumer(queue, func(msg *mq.Message) error {
    fmt.Printf("Received: %s\n", string(msg.Body))
    return nil
})

// 开始消费
if err := consumer.Start(); err != nil {
    // 处理错误
}

// 停止消费
consumer.Stop()
```

### MessageTemplate 消息模板

```go
type MessageTemplate struct {
    // ...
}
```

#### 创建和使用

```go
template := mq.NewMessageTemplate(queue)

// 发送消息
err := template.Send([]byte("hello"))

// 发送带消息头的消息
err := template.SendWithHeaders([]byte("hello"), map[string]string{
    "correlation-id": "12345",
})

// 接收消息
msg, err := template.Receive()

// 带超时接收消息
msg, err := template.ReceiveWithTimeout(5 * time.Second)
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/mq"
)

func main() {
    // 创建队列
    queue := mq.NewInMemoryQueue("my-queue")

    // 发送消息
    queue.Send(mq.Message{
        Body: []byte("hello world"),
        Headers: map[string]string{
            "content-type": "text/plain",
        },
    })

    // 接收消息
    msg, err := queue.Receive()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Received:", string(msg.Body))
    msg.Ack()
}
```

---

## API 参考

### 使用工厂管理队列

```go
factory := mq.NewMessageQueueFactory()

// 创建多个队列
orderQueue := factory.CreateInMemoryQueue("orders")
emailQueue := factory.CreateInMemoryQueue("emails")
notificationQueue := factory.CreateInMemoryQueue("notifications")

// 获取队列
queue, err := factory.GetQueue("orders")
if err != nil {
    // 处理错误
}

// 列出所有队列
names := factory.ListQueues()
fmt.Println("Queues:", names)
```

### 使用发布者和消费者

```go
queue := mq.NewInMemoryQueue("events")

// 创建发布者
publisher := mq.NewMessagePublisher(queue)

// 发布消息
publisher.Publish([]byte("user.created"), map[string]string{
    "event-type": "user.created",
    "user-id":    "12345",
})

// 创建消费者
consumer := mq.NewMessageConsumer(queue, func(msg *mq.Message) error {
    eventType := msg.GetHeader("event-type")
    fmt.Printf("Received event: %s\n", eventType)
    return nil
})

// 开始消费
consumer.Start()
defer consumer.Stop()
```

### 使用消息模板

```go
queue := mq.NewInMemoryQueue("tasks")
template := mq.NewMessageTemplate(queue)

// 发送任务
template.Send([]byte(`{"task": "send_email", "to": "user@example.com"}`))

// 接收任务
msg, err := template.ReceiveWithTimeout(5 * time.Second)
if err != nil {
    fmt.Println("Timeout waiting for task")
    return
}

// 处理任务
processTask(msg.Body)
msg.Ack()
```

### 死信队列

```go
// 创建死信队列
dlq := mq.NewInMemoryQueue("orders.dlq")

// 创建主队列，配置重试和死信队列
queue := mq.NewInMemoryQueue(
    "orders",
    mq.WithMaxRetries(3),
    mq.WithDeadLetterQueue(dlq),
)

// 发送消息
queue.Send(mq.Message{
    Body: []byte("order data"),
})

// 消费消息，处理失败会自动重试
consumer := mq.NewMessageConsumer(queue, func(msg *mq.Message) error {
    err := processOrder(msg.Body)
    if err != nil {
        return err // 消息会 Nack 并重新入队
    }
    return nil
})

consumer.Start()
```

---

## 使用示例

### 场景 1: 异步订单处理

用户下单后异步处理订单，提高接口响应速度：

```go
func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*Order, error) {
    // 创建订单记录
    order := s.createOrder(req)

    // 发送订单处理消息
    s.template.SendWithHeaders(
        order.ID,
        map[string]string{
            "order-id": order.ID,
            "action":   "process",
        },
    )

    // 立即返回
    return order, nil
}

func (s *OrderService) StartOrderProcessor() {
    consumer := mq.NewMessageConsumer(s.orderQueue, func(msg *mq.Message) error {
        orderID := msg.GetHeader("order-id")
        return s.processOrder(orderID)
    })
    consumer.Start()
}
```

**最佳实践**:
- 核心流程同步，非核心流程异步
- 使用消息头传递元数据
- 确保消息处理幂等性

### 场景 2: 事件驱动架构

使用消息队列实现模块间解耦，构建事件驱动架构：

```go
// 用户服务
func (s *UserService) CreateUser(req *CreateUserRequest) error {
    user := s.createUser(req)

    // 发布用户创建事件
    s.publisher.PublishJSON([]byte(fmt.Sprintf(
        `{"event": "user.created", "user_id": "%s"}`,
        user.ID,
    )))

    return nil
}

// 邮件服务（监听用户创建事件）
func (s *EmailService) StartListener() {
    consumer := mq.NewMessageConsumer(s.eventQueue, func(msg *mq.Message) error {
        var event map[string]string
        json.Unmarshal(msg.Body, &event)

        if event["event"] == "user.created" {
            return s.sendWelcomeEmail(event["user_id"])
        }
        return nil
    })
    consumer.Start()
}

// 分析服务（监听用户创建事件）
func (s *AnalyticsService) StartListener() {
    consumer := mq.NewMessageConsumer(s.eventQueue, func(msg *mq.Message) error {
        var event map[string]string
        json.Unmarshal(msg.Body, &event)

        if event["event"] == "user.created" {
            return s.trackUserCreation(event["user_id"])
        }
        return nil
    })
    consumer.Start()
}
```

**最佳实践**:
- 使用统一的事件格式
- 每个服务独立消费队列
- 事件处理失败不影响其他服务

### 场景 3: 任务队列

使用消息队列实现任务队列，支持后台异步任务处理：

```go
type TaskWorker struct {
    template *mq.MessageTemplate
}

func (w *TaskWorker) SubmitTask(task Task) error {
    data, _ := json.Marshal(task)
    return w.template.SendWithHeaders(data, map[string]string{
        "task-type": task.Type,
        "task-id":   task.ID,
    })
}

func (w *TaskWorker) StartWorker() {
    for {
        msg, err := w.template.ReceiveWithTimeout(1 * time.Second)
        if err != nil {
            continue
        }

        taskType := msg.GetHeader("task-type")
        switch taskType {
        case "email":
            w.processEmailTask(msg.Body)
        case "report":
            w.processReportTask(msg.Body)
        case "cleanup":
            w.processCleanupTask(msg.Body)
        }

        msg.Ack()
    }
}
```

**最佳实践**:
- 使用消息头区分任务类型
- 设置合理的超时时间
- 任务处理失败时记录日志并重试

### 场景 4: 日志聚合

使用消息队列聚合日志，统一处理和分析：

```go
func (s *LogService) CollectLog(log LogEntry) error {
    data, _ := json.Marshal(log)
    return s.template.SendWithHeaders(data, map[string]string{
        "level":     log.Level,
        "service":   log.Service,
        "timestamp": log.Timestamp,
    })
}

func (s *LogService) StartLogProcessor() {
    consumer := mq.NewMessageConsumer(s.logQueue, func(msg *mq.Message) error {
        var log LogEntry
        json.Unmarshal(msg.Body, &log)

        // 存储到数据库
        if err := s.storeLog(log); err != nil {
            return err
        }

        // 告警处理
        if log.Level == "ERROR" || log.Level == "CRITICAL" {
            s.sendAlert(log)
        }

        return nil
    })
    consumer.Start()
}
```

**最佳实践**:
- 日志结构化存储
- 错误日志触发告警
- 定期清理旧日志

---

## 最佳实践

### 1. 消息处理的幂等性

```go
// ✅ 推荐：确保消息处理幂等性
func processOrder(msg *mq.Message) error {
    orderID := msg.GetHeader("order-id")
    
    // 检查是否已处理过此订单
    if s.isOrderProcessed(orderID) {
        return nil // 已处理，直接返回
    }
    
    // 处理订单
    return s.handleOrder(orderID)
}

// ⚠️ 不推荐：不考虑幂等性
func processOrder(msg *mq.Message) error {
    orderID := msg.GetHeader("order-id")
    return s.handleOrder(orderID) // 可能重复处理
}
```

### 2. 使用死信队列处理失败消息

```go
// ✅ 推荐：配置死信队列
dlq := mq.NewInMemoryQueue("orders.dlq")
queue := mq.NewInMemoryQueue(
    "orders",
    mq.WithMaxRetries(3),
    mq.WithDeadLetterQueue(dlq),
)

// ⚠️ 不推荐：没有死信队列，失败消息会丢失
queue := mq.NewInMemoryQueue("orders")
```

### 3. 消息头使用规范

```go
// ✅ 推荐：使用标准消息头
publisher.PublishWithHeaders([]byte("message body"), map[string]string{
    "correlation-id": generateUUID(), // 用于追踪消息
    "timestamp":      time.Now().Format(time.RFC3339),
    "source":         "order-service",
    "event-type":     "order.created",
})

// ⚠️ 不推荐：没有标准化的消息头
publisher.Publish([]byte("message body"), map[string]string{
    "id": "12345", // 不够具体
})
```

### 4. 合理设置超时时间

```go
// ✅ 推荐：根据业务需求设置合理超时
msg, err := template.ReceiveWithTimeout(30 * time.Second)

// ⚠️ 不推荐：超时时间过长或过短
msg, err := template.ReceiveWithTimeout(1 * time.Millisecond) // 太短
msg, err := template.ReceiveWithTimeout(1 * time.Hour)       // 太长
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将消息队列相关组件注册为 Bean
container.Register(
    reflect.TypeOf(&mq.MessageTemplate{}),
    core.Bean(createMessageTemplate()),
    core.Singleton(),
)

// 注入使用
type OrderService struct {
    Template *mq.MessageTemplate `inject:"messageTemplate"`
}

func (s *OrderService) ProcessOrder(orderID string) error {
    return s.Template.Send([]byte(orderID))
}
```

### 6. 设计要点

- `InMemoryQueue` 基于 `container/list` 和 `sync.Cond` 实现
- 消息确认使用 Ack/Nack 模式，支持重新入队
- 重试次数超过最大值后进入死信队列
- 消费者在独立 goroutine 中运行
- 零外部依赖，仅使用 Go 标准库