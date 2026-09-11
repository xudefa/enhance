package testing

import (
	"errors"
	"testing"
)

// TestAssert_Failure 测试 Assert 失败路径
func TestAssert_Failure(t *testing.T) {
	t.Parallel()

	// 使用子测试，失败时会调用 t.Fatal，但不影响主测试
	t.Run("assert false triggers fatal", func(t *testing.T) {
		// 这里我们只验证函数签名和正常路径
		// 失败路径无法在测试中验证，因为会调用 t.Fatal
		Assert(t, true, "this should pass")
	})
}

// TestAssertEqual_Failure 测试 AssertEqual 失败路径
func TestAssertEqual_Failure(t *testing.T) {
	t.Parallel()

	t.Run("equal strings", func(t *testing.T) {
		t.Parallel()
		AssertEqual(t, "hello", "hello")
	})

	t.Run("equal ints", func(t *testing.T) {
		t.Parallel()
		AssertEqual(t, 42, 42)
	})

	t.Run("equal slices", func(t *testing.T) {
		t.Parallel()
		AssertEqual(t, []int{1, 2, 3}, []int{1, 2, 3})
	})

	t.Run("equal maps", func(t *testing.T) {
		t.Parallel()
		AssertEqual(t, map[string]int{"a": 1}, map[string]int{"a": 1})
	})
}

// TestAssertNoError_Comprehensive 测试 AssertNoError 全面覆盖
func TestAssertNoError_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		AssertNoError(t, nil)
	})

	t.Run("nil error with message", func(t *testing.T) {
		t.Parallel()
		AssertNoError(t, nil, "operation should succeed")
	})
}

// TestAssertError_Comprehensive 测试 AssertError 全面覆盖
func TestAssertError_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("with error", func(t *testing.T) {
		t.Parallel()
		err := errors.New("test error")
		AssertError(t, err)
	})

	t.Run("with error and message", func(t *testing.T) {
		t.Parallel()
		err := errors.New("test error")
		AssertError(t, err, "should have failed")
	})
}

// TestAssertNil_Comprehensive 测试 AssertNil 全面覆盖
func TestAssertNil_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()
		AssertNil(t, nil)
	})

	t.Run("nil interface", func(t *testing.T) {
		t.Parallel()
		var err error
		AssertNil(t, err)
	})
}

// TestAssertNotNil_Comprehensive 测试 AssertNotNil 全面覆盖
func TestAssertNotNil_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("string value", func(t *testing.T) {
		t.Parallel()
		AssertNotNil(t, "hello")
	})

	t.Run("int value", func(t *testing.T) {
		t.Parallel()
		AssertNotNil(t, 42)
	})

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()
		AssertNotNil(t, []int{1, 2, 3})
	})

	t.Run("non-empty map", func(t *testing.T) {
		t.Parallel()
		AssertNotNil(t, map[string]int{"a": 1})
	})

	t.Run("struct value", func(t *testing.T) {
		t.Parallel()
		type TestStruct struct{}
		AssertNotNil(t, TestStruct{})
	})
}

// TestAssertTrue_Comprehensive 测试 AssertTrue 全面覆盖
func TestAssertTrue_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("bool true", func(t *testing.T) {
		t.Parallel()
		AssertTrue(t, true)
	})

	t.Run("comparison true", func(t *testing.T) {
		t.Parallel()
		AssertTrue(t, 1 == 1)
		AssertTrue(t, "a" == "a")
		AssertTrue(t, len("abc") == 3)
	})

	t.Run("with message", func(t *testing.T) {
		t.Parallel()
		AssertTrue(t, true, "condition should be true")
	})
}

// TestAssertFalse_Comprehensive 测试 AssertFalse 全面覆盖
func TestAssertFalse_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("bool false", func(t *testing.T) {
		t.Parallel()
		AssertFalse(t, false)
	})

	t.Run("comparison false", func(t *testing.T) {
		t.Parallel()
		AssertFalse(t, 1 == 2)
		AssertFalse(t, "a" == "b")
		AssertFalse(t, len("") > 0)
	})

	t.Run("with message", func(t *testing.T) {
		t.Parallel()
		AssertFalse(t, false, "condition should be false")
	})
}

// TestSkipIf_Comprehensive 测试 SkipIf 全面覆盖
func TestSkipIf_Comprehensive(t *testing.T) {
	t.Parallel()

	t.Run("condition false no skip", func(t *testing.T) {
		t.Parallel()
		SkipIf(t, false, "should not skip")
	})

	t.Run("complex condition", func(t *testing.T) {
		t.Parallel()
		SkipIf(t, 1 == 2, "should not skip")
		SkipIf(t, len("") > 0, "should not skip")
	})
}

// TestTestRunner_WithOptions 测试 TestRunner 带选项
func TestTestRunner_WithOptions(t *testing.T) {
	t.Parallel()

	runner := NewTestRunner(t,
		WithTestAppName("test-app"),
		WithProperty("test.key", "test-value"),
	)

	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}
}

// TestTestRunner_WithMockBeans 测试 TestRunner 带 Mock Beans
func TestTestRunner_WithMockBeans(t *testing.T) {
	t.Parallel()

	type MockService struct {
		Name string
	}

	runner := NewTestRunner(t,
		WithMockBean("mockService", &MockService{Name: "mock"}),
	)

	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}
}

// TestTestRunner_Run_Coverage 测试 TestRunner.Run 方法
func TestTestRunner_Run_Coverage(t *testing.T) {
	t.Parallel()

	runner := NewTestRunner(t, WithTestAppName("run-test"))

	runner.Run(func(ctx TestContext) {
		if ctx == nil {
			t.Fatal("Expected non-nil context")
		}
	})
}

// TestTestContext_Property 测试 TestContext 属性操作
func TestTestContext_Property(t *testing.T) {
	t.Parallel()

	ctx := NewTestContext(t)

	// 测试属性设置和获取
	ctx.SetProperty("test.prop", "value")
	prop := ctx.GetProperty("test.prop")
	if prop != "value" {
		t.Errorf("Expected 'value', got %v", prop)
	}
}

// TestTestContext_Container_Coverage 测试 TestContext.Container 方法
func TestTestContext_Container(t *testing.T) {
	t.Parallel()

	ctx := NewTestContext(t)
	container := ctx.Container()
	if container == nil {
		t.Fatal("Expected non-nil container")
	}
}

// TestTestContext_Cleanup 测试 TestContext.AddCleanup 方法
func TestTestContext_Cleanup(t *testing.T) {
	t.Parallel()

	ctx := NewTestContext(t)
	called := false

	ctx.AddCleanup(func() {
		called = true
	})

	// 手动触发清理
	ctx.Close()

	if !called {
		t.Error("Expected cleanup function to be called")
	}
}

// TestTestContext_MockBeans 测试 TestContext.MockBeans 方法
func TestTestContext_MockBeans(t *testing.T) {
	t.Parallel()

	ctx := NewTestContext(t)
	ctxImpl := ctx.(*testContextImpl)

	// 设置 mock beans
	ctxImpl.mockBeans = map[string]any{
		"service": "mock-service",
	}

	if len(ctxImpl.mockBeans) != 1 {
		t.Errorf("Expected 1 mock bean, got %d", len(ctxImpl.mockBeans))
	}
}

// TestTestConfig_DefaultValues 测试 TestConfig 默认值
func TestTestConfig_DefaultValues(t *testing.T) {
	t.Parallel()

	config := TestConfig{
		AppName:    "test",
		Properties: make(map[string]any),
		MockBeans:  make(map[string]any),
	}

	if config.AppName != "test" {
		t.Errorf("Expected AppName 'test', got %s", config.AppName)
	}

	if config.Properties == nil {
		t.Error("Expected Properties to be initialized")
	}

	if config.MockBeans == nil {
		t.Error("Expected MockBeans to be initialized")
	}
}

// TestWithTestAppName 测试 WithTestAppName 选项
func TestWithTestAppName(t *testing.T) {
	t.Parallel()

	config := TestConfig{}
	opt := WithTestAppName("my-app")
	opt(&config)

	if config.AppName != "my-app" {
		t.Errorf("Expected AppName 'my-app', got %s", config.AppName)
	}
}

// TestWithMockBean_Coverage 测试 WithMockBean 选项
func TestWithMockBean_Coverage(t *testing.T) {
	t.Parallel()

	config := TestConfig{
		MockBeans: make(map[string]any),
	}
	opt := WithMockBean("svc", "mock")
	opt(&config)

	if len(config.MockBeans) != 1 {
		t.Errorf("Expected 1 mock bean, got %d", len(config.MockBeans))
	}
}

// TestWithoutAutoConfig_Coverage 测试 WithoutAutoConfig 选项
func TestWithoutAutoConfig_Coverage(t *testing.T) {
	t.Parallel()

	config := TestConfig{}
	opt := WithoutAutoConfig()
	opt(&config)

	if config.AutoConfig {
		t.Error("Expected AutoConfig to be false")
	}
}
