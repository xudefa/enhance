// Package mq 提供消息队列支持，用于 enhance 框架。
package mq

import (
	"container/list"
	"fmt"
	"log/slog"
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
	wg       sync.WaitGroup
}

var messageIDCounter atomic.Int64

// generateMessageID 生成消息 ID。
func generateMessageID() string {
	id := messageIDCounter.Add(1)
	return fmt.Sprintf("msg-%d", id)
}

// AcquireMessage 从池中获取 Message 对象。
func AcquireMessage() *Message {
	msg, ok := messagePool.Get().(*Message)
	if !ok {
		msg = &Message{
			Headers: make(map[string]string, 8),
		}
	}
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
	if !m.acknowledged.CompareAndSwap(0, 2) {
		return // 已确认或已拒绝，无需重复处理
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
		q.mu.Lock()
		defer q.mu.Unlock()

		if requeue && msg.RetryCount < msg.MaxRetries {
			msg.RetryCount++
			msg.acknowledged.Store(0) // 重置确认状态以便重新入队
			q.messages.PushBack(msg)
			q.cond.Signal()
		} else if q.deadLetterQueue != nil {
			msg.acknowledged.Store(0) // 重置确认状态，使 DLQ 消费者可以 Ack/Nack
			if err := q.deadLetterQueue.Send(msg); err != nil {
				slog.Error("failed to send to dead letter queue", "error", err)
			}
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
		if q.stopped.Load() {
			return nil, ErrStopped
		}
		q.cond.Wait()
	}

	elem := q.messages.Front()
	if elem == nil {
		return nil, fmt.Errorf("no messages available")
	}

	q.messages.Remove(elem)
	msg, _ := elem.Value.(*Message)

	return msg, nil
}

// ReceiveWithTimeout 实现 Queue 接口
func (q *InMemoryQueue) ReceiveWithTimeout(timeout time.Duration) (*Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	deadline := time.Now().Add(timeout)

	for q.messages.Len() == 0 {
		if q.stopped.Load() {
			return nil, ErrStopped
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("receive message timeout")
		}

		// 使用 select + done channel 防止定时器唤醒丢失：
		// 定时器回调先获取互斥锁再 Broadcast，确保唤醒发生在 Wait() 挂起之后。
		done := make(chan struct{})
		timer := time.AfterFunc(remaining, func() {
			select {
			case <-done:
				return
			default:
			}
			q.mu.Lock()
			q.cond.Broadcast()
			q.mu.Unlock()
		})

		q.cond.Wait()

		// 确保清理定时器，避免泄漏
		timer.Stop()
		close(done)
	}

	elem := q.messages.Front()
	if elem == nil {
		return nil, fmt.Errorf("no messages available")
	}

	q.messages.Remove(elem)
	msg, _ := elem.Value.(*Message)

	return msg, nil
}

// Consume 实现 Queue 接口
func (q *InMemoryQueue) Consume(handler MessageHandler) error {
	if q.consuming.Swap(true) {
		return fmt.Errorf("queue is already being consumed, cannot consume again")
	}

	q.mu.Lock()
	q.stopChan = make(chan struct{})
	q.stopped.Store(false)
	stopChan := q.stopChan
	q.mu.Unlock()

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[MQ] message handler panic recovered in queue %s: %v\n", q.name, r)
			}
		}()

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
	if !q.consuming.Load() {
		q.mu.Unlock()
		return
	}

	q.consuming.Store(false)
	q.stopped.Store(true)
	q.cond.Broadcast()

	// 安全关闭 channel，避免重复关闭 panic
	select {
	case <-q.stopChan:
		// channel 已经关闭
	default:
		close(q.stopChan)
	}
	q.mu.Unlock()

	// 等待消费者 goroutine 退出
	q.wg.Wait()
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
