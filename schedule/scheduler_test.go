package schedule

import (
	"context"
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
