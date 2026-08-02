package async

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAsyncExecutor_SubmitAndGet(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	defer executor.Shutdown()

	future := executor.Submit(func() (any, error) {
		time.Sleep(50 * time.Millisecond)
		return "hello", nil
	})

	result, err := future.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
}

func TestAsyncExecutor_SubmitWithError(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	defer executor.Shutdown()

	expectedErr := errors.New("task failed")
	future := executor.Submit(func() (any, error) {
		return nil, expectedErr
	})

	_, err := future.Get()
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestAsyncExecutor_SubmitVoid(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	defer executor.Shutdown()

	called := false
	future := executor.SubmitVoid(func() error {
		called = true
		return nil
	})

	_, err := future.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !called {
		t.Error("task should have been called")
	}
}

func TestFuture_GetWithTimeout(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	defer executor.Shutdown()

	future := executor.Submit(func() (any, error) {
		time.Sleep(200 * time.Millisecond)
		return "result", nil
	})

	_, err := future.GetWithTimeout(50 * time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestFuture_GetWithContext(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	defer executor.Shutdown()

	future := executor.Submit(func() (any, error) {
		time.Sleep(200 * time.Millisecond)
		return "result", nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := future.GetWithContext(ctx)
	if err == nil {
		t.Fatal("expected context deadline error")
	}
}

func TestFuture_IsDone(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	defer executor.Shutdown()

	future := executor.Submit(func() (any, error) {
		time.Sleep(50 * time.Millisecond)
		return "done", nil
	})

	if future.IsDone() {
		t.Error("future should not be done yet")
	}

	_, _ = future.Get()

	if !future.IsDone() {
		t.Error("future should be done")
	}
}

func TestAsyncExecutor_Start(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)

	// 默认就是 running=true（懒启动模式，允许提交）
	if !executor.IsRunning() {
		t.Error("executor should be running by default (lazy start mode)")
	}

	// 重复调用 Start 应该幂等
	executor.Start()
	if !executor.IsRunning() {
		t.Error("executor should be running after Start()")
	}

	executor.Shutdown()
}

func TestAsyncExecutor_Shutdown(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	executor.Start()

	for range 5 {
		executor.Submit(func() (any, error) {
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})
	}

	executor.Shutdown()

	if executor.IsRunning() {
		t.Error("executor should not be running after Shutdown()")
	}
}

func TestAsyncExecutor_ShutdownWithTimeout(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	executor.Start()

	for range 3 {
		executor.Submit(func() (any, error) {
			time.Sleep(200 * time.Millisecond)
			return nil, nil
		})
	}

	err := executor.ShutdownWithTimeout(50 * time.Millisecond)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestAsyncExecutor_GetQueueSize(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 1, 10)
	defer executor.Shutdown()

	for range 5 {
		executor.Submit(func() (any, error) {
			time.Sleep(50 * time.Millisecond)
			return nil, nil
		})
	}

	size := executor.GetQueueSize()
	if size < 0 {
		t.Errorf("queue size should be non-negative, got %d", size)
	}
}

func TestAsyncExecutor_ShutdownIdempotent(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)

	executor.Shutdown()
	executor.Shutdown()
	executor.Shutdown()
}

func TestAsyncExecutor_ConcurrentSubmit(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 4, 100)
	defer executor.Shutdown()

	done := make(chan bool, 10)
	for i := range 10 {
		go func(n int) {
			future := executor.Submit(func() (any, error) {
				time.Sleep(10 * time.Millisecond)
				return n, nil
			})
			result, _ := future.Get()
			if result.(int) != n {
				t.Errorf("expected %d, got %v", n, result)
			}
			done <- true
		}(i)
	}

	for range 10 {
		<-done
	}
}

func TestAsyncExecutor_SubmitAfterShutdown(t *testing.T) {
	t.Parallel()
	t.Run("Shutdown 后 Submit 应返回错误而非 panic", func(t *testing.T) {
		executor := NewAsyncExecutor(context.Background(), 2, 10)
		executor.Start()

		// 先提交一个正常任务
		future1 := executor.Submit(func() (any, error) {
			return "ok", nil
		})
		result, _ := future1.Get()
		if result != "ok" {
			t.Errorf("expected 'ok', got %v", result)
		}

		// 关闭执行器
		executor.Shutdown()

		// Shutdown 后再提交应该返回错误，而不是 panic
		future2 := executor.Submit(func() (any, error) {
			return "should not execute", nil
		})

		_, err := future2.Get()
		if err == nil {
			t.Fatal("expected error after shutdown")
		}

		if err.Error() != "executor is shutdown" {
			t.Errorf("expected 'executor is shutdown', got '%v'", err)
		}
	})

	t.Run("ShutdownWithTimeout 后 Submit 应返回错误", func(t *testing.T) {
		executor := NewAsyncExecutor(context.Background(), 2, 10)
		executor.Start()

		// 提交一个慢任务
		executor.Submit(func() (any, error) {
			time.Sleep(500 * time.Millisecond)
			return nil, nil
		})

		// 快速关闭
		_ = executor.ShutdownWithTimeout(10 * time.Millisecond)

		// 再提交应该失败
		future := executor.Submit(func() (any, error) {
			return "test", nil
		})

		_, err := future.Get()
		if err == nil {
			t.Error("expected error after shutdown with timeout")
		}
	})

	t.Run("Shutdown 后 SubmitVoid 应返回错误", func(t *testing.T) {
		executor := NewAsyncExecutor(context.Background(), 2, 10)
		executor.Start()
		executor.Shutdown()

		// 不应该 panic
		future := executor.SubmitVoid(func() error {
			return nil
		})

		_, err := future.Get()
		if err == nil {
			t.Error("expected error after shutdown")
		}
	})
}

func TestAsyncExecutor_PanicRecovery(t *testing.T) {
	t.Parallel()
	executor := NewAsyncExecutor(context.Background(), 2, 10)
	executor.Start()
	defer executor.Shutdown()

	// 提交一个会 panic 的任务
	future := executor.Submit(func() (any, error) {
		panic("test panic")
	})

	result, err := future.Get()
	if err == nil {
		t.Fatal("expected error from panicking task")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if !strings.Contains(err.Error(), "panic") {
		t.Errorf("expected panic error, got: %v", err)
	}

	// 验证执行器在 panic 后仍能正常工作
	future2 := executor.Submit(func() (any, error) {
		return "recovered", nil
	})
	result2, err2 := future2.Get()
	if err2 != nil {
		t.Fatalf("expected no error, got: %v", err2)
	}
	if result2 != "recovered" {
		t.Errorf("expected 'recovered', got %v", result2)
	}
}
