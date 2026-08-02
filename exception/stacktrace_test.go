package exception

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// recordingHandler 记录 panic 时传递给 Handle 的错误，用于断言堆栈跟踪行为。
type recordingHandler struct {
	err          error
	includeStack bool
}

func (h *recordingHandler) Handle(_ context.Context, err error, _ ResponseWriter) *ErrorResponse {
	h.err = err
	return nil
}

func (h *recordingHandler) RegisterResolver(_ ExceptionResolver) {}

func (h *recordingHandler) RegisterException(_ reflect.Type, _ ExceptionResolver) {}

func (h *recordingHandler) RegisterHandlerFunc(_ reflect.Type, _ func(context.Context, error) *ErrorResponse) {}

func (h *recordingHandler) IncludeStackTrace() bool { return h.includeStack }

func invokeMiddleware(t *testing.T, handler ExceptionHandler) {
	t.Helper()
	middleware := ExceptionHandlingMiddleware(handler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	middleware(nextHandler).ServeHTTP(w, req)
}

// TestExceptionHandlingMiddleware_Panic_NoStackTrace 默认配置不应泄露堆栈跟踪（回归测试）。
//
// 背景：middleware 无条件拼接 debug.Stack()，即使 IncludeStackTrace 默认为 false，
// 生产环境中会泄露源码路径等敏感信息。
func TestExceptionHandlingMiddleware_Panic_NoStackTrace(t *testing.T) {
	t.Parallel()
	rec := &recordingHandler{}
	invokeMiddleware(t, rec)

	if rec.err == nil {
		t.Fatal("expected error to be captured")
	}
	if strings.Contains(rec.err.Error(), "goroutine") {
		t.Errorf("默认配置下不应包含堆栈跟踪，got: %v", rec.err)
	}
}

// TestExceptionHandlingMiddleware_Panic_WithStackTrace IncludeStackTrace=true 时包含堆栈跟踪。
func TestExceptionHandlingMiddleware_Panic_WithStackTrace(t *testing.T) {
	t.Parallel()
	rec := &recordingHandler{includeStack: true}
	invokeMiddleware(t, rec)

	if rec.err == nil {
		t.Fatal("expected error to be captured")
	}
	if !strings.Contains(rec.err.Error(), "goroutine") {
		t.Errorf("IncludeStackTrace=true 时应包含堆栈跟踪，got: %v", rec.err)
	}
}

// TestDefaultExceptionHandler_IncludeStackTrace 验证默认异常处理器暴露堆栈跟踪配置。
func TestDefaultExceptionHandler_IncludeStackTrace(t *testing.T) {
	t.Parallel()

	handler := NewDefaultExceptionHandler()
	if handler.(*DefaultExceptionHandler).IncludeStackTrace() {
		t.Error("默认配置应为 false")
	}

	handler = NewDefaultExceptionHandler(WithIncludeStackTrace(true))
	if !handler.(*DefaultExceptionHandler).IncludeStackTrace() {
		t.Error("WithIncludeStackTrace(true) 后应为 true")
	}
}
