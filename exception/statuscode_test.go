package exception

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// recorderRW 将 httptest.ResponseRecorder 适配为 ResponseWriter 接口。
//
// WriteHeader 委托给 httptest.ResponseRecorder，非法状态码会触发其 panic，
// 用于验证处理器在写入非法 HTTP 状态码时不会再次 panic。
type recorderRW struct {
	rec *httptest.ResponseRecorder
}

func (w *recorderRW) SetStatusCode(code int) { w.rec.WriteHeader(code) }
func (w *recorderRW) SetHeader(k, v string)  { w.rec.Header().Set(k, v) }
func (w *recorderRW) Write(data []byte) error {
	_, err := w.rec.Write(data)
	return err
}
func (w *recorderRW) Context() context.Context { return context.Background() }

// outOfRangeErr 用于测试超出范围的状态码。
type outOfRangeErr struct{}

func (outOfRangeErr) Error() string { return "out of range error" }

// zeroCodeErr 用于测试零值状态码。
type zeroCodeErr struct{}

func (zeroCodeErr) Error() string { return "zero code error" }

// validCodeErr 用于测试有效状态码。
type validCodeErr struct{}

func (validCodeErr) Error() string { return "valid code error" }

// TestDefaultExceptionHandler_OutOfRangeStatusCode 超出 HTTP 范围的状态码应被钳制为 500。
//
// net/http 的 WriteHeader 对 <100 或 >999 的状态码会 panic，异常处理本身不能再次 panic。
func TestDefaultExceptionHandler_OutOfRangeStatusCode(t *testing.T) {
	t.Parallel()
	handler := NewDefaultExceptionHandler()

	handler.RegisterHandlerFunc(reflect.TypeOf(outOfRangeErr{}), func(ctx context.Context, err error) *ErrorResponse {
		return NewErrorResponse(99999, "custom business error", "", "", nil)
	})

	rec := httptest.NewRecorder()
	handler.Handle(context.Background(), outOfRangeErr{}, &recorderRW{rec: rec})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("out-of-range status code should be clamped to 500, got %d", rec.Code)
	}
}

// TestDefaultExceptionHandler_ZeroStatusCode 零值状态码（未设置）应被钳制为 500。
func TestDefaultExceptionHandler_ZeroStatusCode(t *testing.T) {
	t.Parallel()
	handler := NewDefaultExceptionHandler()

	handler.RegisterHandlerFunc(reflect.TypeOf(zeroCodeErr{}), func(ctx context.Context, err error) *ErrorResponse {
		return NewErrorResponse(0, "unset code", "", "", nil)
	})

	rec := httptest.NewRecorder()
	handler.Handle(context.Background(), zeroCodeErr{}, &recorderRW{rec: rec})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("zero status code should be clamped to 500, got %d", rec.Code)
	}
}

// TestDefaultExceptionHandler_ValidStatusCode 合法状态码原样透传。
func TestDefaultExceptionHandler_ValidStatusCode(t *testing.T) {
	t.Parallel()
	handler := NewDefaultExceptionHandler()

	handler.RegisterHandlerFunc(reflect.TypeOf(validCodeErr{}), func(ctx context.Context, err error) *ErrorResponse {
		return NewErrorResponse(422, "unprocessable entity", "", "", nil)
	})

	rec := httptest.NewRecorder()
	handler.Handle(context.Background(), validCodeErr{}, &recorderRW{rec: rec})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("valid status code should pass through, got %d", rec.Code)
	}
}
