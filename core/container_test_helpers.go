package core

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/xudefa/enhance/core/lifecycle"
)

// TestService is a test bean for container tests.
type TestService struct {
	Name string
}

// TestRepository is a test bean for container tests.
type TestRepository struct {
	DB string
}

// testLifecycleBeanImpl implements lifecycle.Lifecycle for testing.
type testLifecycleBeanImpl struct {
	InitCalled    bool
	DestroyCalled bool
	InitError     error
	DestroyError  error
}

// Init 执行测试生命周期 Bean 的初始化并标记已调用。
func (b *testLifecycleBeanImpl) Init() error {
	b.InitCalled = true
	return b.InitError
}

// Destroy 执行测试生命周期 Bean 的销毁并标记已调用。
func (b *testLifecycleBeanImpl) Destroy() error {
	b.DestroyCalled = true
	return b.DestroyError
}

// testContainerBean is a simple bean for container tests.
type testContainerBean struct {
	ID string
}

// testContainerBeanWithConstructor is a bean with constructor for testing.
type testContainerBeanWithConstructor struct {
	Value string
}

func newTestContainerBeanWithConstructor() *testContainerBeanWithConstructor {
	return &testContainerBeanWithConstructor{Value: "constructed"}
}

// testContainerBeanWithErrorConstructor is a bean whose constructor returns an error.
type testContainerBeanWithErrorConstructor struct{}

func newTestContainerBeanWithErrorConstructor() (*testContainerBeanWithErrorConstructor, error) {
	return nil, fmt.Errorf("construct test bean: %w", errIntentionalError)
}

// errIntentionalError is a sentinel error for testing.
var errIntentionalError = &testError{msg: "intentional error"}

// testError is a simple error implementation for testing.
type testError struct{ msg string }

// Error 返回测试错误的文本信息。
func (e *testError) Error() string { return e.msg }

// mockBeanDefinition is a mock BeanDef for testing.
type mockBeanDefinition struct {
	BeanName       string
	BeanType       reflect.Type
	BeanInstance   any
	BeanInit       func(instance any) error
	BeanDestroy    func(instance any) error
	BeanOrder      int
	BeanLazyInit   bool
	BeanConditions []string
	BeanPrimary    bool
}

// Name 返回 mock Bean 定义的名称。
func (m *mockBeanDefinition) Name() string { return m.BeanName }

// Type 返回 mock Bean 定义的类型。
func (m *mockBeanDefinition) Type() reflect.Type { return m.BeanType }

// Instance 返回 mock Bean 定义的实例。
func (m *mockBeanDefinition) Instance() any { return m.BeanInstance }

// Init 返回 mock Bean 定义的初始化函数。
func (m *mockBeanDefinition) Init() func(instance any) error { return m.BeanInit }

// Destroy 返回 mock Bean 定义的销毁函数。
func (m *mockBeanDefinition) Destroy() func(instance any) error { return m.BeanDestroy }

// Order 返回 mock Bean 定义的初始化顺序。
func (m *mockBeanDefinition) Order() int { return m.BeanOrder }

// LazyInit 返回 mock Bean 定义是否延迟初始化。
func (m *mockBeanDefinition) LazyInit() bool { return m.BeanLazyInit }

// Conditions 返回 mock Bean 定义的启用条件列表。
func (m *mockBeanDefinition) Conditions() []string { return m.BeanConditions }

// Primary 返回 mock Bean 定义是否为主 Bean。
func (m *mockBeanDefinition) Primary() bool { return m.BeanPrimary }

// FactoryCircularA is a test bean for factory circular dependency testing.
type FactoryCircularA struct{}

// FactoryCircularB is a test bean for factory circular dependency testing.
type FactoryCircularB struct{}

// LifecycleBeanImpl implements lifecycle.Lifecycle for testing.
type LifecycleBeanImpl struct {
	InitCalled    bool
	DestroyCalled bool
	InitError     error
	DestroyError  error
}

// Init 执行测试 Bean 的初始化并标记已调用。
func (b *LifecycleBeanImpl) Init() error {
	b.InitCalled = true
	return b.InitError
}

// Destroy 执行测试 Bean 的销毁并标记已调用。
func (b *LifecycleBeanImpl) Destroy() error {
	b.DestroyCalled = true
	return b.DestroyError
}

// PhaseRecorder records lifecycle phase changes for testing.
type PhaseRecorder struct {
	mu        sync.Mutex
	phases    []lifecycle.Phase
	beanNames []string
}

// OnPhaseChange 记录一次 Bean 生命周期阶段变更。
func (r *PhaseRecorder) OnPhaseChange(beanName string, bean any, phase lifecycle.Phase) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.phases = append(r.phases, phase)
	r.beanNames = append(r.beanNames, beanName)
}

// GetPhases 返回已记录的阶段列表副本。
func (r *PhaseRecorder) GetPhases() []lifecycle.Phase {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]lifecycle.Phase{}, r.phases...)
}

// GetBeanNames 返回已记录的 Bean 名称列表副本。
func (r *PhaseRecorder) GetBeanNames() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string{}, r.beanNames...)
}

// A is a test struct for circular dependency validation.
type A struct {
	B *B `inject:""`
}

// B is a test struct for circular dependency validation.
type B struct {
	A *A `inject:""`
}
