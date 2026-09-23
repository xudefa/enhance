package health

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestStatus_String_AllValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status Status
		want   string
	}{
		{StatusUp, "UP"},
		{StatusDown, "DOWN"},
		{StatusDegraded, "DEGRADED"},
		{StatusOutage, "OUTAGE"},
		{StatusUnknown, "UNKNOWN"},
		{Status(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

type mockIndicator struct {
	name   string
	health Health
}

func (i *mockIndicator) Name() string { return i.name }

func (i *mockIndicator) Health(_ context.Context) Health {
	return i.health
}

func TestAggregatorHelper_NewAggregator(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	if aggregator == nil {
		t.Fatal("NewAggregator returned nil")
	}
	if len(aggregator.Indicators()) != 0 {
		t.Errorf("expected 0 indicators, got %d", len(aggregator.Indicators()))
	}
}

func TestAggregatorHelper_AddAndList(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&mockIndicator{name: "a", health: Health{Status: StatusUp}})
	aggregator.AddIndicator(&mockIndicator{name: "b", health: Health{Status: StatusDown}})

	if len(aggregator.Indicators()) != 2 {
		t.Errorf("expected 2 indicators, got %d", len(aggregator.Indicators()))
	}
}

func TestAggregatorHelper_IndicatorsReturnsCopy(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&mockIndicator{name: "a", health: Health{Status: StatusUp}})

	indicators := aggregator.Indicators()
	indicators = append(indicators, &mockIndicator{name: "b", health: Health{Status: StatusDown}})

	if len(aggregator.Indicators()) != 1 {
		t.Error("modifying returned slice should not affect aggregator")
	}
}

func TestAggregatorHelper_AggregateAllUp(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusUp}})
	aggregator.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusUp}})

	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusUp {
		t.Errorf("expected UP, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_AggregateOneDown(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusUp}})
	aggregator.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusDown}})

	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusDown {
		t.Errorf("expected DOWN, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_AggregateOutage(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusDown}})
	aggregator.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusOutage}})

	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusOutage {
		t.Errorf("expected OUTAGE, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_AggregateDegraded(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusUp}})
	aggregator.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusDegraded}})

	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusDegraded {
		t.Errorf("expected DEGRADED, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_AggregateWithError(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&mockIndicator{
		name:   "failing",
		health: Health{Status: StatusDown, Error: fmt.Errorf("connection refused")},
	})

	aggregateResult := aggregator.Aggregate(context.Background())

	d, ok := aggregateResult.Details["failing"]
	if !ok {
		t.Fatal("should contain failing details")
	}
	detail, _ := d.(map[string]any)
	if _, ok := detail["error"]; !ok {
		t.Error("should contain error in detail")
	}
}

func TestAggregatorHelper_AggregateEmpty(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusUp {
		t.Errorf("expected UP for empty aggregator, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_Timeout(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&slowMockIndicator{delay: 10 * time.Second})

	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusDown {
		t.Errorf("expected DOWN for timeout, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_CancelledContext(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&slowMockIndicator{delay: 10 * time.Second})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	aggregateResult := aggregator.Aggregate(ctx)

	if aggregateResult.Status != StatusDown {
		t.Errorf("expected DOWN for cancelled context, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_PanicRecovery(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	aggregator.AddIndicator(&panicMockIndicator{})

	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusDown {
		t.Errorf("expected DOWN for panic, got %s", aggregateResult.Status)
	}
}

func TestAggregatorHelper_ConcurrentAggregate(t *testing.T) {
	t.Parallel()
	aggregator := NewAggregator()
	for i := 0; i < 10; i++ {
		aggregator.AddIndicator(&mockIndicator{
			name:   fmt.Sprintf("ind-%d", i),
			health: Health{Status: StatusUp},
		})
	}

	aggregateResult := aggregator.Aggregate(context.Background())

	if aggregateResult.Status != StatusUp {
		t.Errorf("expected UP, got %s", aggregateResult.Status)
	}
}

type slowMockIndicator struct {
	delay time.Duration
}

func (s *slowMockIndicator) Name() string { return "slow" }

func (s *slowMockIndicator) Health(ctx context.Context) Health {
	select {
	case <-time.After(s.delay):
		return Health{Status: StatusUp}
	case <-ctx.Done():
		return Health{Status: StatusDown}
	}
}

type panicMockIndicator struct{}

func (p *panicMockIndicator) Name() string { return "panic" }

func (p *panicMockIndicator) Health(_ context.Context) Health {
	panic("test panic")
}

func TestDefaultIndicatorTimeoutValue(t *testing.T) {
	t.Parallel()
	if DefaultIndicatorTimeout != 5*time.Second {
		t.Errorf("DefaultIndicatorTimeout = %v, want 5s", DefaultIndicatorTimeout)
	}
}
