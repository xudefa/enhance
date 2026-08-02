// Package exception 提供异常处理和错误响应功能，用于 enhance 框架。
package exception

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

func (e ErrorCode) Error() string {
	return fmt.Sprintf("[%d] %s (%s)", e.Code, e.Message, e.Detail)
}

// WithMessage 创建带有自定义消息的错误码。
func (e ErrorCode) WithMessage(msg string) ErrorCode {
	return ErrorCode{
		Code:    e.Code,
		Message: msg,
		Detail:  e.Detail,
	}
}

// WithDetail 创建带有自定义详情的错误码。
func (e ErrorCode) WithDetail(detail string) ErrorCode {
	return ErrorCode{
		Code:    e.Code,
		Message: e.Message,
		Detail:  detail,
	}
}

// NewErrorCodeRegistry 创建错误码注册表
func NewErrorCodeRegistry() *ErrorCodeRegistry {
	return &ErrorCodeRegistry{}
}

// Register 注册错误码
func (r *ErrorCodeRegistry) Register(code ErrorCode) {
	r.codes.Store(code.Detail, code)
}

// Get 根据开发者调试信息获取错误码
func (r *ErrorCodeRegistry) Get(detail string) (ErrorCode, bool) {
	v, ok := r.codes.Load(detail)
	if !ok {
		return ErrorCode{}, false
	}
	ec, _ := v.(ErrorCode)
	return ec, true
}

// MustGet 根据开发者调试信息获取错误码，不存在则 panic
func (r *ErrorCodeRegistry) MustGet(detail string) ErrorCode {
	code, ok := r.Get(detail)
	if !ok {
		panic(fmt.Sprintf("error code not registered: %s", detail))
	}
	return code
}

// GetAll 获取所有注册的错误码
func (r *ErrorCodeRegistry) GetAll() []ErrorCode {
	var codes []ErrorCode
	r.codes.Range(func(key, value any) bool {
		ec, _ := value.(ErrorCode)
		codes = append(codes, ec)
		return true
	})
	return codes
}

// 全局错误码注册表
var globalErrorCodeRegistry = NewErrorCodeRegistry()

// GlobalErrorCodeRegistry 返回全局错误码注册表
func GlobalErrorCodeRegistry() *ErrorCodeRegistry {
	return globalErrorCodeRegistry
}

// RegisterErrorCode 注册错误码到全局注册表
func RegisterErrorCode(code ErrorCode) {
	globalErrorCodeRegistry.Register(code)
}

// GetErrorCode 从全局注册表获取错误码
func GetErrorCode(detail string) (ErrorCode, bool) {
	return globalErrorCodeRegistry.Get(detail)
}

// ErrorCodeExceptionResolver 错误码异常解析器
//
// 将 ErrorCode 类型的错误转换为统一的错误响应。
type ErrorCodeExceptionResolver struct {
	order int
}

// NewErrorCodeExceptionResolver 创建错误码异常解析器
func NewErrorCodeExceptionResolver() *ErrorCodeExceptionResolver {
	return &ErrorCodeExceptionResolver{
		order: 50,
	}
}

// Resolve 解析错误码并返回错误响应
func (r *ErrorCodeExceptionResolver) Resolve(ctx context.Context, err error) *ErrorResponse {
	var codeErr ErrorCode
	if !asErrorCode(err, &codeErr) {
		return nil
	}

	// 使用 NewErrorResponse 统一构建响应，确保 Timestamp 被正确填充
	return NewErrorResponse(codeErr.Code, codeErr.Message, "", "", codeErr.Detail)
}

// Supports 判断是否能处理该错误
func (r *ErrorCodeExceptionResolver) Supports(err error) bool {
	var codeErr ErrorCode
	return asErrorCode(err, &codeErr)
}

// Order 返回解析器优先级
func (r *ErrorCodeExceptionResolver) Order() int {
	return r.order
}

// asErrorCode 尝试将错误转换为 ErrorCode
//
// 使用 errors.As 沿错误链查找，使 fmt.Errorf("wrap: %w", err) 等包装错误也能被识别。
func asErrorCode(err error, target *ErrorCode) bool {
	if err == nil {
		return false
	}

	// 匹配 ErrorCode 值类型（ErrorCode 本身实现了 error 接口）
	var code ErrorCode
	if errors.As(err, &code) {
		*target = code
		return true
	}

	// 匹配实现了 ErrorCode() ErrorCode 接口的类型（如 *BusinessError）
	type errorCodeInterface interface {
		ErrorCode() ErrorCode
	}
	var iface errorCodeInterface
	if errors.As(err, &iface) {
		*target = iface.ErrorCode()
		return true
	}

	return false
}

// BusinessError 业务错误
//
// 包装 ErrorCode 并支持附加详细信息。
type BusinessError struct {
	mu      sync.Mutex
	code    ErrorCode
	details map[string]any
}

// New 创建业务错误
func New(code ErrorCode) *BusinessError {
	return &BusinessError{
		code:    code,
		details: make(map[string]any),
	}
}

// WithDetail 添加详细信息
func (e *BusinessError) WithDetail(key string, value any) *BusinessError {
	e.mu.Lock()
	e.details[key] = value
	e.mu.Unlock()
	return e
}

// WithDetails 批量添加详细信息
func (e *BusinessError) WithDetails(details map[string]any) *BusinessError {
	e.mu.Lock()
	for k, v := range details {
		e.details[k] = v
	}
	e.mu.Unlock()
	return e
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.details) == 0 {
		return e.code.Error()
	}
	return fmt.Sprintf("[%d] %s (%s) details=%v", e.code.Code, e.code.Message, e.code.Detail, e.details)
}

// GetDetails 获取详细信息
func (e *BusinessError) GetDetails() map[string]any {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make(map[string]any, len(e.details))
	for k, v := range e.details {
		result[k] = v
	}
	return result
}

// ErrorCode 返回错误码
func (e *BusinessError) ErrorCode() ErrorCode {
	return e.code
}

// 预定义的业务错误码（不与其他包冲突）
var (
	ErrCodeBadRequest          = ErrorCode{400, "请求参数错误", "bad_request"}
	ErrCodeUnauthorized        = ErrorCode{401, "未授权", "unauthorized"}
	ErrCodeForbidden           = ErrorCode{403, "禁止访问", "forbidden"}
	ErrCodeNotFound            = ErrorCode{404, "资源不存在", "not_found"}
	ErrCodeMethodNotAllowed    = ErrorCode{405, "方法不允许", "method_not_allowed"}
	ErrCodeConflict            = ErrorCode{409, "资源冲突", "conflict"}
	ErrCodeInternalServerError = ErrorCode{500, "服务器内部错误", "internal_server_error"}
	ErrCodeServiceUnavailable  = ErrorCode{503, "服务不可用", "service_unavailable"}
)

func init() {
	RegisterErrorCode(ErrCodeBadRequest)
	RegisterErrorCode(ErrCodeUnauthorized)
	RegisterErrorCode(ErrCodeForbidden)
	RegisterErrorCode(ErrCodeNotFound)
	RegisterErrorCode(ErrCodeMethodNotAllowed)
	RegisterErrorCode(ErrCodeConflict)
	RegisterErrorCode(ErrCodeInternalServerError)
	RegisterErrorCode(ErrCodeServiceUnavailable)
}
