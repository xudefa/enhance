package schedule

import (
	"context"
	"testing"
	"time"
)

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
