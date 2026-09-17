package spel

import (
	"testing"
)

func TestEvaluate_ArithmeticSubtraction(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("20 - 5", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != float64(15) {
		t.Errorf("got %v, want 15", got)
	}
}

func TestEvaluate_ArithmeticMultiplication(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("3 * 7", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != float64(21) {
		t.Errorf("got %v, want 21", got)
	}
}

func TestEvaluate_ArithmeticDivision(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("100 / 4", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != float64(25) {
		t.Errorf("got %v, want 25", got)
	}
}

func TestEvaluate_TernaryWithVariable(t *testing.T) {
	t.Parallel()
	ctx := NewStandardEvaluationContext(testUser{Age: 20})
	ctx.SetVariable("adult", "adult")

	expr, err := ParseExpression("Age > 18 ? adult : 'minor'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "adult" {
		t.Errorf("got %v, want 'adult'", got)
	}
}

func TestEvaluate_LogicalAndFalse(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("true && false", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != false {
		t.Errorf("got %v, want false", got)
	}
}

func TestEvaluate_ComparisonLessEqual(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 30}

	got, err := Evaluate("Age <= 30", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestEvaluate_SetValueOnPropertyExpression(t *testing.T) {
	t.Parallel()
	expr := &propertyExpressionImpl{property: "Name"}
	user := &testUser{Name: "Old"}
	ctx := NewStandardEvaluationContext(user)

	err := expr.SetValue(ctx, "New")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "New" {
		t.Errorf("got %v, want 'New'", user.Name)
	}
}

func TestEvaluate_SetValueOnComplexExpression(t *testing.T) {
	t.Parallel()
	expr := &complexExpressionImpl{raw: "a + b"}
	ctx := NewStandardEvaluationContext(nil)

	err := expr.SetValue(ctx, 42)
	if err == nil {
		t.Error("expected error for setting value on complex expression")
	}
}

func TestEvaluate_InvalidLogicalExpression(t *testing.T) {
	t.Parallel()
	expr := &complexExpressionImpl{raw: "invalid"}
	ctx := NewStandardEvaluationContext(nil)

	_, err := expr.GetValue(ctx)
	if err == nil {
		t.Error("expected error for invalid logical expression")
	}
}

func TestEvaluate_ComparisonWithMixedTypes(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("'a' > 1", nil)
	if err == nil {
		t.Error("expected error for non-numeric comparison")
	}
}

func TestEvaluate_ArithmeticWithMixedTypes(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("'a' + 1", nil)
	if err == nil {
		t.Error("expected error for non-numeric arithmetic")
	}
}

func TestEvaluate_LogicalWithMultipleOperators(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 25}

	got, err := Evaluate("Age > 18 && Age < 30 && true", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestEvaluate_ComplexPropertyChain(t *testing.T) {
	t.Parallel()
	type Inner struct {
		Value int
	}
	type Outer struct {
		Inner Inner
	}

	o := Outer{Inner: Inner{Value: 42}}
	got, err := Evaluate("Inner.Value", o)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("got %v, want 42", got)
	}
}

func TestEvaluate_SetValueOnPropertyExpression_NilRoot(t *testing.T) {
	t.Parallel()
	expr := &propertyExpressionImpl{property: "Name"}
	ctx := NewStandardEvaluationContext(nil)

	err := expr.SetValue(ctx, "value")
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestEvaluate_ArithmeticWithNegativeNumbers(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("-5 + 10", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != float64(5) {
		t.Errorf("got %v, want 5", got)
	}
}

func TestEvaluate_ArithmeticWithFloats(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("1.5 + 2.5", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != float64(4) {
		t.Errorf("got %v, want 4", got)
	}
}

func TestEvaluate_LogicalOrBothTrue(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("true || true", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestEvaluate_ComparisonEqualFloat(t *testing.T) {
	t.Parallel()
	got, err := Evaluate("1.0 == 1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestEvaluate_TernaryNested(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 25}

	got, err := Evaluate("Age > 18 ? 1 : 0", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != int64(1) {
		t.Errorf("got %v (%T), want 1", got, got)
	}
}

func TestEvaluate_PropertyChainWithNilRoot(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("Name", nil)
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestEvaluate_EqualityWithVariousTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		left     any
		right    any
		expected bool
	}{
		{"nil nil", nil, nil, true},
		{"nil non-nil", nil, "a", false},
		{"non-nil nil", "a", nil, false},
		{"string equal", "hello", "hello", true},
		{"string not equal", "hello", "world", false},
		{"int equal", 42, 42, true},
		{"int not equal", 42, 43, false},
		{"bool equal", true, true, true},
		{"bool not equal", true, false, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := equals(tt.left, tt.right); got != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.expected)
			}
		})
	}
}

func TestEvaluateArithmeticUnsupported(t *testing.T) {
	t.Parallel()
	_, err := arithmetic("a", "b", "~")
	if err == nil {
		t.Error("expected error for unsupported arithmetic operator")
	}
}
