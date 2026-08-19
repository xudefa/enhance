package testing

import (
	"fmt"
	"reflect"
	"testing"
)

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
	val := resp.Header("Content-Type")
	if val != "" {
		t.Errorf("expected empty header, got %q", val)
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
	val := ctx.GetProperty("key1")
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestNewTestContext_GetProperty_NotFound(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)
	val := ctx.GetProperty("nonexistent")
	if val != nil {
		t.Errorf("expected nil for nonexistent property, got %v", val)
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

func TestTestContext_Register(t *testing.T) {
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

	result, err := mock.Call("Compute", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 100 {
		t.Errorf("expected 100, got %v", result)
	}
}

func TestMock_ExpectTimes_NotEnoughCalls(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.ExpectTimes("Do", []any{}, nil, nil, 3)

	_, _ = mock.Call("Do")

	err := mock.Verify()
	if err == nil {
		t.Fatal("expected verification to fail with partial calls")
	}
}

func TestTestContext_Fatalf_Coverage(t *testing.T) {
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

	result := recorder.Times(2)
	if result == nil {
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
	val := resp.Header("X-Custom")
	if val != "value" {
		t.Errorf("expected 'value', got %q", val)
	}
	val = resp.Header("Nonexistent")
	if val != "" {
		t.Errorf("expected empty string for missing header, got %q", val)
	}
}
