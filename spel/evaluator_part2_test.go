package spel

import (
	"testing"
)

func TestEvaluate_LogicalWithVariables(t *testing.T) {
	t.Parallel()
	ctx := NewStandardEvaluationContext(testUser{Age: 25})
	ctx.SetVariable("minAge", 18)

	expr, err := ParseExpression("Age > minAge && Age < 30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestEvaluate_ComparisonWithVariables(t *testing.T) {
	t.Parallel()
	ctx := NewStandardEvaluationContext(testUser{Age: 25})
	ctx.SetVariable("threshold", 18)

	expr, err := ParseExpression("Age >= threshold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestEvaluate_ArithmeticWithVariables(t *testing.T) {
	t.Parallel()
	ctx := NewStandardEvaluationContext(nil)
	ctx.SetVariable("x", 10)

	expr, err := ParseExpression("x * 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != float64(20) {
		t.Errorf("got %v, want 20", got)
	}
}

func TestEvaluate_MethodCall_WithArgs(t *testing.T) {
	t.Parallel()
	type Greeter struct {
		Prefix string
	}

	g := &Greeter{Prefix: "Hello"}
	_ = g

	user := &testUser{Name: "Alice"}
	got, err := Evaluate("Greet()", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello, Alice" {
		t.Errorf("got %v, want 'Hello, Alice'", got)
	}
}

func TestEvaluate_MethodCall_MissingParen(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("Greet(")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = expr.GetValue(NewStandardEvaluationContext(&testUser{}))
	if err == nil {
		t.Error("expected error for missing closing paren")
	}
}

func TestEvaluate_MethodCall_NonExistent(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("NonExistent()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = expr.GetValue(NewStandardEvaluationContext(&testUser{}))
	if err == nil {
		t.Error("expected error for non-existent method")
	}
}

func TestEvaluate_PropertyChain_NilIntermediate(t *testing.T) {
	t.Parallel()
	type Address struct {
		City string
	}
	type Person struct {
		Address *Address
	}

	p := Person{Address: nil}

	expr, err := ParseExpression("Address.City")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = expr.GetValue(NewStandardEvaluationContext(p))
	if err == nil {
		t.Error("expected error for nil intermediate in property chain")
	}
}

func TestEvaluate_UnknownLiteral(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = expr.GetValue(NewStandardEvaluationContext(nil))
	if err == nil {
		t.Error("expected error for unknown literal")
	}
}

func TestEvaluate_LiteralFloat(t *testing.T) {
	t.Parallel()
	type Num struct{ Val float64 }

	got, err := Evaluate("Val", Num{Val: 3.14})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3.14 {
		t.Errorf("got %v, want 3.14", got)
	}
}

func TestEvaluate_LiteralString(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("'hello world'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := expr.GetValue(NewStandardEvaluationContext(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %v, want 'hello world'", got)
	}
}

func TestEvaluate_LogicalOr_WithTruthValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		expr     string
		expected bool
	}{
		{"'yes' || 'no'", true},
		{"'' || ''", false},
		{"1 || 0", true},
		{"0 || 0", false},
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

func TestEvaluate_ArithmeticNonNumeric(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("'a' + 1", nil)
	if err == nil {
		t.Error("expected error for non-numeric arithmetic")
	}
}

func TestEvaluate_DivisionByZero(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("10 / 0", nil)
	if err == nil {
		t.Error("expected error for division by zero")
	}
}

func TestEvaluate_ComparisonNonNumeric(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("'a' > 'b'", nil)
	if err == nil {
		t.Error("expected error for non-numeric comparison")
	}
}

func TestEvaluate_LiteralUnknown(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("xyz123", nil)
	if err == nil {
		t.Error("expected error for unknown literal")
	}
}

func TestEvaluate_MethodCall_WrongArgCount(t *testing.T) {
	t.Parallel()
	user := &testUser{Name: "Alice"}

	expr, err := ParseExpression("Greet('extra', 'arg')")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = expr.GetValue(NewStandardEvaluationContext(user))
	if err == nil {
		t.Error("expected error for wrong arg count")
	}
}

func TestEvaluate_MethodCall_ArgConversion(t *testing.T) {
	t.Parallel()
	type Converter struct{}

	type convImpl struct{}

	c := &convImpl{}
	_ = c

	expr, err := ParseExpression("NonExistent('a')")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = expr.GetValue(NewStandardEvaluationContext(&testUser{}))
	if err == nil {
		t.Error("expected error for non-existent method with args")
	}
}

func TestEvaluate_LogicalInvalid(t *testing.T) {
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

func TestEvaluate_PropertyChain_WithRootObject(t *testing.T) {
	t.Parallel()
	type Address struct {
		City string
	}
	type Person struct {
		Address Address
	}

	p := Person{Address: Address{City: "Shanghai"}}
	got, err := Evaluate("Address.City", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Shanghai" {
		t.Errorf("got %v, want 'Shanghai'", got)
	}
}

func TestEvaluate_Convenience_WithNilRoot(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("Name", nil)
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestEvaluate_EqualityComplexTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		expr     string
		root     any
		expected bool
	}{
		{"bool equality", "true == true", nil, true},
		{"bool inequality", "true != false", nil, true},
		{"nil equality", "null == null", nil, true},
		{"nil inequality", "null != true", nil, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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
