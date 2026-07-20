// Package exception 提供异常处理和错误响应功能，用于 enhance 框架。
//
// 该模块提供统一的异常处理机制，包括错误响应格式、错误码管理、异常处理器等。
// 参考 Spring 的 @ExceptionHandler 设计。
//
// # 架构设计
//
//   - ExceptionHandler: 异常处理器接口，核心异常处理入口
//   - ExceptionResolver: 异常解析器接口，解析异常并生成统一错误响应
//   - Logger: 日志接口，抽象层不依赖具体日志实现
//   - MetricsRecorder: 指标记录器接口，用于记录异常指标到监控系统
//   - ResponseWriter: 响应写入器接口，适配不同 HTTP 框架的响应写入
//   - ErrorCode: 业务错误码，包含 HTTP 状态码、用户可见消息和开发者调试信息
//   - ErrorResponse: 统一错误响应格式
//
// # 核心功能
//
//   - 统一错误响应: 提供标准化的错误响应格式
//   - 错误码管理: 支持业务错误码定义和管理
//   - 异常处理: 支持全局异常捕获和处理
//   - 错误日志: 自动记录错误日志
//   - 安全过滤: 过滤敏感信息，防止信息泄露
//
// # 使用方式
//
// 定义错误码：
//
//	var (
//	    ErrUserNotFound     = exception.ErrorCode{404, "用户不存在", "user_not_found"}
//	    ErrInvalidParameter = exception.ErrorCode{400, "参数错误", "invalid_parameter"}
//	)
//
// 创建错误响应：
//
//	resp := exception.NewErrorResponse(404, "用户不存在", "", "", nil)
//
// 注册异常处理器：
//
//	handler := exception.NewExceptionHandler(config)
//	handler.RegisterResolver(resolver)
//
// # 错误响应格式
//
//	{
//	  "code": 404,
//	  "message": "用户不存在",
//	  "requestId": "req-123",
//	  "traceId": "trace-456",
//	  "details": {...},
//	  "timestamp": 1704067200000
//	}
package exception

import (
	"context"
	"reflect"
	"sync"
)

// Logger 日志接口。
//
// 抽象层，不依赖具体日志实现，支持 slog、zap 等多种日志框架适配。
type Logger interface {
	// Error 记录错误日志。
	Error(ctx context.Context, msg string, keyValues ...KeyValue)
}

// KeyValue 日志键值对。
type KeyValue struct {
	Key   string // 键名
	Value any    // 值
}

// MetricsRecorder 指标记录器接口。
//
// 抽象层，用于记录异常指标到监控系统。
type MetricsRecorder interface {
	// RecordException 记录一次异常事件。
	RecordException(exceptionType string, statusCode int)
}

// ResponseWriter 响应写入器接口。
//
// 抽象层，适配不同 HTTP 框架的响应写入。
type ResponseWriter interface {
	// SetStatusCode 设置 HTTP 状态码。
	SetStatusCode(code int)

	// SetHeader 设置响应头。
	SetHeader(key, value string)

	// Write 写入响应体。
	Write(data []byte) error
}

// ErrorResponse 统一错误响应。
//
// 所有异常处理最终都转换为此结构体，确保 API 错误响应格式一致。
type ErrorResponse struct {
	Code      int    `json:"code"`                // HTTP 状态码
	Message   string `json:"message"`             // 错误消息
	RequestID string `json:"requestId,omitempty"` // 请求 ID
	TraceID   string `json:"traceId,omitempty"`   // 链路追踪 ID
	Details   any    `json:"details,omitempty"`   // 错误详情
	Timestamp int64  `json:"timestamp"`           // 时间戳
}

// ExceptionHandler 异常处理器接口。
//
// 核心异常处理入口，支持注册自定义解析器和处理函数。
type ExceptionHandler interface {
	// Handle 处理异常并返回统一错误响应。
	Handle(ctx context.Context, err error, response ResponseWriter) *ErrorResponse

	// RegisterResolver 注册通用异常解析器。
	RegisterResolver(resolver ExceptionResolver)

	// RegisterException 为指定异常类型注册解析器。
	RegisterException(exceptionType reflect.Type, resolver ExceptionResolver)

	// RegisterHandlerFunc 为指定异常类型注册处理函数。
	RegisterHandlerFunc(exceptionType reflect.Type, handler func(ctx context.Context, err error) *ErrorResponse)
}

// ExceptionResolver 异常解析器接口。
//
// 解析异常并生成统一错误响应，支持优先级排序。
// 多个解析器按 Order() 返回值从小到大依次尝试。
type ExceptionResolver interface {
	// Resolve 解析异常并返回错误响应。
	Resolve(ctx context.Context, err error) *ErrorResponse

	// Supports 判断是否能处理该异常。
	Supports(err error) bool

	// Order 返回解析器优先级，值越小优先级越高。
	Order() int
}

// ExceptionHandlerConfig 异常处理器配置。
type ExceptionHandlerConfig struct {
	Logger            Logger          // 日志记录器
	MetricsRecorder   MetricsRecorder // 指标记录器
	IncludeStackTrace bool            // 是否在响应中包含堆栈跟踪
}

// ErrorCode 业务错误码。
//
// 定义统一的错误码体系，包含 HTTP 状态码、用户可见消息和开发者调试信息。
//
// 示例:
//
//	var (
//	    ErrUserNotFound     = ErrorCode{404, "用户不存在", "user_not_found"}
//	    ErrInvalidParameter = ErrorCode{400, "参数错误", "invalid_parameter"}
//	)
type ErrorCode struct {
	Code    int    // HTTP 状态码
	Message string // 用户可见消息
	Detail  string // 开发者调试信息
}

// ErrorCodeRegistry 错误码注册表。
//
// 管理所有业务错误码的注册和查询。
// 使用 sync.Map 优化读多写少场景的并发性能。
type ErrorCodeRegistry struct {
	codes sync.Map // map[string]ErrorCode
}

// 以下类型定义在其他文件中，此处仅作文档说明：
// - DefaultExceptionHandler: handler.go（包含完整实现）
// - SecurityAdapter: security_adapter.go（包含完整实现）
// - ErrorMiddleware: middleware.go（包含完整实现）
// - BuiltinResolver: builtin.go（包含完整实现）
// - ErrorLogger: logging.go（包含完整实现）
// - ExceptionHandlerBuilder: builder.go（包含完整实现）
