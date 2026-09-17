package spel

import (
	"testing"
)

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

func TestPropertyExpression_GetValue_WithVariable(t *testing.T) {
	t.Parallel()
	expr := &propertyExpressionImpl{property: "myVar"}
	ctx := NewStandardEvaluationContext(nil)
	ctx.SetVariable("myVar", 42)

	got, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Errorf("got %v, want 42", got)
	}
}

func TestPropertyExpression_GetValue_WithTag(t *testing.T) {
	t.Parallel()
	type TaggedUser struct {
		FullName string `json:"full_name"`
	}

	accessor := NewReflectPropertyAccessor()
	u := TaggedUser{FullName: "Alice"}

	got, err := accessor.GetProperty(u, "full_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Alice" {
		t.Errorf("got %v, want 'Alice'", got)
	}
}
