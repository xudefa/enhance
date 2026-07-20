package core

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/xudefa/enhance/core/lifecycle"
	"github.com/xudefa/enhance/core/registry"
)

// ==================== 测试用 Bean ====================

type TestService struct {
	Name string
}

type TestRepository struct {
	Service *TestService
}

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

// ==================== BeanID 生成测试 ====================

func TestGenerateBeanID(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf((*TestService)(nil)).Elem()

	// 测试无自定义名称
	beanID := container.Generate(typ)
	expected := "github.com/xudefa/enhance/core.TestService"
	if beanID != expected {
		t.Errorf("Expected %s, got %s", expected, beanID)
	}

	// 测试有自定义名称
	beanID = container.Generate(typ, "myBean")
	expected = "github.com/xudefa/enhance/core.TestService#myBean"
	if beanID != expected {
		t.Errorf("Expected %s, got %s", expected, beanID)
	}

	// 测试已有标准格式
	beanID = container.Generate(typ, "github.com/xudefa/enhance/core.TestService#custom")
	expected = "github.com/xudefa/enhance/core.TestService#custom"
	if beanID != expected {
		t.Errorf("Expected %s, got %s", expected, beanID)
	}
}

func TestParseBeanID(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 测试标准格式
	pkg, typ, custom := container.Parse("github.com/example.MyType#myBean")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "MyType" {
		t.Errorf("Expected type 'MyType', got '%s'", typ)
	}
	if custom != "myBean" {
		t.Errorf("Expected custom 'myBean', got '%s'", custom)
	}

	// 测试无自定义名称
	pkg, typ, custom = container.Parse("github.com/example.MyType")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "MyType" {
		t.Errorf("Expected type 'MyType', got '%s'", typ)
	}
	if custom != "" {
		t.Errorf("Expected empty custom, got '%s'", custom)
	}
}

// ==================== 注册边界条件测试 ====================

func TestRegisterNilType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	def := registry.BeanDef{
		Type: nil,
	}
	err := container.RegisterBean(def)
	if err == nil {
		t.Error("Expected error for nil type")
	}
}

func TestRegisterDuplicateBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("duplicate"))
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	// 重复注册应该不报错（幂等）
	err = Register[*TestService](container, WithName[*TestService]("duplicate"))
	if err != nil {
		t.Errorf("Expected no error for duplicate registration, got: %v", err)
	}
}

func TestRegisterAfterInitialize(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 初始化后再注册应该报错
	err = Register[*TestService](container, WithName[*TestService]("test2"))
	if err == nil {
		t.Error("Expected error when registering after initialization")
	}
}

func TestRegisterWithEmptyName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 不提供名称，应该使用默认生成
	err := Register[*TestService](container)
	if err != nil {
		t.Fatalf("Register with empty name failed: %v", err)
	}

	// 应该能通过类型获取（使用指针类型）
	services, err := container.Get(reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("Get by type failed: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
}

// ==================== 作用域测试 ====================

func TestSingletonScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := int32(0)
	err := Register[*TestService](container, WithName[*TestService]("singleton"), WithFactory[*TestService](func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return &TestService{Name: "singleton"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 并发获取多次
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc, err := GetByName[*TestService](container, "singleton")
			if err != nil {
				t.Errorf("Get failed: %v", err)
			}
			if svc.Name != "singleton" {
				t.Errorf("Expected name 'singleton', got '%s'", svc.Name)
			}
		}()
	}
	wg.Wait()

	// 工厂只应该被调用一次
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected factory to be called once, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestPrototypeScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := int32(0)
	err := Register[*TestService](container, WithName[*TestService]("prototype"), WithFactory[*TestService](func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return &TestService{Name: fmt.Sprintf("instance-%d", atomic.LoadInt32(&callCount))}, nil
	}), WithScope[*TestService]("prototype"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 获取多次
	svc1, _ := GetByName[*TestService](container, "prototype")
	svc2, _ := GetByName[*TestService](container, "prototype")
	svc3, _ := GetByName[*TestService](container, "prototype")

	// 每次应该不同
	if svc1 == svc2 || svc2 == svc3 || svc1 == svc3 {
		t.Error("Expected different instances for prototype scope")
	}

	// 工厂应该被调用 3 次
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("Expected factory to be called 3 times, got %d", atomic.LoadInt32(&callCount))
	}
}

// ==================== 生命周期测试 ====================

func TestLifecycleInitCallback(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestService](container, WithName[*TestService]("initTest"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "init"}, nil
	}), WithInit[*TestService](func(bean any) error {
		initCalled = true
		svc := bean.(*TestService)
		svc.Name = "initialized"
		return nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !initCalled {
		t.Error("Expected Init callback to be called")
	}

	// 验证初始化回调修改了实例
	svc, _ := GetByName[*TestService](container, "initTest")
	if svc.Name != "initialized" {
		t.Errorf("Expected name 'initialized', got '%s'", svc.Name)
	}
}

func TestLifecycleDestroyCallback(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	destroyCalled := false
	err := Register[*TestService](container, WithName[*TestService]("destroyTest"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "destroy"}, nil
	}), WithDestroy[*TestService](func(bean any) error {
		destroyCalled = true
		return nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	if !destroyCalled {
		t.Error("Expected Destroy callback to be called")
	}
}

func TestLifecycleBeanInterface(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	bean := &LifecycleBeanImpl{}
	err := container.RegisterBean(registry.BeanDef{
		Type: reflect.TypeOf((*LifecycleBeanImpl)(nil)).Elem(),
		Name: "lifecycleBean",
		Factory: func(c ...any) (any, error) {
			t.Logf("Factory called, bean type: %T", bean)
			// 检查是否实现了 lifecycle.LifecycleBean 接口
			if _, ok := any(bean).(lifecycle.LifecycleBean); ok {
				t.Log("Bean implements lifecycle.LifecycleBean")
			} else {
				t.Log("Bean does NOT implement lifecycle.LifecycleBean")
			}
			return bean, nil
		},
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	t.Logf("After Initialize: InitCalled=%v", bean.InitCalled)

	if !bean.InitCalled {
		t.Error("Expected LifecycleBean.Init() to be called")
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	t.Logf("After Destroy: DestroyCalled=%v", bean.DestroyCalled)

	if !bean.DestroyCalled {
		t.Error("Expected LifecycleBean.Destroy() to be called")
	}
}

func TestLifecyclePhaseNotification(t *testing.T) {
	t.Parallel()
	// 直接测试生命周期管理器的通知功能
	mgr := lifecycle.NewLifecycleManager()
	recorder := &PhaseRecorder{}

	// 注册监听器
	mgr.RegisterListener(recorder)

	bean := &TestService{Name: "test"}
	mgr.NotifyPhaseChange("testBean", bean, lifecycle.PhaseInitialized)

	phases := recorder.GetPhases()
	if len(phases) != 1 {
		t.Errorf("Expected 1 phase notification, got %d", len(phases))
	}

	if phases[0] != lifecycle.PhaseInitialized {
		t.Errorf("Expected PhaseInitialized, got %v", phases[0])
	}
}

func TestLifecycleInitError(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	expectedErr := fmt.Errorf("init error")
	err := Register[*TestService](container, WithName[*TestService]("initError"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "init"}, nil
	}), WithInit[*TestService](func(bean any) error {
		return expectedErr
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err == nil {
		t.Fatal("Expected Initialize to fail")
	}

	if err != expectedErr {
		t.Errorf("Expected init error, got: %v", err)
	}
}

func TestLifecycleFactoryError(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	expectedErr := fmt.Errorf("factory error")
	err := Register[*TestService](container, WithName[*TestService]("factoryError"), WithFactory[*TestService](func(c ...any) (any, error) {
		return nil, expectedErr
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err == nil {
		t.Fatal("Expected Initialize to fail")
	}

	if err != expectedErr {
		t.Errorf("Expected factory error, got: %v", err)
	}
}

func TestLifecycleDestroyOrder(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	var destroyOrder []string
	var mu sync.Mutex

	err := Register[*TestService](container, WithName[*TestService]("first"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "first"}, nil
	}), WithDestroy[*TestService](func(bean any) error {
		mu.Lock()
		defer mu.Unlock()
		destroyOrder = append(destroyOrder, "first")
		return nil
	}))
	if err != nil {
		t.Fatalf("Register first failed: %v", err)
	}

	err = Register[*TestService](container, WithName[*TestService]("second"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "second"}, nil
	}), WithDestroy[*TestService](func(bean any) error {
		mu.Lock()
		defer mu.Unlock()
		destroyOrder = append(destroyOrder, "second")
		return nil
	}))
	if err != nil {
		t.Fatalf("Register second failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	t.Logf("Before Destroy: destroyOrder=%v", destroyOrder)

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	t.Logf("After Destroy: destroyOrder=%v", destroyOrder)

	// 验证逆序销毁
	if len(destroyOrder) != 2 {
		t.Fatalf("Expected 2 destroy calls, got %d", len(destroyOrder))
	}

	if destroyOrder[0] != "second" || destroyOrder[1] != "first" {
		t.Errorf("Expected reverse order destroy, got %v", destroyOrder)
	}
}

func TestLifecycleDestroyError(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	expectedErr := fmt.Errorf("destroy error")
	err := Register[*TestService](container, WithName[*TestService]("destroyError"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "destroy"}, nil
	}), WithDestroy[*TestService](func(bean any) error {
		return expectedErr
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = container.Destroy()
	if err == nil {
		t.Fatal("Expected Destroy to fail")
	}

	if err != expectedErr {
		t.Errorf("Expected destroy error, got: %v", err)
	}
}

func TestLifecycleMultipleDestroyCallbacks(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	var callCount int32
	err := Register[*TestService](container, WithName[*TestService]("multiDestroy"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "multi"}, nil
	}), WithDestroy[*TestService](func(bean any) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 多次销毁应该只调用一次
	err = container.Destroy()
	if err != nil {
		t.Fatalf("First Destroy failed: %v", err)
	}

	firstCount := atomic.LoadInt32(&callCount)

	// 再次销毁不应该再调用
	err = container.Destroy()
	if err == nil {
		t.Error("Expected error when destroying already destroyed container")
	}

	if atomic.LoadInt32(&callCount) != firstCount {
		t.Errorf("Destroy callback should not be called again, expected %d, got %d", firstCount, atomic.LoadInt32(&callCount))
	}
}

func TestLifecycleLazyInitialization(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestService](container, WithName[*TestService]("lazy"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "lazy"}, nil
	}), WithInit[*TestService](func(bean any) error {
		initCalled = true
		return nil
	}), WithLazy[*TestService](true))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 懒加载的 Bean 不应该在 Initialize 时创建
	if initCalled {
		t.Error("Expected lazy bean not to be initialized")
	}

	// 获取时才创建
	_, err = GetByName[*TestService](container, "lazy")
	if err != nil {
		t.Fatalf("Get lazy bean failed: %v", err)
	}

	if !initCalled {
		t.Error("Expected lazy bean to be initialized when accessed")
	}
}

func TestLifecycleGetBeforeInitialize(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestService](container, WithName[*TestService]("beforeInit"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "before"}, nil
	}), WithInit[*TestService](func(bean any) error {
		initCalled = true
		return nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 在 Initialize 之前获取 Bean
	svc, err := GetByName[*TestService](container, "beforeInit")
	if err != nil {
		t.Fatalf("Get before initialize failed: %v", err)
	}

	if svc.Name != "before" {
		t.Errorf("Expected name 'before', got '%s'", svc.Name)
	}

	if !initCalled {
		t.Error("Expected Init to be called when accessing bean before Initialize")
	}
}

// ==================== 容器状态测试 ====================

func TestContainerDestroyState(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后获取 Bean 应该报错
	_, err = GetByName[*TestService](container, "test")
	if err == nil {
		t.Error("Expected error when getting bean from destroyed container")
	}

	// 销毁后再次销毁应该报错
	err = container.Destroy()
	if err == nil {
		t.Error("Expected error when destroying already destroyed container")
	}

	// 销毁后初始化应该报错
	err = container.Initialize()
	if err == nil {
		t.Error("Expected error when initializing destroyed container")
	}
}

func TestContainerAlreadyInitialized(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("First Initialize failed: %v", err)
	}

	// 重复初始化应该报错
	err = container.Initialize()
	if err == nil {
		t.Error("Expected error when initializing already initialized container")
	}
}

// ==================== 类型安全测试 ====================

func TestHasWithTypeCheck(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("testService"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 应该存在
	if !Has[*TestService](container, "testService") {
		t.Error("Expected Has to return true for registered bean")
	}

	// 类型不匹配应该返回 false
	if Has[*TestRepository](container, "testService") {
		t.Error("Expected Has to return false for mismatched type")
	}

	// 名称不存在应该返回 false
	if Has[*TestService](container, "nonexistent") {
		t.Error("Expected Has to return false for nonexistent bean")
	}
}

func TestGetByType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册多个同类型的 Bean
	err := Register[*TestService](container, WithName[*TestService]("svc1"))
	if err != nil {
		t.Fatalf("Register svc1 failed: %v", err)
	}

	err = Register[*TestService](container, WithName[*TestService]("svc2"))
	if err != nil {
		t.Fatalf("Register svc2 failed: %v", err)
	}

	// 使用指针类型查询（与注册时一致）
	services, err := container.Get(reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("Get by type failed: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}
}

func TestGetByTypeNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_, err := container.Get(reflect.TypeOf((*TestService)(nil)).Elem())
	if err == nil {
		t.Error("Expected error when getting nonexistent type")
	}
}

func TestGetByNameNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	_, err := GetByName[*TestService](container, "nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent bean")
	}
}

// ==================== 父容器测试 ====================

func TestParentContainer(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	err := Register[*TestService](parent, WithName[*TestService]("parentBean"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "parent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register to parent failed: %v", err)
	}

	// 通过类型断言访问 SetParent
	childExt := child.(ContainerExt)
	childExt.SetParent(parent)

	// 子容器应该能获取父容器的 Bean
	svc, err := GetByName[*TestService](child, "parentBean")
	if err != nil {
		t.Fatalf("Get from parent failed: %v", err)
	}

	if svc.Name != "parent" {
		t.Errorf("Expected name 'parent', got '%s'", svc.Name)
	}
}

func TestParentContainerGetByType(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	err := Register[*TestService](parent, WithName[*TestService]("parentBean"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "parent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register to parent failed: %v", err)
	}

	// 通过类型断言访问 SetParent
	childExt := child.(ContainerExt)
	childExt.SetParent(parent)

	// 子容器应该能通过类型获取父容器的 Bean（使用指针类型）
	services, err := child.Get(reflect.TypeOf((*TestService)(nil)))
	if err != nil {
		t.Fatalf("Get by type from parent failed: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service from parent, got %d", len(services))
	}
}

// ==================== 并发安全测试 ====================

func TestConcurrentRegister(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("concurrent%d", i)
			err := Register[*TestService](container, WithName[*TestService](name), WithFactory[*TestService](func(c ...any) (any, error) {
				return &TestService{Name: name}, nil
			}))
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent register failed: %v", err)
	}

	// 验证所有 Bean 都注册成功
	ext := container.(ContainerExt)
	count := ext.BeanCount()
	if count != 10 {
		t.Errorf("Expected 10 beans, got %d", count)
	}
}

func TestConcurrentGetAndInitialize(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("concurrent"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "concurrent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 20)

	// 并发初始化和获取
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := container.Initialize(); err != nil && err != ErrContainerAlreadyInitialized {
				errChan <- err
			}
		}()
		go func() {
			defer wg.Done()
			_, err := GetByName[*TestService](container, "concurrent")
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent get/initialize failed: %v", err)
	}
}

// ==================== BeanCount 测试 ====================

func TestBeanCount(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	if ext.BeanCount() != 0 {
		t.Errorf("Expected 0 beans, got %d", ext.BeanCount())
	}

	Register[*TestService](container, WithName[*TestService]("svc1"))
	Register[*TestService](container, WithName[*TestService]("svc2"))
	Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	if ext.BeanCount() != 3 {
		t.Errorf("Expected 3 beans, got %d", ext.BeanCount())
	}

	// 使用指针类型查询（与注册时一致）
	typ := reflect.TypeOf((*TestService)(nil))
	if ext.BeanCountType(typ) != 2 {
		t.Errorf("Expected 2 TestService beans, got %d", ext.BeanCountType(typ))
	}
}

// ==================== GetAll 测试 ====================

func TestGetAll(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc1"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc1"}, nil
	}))
	Register[*TestService](container, WithName[*TestService]("svc2"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc2"}, nil
	}))

	// 初始化容器以实例化所有 Bean
	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	instances := container.GetAll()
	if len(instances) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(instances))
	}
}

// ==================== 边界条件测试 ====================

func TestEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	if ext.BeanCount() != 0 {
		t.Errorf("Expected 0 beans in empty container, got %d", ext.BeanCount())
	}

	types := ext.Types()
	if len(types) != 0 {
		t.Errorf("Expected 0 types in empty container, got %d", len(types))
	}

	instances := container.GetAll()
	if len(instances) != 0 {
		t.Errorf("Expected 0 instances in empty container, got %d", len(instances))
	}
}

func TestRegisterWithNilFactory(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 不提供 Factory，应该使用默认工厂
	err := Register[*TestService](container, WithName[*TestService]("nilFactory"))
	if err != nil {
		t.Fatalf("Register with nil factory failed: %v", err)
	}

	svc, err := GetByName[*TestService](container, "nilFactory")
	if err != nil {
		t.Fatalf("Get with nil factory failed: %v", err)
	}

	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestLifecycleListenerConcurrent(t *testing.T) {
	t.Parallel()
	mgr := lifecycle.NewLifecycleManager()
	recorder := &PhaseRecorder{}

	// 并发注册监听器
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.RegisterListener(recorder)
		}()
	}
	wg.Wait()

	err := Register[*TestService](NewContainer(), WithName[*TestService]("concurrentListener"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "test"}, nil
	}), WithInit[*TestService](func(bean any) error {
		return nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
}

// ==================== 补充单测：提高覆盖率 ====================

func TestHasTypeBasic(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 未注册时应该返回 false
	if container.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return false for unregistered type")
	}

	// 注册后应该返回 true
	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !container.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return true for registered type")
	}
}

func TestGetParent(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	// 默认没有父容器
	if child.(ContainerExt).GetParent() != nil {
		t.Error("Expected parent to be nil by default")
	}

	// 设置父容器
	child.(ContainerExt).SetParent(parent)

	// 应该能获取到父容器
	if child.(ContainerExt).GetParent() != parent {
		t.Error("Expected to get parent container")
	}
}

func TestMustGet(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "must-get"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 成功获取
	svc := MustGet[*TestService](container, "test")
	if svc.Name != "must-get" {
		t.Errorf("Expected name 'must-get', got '%s'", svc.Name)
	}

	// 失败时应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nonexistent bean")
		}
	}()
	MustGet[*TestService](container, "nonexistent")
}

func TestGetByNameByType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("svc1"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 不指定名称，按类型获取
	svc, err := GetByName[*TestService](container, "")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestGetAllWithErrors(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册一个会失败的 Bean
	err := Register[*TestService](container, WithName[*TestService]("fail"), WithFactory[*TestService](func(c ...any) (any, error) {
		return nil, fmt.Errorf("factory error")
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// GetAll 应该跳过错误的 Bean
	all := container.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected 0 beans (factory error), got %d", len(all))
	}
}

func TestResolveBeanNotFound(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 尝试获取不存在的 Bean
	_, err := container.GetByTypeAndName("nonexistent", reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error for nonexistent bean")
	}
}

func TestCreateAndInitializeCached(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	callCount := int32(0)
	err := Register[*TestService](container, WithName[*TestService]("cached"), WithFactory[*TestService](func(c ...any) (any, error) {
		atomic.AddInt32(&callCount, 1)
		return &TestService{Name: "cached"}, nil
	}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 第一次获取
	svc1, err := GetByName[*TestService](container, "cached")
	if err != nil {
		t.Fatalf("First GetByName failed: %v", err)
	}

	// 第二次获取（应该从缓存）
	svc2, err := GetByName[*TestService](container, "cached")
	if err != nil {
		t.Fatalf("Second GetByName failed: %v", err)
	}

	// 应该是同一个实例
	if svc1 != svc2 {
		t.Error("Expected same cached instance")
	}

	// 工厂只应该被调用一次
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected factory to be called once, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestGetDestroyedContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Get 应该报错
	_, err = container.Get(reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error when getting from destroyed container")
	}

	// 销毁后 GetByTypeAndName 应该报错
	_, err = container.GetByTypeAndName("test", reflect.TypeOf((*TestService)(nil)))
	if err == nil {
		t.Error("Expected error when getting by name from destroyed container")
	}

	// 销毁后 GetAll 应该返回空
	all := container.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected empty list from destroyed container, got %d", len(all))
	}
}

func TestInitializeDestroyedContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后初始化应该报错
	err = container.Initialize()
	if err == nil {
		t.Error("Expected error when initializing destroyed container")
	}
}

func TestRegisterBeanWithCustomType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 使用 WithType 覆盖类型
	err := Register[*TestService](container,
		WithName[*TestService]("customType"),
		WithType[*TestService](reflect.TypeOf((*TestRepository)(nil))),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "custom"}, nil
		}))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 应该能通过覆盖后的类型获取
	services, err := container.Get(reflect.TypeOf((*TestRepository)(nil)))
	if err != nil {
		t.Fatalf("Get by custom type failed: %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}
}

func TestRegisterPrimaryBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册非首选 Bean
	err := Register[*TestService](container, WithName[*TestService]("secondary"))
	if err != nil {
		t.Fatalf("Register secondary failed: %v", err)
	}

	// 注册首选 Bean
	err = Register[*TestService](container, WithName[*TestService]("primary"), WithPrimary[*TestService](true))
	if err != nil {
		t.Fatalf("Register primary failed: %v", err)
	}

	// 检查 Bean 数量
	ext := container.(ContainerExt)
	if ext.BeanCountType(reflect.TypeOf((*TestService)(nil))) != 2 {
		t.Errorf("Expected 2 beans, got %d", ext.BeanCountType(reflect.TypeOf((*TestService)(nil))))
	}
}

func TestGetWithParentContainer(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	child := NewContainer()

	err := Register[*TestService](parent, WithName[*TestService]("parentBean"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "from-parent"}, nil
	}))
	if err != nil {
		t.Fatalf("Register to parent failed: %v", err)
	}

	child.(ContainerExt).SetParent(parent)

	// 子容器应该能获取父容器的 Bean
	svc, err := GetByName[*TestService](child, "parentBean")
	if err != nil {
		t.Fatalf("Get from parent failed: %v", err)
	}

	if svc.Name != "from-parent" {
		t.Errorf("Expected name 'from-parent', got '%s'", svc.Name)
	}
}

func TestTypes(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc"))
	Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	types := container.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(types))
	}
}

func TestTypesEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	types := container.Types()
	if len(types) != 0 {
		t.Errorf("Expected 0 types in empty container, got %d", len(types))
	}
}

func TestTypesWithMultipleSameTypeBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc1"))
	Register[*TestService](container, WithName[*TestService]("svc2"))
	Register[*TestService](container, WithName[*TestService]("svc3"))
	Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	types := container.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(types))
	}
}

func TestTypesAfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc"))
	Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Bean 定义仍然存在
	types := container.Types()
	if len(types) != 2 {
		t.Errorf("Expected 2 types after destroy, got %d", len(types))
	}
}

func TestTypesWithInterfaceBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[fmt.Stringer](container, WithName[fmt.Stringer]("stringer"), WithFactory[fmt.Stringer](func(c ...any) (any, error) {
		return &TestService{Name: "test"}, nil
	}))

	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}

	// 验证类型是接口类型
	expectedType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	if types[0] != expectedType {
		t.Errorf("Expected type %v, got %v", expectedType, types[0])
	}
}

func TestTypesWithFactoryBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("factory"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "factory-created"}, nil
	}))

	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}
}

func TestTypesWithLazyBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("lazy"), WithLazy[*TestService](true))

	// 懒加载 Bean 注册后应该出现在 Types 中
	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}
}

func TestTypesWithPrototypeBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("prototype"), WithScope[*TestService]("prototype"))

	types := container.Types()
	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}
}

func TestBeanCountType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	Register[*TestService](container, WithName[*TestService]("svc1"))
	Register[*TestService](container, WithName[*TestService]("svc2"))
	Register[*TestRepository](container, WithName[*TestRepository]("repo"))

	// 测试类型计数
	svcCount := ext.BeanCountType(reflect.TypeOf((*TestService)(nil)))
	if svcCount != 2 {
		t.Errorf("Expected 2 TestService beans, got %d", svcCount)
	}

	repoCount := ext.BeanCountType(reflect.TypeOf((*TestRepository)(nil)))
	if repoCount != 1 {
		t.Errorf("Expected 1 TestRepository bean, got %d", repoCount)
	}

	// 测试不存在的类型
	nonExistentCount := ext.BeanCountType(reflect.TypeOf((*string)(nil)))
	if nonExistentCount != 0 {
		t.Errorf("Expected 0 for non-existent type, got %d", nonExistentCount)
	}
}

func TestHasWithNonExistentBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 不存在的 Bean 应该返回 false
	if container.Has("nonexistent", reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected Has to return false for non-existent bean")
	}
}

func TestHasWithWrongType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("test"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 名称存在但类型不匹配应该返回 false
	if container.Has("test", reflect.TypeOf((*TestRepository)(nil))) {
		t.Error("Expected Has to return false for mismatched type")
	}
}

func TestGetByNameMultipleBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("svc1"))
	if err != nil {
		t.Fatalf("Register svc1 failed: %v", err)
	}

	err = Register[*TestService](container, WithName[*TestService]("svc2"))
	if err != nil {
		t.Fatalf("Register svc2 failed: %v", err)
	}

	// 不指定名称时应该返回第一个
	svc, err := GetByName[*TestService](container, "")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if svc == nil {
		t.Error("Expected non-nil service")
	}
}

func TestInitializeWithLazyBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestService](container,
		WithName[*TestService]("lazy"),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "lazy"}, nil
		}),
		WithInit[*TestService](func(bean any) error {
			initCalled = true
			return nil
		}),
		WithLazy[*TestService](true))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Initialize 不应该触发懒加载 Bean 的初始化
	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if initCalled {
		t.Error("Expected lazy bean not to be initialized during Initialize")
	}
}

func TestInitializeWithPrototypeBean(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	initCalled := false
	err := Register[*TestService](container,
		WithName[*TestService]("prototype"),
		WithFactory[*TestService](func(c ...any) (any, error) {
			return &TestService{Name: "prototype"}, nil
		}),
		WithInit[*TestService](func(bean any) error {
			initCalled = true
			return nil
		}),
		WithScope[*TestService]("prototype"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Initialize 不应该初始化 prototype Bean
	err = container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if initCalled {
		t.Error("Expected prototype bean not to be initialized during Initialize")
	}
}

func TestGenerateBeanIDWithEmptyCustomName(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf((*TestService)(nil))

	// 测试空字符串作为自定义名称
	beanID := container.Generate(typ, "")
	expected := "github.com/xudefa/enhance/core.TestService"
	if beanID != expected {
		t.Errorf("Expected %s, got %s", expected, beanID)
	}
}

func TestParseBeanIDWithNoDot(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 测试没有点号的 Bean ID
	pkg, typ, custom := container.Parse("SimpleType")
	if pkg != "" {
		t.Errorf("Expected empty pkg, got '%s'", pkg)
	}
	if typ != "SimpleType" {
		t.Errorf("Expected type 'SimpleType', got '%s'", typ)
	}
	if custom != "" {
		t.Errorf("Expected empty custom, got '%s'", custom)
	}
}

func TestParseBeanIDWithCustomName(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 测试带自定义名称的 Bean ID
	pkg, typ, custom := container.Parse("github.com/example.MyType#custom")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "MyType" {
		t.Errorf("Expected type 'MyType', got '%s'", typ)
	}
	if custom != "custom" {
		t.Errorf("Expected custom 'custom', got '%s'", custom)
	}
}

func TestValidateSuccess(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container)
	Register[*TestRepository](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestValidateMissingDependency(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type TestBean struct {
		Service *TestService `inject:""`
	}

	Register[*TestBean](container)
	// 注意：没有注册 TestService

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("Expected validation error for missing dependency")
	}
}

type A struct {
	B *B `inject:""`
}

type B struct {
	A *A `inject:""`
}

func TestValidateCircularDependency(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*A](container)
	Register[*B](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("Expected validation error for circular dependency")
	}
}

func TestValidateAlreadyInitialized(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	Register[*TestService](container)

	// 先初始化
	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ext := container.(ContainerExt)
	err = ext.Validate()
	if err == nil {
		t.Error("Expected error when validating initialized container")
	}
}

func TestValidateWithParent(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	Register[*TestService](parent)

	child := NewContainer()
	ext := child.(ContainerExt)
	ext.SetParent(parent)

	type TestBean struct {
		Service *TestService `inject:""`
	}

	Register[*TestBean](child)

	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed with parent: %v", err)
	}
}

func TestValidateWithEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed for empty container: %v", err)
	}
}

func TestValidateWithNonStructType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	Register[*TestService](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed for non-struct type: %v", err)
	}
}

func TestValidateWithMultipleDependencies(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type Config struct {
		Timeout int `value:"timeout"`
	}

	type ServiceA struct {
		Config *Config `inject:""`
	}

	type ServiceB struct {
		Config *Config `inject:""`
	}

	Register[*Config](container)
	Register[*ServiceA](container)
	Register[*ServiceB](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestValidateWithPartialDependencies(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	type DependencyA struct {
		Name string
	}

	type DependencyB struct {
		Name string
	}

	type Service struct {
		DepA *DependencyA `inject:""`
		DepB *DependencyB `inject:""`
	}

	Register[*DependencyA](container)
	Register[*Service](container)

	ext := container.(ContainerExt)
	err := ext.Validate()
	if err == nil {
		t.Error("Expected validation error for partial dependencies")
	}
}

// ==================== BeanGet 接口测试 ====================

func TestListBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc1"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc1"}, nil
	}))
	Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	beanGet := container.(BeanGet)
	beanDefs := beanGet.ListBeans()
	if len(beanDefs) != 2 {
		t.Errorf("Expected 2 bean defs, got %d", len(beanDefs))
	}

	// 验证 Bean 信息
	foundSvc := false
	foundRepo := false
	for id, def := range beanDefs {
		if def.Type == nil {
			continue
		}
		typeStr := def.Type.String()
		if typeStr == "*core.TestService" {
			foundSvc = true
		}
		if typeStr == "*core.TestRepository" {
			foundRepo = true
		}
		_ = id
	}
	if !foundSvc || !foundRepo {
		t.Error("Expected to find both TestService and TestRepository beans")
	}
}

func TestListBeansEmptyContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	beanGet := container.(BeanGet)
	beanDefs := beanGet.ListBeans()
	if len(beanDefs) != 0 {
		t.Errorf("Expected 0 beans in empty container, got %d", len(beanDefs))
	}
}

func TestGetBeanDef(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	err := Register[*TestService](container, WithName[*TestService]("myService"))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 获取存在的 BeanDef
	beanGet := container.(BeanGet)
	beanDefs := beanGet.ListBeans()
	expectedID := "github.com/xudefa/enhance/core.TestService#myService"
	def, ok := beanDefs[expectedID]
	if !ok {
		t.Error("Expected to find BeanDef")
	}
	if def == nil {
		t.Error("Expected non-nil BeanDef")
	}

	// 获取不存在的 BeanDef
	_, ok = beanDefs["nonexistent"]
	if ok {
		t.Error("Expected not to find BeanDef for nonexistent bean")
	}
}

func TestListBeansWithInitializedContainer(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc1"), WithFactory[*TestService](func(c ...any) (any, error) {
		return &TestService{Name: "svc1"}, nil
	}))

	beanGet := container.(BeanGet)

	// 初始化前
	beansBefore := beanGet.ListBeans()
	if len(beansBefore) != 1 {
		t.Errorf("Expected 1 bean before initialization, got %d", len(beansBefore))
	}

	// 初始化后
	err := container.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	beansAfter := beanGet.ListBeans()
	if len(beansAfter) != 1 {
		t.Errorf("Expected 1 bean after initialization, got %d", len(beansAfter))
	}
}

func TestHasTypeWithMultipleTypes(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc1"))
	Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	if !container.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return true for TestService")
	}

	if !container.HasType(reflect.TypeOf((*TestRepository)(nil))) {
		t.Error("Expected HasType to return true for TestRepository")
	}
}

func TestBeanCountTypeWithPrototypeScope(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	Register[*TestService](container, WithName[*TestService]("proto1"), WithScope[*TestService]("prototype"))
	Register[*TestService](container, WithName[*TestService]("proto2"), WithScope[*TestService]("prototype"))

	count := ext.BeanCountType(reflect.TypeOf((*TestService)(nil)))
	if count != 2 {
		t.Errorf("Expected 2 prototype beans, got %d", count)
	}
}

func TestListBeansWithNilType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册 nil 类型的 BeanDef 应该报错
	def := registry.BeanDef{
		Type: nil,
	}
	err := container.RegisterBean(def)
	if err == nil {
		t.Error("Expected error when registering bean with nil type")
	}
}

func TestHasTypeWithInterfaceType(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	// 注册接口类型的 Bean
	Register[fmt.Stringer](container, WithName[fmt.Stringer]("stringer"), WithFactory[fmt.Stringer](func(c ...any) (any, error) {
		return &TestService{Name: "test"}, nil
	}))

	if !container.HasType(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
		t.Error("Expected HasType to return true for interface type")
	}
}

func TestBeanCountTypeAfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	ext := container.(ContainerExt)

	Register[*TestService](container, WithName[*TestService]("svc1"))
	Register[*TestService](container, WithName[*TestService]("svc2"))

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Bean 定义仍然存在，但实例被清除
	count := ext.BeanCountType(reflect.TypeOf((*TestService)(nil)))
	if count != 2 {
		t.Errorf("Expected 2 bean defs after destroy, got %d", count)
	}
}

func TestListBeansAfterDestroy(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc1"))

	err := container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// 销毁后 Bean 定义仍然存在
	beanGet := container.(BeanGet)
	beanDefs := beanGet.ListBeans()
	if len(beanDefs) != 1 {
		t.Errorf("Expected 1 bean def after destroy, got %d", len(beanDefs))
	}
}

func TestGetBeanDefWithMultipleBeans(t *testing.T) {
	t.Parallel()
	container := NewContainer()

	Register[*TestService](container, WithName[*TestService]("svc1"))
	Register[*TestService](container, WithName[*TestService]("svc2"))
	Register[*TestRepository](container, WithName[*TestRepository]("repo1"))

	// 获取所有 BeanDef
	beanGet := container.(BeanGet)
	defs := beanGet.ListBeans()

	if len(defs) != 3 {
		t.Errorf("Expected 3 bean defs, got %d", len(defs))
	}
}

func TestHasTypeWithParentContainer(t *testing.T) {
	t.Parallel()
	parent := NewContainer()
	Register[*TestService](parent, WithName[*TestService]("parentSvc"))

	child := NewContainer()
	child.(ContainerExt).SetParent(parent)

	// HasType 只检查当前容器，不检查父容器
	if child.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return false for child container (does not check parent)")
	}

	// 但可以通过父容器获取
	if !parent.HasType(reflect.TypeOf((*TestService)(nil))) {
		t.Error("Expected HasType to return true for parent container")
	}
}
