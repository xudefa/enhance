package mq

import (
	"sync"
	"testing"
	"time"
)

func TestMessageQueueFactory_CreateInMemoryQueue(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()
	queue := factory.CreateInMemoryQueue("test-queue")

	if queue == nil {
		t.Fatal("CreateInMemoryQueue() returned nil")
	}
	if queue.Name() != "test-queue" {
		t.Errorf("expected queue name 'test-queue', got %v", queue.Name())
	}
}

func TestMessageQueueFactory_GetQueue(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()
	factory.CreateInMemoryQueue("test-queue")

	queue, err := factory.GetQueue("test-queue")
	if err != nil {
		t.Fatalf("GetQueue() error = %v", err)
	}
	if queue.Name() != "test-queue" {
		t.Errorf("expected queue name 'test-queue', got %v", queue.Name())
	}
}

func TestMessageQueueFactory_GetQueue_NotFound(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()

	_, err := factory.GetQueue("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent queue")
	}
}

func TestMessageQueueFactory_DeleteQueue(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()
	factory.CreateInMemoryQueue("test-queue")

	err := factory.DeleteQueue("test-queue")
	if err != nil {
		t.Fatalf("DeleteQueue() error = %v", err)
	}

	_, err = factory.GetQueue("test-queue")
	if err == nil {
		t.Fatal("queue should be deleted")
	}
}

func TestMessageQueueFactory_DeleteQueue_NotFound(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()

	err := factory.DeleteQueue("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent queue")
	}
}

func TestMessageQueueFactory_ListQueues(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()
	factory.CreateInMemoryQueue("queue-1")
	factory.CreateInMemoryQueue("queue-2")
	factory.CreateInMemoryQueue("queue-3")

	names := factory.ListQueues()
	if len(names) != 3 {
		t.Errorf("expected 3 queues, got %d", len(names))
	}
}

func TestMessageQueueFactory_ListQueues_Empty(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()

	names := factory.ListQueues()
	if len(names) != 0 {
		t.Errorf("expected 0 queues, got %d", len(names))
	}
}

func TestMessageQueueFactory_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	factory := NewMessageQueueFactory()
	var wg sync.WaitGroup

	// 并发创建队列
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			factory.CreateInMemoryQueue("queue-" + string(rune('0'+id)))
		}(i)
	}

	wg.Wait()

	names := factory.ListQueues()
	if len(names) != 10 {
		t.Errorf("expected 10 queues, got %d", len(names))
	}
}

func TestMessagePublisher_Publish(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	publisher := NewMessagePublisher(queue)

	err := publisher.Publish([]byte("hello"), map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("expected queue size 1, got %d", queue.Size())
	}
}

func TestMessagePublisher_PublishJSON(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	publisher := NewMessagePublisher(queue)

	err := publisher.PublishJSON([]byte(`{"key":"value"}`))
	if err != nil {
		t.Fatalf("PublishJSON() error = %v", err)
	}

	msg, err := queue.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}

	if msg.GetHeader("content-type") != "application/json" {
		t.Errorf("expected content-type 'application/json', got %v", msg.GetHeader("content-type"))
	}
}

func TestMessageConsumer_StartStop(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	received := make(chan bool, 1)

	consumer := NewMessageConsumer(queue, func(msg *Message) error {
		received <- true
		return nil
	})

	err := consumer.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// 发送消息
	queue.Send(&Message{Body: []byte("test")})

	// 等待消息被处理
	select {
	case <-received:
		// 成功
	case <-time.After(100 * time.Millisecond):
		t.Fatal("message not received within timeout")
	}

	consumer.Stop()
}

func TestMessageTemplate_Send(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	template := NewMessageTemplate(queue)

	err := template.Send([]byte("hello"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("expected queue size 1, got %d", queue.Size())
	}
}

func TestMessageTemplate_SendWithHeaders(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	template := NewMessageTemplate(queue)

	headers := map[string]string{"key": "value"}
	err := template.SendWithHeaders([]byte("hello"), headers)
	if err != nil {
		t.Fatalf("SendWithHeaders() error = %v", err)
	}

	msg, err := queue.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}

	if msg.GetHeader("key") != "value" {
		t.Errorf("expected header 'value', got %v", msg.GetHeader("key"))
	}
}

func TestMessageTemplate_Receive(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	template := NewMessageTemplate(queue)

	queue.Send(&Message{Body: []byte("test")})

	msg, err := template.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}

	if string(msg.Body) != "test" {
		t.Errorf("expected body 'test', got %v", string(msg.Body))
	}
}

func TestMessageTemplate_ReceiveWithTimeout(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	template := NewMessageTemplate(queue)

	queue.Send(&Message{Body: []byte("test")})

	msg, err := template.ReceiveWithTimeout(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("ReceiveWithTimeout() error = %v", err)
	}

	if string(msg.Body) != "test" {
		t.Errorf("expected body 'test', got %v", string(msg.Body))
	}
}

func TestMessageTemplate_Purge(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	template := NewMessageTemplate(queue)

	queue.Send(&Message{Body: []byte("test")})

	err := template.Purge()
	if err != nil {
		t.Fatalf("Purge() error = %v", err)
	}

	if queue.Size() != 0 {
		t.Errorf("expected queue size 0 after purge, got %d", queue.Size())
	}
}

func TestMessageTemplate_Size(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test")
	template := NewMessageTemplate(queue)

	queue.Send(&Message{Body: []byte("test")})

	if template.Size() != 1 {
		t.Errorf("expected size 1, got %d", template.Size())
	}
}

func TestAcquireAndReleaseMessage(t *testing.T) {
	t.Parallel()

	// 测试从池中获取消息
	msg := AcquireMessage()
	if msg == nil {
		t.Fatal("AcquireMessage() returned nil")
	}

	// 验证消息已重置
	if msg.Headers == nil {
		t.Error("expected Headers to be initialized")
	}
	if msg.RetryCount != 0 {
		t.Errorf("expected RetryCount 0, got %d", msg.RetryCount)
	}
	if msg.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected MaxRetries %d, got %d", DefaultMaxRetries, msg.MaxRetries)
	}

	// 设置一些值
	msg.Body = []byte("test body")
	msg.ID = "test-id"

	// 释放回池
	ReleaseMessage(msg)

	// 再次获取，应该被重置
	msg2 := AcquireMessage()
	if msg2.Body != nil {
		t.Error("expected Body to be nil after release")
	}
	if msg2.ID != "" {
		t.Error("expected ID to be empty after release")
	}
}

func TestMessage_IsAcknowledged(t *testing.T) {
	t.Parallel()

	msg := &Message{}

	// 初始状态应该是未确认
	if msg.IsAcknowledged() {
		t.Error("expected message to not be acknowledged initially")
	}

	// 确认消息
	msg.Ack()

	// 现在应该已确认
	if !msg.IsAcknowledged() {
		t.Error("expected message to be acknowledged after Ack()")
	}
}

func TestMessage_SetHeader(t *testing.T) {
	t.Parallel()

	msg := &Message{}

	// 设置header
	msg.SetHeader("key1", "value1")

	// 获取header
	if msg.GetHeader("key1") != "value1" {
		t.Errorf("expected header value 'value1', got %s", msg.GetHeader("key1"))
	}

	// 测试nil Headers
	msg2 := &Message{}
	msg2.SetHeader("key2", "value2")
	if msg2.GetHeader("key2") != "value2" {
		t.Errorf("expected header value 'value2', got %s", msg2.GetHeader("key2"))
	}
}

func TestMessage_GetHeader_NilHeaders(t *testing.T) {
	t.Parallel()

	msg := &Message{}

	// 从nil Headers获取应该返回空字符串
	if msg.GetHeader("key") != "" {
		t.Errorf("expected empty string for nil Headers, got %s", msg.GetHeader("key"))
	}
}

func TestWithMaxRetries(t *testing.T) {
	t.Parallel()

	queue := NewInMemoryQueue("test", WithMaxRetries(5))

	if queue.maxRetries != 5 {
		t.Errorf("expected maxRetries 5, got %d", queue.maxRetries)
	}
}

func TestWithDeadLetterQueue(t *testing.T) {
	t.Parallel()

	dlq := NewInMemoryQueue("dead-letter")
	queue := NewInMemoryQueue("test", WithDeadLetterQueue(dlq))

	if queue.deadLetterQueue != dlq {
		t.Error("expected deadLetterQueue to be set")
	}
}
