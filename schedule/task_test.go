package schedule

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFunctionTask_Execute(t *testing.T) {
	t.Parallel()

	called := false
	task := NewTask("test", "* * * * * *", func(ctx context.Context) error {
		called = true
		return nil
	})

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestFunctionTask_Execute_Error(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("task failed")
	task := NewTask("failing", "* * * * * *", func(ctx context.Context) error {
		return expectedErr
	})

	err := task.Execute(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestFixedDelayTask_Execute(t *testing.T) {
	t.Parallel()

	called := false
	task := NewFixedDelayTask("delay", time.Second, func(ctx context.Context) error {
		called = true
		return nil
	})

	if task.Name() != "delay" {
		t.Errorf("expected name 'delay', got %q", task.Name())
	}
	if task.Cron() != "@fixed-delay(1s)" {
		t.Errorf("expected cron '@fixed-delay(1s)', got %q", task.Cron())
	}

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestFixedRateTask_Execute(t *testing.T) {
	t.Parallel()

	called := false
	task := NewFixedRateTask("rate", 500*time.Millisecond, func(ctx context.Context) error {
		called = true
		return nil
	})

	if task.Name() != "rate" {
		t.Errorf("expected name 'rate', got %q", task.Name())
	}
	if task.Cron() != "@fixed-rate(500ms)" {
		t.Errorf("expected cron '@fixed-rate(500ms)', got %q", task.Cron())
	}

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestParseCronExpression_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr string
	}{
		{"empty", ""},
		{"too few fields", "* * * *"},
		{"too many fields", "* * * * * * *"},
		{"invalid second", "60 * * * * *"},
		{"invalid minute", "0 60 * * * *"},
		{"invalid hour", "0 0 24 * * *"},
		{"invalid day", "0 0 0 32 * *"},
		{"invalid month", "0 0 0 * 13 *"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseCronExpression(tt.expr)
			if err == nil {
				t.Errorf("expected error for %q, got nil", tt.expr)
			}
		})
	}
}

func TestParseCronExpression_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr string
	}{
		{"wildcard", "* * * * * *"},
		{"every 5 minutes", "0 */5 * * * *"},
		{"specific values", "0,15,30,45 * * * * *"},
		{"range", "0 0-5 * * * *"},
		{"step", "0/10 * * * * *"},
		{"day names", "0 0 0 * * MON-FRI"},
		{"month names", "0 0 0 1 JAN *"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ce, err := ParseCronExpression(tt.expr)
			if err != nil {
				t.Errorf("unexpected error for %q: %v", tt.expr, err)
			}
			if ce == nil {
				t.Error("expected non-nil CronExpression")
			}
		})
	}
}

func TestCronExpression_Next_EverySecond(t *testing.T) {
	t.Parallel()

	ce, err := ParseCronExpression("* * * * * *")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	from := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	next := ce.Next(from)
	expected := time.Date(2024, 1, 1, 12, 0, 1, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestCronExpression_Next_Daily(t *testing.T) {
	t.Parallel()

	ce, err := ParseCronExpression("0 0 0 * * *")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	from := time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC)
	next := ce.Next(from)
	expected := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestCronExpression_Next_Weekday(t *testing.T) {
	t.Parallel()

	ce, err := ParseCronExpression("0 0 0 * * MON-FRI")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// 2024-01-06 is Saturday
	from := time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC)
	next := ce.Next(from)
	expected := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC) // Monday

	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}
