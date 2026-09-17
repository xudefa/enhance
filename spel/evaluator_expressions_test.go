package spel

import (
	"testing"
)

func TestEvaluateMethodCall_Simple(t *testing.T) {
	t.Parallel()

	user := &testUser{Name: "Alice"}
	got, err := Evaluate("Greet()", user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello, Alice" {
		t.Errorf("Greet() = %v, want 'Hello, Alice'", got)
	}
}

func TestEvaluateMethodCall_MissingClosingParen(t *testing.T) {
	t.Parallel()

	expr, _ := ParseExpression("Greet(")
	_, err := expr.GetValue(NewStandardEvaluationContext(&testUser{}))
	if err == nil {
		t.Error("expected error for missing closing paren")
	}
}

func TestEvaluateMethodCall_NoMethod(t *testing.T) {
	t.Parallel()

	expr, _ := ParseExpression("NonExistent()")
	_, err := expr.GetValue(NewStandardEvaluationContext(&testUser{}))
	if err == nil {
		t.Error("expected error for non-existent method")
	}
}

func TestEvaluateMethodCall_NilRootObject(t *testing.T) {
	t.Parallel()

	expr, _ := ParseExpression("Greet()")
	_, err := expr.GetValue(NewStandardEvaluationContext(nil))
	if err == nil {
		t.Error("expected error when root object is nil")
	}
}

func TestEvaluatePropertyChain_SingleLevel(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
	}
	p := Person{Name: "Bob"}

	got, err := Evaluate("Name", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Bob" {
		t.Errorf("Name = %v, want 'Bob'", got)
	}
}

func TestEvaluatePropertyChain_MultiLevel(t *testing.T) {
	t.Parallel()

	type Address struct{ City string }
	type Person struct{ Address Address }

	p := Person{Address: Address{City: "Beijing"}}

	got, err := Evaluate("Address.City", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Beijing" {
		t.Errorf("Address.City = %v, want 'Beijing'", got)
	}
}

func TestEvaluatePropertyChain_NilIntermediate(t *testing.T) {
	t.Parallel()

	type Address struct{ City string }
	type Person struct{ Address *Address }

	p := Person{Address: nil}

	_, err := Evaluate("Address.City", p)
	if err == nil {
		t.Error("expected error for nil intermediate property")
	}
}

func TestEvaluateLiteral_String(t *testing.T) {
	t.Parallel()

	got, err := Evaluate("'hello'", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("string literal = %v, want 'hello'", got)
	}
}

func TestEvaluateLiteral_True(t *testing.T) {
	t.Parallel()

	got, err := Evaluate("true", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Errorf("true literal = %v, want true", got)
	}
}

func TestEvaluateLiteral_False(t *testing.T) {
	t.Parallel()

	got, err := Evaluate("false", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != false {
		t.Errorf("false literal = %v, want false", got)
	}
}

func TestEvaluateLiteral_Null(t *testing.T) {
	t.Parallel()

	got, err := Evaluate("null", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("null literal = %v, want nil", got)
	}
}

func TestEvaluateLiteral_Integer(t *testing.T) {
	t.Parallel()

	got, err := Evaluate("42", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != int64(42) {
		t.Errorf("integer literal = %v (%T), want 42", got, got)
	}
}

func TestEvaluateLiteral_NegativeInteger(t *testing.T) {
	t.Parallel()

	got, err := Evaluate("-7", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != int64(-7) {
		t.Errorf("negative integer literal = %v, want -7", got)
	}
}

func TestEvaluateLiteral_Unknown(t *testing.T) {
	t.Parallel()

	_, err := Evaluate("unknown_literal", nil)
	if err == nil {
		t.Error("expected error for unknown literal")
	}
}

func TestCompareValues_Equal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		left  any
		right any
		want  bool
	}{
		{"both nil", nil, nil, true},
		{"nil and string", nil, "a", false},
		{"strings equal", "hello", "hello", true},
		{"strings not equal", "hello", "world", false},
		{"ints equal", 42, 42, true},
		{"ints not equal", 42, 43, false},
		{"bools equal", true, true, true},
		{"bools not equal", true, false, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := compareValues(tt.left, tt.right, "==")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("compareValues(%v, %v, '==') = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestCompareValues_NotEqual(t *testing.T) {
	t.Parallel()

	got, err := compareValues("a", "b", "!=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true for 'a' != 'b'")
	}
}

func TestCompareValues_Ordering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		left  any
		right any
		op    string
		want  bool
	}{
		{"5 > 3", 5, 3, ">", true},
		{"3 > 5", 3, 5, ">", false},
		{"3 < 5", 3, 5, "<", true},
		{"5 < 3", 5, 3, "<", false},
		{"5 >= 5", 5, 5, ">=", true},
		{"5 >= 6", 5, 6, ">=", false},
		{"5 <= 5", 5, 5, "<=", true},
		{"6 <= 5", 6, 5, "<=", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := compareValues(tt.left, tt.right, tt.op)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("compareValues(%v, %v, %q) = %v, want %v", tt.left, tt.right, tt.op, got, tt.want)
			}
		})
	}
}

func TestCompareValues_UnsupportedOp(t *testing.T) {
	t.Parallel()

	_, err := compareValues(1, 2, "~")
	if err == nil {
		t.Error("expected error for unsupported operator")
	}
}

func TestCompareValues_NonNumericOrdering(t *testing.T) {
	t.Parallel()

	_, err := compareValues("a", "b", ">")
	if err == nil {
		t.Error("expected error for non-numeric comparison")
	}
}
