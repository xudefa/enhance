package aop

import (
	"context"
	"testing"
)

// TestPanicInfo_ContainsStack 测试 PanicInfo 包含堆栈信息
func TestPanicInfo_ContainsStack(t *testing.T) {
	t.Parallel()
	// 创建一个会 panic 的函数
	defer func() {
		if r := recover(); r != nil {
			// 验证 panic 值是 PanicInfo 类型
			panicInfo, ok := r.(*PanicInfo)
			if !ok {
				t.Fatalf("expected *PanicInfo, got %T", r)
			}

			// 验证包含原始 panic 值
			if panicInfo.Value == nil {
				t.Error("PanicInfo.Value should not be nil")
			}

			// 验证包含堆栈信息
			if len(panicInfo.Stack) == 0 {
				t.Error("PanicInfo.Stack should not be empty")
			}

			// 验证 Error() 方法正常工作
			errMsg := panicInfo.Error()
			if errMsg == "" {
				t.Error("PanicInfo.Error() should not return empty string")
			}

			t.Logf("PanicInfo captured successfully, stack size: %d bytes", len(panicInfo.Stack))
		}
	}()

	// 创建一个会触发 panic 的代理
	executor := NewChainExecutor(WithRecovery())

	// 创建一个 mock JoinPoint
	joinPoint := &mockJoinPointForPanicTest{
		proceedFunc: func() (any, error) {
			panic("test panic message")
		},
	}

	// 创建 Invocation
	inv := NewInvocation(
		joinPoint,
		func() (any, error) {
			panic("test panic message")
		},
	)

	// 执行会触发 panic
	executor.Execute(inv, nil, func(args ...any) any {
		panic("test panic message")
	})

	t.Fatal("should have panicked")
}

// TestPanicInfo_ErrorFormat 测试 PanicInfo 的 Error 输出格式
func TestPanicInfo_ErrorFormat(t *testing.T) {
	t.Parallel()
	panicInfo := &PanicInfo{
		Value: "test panic",
		Stack: []byte("stack trace here"),
	}

	errMsg := panicInfo.Error()

	// 验证包含 panic 值
	if !containsStr(errMsg, "test panic") {
		t.Errorf("Error() should contain panic value, got: %s", errMsg)
	}

	// 验证包含堆栈
	if !containsStr(errMsg, "stack trace here") {
		t.Errorf("Error() should contain stack trace, got: %s", errMsg)
	}
}

// TestPanicInfo_ImplementsError 测试 PanicInfo 实现 error 接口
func TestPanicInfo_ImplementsError(t *testing.T) {
	t.Parallel()
	var _ error = (*PanicInfo)(nil)
}

// 辅助函数
func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsStrMiddle(s, substr)))
}

func containsStrMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mockJoinPointForPanicTest 用于 panic 测试的 mock JoinPoint
type mockJoinPointForPanicTest struct {
	proceedFunc func() (any, error)
}

func (m *mockJoinPointForPanicTest) Target() any           { return nil }
func (m *mockJoinPointForPanicTest) Method() string        { return "" }
func (m *mockJoinPointForPanicTest) Args() []any           { return nil }
func (m *mockJoinPointForPanicTest) Proceed() (any, error) { return m.proceedFunc() }
func (m *mockJoinPointForPanicTest) ProceedWithArgs(args []any) (any, error) {
	return m.proceedFunc()
}
func (m *mockJoinPointForPanicTest) Context() context.Context { return context.Background() }
func (m *mockJoinPointForPanicTest) GetResult() any           { return nil }
func (m *mockJoinPointForPanicTest) GetError() error          { return nil }
func (m *mockJoinPointForPanicTest) SetResult(v any)          {}
func (m *mockJoinPointForPanicTest) SetError(err error)       {}
