package core

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/xudefa/enhance/core/lifecycle"
	"github.com/xudefa/enhance/core/registry"
)

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

	if !bean.InitCalled {
		t.Error("Expected LifecycleBean.Init() to be called")
	}

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

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

	if !errors.Is(err, expectedErr) {
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

	if !errors.Is(err, expectedErr) {
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

	err = container.Destroy()
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

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

	if !errors.Is(err, expectedErr) {
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

	// 验证监听器注册成功
	mgr.NotifyPhaseChange("test", &TestService{}, lifecycle.PhaseInitialized)
	phases := recorder.GetPhases()
	if len(phases) == 0 {
		t.Error("Expected phase notifications")
	}
}
