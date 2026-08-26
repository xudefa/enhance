package condition

import (
	"testing"
)

func TestAll_Empty(t *testing.T) {
	t.Parallel()
	c := All()
	if !c.Matches(nil) {
		t.Error("All() empty should match")
	}
	if got := c.String(); got != "All()" {
		t.Errorf("String() = %q", got)
	}
}

func TestAll_WithDescription(t *testing.T) {
	t.Parallel()
	f := AllWithOptions(WithDescription("custom-desc"))
	c := f(OnProperty("x"))
	if got := c.String(); got != "All(custom-desc)" {
		t.Errorf("String() = %q, want %q", got, "All(custom-desc)")
	}
}

func TestAll_WithLazyEvaluation(t *testing.T) {
	t.Parallel()
	f := AllWithOptions(WithLazyEvaluation(), WithDescription("lazy"))
	c := f(OnProperty("x"), OnProperty("y"))
	if got := c.String(); got != "All(lazy)" {
		t.Errorf("String() = %q", got)
	}
}

func TestAny_Empty(t *testing.T) {
	t.Parallel()
	c := Any()
	if c.Matches(nil) {
		t.Error("Any() empty should not match")
	}
	if got := c.String(); got != "Any()" {
		t.Errorf("String() = %q", got)
	}
}

func TestAny_WithDescription(t *testing.T) {
	t.Parallel()
	f := AllWithOptions(WithDescription("any-desc"))
	c := f(OnProperty("x"))
	if got := c.String(); got != "All(any-desc)" {
		t.Errorf("String() = %q, want %q", got, "All(any-desc)")
	}
}

func TestNot_Empty(t *testing.T) {
	t.Parallel()
	c := Not(All())
	if c.Matches(nil) {
		t.Error("Not(All()) should not match (inverts true to false)")
	}
	if got := c.String(); got != "Not(All())" {
		t.Errorf("String() = %q", got)
	}
}

func TestAll_ShortCircuit(t *testing.T) {
	t.Parallel()
	callCount := 0
	trueCond := ConditionFunc(func(ctx ConditionContext) bool {
		return true
	})
	falseCond := ConditionFunc(func(ctx ConditionContext) bool {
		callCount++
		return false
	})
	// All: falseCond first, trueCond second
	// falseCond returns false → All should short-circuit, trueCond not called
	c := All(falseCond, trueCond)
	if c.Matches(nil) {
		t.Error("All should not match")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d (short-circuit failed)", callCount)
	}
}

func TestAny_ShortCircuit(t *testing.T) {
	t.Parallel()
	callCount := 0
	trueCond := ConditionFunc(func(ctx ConditionContext) bool {
		callCount++
		return true
	})
	falseCond := ConditionFunc(func(ctx ConditionContext) bool {
		callCount++
		return false
	})
	// Any: trueCond first → short-circuits, falseCond not called
	c := Any(trueCond, falseCond)
	if !c.Matches(nil) {
		t.Error("Any should match")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d (short-circuit failed)", callCount)
	}
}

func TestAll_SingleCondition(t *testing.T) {
	t.Parallel()
	c := All(OnProperty("x"))
	if got := c.String(); got != "All(OnProperty(x))" {
		t.Errorf("String() = %q", got)
	}
}

func TestAny_SingleCondition(t *testing.T) {
	t.Parallel()
	c := Any(OnProperty("x"))
	if got := c.String(); got != "Any(OnProperty(x))" {
		t.Errorf("String() = %q", got)
	}
}

func TestJoinConditions(t *testing.T) {
	t.Parallel()
	result := joinConditions([]Condition{OnProperty("a"), OnProperty("b")}, ", ")
	expected := "OnProperty(a), OnProperty(b)"
	if result != expected {
		t.Errorf("joinConditions() = %q, want %q", result, expected)
	}
}

func TestJoinConditions_Empty(t *testing.T) {
	t.Parallel()
	result := joinConditions(nil, ", ")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
