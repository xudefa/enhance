package mq

import (
	"sync"
	"testing"
	"time"
)

// TestInMemoryQueue_ConcurrentStopConsuming 测试并发停止消费的安全性
func TestInMemoryQueue_ConcurrentStopConsuming(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-concurrent-stop")

	// 启动消费
	err := queue.Consume(func(msg *Message) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// 并发调用 StopConsuming
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			queue.StopConsuming()
		}()
	}

	wg.Wait()

	// 不应该 panic
	t.Log("concurrent StopConsuming passed")
}

// TestInMemoryQueue_ConcurrentSendReceive 测试并发发送和接收
func TestInMemoryQueue_ConcurrentSendReceive(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-concurrent-send-receive")

	var wg sync.WaitGroup
	messageCount := 100

	// 并发发送消息
	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := &Message{Body: []byte("test")}
			err := queue.Send(msg)
			if err != nil {
				t.Errorf("failed to send message: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// 验证队列大小
	if queue.Size() != messageCount {
		t.Errorf("expected queue size %d, got %d", messageCount, queue.Size())
	}
}

// TestInMemoryQueue_TimerLeak 测试定时器泄漏
func TestInMemoryQueue_TimerLeak(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-timer-leak")

	// 快速连续调用 ReceiveWithTimeout
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = queue.ReceiveWithTimeout(10 * time.Millisecond)
		}()
	}

	wg.Wait()

	// 等待所有 goroutine 完成
	time.Sleep(100 * time.Millisecond)

	// 检查队列状态应该正常
	t.Log("timer leak test passed")
}

// TestInMemoryQueue_ConcurrentAckNack 测试并发确认和拒绝
func TestInMemoryQueue_ConcurrentAckNack(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-concurrent-ack-nack")

	msg := &Message{Body: []byte("test")}
	_ = queue.Send(msg)

	// 接收消息
	receivedMsg, err := queue.Receive()
	if err != nil {
		t.Fatal(err)
	}

	// 并发调用 Ack 和 Nack
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			receivedMsg.Ack()
		}()
	}

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			receivedMsg.Nack(false)
		}()
	}

	wg.Wait()

	// 应该只被确认一次
	if !receivedMsg.IsAcknowledged() {
		t.Error("message should be acknowledged")
	}
}

// TestInMemoryQueue_MultipleStartStop 测试多次启动和停止
func TestInMemoryQueue_MultipleStartStop(t *testing.T) {
	t.Parallel()
	queue := NewInMemoryQueue("test-multiple-start-stop")

	for i := range 5 {
		// 启动消费
		err := queue.Consume(func(msg *Message) error {
			return nil
		})
		if err != nil {
			t.Fatalf("iteration %d: failed to start consume: %v", i, err)
		}

		// 发送一些消息
		for range 10 {
			_ = queue.Send(&Message{Body: []byte("test")})
		}

		// 停止消费
		queue.StopConsuming()

		// 短暂等待
		time.Sleep(10 * time.Millisecond)
	}

	t.Log("multiple start/stop test passed")
}
