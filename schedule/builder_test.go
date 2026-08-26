package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
)

func TestSchedulerBuilder_DefaultPoolSize(t *testing.T) {
	t.Parallel()

	builder := NewSchedulerBuilder()
	if builder.poolSize != DefaultSchedulePoolSize {
		t.Errorf("expected default pool size %d, got %d", DefaultSchedulePoolSize, builder.poolSize)
	}
	if builder.ctx != nil {
		t.Error("expected nil default context")
	}
	if builder.errorHandler != nil {
		t.Error("expected nil default error handler")
	}
	if builder.logger != nil {
		t.Error("expected nil default logger")
	}
	if len(builder.tasks) != 0 {
		t.Errorf("expected empty tasks, got %d", len(builder.tasks))
	}
}

func TestSchedulerBuilder_PoolSize(t *testing.T) {
	t.Parallel()

	b := NewSchedulerBuilder().PoolSize(10)
	if b.poolSize != 10 {
		t.Errorf("expected pool size 10, got %d", b.poolSize)
	}
}

func TestSchedulerBuilder_Context_Nil(t *testing.T) {
	t.Parallel()

	b := NewSchedulerBuilder()
	scheduler := b.Build()

	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestSchedulerBuilder_Build_WithAllOptions(t *testing.T) {
	t.Parallel()

	var errHandled string
	customLogger := log.NewLoggerBuilder().Build()

	scheduler := NewSchedulerBuilder().
		Context(context.Background()).
		PoolSize(8).
		ErrorHandler(func(taskName string, err error) {
			errHandled = taskName
		}).
		Logger(customLogger).
		WithCronTask("test-task", "* * * * * *", func(ctx context.Context) error {
			return nil
		}).
		WithFixedDelayTask("delay-task", 100*time.Millisecond, func(ctx context.Context) error {
			return nil
		}).
		WithFixedRateTask("rate-task", 100*time.Millisecond, func(ctx context.Context) error {
			return nil
		}).
		Build()

	if len(scheduler.RegisteredTasks()) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(scheduler.RegisteredTasks()))
	}

	_ = errHandled
}

func TestScheduleHelper_GetTaskCount(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())
	helper := NewScheduleHelper(scheduler)

	if helper.GetTaskCount() != 0 {
		t.Errorf("expected 0 tasks, got %d", helper.GetTaskCount())
	}

	_ = helper.RegisterCronTask("t1", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	if helper.GetTaskCount() != 1 {
		t.Errorf("expected 1 task, got %d", helper.GetTaskCount())
	}
}

func TestScheduleHelper_HasTask(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())
	helper := NewScheduleHelper(scheduler)

	_ = helper.RegisterCronTask("existing", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	if !helper.HasTask("existing") {
		t.Error("expected 'existing' task to exist")
	}
	if helper.HasTask("nonexistent") {
		t.Error("expected 'nonexistent' task to not exist")
	}
}
