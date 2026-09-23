package condition

import (
	"os"
	"testing"
)

// TestOnExpression 验证 OnExpression 条件的行为:
//  1. 简单比较表达式
//  2. 属性占位符表达式
//  3. 逻辑运算表达式
//  4. 表达式解析失败时返回 false
func TestOnExpression(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			switch key {
			case "server.port":
				return 9090, true
			case "app.env":
				return "prod", true
			case "debug.enabled":
				return false, true
			default:
				return nil, false
			}
		},
	}

	// 测试简单比较表达式
	c1 := OnExpression("9090 > 8080")
	if !c1.Matches(ctx) {
		t.Fatal("expected expression '9090 > 8080' to match")
	}

	// 测试属性占位符表达式
	c2 := OnExpression("${server.port} > 8080")
	if !c2.Matches(ctx) {
		t.Fatal("expected expression '${server.port} > 8080' to match")
	}

	// 测试不匹配的表达式
	c3 := OnExpression("${server.port} < 8080")
	if c3.Matches(ctx) {
		t.Fatal("expected expression '${server.port} < 8080' to not match")
	}

	// 测试字符串比较
	c4 := OnExpression("${app.env} == 'prod'")
	if !c4.Matches(ctx) {
		t.Fatal("expected expression '${app.env} == 'prod'' to match")
	}

	// 测试逻辑与运算
	c5 := OnExpression("${app.env} == 'prod' && ${debug.enabled} == false")
	if !c5.Matches(ctx) {
		t.Fatal("expected logical AND expression to match")
	}
}

// TestOnResourceExists 验证 OnResourceExists 条件的行为:
//  1. classpath: 前缀资源
//  2. file: 前缀资源
//  3. 无前缀资源（默认 classpath）
//  4. 资源不存在时不匹配
func TestOnResourceExists(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{}

	// 测试存在的文件（使用当前目录的 README.md）
	c1 := OnResourceExists("file:README.md")
	if !c1.Matches(ctx) {
		t.Fatal("expected OnResourceExists('file:README.md') to match")
	}

	// 测试不存在的文件
	c2 := OnResourceExists("file:/nonexistent/file.txt")
	if c2.Matches(ctx) {
		t.Fatal("expected OnResourceExists('file:/nonexistent/file.txt') to not match")
	}

	// 测试 classpath: 前缀
	c3 := OnResourceExists("classpath:README.md")
	if !c3.Matches(ctx) {
		t.Fatal("expected OnResourceExists('classpath:README.md') to match")
	}

	// 测试无前缀（默认 classpath）
	c4 := OnResourceExists("README.md")
	if !c4.Matches(ctx) {
		t.Fatal("expected OnResourceExists('README.md') to match")
	}
}

// TestOnResourceMissing 验证 OnResourceMissing 条件的行为:
//  1. 资源不存在时匹配
//  2. 资源存在时不匹配
func TestOnResourceMissing(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{}

	c1 := OnResourceMissing("file:/nonexistent/file.txt")
	if !c1.Matches(ctx) {
		t.Fatal("expected OnResourceMissing to match for nonexistent file")
	}

	c2 := OnResourceMissing("file:README.md")
	if c2.Matches(ctx) {
		t.Fatal("expected OnResourceMissing to not match for existing file")
	}
}

// TestOnEnvVarExists 验证 OnEnvVarExists 条件的行为:
//  1. 环境变量存在时匹配
//  2. 环境变量不存在时不匹配
func TestOnEnvVarExists(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{}

	// 设置测试环境变量
	_ = os.Setenv("TEST_ENV_VAR", "test_value")
	defer func() { _ = os.Unsetenv("TEST_ENV_VAR") }()

	c1 := OnEnvVarExists("TEST_ENV_VAR")
	if !c1.Matches(ctx) {
		t.Fatal("expected OnEnvVarExists to match for existing env var")
	}

	c2 := OnEnvVarExists("NONEXISTENT_ENV_VAR_XYZ")
	if c2.Matches(ctx) {
		t.Fatal("expected OnEnvVarExists to not match for nonexistent env var")
	}
}

// TestOnEnvVarMissing 验证 OnEnvVarMissing 条件的行为:
//  1. 环境变量不存在时匹配
//  2. 环境变量存在时不匹配
func TestOnEnvVarMissing(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{}

	_ = os.Setenv("TEST_ENV_VAR_2", "test_value")
	defer func() { _ = os.Unsetenv("TEST_ENV_VAR_2") }()

	c1 := OnEnvVarMissing("NONEXISTENT_ENV_VAR_XYZ")
	if !c1.Matches(ctx) {
		t.Fatal("expected OnEnvVarMissing to match for nonexistent env var")
	}

	c2 := OnEnvVarMissing("TEST_ENV_VAR_2")
	if c2.Matches(ctx) {
		t.Fatal("expected OnEnvVarMissing to not match for existing env var")
	}
}

func testConditionBuilderSingle(t *testing.T, ctx ConditionContext) {
	t.Helper()
	b1 := New().OnProperty("feature.enabled", "true").Build()
	if !b1.Matches(ctx) {
		t.Fatal("expected single condition to match")
	}
}

func testConditionBuilderAnd(t *testing.T, ctx ConditionContext) {
	t.Helper()
	b2 := New().
		OnProperty("feature.enabled", "true").
		And().
		OnBean("dataSource").
		Build()
	if !b2.Matches(ctx) {
		t.Fatal("expected AND condition to match")
	}
}

func testConditionBuilderOr(t *testing.T, ctx ConditionContext) {
	t.Helper()
	b3 := New().
		OnProperty("nonexistent").
		Or().
		OnProperty("feature.enabled").
		Build()
	if !b3.Matches(ctx) {
		t.Fatal("expected OR condition to match when one matches")
	}
}

func testConditionBuilderNot(t *testing.T, ctx ConditionContext) {
	t.Helper()
	b4 := New().
		Not().
		OnProperty("nonexistent").
		Build()
	if !b4.Matches(ctx) {
		t.Fatal("expected NOT condition to match when condition is false")
	}
}

func testConditionBuilderMixed(t *testing.T, ctx ConditionContext) {
	t.Helper()
	b5 := New().
		OnProperty("feature.enabled", "true").
		And().
		OnBean("dataSource").
		Or().
		OnProperty("app.name").
		Build()
	if !b5.Matches(ctx) {
		t.Fatal("expected mixed AND/OR condition to match")
	}
}

func testConditionBuilderEmpty(t *testing.T, ctx ConditionContext) {
	t.Helper()
	b6 := New().Build()
	if !b6.Matches(ctx) {
		t.Fatal("expected empty builder to always match")
	}
}

// TestConditionBuilder 验证 ConditionBuilder 的流式 DSL:
//  1. 单个条件构建
//  2. And 操作
//  3. Or 操作
//  4. Not 操作
//  5. 混合操作
//  6. 空构建
func TestConditionBuilder(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			switch key {
			case "feature.enabled":
				return "true", true
			case "app.name":
				return "test", true
			default:
				return nil, false
			}
		},
		hasBeanFn: func(id string) bool {
			return id == "dataSource"
		},
	}

	testConditionBuilderSingle(t, ctx)
	testConditionBuilderAnd(t, ctx)
	testConditionBuilderOr(t, ctx)
	testConditionBuilderNot(t, ctx)
	testConditionBuilderMixed(t, ctx)
	testConditionBuilderEmpty(t, ctx)
}

// TestConditionBuilderExtended 验证扩展条件的构建器支持:
//  1. OnExpression
//  2. OnResourceExists
//  3. OnEnvVarExists
func TestConditionBuilderExtended(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			switch key {
			case "server.port":
				return 9090, true
			default:
				return nil, false
			}
		},
	}

	_ = os.Setenv("TEST_BUILDER_ENV", "yes")
	defer func() { _ = os.Unsetenv("TEST_BUILDER_ENV") }()

	// 测试 OnExpression
	b1 := New().OnExpression("${server.port} > 8080").Build()
	if !b1.Matches(ctx) {
		t.Fatal("expected OnExpression in builder to match")
	}

	// 测试 OnResourceExists
	b2 := New().OnResourceExists("file:README.md").Build()
	if !b2.Matches(ctx) {
		t.Fatal("expected OnResourceExists in builder to match")
	}

	// 测试 OnEnvVarExists
	b3 := New().OnEnvVarExists("TEST_BUILDER_ENV").Build()
	if !b3.Matches(ctx) {
		t.Fatal("expected OnEnvVarExists in builder to match")
	}

	// 测试组合使用
	b4 := New().
		OnExpression("${server.port} > 8080").
		And().
		OnEnvVarExists("TEST_BUILDER_ENV").
		And().
		OnResourceExists("file:README.md").
		Build()
	if !b4.Matches(ctx) {
		t.Fatal("expected combined conditions to match")
	}
}

func testAllWithDSLOr(t *testing.T, ctx ConditionContext) {
	t.Helper()
	dsl1 := AllWith(
		OnProperty("feature.enabled", "true"),
		OnBean("dataSource"),
	).Or(
		OnProfile("dev"),
	).Build()

	if !dsl1.Matches(ctx) {
		t.Fatal("expected All(...).Or(...) DSL to match")
	}
}

func testAllWithDSLAnd(t *testing.T, ctx ConditionContext) {
	t.Helper()
	dsl2 := AllWith(
		OnProperty("feature.enabled", "true"),
	).And(
		OnProperty("app.mode"),
	).Build()

	if !dsl2.Matches(ctx) {
		t.Fatal("expected All(...).And(...) DSL to match")
	}
}

func testAllWithDSLNested(t *testing.T, ctx ConditionContext) {
	t.Helper()
	dsl3 := AllWith(
		OnProperty("feature.enabled", "true"),
	).Or(
		OnProfile("dev"),
		OnProfile("test"),
	).And(
		OnBean("dataSource"),
	).Build()

	if !dsl3.Matches(ctx) {
		t.Fatal("expected nested DSL to match")
	}
}

// TestAllWithDSL 验证改进的 DSL:
//  1. All(...).Or(...) 链式调用
//  2. All(...).And(...) 链式调用
//  3. 嵌套组合
func TestAllWithDSL(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) {
			switch key {
			case "feature.enabled":
				return "true", true
			case "app.mode":
				return "dev", true
			default:
				return nil, false
			}
		},
		hasBeanFn: func(id string) bool {
			return id == "dataSource"
		},
	}

	testAllWithDSLOr(t, ctx)
	testAllWithDSLAnd(t, ctx)
	testAllWithDSLNested(t, ctx)
}

// TestConditionString 验证新增条件的 String 输出:
//  1. OnExpression 可读描述
//  2. OnResourceExists 可读描述
//  3. OnEnvVarExists 可读描述
//  4. OnResourceMissing 可读描述
//  5. OnEnvVarMissing 可读描述
func TestConditionString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		condition Condition
		expected  string
	}{
		{OnExpression("${server.port} > 8080"), "OnExpression(${server.port} > 8080)"},
		{OnResourceExists("classpath:config.yml"), "OnResourceExists(classpath:config.yml)"},
		{OnResourceMissing("file:/etc/app/config.json"), "OnResourceMissing(file:/etc/app/config.json)"},
		{OnEnvVarExists("DATABASE_URL"), "OnEnvVarExists(DATABASE_URL)"},
		{OnEnvVarMissing("REDIS_HOST"), "OnEnvVarMissing(REDIS_HOST)"},
	}

	for _, tt := range tests {
		if tt.condition.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.condition.String())
		}
	}
}

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
		propValue, ok := ctx.GetProperty("enabled")
		return ok && propValue == "true"
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
		propValue, ok := ctx.GetProperty("flag")
		return ok && propValue == "on"
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
		propValue, ok := ctx.GetProperty("ok")
		return ok && propValue == true
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
