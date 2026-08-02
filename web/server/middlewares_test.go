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

func TestCORSMiddleware_PreflightOnlyHeaders(t *testing.T) {
	t.Parallel()
	config := DefaultCORSConfig()
	middleware := CORSMiddleware(config)

	// 预检请求应返回 Allow-Methods/Allow-Headers/Max-Age
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)
	middleware(ctx)

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("preflight response should include Access-Control-Allow-Methods")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("preflight response should include Access-Control-Allow-Headers")
	}
	if w.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("preflight response should include Access-Control-Max-Age")
	}
}

func TestCORSMiddleware_ActualRequestNoPreflightHeaders(t *testing.T) {
	t.Parallel()
	config := DefaultCORSConfig()
	middleware := CORSMiddleware(config)

	// 实际请求不应携带仅预检所需的头
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)
	middleware(ctx)

	if w.Header().Get("Access-Control-Allow-Methods") != "" {
		t.Error("actual response should not include Access-Control-Allow-Methods")
	}
	if w.Header().Get("Access-Control-Max-Age") != "" {
		t.Error("actual response should not include Access-Control-Max-Age")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("Allow-Origin = %s, want http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_Preflight_DisallowedOriginRejected(t *testing.T) {
	t.Parallel()
	config := CORSConfig{
		AllowOrigins: []string{"http://allowed.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
	}
	middleware := CORSMiddleware(config)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)
	middleware(ctx)

	if w.Code != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d for disallowed origin", w.Code, http.StatusForbidden)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("disallowed origin should not receive Allow-Origin header")
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

	// GzipMiddleware 当前为 no-op，不设置 Content-Encoding 头
	if w.Header().Get("Content-Encoding") != "" {
		t.Errorf("Content-Encoding should be empty, got %s", w.Header().Get("Content-Encoding"))
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

	if req.Header.Get("X-Real-IP") != "1.2.3.4" {
		t.Errorf("X-Real-IP = %s, want 1.2.3.4", req.Header.Get("X-Real-IP"))
	}
	if w.Header().Get("X-Real-IP") != "" {
		t.Errorf("X-Real-IP should not leak to response header, got %s", w.Header().Get("X-Real-IP"))
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

	if req.Header.Get("X-Real-IP") != "5.6.7.8" {
		t.Errorf("X-Real-IP = %s, want 5.6.7.8", req.Header.Get("X-Real-IP"))
	}
}

func TestRealIPMiddleware_FallbackToRemoteAddr(t *testing.T) {
	t.Parallel()
	middleware := RealIPMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	ctx := NewContext(w, req)

	middleware(ctx)

	if req.Header.Get("X-Real-IP") != "10.0.0.1" {
		t.Errorf("X-Real-IP = %s, want 10.0.0.1", req.Header.Get("X-Real-IP"))
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
