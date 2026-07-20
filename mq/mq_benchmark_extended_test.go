package mq

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkMessageQueueFactory_CreateQueue(b *testing.B) {
	factory := NewMessageQueueFactory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", i))
	}
}

func BenchmarkMessageQueueFactory_GetQueue(b *testing.B) {
	factory := NewMessageQueueFactory()
	for i := 0; i < 100; i++ {
		factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = factory.GetQueue(fmt.Sprintf("queue-%d", i%100))
	}
}

func BenchmarkMessageQueueFactory_DeleteQueue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		factory := NewMessageQueueFactory()
		for j := 0; j < 10; j++ {
			factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", j))
		}
		factory.DeleteQueue("queue-5")
	}
}

func BenchmarkMessageQueueFactory_ListQueues(b *testing.B) {
	factory := NewMessageQueueFactory()
	for i := 0; i < 100; i++ {
		factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = factory.ListQueues()
	}
}

func BenchmarkMessageQueueFactory_ConcurrentCreate(b *testing.B) {
	factory := NewMessageQueueFactory()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", i))
			i++
		}
	})
}

func BenchmarkMessageQueueFactory_ConcurrentGet(b *testing.B) {
	factory := NewMessageQueueFactory()
	for i := 0; i < 100; i++ {
		factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", i))
	}
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = factory.GetQueue(fmt.Sprintf("queue-%d", i%100))
			i++
		}
	})
}

func BenchmarkMessageQueueFactory_ConcurrentMixed(b *testing.B) {
	factory := NewMessageQueueFactory()
	for i := 0; i < 50; i++ {
		factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			op := i % 3
			switch op {
			case 0:
				factory.CreateInMemoryQueue(fmt.Sprintf("queue-new-%d", i))
			case 1:
				_, _ = factory.GetQueue(fmt.Sprintf("queue-%d", i%50))
			case 2:
				_ = factory.ListQueues()
			}
			i++
		}
	})
}

func TestMessageQueueFactory_ConcurrentCreateGet(t *testing.T) {
	t.Parallel()
	factory := NewMessageQueueFactory()
	var wg sync.WaitGroup
	const goroutines = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				queueName := fmt.Sprintf("queue-%d-%d", gid, i)
				factory.CreateInMemoryQueue(queueName)
				q, err := factory.GetQueue(queueName)
				if err != nil {
					t.Errorf("failed to get queue %s: %v", queueName, err)
				}
				if q.Name() != queueName {
					t.Errorf("expected queue name %s, got %s", queueName, q.Name())
				}
			}
		}(g)
	}

	wg.Wait()

	queues := factory.ListQueues()
	if len(queues) != goroutines*100 {
		t.Errorf("expected %d queues, got %d", goroutines*100, len(queues))
	}
}

func TestMessageQueueFactory_ConcurrentDelete(t *testing.T) {
	t.Parallel()
	factory := NewMessageQueueFactory()
	const queues = 100

	for i := 0; i < queues; i++ {
		factory.CreateInMemoryQueue(fmt.Sprintf("queue-%d", i))
	}

	var wg sync.WaitGroup
	for i := 0; i < queues; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = factory.DeleteQueue(fmt.Sprintf("queue-%d", idx))
		}(i)
	}

	wg.Wait()

	queuesLeft := factory.ListQueues()
	if len(queuesLeft) != 0 {
		t.Errorf("expected 0 queues after deletion, got %d", len(queuesLeft))
	}
}
