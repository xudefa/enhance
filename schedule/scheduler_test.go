package schedule

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
)

func TestNewTask(t *testing.T) {
	t.Parallel()

	var executed int32
	task := NewTask("test-task", "0 * * * * *", func(ctx context.Context) error {
		atomic.AddInt32(&executed, 1)
		return nil
	})

	if task.Name() != "test-task" {
		t.Errorf("expected name 'test-task', got %s", task.Name())
	}

	if task.Cron() != "0 * * * * *" {
		t.Errorf("expected cron '0 * * * * *', got %s", task.Cron())
	}

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("execute failed: %v", err)
	}

	if atomic.LoadInt32(&executed) != 1 {
		t.Error("task should have been executed")
	}
}

func TestParseCronExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"every_second", "0 * * * * *", false},
		{"every_minute", "0 */5 * * * *", false},
		{"every_hour", "0 0 */1 * * *", false},
		{"daily", "0 0 0 * * *", false},
		{"workday", "0 0 0 * * MON-FRI", false},
		{"invalid_fields", "0 * * * *", true},
		{"too_many_fields", "0 * * * * * *", true},
		{"invalid_value", "60 * * * * *", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ce, err := ParseCronExpression(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if ce == nil {
				t.Error("expected cron expression, got nil")
			}
		})
	}
}

func TestCronExpression_Next(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expr     string
		from     time.Time
		expected time.Time
	}{
		{
			name:     "every_minute",
			expr:     "0 * * * * *",
			from:     time.Date(2024, 1, 1, 10, 30, 45, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 10, 31, 0, 0, time.UTC),
		},
		{
			name:     "every_5_minutes",
			expr:     "0 */5 * * * *",
			from:     time.Date(2024, 1, 1, 10, 32, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 10, 35, 0, 0, time.UTC),
		},
		{
			name:     "daily_midnight",
			expr:     "0 0 0 * * *",
			from:     time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ce, err := ParseCronExpression(tt.expr)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			next := ce.Next(tt.from)
			if !next.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, next)
			}
		})
	}
}

func TestScheduler_Register(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task := NewTask("test-task", "0 * * * * *", func(ctx context.Context) error {
		return nil
	})

	err := scheduler.Register(task)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	tasks := scheduler.RegisteredTasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	err = scheduler.Register(task)
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestScheduler_Unregister(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task := NewTask("test-task", "0 * * * * *", func(ctx context.Context) error {
		return nil
	})

	_ = scheduler.Register(task)

	if !scheduler.Unregister("test-task") {
		t.Error("unregister should succeed")
	}

	tasks := scheduler.RegisteredTasks()
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after unregister, got %d", len(tasks))
	}

	if scheduler.Unregister("non-existent") {
		t.Error("unregister non-existent task should fail")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	if scheduler.IsRunning() {
		t.Error("should not be running initially")
	}

	ctx := context.Background()
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("should be running after start")
	}

	err = scheduler.Start(ctx)
	if err == nil {
		t.Error("should error when already running")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = scheduler.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	if scheduler.IsRunning() {
		t.Error("should not be running after shutdown")
	}
}

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

	ctx := context.Background()
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

	if maxConcurrent > 2 {
		t.Errorf("expected max concurrent <= 2, got %d", maxConcurrent)
	}
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

func TestCronExpression_Next_BoundaryCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expr     string
		from     time.Time
		expected time.Time
	}{
		{
			name:     "end_of_minute",
			expr:     "0 * * * * *",
			from:     time.Date(2024, 1, 1, 10, 30, 59, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 10, 31, 0, 0, time.UTC),
		},
		{
			name:     "end_of_hour",
			expr:     "0 0 * * * *",
			from:     time.Date(2024, 1, 1, 10, 59, 59, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			name:     "end_of_day",
			expr:     "0 0 0 * * *",
			from:     time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "end_of_month",
			expr:     "0 0 0 1 * *",
			from:     time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ce, err := ParseCronExpression(tt.expr)
			if err != nil {
				t.Fatalf("failed to parse cron expression: %v", err)
			}

			next := ce.Next(tt.from)
			if !next.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, next)
			}
		})
	}
}

func TestParseCronExpression_ComplexExpressions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"step_with_start", "5/10 * * * * *", false},
		{"range_with_step", "0-30/5 * * * * *", false},
		{"multiple_ranges", "0,15,30,45 * * * * *", false},
		{"month_names", "0 0 0 1 JAN,MAR,JUL *", false},
		{"day_range", "0 0 0 * * MON-WED", false},
		{"invalid_step", "*/0 * * * * *", true},
		{"invalid_range_order", "30-10 * * * * *", true},
		{"out_of_range_second", "60 * * * * *", true},
		{"out_of_range_minute", "0 60 * * * *", true},
		{"out_of_range_hour", "0 0 24 * * *", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseCronExpression(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFixedDelayTask_Properties(t *testing.T) {
	t.Parallel()

	delay := 5 * time.Second
	task := NewFixedDelayTask("test-delay", delay, func(ctx context.Context) error {
		return nil
	})

	if task.Name() != "test-delay" {
		t.Errorf("expected name 'test-delay', got %s", task.Name())
	}

	cron := task.Cron()
	if cron != "@fixed-delay(5s)" {
		t.Errorf("expected cron '@fixed-delay(5s)', got %s", cron)
	}

	ft := task.(*fixedDelayTask)
	if ft.FixedDelay() != delay {
		t.Errorf("expected delay %v, got %v", delay, ft.FixedDelay())
	}

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("execute failed: %v", err)
	}
}

func TestFixedRateTask_Properties(t *testing.T) {
	t.Parallel()

	interval := 10 * time.Second
	task := NewFixedRateTask("test-rate", interval, func(ctx context.Context) error {
		return nil
	})

	if task.Name() != "test-rate" {
		t.Errorf("expected name 'test-rate', got %s", task.Name())
	}

	cron := task.Cron()
	if cron != "@fixed-rate(10s)" {
		t.Errorf("expected cron '@fixed-rate(10s)', got %s", cron)
	}

	ft := task.(*fixedRateTask)
	if ft.Interval() != interval {
		t.Errorf("expected interval %v, got %v", interval, ft.Interval())
	}

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("execute failed: %v", err)
	}
}

func TestScheduler_UnregisterNonExistent(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	if scheduler.Unregister("nonexistent") {
		t.Error("should return false for non-existent task")
	}
}

func TestScheduler_RegisterDuplicate(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task := NewTask("duplicate-task", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	if err := scheduler.Register(task); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	if err := scheduler.Register(task); err == nil {
		t.Error("second register should fail")
	}
}

func TestScheduler_RegisteredTasks(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task1 := NewTask("task-1", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	task2 := NewTask("task-2", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	_ = scheduler.Register(task1)
	_ = scheduler.Register(task2)

	tasks := scheduler.RegisteredTasks()
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestScheduler_WithLogger(t *testing.T) {
	t.Parallel()

	customLogger := log.NewLoggerBuilder().Build()

	scheduler := NewScheduler(context.Background(),
		WithLogger(customLogger),
	)

	if scheduler == nil {
		t.Error("expected non-nil scheduler")
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

func TestScheduler_Close(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task := NewTask("close-task", "* * * * * *", func(ctx context.Context) error {
		return nil
	})

	_ = scheduler.Register(task)

	ctx := context.Background()
	_ = scheduler.Start(ctx)

	// Close should shutdown gracefully
	scheduler.Close()

	if scheduler.IsRunning() {
		t.Error("should not be running after close")
	}
}

func TestScheduler_Close_NotRunning(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	// Close when not running should not error
	scheduler.Close()
}
