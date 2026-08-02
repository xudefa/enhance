package exception

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"sync"
)

// DefaultExceptionHandler 默认异常处理器实现。
//
// DefaultExceptionHandler 是 ExceptionHandler 接口的主要实现，提供了完整的异常处理功能：
// - 使用解析器链处理不同类型的异常
// - 支持日志和监控集成
// - 自动写入 HTTP 响应
// - 支持基于类型的异常处理函数注册
type DefaultExceptionHandler struct {
	chain   *ResolverChain
	config  ExceptionHandlerConfig
	mu      sync.RWMutex
	typeMap map[reflect.Type]func(ctx context.Context, err error) *ErrorResponse
}

// NewDefaultExceptionHandler 创建默认异常处理器。
//
// 支持通过 Option 函数配置日志记录器、指标记录器和堆栈跟踪选项。
// 默认情况下，会注册内置异常解析器和默认异常解析器。
//
// 示例：
//
//	handler := exception.NewDefaultExceptionHandler(
//		exception.WithLogger(logger),
//		exception.WithMetricsRecorder(metrics),
//		exception.WithIncludeStackTrace(true),
//	)
func NewDefaultExceptionHandler(opts ...Option) ExceptionHandler {
	config := ExceptionHandlerConfig{
		IncludeStackTrace: false,
	}

	for _, opt := range opts {
		opt(&config)
	}

	chain := NewResolverChain()
	chain.AddResolver(NewBuiltinExceptionResolver())
	chain.AddResolver(NewDefaultExceptionResolver())

	return &DefaultExceptionHandler{
		chain:   chain,
		config:  config,
		typeMap: make(map[reflect.Type]func(ctx context.Context, err error) *ErrorResponse),
	}
}

// Handle 处理异常。
//
// Handle 方法是异常处理的核心方法，它会：
// 1. 检查异常是否为 nil
// 2. 查找类型匹配的处理函数
// 3. 使用解析器链查找合适的解析器
// 4. 如果没有找到解析器，返回 500 错误
// 5. 将响应写入 ResponseWriter
// 6. 记录日志和指标（如果配置）
func (h *DefaultExceptionHandler) Handle(ctx context.Context, err error, response ResponseWriter) *ErrorResponse {
	if err == nil {
		return nil
	}

	var resp *ErrorResponse

	// 复制handler，避免在锁内执行用户回调导致死锁
	h.mu.RLock()
	errType := reflect.TypeOf(err)
	handlerFunc, ok := h.typeMap[errType]
	if !ok {
		// 处理包装错误：遍历已注册类型，用 errors.As 匹配错误链，
		// 否则 fmt.Errorf("wrap: %w") 等包装错误永远匹配不到处理函数
		errorIface := reflect.TypeOf((*error)(nil)).Elem()
		for t, fn := range h.typeMap {
			if !t.Implements(errorIface) && t.Kind() != reflect.Interface {
				continue
			}
			if errors.As(err, reflect.New(t).Interface()) {
				handlerFunc = fn
				break
			}
		}
	}
	h.mu.RUnlock()

	if handlerFunc != nil {
		resp = handlerFunc(ctx, err)
	}

	if resp == nil {
		resp = h.chain.Resolve(ctx, err)
	}

	if resp == nil {
		resp = NewErrorResponse(500, "Internal Server Error", "", "", nil)
	}

	if response != nil {
		data, marshalErr := json.Marshal(resp)
		if marshalErr != nil {
			// JSON 序列化失败时回退为纯文本 500，并记录日志
			response.SetStatusCode(http.StatusInternalServerError)
			response.SetHeader("Content-Type", "text/plain; charset=utf-8")
			if writeErr := response.Write([]byte("Internal Server Error")); writeErr != nil {
				return resp
			}
			if h.config.Logger != nil {
				h.config.Logger.Error(ctx, "failed to marshal error response",
					KeyValue{Key: "error", Value: marshalErr.Error()},
				)
			}
			return resp
		}
		// 钳制状态码到合法 HTTP 范围 [100, 999]，防止 net/http 的 WriteHeader panic
		statusCode := resp.Code
		if statusCode < 100 || statusCode > 999 {
			statusCode = http.StatusInternalServerError
		}
		response.SetStatusCode(statusCode)
		response.SetHeader("Content-Type", "application/json")
		if writeErr := response.Write(data); writeErr != nil {
			return resp
		}
	}

	if h.config.Logger != nil {
		h.config.Logger.Error(ctx, "exception handled",
			KeyValue{Key: "exception_type", Value: reflect.TypeOf(err).String()},
			KeyValue{Key: "message", Value: resp.Message},
			KeyValue{Key: "code", Value: resp.Code},
		)
	}

	if h.config.MetricsRecorder != nil {
		h.config.MetricsRecorder.RecordException(reflect.TypeOf(err).String(), resp.Code)
	}

	return resp
}

// IncludeStackTrace 返回是否在错误响应中包含堆栈跟踪信息。
//
// 供 ExceptionHandlingMiddleware 判断 panic 恢复时是否附加堆栈跟踪，
// 生产环境默认关闭，避免泄露源码路径等敏感信息。
func (h *DefaultExceptionHandler) IncludeStackTrace() bool {
	return h.config.IncludeStackTrace
}

// RegisterResolver 注册解析器。
//
// 将自定义解析器添加到解析器链中，解析器会自动按 Order 排序。
// 优先级数值越小，优先级越高。
func (h *DefaultExceptionHandler) RegisterResolver(resolver ExceptionResolver) {
	h.chain.AddResolver(resolver)
}

// RegisterException 注册异常类型和解析器。
//
// 为特定异常类型注册专用解析器，与 RegisterHandlerFunc 等效：
// 注册后该异常类型及其包装错误会优先生成响应，而非加入解析器链。
func (h *DefaultExceptionHandler) RegisterException(exceptionType reflect.Type, resolver ExceptionResolver) {
	h.mu.Lock()
	h.typeMap[exceptionType] = func(ctx context.Context, err error) *ErrorResponse {
		return resolver.Resolve(ctx, err)
	}
	h.mu.Unlock()
}

// RegisterHandlerFunc 注册异常类型和处理函数。
//
// 为特定异常类型注册处理函数，该函数会优先于解析器链执行。
// 这是一种更直接、更高效的异常处理方式。
func (h *DefaultExceptionHandler) RegisterHandlerFunc(exceptionType reflect.Type, handler func(ctx context.Context, err error) *ErrorResponse) {
	h.mu.Lock()
	h.typeMap[exceptionType] = handler
	h.mu.Unlock()
}

// Option 配置选项。
//
// Option 是用于配置 ExceptionHandler 的函数类型。
type Option func(*ExceptionHandlerConfig)

// WithLogger 设置日志记录器。
//
// 为异常处理器配置日志记录器，所有异常处理都会被记录。
func WithLogger(logger Logger) Option {
	return func(c *ExceptionHandlerConfig) {
		c.Logger = logger
	}
}

// WithMetricsRecorder 设置指标记录器。
//
// 为异常处理器配置指标记录器，所有异常都会被记录为指标。
func WithMetricsRecorder(recorder MetricsRecorder) Option {
	return func(c *ExceptionHandlerConfig) {
		c.MetricsRecorder = recorder
	}
}

// WithIncludeStackTrace 设置是否包含堆栈信息。
//
// 配置是否在错误响应中包含堆栈跟踪信息。
// 注意：生产环境中建议关闭此选项以避免泄露敏感信息。
func WithIncludeStackTrace(include bool) Option {
	return func(c *ExceptionHandlerConfig) {
		c.IncludeStackTrace = include
	}
}
