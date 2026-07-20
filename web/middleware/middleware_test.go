package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xudefa/enhance/log"
)

func TestRequestID(t *testing.T) {
	t.Parallel()
	mw := RequestID()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected request ID to be set")
	}
}

func TestRequestIDWithExistingID(t *testing.T) {
	t.Parallel()
	mw := RequestID()

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID != "existing-id" {
		t.Errorf("expected existing-id, got %s", requestID)
	}
}

func TestRequestIDCustomConfig(t *testing.T) {
	t.Parallel()
	config := RequestIDConfig{
		HeaderName: "X-Trace-ID",
		Generator: func() string {
			return "custom-id"
		},
	}

	mw := RequestID(config)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)

	traceID := rec.Header().Get("X-Trace-ID")
	if traceID != "custom-id" {
		t.Errorf("expected custom-id, got %s", traceID)
	}
}

func TestDefaultRequestIDConfig(t *testing.T) {
	t.Parallel()
	config := DefaultRequestIDConfig()

	if config.HeaderName != "X-Request-ID" {
		t.Errorf("expected X-Request-ID, got %s", config.HeaderName)
	}

	if config.Generator == nil {
		t.Error("expected generator to be set")
	}

	id := config.Generator()
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestAccessLog(t *testing.T) {
	t.Parallel()
	mw := AccessLog()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)
}

func TestAccessLogCustomConfig(t *testing.T) {
	t.Parallel()
	config := AccessLogConfig{
		SlowThreshold: 0,
		Logger:        log.Build(),
	}

	mw := AccessLog(config)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)
}

func TestDefaultAccessLogConfig(t *testing.T) {
	t.Parallel()
	config := DefaultAccessLogConfig()

	if config.SlowThreshold == 0 {
		t.Error("expected slow threshold to be set")
	}

	if config.Logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestError(t *testing.T) {
	t.Parallel()
	mw := Error()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)
}

func TestErrorRecover(t *testing.T) {
	t.Parallel()
	mw := Error()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Fatal("should not panic")
		}
	}()

	ctx := newMockContext(req, rec)
	ctx.panicOnNext = true

	mw(ctx)
}

func TestDefaultErrorConfig(t *testing.T) {
	t.Parallel()
	config := DefaultErrorConfig()

	if config.Logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestCORS(t *testing.T) {
	t.Parallel()
	mw := CORS()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected *, got %s", origin)
	}
}

func TestCORSOptions(t *testing.T) {
	t.Parallel()
	mw := CORS()

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected methods to be set")
	}
}

func TestCORSCustomConfig(t *testing.T) {
	t.Parallel()
	config := CORSConfig{
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
	}

	mw := CORS(config)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	ctx := newMockContext(req, rec)

	mw(ctx)

	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", origin)
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	t.Parallel()
	config := DefaultCORSConfig()

	if len(config.AllowOrigins) == 0 {
		t.Error("expected allow origins to be set")
	}

	if len(config.AllowMethods) == 0 {
		t.Error("expected allow methods to be set")
	}

	if len(config.AllowHeaders) == 0 {
		t.Error("expected allow headers to be set")
	}

	if !config.AllowCredentials {
		t.Error("expected allow credentials to be true")
	}

	if config.MaxAge == 0 {
		t.Error("expected max age to be set")
	}
}

func TestJoinStrings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: "",
		},
		{
			name:     "single element",
			input:    []string{"GET"},
			expected: "GET",
		},
		{
			name:     "multiple elements",
			input:    []string{"GET", "POST", "PUT"},
			expected: "GET, POST, PUT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// mockContext 模拟 Context 用于测试。
type mockContext struct {
	req         *http.Request
	res         http.ResponseWriter
	params      map[string]string
	panicOnNext bool
}

func newMockContext(req *http.Request, res http.ResponseWriter) *mockContext {
	return &mockContext{
		req:    req,
		res:    res,
		params: make(map[string]string),
	}
}

func (m *mockContext) RequestMethod() string {
	return m.req.Method
}

func (m *mockContext) RequestURI() string {
	return m.req.RequestURI
}

func (m *mockContext) PathParam(name string) string {
	return m.params[name]
}

func (m *mockContext) Query(name string) string {
	return m.req.URL.Query().Get(name)
}

func (m *mockContext) QueryDefault(name, defaultVal string) string {
	val := m.req.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	return val
}

func (m *mockContext) Header(key string) string {
	return m.req.Header.Get(key)
}

func (m *mockContext) BindJSON(target any) error {
	return nil
}

func (m *mockContext) SetStatusCode(code int) {
	m.res.WriteHeader(code)
}

func (m *mockContext) SetHeader(key, value string) {
	m.res.Header().Set(key, value)
}

func (m *mockContext) JSON(code int, data any) error {
	return nil
}

func (m *mockContext) String(code int, format string, args ...any) {
}

func (m *mockContext) AbortWithStatus(code int) {
	m.res.WriteHeader(code)
}

func (m *mockContext) AbortWithStatusJSON(code int, body any) {
	m.res.WriteHeader(code)
}

func (m *mockContext) Next() {
	if m.panicOnNext {
		panic("test panic")
	}
}

func (m *mockContext) IsAborted() bool {
	return false
}

func (m *mockContext) Context() context.Context {
	return m.req.Context()
}

func (m *mockContext) SetContext(ctx context.Context) {
	m.req = m.req.WithContext(ctx)
}

func BenchmarkRequestID(b *testing.B) {
	mw := RequestID()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := newMockContext(req, rec)
		mw(ctx)
	}
}

func BenchmarkAccessLog(b *testing.B) {
	mw := AccessLog()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := newMockContext(req, rec)
		mw(ctx)
	}
}

func BenchmarkError(b *testing.B) {
	mw := Error()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := newMockContext(req, rec)
		mw(ctx)
	}
}

func BenchmarkCORS(b *testing.B) {
	mw := CORS()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := newMockContext(req, rec)
		mw(ctx)
	}
}

func BenchmarkCORSOptions(b *testing.B) {
	mw := CORS()

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := newMockContext(req, rec)
		mw(ctx)
	}
}
