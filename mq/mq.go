// Package mq 提供消息队列支持，用于 enhance 框架。
package mq

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// messagePool 复用 Message 对象的对象池。
var messagePool = sync.Pool{
	New: func() any {
		return &Message{
			Headers: make(map[string]string, 8),
		}
	},
}

// InMemoryQueue 内存消息队列实现。
type InMemoryQueue struct {
	mu sync.Mutex
	BaseQueue
	messages *list.List
	cond     *sync.Cond
}

var messageIDCounter atomic.Int64

// generateMessageID 生成消息 ID。
func generateMessageID() string {
	id := messageIDCounter.Add(1)
	return fmt.Sprintf("msg-%d", id)
}

// AcquireMessage 从池中获取 Message 对象。
func AcquireMessage() *Message {
	msg := messagePool.Get().(*Message)
	if msg.Headers != nil {
		for k := range msg.Headers {
			delete(msg.Headers, k)
		}
	} else {
		msg.Headers = make(map[string]string, 8)
	}
	msg.acknowledged.Store(0)
	msg.RetryCount = 0
	msg.MaxRetries = DefaultMaxRetries
	msg.Body = nil
	msg.QueueName = ""
	msg.Timestamp = time.Time{}
	msg.ID = ""
	msg.ack = nil
	msg.nack = nil
	return msg
}

// ReleaseMessage 归还 Message 对象到池中。
func ReleaseMessage(msg *Message) {
	if msg == nil {
		return
	}
	if msg.Headers != nil {
		for k := range msg.Headers {
			delete(msg.Headers, k)
		}
	}
	msg.acknowledged.Store(0)
	msg.RetryCount = 0
	msg.Body = nil
	msg.QueueName = ""
	msg.Timestamp = time.Time{}
	msg.ID = ""
	msg.ack = nil
	msg.nack = nil
	messagePool.Put(msg)
}

// Ack 确认消息。
//
// 线程安全，多次调用只会执行一次。
func (m *Message) Ack() {
	if m.acknowledged.CompareAndSwap(0, 1) {
		if m.ack != nil {
			m.ack()
		}
	}
}

// Nack 拒绝消息。
//
// 线程安全，多次调用只会执行一次。
func (m *Message) Nack(requeue bool) {
	if m.acknowledged.Load() == 1 {
		return // 已确认，无需重复处理
	}
	if m.nack != nil {
		m.nack(requeue)
	}
}

// IsAcknowledged 检查是否已确认。
func (m *Message) IsAcknowledged() bool {
	return m.acknowledged.Load() == 1
}

// GetHeader 获取消息头。
func (m *Message) GetHeader(key string) string {
	if m.Headers == nil {
		return ""
	}
	return m.Headers[key]
}

// SetHeader 设置消息头。
func (m *Message) SetHeader(key, value string) {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
}

// WithMaxRetries 设置最大重试次数。
func WithMaxRetries(maxRetries int) QueueOption {
	return func(q *BaseQueue) {
		q.maxRetries = maxRetries
	}
}

// WithDeadLetterQueue 设置死信队列。
func WithDeadLetterQueue(dlq Queue) QueueOption {
	return func(q *BaseQueue) {
		q.deadLetterQueue = dlq
	}
}

// Name 实现 Queue 接口。
func (q *BaseQueue) Name() string {
	return q.name
}

// NewInMemoryQueue 创建内存消息队列。
func NewInMemoryQueue(name string, opts ...QueueOption) *InMemoryQueue {
	queue := &InMemoryQueue{
		BaseQueue: BaseQueue{
			name:       name,
			maxRetries: DefaultMaxRetries,
			stopChan:   make(chan struct{}),
		},
		messages: list.New(),
	}

	queue.cond = sync.NewCond(&queue.mu)

	for _, opt := range opts {
		opt(&queue.BaseQueue)
	}

	return queue
}

// Send 实现 Queue 接口
func (q *InMemoryQueue) Send(msg *Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	msg.QueueName = q.name
	msg.Timestamp = time.Now()
	msg.MaxRetries = q.maxRetries
	msg.ID = generateMessageID()

	// 设置确认函数
	msg.ack = func() {
		// 消息已确认，无需额外操作
	}

	msg.nack = func(requeue bool) {
		if !msg.acknowledged.CompareAndSwap(0, 1) {
			return // 已确认，无需重复处理
		}

		q.mu.Lock()
		defer q.mu.Unlock()

		if requeue && msg.RetryCount < msg.MaxRetries {
			msg.RetryCount++
			msg.acknowledged.Store(0) // 重置确认状态以便重新入队
			q.messages.PushBack(msg)
			q.cond.Signal()
		} else if q.deadLetterQueue != nil {
			_ = q.deadLetterQueue.Send(msg)
		}
	}

	q.messages.PushBack(msg)
	q.cond.Signal()

	return nil
}

// Receive 实现 Queue 接口
func (q *InMemoryQueue) Receive() (*Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.messages.Len() == 0 {
		q.cond.Wait()
	}

	elem := q.messages.Front()
	if elem == nil {
		return nil, fmt.Errorf("no messages available")
	}

	q.messages.Remove(elem)
	msg := elem.Value.(*Message)

	return msg, nil
}

// ReceiveWithTimeout 实现 Queue 接口
func (q *InMemoryQueue) ReceiveWithTimeout(timeout time.Duration) (*Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	deadline := time.Now().Add(timeout)

	for q.messages.Len() == 0 {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("receive message timeout")
		}

		// 使用定时器等待
		timer := time.AfterFunc(remaining, func() {
			q.cond.Broadcast()
		})

		q.cond.Wait()

		// 确保清理定时器，避免泄漏
		timer.Stop()
	}

	elem := q.messages.Front()
	if elem == nil {
		return nil, fmt.Errorf("no messages available")
	}

	q.messages.Remove(elem)
	msg := elem.Value.(*Message)

	return msg, nil
}

// Consume 实现 Queue 接口
func (q *InMemoryQueue) Consume(handler MessageHandler) error {
	if q.consuming.Swap(true) {
		return fmt.Errorf("queue is already being consumed, cannot consume again")
	}

	q.mu.Lock()
	q.stopChan = make(chan struct{})
	stopChan := q.stopChan
	q.mu.Unlock()

	go func() {
		for {
			if !q.consuming.Load() {
				return
			}

			select {
			case <-stopChan:
				return
			default:
				msg, err := q.ReceiveWithTimeout(DefaultReceiveTimeout)
				if err != nil {
					continue
				}

				if err := handler(msg); err != nil {
					msg.Nack(true)
					continue
				}
				msg.Ack()
			}
		}
	}()

	return nil
}

// StopConsuming 实现 Queue 接口
func (q *InMemoryQueue) StopConsuming() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.consuming.Load() {
		return
	}

	q.consuming.Store(false)

	// 安全关闭 channel，避免重复关闭 panic
	select {
	case <-q.stopChan:
		// channel 已经关闭
	default:
		close(q.stopChan)
	}
}

// Purge 实现 Queue 接口
func (q *InMemoryQueue) Purge() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.messages.Init()
	return nil
}

// Close 实现 Queue 接口
func (q *InMemoryQueue) Close() error {
	q.StopConsuming()
	return q.Purge()
}

// Size 实现 Queue 接口
func (q *InMemoryQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.messages.Len()
}

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
	return val.(Queue), nil
}

// DeleteQueue 删除队列
func (f *MessageQueueFactory) DeleteQueue(name string) error {
	val, ok := f.queues.Load(name)
	if !ok {
		return fmt.Errorf("queue %s does not exist", name)
	}
	queue := val.(Queue)
	_ = queue.Close()
	f.queues.Delete(name)
	return nil
}

// ListQueues 列出所有队列
func (f *MessageQueueFactory) ListQueues() []string {
	var names []string
	f.queues.Range(func(key, value any) bool {
		names = append(names, key.(string))
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
