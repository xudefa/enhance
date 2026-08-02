package exception

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
)

// ExceptionHandlingMiddleware HTTP 异常处理中间件
//
// ExceptionHandlingMiddleware 是一个 HTTP 中间件，用于自动捕获和处理 HTTP 处理过程中的异常。
// 它会捕获 panic 和返回的异常，并通过 ExceptionHandler 处理它们。
//
// 使用示例：
//
//	handler := exception.NewDefaultExceptionHandler()
//	middleware := exception.ExceptionHandlingMiddleware(handler)
//	http.Handle("/", middleware(httpHandler))
func ExceptionHandlingMiddleware(handler ExceptionHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var committed bool
			aw := &httpResponseWriter{ResponseWriter: w, ctx: r.Context(), committed: &committed}
			cw := &committedWriter{ResponseWriter: w, committed: &committed}

			defer func() {
				if rec := recover(); rec != nil {
					// 处理器为 nil 时无法处理异常，兜底返回 500
					if handler == nil {
						if !committed {
							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						}
						return
					}

					// 下游已写入响应头/响应体，不能再次写入，避免重复响应
					if committed {
						return
					}

					err := newRecoveredError(rec, handler)

					// 嵌套 recover：异常处理过程中再次 panic 时兜底，避免向上冒泡
					defer func() { _ = recover() }()
					handler.Handle(r.Context(), err, aw)
				}
			}()

			// 传入包装后的 ResponseWriter，统一跟踪响应提交状态
			next.ServeHTTP(cw, r)
		})
	}
}

// stackTraceProvider 可选接口，暴露是否包含堆栈跟踪的配置。
//
// DefaultExceptionHandler 实现了该接口；自定义 ExceptionHandler 未实现时
// 默认不附加堆栈跟踪，避免生产环境泄露敏感信息。
type stackTraceProvider interface {
	IncludeStackTrace() bool
}

// newRecoveredError 构造 panic 恢复错误。
//
// 仅当处理器显式开启堆栈跟踪时才调用 debug.Stack()，
// 默认情况下错误只包含 panic 值本身。
func newRecoveredError(rec any, handler ExceptionHandler) error {
	if provider, ok := handler.(stackTraceProvider); ok && provider.IncludeStackTrace() {
		return fmt.Errorf("panic recovered: %v\n%s", rec, debug.Stack())
	}
	return fmt.Errorf("panic recovered: %v", rec)
}

// httpResponseWriter 适配 http.ResponseWriter 到 ResponseWriter 接口
//
// httpResponseWriter 是一个适配器，将标准库的 http.ResponseWriter 适配到包中的 ResponseWriter 接口。
// 这使得异常处理器可以与标准 HTTP 处理器无缝集成。
type httpResponseWriter struct {
	http.ResponseWriter
	ctx       context.Context
	committed *bool
}

// committedWriter 标准库 ResponseWriter 包装器，跟踪响应提交状态
//
// 兼容 http.ResponseWriter 接口，与 httpResponseWriter 共享提交状态，
// 用于检测下游处理器是否已写入响应。
type committedWriter struct {
	http.ResponseWriter
	committed *bool
}

// Context 返回请求上下文
func (w *httpResponseWriter) Context() context.Context {
	return w.ctx
}

// SetStatusCode 设置 HTTP 状态码
//
// 调用底层 http.ResponseWriter 的 WriteHeader 方法。
// 响应已提交时忽略，避免重复写入状态码。
func (w *httpResponseWriter) SetStatusCode(code int) {
	w.WriteHeader(code)
}

// SetHeader 设置 HTTP 头
//
// 调用底层 http.ResponseWriter 的 Header().Set 方法。
func (w *httpResponseWriter) SetHeader(key, value string) {
	w.Header().Set(key, value)
}

// Write 写入响应体
//
// 调用底层 http.ResponseWriter 的 Write 方法。
func (w *httpResponseWriter) Write(data []byte) error {
	_, err := w.ResponseWriter.Write(data)
	return err
}

// WriteHeader 写入响应头并标记响应已提交
//
// 响应已提交后再次调用直接忽略，防止产生重复状态码。
// committed 未设置时不做跟踪（兼容未显式初始化的用法）。
func (w *httpResponseWriter) WriteHeader(code int) {
	if w.committed != nil && *w.committed {
		return
	}
	if w.committed != nil {
		*w.committed = true
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write 写入响应体并标记响应已提交
func (w *committedWriter) Write(data []byte) (int, error) {
	*w.committed = true
	return w.ResponseWriter.Write(data)
}

// WriteHeader 写入响应头并标记响应已提交
//
// 响应已提交后再次调用直接忽略，防止产生重复状态码。
func (w *committedWriter) WriteHeader(code int) {
	if *w.committed {
		return
	}
	*w.committed = true
	w.ResponseWriter.WriteHeader(code)
}
