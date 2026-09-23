package async

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkAsyncExecutor_Submit(b *testing.B) {
	executor := NewAsyncExecutor(context.Background(), 4, 100)
	defer executor.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := executor.Submit(func() (any, error) {
			return i, nil
		})
		_, _ = future.Get()
	}
}

func BenchmarkAsyncExecutor_SubmitVoid(b *testing.B) {
	executor := NewAsyncExecutor(context.Background(), 4, 100)
	defer executor.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := executor.SubmitVoid(func() error {
			return nil
		})
		_, _ = future.Get()
	}
}

func BenchmarkAsyncExecutor_Concurrent(b *testing.B) {
	executor := NewAsyncExecutor(context.Background(), 8, 1000)
	defer executor.Shutdown()

	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			future := executor.Submit(func() (any, error) {
				return seq, nil
			})
			_, _ = future.Get()
			seq++
		}
	})
}

func BenchmarkAsyncExecutor_DifferentWorkerCounts(b *testing.B) {
	b.Run("1-Worker", func(b *testing.B) {
		benchAsyncExecutorWorkers(b, 1)
	})

	b.Run("4-Workers", func(b *testing.B) {
		benchAsyncExecutorWorkers(b, 4)
	})

	b.Run("8-Workers", func(b *testing.B) {
		benchAsyncExecutorWorkers(b, 8)
	})

	b.Run("16-Workers", func(b *testing.B) {
		benchAsyncExecutorWorkers(b, 16)
	})
}

func benchAsyncExecutorWorkers(b *testing.B, workers int) {
	executor := NewAsyncExecutor(context.Background(), workers, 100)
	defer executor.Shutdown()

	b.RunParallel(func(pb *testing.PB) {
		seq := 0
		for pb.Next() {
			future := executor.Submit(func() (any, error) {
				return seq, nil
			})
			_, _ = future.Get()
			seq++
		}
	})
}

func BenchmarkFuture_GetWithTimeout(b *testing.B) {
	executor := NewAsyncExecutor(context.Background(), 4, 100)
	defer executor.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		future := executor.Submit(func() (any, error) {
			return fmt.Sprintf("result-%d", i), nil
		})
		_, _ = future.GetWithTimeout(1000000) // 1 second timeout
	}
}
