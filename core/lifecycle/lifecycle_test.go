package lifecycle

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

type TestLifecycleBean struct {
	InitCalled    bool
	DestroyCalled bool
	InitError     error
	DestroyError  error
}

func (b *TestLifecycleBean) Init() error {
	b.InitCalled = true
	return b.InitError
}

func (b *TestLifecycleBean) Destroy() error {
	b.DestroyCalled = true
	return b.DestroyError
}

func TestInvokeInitWithFunc(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	initCalled := false
	bean := &TestLifecycleBean{}

	err := mgr.InvokeInit("testBean", bean, func(b any) error {
		initCalled = true
		return nil
	})

	if err != nil {
		t.Fatalf("InvokeInit failed: %v", err)
	}

	if !initCalled {
		t.Error("Expected init callback to be called")
	}

	if !bean.InitCalled {
		t.Error("Expected LifecycleBean.Init() to be called")
	}
}

func TestInvokeInitWithLifecycleBean(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	bean := &TestLifecycleBean{}

	err := mgr.InvokeInit("testBean", bean, nil)

	if err != nil {
		t.Fatalf("InvokeInit failed: %v", err)
	}

	if !bean.InitCalled {
		t.Error("Expected LifecycleBean.Init() to be called")
	}
}

func TestInvokeInitError(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	expectedErr := fmt.Errorf("init error")
	bean := &TestLifecycleBean{InitError: expectedErr}

	err := mgr.InvokeInit("testBean", bean, nil)

	if err == nil {
		t.Fatal("Expected InvokeInit to fail")
	}

	if err != expectedErr {
		t.Errorf("Expected init error, got: %v", err)
	}
}

func TestInvokeInitFuncError(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	expectedErr := fmt.Errorf("func init error")
	bean := &TestLifecycleBean{}

	err := mgr.InvokeInit("testBean", bean, func(b any) error {
		return expectedErr
	})

	if err == nil {
		t.Fatal("Expected InvokeInit to fail")
	}

	if err != expectedErr {
		t.Errorf("Expected func init error, got: %v", err)
	}
}

func TestInvokeDestroyWithFunc(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	destroyCalled := false
	bean := &TestLifecycleBean{}

	err := mgr.InvokeDestroy("testBean", bean, func(b any) error {
		destroyCalled = true
		return nil
	})

	if err != nil {
		t.Fatalf("InvokeDestroy failed: %v", err)
	}

	if !destroyCalled {
		t.Error("Expected destroy callback to be called")
	}

	if !bean.DestroyCalled {
		t.Error("Expected LifecycleBean.Destroy() to be called")
	}
}

func TestInvokeDestroyWithLifecycleBean(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	bean := &TestLifecycleBean{}

	err := mgr.InvokeDestroy("testBean", bean, nil)

	if err != nil {
		t.Fatalf("InvokeDestroy failed: %v", err)
	}

	if !bean.DestroyCalled {
		t.Error("Expected LifecycleBean.Destroy() to be called")
	}
}

func TestInvokeDestroyError(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	expectedErr := fmt.Errorf("destroy error")
	bean := &TestLifecycleBean{DestroyError: expectedErr}

	err := mgr.InvokeDestroy("testBean", bean, nil)

	if err == nil {
		t.Fatal("Expected InvokeDestroy to fail")
	}

	if err != expectedErr {
		t.Errorf("Expected destroy error, got: %v", err)
	}
}

func TestDestroyAll(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	var destroyOrder []string
	var mu sync.Mutex

	bean1 := &TestLifecycleBean{}
	bean2 := &TestLifecycleBean{}
	bean3 := &TestLifecycleBean{}

	mgr.RegisterBean("bean1", bean1, func(b any) error {
		mu.Lock()
		defer mu.Unlock()
		destroyOrder = append(destroyOrder, "bean1")
		return nil
	})

	mgr.RegisterBean("bean2", bean2, func(b any) error {
		mu.Lock()
		defer mu.Unlock()
		destroyOrder = append(destroyOrder, "bean2")
		return nil
	})

	mgr.RegisterBean("bean3", bean3, func(b any) error {
		mu.Lock()
		defer mu.Unlock()
		destroyOrder = append(destroyOrder, "bean3")
		return nil
	})

	err := mgr.DestroyAll()
	if err != nil {
		t.Fatalf("DestroyAll failed: %v", err)
	}

	// Verify reverse order destruction
	if len(destroyOrder) != 3 {
		t.Fatalf("Expected 3 destroy calls, got %d", len(destroyOrder))
	}

	if destroyOrder[0] != "bean3" || destroyOrder[1] != "bean2" || destroyOrder[2] != "bean1" {
		t.Errorf("Expected reverse order destroy [bean3, bean2, bean1], got %v", destroyOrder)
	}

	// Verify all beans were destroyed
	if !bean1.DestroyCalled || !bean2.DestroyCalled || !bean3.DestroyCalled {
		t.Error("Expected all beans to be destroyed")
	}
}

func TestDestroyAllWithError(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	expectedErr := fmt.Errorf("destroy error")

	bean1 := &TestLifecycleBean{}
	bean2 := &TestLifecycleBean{DestroyError: expectedErr}
	bean3 := &TestLifecycleBean{}

	mgr.RegisterBean("bean1", bean1, nil)
	mgr.RegisterBean("bean2", bean2, nil)
	mgr.RegisterBean("bean3", bean3, nil)

	err := mgr.DestroyAll()

	if err == nil {
		t.Fatal("Expected DestroyAll to fail")
	}

	if err != expectedErr {
		t.Errorf("Expected destroy error, got: %v", err)
	}

	// All beans should still be destroyed despite error
	if !bean1.DestroyCalled || !bean2.DestroyCalled || !bean3.DestroyCalled {
		t.Error("Expected all beans to be destroyed despite error")
	}
}

func TestDestroyAllEmpty(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	err := mgr.DestroyAll()
	if err != nil {
		t.Errorf("Expected no error for empty destroy, got: %v", err)
	}
}

func TestDestroyAllTwice(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	callCount := int32(0)
	bean := &TestLifecycleBean{}

	mgr.RegisterBean("bean", bean, func(b any) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	})

	// First destroy
	err := mgr.DestroyAll()
	if err != nil {
		t.Fatalf("First DestroyAll failed: %v", err)
	}

	firstCount := atomic.LoadInt32(&callCount)

	// Second destroy should not call callbacks again
	err = mgr.DestroyAll()
	if err != nil {
		t.Fatalf("Second DestroyAll failed: %v", err)
	}

	if atomic.LoadInt32(&callCount) != firstCount {
		t.Errorf("Destroy callback should not be called again, expected %d, got %d", firstCount, atomic.LoadInt32(&callCount))
	}
}

func TestLifecycleListener(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	var phases []Phase
	var beanNames []string
	var mu sync.Mutex

	listener := LifecycleListenerFunc(func(beanName string, bean any, phase Phase) {
		mu.Lock()
		defer mu.Unlock()
		phases = append(phases, phase)
		beanNames = append(beanNames, beanName)
	})

	mgr.RegisterListener(listener)

	bean := &TestLifecycleBean{}
	mgr.InvokeInit("testBean", bean, nil)

	mu.Lock()
	phaseCount := len(phases)
	mu.Unlock()

	if phaseCount != 1 {
		t.Errorf("Expected 1 phase notification, got %d", phaseCount)
	}

	if phases[0] != PhaseInitialized {
		t.Errorf("Expected PhaseInitialized, got %v", phases[0])
	}

	if beanNames[0] != "testBean" {
		t.Errorf("Expected bean name 'testBean', got '%s'", beanNames[0])
	}
}

func TestMultipleListeners(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	var callCount1, callCount2 int32

	listener1 := LifecycleListenerFunc(func(beanName string, bean any, phase Phase) {
		atomic.AddInt32(&callCount1, 1)
	})

	listener2 := LifecycleListenerFunc(func(beanName string, bean any, phase Phase) {
		atomic.AddInt32(&callCount2, 1)
	})

	mgr.RegisterListener(listener1)
	mgr.RegisterListener(listener2)

	bean := &TestLifecycleBean{}
	mgr.InvokeInit("testBean", bean, nil)

	if atomic.LoadInt32(&callCount1) != 1 {
		t.Errorf("Expected listener1 to be called once, got %d", atomic.LoadInt32(&callCount1))
	}

	if atomic.LoadInt32(&callCount2) != 1 {
		t.Errorf("Expected listener2 to be called once, got %d", atomic.LoadInt32(&callCount2))
	}
}

func TestNotifyPhaseChange(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	var receivedPhase Phase
	var receivedBeanName string
	var mu sync.Mutex

	listener := LifecycleListenerFunc(func(beanName string, bean any, phase Phase) {
		mu.Lock()
		defer mu.Unlock()
		receivedPhase = phase
		receivedBeanName = beanName
	})

	mgr.RegisterListener(listener)

	bean := &TestLifecycleBean{}
	mgr.NotifyPhaseChange("myBean", bean, PhaseDestroyed)

	mu.Lock()
	defer mu.Unlock()

	if receivedPhase != PhaseDestroyed {
		t.Errorf("Expected PhaseDestroyed, got %v", receivedPhase)
	}

	if receivedBeanName != "myBean" {
		t.Errorf("Expected bean name 'myBean', got '%s'", receivedBeanName)
	}
}

func TestConcurrentListeners(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	var callCount int32

	for i := 0; i < 10; i++ {
		listener := LifecycleListenerFunc(func(beanName string, bean any, phase Phase) {
			atomic.AddInt32(&callCount, 1)
		})
		mgr.RegisterListener(listener)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bean := &TestLifecycleBean{}
			mgr.NotifyPhaseChange("testBean", bean, PhaseInitialized)
		}()
	}
	wg.Wait()

	// 10 listeners * 100 notifications = 1000 calls
	expected := int32(1000)
	if atomic.LoadInt32(&callCount) != expected {
		t.Errorf("Expected %d listener calls, got %d", expected, atomic.LoadInt32(&callCount))
	}
}

func TestRegisterBeanConcurrent(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			bean := &TestLifecycleBean{}
			mgr.RegisterBean("bean", bean, nil)
		}(i)
	}
	wg.Wait()
}
