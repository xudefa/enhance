package spel

import (
	"testing"
)

type testUser struct {
	Name string
	Age  int
}

func (u *testUser) Greet() string {
	return "Hello, " + u.Name
}

func TestParseExpression_Simple(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if expr.String() != "name" {
		t.Errorf("expected 'name', got %v", expr.String())
	}
}

func TestParseExpression_Empty(t *testing.T) {
	t.Parallel()
	_, err := ParseExpression("")
	if err == nil {
		t.Error("expected error for empty expression")
	}
}

func TestParseExpression_Complex(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("age > 18")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if expr == nil {
		t.Error("expected expression to be created")
	}
}

func TestEvaluate_Property(t *testing.T) {
	t.Parallel()
	user := testUser{Name: "Alice", Age: 30}

	got, err := Evaluate("Name", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "Alice" {
		t.Errorf("expected 'Alice', got %v", got)
	}
}

func TestEvaluate_Literal_String(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("'hello world'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "hello world" {
		t.Errorf("expected 'hello world', got %v", got)
	}
}

func TestEvaluate_Literal_Number(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != int64(42) {
		t.Errorf("expected 42, got %v", got)
	}
}

func TestEvaluate_Literal_Boolean(t *testing.T) {
	t.Parallel()
	tests := []struct {
		expr     string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			expr, err := ParseExpression(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ctx := NewStandardEvaluationContext(nil)
			got, err := expr.GetValue(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestEvaluate_Literal_Null(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("null")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestEvaluate_Comparison(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 30}

	expr, err := ParseExpression("Age > 18")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(user)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != true {
		t.Errorf("expected true, got %v", got)
	}
}

func TestEvaluate_Arithmetic(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("10 + 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != float64(15) {
		t.Errorf("expected 15, got %v", got)
	}
}

func TestEvaluate_Logical(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 25}

	expr, err := ParseExpression("Age > 18 && Age < 30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(user)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != true {
		t.Errorf("expected true, got %v", got)
	}
}

func TestEvaluate_Ternary(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 20}

	expr, err := ParseExpression("Age > 18 ? 'adult' : 'minor'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(user)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "adult" {
		t.Errorf("expected 'adult', got %v", got)
	}
}

func TestEvaluate_PropertyChain(t *testing.T) {
	t.Parallel()
	type Address struct {
		City string
	}
	type Person struct {
		Address Address
	}

	p := Person{Address: Address{City: "Beijing"}}

	expr, err := ParseExpression("Address.City")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(p)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "Beijing" {
		t.Errorf("expected 'Beijing', got %v", got)
	}
}

func TestEvaluate_MethodCall(t *testing.T) {
	t.Parallel()
	user := &testUser{Name: "Alice"}

	expr, err := ParseExpression("Greet()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(user)
	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "Hello, Alice" {
		t.Errorf("expected 'Hello, Alice', got %v", got)
	}
}
