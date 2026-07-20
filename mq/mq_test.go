package mq

import (
	"sync"
	"testing"
	"time"
)

func TestInMemoryQueue_SendReceive(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	err := queue.Send(&Message{
		Body: []byte("hello world"),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("expected queue size 1, got %d", queue.Size())
	}

	msg, err := queue.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if string(msg.Body) != "hello world" {
		t.Errorf("expected body 'hello world', got %s", string(msg.Body))
	}

	if msg.QueueName != "test-queue" {
		t.Errorf("expected queue name 'test-queue', got %s", msg.QueueName)
	}
}

func TestInMemoryQueue_Ack(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	err := queue.Send(&Message{
		Body: []byte("test message"),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	msg, err := queue.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg.IsAcknowledged() {
		t.Error("expected message to not be acknowledged")
	}

	msg.Ack()

	if !msg.IsAcknowledged() {
		t.Error("expected message to be acknowledged")
	}
}

func TestInMemoryQueue_Nack(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	err := queue.Send(&Message{
		Body: []byte("test message"),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	msg, err := queue.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	// Nack 是同步操作，消息立即重新入队
	msg.Nack(true)

	// 消息应该被重新入队
	if queue.Size() != 1 {
		t.Errorf("expected queue size 1 after nack, got %d", queue.Size())
	}
}

func TestInMemoryQueue_NackWithMaxRetries(t *testing.T) {
	t.Parallel()
	dlq := NewInMemoryQueue("dead-letter-queue")
	queue := NewInMemoryQueue("test-queue",
		WithMaxRetries(2),
		WithDeadLetterQueue(dlq),
	)

	err := queue.Send(&Message{
		Body: []byte("test message"),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 第一次接收并拒绝
	msg, err := queue.ReceiveWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}
	msg.Nack(true)

	// 消息已重新入队，立即检查
	if queue.Size() != 1 {
		t.Errorf("expected queue size 1 after first nack, got %d", queue.Size())
	}

	// 第二次接收并拒绝
	msg, err = queue.ReceiveWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}
	msg.Nack(true)

	// 消息已重新入队，立即检查
	if queue.Size() != 1 {
		t.Errorf("expected queue size 1 after second nack, got %d", queue.Size())
	}

	// 第三次接收并拒绝（超过最大重试次数，应该进入死信队列）
	msg, err = queue.ReceiveWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}
	msg.Nack(true)

	// 消息已进入死信队列，立即检查
	if queue.Size() != 0 {
		t.Errorf("expected queue size 0, got %d", queue.Size())
	}

	if dlq.Size() != 1 {
		t.Errorf("expected dead letter queue size 1, got %d", dlq.Size())
	}
}

func TestInMemoryQueue_ReceiveWithTimeout(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	// 超时接收应该返回错误
	_, err := queue.ReceiveWithTimeout(100 * time.Millisecond)
	if err == nil {
		t.Error("expected timeout error")
	}

	// 发送消息
	err = queue.Send(&Message{
		Body: []byte("test message"),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 接收消息
	msg, err := queue.ReceiveWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if string(msg.Body) != "test message" {
		t.Errorf("expected body 'test message', got %s", string(msg.Body))
	}
}

func TestInMemoryQueue_Consume(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	var received []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	err := queue.Consume(func(msg *Message) error {
		mu.Lock()
		received = append(received, string(msg.Body))
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to start consuming: %v", err)
	}
	defer queue.StopConsuming()

	// 发送多条消息
	wg.Add(3)
	for range 3 {
		err = queue.Send(&Message{
			Body: []byte("message"),
		})
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}
	}

	// 等待所有消息被消费
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for messages to be consumed")
	}

	mu.Lock()
	if len(received) != 3 {
		t.Errorf("expected 3 messages, got %d", len(received))
	}
	mu.Unlock()
}

func TestInMemoryQueue_Purge(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	// 发送消息
	for range 5 {
		err := queue.Send(&Message{
			Body: []byte("message"),
		})
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}
	}

	if queue.Size() != 5 {
		t.Errorf("expected queue size 5, got %d", queue.Size())
	}

	// 清空队列
	err := queue.Purge()
	if err != nil {
		t.Fatalf("Failed to purge queue: %v", err)
	}

	if queue.Size() != 0 {
		t.Errorf("expected queue size 0 after purge, got %d", queue.Size())
	}
}

func TestInMemoryQueue_Close(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	err := queue.Send(&Message{
		Body: []byte("test message"),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	err = queue.Close()
	if err != nil {
		t.Fatalf("Failed to close queue: %v", err)
	}

	if queue.Size() != 0 {
		t.Errorf("expected queue size 0 after close, got %d", queue.Size())
	}
}

func TestMessageQueueFactory(t *testing.T) {
	t.Parallel()
	factory := NewMessageQueueFactory()

	// 创建队列
	queue1 := factory.CreateInMemoryQueue("queue1")
	queue2 := factory.CreateInMemoryQueue("queue2")

	if queue1.Name() != "queue1" {
		t.Errorf("expected queue name 'queue1', got %s", queue1.Name())
	}

	if queue2.Name() != "queue2" {
		t.Errorf("expected queue name 'queue2', got %s", queue2.Name())
	}

	// 获取队列
	q, err := factory.GetQueue("queue1")
	if err != nil {
		t.Fatalf("Failed to get queue: %v", err)
	}

	if q.Name() != "queue1" {
		t.Errorf("expected queue name 'queue1', got %s", q.Name())
	}

	// 获取不存在的队列
	_, err = factory.GetQueue("nonexistent")
	if err == nil {
		t.Error("expected error when getting nonexistent queue")
	}

	// 列出所有队列
	queues := factory.ListQueues()
	if len(queues) != 2 {
		t.Errorf("expected 2 queues, got %d", len(queues))
	}

	// 删除队列
	err = factory.DeleteQueue("queue1")
	if err != nil {
		t.Fatalf("Failed to delete queue: %v", err)
	}

	queues = factory.ListQueues()
	if len(queues) != 1 {
		t.Errorf("expected 1 queue after deletion, got %d", len(queues))
	}
}

func TestMessagePublisher(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")
	publisher := NewMessagePublisher(queue)

	// 发布消息
	err := publisher.Publish([]byte("hello"), map[string]string{
		"key": "value",
	})
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("expected queue size 1, got %d", queue.Size())
	}

	msg, err := queue.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if string(msg.Body) != "hello" {
		t.Errorf("expected body 'hello', got %s", string(msg.Body))
	}

	if msg.GetHeader("key") != "value" {
		t.Errorf("expected header 'key' to be 'value', got %s", msg.GetHeader("key"))
	}
}

func TestMessagePublisher_PublishJSON(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")
	publisher := NewMessagePublisher(queue)

	err := publisher.PublishJSON([]byte(`{"key": "value"}`))
	if err != nil {
		t.Fatalf("Failed to publish JSON message: %v", err)
	}

	msg, err := queue.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg.GetHeader("content-type") != "application/json" {
		t.Errorf("expected content-type header to be 'application/json', got %s", msg.GetHeader("content-type"))
	}
}

func TestMessageConsumer(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	var received []byte
	var wg sync.WaitGroup

	consumer := NewMessageConsumer(queue, func(msg *Message) error {
		received = msg.Body
		wg.Done()
		return nil
	})

	err := consumer.Start()
	if err != nil {
		t.Fatalf("Failed to start consumer: %v", err)
	}
	defer consumer.Stop()

	wg.Add(1)
	err = queue.Send(&Message{
		Body: []byte("test message"),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	if string(received) != "test message" {
		t.Errorf("expected received message 'test message', got %s", string(received))
	}
}

func TestMessageTemplate(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")
	template := NewMessageTemplate(queue)

	// 发送消息
	err := template.Send([]byte("hello"))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if template.Size() != 1 {
		t.Errorf("expected queue size 1, got %d", template.Size())
	}

	// 接收消息
	msg, err := template.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if string(msg.Body) != "hello" {
		t.Errorf("expected body 'hello', got %s", string(msg.Body))
	}
}

func TestMessageTemplate_SendWithHeaders(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")
	template := NewMessageTemplate(queue)

	err := template.SendWithHeaders([]byte("hello"), map[string]string{
		"key": "value",
	})
	if err != nil {
		t.Fatalf("Failed to send message with headers: %v", err)
	}

	msg, err := template.Receive()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msg.GetHeader("key") != "value" {
		t.Errorf("expected header 'key' to be 'value', got %s", msg.GetHeader("key"))
	}
}

func TestMessageTemplate_Purge(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")
	template := NewMessageTemplate(queue)

	// 发送消息
	for range 5 {
		err := template.Send([]byte("message"))
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}
	}

	if template.Size() != 5 {
		t.Errorf("expected queue size 5, got %d", template.Size())
	}

	// 清空队列
	err := template.Purge()
	if err != nil {
		t.Fatalf("Failed to purge queue: %v", err)
	}

	if template.Size() != 0 {
		t.Errorf("expected queue size 0 after purge, got %d", template.Size())
	}
}

func TestMessage_Headers(t *testing.T) {
	t.Parallel()
	msg := Message{}

	msg.SetHeader("key1", "value1")
	msg.SetHeader("key2", "value2")

	if msg.GetHeader("key1") != "value1" {
		t.Errorf("expected header 'key1' to be 'value1', got %s", msg.GetHeader("key1"))
	}

	if msg.GetHeader("key2") != "value2" {
		t.Errorf("expected header 'key2' to be 'value2', got %s", msg.GetHeader("key2"))
	}

	if msg.GetHeader("nonexistent") != "" {
		t.Errorf("expected empty header for nonexistent key")
	}
}

func TestInMemoryQueue_DoubleConsume(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	err := queue.Consume(func(msg *Message) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to start consuming: %v", err)
	}
	defer queue.StopConsuming()

	// 再次消费应该返回错误
	err = queue.Consume(func(msg *Message) error {
		return nil
	})
	if err == nil {
		t.Error("expected error when consuming from already consumed queue")
	}
}

func TestMessage_IDGeneration(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-queue")

	ids := make(map[string]bool)
	for range 10 {
		err := queue.Send(&Message{
			Body: []byte("message"),
		})
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		msg, err := queue.Receive()
		if err != nil {
			t.Fatalf("Failed to receive message: %v", err)
		}

		if ids[msg.ID] {
			t.Errorf("duplicate message ID: %s", msg.ID)
		}
		ids[msg.ID] = true
	}
}
