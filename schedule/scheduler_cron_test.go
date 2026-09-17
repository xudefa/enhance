package schedule

import (
	"testing"
	"time"
)

func TestCronExpression_Next_BoundaryCases(t *testing.T) {
	t.Parallel()

	tests := testCronExpressionNextBoundaryCases()

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

type cronBoundaryCase struct {
	name     string
	expr     string
	from     time.Time
	expected time.Time
}

func testCronExpressionNextBoundaryCases() []cronBoundaryCase {
	return []cronBoundaryCase{
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
