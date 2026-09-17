package actuator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xudefa/enhance/actuator/health"
)

func TestHealthIndicatorBuilder_Build_Success(t *testing.T) {
	t.Parallel()
	indicator := health.NewIndicatorBuilder().
		Name("test").
		CheckFunc(func(ctx context.Context) error { return nil }).
		Timeout(2*time.Second).
		Detail("key", "value").
		Build()

	if indicator.Name() != "test" {
		t.Errorf("Expected name 'test', got %s", indicator.Name())
	}

	healthResult := indicator.Health(context.Background())
	if healthResult.Status != health.StatusUp {
		t.Errorf("Expected status UP, got %s", healthResult.Status)
	}
	if healthResult.Details["key"] != "value" {
		t.Errorf("Expected detail key=value, got %v", healthResult.Details["key"])
	}
}

func TestHealthIndicatorBuilder_Build_Failure(t *testing.T) {
	t.Parallel()
	indicator := health.NewIndicatorBuilder().
		Name("test").
		CheckFunc(func(ctx context.Context) error { return errors.New("connection failed") }).
		Build()

	healthResult := indicator.Health(context.Background())
	if healthResult.Status != health.StatusDown {
		t.Errorf("Expected status DOWN, got %s", healthResult.Status)
	}
	if healthResult.Error == nil || healthResult.Error.Error() != "connection failed" {
		t.Errorf("Expected error 'connection failed', got %v", healthResult.Error)
	}
}

func TestHealthIndicatorBuilder_Build_NoCheckFunc(t *testing.T) {
	t.Parallel()
	indicator := health.NewIndicatorBuilder().
		Name("test").
		Build()

	healthResult := indicator.Health(context.Background())
	if healthResult.Status != health.StatusUnknown {
		t.Errorf("Expected status UNKNOWN, got %s", healthResult.Status)
	}
}

func TestHealthIndicatorBuilder_Build_Timeout(t *testing.T) {
	t.Parallel()
	indicator := health.NewIndicatorBuilder().
		Name("test").
		CheckFunc(func(ctx context.Context) error {
			// 模拟一个会检查上下文的操作
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil
			}
		}).
		Timeout(50 * time.Millisecond).
		Build()

	healthResult := indicator.Health(context.Background())
	if healthResult.Status != health.StatusDown {
		t.Errorf("Expected status DOWN due to timeout, got %s", healthResult.Status)
	}
}
