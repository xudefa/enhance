package spel

import (
	"testing"
)

func TestEvaluate_Equality(t *testing.T) {
	t.Parallel()
	tests := []struct {
		expr     string
		root     any
		expected any
	}{
		{"Name == 'Alice'", testUser{Name: "Alice"}, true},
		{"Name == 'Bob'", testUser{Name: "Alice"}, false},
		{"Name != 'Bob'", testUser{Name: "Alice"}, true},
		{"Name != 'Alice'", testUser{Name: "Alice"}, false},
		{"Age == 30", testUser{Age: 30}, true},
		{"Age != 25", testUser{Age: 30}, true},
		{"1 == 1", nil, true},
		{"1 != 2", nil, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.expr, func(t *testing.T) {
			t.Parallel()
			got, err := Evaluate(tt.expr, tt.root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEvaluate_LogicalOr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		expr     string
		root     any
		expected bool
	}{
		{"true || false", nil, true},
		{"false || false", nil, false},
		{"false || true", nil, true},
		{"true || true", nil, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.expr, func(t *testing.T) {
			t.Parallel()
			got, err := Evaluate(tt.expr, tt.root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEvaluate_ArithmeticOperators(t *testing.T) {
	t.Parallel()
	tests := []struct {
		expr     string
		expected float64
	}{
		{"10 - 3", 7},
		{"4 * 5", 20},
		{"100 / 10", 10},
		{"2 + 3", 5},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.expr, func(t *testing.T) {
			t.Parallel()
			got, err := Evaluate(tt.expr, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEvaluate_ComparisonOperators(t *testing.T) {
	t.Parallel()
	tests := []struct {
		expr     string
		root     any
		expected bool
	}{
		{"Age >= 30", testUser{Age: 30}, true},
		{"Age >= 31", testUser{Age: 30}, false},
		{"Age <= 30", testUser{Age: 30}, true},
		{"Age <= 29", testUser{Age: 30}, false},
		{"Age < 31", testUser{Age: 30}, true},
		{"Age < 30", testUser{Age: 30}, false},
		{"Age > 29", testUser{Age: 30}, true},
		{"Age > 30", testUser{Age: 30}, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.expr, func(t *testing.T) {
			t.Parallel()
			got, err := Evaluate(tt.expr, tt.root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEvaluate_TernaryFalse(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 15}

	got, err := Evaluate("Age > 18 ? 'adult' : 'minor'", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "minor" {
		t.Errorf("got %v, want 'minor'", got)
	}
}

func TestEvaluate_TernaryInvalid(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("true ?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = expr.GetValue(NewStandardEvaluationContext(nil))
	if err == nil {
		t.Error("expected error for invalid ternary")
	}
}
