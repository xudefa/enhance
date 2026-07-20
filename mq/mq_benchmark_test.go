package mq

import (
	"fmt"
	"testing"
)

// BenchmarkInMemoryQueue_Send 测试发送性能
func BenchmarkInMemoryQueue_Send(b *testing.B) {
	queue := NewInMemoryQueue("bench-send")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := AcquireMessage()
		msg.Body = []byte(fmt.Sprintf("msg-%d", i))
		_ = queue.Send(msg)
		ReleaseMessage(msg)
	}
}

// BenchmarkInMemoryQueue_Receive 测试接收性能
func BenchmarkInMemoryQueue_Receive(b *testing.B) {
	queue := NewInMemoryQueue("bench-receive")
	// 预填充消息
	for i := 0; i < b.N; i++ {
		msg := AcquireMessage()
		msg.Body = []byte("test")
		_ = queue.Send(msg)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg, _ := queue.Receive()
		ReleaseMessage(msg)
	}
}

// BenchmarkInMemoryQueue_SendReceive 测试发送+接收循环性能
func BenchmarkInMemoryQueue_SendReceive(b *testing.B) {
	queue := NewInMemoryQueue("bench-send-receive")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := AcquireMessage()
		msg.Body = []byte("test")
		_ = queue.Send(msg)
		receivedMsg, _ := queue.Receive()
		ReleaseMessage(msg)
		ReleaseMessage(receivedMsg)
	}
}

// BenchmarkInMemoryQueue_Concurrent_Send 测试并发发送性能
func BenchmarkInMemoryQueue_Concurrent_Send(b *testing.B) {
	queue := NewInMemoryQueue("bench-concurrent-send")
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			msg := AcquireMessage()
			msg.Body = []byte("test")
			_ = queue.Send(msg)
			ReleaseMessage(msg)
			i++
		}
	})
}

// BenchmarkInMemoryQueue_Concurrent_SendReceive 测试并发发送接收性能
func BenchmarkInMemoryQueue_Concurrent_SendReceive(b *testing.B) {
	queue := NewInMemoryQueue("bench-concurrent")
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			msg := AcquireMessage()
			msg.Body = []byte("test")
			_ = queue.Send(msg)
			receivedMsg, _ := queue.Receive()
			ReleaseMessage(msg)
			ReleaseMessage(receivedMsg)
			i++
		}
	})
}

// BenchmarkInMemoryQueue_DifferentSizes 测试不同队列大小的性能
func BenchmarkInMemoryQueue_DifferentSizes(b *testing.B) {
	b.Run("Empty-Queue", func(b *testing.B) {
		queue := NewInMemoryQueue("empty")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			msg := AcquireMessage()
			msg.Body = []byte("test")
			_ = queue.Send(msg)
			receivedMsg, _ := queue.Receive()
			ReleaseMessage(msg)
			ReleaseMessage(receivedMsg)
		}
	})

	b.Run("Small-Queue-100", func(b *testing.B) {
		queue := NewInMemoryQueue("small")
		for i := 0; i < 100; i++ {
			msg := AcquireMessage()
			msg.Body = []byte("test")
			_ = queue.Send(msg)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			receivedMsg, _ := queue.Receive()
			msg := AcquireMessage()
			msg.Body = []byte("test")
			_ = queue.Send(msg)
			ReleaseMessage(receivedMsg)
			ReleaseMessage(msg)
		}
	})

	b.Run("Large-Queue-10000", func(b *testing.B) {
		queue := NewInMemoryQueue("large")
		for i := 0; i < 10000; i++ {
			msg := AcquireMessage()
			msg.Body = []byte("test")
			_ = queue.Send(msg)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			receivedMsg, _ := queue.Receive()
			msg := AcquireMessage()
			msg.Body = []byte("test")
			_ = queue.Send(msg)
			ReleaseMessage(receivedMsg)
			ReleaseMessage(msg)
		}
	})
}

// BenchmarkMessage_Ack 测试消息确认性能
func BenchmarkMessage_Ack(b *testing.B) {
	msg := &Message{Body: []byte("test")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.Ack()
	}
}

// BenchmarkMessage_Nack 测试消息拒绝性能
func BenchmarkMessage_Nack(b *testing.B) {
	msg := &Message{Body: []byte("test")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.Nack(true)
	}
}
