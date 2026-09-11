package core

import (
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

func (b *testLifecycleBeanImpl) Init() error {
	b.InitCalled = true
	return b.InitError
}

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
	return nil, errIntentionalError
}

// errIntentionalError is a sentinel error for testing.
var errIntentionalError = &testError{msg: "intentional error"}

// testError is a simple error implementation for testing.
type testError struct{ msg string }

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

func (m *mockBeanDefinition) Name() string                      { return m.BeanName }
func (m *mockBeanDefinition) Type() reflect.Type                { return m.BeanType }
func (m *mockBeanDefinition) Instance() any                     { return m.BeanInstance }
func (m *mockBeanDefinition) Init() func(instance any) error    { return m.BeanInit }
func (m *mockBeanDefinition) Destroy() func(instance any) error { return m.BeanDestroy }
func (m *mockBeanDefinition) Order() int                        { return m.BeanOrder }
func (m *mockBeanDefinition) LazyInit() bool                    { return m.BeanLazyInit }
func (m *mockBeanDefinition) Conditions() []string              { return m.BeanConditions }
func (m *mockBeanDefinition) Primary() bool                     { return m.BeanPrimary }

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

func (b *LifecycleBeanImpl) Init() error {
	b.InitCalled = true
	return b.InitError
}

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

func (r *PhaseRecorder) OnPhaseChange(beanName string, bean any, phase lifecycle.Phase) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.phases = append(r.phases, phase)
	r.beanNames = append(r.beanNames, beanName)
}

func (r *PhaseRecorder) GetPhases() []lifecycle.Phase {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]lifecycle.Phase{}, r.phases...)
}

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
