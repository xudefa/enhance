package exception

import (
	"context"
	"fmt"
	"testing"
)

// TestErrorCodeExceptionResolver_WrappedBusinessError 验证被包装的 BusinessError 也能被解析（回归测试）。
//
// 背景：asErrorCode 使用直接类型断言，fmt.Errorf("wrap: %w", businessErr)
// 包装后的错误无法匹配，导致返回 500 而非业务错误码。
func TestErrorCodeExceptionResolver_WrappedBusinessError(t *testing.T) {
	t.Parallel()
	resolver := NewErrorCodeExceptionResolver()
	businessErr := New(ErrCodeNotFound).WithDetail("id", 42)
	wrapped := fmt.Errorf("wrap: %w", businessErr)

	if !resolver.Supports(wrapped) {
		t.Error("expected Supports to match wrapped business error")
	}

	resp := resolver.Resolve(context.Background(), wrapped)
	if resp == nil {
		t.Fatal("expected response for wrapped error")
	}
	if resp.Code != ErrCodeNotFound.Code {
		t.Errorf("expected code %d, got %d", ErrCodeNotFound.Code, resp.Code)
	}
	if resp.Message != ErrCodeNotFound.Message {
		t.Errorf("expected message %q, got %q", ErrCodeNotFound.Message, resp.Message)
	}
}

// TestErrorCodeExceptionResolver_WrappedErrorCode 验证被包装的 ErrorCode 值也能被解析（回归测试）。
func TestErrorCodeExceptionResolver_WrappedErrorCode(t *testing.T) {
	t.Parallel()
	resolver := NewErrorCodeExceptionResolver()
	wrapped := fmt.Errorf("wrap: %w", ErrCodeNotFound)

	if !resolver.Supports(wrapped) {
		t.Error("expected Supports to match wrapped ErrorCode")
	}

	resp := resolver.Resolve(context.Background(), wrapped)
	if resp == nil {
		t.Fatal("expected response for wrapped error")
	}
	if resp.Code != ErrCodeNotFound.Code {
		t.Errorf("expected code %d, got %d", ErrCodeNotFound.Code, resp.Code)
	}
}

// TestAsErrorCode_Wrapped 验证 asErrorCode 支持错误链。
func TestAsErrorCode_Wrapped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantOK   bool
	}{
		{"direct ErrorCode", ErrCodeNotFound, 404, true},
		{"wrapped ErrorCode", fmt.Errorf("wrap: %w", ErrCodeNotFound), 404, true},
		{"wrapped BusinessError", fmt.Errorf("wrap: %w", New(ErrCodeBadRequest)), 400, true},
		{"non error code", fmt.Errorf("plain error"), 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var code ErrorCode
			ok := asErrorCode(tt.err, &code)
			if ok != tt.wantOK {
				t.Errorf("asErrorCode(%v) ok = %v, want %v", tt.err, ok, tt.wantOK)
				return
			}
			if ok && code.Code != tt.wantCode {
				t.Errorf("expected code %d, got %d", tt.wantCode, code.Code)
			}
		})
	}
}
