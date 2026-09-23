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
