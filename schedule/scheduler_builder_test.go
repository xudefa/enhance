package schedule

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
)

func TestSchedulerBuilder(t *testing.T) {
	t.Parallel()

	var executed int32

	builder := NewSchedulerBuilder().
		PoolSize(5).
		WithCronTask("cron-task", "* * * * * *", func(ctx context.Context) error {
			atomic.AddInt32(&executed, 1)
			return nil
		}).
		WithFixedDelayTask("delay-task", 100*time.Millisecond, func(ctx context.Context) error {
			atomic.AddInt32(&executed, 1)
			return nil
		}).
		WithFixedRateTask("rate-task", 100*time.Millisecond, func(ctx context.Context) error {
			atomic.AddInt32(&executed, 1)
			return nil
		})

	scheduler := builder.Build()

	if len(scheduler.RegisteredTasks()) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(scheduler.RegisteredTasks()))
	}

	ctx := context.Background()
	_ = scheduler.Start(ctx)

	time.Sleep(250 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_ = scheduler.Shutdown(shutdownCtx)

	count := atomic.LoadInt32(&executed)
	if count < 3 {
		t.Errorf("expected at least 3 executions, got %d", count)
	}
}

func TestScheduleHelper(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())
	helper := NewScheduleHelper(scheduler)

	_ = helper.RegisterCronTask("helper-cron", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	_ = helper.RegisterFixedDelayTask("helper-delay", 100*time.Millisecond, func(ctx context.Context) error {
		return nil
	})

	_ = helper.RegisterFixedRateTask("helper-rate", 100*time.Millisecond, func(ctx context.Context) error {
		return nil
	})

	if !helper.HasTask("helper-cron") {
		t.Error("expected helper-cron task to exist")
	}

	if helper.HasTask("nonexistent") {
		t.Error("expected nonexistent task to not exist")
	}

	if helper.GetTaskCount() != 3 {
		t.Errorf("expected 3 tasks, got %d", helper.GetTaskCount())
	}

	_ = helper.UnregisterTask("helper-cron")

	if helper.HasTask("helper-cron") {
		t.Error("expected helper-cron task to be unregistered")
	}
}

func TestSchedulerBuilder_WithTask(t *testing.T) {
	t.Parallel()

	task := NewTask("manual-task", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	builder := NewSchedulerBuilder().
		PoolSize(5).
		WithTask(task)

	scheduler := builder.Build()

	if len(scheduler.RegisteredTasks()) != 1 {
		t.Errorf("expected 1 task, got %d", len(scheduler.RegisteredTasks()))
	}
}

func TestSchedulerBuilder_MustBuild(t *testing.T) {
	t.Parallel()

	builder := NewSchedulerBuilder().
		PoolSize(3)

	scheduler := builder.MustBuild()

	if scheduler == nil {
		t.Error("expected non-nil scheduler")
	}

	// scheduler should not be running until Start is called
	if scheduler.IsRunning() {
		t.Error("scheduler should not be running before Start is called")
	}
}

func TestSchedulerBuilder_ErrorHandler(t *testing.T) {
	t.Parallel()

	var errHandled string
	var mu sync.Mutex

	builder := NewSchedulerBuilder().
		ErrorHandler(func(taskName string, err error) {
			mu.Lock()
			defer mu.Unlock()
			errHandled = taskName
		}).
		WithCronTask("error-task", "* * * * * *", func(ctx context.Context) error {
			return context.Canceled
		})

	scheduler := builder.Build()

	ctx := context.Background()
	_ = scheduler.Start(ctx)

	time.Sleep(1200 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_ = scheduler.Shutdown(shutdownCtx)

	mu.Lock()
	defer mu.Unlock()

	if errHandled != "error-task" {
		t.Errorf("expected error handler to be called for 'error-task', got '%s'", errHandled)
	}
}

func TestSchedulerBuilder_Logger(t *testing.T) {
	t.Parallel()

	// Create a custom logger (using default builder)
	customLogger := log.NewLoggerBuilder().Build()

	builder := NewSchedulerBuilder().
		Logger(customLogger).
		WithCronTask("test-task", "* * * * * *", func(ctx context.Context) error {
			return nil
		})

	scheduler := builder.Build()

	if scheduler == nil {
		t.Error("expected non-nil scheduler")
	}
}

func TestScheduleHelper_StartAndBlock(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())
	helper := NewScheduleHelper(scheduler)

	_ = helper.RegisterCronTask("block-task", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// StartAndBlock should block until context is done
	err := helper.StartAndBlock(ctx)

	if err != nil {
		// Context timeout is expected
		if err != context.DeadlineExceeded {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestSchedulerBuilder_Context(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	builder := NewSchedulerBuilder().Context(ctx)

	if builder == nil {
		t.Error("expected non-nil builder")
	}

	scheduler := builder.Build()
	if scheduler == nil {
		t.Error("expected non-nil scheduler")
	}
}
