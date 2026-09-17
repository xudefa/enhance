// Package exception 提供异常处理和错误响应功能，用于 enhance 框架。
package exception

import (
	"encoding/json"
	"time"
)

// ErrorResponseOption 错误响应的可选参数。
type ErrorResponseOption func(*errorResponseOptions)

// errorResponseOptions 保存错误响应的可选参数。
type errorResponseOptions struct {
	requestID string
	traceID   string
	details   any
}

// WithRequestID 设置请求 ID。
func WithRequestID(requestID string) ErrorResponseOption {
	return func(o *errorResponseOptions) {
		o.requestID = requestID
	}
}

// WithTraceID 设置追踪 ID。
func WithTraceID(traceID string) ErrorResponseOption {
	return func(o *errorResponseOptions) {
		o.traceID = traceID
	}
}

// WithDetails 设置详细信息（任意类型）。
func WithDetails(details any) ErrorResponseOption {
	return func(o *errorResponseOptions) {
		o.details = details
	}
}

// NewErrorResponse 创建新的错误响应
//
// 创建一个 ErrorResponse 实例，自动设置时间戳为当前时间。
//
// 参数：
//   - code: HTTP 状态码
//   - message: 错误消息
//   - opts: 可选参数（WithRequestID / WithTraceID / WithDetails）
//
// 返回一个初始化好的 ErrorResponse 实例。
func NewErrorResponse(code int, message string, opts ...ErrorResponseOption) *ErrorResponse {
	options := &errorResponseOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: options.requestID,
		TraceID:   options.traceID,
		Details:   options.details,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ToJSON 将错误响应转换为JSON字节
func (e *ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
