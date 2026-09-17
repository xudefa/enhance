package schedule

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_TaskExecution(t *testing.T) {
	t.Parallel()

	var executed int32
	var mu sync.Mutex
	execTimes := make([]time.Time, 0)

	task := NewTask("frequent-task", "* * * * * *", func(ctx context.Context) error {
		atomic.AddInt32(&executed, 1)
		mu.Lock()
		execTimes = append(execTimes, time.Now())
		mu.Unlock()
		return nil
	})

	scheduler := NewScheduler(context.Background(), WithPoolSize(5))

	err := scheduler.Register(task)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	ctx := context.Background()
	err = scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(2500 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = scheduler.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	if atomic.LoadInt32(&executed) < 2 {
		t.Errorf("expected at least 2 executions, got %d", atomic.LoadInt32(&executed))
	}
}

func TestScheduler_ErrorHandler(t *testing.T) {
	t.Parallel()

	var errHandled string
	var mu sync.Mutex

	errorHandler := func(taskName string, err error) {
		mu.Lock()
		defer mu.Unlock()
		errHandled = taskName
	}

	task := NewTask("failing-task", "* * * * * *", func(ctx context.Context) error {
		return context.Canceled
	})

	scheduler := NewScheduler(context.Background(), WithErrorHandler(errorHandler))

	err := scheduler.Register(task)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	ctx := context.Background()
	err = scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_ = scheduler.Shutdown(shutdownCtx)

	mu.Lock()
	defer mu.Unlock()

	if errHandled != "failing-task" {
		t.Errorf("expected error handler to be called for 'failing-task', got %s", errHandled)
	}
}

func TestScheduler_ConcurrentControl(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	maxConcurrent := testSchedulerConcurrentMax(t, ctx)

	if maxConcurrent > 2 {
		t.Errorf("expected max concurrent <= 2, got %d", maxConcurrent)
	}
}

func testSchedulerConcurrentMax(t *testing.T, ctx context.Context) int32 {
	var concurrent int32
	var maxConcurrent int32
	var mu sync.Mutex

	task := NewTask("slow-task", "* * * * * *", func(ctx context.Context) error {
		current := atomic.AddInt32(&concurrent, 1)
		mu.Lock()
		if current > maxConcurrent {
			maxConcurrent = current
		}
		mu.Unlock()

		time.Sleep(500 * time.Millisecond)
		atomic.AddInt32(&concurrent, -1)
		return nil
	})

	scheduler := NewScheduler(context.Background(), WithPoolSize(2))

	for i := 0; i < 5; i++ {
		err := scheduler.Register(NewTask(
			"slow-task",
			"* * * * * *",
			task.Execute,
		))
		if err != nil {
			t.Logf("task %d register error (expected): %v", i, err)
		}
	}

	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_ = scheduler.Shutdown(shutdownCtx)

	mu.Lock()
	defer mu.Unlock()
	return maxConcurrent
}

func TestScheduler_FixedDelayTask(t *testing.T) {
	t.Parallel()

	var executed int32

	task := NewFixedDelayTask("fixed-delay-task", 100*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&executed, 1)
		return nil
	})

	scheduler := NewScheduler(context.Background())
	_ = scheduler.Register(task)

	ctx := context.Background()
	_ = scheduler.Start(ctx)

	time.Sleep(350 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_ = scheduler.Shutdown(shutdownCtx)

	count := atomic.LoadInt32(&executed)
	if count < 2 {
		t.Errorf("expected at least 2 executions, got %d", count)
	}
}

func TestScheduler_FixedRateTask(t *testing.T) {
	t.Parallel()

	var executed int32

	task := NewFixedRateTask("fixed-rate-task", 100*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&executed, 1)
		return nil
	})

	scheduler := NewScheduler(context.Background())
	_ = scheduler.Register(task)

	ctx := context.Background()
	_ = scheduler.Start(ctx)

	time.Sleep(350 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_ = scheduler.Shutdown(shutdownCtx)

	count := atomic.LoadInt32(&executed)
	if count < 2 {
		t.Errorf("expected at least 2 executions, got %d", count)
	}
}
