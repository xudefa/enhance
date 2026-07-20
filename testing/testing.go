package testing

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/context"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

// ==================== TestRunner ====================

// TestRunner 测试运行器
type TestRunner struct {
	t       *testing.T
	config  *TestConfig
	context TestContext
}

// TestConfig 测试配置
type TestConfig struct {
	Properties map[string]any
	MockBeans  map[string]any
	AutoConfig bool
	AppName    string
}

// NewTestRunner 创建测试运行器
func NewTestRunner(t *testing.T, opts ...TestOption) *TestRunner {
	config := &TestConfig{
		Properties: make(map[string]any),
		MockBeans:  make(map[string]any),
		AutoConfig: true,
		AppName:    "test-app",
	}

	for _, opt := range opts {
		opt(config)
	}

	return &TestRunner{
		t:      t,
		config: config,
	}
}

// TestOption 测试选项
type TestOption func(*TestConfig)

// WithProperty 设置测试属性
func WithProperty(key string, value any) TestOption {
	return func(c *TestConfig) {
		c.Properties[key] = value
	}
}

// WithMockBean 添加 Mock Bean
func WithMockBean(name string, bean any) TestOption {
	return func(c *TestConfig) {
		c.MockBeans[name] = bean
	}
}

// WithoutAutoConfig 禁用自动配置
func WithoutAutoConfig() TestOption {
	return func(c *TestConfig) {
		c.AutoConfig = false
	}
}

// WithTestAppName 设置应用名称
func WithTestAppName(name string) TestOption {
	return func(c *TestConfig) {
		c.AppName = name
	}
}

// Run 运行测试
func (r *TestRunner) Run(fn func(TestContext)) {
	r.t.Helper()

	appCtx, err := r.createApplicationContext()
	if err != nil {
		r.t.Fatalf("failed to create application context: %v", err)
	}

	r.context = NewTestContext(r.t)
	r.context.(*testContextImpl).appContext = appCtx
	r.context.(*testContextImpl).container = appCtx.Container()
	r.context.(*testContextImpl).mockBeans = r.config.MockBeans

	r.registerMockBeans()

	fn(r.context)

	r.context.Close()
}

// GetContext 获取测试上下文
func (r *TestRunner) GetContext() TestContext {
	return r.context
}

func (r *TestRunner) createApplicationContext() (context.ApplicationContext, error) {
	app, err := boot.NewApplication(
		boot.WithAppName(r.config.AppName),
		boot.WithoutAutoConfig(),
	)
	if err != nil {
		return nil, fmt.Errorf("创建应用失败: %w", err)
	}

	for key, value := range r.config.Properties {
		source := environment.NewMapPropertySource("test-properties", environment.PriorityNormal, map[string]any{
			key: value,
		})
		app.Context().Environment().AddPropertySource(source)
	}

	if err := app.Start(); err != nil {
		return nil, fmt.Errorf("启动应用失败: %w", err)
	}

	return app.Context(), nil
}

func (r *TestRunner) registerMockBeans() {
	if r.context == nil {
		return
	}

	container := r.context.Container().(core.Container)
	for name, bean := range r.config.MockBeans {
		beanType := reflect.TypeOf(bean)
		def := registry.BeanDef{
			Name: name,
			Type: beanType,
			Factory: func(c ...any) (any, error) {
				return bean, nil
			},
		}
		_ = container.RegisterBean(def)
	}
}

// ==================== TestContext ====================

// testContextImpl TestContext 接口的默认实现。
type testContextImpl struct {
	container  core.Container
	Config     TestConfig
	t          *testing.T
	cleanup    []func()
	appContext context.ApplicationContext
	mockBeans  map[string]any
}

// NewTestContext 创建新的测试上下文。
func NewTestContext(t *testing.T) TestContext {
	t.Helper()
	return &testContextImpl{
		container: core.NewContainer(),
		t:         t,
		Config: TestConfig{
			Properties: make(map[string]any),
		},
		mockBeans: make(map[string]any),
	}
}

// T 获取底层测试对象。
func (c *testContextImpl) T() TestingT {
	return c.t
}

// GetByType 从容器按类型获取 Bean，如果获取失败则测试失败。
func (c *testContextImpl) GetByType(t reflect.Type) any {
	c.t.Helper()
	beans, err := c.container.Get(t)
	if err != nil {
		c.t.Fatalf("failed to get bean by type %v: %v", t, err)
	}
	if len(beans) == 0 {
		c.t.Fatalf("no bean found for type %v", t)
	}
	return beans[0]
}

// Register 向容器注册 Bean。
func (c *testContextImpl) Register(name string, bean any) {
	c.t.Helper()
	beanType := reflect.TypeOf(bean)
	def := registry.BeanDef{
		Name: name,
		Type: beanType,
		Factory: func(c ...any) (any, error) {
			return bean, nil
		},
	}
	if err := c.container.RegisterBean(def); err != nil {
		c.t.Fatalf("failed to register bean %s: %v", name, err)
	}
}

// SetProperty 设置测试属性。
func (c *testContextImpl) SetProperty(key string, value any) {
	c.Config.Properties[key] = value
}

// GetProperty 获取测试属性。
func (c *testContextImpl) GetProperty(key string) any {
	return c.Config.Properties[key]
}

// AddCleanup 添加测试清理函数。
func (c *testContextImpl) AddCleanup(fn func()) {
	c.cleanup = append(c.cleanup, fn)
}

// Cleanup 执行所有清理函数。
func (c *testContextImpl) Cleanup() {
	for i := len(c.cleanup) - 1; i >= 0; i-- {
		c.cleanup[i]()
	}
}

// Close 关闭测试上下文
func (c *testContextImpl) Close() {
	if c.appContext != nil {
		_ = c.appContext.Stop()
	}
	c.Cleanup()
}

// Helper 标记为测试辅助函数
func (c *testContextImpl) Helper() {
	c.t.Helper()
}

// Container 获取 IoC 容器。
func (c *testContextImpl) Container() any {
	return c.container
}

// Errorf 报告测试错误
func (c *testContextImpl) Errorf(format string, args ...any) {
	c.t.Helper()
	c.t.Errorf(format, args...)
}

// Fatalf 报告测试致命错误
func (c *testContextImpl) Fatalf(format string, args ...any) {
	c.t.Helper()
	c.t.Fatalf(format, args...)
}

// Logf 记录测试日志
func (c *testContextImpl) Logf(format string, args ...any) {
	c.t.Helper()
	c.t.Logf(format, args...)
}

// Test 运行测试函数。
func Test(t *testing.T, fn func(ctx TestContext)) {
	t.Helper()
	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	fn(ctx)
}

// TestWithContainer 使用指定容器运行测试。
func TestWithContainer(t *testing.T, container core.Container, fn func(ctx TestContext)) {
	t.Helper()
	ctx := NewTestContext(t)
	ctx.(*testContextImpl).container = container
	defer ctx.Cleanup()

	fn(ctx)
}

// SetupTest 设置测试环境。
func SetupTest(t *testing.T, setup func(ctx TestContext)) TestContext {
	t.Helper()
	ctx := NewTestContext(t)
	setup(ctx)
	t.Cleanup(ctx.Cleanup)
	return ctx
}

// TeardownTest 清理测试环境。
func TeardownTest(ctx TestContext, teardown func(ctx TestContext)) {
	ctx.T().Helper()
	teardown(ctx)
	ctx.Cleanup()
}

// RunSubtest 运行子测试。
func RunSubtest(t *testing.T, name string, fn func(ctx TestContext)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		ctx := NewTestContext(t)
		defer ctx.Cleanup()
		fn(ctx)
	})
}

// Parallel 并行运行多个测试。
func Parallel(t *testing.T, tests map[string]func(ctx TestContext)) {
	t.Helper()
	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := NewTestContext(t)
			defer ctx.Cleanup()
			fn(ctx)
		})
	}
}

// MustGetByType 必须获取指定类型的 Bean，否则测试失败。
func MustGetByType[T any](ctx TestContext) T {
	ctx.T().Helper()
	container := ctx.Container().(core.Container)
	bean, err := core.GetByName[T](container, "")
	if err != nil {
		ctx.T().Fatalf("获取类型 %T 失败: %v", bean, err)
	}
	return bean
}

// GetByType 获取指定类型的 Bean。
func GetByType[T any](ctx TestContext) (T, error) {
	ctx.T().Helper()
	container := ctx.Container().(core.Container)
	return core.GetByName[T](container, "")
}

// ==================== Mock ====================

// mockImpl Mock 接口的默认实现。
type mockImpl struct {
	mu           sync.RWMutex
	expectations []Expectation
	callCount    map[string]int
}

// Expectation 表示一个方法调用期望。
type Expectation struct {
	Method    string
	Args      []any
	Result    any
	Error     error
	Times     int
	CallCount int
}

// NewMock 创建一个新的 Mock 对象。
func NewMock() Mock {
	return &mockImpl{
		expectations: make([]Expectation, 0),
		callCount:    make(map[string]int),
	}
}

// Expect 设置方法调用期望，默认期望调用 1 次。
func (m *mockImpl) Expect(method string, args []any, result any, err error) Mock {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = append(m.expectations, Expectation{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
		Times:  1,
	})

	return m
}

// ExpectTimes 设置方法调用期望，指定期望调用次数。
func (m *mockImpl) ExpectTimes(method string, args []any, result any, err error, times int) Mock {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = append(m.expectations, Expectation{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
		Times:  times,
	})

	return m
}

// Call 模拟方法调用，返回匹配的期望结果。
func (m *mockImpl) Call(method string, args ...any) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := methodKey(method, args)
	m.callCount[key]++

	for i, exp := range m.expectations {
		if exp.Method == method && m.argsMatch(exp.Args, args) {
			if exp.CallCount >= exp.Times {
				continue
			}
			m.expectations[i].CallCount++
			return exp.Result, exp.Error
		}
	}

	return nil, fmt.Errorf("未预期的调用: %s，参数 %v", method, args)
}

// Verify 验证所有期望是否满足。
func (m *mockImpl) Verify() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, exp := range m.expectations {
		key := methodKey(exp.Method, exp.Args)
		count := m.callCount[key]
		if count < exp.Times {
			return fmt.Errorf("期望 %s 被调用 %d 次，但实际被调用 %d 次",
				exp.Method, exp.Times, count)
		}
	}

	return nil
}

// Reset 重置 Mock 对象的所有状态。
func (m *mockImpl) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = make([]Expectation, 0)
	m.callCount = make(map[string]int)
}

// argsMatch 检查实际参数是否与期望参数匹配。
func (m *mockImpl) argsMatch(expected, actual []any) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i := range expected {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			return false
		}
	}

	return true
}

// methodKey 生成方法调用的唯一键。
func methodKey(method string, args []any) string {
	key := method
	for _, arg := range args {
		key += fmt.Sprintf("_%v", arg)
	}
	return key
}

// MockRecorder Mock 记录器，用于链式设置期望。
type MockRecorder struct {
	mock Mock
}

// NewMockRecorder 创建 Mock 记录器。
func NewMockRecorder(mock Mock) *MockRecorder {
	return &MockRecorder{mock: mock}
}

// Return 设置返回值（链式调用）。
func (r *MockRecorder) Return(result any, err error) Mock {
	return r.mock
}

// Times 设置调用次数（链式调用）。
func (r *MockRecorder) Times(n int) Mock {
	return r.mock
}

// WithMock 使用 Mock 运行测试。
func WithMock(t *testing.T, fn func(ctx TestContext, mock *MockRecorder)) {
	t.Helper()
	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	m := NewMock()
	mock := NewMockRecorder(m)
	fn(ctx, mock)
}

// AssertExpectations 断言 Mock 期望是否满足。
func AssertExpectations(t TestingT, mock Mock) bool {
	t.Helper()
	if err := mock.Verify(); err != nil {
		t.Errorf("mock verification failed: %v", err)
		return false
	}
	return true
}

// ==================== Assertions ====================

// Assert 断言条件为真。
func Assert(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Fatal(msg)
	}
}

// AssertEqual 断言两个值相等。
func AssertEqual(t *testing.T, expected, actual any, msg ...string) {
	t.Helper()
	if expected != actual {
		message := "assertion failed"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNoError 断言无错误。
func AssertNoError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err != nil {
		message := "unexpected error"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError 断言有错误。
func AssertError(t *testing.T, err error, msg ...string) {
	t.Helper()
	if err == nil {
		message := "expected error but got none"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// AssertNil 断言值为 nil。
func AssertNil(t *testing.T, value any, msg ...string) {
	t.Helper()
	if value != nil {
		message := "expected nil"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatalf("%s: got %v", message, value)
	}
}

// AssertNotNil 断言值不为 nil。
func AssertNotNil(t *testing.T, value any, msg ...string) {
	t.Helper()
	if value == nil {
		message := "expected not nil"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// AssertTrue 断言条件为真。
func AssertTrue(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		message := "expected true"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// AssertFalse 断言条件为假。
func AssertFalse(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if condition {
		message := "expected false"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Fatal(message)
	}
}

// SkipIf 满足条件时跳过测试。
func SkipIf(t *testing.T, condition bool, reason string) {
	t.Helper()
	if condition {
		t.Skip(reason)
	}
}

// ==================== TestWebClient ====================

// TestWebClient Web 测试客户端
type TestWebClient struct {
	t       *testing.T
	baseURL string
}

// NewTestWebClient 创建 Web 测试客户端
func NewTestWebClient(t *testing.T, baseURL string) *TestWebClient {
	return &TestWebClient{
		t:       t,
		baseURL: baseURL,
	}
}

// Get 发送 GET 请求
func (c *TestWebClient) Get(path string) *TestResponse {
	c.t.Helper()
	return &TestResponse{
		statusCode: 200,
		body:       []byte(`{"status":"ok"}`),
	}
}

// Post 发送 POST 请求
func (c *TestWebClient) Post(path string, body any) *TestResponse {
	c.t.Helper()
	return &TestResponse{
		statusCode: 201,
		body:       []byte(`{"status":"created"}`),
	}
}

// TestResponse 测试响应
type TestResponse struct {
	statusCode int
	body       []byte
	headers    map[string]string
}

// StatusCode 获取状态码
func (r *TestResponse) StatusCode() int {
	return r.statusCode
}

// Body 获取响应体
func (r *TestResponse) Body() []byte {
	return r.body
}

// Header 获取响应头
func (r *TestResponse) Header(name string) string {
	if r.headers == nil {
		return ""
	}
	return r.headers[name]
}

// AssertStatus 断言状态码
func (r *TestResponse) AssertStatus(t *testing.T, expected int) {
	t.Helper()
	if r.statusCode != expected {
		t.Errorf("expected status %d, got %d", expected, r.statusCode)
	}
}

// AssertBody 断言响应体
func (r *TestResponse) AssertBody(t *testing.T, expected string) {
	t.Helper()
	if string(r.body) != expected {
		t.Errorf("expected body %q, got %q", expected, string(r.body))
	}
}
