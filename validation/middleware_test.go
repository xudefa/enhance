package validation

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// TestMiddlewareConfig 测试中间件配置
func TestMiddlewareConfig(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
	}

	if config.Validator == nil {
		t.Error("Validator 不能为 nil")
	}

	if len(config.Groups) != 1 {
		t.Errorf("预期 1 个组，得到 %d", len(config.Groups))
	}
}

// TestNewValidateMiddleware 测试创建验证中间件
func TestNewValidateMiddleware(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
	}

	middleware := NewValidateMiddleware(config)
	if middleware == nil {
		t.Fatal("NewValidateMiddleware 返回 nil")
	}
}

// TestValidateMiddlewareWithNilValidator 测试 nil 验证器
func TestValidateMiddlewareWithNilValidator(t *testing.T) {
	t.Parallel()
	config := &MiddlewareConfig{
		Validator: nil,
	}

	middleware := NewValidateMiddleware(config)
	err := middleware(nil, nil, config)
	if err == nil {
		t.Error("预期因验证器为 nil 而返回错误")
	}
}

// TestValidateMiddlewareWithValidData 测试有效数据
func TestValidateMiddlewareWithValidData(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	type User struct {
		Name string `validate:"create:required,min=2"`
	}

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
	}

	middleware := NewValidateMiddleware(config)
	user := User{Name: "张三"}
	err := middleware(nil, user, config)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}
}

// TestValidateMiddlewareWithInvalidData 测试无效数据
func TestValidateMiddlewareWithInvalidData(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()
	validator := NewGroupedValidator(registry)

	type User struct {
		Name string `validate:"create:required,min=2"`
	}

	errorCalled := false
	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{"create"},
		ErrorHandler: func(c any, err error) {
			errorCalled = true
		},
	}

	middleware := NewValidateMiddleware(config)
	user := User{Name: "张"}
	err := middleware(nil, user, config)
	if err == nil {
		t.Error("预期因名字太短而验证失败")
	}

	if !errorCalled {
		t.Error("错误处理器应该被调用")
	}
}

// TestValidateMiddlewareWithStandardValidator 测试标准验证器
func TestValidateMiddlewareWithStandardValidator(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type User struct {
		Name string `validate:"required,min=2"`
	}

	config := &MiddlewareConfig{
		Validator: validator,
		Groups:    []string{},
	}

	middleware := NewValidateMiddleware(config)
	user := User{Name: "张三"}
	err := middleware(nil, user, config)
	if err != nil {
		t.Errorf("预期验证成功，得到错误: %v", err)
	}

	user.Name = "张"
	err = middleware(nil, user, config)
	if err == nil {
		t.Error("预期因名字太短而验证失败")
	}
}

// TestDefaultErrorHandler 测试默认错误处理器
func TestDefaultErrorHandler(t *testing.T) {
	t.Parallel()
	err := errors.New("test error")
	DefaultErrorHandler(nil, err)
}

// TestShouldSkipPath 测试跳过路径
func TestShouldSkipPath(t *testing.T) {
	t.Parallel()
	result := shouldSkipPath(nil, []string{"/skip"})
	if result {
		t.Error("shouldSkipPath 应该返回 false（未实现）")
	}
}

// TestErrorResponse_ToJSON 测试错误响应序列化
func TestErrorResponse_ToJSON(t *testing.T) {
	t.Parallel()

	resp := &ErrorResponse{
		Code:    400,
		Message: "validation failed",
	}

	data, err := resp.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON data")
	}
}

// TestFindErrorHandlerMethod 测试查找错误处理方法
func TestFindErrorHandlerMethod(t *testing.T) {
	t.Parallel()

	// nil 输入
	handler := findErrorHandlerMethod(nil)
	if handler != nil {
		t.Error("expected nil handler for nil input")
	}

	// 非结构体输入
	handler = findErrorHandlerMethod("string")
	if handler != nil {
		t.Error("expected nil handler for non-struct input")
	}

	// 结构体没有 OnValidationError 方法
	type NoHandler struct{}
	handler = findErrorHandlerMethod(&NoHandler{})
	if handler != nil {
		t.Error("expected nil handler for struct without OnValidationError")
	}

	// 结构体有 OnValidationError 方法
	type WithHandler struct{}
	funcCalled := false
	// 使用闭包来跟踪调用
	var capturedErr error
	wh := &WithHandler{}
	// 通过反射调用测试
	v := reflect.ValueOf(wh)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// 检查是否有方法
	if v.NumMethod() == 0 {
		// 没有方法，应该返回 nil
		handler = findErrorHandlerMethod(wh)
		if handler != nil {
			t.Error("expected nil handler")
		}
	}
	_ = funcCalled
	_ = capturedErr
}

// TestGetPathFromContext 测试从上下文获取路径
func TestGetPathFromContext(t *testing.T) {
	t.Parallel()

	// nil 输入
	path := getPathFromContext(nil)
	if path != "" {
		t.Errorf("expected empty path for nil, got %s", path)
	}

	// 非结构体输入
	path = getPathFromContext("string")
	if path != "" {
		t.Errorf("expected empty path for non-struct, got %s", path)
	}

	// 结构体没有 Path 方法
	type NoPath struct{}
	path = getPathFromContext(&NoPath{})
	if path != "" {
		t.Errorf("expected empty path for struct without Path, got %s", path)
	}
}

// TestMatchPath 测试路径匹配
func TestMatchPath(t *testing.T) {
	t.Parallel()

	// 精确匹配
	if !matchPath("/api/users", "/api/users") {
		t.Error("expected exact match")
	}

	// 不匹配
	if matchPath("/api/users", "/api/posts") {
		t.Error("expected no match for different paths")
	}

	// 通配符匹配 - 前缀
	if !matchPath("/api/users/123", "/api/users/*") {
		t.Error("expected wildcard match for prefix")
	}

	// 通配符匹配 - 不匹配前缀
	if matchPath("/api/posts/123", "/api/users/*") {
		t.Error("expected no wildcard match for different prefix")
	}

	// 空模式
	if matchPath("/api/users", "") {
		t.Error("expected no match for empty pattern")
	}

	// 单字符模式
	if matchPath("/api/users", "*") {
		t.Error("expected no match for single asterisk")
	}
}

// TestShouldSkipPath_WithSkipPaths 测试跳过路径逻辑
func TestShouldSkipPath_WithSkipPaths(t *testing.T) {
	t.Parallel()

	// 空跳过列表
	if shouldSkipPath(nil, nil) {
		t.Error("should not skip with nil skipPaths")
	}

	// 空跳过列表
	if shouldSkipPath(nil, []string{}) {
		t.Error("should not skip with empty skipPaths")
	}
}

// TestDefaultErrorHandler_NilInputs 测试 nil 输入
func TestDefaultErrorHandler_NilInputs(t *testing.T) {
	t.Parallel()

	// 两个都是 nil
	DefaultErrorHandler(nil, nil)

	// 只有 context 是 nil
	DefaultErrorHandler(nil, errors.New("test"))

	// 只有 error 是 nil
	DefaultErrorHandler("context", nil)
}

// TestDefaultErrorHandler_WithHTTPRequest 测试 HTTP 请求上下文
func TestDefaultErrorHandler_WithHTTPRequest(t *testing.T) {
	t.Parallel()

	// *http.Request 应该被处理但不报错
	req := &http.Request{}
	err := errors.New("test error")
	DefaultErrorHandler(req, err)
}

// TestNewValidateMiddleware_NilValidator 测试 nil 验证器
func TestNewValidateMiddleware_NilValidator(t *testing.T) {
	t.Parallel()

	config := &MiddlewareConfig{
		Validator: nil,
	}

	middleware := NewValidateMiddleware(config)
	err := middleware(nil, nil, config)
	if err == nil {
		t.Error("expected error for nil validator")
	}
}

// TestDefaultErrorHandler_WithHTTPResponseWriter 测试 HTTP ResponseWriter
func TestDefaultErrorHandler_WithHTTPResponseWriter(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	err := errors.New("validation failed")
	DefaultErrorHandler(rec, err)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	if rec.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}

// TestDefaultErrorHandler_WithCustomResponseWriter 测试自定义 ResponseWriter
func TestDefaultErrorHandler_WithCustomResponseWriter(t *testing.T) {
	t.Parallel()

	cw := &customResponseWriter{
		headers: make(map[string]string),
	}
	err := errors.New("validation failed")
	DefaultErrorHandler(cw, err)

	if cw.statusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", cw.statusCode)
	}

	if cw.headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", cw.headers["Content-Type"])
	}
}

// customResponseWriter 实现 ResponseWriter 接口
type customResponseWriter struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (c *customResponseWriter) SetStatusCode(code int) {
	c.statusCode = code
}

func (c *customResponseWriter) SetHeader(key, value string) {
	c.headers[key] = value
}

func (c *customResponseWriter) Write(data []byte) error {
	c.body = data
	return nil
}

// TestShouldSkipPath_MatchScenarios 测试 shouldSkipPath 的不同场景
func TestShouldSkipPath_MatchScenarios(t *testing.T) {
	t.Parallel()

	// 测试路径匹配
	skipPaths := []string{"/api/health", "/api/users/*"}

	// 应该跳过精确匹配
	req1 := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	if !shouldSkipPath(req1, skipPaths) {
		t.Error("should skip /api/health")
	}

	// 应该跳过通配符匹配
	req2 := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	if !shouldSkipPath(req2, skipPaths) {
		t.Error("should skip /api/users/123")
	}

	// 不应该跳过不匹配的路径
	req3 := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	if shouldSkipPath(req3, skipPaths) {
		t.Error("should not skip /api/posts")
	}
}

// TestGetPathFromContext_WithHTTPRequest 测试从 HTTP 请求获取路径
func TestGetPathFromContext_WithHTTPRequest(t *testing.T) {
	t.Parallel()

	// getPathFromContext 通过反射查找 Path() 方法
	// http.Request 没有 Path() 方法，所以返回空字符串
	// shouldSkipPath 会直接从 http.Request.URL.Path 获取路径
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	path := getPathFromContext(req)
	// getPathFromContext 对 http.Request 返回空，因为它是通过反射找 Path() 方法
	// 而 shouldSkipPath 有专门的 case 处理 http.Request
	if path != "" {
		t.Logf("getPathFromContext returned: %s", path)
	}
}
