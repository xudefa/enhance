// Package mq 提供消息队列支持，用于 enhance 框架。
//
// 该模块提供统一的消息队列抽象接口，支持多种消息中间件集成。
// 包含消息发送、消费、消息转换等消息中间件功能。
//
// # 架构设计
//
//   - Queue: 消息队列接口，定义统一的消息操作
//   - Message: 消息对象，包含消息头、消息体等
//   - MessageHandler: 消息处理器函数类型
//   - QueueOption: 队列配置选项函数
//
// # 核心功能
//
//   - 消息发送: 支持同步和异步消息发送
//   - 消息消费: 支持消息监听和自动消费
//   - 消息确认: 支持消息确认（Ack/Nack）和重试机制
//   - 死信队列: 支持死信队列处理失败消息
//   - 对象池: 使用 sync.Pool 复用 Message 对象，减少 GC 压力
//
// # 使用方式
//
// 创建消息队列：
//
//	queue := mq.NewInMemoryQueue("my-queue")
//
// 发送消息：
//
//	msg := mq.AcquireMessage()
//	msg.Body = []byte(`{"userId": 123}`)
//	err := queue.Send(msg)
//
// 消费消息：
//
//	queue.Consume(func(msg *mq.Message) error {
//	    // 处理消息
//	    msg.Ack()
//	    return nil
//	})
//
// # 集成后端
//
// 具体实现位于 starter 子包：
//
//   - starter/kafka: Apache Kafka 集成
//   - starter/rabbitmq: RabbitMQ 集成
package mq

import (
	"errors"
	"sync/atomic"
	"time"
)

// 默认配置常量。
const (
	// DefaultMaxRetries 默认最大重试次数。
	DefaultMaxRetries = 3
	// DefaultReceiveTimeout 默认接收超时时间。
	DefaultReceiveTimeout = 1 * time.Second
)

// ErrStopped 队列已停止的错误。
var ErrStopped = errors.New("queue has been stopped")

// Message 消息对象。
//
// 表示队列中的一条消息，支持并发安全的 Ack/Nack 操作。
// 使用 atomic.Int32 保证 acknowledged 状态的线程安全。
type Message struct {
	// ID 消息 ID。
	ID string
	// Body 消息体。
	Body []byte
	// Headers 消息头。
	Headers map[string]string
	// Timestamp 消息时间戳。
	Timestamp time.Time
	// QueueName 队列名称。
	QueueName string
	// RetryCount 重试次数。
	RetryCount int
	// MaxRetries 最大重试次数。
	MaxRetries int
	// ack 确认函数（内部使用）。
	ack func()
	// nack 拒绝函数（内部使用）。
	nack func(requeue bool)
	// acknowledged 是否已确认（0=未确认, 1=已确认）（内部使用）。
	acknowledged atomic.Int32
}

// MessageHandler 消息处理器函数类型。
type MessageHandler func(msg *Message) error

// QueueOption 队列配置选项函数。
type QueueOption func(*BaseQueue)

// Queue 消息队列接口。
type Queue interface {
	// Send 发送消息。
	Send(msg *Message) error
	// Receive 接收消息。
	Receive() (*Message, error)
	// ReceiveWithTimeout 带超时接收消息。
	ReceiveWithTimeout(timeout time.Duration) (*Message, error)
	// Consume 消费消息（持续监听）。
	Consume(handler MessageHandler) error
	// StopConsuming 停止消费。
	StopConsuming()
	// Purge 清空队列。
	Purge() error
	// Close 关闭队列。
	Close() error
	// Name 获取队列名称。
	Name() string
	// Size 获取队列大小。
	Size() int
}

// BaseQueue 基础队列结构，提供队列的公共字段和状态管理。
type BaseQueue struct {
	name            string
	maxRetries      int
	deadLetterQueue Queue
	consuming       atomic.Bool
	stopped         atomic.Bool
	stopChan        chan struct{}
}
