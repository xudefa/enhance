package exception

import (
	"context"
	"errors"
	"testing"
)

// TestErrorCodeRegistry_MustGet_Coverage 测试 MustGet 方法
func TestErrorCodeRegistry_MustGet_Coverage(t *testing.T) {
	t.Parallel()

	registry := NewErrorCodeRegistry()
	registry.Register(ErrorCode{
		Code:    400,
		Message: "Test error",
		Detail:  "test.error",
	})

	// 测试正常获取
	code := registry.MustGet("test.error")
	if code.Code != 400 {
		t.Errorf("Expected 400, got %d", code.Code)
	}

	// 测试不存在时 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for unregistered error code")
		}
	}()
	registry.MustGet("nonexistent")
}

// TestErrorCodeRegistry_GetAll_Coverage 测试 GetAll 方法
func TestErrorCodeRegistry_GetAll_Coverage(t *testing.T) {
	t.Parallel()

	registry := NewErrorCodeRegistry()
	registry.Register(ErrorCode{
		Code:    400,
		Message: "Test error 1",
		Detail:  "test.error1",
	})
	registry.Register(ErrorCode{
		Code:    500,
		Message: "Test error 2",
		Detail:  "test.error2",
	})

	codes := registry.GetAll()
	if len(codes) != 2 {
		t.Errorf("Expected 2 codes, got %d", len(codes))
	}
}

// TestGlobalErrorCodeRegistry_Coverage 测试全局错误码注册表
func TestGlobalErrorCodeRegistry_Coverage(t *testing.T) {
	t.Parallel()

	registry := GlobalErrorCodeRegistry()
	if registry == nil {
		t.Fatal("Expected non-nil global registry")
	}
}

// TestRegisterErrorCode_Coverage 测试注册错误码
func TestRegisterErrorCode_Coverage(t *testing.T) {
	t.Parallel()

	code := ErrorCode{
		Code:    400,
		Message: "Global test error",
		Detail:  "global.test",
	}

	RegisterErrorCode(code)

	retrieved, ok := GetErrorCode("global.test")
	if !ok {
		t.Fatal("Expected to find registered error code")
	}
	if retrieved.Code != 400 {
		t.Errorf("Expected 400, got %d", retrieved.Code)
	}
}

// TestGetErrorCode_NotFound_Coverage 测试获取不存在的错误码
func TestGetErrorCode_NotFound_Coverage(t *testing.T) {
	t.Parallel()

	_, ok := GetErrorCode("nonexistent.error")
	if ok {
		t.Error("Expected false for unregistered error code")
	}
}

// TestErrorCodeExceptionResolver_Coverage 测试错误码解析器（补充测试）
func TestErrorCodeExceptionResolver_Coverage(t *testing.T) {
	t.Parallel()

	resolver := NewErrorCodeExceptionResolver()

	if resolver.Order() != 50 {
		t.Errorf("Expected order 50, got %d", resolver.Order())
	}

	// 测试 Supports 方法
	code := ErrorCode{
		Code:    400,
		Message: "Resolver test error",
		Detail:  "resolver.test",
	}
	businessErr := New(code)

	if !resolver.Supports(businessErr) {
		t.Error("Expected resolver to support ErrorCode")
	}

	// 测试不支持的错误
	regularErr := errors.New("regular error")
	if resolver.Supports(regularErr) {
		t.Error("Expected resolver to not support regular error")
	}

	// 测试 Resolve 方法
	ctx := context.Background()
	response := resolver.Resolve(ctx, businessErr)
	if response == nil {
		t.Fatal("Expected non-nil response")
	}
	if response.Code != 400 {
		t.Errorf("Expected 400, got %d", response.Code)
	}

	// 测试 Resolve 不支持的错误
	response = resolver.Resolve(ctx, regularErr)
	if response != nil {
		t.Error("Expected nil response for unsupported error")
	}
}

// TestBusinessError_WithDetail_Coverage 测试 WithDetail 方法
func TestBusinessError_WithDetail_Coverage(t *testing.T) {
	t.Parallel()

	code := ErrorCode{
		Code:    400,
		Message: "Detail test error",
		Detail:  "detail.test",
	}

	err := New(code).
		WithDetail("key1", "value1").
		WithDetail("key2", 123)

	if err.code.Code != 400 {
		t.Errorf("Expected 400, got %d", err.code.Code)
	}
}

// TestBusinessError_WithDetails_Coverage 测试 WithDetails 方法
func TestBusinessError_WithDetails_Coverage(t *testing.T) {
	t.Parallel()

	code := ErrorCode{
		Code:    400,
		Message: "Details test error",
		Detail:  "details.test",
	}

	details := map[string]any{
		"field1": "value1",
		"field2": 456,
	}

	err := New(code).WithDetails(details)

	if err.code.Code != 400 {
		t.Errorf("Expected 400, got %d", err.code.Code)
	}
}

// TestBusinessError_ErrorMethod_Coverage 测试 Error 方法
func TestBusinessError_ErrorMethod_Coverage(t *testing.T) {
	t.Parallel()

	code := ErrorCode{
		Code:    400,
		Message: "Error test message",
		Detail:  "error.test",
	}

	err := New(code)
	// Error() 方法返回格式化的错误信息
	if len(err.Error()) == 0 {
		t.Error("Expected non-empty error message")
	}
}

// TestBusinessError_ErrorMethod_WithDetails_Coverage 测试带详细信息的 Error 方法
func TestBusinessError_ErrorMethod_WithDetails_Coverage(t *testing.T) {
	t.Parallel()

	code := ErrorCode{
		Code:    400,
		Message: "Error with details",
		Detail:  "error.details",
	}

	err := New(code).WithDetail("key", "value")
	// 应该包含 details
	errMsg := err.Error()
	if len(errMsg) == 0 {
		t.Error("Expected non-empty error message")
	}
}

// TestBusinessError_ErrorCode_Coverage 测试 ErrorCode 方法
func TestBusinessError_ErrorCode_Coverage(t *testing.T) {
	t.Parallel()

	code := ErrorCode{
		Code:    400,
		Message: "Unwrap test message",
		Detail:  "unwrap.test",
	}

	err := New(code)

	// 使用 ErrorCode() 方法获取错误码
	errorCode := err.ErrorCode()
	if errorCode.Code != 400 {
		t.Errorf("Expected 400, got %d", errorCode.Code)
	}
	if errorCode.Message != "Unwrap test message" {
		t.Errorf("Expected 'Unwrap test message', got %s", errorCode.Message)
	}
}

// TestBusinessError_GetDetails_Coverage 测试 GetDetails 方法
func TestBusinessError_GetDetails_Coverage(t *testing.T) {
	t.Parallel()

	code := ErrorCode{
		Code:    400,
		Message: "Details test",
		Detail:  "details.test",
	}

	err := New(code).WithDetail("key1", "value1")

	details := err.GetDetails()
	if len(details) != 1 {
		t.Errorf("Expected 1 detail, got %d", len(details))
	}
	if details["key1"] != "value1" {
		t.Errorf("Expected value1, got %v", details["key1"])
	}
}

// TestNewErrorResponse_Builder_Coverage 测试 ErrorResponse 构建器
func TestNewErrorResponse_Builder_Coverage(t *testing.T) {
	t.Parallel()

	response := NewErrorResponseBuilder().
		Code(400).
		Message("Builder test message").
		RequestID("req-123").
		TraceID("trace-456").
		Details(map[string]any{"key": "value"}).
		Build()

	if response.Code != 400 {
		t.Errorf("Expected 400, got %d", response.Code)
	}
	if response.Message != "Builder test message" {
		t.Errorf("Expected 'Builder test message', got %s", response.Message)
	}
	if response.RequestID != "req-123" {
		t.Errorf("Expected req-123, got %s", response.RequestID)
	}
	if response.TraceID != "trace-456" {
		t.Errorf("Expected trace-456, got %s", response.TraceID)
	}
}

// TestNewErrorResponseBuilder_ToJSON_Coverage 测试 ToJSON 方法
func TestNewErrorResponseBuilder_ToJSON_Coverage(t *testing.T) {
	t.Parallel()

	jsonBytes, err := NewErrorResponseBuilder().
		Code(400).
		Message("JSON test").
		ToJSON()

	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Error("Expected non-empty JSON")
	}
}

// TestExceptionHandler_Builder_Coverage 测试 ExceptionHandler 构建器
func TestExceptionHandler_Builder_Coverage(t *testing.T) {
	t.Parallel()

	handler, err := NewExceptionHandlerBuilder().
		IncludeStackTrace(true).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}
