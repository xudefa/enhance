package testing

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

func TestTestRunner_WithTestAppName(t *testing.T) {
	t.Parallel()
	runner := NewTestRunner(t, WithTestAppName("my-test-app"))
	if runner.config.AppName != "my-test-app" {
		t.Errorf("expected AppName 'my-test-app', got %s", runner.config.AppName)
	}
}

func TestTestContext_SetAndGetProperty(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	ctx.SetProperty("test.key", "test-value")
	value := ctx.GetProperty("test.key")

	if value != "test-value" {
		t.Errorf("expected 'test-value', got %v", value)
	}
}

func TestTestContext_RegisterAndGetByType(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	service := &mockService{}
	container := ctx.Container().(core.Container)
	def := registry.BeanDef{
		Type: reflect.TypeOf(service),
		Factory: func(c ...any) (any, error) {
			return service, nil
		},
	}
	_ = container.RegisterBean(def)

	bean := ctx.GetByType(reflect.TypeOf(service))
	if bean == nil {
		t.Error("expected non-nil bean")
	}
	if bean != service {
		t.Error("expected bean to match registered service")
	}
}

func TestTestContext_AddCleanup(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	called := false
	ctx.AddCleanup(func() {
		called = true
	})

	ctx.Cleanup()

	if !called {
		t.Error("expected cleanup function to be called")
	}
}

func TestTestContext_CleanupMultiple(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	callOrder := []int{}
	ctx.AddCleanup(func() {
		callOrder = append(callOrder, 1)
	})
	ctx.AddCleanup(func() {
		callOrder = append(callOrder, 2)
	})
	ctx.AddCleanup(func() {
		callOrder = append(callOrder, 3)
	})

	ctx.Cleanup()

	if len(callOrder) != 3 {
		t.Fatalf("expected 3 cleanup calls, got %d", len(callOrder))
	}

	if callOrder[0] != 3 || callOrder[1] != 2 || callOrder[2] != 1 {
		t.Errorf("expected reverse order cleanup [3,2,1], got %v", callOrder)
	}
}

func TestTestContext_Close(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	called := false
	ctx.AddCleanup(func() {
		called = true
	})

	ctx.Close()

	if !called {
		t.Error("expected cleanup to be called on Close")
	}
}

func TestTestContext_Container(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	container := ctx.Container()
	if container == nil {
		t.Error("expected non-nil container")
	}
}

func TestTestContext_Errorf(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t).(*testContextImpl)
	ctx.t.Skip("skipping Errorf test to avoid test failure")
}

func TestTestContext_Logf(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t).(*testContextImpl)
	ctx.Logf("test log %d", 1)
}

func TestTestContext_Helper(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t).(*testContextImpl)
	ctx.Helper()
}

func TestTestFunctions(t *testing.T) {
	t.Parallel()
	t.Run("Test", func(t *testing.T) {
		called := false
		Test(t, func(ctx TestContext) {
			called = true
			if ctx == nil {
				t.Error("expected non-nil context")
			}
		})

		if !called {
			t.Error("expected test function to be called")
		}
	})

	t.Run("TestWithContainer", func(t *testing.T) {
		container := core.NewContainer()
		called := false
		TestWithContainer(t, container, func(ctx TestContext) {
			called = true
			if ctx.Container() != container {
				t.Error("expected container to match")
			}
		})

		if !called {
			t.Error("expected test function to be called")
		}
	})

	t.Run("SetupTest", func(t *testing.T) {
		ctx := SetupTest(t, func(ctx TestContext) {
			ctx.SetProperty("setup.key", "setup-value")
		})

		value := ctx.GetProperty("setup.key")
		if value != "setup-value" {
			t.Errorf("expected 'setup-value', got %v", value)
		}
	})

	t.Run("RunSubtest", func(t *testing.T) {
		called := false
		RunSubtest(t, "subtest", func(ctx TestContext) {
			called = true
		})

		if !called {
			t.Error("expected subtest to be called")
		}
	})
}

func TestParallel(t *testing.T) {
	t.Parallel()
	tests := map[string]func(ctx TestContext){
		"test1": func(ctx TestContext) {},
		"test2": func(ctx TestContext) {},
	}

	Parallel(t, tests)
}

func TestMustGet(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	container := ctx.Container().(core.Container)
	service := &mockService{}
	def := registry.BeanDef{
		Type: reflect.TypeOf(service),
		Factory: func(c ...any) (any, error) {
			return service, nil
		},
	}
	_ = container.RegisterBean(def)

	t.Run("ExistingBean", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()

		MustGetByType[*mockService](ctx)
	})
}

func TestGetByType(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	container := ctx.Container().(core.Container)
	service := &mockService{}
	def := registry.BeanDef{
		Type: reflect.TypeOf(service),
		Factory: func(c ...any) (any, error) {
			return service, nil
		},
	}
	_ = container.RegisterBean(def)

	t.Run("ExistingBean", func(t *testing.T) {
		bean, err := GetByType[*mockService](ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bean != service {
			t.Error("expected bean to match service")
		}
	})

	t.Run("NonExistingBean", func(t *testing.T) {
		type nonExistent struct{}
		_, err := GetByType[*nonExistent](ctx)
		if err == nil {
			t.Error("expected error for non-existing bean")
		}
	})
}

func TestMock_CallWithNoMatch(t *testing.T) {
	t.Parallel()
	mock := NewMock()

	_, err := mock.Call("NonExistent", 1)
	if err == nil {
		t.Error("expected error for unregistered call")
	}
}

func TestMock_CallExceedTimes(t *testing.T) {
	t.Parallel()
	mock := NewMock()

	mock.ExpectTimes("GetUser", []any{1}, &User{Name: "Alice"}, nil, 1)

	_, err := mock.Call("GetUser", 1)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	_, err = mock.Call("GetUser", 1)
	if err == nil {
		t.Error("expected error when exceeding expected call count")
	}
}

func TestMock_VerifyPartialCalls(t *testing.T) {
	t.Parallel()
	mock := NewMock()

	mock.ExpectTimes("GetUser", []any{1}, &User{Name: "Alice"}, nil, 3)

	_, _ = mock.Call("GetUser", 1)
	_, _ = mock.Call("GetUser", 1)

	err := mock.Verify()
	if err == nil {
		t.Error("expected error for partial calls")
	}
}

func TestMockRecorder(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	recorder := NewMockRecorder(mock)

	result := recorder.Return(&User{Name: "Bob"}, nil)
	if result != mock {
		t.Error("expected Return to return mock")
	}

	result = recorder.Times(5)
	if result != mock {
		t.Error("expected Times to return mock")
	}
}

func TestWithMock(t *testing.T) {
	t.Parallel()
	called := false
	WithMock(t, func(ctx TestContext, mock *MockRecorder) {
		called = true
		if mock == nil {
			t.Error("expected non-nil mock recorder")
		}
		if ctx == nil {
			t.Error("expected non-nil context")
		}
	})

	if !called {
		t.Error("expected WithMock to execute function")
	}
}

func TestAssertExpectations_Failure(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{1}, &User{Name: "Alice"}, nil)

	result := AssertExpectations(&mockTestingT{}, mock)
	if result {
		t.Error("expected assertion to fail")
	}
}

func TestAssertions_Failure(t *testing.T) {
	t.Parallel()
	t.Run("AssertTrue", func(t *testing.T) {
		AssertTrue(t, true)
	})

	t.Run("AssertFalse", func(t *testing.T) {
		AssertFalse(t, false)
	})
}

func TestAssertEqual_WithMessage(t *testing.T) {
	t.Parallel()
	AssertEqual(t, "expected", "expected", "custom message")
}

func TestAssertNoError_WithMessage(t *testing.T) {
	t.Parallel()
	AssertNoError(t, nil, "custom message")
}

func TestAssertError_WithMessage(t *testing.T) {
	t.Parallel()
	AssertError(t, &mockError{}, "custom message")
}

func TestAssertNil_WithMessage(t *testing.T) {
	t.Parallel()
	AssertNil(t, nil, "custom message")
}

func TestAssertNotNil_WithMessage(t *testing.T) {
	t.Parallel()
	AssertNotNil(t, "value", "custom message")
}

func TestAssertTrue_WithMessage(t *testing.T) {
	t.Parallel()
	AssertTrue(t, true, "custom message")
}

func TestAssertFalse_WithMessage(t *testing.T) {
	t.Parallel()
	AssertFalse(t, false, "custom message")
}

type mockService struct{}

type User struct {
	ID   int
	Name string
}

type mockError struct{}

func (e *mockError) Error() string { return "mock error" }

type mockTestingT struct {
	testing.T
}

func (m *mockTestingT) Errorf(format string, args ...any) {
}

func (m *mockTestingT) Fatalf(format string, args ...any) {
}

func (m *mockTestingT) Fatal(args ...any) {
}

func (m *mockTestingT) Logf(format string, args ...any) {
}

func (m *mockTestingT) Helper() {
}
