package exception

import (
	"testing"
)

// TestErrorCode 验证 ErrorCode 的基本功能:
//  1. 错误码定义
//  2. Error() 方法
//  3. WithMessage 和 WithDetail
func TestErrorCode(t *testing.T) {
	t.Parallel()
	code := ErrorCode{
		Code:    404,
		Message: "用户不存在",
		Detail:  "user_not_found",
	}

	// 测试 Error() 方法
	expected := "[404] 用户不存在 (user_not_found)"
	if code.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, code.Error())
	}

	// 测试 WithMessage
	newCode := code.WithMessage("自定义消息")
	if newCode.Message != "自定义消息" {
		t.Errorf("expected message '自定义消息', got '%s'", newCode.Message)
	}
	if newCode.Detail != "user_not_found" {
		t.Errorf("expected detail 'user_not_found', got '%s'", newCode.Detail)
	}

	// 测试 WithDetail
	newCode2 := code.WithDetail("id=123")
	if newCode2.Detail != "id=123" {
		t.Errorf("expected detail 'id=123', got '%s'", newCode2.Detail)
	}
	if newCode2.Message != "用户不存在" {
		t.Errorf("expected message '用户不存在', got '%s'", newCode2.Message)
	}
}

// TestErrorCodeRegistry 验证错误码注册表功能:
//  1. 注册和查询
//  2. MustGet panic 行为
//  3. GetAll 获取所有
func TestErrorCodeRegistry(t *testing.T) {
	t.Parallel()
	registry := NewErrorCodeRegistry()

	code1 := ErrorCode{400, "参数错误", "invalid_param"}
	code2 := ErrorCode{500, "服务器错误", "server_error"}

	// 测试注册和查询
	registry.Register(code1)
	registry.Register(code2)

	got, ok := registry.Get("invalid_param")
	if !ok {
		t.Error("expected to find 'invalid_param'")
	}
	if got.Message != "参数错误" {
		t.Errorf("expected message '参数错误', got '%s'", got.Message)
	}

	// 测试不存在的错误码
	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("expected not to find 'nonexistent'")
	}

	// 测试 MustGet panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustGet to panic for nonexistent code")
		}
	}()
	_ = registry.MustGet("nonexistent")
}

// TestBusinessError 验证 BusinessError 功能:
//  1. 创建业务错误
//  2. 添加详细信息
//  3. Error() 方法
//  4. ErrorCode() 方法
func TestBusinessError(t *testing.T) {
	t.Parallel()
	code := ErrorCode{404, "用户不存在", "user_not_found"}

	// 测试创建业务错误
	bizErr := New(code)
	returnedCode := bizErr.ErrorCode()
	if returnedCode.Code != 404 {
		t.Errorf("expected code 404, got %d", returnedCode.Code)
	}

	// 测试添加详细信息
	_ = bizErr.WithDetail("id", "123").WithDetail("type", "user")
	details := bizErr.GetDetails()
	if details["id"] != "123" {
		t.Errorf("expected detail id='123', got '%v'", details["id"])
	}
	if details["type"] != "user" {
		t.Errorf("expected detail type='user', got '%v'", details["type"])
	}

	// 测试 Error() 方法
	errMsg := bizErr.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}

	// 测试 ErrorCode() 方法
	if returnedCode.Detail != "user_not_found" {
		t.Errorf("expected detail 'user_not_found', got '%s'", returnedCode.Detail)
	}
}

// TestPredefinedErrorCodes 验证预定义错误码:
//  1. 所有预定义错误码都已注册
//  2. 可以从全局注册表获取
func TestPredefinedErrorCodes(t *testing.T) {
	t.Parallel()
	predefinedCodes := []ErrorCode{
		ErrCodeBadRequest,
		ErrCodeUnauthorized,
		ErrCodeForbidden,
		ErrCodeNotFound,
		ErrCodeMethodNotAllowed,
		ErrCodeConflict,
		ErrCodeInternalServerError,
		ErrCodeServiceUnavailable,
	}

	for _, code := range predefinedCodes {
		got, ok := GetErrorCode(code.Detail)
		if !ok {
			t.Errorf("expected to find code '%s' in global registry", code.Detail)
		}
		if got.Code != code.Code {
			t.Errorf("expected code %d, got %d for '%s'", code.Code, got.Code, code.Detail)
		}
	}
}

// TestGlobalRegistry 验证全局注册表功能:
//  1. 注册自定义错误码
//  2. 从全局注册表查询
func TestGlobalRegistry(t *testing.T) {
	t.Parallel()
	customCode := ErrorCode{418, "我是茶壶", "im_a_teapot"}
	RegisterErrorCode(customCode)

	got, ok := GetErrorCode("im_a_teapot")
	if !ok {
		t.Error("expected to find custom code in global registry")
	}
	if got.Code != 418 {
		t.Errorf("expected code 418, got %d", got.Code)
	}
}

// TestErrorCodeExceptionResolver 验证错误码异常解析器:
//  1. 支持 ErrorCode 类型
//  2. 解析生成正确的 ErrorResponse
func TestErrorCodeExceptionResolver(t *testing.T) {
	t.Parallel()
	resolver := NewErrorCodeExceptionResolver()

	// 测试 Supports
	code := ErrorCode{404, "用户不存在", "user_not_found"}
	if !resolver.Supports(code) {
		t.Error("expected resolver to support ErrorCode")
	}

	// 测试 Resolve
	response := resolver.Resolve(nil, code)
	if response == nil || response.Code != 404 || response.Message != "用户不存在" {
		t.Fatalf("expected code 404 and message '用户不存在', got code=%v, message=%v", response != nil, response != nil)
	}
	if response.Details != "user_not_found" {
		t.Errorf("expected details 'user_not_found', got '%v'", response.Details)
	}

	// 测试不支持的错误类型
	if resolver.Supports(nil) {
		t.Error("expected resolver to not support nil error")
	}
}
