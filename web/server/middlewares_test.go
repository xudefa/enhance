package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xudefa/enhance/web/mvc"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()
	config := DefaultRequestIDConfig()
	middleware := RequestIDMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("RequestID middleware should set X-Request-ID header")
	}
}

func TestRequestIDMiddleware_ExistingID(t *testing.T) {
	t.Parallel()
	config := DefaultRequestIDConfig()
	middleware := RequestIDMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if w.Header().Get("X-Request-ID") != "existing-id" {
		t.Errorf("RequestID = %s, want existing-id", w.Header().Get("X-Request-ID"))
	}
}

func TestRequestIDMiddleware_Validate(t *testing.T) {
	t.Parallel()
	config := RequestIDConfig{}
	if err := config.Validate(); err == nil {
		t.Error("Validate() should error for empty config")
	}
}

func TestAccessLogMiddleware(t *testing.T) {
	t.Parallel()
	config := DefaultAccessLogConfig()
	middleware := AccessLogMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	// 不应该 panic 且应该调用 Next
}

func TestAccessLogMiddleware_SlowRequest(t *testing.T) {
	t.Parallel()
	config := AccessLogConfig{
		SlowThreshold: 1 * time.Millisecond,
		Logger:        DefaultAccessLogConfig().Logger,
	}
	middleware := AccessLogMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	// 添加慢速处理器
	ctx.WithMiddleware([]mvc.MiddlewareFunc{middleware}, func(ctx mvc.Context) {
		time.Sleep(10 * time.Millisecond)
	})
	ctx.Next()
}

func TestAccessLogMiddleware_Validate(t *testing.T) {
	t.Parallel()
	config := AccessLogConfig{}
	if err := config.Validate(); err == nil {
		t.Error("Validate() should error for nil logger")
	}
}

func TestErrorMiddleware(t *testing.T) {
	t.Parallel()
	config := DefaultErrorConfig()
	middleware := ErrorMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	ctx.WithMiddleware([]mvc.MiddlewareFunc{middleware}, func(ctx mvc.Context) {
		panic("test panic")
	})
	ctx.Next()

	if w.Code != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestErrorMiddleware_Validate(t *testing.T) {
	t.Parallel()
	config := ErrorConfig{}
	if err := config.Validate(); err == nil {
		t.Error("Validate() should error for nil logger")
	}
}

func TestCORSMiddleware(t *testing.T) {
	t.Parallel()
	config := DefaultCORSConfig()
	middleware := CORSMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("Allow-Origin = %s, want http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	t.Parallel()
	config := DefaultCORSConfig()
	middleware := CORSMiddleware(config)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if w.Code != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestCORSMiddleware_Validate(t *testing.T) {
	t.Parallel()
	config := CORSConfig{}
	if err := config.Validate(); err == nil {
		t.Error("Validate() should error for empty config")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()
	middleware := RecoveryMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	ctx.WithMiddleware([]mvc.MiddlewareFunc{middleware}, func(ctx mvc.Context) {
		panic("test panic")
	})
	ctx.Next()

	if w.Code != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()
	middleware := LoggingMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	// 不应该 panic
}

func TestGzipMiddleware(t *testing.T) {
	t.Parallel()
	middleware := GzipMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Content-Encoding = %s, want gzip", w.Header().Get("Content-Encoding"))
	}
}

func TestRealIPMiddleware(t *testing.T) {
	t.Parallel()
	middleware := RealIPMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "1.2.3.4")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if w.Header().Get("X-Real-IP") != "1.2.3.4" {
		t.Errorf("X-Real-IP = %s, want 1.2.3.4", w.Header().Get("X-Real-IP"))
	}
}

func TestRealIPMiddleware_XForwardedFor(t *testing.T) {
	t.Parallel()
	middleware := RealIPMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "5.6.7.8")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if w.Header().Get("X-Real-IP") != "5.6.7.8" {
		t.Errorf("X-Real-IP = %s, want 5.6.7.8", w.Header().Get("X-Real-IP"))
	}
}

func TestJoin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		parts []string
		want  string
	}{
		{[]string{"a", "b", "c"}, "a, b, c"},
		{[]string{"a"}, "a"},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		if got := join(tt.parts); got != tt.want {
			t.Errorf("join(%v) = %s, want %s", tt.parts, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello", "ell", true},
		{"hello", "xyz", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		if got := contains(tt.s, tt.substr); got != tt.want {
			t.Errorf("contains(%s, %s) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}
