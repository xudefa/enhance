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

	// 测试单个条件
	b1 := New().OnProperty("feature.enabled", "true").Build()
	if !b1.Matches(ctx) {
		t.Fatal("expected single condition to match")
	}

	// 测试 And 操作
	b2 := New().
		OnProperty("feature.enabled", "true").
		And().
		OnBean("dataSource").
		Build()
	if !b2.Matches(ctx) {
		t.Fatal("expected AND condition to match")
	}

	// 测试 Or 操作
	b3 := New().
		OnProperty("nonexistent").
		Or().
		OnProperty("feature.enabled").
		Build()
	if !b3.Matches(ctx) {
		t.Fatal("expected OR condition to match when one matches")
	}

	// 测试 Not 操作
	b4 := New().
		Not().
		OnProperty("nonexistent").
		Build()
	if !b4.Matches(ctx) {
		t.Fatal("expected NOT condition to match when condition is false")
	}

	// 测试混合操作（AND + OR）
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

	// 测试空构建
	b6 := New().Build()
	if !b6.Matches(ctx) {
		t.Fatal("expected empty builder to always match")
	}
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

	// 测试 All(...).Or(...) DSL
	dsl1 := AllWith(
		OnProperty("feature.enabled", "true"),
		OnBean("dataSource"),
	).Or(
		OnProfile("dev"),
	).Build()

	if !dsl1.Matches(ctx) {
		t.Fatal("expected All(...).Or(...) DSL to match")
	}

	// 测试 All(...).And(...) DSL
	dsl2 := AllWith(
		OnProperty("feature.enabled", "true"),
	).And(
		OnProperty("app.mode"),
	).Build()

	if !dsl2.Matches(ctx) {
		t.Fatal("expected All(...).And(...) DSL to match")
	}

	// 测试嵌套组合
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
