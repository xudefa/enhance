package lifecycle

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkHookRegistry_Register(b *testing.B) {
	registry := NewHookRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Register(OnStartFunc(func(ctx context.Context) error {
			return nil
		}))
	}
}

func BenchmarkHookRegistry_StartAll(b *testing.B) {
	registry := NewHookRegistry()
	for i := 0; i < 100; i++ {
		registry.Register(OnStartFunc(func(ctx context.Context) error {
			return nil
		}))
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.StartAll(ctx)
	}
}

func BenchmarkHookRegistry_StopAll(b *testing.B) {
	registry := NewHookRegistry()
	for i := 0; i < 100; i++ {
		registry.Register(OnStopFunc(func(ctx context.Context) error {
			return nil
		}))
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.StopAll(ctx)
	}
}

func BenchmarkHookRegistry_Count(b *testing.B) {
	registry := NewHookRegistry()
	for i := 0; i < 1000; i++ {
		registry.Register(OnStartFunc(func(ctx context.Context) error {
			return nil
		}))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.Count()
	}
}

func BenchmarkHookRegistry_GetAll(b *testing.B) {
	registry := NewHookRegistry()
	for i := 0; i < 100; i++ {
		registry.Register(OnStartFunc(func(ctx context.Context) error {
			return nil
		}))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetAll()
	}
}

func BenchmarkHookRegistry_ConcurrentRegister(b *testing.B) {
	registry := NewHookRegistry()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			registry.Register(OnStartFunc(func(ctx context.Context) error {
				return nil
			}))
			i++
		}
	})
}

func BenchmarkHookRegistry_ConcurrentStartAll(b *testing.B) {
	registry := NewHookRegistry()
	for i := 0; i < 50; i++ {
		registry.Register(OnStartFunc(func(ctx context.Context) error {
			return nil
		}))
	}
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = registry.StartAll(ctx)
		}
	})
}

func BenchmarkHookRegistry_ConcurrentMixed(b *testing.B) {
	registry := NewHookRegistry()
	for i := 0; i < 50; i++ {
		registry.Register(OnStartFunc(func(ctx context.Context) error {
			return nil
		}))
	}
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			op := i % 4
			switch op {
			case 0:
				registry.Register(OnStartFunc(func(ctx context.Context) error {
					return nil
				}))
			case 1:
				_ = registry.StartAll(ctx)
			case 2:
				_ = registry.Count()
			case 3:
				_ = registry.GetAll()
			}
			i++
		}
	})
}

func TestHookRegistry_ConcurrentRegisterStart(t *testing.T) {
	t.Parallel()
	registry := NewHookRegistry()
	var wg sync.WaitGroup
	var startCount atomic.Int64
	const goroutines = 50

	hook := OnStartFunc(func(ctx context.Context) error {
		startCount.Add(1)
		return nil
	})

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				registry.Register(hook)
			}
		}()
	}

	wg.Wait()

	ctx := context.Background()
	err := registry.StartAll(ctx)
	if err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	expected := int64(goroutines * 10)
	if startCount.Load() != expected {
		t.Errorf("expected %d starts, got %d", expected, startCount.Load())
	}
}

func TestHookRegistry_ConcurrentStopAll(t *testing.T) {
	t.Parallel()
	registry := NewHookRegistry()
	var wg sync.WaitGroup
	var stopCount atomic.Int64

	for i := 0; i < 100; i++ {
		registry.Register(OnStopFunc(func(ctx context.Context) error {
			stopCount.Add(1)
			return nil
		}))
	}

	ctx := context.Background()
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = registry.StopAll(ctx)
		}()
	}

	wg.Wait()

	expected := int64(100 * 10)
	if stopCount.Load() != expected {
		t.Errorf("expected %d stops, got %d", expected, stopCount.Load())
	}
}

func TestHookRegistry_StressMixedOperations(t *testing.T) {
	t.Parallel()
	registry := NewHookRegistry()
	var wg sync.WaitGroup
	const goroutines = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < 100; i++ {
				op := (gid + i) % 4
				switch op {
				case 0:
					registry.Register(OnStartFunc(func(ctx context.Context) error {
						return nil
					}))
				case 1:
					_ = registry.StartAll(ctx)
				case 2:
					_ = registry.StopAll(ctx)
				case 3:
					_ = registry.Count()
				}
			}
		}(g)
	}

	wg.Wait()
}
