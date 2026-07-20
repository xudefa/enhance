package testing

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

type mockService struct{}

type User struct {
	Name string
}

func TestTestRunner_BasicUsage(t *testing.T) {
	t.Parallel()
	runner := NewTestRunner(t)
	if runner == nil || runner.config.AppName != "test-app" || !runner.config.AutoConfig {
		t.Fatalf("expected valid runner with AppName=test-app and AutoConfig=true, got %+v", runner)
	}
}

func TestTestRunner_WithOptions(t *testing.T) {
	t.Parallel()
	runner := NewTestRunner(t,
		WithProperty("test.key", "test-value"),
		WithMockBean("mockService", &mockService{}),
		WithoutAutoConfig(),
	)

	if runner.config.Properties["test.key"] != "test-value" {
		t.Errorf("property test.key = %v, want %v", runner.config.Properties["test.key"], "test-value")
	}

	if _, ok := runner.config.MockBeans["mockService"]; !ok {
		t.Error("expected mockService to be registered")
	}

	if runner.config.AutoConfig {
		t.Error("expected AutoConfig to be false")
	}
}

func TestTestContext_GetByType(t *testing.T) {
	t.Parallel()
	ctx := NewTestContext(t)

	container := ctx.Container().(core.Container)
	def := registry.BeanDef{
		Type: reflect.TypeOf(&mockService{}),
		Factory: func(c ...any) (any, error) {
			return &mockService{}, nil
		},
	}
	_ = container.RegisterBean(def)

	bean := ctx.GetByType(reflect.TypeOf(&mockService{}))
	if bean == nil {
		t.Error("expected non-nil bean")
	}
}

func TestMock_ExpectAndCall(t *testing.T) {
	t.Parallel()
	mock := NewMock()

	mock.Expect("GetUser", []any{1}, &User{Name: "Alice"}, nil)

	result, err := mock.Call("GetUser", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, ok := result.(*User)
	if !ok {
		t.Fatalf("expected *User, got %T", result)
	}
	if user.Name != "Alice" {
		t.Errorf("User.Name = %q, want %q", user.Name, "Alice")
	}
}

func TestMock_VerifyExpectations(t *testing.T) {
	t.Parallel()
	mock := NewMock()

	mock.Expect("GetUser", []any{1}, &User{Name: "Alice"}, nil)

	_, _ = mock.Call("GetUser", 1)

	err := mock.Verify()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMock_UnmetExpectations(t *testing.T) {
	t.Parallel()
	mock := NewMock()

	mock.Expect("GetUser", []any{1}, &User{Name: "Alice"}, nil)

	err := mock.Verify()
	if err == nil {
		t.Error("expected error for unmet expectations")
	}
}

func TestMock_ExpectTimes(t *testing.T) {
	t.Parallel()
	mock := NewMock()

	mock.ExpectTimes("GetUser", []any{1}, &User{Name: "Alice"}, nil, 3)

	for range 3 {
		_, err := mock.Call("GetUser", 1)
		if err != nil {
			t.Fatalf("call failed: %v", err)
		}
	}

	err := mock.Verify()
	if err != nil {
		t.Errorf("verification failed: %v", err)
	}
}

func TestMock_Reset(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{1}, &User{Name: "Alice"}, nil)
	mock.Reset()

	_, err := mock.Call("GetUser", 1)
	if err == nil {
		t.Error("expected error after reset")
	}
}

func TestAssertions(t *testing.T) {
	t.Parallel()
	t.Run("AssertEqual", func(t *testing.T) {
		AssertEqual(t, "expected", "expected")
	})

	t.Run("AssertNoError", func(t *testing.T) {
		AssertNoError(t, nil)
	})

	t.Run("AssertError", func(t *testing.T) {
		AssertError(t, &mockError{})
	})

	t.Run("AssertTrue", func(t *testing.T) {
		Assert(t, true, "expected true")
	})

	t.Run("AssertNil", func(t *testing.T) {
		AssertNil(t, nil)
	})

	t.Run("AssertNotNil", func(t *testing.T) {
		AssertNotNil(t, "value")
	})
}

func TestAssertExpectations(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{1}, &User{Name: "Alice"}, nil)
	_, _ = mock.Call("GetUser", 1)

	if !AssertExpectations(t, mock) {
		t.Error("expected assertions to pass")
	}
}

type mockError struct{}

func (e *mockError) Error() string {
	return "mock error"
}
