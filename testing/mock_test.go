package testing

import (
	"errors"
	"testing"
)

func TestMock_ExpectAndCall(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{"user123"}, "user data", nil)

	result, err := mock.Call("GetUser", "user123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "user data" {
		t.Errorf("expected 'user data', got %v", result)
	}
}

func TestMock_ExpectTimes_FullCalls(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.ExpectTimes("GetUser", []any{"user123"}, "user data", nil, 3)

	for i := 0; i < 3; i++ {
		_, err := mock.Call("GetUser", "user123")
		if err != nil {
			t.Fatalf("call %d failed: %v", i+1, err)
		}
	}

	err := mock.Verify()
	if err != nil {
		t.Errorf("verification failed: %v", err)
	}
}

func TestMock_Verify_Success(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{"user123"}, "user data", nil)

	_, _ = mock.Call("GetUser", "user123")

	err := mock.Verify()
	if err != nil {
		t.Errorf("verification failed: %v", err)
	}
}

func TestMock_Reset(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{"user123"}, "user data", nil)

	mock.Reset()

	_, err := mock.Call("GetUser", "user123")
	if err == nil {
		t.Error("expected error after reset")
	}
}

func TestMock_Call_Unexpected(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{"user123"}, "user data", nil)

	_, err := mock.Call("DeleteUser", "user123")
	if err == nil {
		t.Error("expected error for unexpected method call")
	}
}

func TestMock_Call_WithArgsMismatch(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{"user123"}, "user data", nil)

	_, err := mock.Call("GetUser", "user456")
	if err == nil {
		t.Error("expected error for args mismatch")
	}
}

func TestMock_Call_WithError(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	expectedErr := errors.New("test error")
	mock.Expect("GetUser", []any{"user123"}, nil, expectedErr)

	_, err := mock.Call("GetUser", "user123")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %v", err)
	}
}

func TestMock_MultipleExpectations(t *testing.T) {
	t.Parallel()
	mock := NewMock()
	mock.Expect("GetUser", []any{"user1"}, "user1 data", nil)
	mock.Expect("GetUser", []any{"user2"}, "user2 data", nil)

	result1, err1 := mock.Call("GetUser", "user1")
	if err1 != nil {
		t.Fatalf("first call failed: %v", err1)
	}
	if result1 != "user1 data" {
		t.Errorf("expected 'user1 data', got %v", result1)
	}

	result2, err2 := mock.Call("GetUser", "user2")
	if err2 != nil {
		t.Fatalf("second call failed: %v", err2)
	}
	if result2 != "user2 data" {
		t.Errorf("expected 'user2 data', got %v", result2)
	}
}

func TestMethodKey(t *testing.T) {
	t.Parallel()
	key1 := methodKey("GetUser", []any{"user123"})
	key2 := methodKey("GetUser", []any{"user456"})

	if key1 == key2 {
		t.Error("expected different keys for different args")
	}

	key3 := methodKey("GetUser", []any{"user123"})
	if key1 != key3 {
		t.Error("expected same key for same args")
	}
}

func TestArgsMatch(t *testing.T) {
	t.Parallel()
	mock := &mockImpl{}

	t.Run("matching args", func(t *testing.T) {
		t.Parallel()
		if !mock.argsMatch([]any{1, "test"}, []any{1, "test"}) {
			t.Error("expected args to match")
		}
	})

	t.Run("different length", func(t *testing.T) {
		t.Parallel()
		if mock.argsMatch([]any{1}, []any{1, "test"}) {
			t.Error("expected args to not match")
		}
	})

	t.Run("different values", func(t *testing.T) {
		t.Parallel()
		if mock.argsMatch([]any{1, "test"}, []any{1, "other"}) {
			t.Error("expected args to not match")
		}
	})

	t.Run("empty args", func(t *testing.T) {
		t.Parallel()
		if !mock.argsMatch([]any{}, []any{}) {
			t.Error("expected empty args to match")
		}
	})
}
