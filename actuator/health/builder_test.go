package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIndicatorBuilder_Build_Up(t *testing.T) {
	t.Parallel()
	indicator := NewIndicatorBuilder().
		Name("database").
		CheckFunc(func(ctx context.Context) error {
			return nil
		}).
		Timeout(5*time.Second).
		Detail("type", "postgres").
		Build()

	healthResult := indicator.Health(context.Background())

	if healthResult.Status != StatusUp {
		t.Errorf("expected UP, got %s", healthResult.Status)
	}
	if healthResult.Details["type"] != "postgres" {
		t.Errorf("expected postgres, got %v", healthResult.Details["type"])
	}
	if healthResult.Error != nil {
		t.Errorf("expected no error, got %v", healthResult.Error)
	}
}

func TestIndicatorBuilder_Build_Down(t *testing.T) {
	t.Parallel()
	indicator := NewIndicatorBuilder().
		Name("redis").
		CheckFunc(func(ctx context.Context) error {
			return errors.New("connection refused")
		}).
		Build()

	healthResult := indicator.Health(context.Background())

	if healthResult.Status != StatusDown {
		t.Errorf("expected DOWN, got %s", healthResult.Status)
	}
	if healthResult.Error == nil {
		t.Error("expected error")
	}
}

func TestIndicatorBuilder_Build_Timeout(t *testing.T) {
	t.Parallel()
	indicator := NewIndicatorBuilder().
		Name("slow-service").
		CheckFunc(func(ctx context.Context) error {
			select {
			case <-time.After(10 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}).
		Timeout(100 * time.Millisecond).
		Build()

	start := time.Now()
	healthResult := indicator.Health(context.Background())
	elapsed := time.Since(start)

	if healthResult.Status != StatusDown {
		t.Errorf("expected DOWN due to timeout, got %s", healthResult.Status)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("expected timeout around 100ms, took %v", elapsed)
	}
}

func TestIndicatorBuilder_Build_NoCheckFunc(t *testing.T) {
	t.Parallel()
	indicator := NewIndicatorBuilder().
		Name("unknown").
		Build()

	healthResult := indicator.Health(context.Background())

	if healthResult.Status != StatusUnknown {
		t.Errorf("expected UNKNOWN, got %s", healthResult.Status)
	}
}

func TestIndicatorBuilder_ChainConfiguration(t *testing.T) {
	t.Parallel()
	indicator := NewIndicatorBuilder().
		Name("test-service").
		Timeout(3*time.Second).
		Detail("version", "1.0.0").
		Detail("env", "production").
		CheckFunc(func(ctx context.Context) error {
			return nil
		}).
		Build()

	healthResult := indicator.Health(context.Background())

	if healthResult.Status != StatusUp {
		t.Errorf("expected UP, got %s", healthResult.Status)
	}
	if healthResult.Details["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", healthResult.Details["version"])
	}
	if healthResult.Details["env"] != "production" {
		t.Errorf("expected env production, got %v", healthResult.Details["env"])
	}
}

func TestIndicatorBuilder_MultipleIndicators(t *testing.T) {
	t.Parallel()
	registry := NewAggregator()

	dbIndicator := NewIndicatorBuilder().
		Name("database").
		CheckFunc(func(ctx context.Context) error {
			return nil
		}).
		Build()

	redisIndicator := NewIndicatorBuilder().
		Name("redis").
		CheckFunc(func(ctx context.Context) error {
			return errors.New("redis down")
		}).
		Build()

	registry.AddIndicator(dbIndicator)
	registry.AddIndicator(redisIndicator)

	healthResult := registry.Aggregate(context.Background())

	// 数据库应该 UP，Redis 应该 DOWN
	details := healthResult.Details["database"].(map[string]any)
	if details["status"] != "UP" {
		t.Errorf("expected database UP, got %s", details["status"])
	}

	redisDetails := healthResult.Details["redis"].(map[string]any)
	if redisDetails["status"] != "DOWN" {
		t.Errorf("expected redis DOWN, got %s", redisDetails["status"])
	}
}
