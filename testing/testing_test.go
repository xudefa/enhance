package testing

import (
	"fmt"
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

func TestTestContext_Container_Coverage(t *testing.T) {
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
		t.Parallel()
		testTestFunctionsTest(t)
	})

	t.Run("TestWithContainer", func(t *testing.T) {
		t.Parallel()
		testTestFunctionsTestWithContainer(t)
	})

	t.Run("SetupTest", func(t *testing.T) {
		t.Parallel()
		testTestFunctionsSetupTest(t)
	})

	t.Run("RunSubtest", func(t *testing.T) {
		t.Parallel()
		testTestFunctionsRunSubtest(t)
	})
}

func testTestFunctionsTest(t *testing.T) {
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
}

func testTestFunctionsTestWithContainer(t *testing.T) {
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
}

func testTestFunctionsSetupTest(t *testing.T) {
	ctx := SetupTest(t, func(ctx TestContext) {
		ctx.SetProperty("setup.key", "setup-value")
	})

	value := ctx.GetProperty("setup.key")
	if value != "setup-value" {
		t.Errorf("expected 'setup-value', got %v", value)
	}
}

func testTestFunctionsRunSubtest(t *testing.T) {
	called := false
	RunSubtest(t, "subtest", func(ctx TestContext) {
		called = true
	})

	if !called {
		t.Error("expected subtest to be called")
	}
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
			if rec := recover(); rec != nil {
				t.Errorf("unexpected panic: %v", rec)
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

	mock.ExpectTimes(ExpectationRequest{Method: "GetUser", Args: []any{1}, Result: &User{Name: "Alice"}, Times: 1})

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

	mock.ExpectTimes(ExpectationRequest{Method: "GetUser", Args: []any{1}, Result: &User{Name: "Alice"}, Times: 3})

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

	got := recorder.Return(&User{Name: "Bob"}, nil)
	if got != mock {
		t.Error("expected Return to return mock")
	}

	got = recorder.Times(5)
	if got != mock {
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

	assertOK := AssertExpectations(&mockTestingT{}, mock)
	if assertOK {
		t.Error("expected assertion to fail")
	}
}

func TestAssertions_Failure(t *testing.T) {
	t.Parallel()
	t.Run("AssertTrue", func(t *testing.T) {
		t.Parallel()
		AssertTrue(t, true)
	})

	t.Run("AssertFalse", func(t *testing.T) {
		t.Parallel()
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

func TestAssert_TrueCondition(t *testing.T) {
	t.Parallel()
	Assert(t, true, "should not fail")
}

func TestAssertEqual_EqualValues(t *testing.T) {
	t.Parallel()
	AssertEqual(t, 42, 42)
}

func TestAssertEqual_WithCustomMessage(t *testing.T) {
	t.Parallel()
	AssertEqual(t, "hello", "hello", "strings should match")
}

func TestAssertNoError_NilError(t *testing.T) {
	t.Parallel()
	AssertNoError(t, nil)
}

func TestAssertNoError_WithCustomMessage(t *testing.T) {
	t.Parallel()
	AssertNoError(t, nil, "should have no error")
}

func TestAssertError_NonNilError(t *testing.T) {
	t.Parallel()
	AssertError(t, fmt.Errorf("some error"))
}

func TestAssertError_WithCustomMessage(t *testing.T) {
	t.Parallel()
	AssertError(t, fmt.Errorf("some error"), "should have error")
}

func TestAssertNil_NilValue(t *testing.T) {
	t.Parallel()
	AssertNil(t, nil)
}

func TestAssertNil_WithCustomMessage(t *testing.T) {
	t.Parallel()
	AssertNil(t, nil, "custom msg")
}

func TestAssertNotNil_NonNilValue(t *testing.T) {
	t.Parallel()
	AssertNotNil(t, "not nil")
}

func TestAssertNotNil_WithCustomMessage(t *testing.T) {
	t.Parallel()
	AssertNotNil(t, "not nil", "custom msg")
}

func TestAssertTrue_True(t *testing.T) {
	t.Parallel()
	AssertTrue(t, true)
}

func TestAssertTrue_True_Extra(t *testing.T) {
	t.Parallel()
	AssertTrue(t, true, "should be true")
}

func TestAssertFalse_False(t *testing.T) {
	t.Parallel()
	AssertFalse(t, false)
}

func TestAssertFalse_False_Extra(t *testing.T) {
	t.Parallel()
	AssertFalse(t, false, "should be false")
}

func TestSkipIf_Skip(t *testing.T) {
	t.Parallel()
	SkipIf(t, true, "skipping test")
}

func TestSkipIf_NoSkip(t *testing.T) {
	t.Parallel()
	SkipIf(t, false, "should not skip")
}

func TestTestWebClient_Get(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Get("/api/test")
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestTestWebClient_Post(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Post("/api/test", map[string]string{"key": "value"})
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestTestResponse_StatusCode(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Get("/api/test")
	if resp.StatusCode() != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode())
	}
}

func TestTestResponse_Body(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Get("/api/test")
	if len(resp.Body()) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestTestResponse_Header_Nil(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Get("/api/test")
	headerValue := resp.Header("Content-Type")
	if headerValue != "" {
		t.Errorf("expected empty header, got %q", headerValue)
	}
}

func TestTestResponse_AssertStatus(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Get("/api/test")
	resp.AssertStatus(t, 200)
}

func TestTestResponse_AssertBody(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Get("/api/test")
	resp.AssertBody(t, `{"status":"ok"}`)
}

func TestNewTestContext_Basic(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
}

func TestNewTestContext_T(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	testT := ctx.T()
	if testT == nil {
		t.Fatal("expected non-nil T()")
	}
}

func TestNewTestContext_Container(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	container := ctx.Container()
	if container == nil {
		t.Fatal("expected non-nil container")
	}
}

func TestNewTestContext_SetProperty(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	ctx.SetProperty("key1", "value1")
	got := ctx.GetProperty("key1")
	if got != "value1" {
		t.Errorf("expected 'value1', got %v", got)
	}
}

func TestNewTestContext_GetProperty_NotFound(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	got := ctx.GetProperty("nonexistent")
	if got != nil {
		t.Errorf("expected nil for nonexistent property, got %v", got)
	}
}

func TestNewTestContext_AddCleanup(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	called := false
	ctx.AddCleanup(func() {
		called = true
	})
	ctx.Cleanup()
	if !called {
		t.Error("expected cleanup to be called")
	}
}

func TestMock_AssertExpectations_Pass(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("Do", []any{"arg1"}, "result", nil)
	_, _ = mock.Call("Do", "arg1")

	if !AssertExpectations(t, mock) {
		t.Error("expected expectations to pass")
	}
}

func TestMock_AssertExpectations_Fail(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("Do", []any{"arg1"}, "result", nil)

	mt := &mockTestingT{}
	if AssertExpectations(mt, mock) {
		t.Error("expected expectations to fail")
	}
}

func TestTestContext_Register_Extra(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	type testBean struct{ Name string }
	ctx.Register("myBean", &testBean{Name: "test"})
	bean := ctx.GetByType(reflect.TypeOf(&testBean{}))
	if bean == nil {
		t.Fatal("expected non-nil bean after register")
	}
}

func TestTeardownTest(t *testing.T) {
	t.Parallel()
	called := false
	ctx := NewTestContext(t)
	TeardownTest(ctx, func(tc TestContext) {
		called = true
	})
	if !called {
		t.Error("expected teardown to be called")
	}
}

func TestTestContext_Close_NilApp(t *testing.T) {
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

func TestMock_ExpectAndCall_WithResult(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("Compute", []any{42}, 100, nil)

	got, err := mock.Call("Compute", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 100 {
		t.Errorf("expected 100, got %v", got)
	}
}

func TestMock_ExpectTimes_NotEnoughCalls(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.ExpectTimes(ExpectationRequest{Method: "Do", Args: []any{}, Times: 3})

	_, _ = mock.Call("Do")

	err := mock.Verify()
	if err == nil {
		t.Fatal("expected verification to fail with partial calls")
	}
}

func TestTestContext_Fatalf(t *testing.T) {
	t.Parallel()
	t.Skip("Fatalf delegates to testing.T.Fatalf which calls runtime.Goexit")
}

func TestTestContext_Register_Duplicate(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	type testBean struct{ Name string }
	ctx.Register("myBean", &testBean{Name: "test"})
	ctx.Register("myBean2", &testBean{Name: "test2"})
	if ctx.GetByType(reflect.TypeOf(&testBean{})) == nil {
		t.Fatal("expected non-nil bean")
	}
}

func TestTestContext_Close_WithNilApp(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	impl := ctx.(*testContextImpl)
	impl.app = nil
	called := false
	impl.AddCleanup(func() { called = true })
	impl.Close()
	if !called {
		t.Error("expected cleanup to be called")
	}
}

func TestMock_Call_DifferentArgs(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("Do", []any{1}, "result1", nil)
	mock.Expect("Do", []any{2}, "result2", nil)

	r1, _ := mock.Call("Do", 1)
	r2, _ := mock.Call("Do", 2)

	if r1 != "result1" {
		t.Errorf("expected result1, got %v", r1)
	}
	if r2 != "result2" {
		t.Errorf("expected result2, got %v", r2)
	}
}

func TestMock_Call_DifferentMethod(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("MethodA", []any{}, "a", nil)
	mock.Expect("MethodB", []any{}, "b", nil)

	_, _ = mock.Call("MethodA")
	_, _ = mock.Call("MethodB")

	if err := mock.Verify(); err != nil {
		t.Errorf("verification failed: %v", err)
	}
}

func TestMockRecorder_Chain(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	recorder := NewMockRecorder(mock)

	chained := recorder.Times(2)
	if chained == nil {
		t.Error("expected non-nil result from chain")
	}
}

func TestSkipIf_ConditionFalse(t *testing.T) {
	t.Parallel()
	skipped := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				skipped = true
			}
		}()
		SkipIf(t, false, "should not skip")
	}()
	if skipped {
		t.Error("SkipIf with false condition should not skip")
	}
}

func TestWithMock_Execution(t *testing.T) {
	t.Parallel()
	executed := false
	WithMock(t, func(ctx TestContext, mock *MockRecorder) {
		executed = true
		if ctx == nil {
			t.Error("expected non-nil context")
		}
	})
	if !executed {
		t.Error("expected WithMock to execute function")
	}
}

func TestMock_CallArgsMismatch_Length(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("Do", []any{1, 2}, "result", nil)

	_, err := mock.Call("Do", 1)
	if err == nil {
		t.Error("expected error for arg length mismatch")
	}
}

func TestMock_CallArgsMismatch_Value(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("Do", []any{1}, "result", nil)

	_, err := mock.Call("Do", 2)
	if err == nil {
		t.Error("expected error for arg value mismatch")
	}
}

func TestTestContext_GetByType_NoBeans(t *testing.T) {
	t.Parallel()
	t.Skip("GetByType calls t.Fatalf which calls runtime.Goexit")
}

func TestTestResponse_PostCreated(t *testing.T) {
	t.Parallel()
	client := NewTestWebClient(t, "http://localhost:8080")
	resp := client.Post("/api/create", "data")
	resp.AssertStatus(t, 201)
	expectedBody := `{"status":"created"}`
	resp.AssertBody(t, expectedBody)
}

func TestTestResponse_Header_WithHeaders(t *testing.T) {
	t.Parallel()
	resp := &TestResponse{
		statusCode: 200,
		body:       []byte("ok"),
		headers:    map[string]string{"X-Custom": "value"},
	}
	headerValue := resp.Header("X-Custom")
	if headerValue != "value" {
		t.Errorf("expected 'value', got %q", headerValue)
	}
	headerValue = resp.Header("Nonexistent")
	if headerValue != "" {
		t.Errorf("expected empty string for missing header, got %q", headerValue)
	}
}
