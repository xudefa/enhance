package mq

import (
	"fmt"
	"sync"
	"time"
)

// MessageQueueFactory 消息队列工厂
//
// 使用 sync.Map 优化并发访问，避免全局锁竞争。
type MessageQueueFactory struct {
	queues sync.Map // map[string]Queue
}

// NewMessageQueueFactory 创建消息队列工厂
func NewMessageQueueFactory() *MessageQueueFactory {
	return &MessageQueueFactory{}
}

// CreateInMemoryQueue 创建内存消息队列
func (f *MessageQueueFactory) CreateInMemoryQueue(name string, opts ...QueueOption) Queue {
	queue := NewInMemoryQueue(name, opts...)
	f.queues.Store(name, queue)
	return queue
}

// GetQueue 获取队列
func (f *MessageQueueFactory) GetQueue(name string) (Queue, error) {
	val, ok := f.queues.Load(name)
	if !ok {
		return nil, fmt.Errorf("queue %s does not exist", name)
	}
	q, _ := val.(Queue)
	return q, nil
}

// DeleteQueue 删除队列
func (f *MessageQueueFactory) DeleteQueue(name string) error {
	val, ok := f.queues.Load(name)
	if !ok {
		return fmt.Errorf("queue %s does not exist", name)
	}
	queue, ok := val.(Queue)
	if !ok {
		return fmt.Errorf("queue %s has invalid type", name)
	}
	if err := queue.Close(); err != nil {
		return fmt.Errorf("failed to close queue %s: %w", name, err)
	}
	f.queues.Delete(name)
	return nil
}

// ListQueues 列出所有队列
func (f *MessageQueueFactory) ListQueues() []string {
	var names []string
	f.queues.Range(func(key, value any) bool {
		k, _ := key.(string)
		names = append(names, k)
		return true
	})
	return names
}

// MessagePublisher 消息发布者
type MessagePublisher struct {
	queue Queue
}

// NewMessagePublisher 创建消息发布者
func NewMessagePublisher(queue Queue) *MessagePublisher {
	return &MessagePublisher{
		queue: queue,
	}
}

// Publish 发布消息
func (p *MessagePublisher) Publish(body []byte, headers map[string]string) error {
	msg := &Message{
		Body:    body,
		Headers: headers,
	}

	return p.queue.Send(msg)
}

// PublishJSON 发布 JSON 消息
func (p *MessagePublisher) PublishJSON(body []byte) error {
	msg := &Message{
		Body: body,
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}

	return p.queue.Send(msg)
}

// MessageConsumer 消息消费者
type MessageConsumer struct {
	queue   Queue
	handler MessageHandler
}

// NewMessageConsumer 创建消息消费者
func NewMessageConsumer(queue Queue, handler MessageHandler) *MessageConsumer {
	return &MessageConsumer{
		queue:   queue,
		handler: handler,
	}
}

// Start 开始消费
func (c *MessageConsumer) Start() error {
	return c.queue.Consume(c.handler)
}

// Stop 停止消费
func (c *MessageConsumer) Stop() {
	c.queue.StopConsuming()
}

// MessageTemplate 消息模板
//
// 提供便捷的消息发送和接收方法
type MessageTemplate struct {
	queue Queue
}

// NewMessageTemplate 创建消息模板
func NewMessageTemplate(queue Queue) *MessageTemplate {
	return &MessageTemplate{
		queue: queue,
	}
}

// Send 发送消息
func (t *MessageTemplate) Send(body []byte) error {
	return t.queue.Send(&Message{
		Body:      body,
		Timestamp: time.Now(),
	})
}

// SendWithHeaders 发送带消息头的消息
func (t *MessageTemplate) SendWithHeaders(body []byte, headers map[string]string) error {
	return t.queue.Send(&Message{
		Body:      body,
		Headers:   headers,
		Timestamp: time.Now(),
	})
}

// Receive 接收消息
func (t *MessageTemplate) Receive() (*Message, error) {
	return t.queue.Receive()
}

// ReceiveWithTimeout 带超时接收消息
func (t *MessageTemplate) ReceiveWithTimeout(timeout time.Duration) (*Message, error) {
	return t.queue.ReceiveWithTimeout(timeout)
}

// Purge 清空队列
func (t *MessageTemplate) Purge() error {
	return t.queue.Purge()
}

// Size 获取队列大小
func (t *MessageTemplate) Size() int {
	return t.queue.Size()
}
