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
	a := NewAggregator()
	if a == nil {
		t.Fatal("NewAggregator returned nil")
	}
	if len(a.Indicators()) != 0 {
		t.Errorf("expected 0 indicators, got %d", len(a.Indicators()))
	}
}

func TestAggregatorHelper_AddAndList(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&mockIndicator{name: "a", health: Health{Status: StatusUp}})
	a.AddIndicator(&mockIndicator{name: "b", health: Health{Status: StatusDown}})

	if len(a.Indicators()) != 2 {
		t.Errorf("expected 2 indicators, got %d", len(a.Indicators()))
	}
}

func TestAggregatorHelper_IndicatorsReturnsCopy(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&mockIndicator{name: "a", health: Health{Status: StatusUp}})

	indicators := a.Indicators()
	indicators = append(indicators, &mockIndicator{name: "b", health: Health{Status: StatusDown}})

	if len(a.Indicators()) != 1 {
		t.Error("modifying returned slice should not affect aggregator")
	}
}

func TestAggregatorHelper_AggregateAllUp(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusUp}})
	a.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusUp}})

	result := a.Aggregate(context.Background())

	if result.Status != StatusUp {
		t.Errorf("expected UP, got %s", result.Status)
	}
}

func TestAggregatorHelper_AggregateOneDown(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusUp}})
	a.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusDown}})

	result := a.Aggregate(context.Background())

	if result.Status != StatusDown {
		t.Errorf("expected DOWN, got %s", result.Status)
	}
}

func TestAggregatorHelper_AggregateOutage(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusDown}})
	a.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusOutage}})

	result := a.Aggregate(context.Background())

	if result.Status != StatusOutage {
		t.Errorf("expected OUTAGE, got %s", result.Status)
	}
}

func TestAggregatorHelper_AggregateDegraded(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&mockIndicator{name: "db", health: Health{Status: StatusUp}})
	a.AddIndicator(&mockIndicator{name: "cache", health: Health{Status: StatusDegraded}})

	result := a.Aggregate(context.Background())

	if result.Status != StatusDegraded {
		t.Errorf("expected DEGRADED, got %s", result.Status)
	}
}

func TestAggregatorHelper_AggregateWithError(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&mockIndicator{
		name:   "failing",
		health: Health{Status: StatusDown, Error: fmt.Errorf("connection refused")},
	})

	result := a.Aggregate(context.Background())

	d, ok := result.Details["failing"]
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
	a := NewAggregator()
	result := a.Aggregate(context.Background())

	if result.Status != StatusUp {
		t.Errorf("expected UP for empty aggregator, got %s", result.Status)
	}
}

func TestAggregatorHelper_Timeout(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&slowMockIndicator{delay: 10 * time.Second})

	result := a.Aggregate(context.Background())

	if result.Status != StatusDown {
		t.Errorf("expected DOWN for timeout, got %s", result.Status)
	}
}

func TestAggregatorHelper_CancelledContext(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&slowMockIndicator{delay: 10 * time.Second})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result := a.Aggregate(ctx)

	if result.Status != StatusDown {
		t.Errorf("expected DOWN for cancelled context, got %s", result.Status)
	}
}

func TestAggregatorHelper_PanicRecovery(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	a.AddIndicator(&panicMockIndicator{})

	result := a.Aggregate(context.Background())

	if result.Status != StatusDown {
		t.Errorf("expected DOWN for panic, got %s", result.Status)
	}
}

func TestAggregatorHelper_ConcurrentAggregate(t *testing.T) {
	t.Parallel()
	a := NewAggregator()
	for i := 0; i < 10; i++ {
		a.AddIndicator(&mockIndicator{
			name:   fmt.Sprintf("ind-%d", i),
			health: Health{Status: StatusUp},
		})
	}

	result := a.Aggregate(context.Background())

	if result.Status != StatusUp {
		t.Errorf("expected UP, got %s", result.Status)
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
