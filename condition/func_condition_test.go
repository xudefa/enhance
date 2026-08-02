package condition

import (
	"testing"
)

func TestConditionFunc(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "enabled" {
				return "true", true
			}
			return nil, false
		},
	}

	cond := ConditionFunc(func(ctx ConditionContext) bool {
		val, ok := ctx.GetProperty("enabled")
		return ok && val == "true"
	})
	if !cond.Matches(ctx) {
		t.Fatal("ConditionFunc should match")
	}
	if cond.String() != "ConditionFunc(...)" {
		t.Errorf("expected 'ConditionFunc(...)', got %s", cond.String())
	}
}

func TestAlways(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(string) (any, bool) { return nil, false },
	}
	if !Always().Matches(ctx) {
		t.Fatal("Always should always match")
	}
}

func TestNever(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(string) (any, bool) { return "x", true },
	}
	if Never().Matches(ctx) {
		t.Fatal("Never should never match")
	}
}

func TestWhen(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "flag" {
				return "on", true
			}
			return nil, false
		},
	}

	cond := When("flag is on", func(ctx ConditionContext) bool {
		val, ok := ctx.GetProperty("flag")
		return ok && val == "on"
	})
	if !cond.Matches(ctx) {
		t.Fatal("When should match")
	}
	expected := "When(flag is on)"
	if cond.String() != expected {
		t.Errorf("expected %s, got %s", expected, cond.String())
	}
}

func TestFunc(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(string) (any, bool) { return "v", true },
	}
	cond := Func(func(ctx ConditionContext) bool {
		_, ok := ctx.GetProperty("any")
		return ok
	})
	if !cond.Matches(ctx) {
		t.Fatal("Func should match")
	}
}

func TestAllFunc(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			switch key {
			case "a", "b":
				return "1", true
			default:
				return nil, false
			}
		},
	}
	trueFn := func(ctx ConditionContext) bool { _, ok := ctx.GetProperty("a"); return ok }
	falseFn := func(ctx ConditionContext) bool { _, ok := ctx.GetProperty("missing"); return ok }

	if !AllFunc(trueFn, trueFn).Matches(ctx) {
		t.Fatal("AllFunc(true, true) should match")
	}
	if AllFunc(trueFn, falseFn).Matches(ctx) {
		t.Fatal("AllFunc(true, false) should not match")
	}
	if AllFunc(falseFn, falseFn).Matches(ctx) {
		t.Fatal("AllFunc(false, false) should not match")
	}
}

func TestAnyFunc(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "1", true
			}
			return nil, false
		},
	}
	trueFn := func(ctx ConditionContext) bool { _, ok := ctx.GetProperty("a"); return ok }
	falseFn := func(ctx ConditionContext) bool { _, ok := ctx.GetProperty("missing"); return ok }

	if !AnyFunc(trueFn, falseFn).Matches(ctx) {
		t.Fatal("AnyFunc(true, false) should match")
	}
	if AnyFunc(falseFn, falseFn).Matches(ctx) {
		t.Fatal("AnyFunc(false, false) should not match")
	}
}

func TestCustom(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "ok" {
				return true, true
			}
			return nil, false
		},
	}
	cond := Custom("custom-check", func(ctx ConditionContext) bool {
		val, ok := ctx.GetProperty("ok")
		return ok && val == true
	})
	if !cond.Matches(ctx) {
		t.Fatal("Custom should match")
	}
	if cond.String() != "Custom(custom-check)" {
		t.Errorf("expected 'Custom(custom-check)', got %s", cond.String())
	}
}

func TestValAsString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected string
	}{
		{"hello", "hello"},
		{true, "true"},
		{false, "false"},
		{42, "42"},
		{3.14, "3.14"},
		{nil, ""},
		{struct{}{}, ""},
	}
	for _, tt := range tests {
		got := valAsString(tt.input)
		if got != tt.expected {
			t.Errorf("valAsString(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestAllWithOptions(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "k" {
				return "v", true
			}
			return nil, false
		},
	}
	factory := AllWithOptions(WithDescription("custom-all"))
	cond := factory(OnProperty("k"))
	if !cond.Matches(ctx) {
		t.Fatal("AllWithOptions should match")
	}
	if cond.String() != "All(custom-all)" {
		t.Errorf("expected 'All(custom-all)', got %s", cond.String())
	}
}

func TestCompositeConditionStrings(t *testing.T) {
	t.Parallel()
	p1 := OnProperty("x")
	p2 := OnProperty("y")
	all := All(p1, p2)
	any := Any(p1, p2)
	not := Not(p1)

	if all.String() != "All(OnProperty(x), OnProperty(y))" {
		t.Errorf("unexpected All string: %s", all.String())
	}
	if any.String() != "Any(OnProperty(x), OnProperty(y))" {
		t.Errorf("unexpected Any string: %s", any.String())
	}
	if not.String() != "Not(OnProperty(x))" {
		t.Errorf("unexpected Not string: %s", not.String())
	}
}

func TestBuiltinConditionStrings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		condition Condition
		expected  string
	}{
		{OnProperty("k"), "OnProperty(k)"},
		{OnProperty("k", "v"), "OnProperty(k=v)"},
		{OnMissingProperty("k"), "OnMissingProperty(k)"},
		{OnBean("b"), "OnBean(b)"},
		{OnMissingBean("b"), "OnMissingBean(b)"},
		{OnProfile("dev"), "OnProfile(dev)"},
		{OnModuleLoaded("m"), "OnModuleLoaded(m)"},
		{OnMissingModule("m"), "OnMissingModule(m)"},
	}
	for _, tt := range tests {
		if tt.condition.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.condition.String())
		}
	}
}

func TestOnPropertyWithNonStringValues(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			switch key {
			case "int.key":
				return 8080, true
			case "bool.key":
				return true, true
			case "float.key":
				return 3.14, true
			default:
				return nil, false
			}
		},
	}

	if !OnProperty("int.key", "8080").Matches(ctx) {
		t.Fatal("OnProperty should match int value converted to string")
	}
	if !OnProperty("bool.key", "true").Matches(ctx) {
		t.Fatal("OnProperty should match bool value converted to string")
	}
	if !OnProperty("float.key", "3.14").Matches(ctx) {
		t.Fatal("OnProperty should match float value converted to string")
	}
}

func TestOnPropertyEmptyStringValue(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "empty" {
				return "", true
			}
			return nil, false
		},
	}

	cond := OnProperty("empty")
	if cond.Matches(ctx) {
		t.Fatal("OnProperty without expectedValue should not match empty string")
	}
}

func TestCompositeWithSingleCondition(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "v", true
			}
			return nil, false
		},
	}

	if !All(OnProperty("a")).Matches(ctx) {
		t.Fatal("All with single condition should match")
	}
	if !Any(OnProperty("a")).Matches(ctx) {
		t.Fatal("Any with single condition should match")
	}
}

func TestBuilderWithMultipleOperators(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			switch key {
			case "a":
				return "1", true
			case "b":
				return "2", true
			default:
				return nil, false
			}
		},
	}

	b := New().
		OnProperty("a").
		Or().
		OnProperty("b").
		Or().
		OnProperty("missing").
		Build()
	if !b.Matches(ctx) {
		t.Fatal("OR chain should match when first condition matches")
	}

	b2 := New().
		OnProperty("missing").
		Or().
		OnProperty("missing2").
		Or().
		OnProperty("a").
		Build()
	if !b2.Matches(ctx) {
		t.Fatal("OR chain should match when last condition matches")
	}
}

func TestBuilderNotWithOr(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			if key == "a" {
				return "v", true
			}
			return nil, false
		},
	}

	b := New().
		OnProperty("missing").
		Or().
		Not().
		OnProperty("missing2").
		Build()
	if !b.Matches(ctx) {
		t.Fatal("NOT with OR should match when Not condition is true")
	}
}

func TestBuilderNotNegatesNextCondition(t *testing.T) {
	t.Parallel()

	mkCtx := func(keys ...string) ConditionContext {
		return &mockConditionContext{
			envFn: func(k string) (any, bool) {
				for _, key := range keys {
					if k == key {
						return "v", true
					}
				}
				return nil, false
			},
		}
	}

	// A().Not().B() 语义应为 All(A, Not(B))：
	// - A、B 都存在时 → Not(B) 为 false → 整体 false
	// - A 存在、B 缺失时 → Not(B) 为 true → 整体 true
	bothPresent := New().
		OnProperty("a").
		Not().
		OnProperty("b").
		Build()
	if bothPresent.Matches(mkCtx("a", "b")) {
		t.Fatal("expected (a AND NOT b) to be false when a and b both present")
	}
	if b := New().OnProperty("a").Not().OnProperty("b").Build(); !b.Matches(mkCtx("a")) {
		t.Fatal("expected (a AND NOT b) to be true when b is missing")
	}

	// Not() 紧邻其后条件：Not().A() 语义为 Not(A)
	b := New().
		Not().
		OnProperty("a").
		Build()
	if b.Matches(mkCtx("a")) {
		t.Fatal("expected Not(a) to be false when a present")
	}
	if !b.Matches(mkCtx()) {
		t.Fatal("expected Not(a) to be true when a missing")
	}
}
