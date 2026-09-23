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
	joined := joinConditions([]Condition{OnProperty("a"), OnProperty("b")}, ", ")
	expected := "OnProperty(a), OnProperty(b)"
	if joined != expected {
		t.Errorf("joinConditions() = %q, want %q", joined, expected)
	}
}

func TestJoinConditions_Empty(t *testing.T) {
	t.Parallel()
	joined := joinConditions(nil, ", ")
	if joined != "" {
		t.Errorf("expected empty string, got %q", joined)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	b := New()
	if b == nil {
		t.Fatal("New() returned nil")
	}
}

func TestConditionBuilder_BuildEmpty(t *testing.T) {
	t.Parallel()
	b := New()
	c := b.Build()
	if c == nil {
		t.Fatal("Build() returned nil")
	}
	// Empty builder returns alwaysTrue
	if !c.Matches(nil) {
		t.Error("empty Build() should match")
	}
	if c.String() != "AlwaysTrue" {
		t.Errorf("String() = %q, want %q", c.String(), "AlwaysTrue")
	}
}

func TestConditionBuilder_SingleCondition(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "1", true
			}
			return nil, false
		},
	}
	b := New()
	b.OnProperty("a")
	c := b.Build()
	if !c.Matches(ctx) {
		t.Error("single OnProperty should match")
	}
}

func TestConditionBuilder_And(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" || key == "b" {
				return "1", true
			}
			return nil, false
		},
	}
	b := New()
	b.OnProperty("a").And().OnProperty("b")
	c := b.Build()
	if !c.Matches(ctx) {
		t.Error("AND(a,b) should match when both exist")
	}

	// Missing one
	ctx2 := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "1", true
			}
			return nil, false
		},
	}
	if c.Matches(ctx2) {
		t.Error("AND(a,b) should not match when b is missing")
	}
}

func TestConditionBuilder_Or(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "1", true
			}
			return nil, false
		},
	}
	b := New()
	b.OnProperty("a").Or().OnProperty("b")
	c := b.Build()
	if !c.Matches(ctx) {
		t.Error("OR(a,b) should match when a exists")
	}
}

func TestConditionBuilder_Not(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "1", true
			}
			return nil, false
		},
	}
	b := New()
	b.OnProperty("a").And().Not().OnProperty("b")
	c := b.Build()
	// a=true, b=false → NOT(b)=true → All(a, NOT(b)) = true
	if !c.Matches(ctx) {
		t.Error("a AND NOT(b) should match when a exists and b is missing")
	}

	ctx2 := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" || key == "b" {
				return "1", true
			}
			return nil, false
		},
	}
	// a=true, b=true → NOT(b)=false → All(a, NOT(b)) = false
	if c.Matches(ctx2) {
		t.Error("a AND NOT(b) should not match when both a and b exist")
	}
}

func TestConditionBuilder_All(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			return "1", true
		},
	}
	b := New()
	b.All(OnProperty("a"), OnProperty("b"))
	c := b.Build()
	if !c.Matches(ctx) {
		t.Error("All(a,b) should match")
	}
}

func TestConditionBuilder_Any(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			return nil, false
		},
	}
	b := New()
	b.Any(OnProperty("a"), OnProperty("b"))
	c := b.Build()
	if c.Matches(ctx) {
		t.Error("Any(a,b) should not match when neither exists")
	}
}

func TestConditionBuilder_AllMethodsExist(t *testing.T) {
	t.Parallel()
	b := New()
	b.OnMissingProperty("missing").
		OnBean("db").
		OnMissingBean("cache").
		OnProfile("dev").
		OnModuleLoaded("mod1").
		OnMissingModule("mod2").
		OnExpression("1==1").
		OnResourceExists("file:go.mod").
		OnResourceMissing("file:nonexistent_file").
		OnEnvVarExists("PATH").
		OnEnvVarMissing("UNLIKELY_ENV_VAR_FOR_TEST_12345")
	c := b.Build()
	if c == nil {
		t.Fatal("Build() should not return nil after chaining all methods")
	}
}

func TestConditionBuilder_MixedAndOr(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" || key == "c" {
				return "1", true
			}
			return nil, false
		},
	}
	b := New()
	b.OnProperty("a").And().OnProperty("b").Or().OnProperty("c")
	c := b.Build()
	// Should be: (a AND b) OR c → true because c exists
	if !c.Matches(ctx) {
		t.Error("mixed AND/OR should match")
	}
}

func TestConditionBuilder_AllWith(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			return nil, false
		},
	}
	g := AllWith(OnProperty("a"), OnProperty("b"))
	c := g.Build()
	if c.Matches(ctx) {
		t.Error("AllWith(a,b) should not match when neither exists")
	}
}

func TestConditionBuilderGroup_Or(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "1", true
			}
			return nil, false
		},
	}
	g := AllWith(OnProperty("a")).Or(OnProperty("b"))
	c := g.Build()
	if !c.Matches(ctx) {
		t.Error("OR group should match")
	}
}

func TestConditionBuilderGroup_And(t *testing.T) {
	t.Parallel()
	g := AllWith(OnProperty("a")).And(OnProperty("b"))
	c := g.Build()
	if c == nil {
		t.Fatal("Build() should not return nil")
	}
}

func TestConditionBuilderGroup_Empty(t *testing.T) {
	t.Parallel()
	g := ConditionBuilderGroup{}
	c := g.Build()
	if c == nil {
		t.Fatal("empty group Build() should not return nil")
	}
	if !c.Matches(nil) {
		t.Error("empty group should always match")
	}
}

func TestConditionBuilderGroup_SingleGroup(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "x" {
				return "1", true
			}
			return nil, false
		},
	}
	g := AllWith(OnProperty("x"))
	c := g.Build()
	if !c.Matches(ctx) {
		t.Error("single group should match")
	}
}
