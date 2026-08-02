package spel

import (
	"strings"
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

	val, err := Evaluate("Name", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "Alice" {
		t.Errorf("expected 'Alice', got %v", val)
	}
}

func TestEvaluate_Literal_String(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("'hello world'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "hello world" {
		t.Errorf("expected 'hello world', got %v", val)
	}
}

func TestEvaluate_Literal_Number(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != int64(42) {
		t.Errorf("expected 42, got %v", val)
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
			val, err := expr.GetValue(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if val != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, val)
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
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != nil {
		t.Errorf("expected nil, got %v", val)
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
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != true {
		t.Errorf("expected true, got %v", val)
	}
}

func TestEvaluate_Arithmetic(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("10 + 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != float64(15) {
		t.Errorf("expected 15, got %v", val)
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
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != true {
		t.Errorf("expected true, got %v", val)
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
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "adult" {
		t.Errorf("expected 'adult', got %v", val)
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
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "Beijing" {
		t.Errorf("expected 'Beijing', got %v", val)
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
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "Hello, Alice" {
		t.Errorf("expected 'Hello, Alice', got %v", val)
	}
}

func TestStandardEvaluationContext_Variables(t *testing.T) {
	t.Parallel()
	ctx := NewStandardEvaluationContext(nil)

	ctx.SetVariable("name", "Alice")
	val, ok := ctx.GetVariable("name")
	if !ok {
		t.Fatal("variable not found")
	}

	if val != "Alice" {
		t.Errorf("expected 'Alice', got %v", val)
	}
}

func TestStandardEvaluationContext_RootObject(t *testing.T) {
	t.Parallel()
	ctx := NewStandardEvaluationContext("root")

	if ctx.GetRootObject() != "root" {
		t.Errorf("expected 'root', got %v", ctx.GetRootObject())
	}

	ctx.SetRootObject("newRoot")
	if ctx.GetRootObject() != "newRoot" {
		t.Errorf("expected 'newRoot', got %v", ctx.GetRootObject())
	}
}

func TestReflectPropertyAccessor_GetProperty(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := testUser{Name: "Bob"}

	val, err := accessor.GetProperty(user, "Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "Bob" {
		t.Errorf("expected 'Bob', got %v", val)
	}
}

func TestReflectPropertyAccessor_GetProperty_Nil(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()

	_, err := accessor.GetProperty(nil, "Name")
	if err == nil {
		t.Error("expected error for nil target")
	}
}

func TestReflectPropertyAccessor_SetProperty(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := &testUser{Name: "Old"}

	err := accessor.SetProperty(user, "Name", "New")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Name != "New" {
		t.Errorf("expected 'New', got %v", user.Name)
	}
}

func TestExpression_SetValue(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	user := &testUser{Name: "Old"}

	ctx := NewStandardEvaluationContext(user)
	err = expr.SetValue(ctx, "New")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Name != "New" {
		t.Errorf("expected 'New', got %v", user.Name)
	}
}

func TestComplexExpression_DivisionByZero(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("10 / 0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(nil)
	_, err = expr.GetValue(ctx)
	if err == nil {
		t.Error("expected division by zero error")
	}
}

func TestIsTruthy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"nil", nil, false},
		{"true", true, true},
		{"false", false, false},
		{"zero int", 0, false},
		{"non-zero int", 1, true},
		{"empty string", "", false},
		{"non-empty string", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTruthy(tt.input); got != tt.expected {
				t.Errorf("isTruthy(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"42", 42, false},
		{"-10", -10, false},
		{"0", 0, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseInt(tt.input, 10, 64)
			if tt.hasError {
				if err == nil {
					t.Error("expected error")
				}
			} else {
				if got != tt.expected {
					t.Errorf("expected %d, got %d", tt.expected, got)
				}
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"3.14", 3.14, false},
		{"-2.5", -2.5, false},
		{"42", 42.0, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseFloat(tt.input, 64)
			if tt.hasError {
				if err == nil {
					t.Error("expected error")
				}
			} else {
				if got != tt.expected {
					t.Errorf("expected %f, got %f", tt.expected, got)
				}
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"42", true},
		{"3.14", true},
		{"-10", true},
		{"abc", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isNumber(tt.input); got != tt.expected {
				t.Errorf("isNumber(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExpression_Evaluate_Fallback(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("unknownVar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := NewStandardEvaluationContext(map[string]any{})

	_, err = expr.GetValue(ctx)
	if err == nil {
		t.Fatal("expected error for unknown variable")
	}
	// 错误消息应包含 unknown、unable to evaluate 或 non-struct type
	if !strings.Contains(err.Error(), "unknown") &&
		!strings.Contains(err.Error(), "unable to evaluate") &&
		!strings.Contains(err.Error(), "non-struct type") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExpression_SetValue_NotSupported(t *testing.T) {
	t.Parallel()
	expr, err := ParseExpression("a + b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := NewStandardEvaluationContext(nil)

	err = expr.SetValue(ctx, 42)
	if err == nil {
		t.Error("expected error for setting value on complex expression")
	}
}

func TestGlobalSpelParser(t *testing.T) {
	t.Parallel()
	if GlobalSpelParser == nil {
		t.Fatal("GlobalSpelParser should not be nil")
	}

	expr, err := GlobalSpelParser.ParseExpression("name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if expr.String() != "name" {
		t.Errorf("expected 'name', got %v", expr.String())
	}
}

func TestEvaluate_Convenience(t *testing.T) {
	t.Parallel()
	user := testUser{Name: "Test"}

	val, err := Evaluate("Name", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != "Test" {
		t.Errorf("expected 'Test', got %v", val)
	}
}
