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
			val, err := Evaluate(tt.expr, tt.root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %v, want %v", val, tt.expected)
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
			val, err := Evaluate(tt.expr, tt.root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %v, want %v", val, tt.expected)
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
			val, err := Evaluate(tt.expr, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %v, want %v", val, tt.expected)
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
			val, err := Evaluate(tt.expr, tt.root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestPropertyExpression_GetValue_NilRoot(t *testing.T) {
	t.Parallel()
	expr := &propertyExpressionImpl{property: "Name"}
	ctx := NewStandardEvaluationContext(nil)

	_, err := expr.GetValue(ctx)
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestPropertyExpression_SetValue_NilRoot(t *testing.T) {
	t.Parallel()
	expr := &propertyExpressionImpl{property: "Name"}
	ctx := NewStandardEvaluationContext(nil)

	err := expr.SetValue(ctx, "value")
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestPropertyExpression_String(t *testing.T) {
	t.Parallel()
	expr := &propertyExpressionImpl{property: "Name"}
	if expr.String() != "Name" {
		t.Errorf("expected 'Name', got %v", expr.String())
	}
}

func TestComplexExpression_String(t *testing.T) {
	t.Parallel()
	expr := &complexExpressionImpl{raw: "age > 18"}
	if expr.String() != "age > 18" {
		t.Errorf("expected 'age > 18', got %v", expr.String())
	}
}

func TestEvaluate_TernaryFalse(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 15}

	val, err := Evaluate("Age > 18 ? 'adult' : 'minor'", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "minor" {
		t.Errorf("got %v, want 'minor'", val)
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

	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != true {
		t.Errorf("got %v, want true", val)
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

	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != true {
		t.Errorf("got %v, want true", val)
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

	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != float64(20) {
		t.Errorf("got %v, want 20", val)
	}
}

func TestPropertyExpression_GetValue_WithVariable(t *testing.T) {
	t.Parallel()
	expr := &propertyExpressionImpl{property: "myVar"}
	ctx := NewStandardEvaluationContext(nil)
	ctx.SetVariable("myVar", 42)

	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Errorf("got %v, want 42", val)
	}
}

func TestPropertyExpression_GetValue_WithTag(t *testing.T) {
	t.Parallel()
	type TaggedUser struct {
		FullName string `json:"full_name"`
	}

	accessor := NewReflectPropertyAccessor()
	u := TaggedUser{FullName: "Alice"}

	val, err := accessor.GetProperty(u, "full_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Alice" {
		t.Errorf("got %v, want 'Alice'", val)
	}
}

func TestReflectPropertyAccessor_SetProperty_PointerTarget(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := &testUser{Name: "Old"}

	err := accessor.SetProperty(user, "Name", "New")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Name != "New" {
		t.Errorf("got %v, want 'New'", user.Name)
	}
}

func TestReflectPropertyAccessor_GetProperty_PointerTarget(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := &testUser{Name: "Alice"}

	val, err := accessor.GetProperty(user, "Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Alice" {
		t.Errorf("got %v, want 'Alice'", val)
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
	val, err := Evaluate("Greet()", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Hello, Alice" {
		t.Errorf("got %v, want 'Hello, Alice'", val)
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

	val, err := Evaluate("Val", Num{Val: 3.14})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 3.14 {
		t.Errorf("got %v, want 3.14", val)
	}
}

func TestEvaluate_LiteralString(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("'hello world'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := expr.GetValue(NewStandardEvaluationContext(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hello world" {
		t.Errorf("got %v, want 'hello world'", val)
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
			val, err := Evaluate(tt.expr, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %v, want %v", val, tt.expected)
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
	val, err := Evaluate("Address.City", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "Shanghai" {
		t.Errorf("got %v, want 'Shanghai'", val)
	}
}

func TestEvaluate_Convenience_WithNilRoot(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("Name", nil)
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestSplitArgsRespectingQuotes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", []string{}},
		{"single", "a", []string{"a"}},
		{"multiple", "a,b,c", []string{"a", "b", "c"}},
		{"quoted comma", "'a,b',c", []string{"'a,b'", "c"}},
		{"spaces", " a , b ", []string{" a ", " b "}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitArgsRespectingQuotes(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d args, want %d: %v", len(got), len(tt.expected), got)
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("got[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestIsTruthyAdvanced(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"uint positive", uint(1), true},
		{"uint zero", uint(0), false},
		{"float positive", 1.5, true},
		{"float zero", 0.0, false},
		{"struct", struct{}{}, true},
		{"slice", []int{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTruthy(tt.input); got != tt.expected {
				t.Errorf("isTruthy(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
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
			val, err := Evaluate(tt.expr, tt.root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestEvaluate_ArithmeticSubtraction(t *testing.T) {
	t.Parallel()
	val, err := Evaluate("20 - 5", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != float64(15) {
		t.Errorf("got %v, want 15", val)
	}
}

func TestEvaluate_ArithmeticMultiplication(t *testing.T) {
	t.Parallel()
	val, err := Evaluate("3 * 7", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != float64(21) {
		t.Errorf("got %v, want 21", val)
	}
}

func TestEvaluate_ArithmeticDivision(t *testing.T) {
	t.Parallel()
	val, err := Evaluate("100 / 4", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != float64(25) {
		t.Errorf("got %v, want 25", val)
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

	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "adult" {
		t.Errorf("got %v, want 'adult'", val)
	}
}

func TestEvaluate_LogicalAndFalse(t *testing.T) {
	t.Parallel()
	val, err := Evaluate("true && false", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != false {
		t.Errorf("got %v, want false", val)
	}
}

func TestEvaluate_ComparisonLessEqual(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 30}

	val, err := Evaluate("Age <= 30", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != true {
		t.Errorf("got %v, want true", val)
	}
}

func TestReflectPropertyAccessor_SetProperty_ConvertibleTypes(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	target := &testSettable{Count: 0}

	err := accessor.SetProperty(target, "Count", int(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Count != 42 {
		t.Errorf("got %d, want 42", target.Count)
	}
}

func TestReflectPropertyAccessor_SetProperty_InconvertibleTypes(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	target := &testSettable{Name: ""}

	type incompatibleStruct struct{ X int }
	err := accessor.SetProperty(target, "Name", incompatibleStruct{X: 1})
	if err == nil {
		t.Error("expected error for inconvertible types")
	}
}

func TestReflectPropertyAccessor_GetProperty_NotFound(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := testUser{Name: "Alice"}

	_, err := accessor.GetProperty(user, "NonExistent")
	if err == nil {
		t.Error("expected error for non-existent property")
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

	val, err := Evaluate("Age > 18 && Age < 30 && true", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != true {
		t.Errorf("got %v, want true", val)
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
	val, err := Evaluate("Inner.Value", o)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Errorf("got %v, want 42", val)
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
	val, err := Evaluate("-5 + 10", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != float64(5) {
		t.Errorf("got %v, want 5", val)
	}
}

func TestEvaluate_ArithmeticWithFloats(t *testing.T) {
	t.Parallel()
	val, err := Evaluate("1.5 + 2.5", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != float64(4) {
		t.Errorf("got %v, want 4", val)
	}
}

func TestEvaluate_LogicalOrBothTrue(t *testing.T) {
	t.Parallel()
	val, err := Evaluate("true || true", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != true {
		t.Errorf("got %v, want true", val)
	}
}

func TestEvaluate_ComparisonEqualFloat(t *testing.T) {
	t.Parallel()
	val, err := Evaluate("1.0 == 1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != true {
		t.Errorf("got %v, want true", val)
	}
}

func TestEvaluate_TernaryNested(t *testing.T) {
	t.Parallel()
	user := testUser{Age: 25}

	val, err := Evaluate("Age > 18 ? 1 : 0", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != int64(1) {
		t.Errorf("got %v (%T), want 1", val, val)
	}
}

func TestEvaluate_PropertyChainWithNilRoot(t *testing.T) {
	t.Parallel()
	_, err := Evaluate("Name", nil)
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestIsSimplePropertyEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"123", false},
		{"abc", true},
		{"_abc", true},
		{"a1b2", true},
		{"abc_def", true},
		{"abc.def", false},
		{"abc-def", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := isSimpleProperty(tt.input); got != tt.expected {
				t.Errorf("isSimpleProperty(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsLiteralEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"true", true},
		{"false", true},
		{"null", true},
		{"TRUE", true},
		{"123", true},
		{"abc", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := isLiteral(tt.input); got != tt.expected {
				t.Errorf("isLiteral(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
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

func TestToInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected int64
	}{
		{int(42), 42},
		{int8(8), 8},
		{int16(16), 16},
		{int32(32), 32},
		{int64(64), 64},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toInt64(tt.input); got != tt.expected {
				t.Errorf("toInt64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToUint64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected uint64
	}{
		{uint(42), 42},
		{uint8(8), 8},
		{uint16(16), 16},
		{uint32(32), 32},
		{uint64(64), 64},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toUint64(tt.input); got != tt.expected {
				t.Errorf("toUint64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToFloat64Value(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected float64
	}{
		{int(42), 42.0},
		{int8(8), 8.0},
		{int16(16), 16.0},
		{int32(32), 32.0},
		{int64(64), 64.0},
		{uint(10), 10.0},
		{uint8(8), 8.0},
		{uint16(16), 16.0},
		{uint32(32), 32.0},
		{uint64(64), 64.0},
		{float32(1.5), 1.5},
		{float64(2.5), 2.5},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toFloat64Value(tt.input); got != tt.expected {
				t.Errorf("toFloat64Value(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected float64
		ok       bool
	}{
		{"int", int(42), 42.0, true},
		{"int8", int8(8), 8.0, true},
		{"int16", int16(16), 16.0, true},
		{"int32", int32(32), 32.0, true},
		{"int64", int64(64), 64.0, true},
		{"uint", uint(10), 10.0, true},
		{"uint8", uint8(8), 8.0, true},
		{"uint16", uint16(16), 16.0, true},
		{"uint32", uint32(32), 32.0, true},
		{"uint64", uint64(64), 64.0, true},
		{"float32", float32(1.5), 1.5, true},
		{"float64", float64(2.5), 2.5, true},
		{"string", "abc", 0, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && got != tt.expected {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEqualsCrossTypeComparisons(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		left     any
		right    any
		expected bool
	}{
		{"int vs int8", int(8), int8(8), true},
		{"int vs int16", int(16), int16(16), true},
		{"int vs int32", int(32), int32(32), true},
		{"int vs int64", int(64), int64(64), true},
		{"int vs uint8", int(8), uint8(8), true},
		{"int vs uint16", int(16), uint16(16), true},
		{"int vs uint32", int(32), uint32(32), true},
		{"int vs float32", int(1), float32(1), true},
		{"int vs float64", int(1), float64(1), true},
		{"int64 vs int", int64(42), int(42), true},
		{"int64 vs int8", int64(8), int8(8), true},
		{"int64 vs int16", int64(16), int16(16), true},
		{"int64 vs int32", int64(32), int32(32), true},
		{"int64 vs uint8", int64(8), uint8(8), true},
		{"int64 vs uint16", int64(16), uint16(16), true},
		{"int64 vs uint32", int64(32), uint32(32), true},
		{"int64 vs float32", int64(1), float32(1), true},
		{"int64 vs float64", int64(1), float64(1), true},
		{"uint vs uint", uint(42), uint(42), true},
		{"uint vs uint8", uint(8), uint8(8), true},
		{"uint vs uint16", uint(16), uint16(16), true},
		{"uint vs uint32", uint(32), uint32(32), true},
		{"uint vs uint64", uint(64), uint64(64), true},
		{"uint vs float32", uint(1), float32(1), true},
		{"uint vs float64", uint(1), float64(1), true},
		{"uint64 vs int8", uint64(8), int8(8), true},
		{"uint64 vs int16", uint64(16), int16(16), true},
		{"uint64 vs int32", uint64(32), int32(32), true},
		{"uint64 vs uint", uint64(42), uint(42), true},
		{"uint64 vs uint8", uint64(8), uint8(8), true},
		{"uint64 vs uint16", uint64(16), uint16(16), true},
		{"uint64 vs uint32", uint64(32), uint32(32), true},
		{"uint64 vs uint64", uint64(64), uint64(64), true},
		{"uint64 vs float32", uint64(1), float32(1), true},
		{"uint64 vs float64", uint64(1), float64(1), true},
		{"float32 vs int", float32(42), int(42), true},
		{"float32 vs int8", float32(8), int8(8), true},
		{"float32 vs int16", float32(16), int16(16), true},
		{"float32 vs int32", float32(32), int32(32), true},
		{"float32 vs int64", float32(64), int64(64), true},
		{"float32 vs uint", float32(42), uint(42), true},
		{"float32 vs uint8", float32(8), uint8(8), true},
		{"float32 vs uint16", float32(16), uint16(16), true},
		{"float32 vs uint32", float32(32), uint32(32), true},
		{"float32 vs uint64", float32(64), uint64(64), true},
		{"float32 vs float64", float32(1.5), float64(1.5), true},
		{"float64 vs int", float64(42), int(42), true},
		{"float64 vs int8", float64(8), int8(8), true},
		{"float64 vs int16", float64(16), int16(16), true},
		{"float64 vs int32", float64(32), int32(32), true},
		{"float64 vs int64", float64(64), int64(64), true},
		{"float64 vs uint", float64(42), uint(42), true},
		{"float64 vs uint8", float64(8), uint8(8), true},
		{"float64 vs uint16", float64(16), uint16(16), true},
		{"float64 vs uint32", float64(32), uint32(32), true},
		{"float64 vs uint64", float64(64), uint64(64), true},
		{"float64 vs float32", float64(1.5), float32(1.5), true},
		{"int8 vs int", int8(42), int(42), true},
		{"int16 vs int", int16(42), int(42), true},
		{"int32 vs int", int32(42), int(42), true},
		{"uint8 vs uint", uint8(42), uint(42), true},
		{"uint16 vs uint", uint16(42), uint(42), true},
		{"uint32 vs uint", uint32(42), uint(42), true},
		{"float64 vs string", float64(1), "a", false},
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

func TestCompareValuesUnsupported(t *testing.T) {
	t.Parallel()
	_, err := compareValues("a", "b", "~")
	if err == nil {
		t.Error("expected error for unsupported operator")
	}
}

func TestEvaluateArithmeticUnsupported(t *testing.T) {
	t.Parallel()
	_, err := arithmetic("a", "b", "~")
	if err == nil {
		t.Error("expected error for unsupported arithmetic operator")
	}
}
