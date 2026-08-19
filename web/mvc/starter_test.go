package mvc

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
)

// mockServer 模拟 Server 实现
type mockServer struct {
	mu      sync.RWMutex
	started bool
	stopped bool
	handler any
}

func (m *mockServer) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.started = true
	return nil
}

func (m *mockServer) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopped = true
	return nil
}

func (m *mockServer) SetHandler(handler any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handler = handler
}

func (m *mockServer) Use(middleware any) {}

func (m *mockServer) isStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

func (m *mockServer) isStopped() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stopped
}

// mockRouter 模拟 Router 实现
type mockRouter struct {
	routes      []string
	middlewares []MiddlewareFunc
}

func (m *mockRouter) GET(path string, handler HandlerFunc) {
	m.routes = append(m.routes, "GET "+path)
}

func (m *mockRouter) POST(path string, handler HandlerFunc) {
	m.routes = append(m.routes, "POST "+path)
}

func (m *mockRouter) PUT(path string, handler HandlerFunc) {
	m.routes = append(m.routes, "PUT "+path)
}

func (m *mockRouter) DELETE(path string, handler HandlerFunc) {
	m.routes = append(m.routes, "DELETE "+path)
}

func (m *mockRouter) PATCH(path string, handler HandlerFunc) {
	m.routes = append(m.routes, "PATCH "+path)
}

func (m *mockRouter) Group(prefix string) Router {
	return &mockRouter{}
}

func (m *mockRouter) Use(middleware MiddlewareFunc) {
	m.middlewares = append(m.middlewares, middleware)
}

// mockController 模拟 Controller 实现
type mockController struct {
	routesRegistered bool
}

func (c *mockController) Routes(router Router) {
	c.routesRegistered = true
	router.GET("/test", func(ctx Context) {})
}

func TestWebStarter_Lifecycle(t *testing.T) {
	// 注意: 不使用 t.Parallel()，因为测试依赖全局的控制器注册表
	// 清除全局控制器注册
	ClearControllers()

	// 创建 mock 组件
	server := &mockServer{}
	router := &mockRouter{}
	controller := &mockController{}

	// 注册控制器（使用指针，确保 Routes 方法修改的是同一个实例）
	RegisterController(controller)

	// 创建 WebStarter
	starter := NewWebStarter(
		WithConfig(WebConfig{
			Port: 8080,
			Host: "localhost",
		}),
		WithServer(server),
		WithRouter(router),
	)

	// 测试 Name
	if starter.Name() != "web" {
		t.Errorf("expected name 'web', got '%s'", starter.Name())
	}

	// 测试 Dependencies
	deps := starter.Dependencies()
	if deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}

	// 测试 Configure
	err := starter.Configure(nil)
	if err != nil {
		t.Errorf("Configure failed: %v", err)
	}

	// 测试 Start
	err = starter.Start(nil)
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// 等待 goroutine 执行（server.Start 在后台 goroutine 中执行）
	time.Sleep(100 * time.Millisecond)

	// 验证服务器启动
	if !server.isStarted() {
		t.Error("server should be started")
	}

	// 验证控制器路由注册（Start 是同步执行路由注册的，不需要等待 goroutine）
	if !controller.routesRegistered {
		t.Error("controller routes should be registered")
	}

	// 验证路由数量
	if len(router.routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(router.routes))
	}

	// 测试 Stop
	err = starter.Stop(nil)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// 验证服务器停止
	if !server.isStopped() {
		t.Error("server should be stopped")
	}

	// 测试 GetCondition
	cond := starter.GetCondition()
	if cond != nil {
		t.Error("expected nil condition")
	}
}

func TestWebStarter_WithMiddleware(t *testing.T) {
	// 注意: 不使用 t.Parallel()，因为测试依赖全局的控制器注册表
	ClearControllers()

	server := &mockServer{}
	router := &mockRouter{}

	middleware := func(ctx Context) {
		ctx.Next()
	}

	starter := NewWebStarter(
		WithServer(server),
		WithRouter(router),
		WithMiddlewares([]MiddlewareFunc{middleware}),
	)

	// Start 会应用中间件
	_ = starter.Start(nil)
	time.Sleep(100 * time.Millisecond)

	// 验证中间件已注册
	if len(router.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(router.middlewares))
	}
}

func TestWebStarter_Validation(t *testing.T) {
	// 注意: 不使用 t.Parallel()，因为测试依赖全局的控制器注册表
	ClearControllers()

	// 测试缺少 router
	starter := NewWebStarter(
		WithServer(&mockServer{}),
	)

	err := starter.Start(nil)
	if err == nil {
		t.Error("expected error when router is nil")
	}

	// 测试缺少 server
	starter2 := NewWebStarter(
		WithRouter(&mockRouter{}),
	)

	err = starter2.Start(nil)
	if err == nil {
		t.Error("expected error when server is nil")
	}
}

func TestDefaultWebConfig(t *testing.T) {
	t.Parallel()
	config := DefaultWebConfig()

	if config.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Port)
	}

	if config.Host != "0.0.0.0" {
		t.Errorf("expected host '0.0.0.0', got '%s'", config.Host)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", config.Timeout)
	}
}

func TestControllerRegistration(t *testing.T) {
	t.Parallel()
	ClearControllers()

	ctrl1 := &mockController{}
	ctrl2 := &mockController{}

	RegisterController(ctrl1)
	RegisterController(ctrl2)

	controllers := GetControllers()
	if len(controllers) != 2 {
		t.Errorf("expected 2 controllers, got %d", len(controllers))
	}

	// 验证返回的 slice 是副本，但元素是相同的指针
	controllers2 := GetControllers()
	controllers2[0].(*mockController).routesRegistered = true
	// 修改 controllers2 会影响 controllers[0]，因为它们指向同一个对象
	if !controllers[0].(*mockController).routesRegistered {
		t.Error("controllers should point to same objects")
	}

	// 测试清除
	ClearControllers()
	if len(GetControllers()) != 0 {
		t.Error("ClearControllers should clear all controllers")
	}
}

// 测试 Context 接口实现
type mockContext struct {
	method     string
	uri        string
	params     map[string]string
	query      map[string]string
	headers    map[string]string
	statusCode int
	aborted    bool
	nextCalled bool
	ctx        context.Context
}

func (m *mockContext) RequestMethod() string {
	return m.method
}

func (m *mockContext) RequestURI() string {
	return m.uri
}

func (m *mockContext) PathParam(name string) string {
	return m.params[name]
}

func (m *mockContext) Query(name string) string {
	return m.query[name]
}

func (m *mockContext) QueryDefault(name, defaultVal string) string {
	if val, ok := m.query[name]; ok {
		return val
	}
	return defaultVal
}

func (m *mockContext) Header(key string) string {
	return m.headers[key]
}

func (m *mockContext) BindJSON(target any) error {
	return nil
}

func (m *mockContext) SetStatusCode(code int) {
	m.statusCode = code
}

func (m *mockContext) SetHeader(key, value string) {
	m.headers[key] = value
}

func (m *mockContext) JSON(code int, data any) error {
	m.statusCode = code
	return nil
}

func (m *mockContext) String(code int, format string, args ...any) {
	m.statusCode = code
}

func (m *mockContext) AbortWithStatus(code int) {
	m.aborted = true
	m.statusCode = code
}

func (m *mockContext) AbortWithStatusJSON(code int, body any) {
	m.aborted = true
	m.statusCode = code
}

func (m *mockContext) Next() {
	m.nextCalled = true
}

func (m *mockContext) IsAborted() bool {
	return m.aborted
}

func (m *mockContext) Context() context.Context {
	return m.ctx
}

func (m *mockContext) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func TestContextInterface(t *testing.T) {
	t.Parallel()
	ctx := &mockContext{
		method:  http.MethodGet,
		uri:     "/users/123",
		params:  map[string]string{"id": "123"},
		query:   map[string]string{"page": "1"},
		headers: map[string]string{"Content-Type": "application/json"},
	}

	// 测试基本方法
	if ctx.RequestMethod() != http.MethodGet {
		t.Errorf("expected GET, got %s", ctx.RequestMethod())
	}

	if ctx.RequestURI() != "/users/123" {
		t.Errorf("expected /users/123, got %s", ctx.RequestURI())
	}

	// 测试 PathParam
	if ctx.PathParam("id") != "123" {
		t.Errorf("expected id=123, got %s", ctx.PathParam("id"))
	}

	// 测试 Query
	if ctx.Query("page") != "1" {
		t.Errorf("expected page=1, got %s", ctx.Query("page"))
	}

	// 测试 QueryDefault
	if ctx.QueryDefault("size", "10") != "10" {
		t.Errorf("expected size=10, got %s", ctx.QueryDefault("size", "10"))
	}

	// 测试 Header
	if ctx.Header("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %s", ctx.Header("Content-Type"))
	}

	// 测试 SetStatusCode
	ctx.SetStatusCode(http.StatusOK)
	if ctx.statusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", ctx.statusCode)
	}

	// 测试 SetHeader
	ctx.SetHeader("X-Custom", "value")
	if ctx.headers["X-Custom"] != "value" {
		t.Errorf("expected X-Custom=value, got %s", ctx.headers["X-Custom"])
	}

	// 测试 Next
	ctx.Next()
	if !ctx.nextCalled {
		t.Error("next should be called")
	}

	// 测试 IsAborted
	if ctx.IsAborted() {
		t.Error("should not be aborted")
	}

	// 测试 AbortWithStatus
	ctx.AbortWithStatus(http.StatusBadRequest)
	if !ctx.IsAborted() {
		t.Error("should be aborted")
	}

	if ctx.statusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ctx.statusCode)
	}
}

// ==================== mockLogger ====================

type mockLogger struct{}

func (m *mockLogger) Debug(_ context.Context, _ string, _ ...log.KeyValue) {}
func (m *mockLogger) Info(_ context.Context, _ string, _ ...log.KeyValue)  {}
func (m *mockLogger) Warn(_ context.Context, _ string, _ ...log.KeyValue)  {}
func (m *mockLogger) Error(_ context.Context, _ string, _ ...log.KeyValue) {}
func (m *mockLogger) Sync() error                                         { return nil }
func (m *mockLogger) With(_ context.Context, _ ...log.KeyValue) log.Logger {
	return m
}

// ==================== errorMockServer ====================

type errorMockServer struct {
	handler any
}

func (e *errorMockServer) Start() error              { return nil }
func (e *errorMockServer) Stop(_ context.Context) error {
	return errors.New("server stop failed")
}
func (e *errorMockServer) SetHandler(handler any) { e.handler = handler }
func (e *errorMockServer) Use(_ any)              {}

// ==================== Option function tests ====================

func TestWithName(t *testing.T) {
	t.Parallel()
	s := NewWebStarter(WithName("my-app"))
	if s.name != "my-app" {
		t.Errorf("expected name 'my-app', got %q", s.name)
	}
}

func TestWithLogger(t *testing.T) {
	t.Parallel()
	logger := &mockLogger{}
	s := NewWebStarter(WithLogger(logger))
	if s.logger != logger {
		t.Error("expected logger to be set to the provided mockLogger")
	}
}

func TestWithHandler(t *testing.T) {
	t.Parallel()
	handler := http.NewServeMux()
	s := NewWebStarter(WithHandler(handler))
	if s.handler != handler {
		t.Error("expected handler to be set")
	}
}

// ==================== Setter tests ====================

func TestSetRouter(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	router := &mockRouter{}
	s.SetRouter(router)
	if s.router != router {
		t.Error("expected router to be set")
	}
}

func TestSetServer(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	server := &mockServer{}
	s.SetServer(server)
	if s.server != server {
		t.Error("expected server to be set")
	}
}

func TestSetHandler(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	handler := http.NewServeMux()
	s.SetHandler(handler)
	if s.handler != handler {
		t.Error("expected handler to be set")
	}
}

func TestSetMiddlewares(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	mws := []MiddlewareFunc{func(ctx Context) {}, func(ctx Context) {}}
	s.SetMiddlewares(mws)
	if len(s.middlewares) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(s.middlewares))
	}
}

func TestAddMiddleware(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	s.AddMiddleware(func(ctx Context) {})
	s.AddMiddleware(func(ctx Context) {})
	if len(s.middlewares) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(s.middlewares))
	}
}

// ==================== Start with custom handler ====================

func TestStartWithCustomHandler(t *testing.T) {
	ClearControllers()
	server := &mockServer{}
	router := &mockRouter{}
	handler := http.NewServeMux()

	starter := NewWebStarter(
		WithServer(server),
		WithRouter(router),
		WithHandler(handler),
	)

	if err := starter.Start(nil); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	server.mu.RLock()
	got := server.handler
	server.mu.RUnlock()

	if got != handler {
		t.Error("server.SetHandler should be called with custom handler, not router")
	}
}

// ==================== Start/Stop with nil context ====================

func TestStartWithNilContext(t *testing.T) {
	ClearControllers()
	server := &mockServer{}
	router := &mockRouter{}

	starter := NewWebStarter(
		WithServer(server),
		WithRouter(router),
	)

	if err := starter.Start(nil); err != nil {
		t.Fatalf("Start(nil) should not panic, got error: %v", err)
	}
}

func TestStopWithNilContext(t *testing.T) {
	ClearControllers()
	server := &mockServer{}
	router := &mockRouter{}

	starter := NewWebStarter(
		WithServer(server),
		WithRouter(router),
	)

	if err := starter.Stop(nil); err != nil {
		t.Fatalf("Stop(nil) should not panic, got error: %v", err)
	}
}

// ==================== Stop with error ====================

func TestStopWithError(t *testing.T) {
	ClearControllers()
	server := &errorMockServer{}
	router := &mockRouter{}

	starter := NewWebStarter(
		WithServer(server),
		WithRouter(router),
	)

	err := starter.Stop(nil)
	if err == nil {
		t.Fatal("expected error from Stop")
	}
	if err.Error() != "server stop failed" {
		t.Errorf("unexpected error: %v", err)
	}
}

// ==================== DefaultConfig ====================

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected host '0.0.0.0', got %q", cfg.Host)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.Logger == nil {
		t.Error("expected logger to be non-nil")
	}
}

// ==================== Chainable config methods ====================

func TestWebStarterWithServerChainable(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	server := &mockServer{}
	ret := s.WithServer(server)
	if s.server != server {
		t.Error("expected server to be set")
	}
	if ret != s {
		t.Error("WithServer should return *WebStarter for chaining")
	}
}

func TestWebStarterWithRouterChainable(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	router := &mockRouter{}
	ret := s.WithRouter(router)
	if s.router != router {
		t.Error("expected router to be set")
	}
	if ret != s {
		t.Error("WithRouter should return *WebStarter for chaining")
	}
}

func TestWebStarterUseChainable(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	ret := s.Use(func(ctx Context) {})
	if len(s.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(s.middlewares))
	}
	if ret != s {
		t.Error("Use should return *WebStarter for chaining")
	}
}

func TestWebStarterGetRouter(t *testing.T) {
	t.Parallel()
	s := NewWebStarter()
	if s.GetRouter() != nil {
		t.Error("expected nil router initially")
	}
	router := &mockRouter{}
	s.SetRouter(router)
	if s.GetRouter() != router {
		t.Error("expected router to match")
	}
}

func TestDefaultWebConfigAlias(t *testing.T) {
	t.Parallel()
	cfg := DefaultWebConfig()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
}
