package event

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkDeadLetterQueue_Add(b *testing.B) {
	dlq := NewDeadLetterQueue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: "bench.event", EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(time.Second),
		})
	}
}

func BenchmarkDeadLetterQueue_Size(b *testing.B) {
	dlq := NewDeadLetterQueue()
	for i := 0; i < 1000; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: "bench.event", EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(time.Second),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dlq.Size()
	}
}

func BenchmarkDeadLetterQueue_Peek(b *testing.B) {
	dlq := NewDeadLetterQueue()
	for i := 0; i < 100; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: "bench.event", EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(-time.Second),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dlq.Peek()
	}
}

func BenchmarkDeadLetterQueue_Events(b *testing.B) {
	dlq := NewDeadLetterQueue()
	for i := 0; i < 100; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: "bench.event", EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(time.Second),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dlq.Events()
	}
}

func BenchmarkDeadLetterQueue_Stats(b *testing.B) {
	dlq := NewDeadLetterQueue()
	for i := 0; i < 100; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: fmt.Sprintf("event.%d", i%5), EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
			RetryCount:  i % 4,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(-time.Second),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dlq.Stats()
	}
}

func BenchmarkDeadLetterQueue_ConcurrentAdd(b *testing.B) {
	dlq := NewDeadLetterQueue()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			dlq.Add(FailedEvent{
				Event:       &BaseEvent{EventType: "bench.event", EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
				RetryCount:  0,
				MaxRetries:  3,
				NextRetryAt: time.Now().Add(time.Second),
			})
			i++
		}
	})
}

func BenchmarkDeadLetterQueue_ConcurrentSize(b *testing.B) {
	dlq := NewDeadLetterQueue()
	for i := 0; i < 1000; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: "bench.event", EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(time.Second),
		})
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = dlq.Size()
		}
	})
}

func BenchmarkDeadLetterQueue_ConcurrentReadWrite(b *testing.B) {
	dlq := NewDeadLetterQueue()
	var counter atomic.Int64

	// 预填充一些数据
	for i := 0; i < 100; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: "bench.event", EventTime: time.Unix(0, int64(i))},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(time.Second),
		})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 使用原子计数器生成唯一时间戳，避免 key 冲突
			ts := time.Unix(0, counter.Add(1))
			dlq.Add(FailedEvent{
				Event:       &BaseEvent{EventType: "bench.event", EventTime: ts},
				RetryCount:  0,
				MaxRetries:  3,
				NextRetryAt: time.Now().Add(time.Second),
			})
			_ = dlq.Size()
		}
	})
}

func BenchmarkEventBusWithDeadLetter_Publish(b *testing.B) {
	bus := NewEventBusWithDeadLetter(
		WithMaxRetries(0),
	)
	bus.Subscribe("bench.event", func(e ApplicationEvent) {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(&BaseEvent{EventType: "bench.event"})
	}
}

func BenchmarkEventBusWithDeadLetter_ConcurrentPublish(b *testing.B) {
	bus := NewEventBusWithDeadLetter(
		WithMaxRetries(0),
	)
	bus.Subscribe("bench.event", func(e ApplicationEvent) {})
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish(&BaseEvent{EventType: "bench.event"})
		}
	})
}

func BenchmarkEventBusWithDeadLetter_ConcurrentPublishMultipleListeners(b *testing.B) {
	bus := NewEventBusWithDeadLetter(
		WithMaxRetries(0),
	)
	for i := 0; i < 10; i++ {
		bus.Subscribe("bench.event", func(e ApplicationEvent) {})
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Publish(&BaseEvent{EventType: "bench.event"})
		}
	})
}

func BenchmarkDeadLetterQueue_StressAddRemove(b *testing.B) {
	dlq := NewDeadLetterQueue()
	var counter atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ts := time.Now().Add(time.Duration(counter.Add(1)) * time.Nanosecond)
			fe := FailedEvent{
				Event:       &BaseEvent{EventType: "stress.event", EventTime: ts},
				RetryCount:  0,
				MaxRetries:  3,
				NextRetryAt: time.Now().Add(time.Second),
			}
			dlq.Add(fe)
			dlq.Remove(&BaseEvent{EventType: "stress.event", EventTime: ts})
		}
	})
}

func BenchmarkDeadLetterQueue_StressMixedOperations(b *testing.B) {
	dlq := NewDeadLetterQueue()
	var counter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			op := i % 4
			switch op {
			case 0:
				ts := time.Now().Add(time.Duration(counter.Add(1)) * time.Nanosecond)
				dlq.Add(FailedEvent{
					Event:       &BaseEvent{EventType: "stress.event", EventTime: ts},
					RetryCount:  0,
					MaxRetries:  3,
					NextRetryAt: time.Now().Add(time.Second),
				})
			case 1:
				_ = dlq.Size()
			case 2:
				_, _ = dlq.Peek()
			case 3:
				_ = dlq.Stats()
			}
			i++
		}
	})
}

func BenchmarkDeadLetterQueue_GetByType(b *testing.B) {
	dlq := NewDeadLetterQueue()
	for i := 0; i < 500; i++ {
		dlq.Add(FailedEvent{
			Event:       &BaseEvent{EventType: fmt.Sprintf("event.%d", i%10), EventTime: time.Now().Add(time.Duration(i) * time.Nanosecond)},
			RetryCount:  0,
			MaxRetries:  3,
			NextRetryAt: time.Now().Add(time.Second),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dlq.GetByType("event.5")
	}
}

func BenchmarkDeadLetterQueue_RemoveByType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dlq := NewDeadLetterQueue()
		for j := 0; j < 100; j++ {
			dlq.Add(FailedEvent{
				Event:       &BaseEvent{EventType: fmt.Sprintf("event.%d", j%5), EventTime: time.Now().Add(time.Duration(j) * time.Nanosecond)},
				RetryCount:  0,
				MaxRetries:  3,
				NextRetryAt: time.Now().Add(time.Second),
			})
		}
		dlq.RemoveByType("event.2")
	}
}

func TestDeadLetterQueue_StressConcurrentAddRemove(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()
	var wg sync.WaitGroup
	var counter atomic.Int64
	const goroutines = 100
	const opsPerGoroutine = 100

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				idx := counter.Add(1)
				ts := time.Unix(0, idx)
				fe := FailedEvent{
					Event:       &BaseEvent{EventType: "stress.event", EventTime: ts},
					RetryCount:  0,
					MaxRetries:  3,
					NextRetryAt: time.Now().Add(-time.Second),
				}
				dlq.Add(fe)
			}
		}(g)
	}

	wg.Wait()

	if dlq.Size() != goroutines*opsPerGoroutine {
		t.Errorf("expected %d events, got %d", goroutines*opsPerGoroutine, dlq.Size())
	}
}

func TestDeadLetterQueue_StressMixedConcurrent(t *testing.T) {
	t.Parallel()
	dlq := NewDeadLetterQueue()
	var wg sync.WaitGroup
	const goroutines = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				op := (gid + i) % 4
				switch op {
				case 0:
					ts := time.Now().Add(time.Duration(gid*100+i) * time.Nanosecond)
					dlq.Add(FailedEvent{
						Event:       &BaseEvent{EventType: "stress.event", EventTime: ts},
						RetryCount:  0,
						MaxRetries:  3,
						NextRetryAt: time.Now().Add(time.Second),
					})
				case 1:
					_ = dlq.Size()
				case 2:
					_, _ = dlq.Peek()
				case 3:
					_ = dlq.Stats()
				}
			}
		}(g)
	}

	wg.Wait()
}
