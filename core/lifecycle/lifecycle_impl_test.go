package lifecycle

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestLifecycleListenerFunc_OnPhaseChange(t *testing.T) {
	t.Parallel()
	var called atomic.Bool
	listener := LifecycleListenerFunc(func(beanName string, bean any, phase Phase) {
		called.Store(true)
	})
	listener.OnPhaseChange("test", nil, PhaseInitialized)
	if !called.Load() {
		t.Error("expected listener function to be called")
	}
}

func TestNewLifecycleManager(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()
	if mgr == nil {
		t.Fatal("expected non-nil lifecycle manager")
	}
}

func TestRegisterListener_Concurrent(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.RegisterListener(LifecycleListenerFunc(func(string, any, Phase) {}))
		}()
	}
	wg.Wait()

	// verify listeners are registered by triggering a notification
	var count atomic.Int32
	mgr.RegisterListener(LifecycleListenerFunc(func(string, any, Phase) {
		count.Add(1)
	}))
	mgr.NotifyPhaseChange("test", nil, PhaseInitialized)
	if count.Load() != 1 {
		t.Errorf("expected 1 listener call, got %d", count.Load())
	}
}

func TestInvokeInit_NilInitFunc_NonLifecycleBean(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	type PlainBean struct{ Name string }
	bean := &PlainBean{Name: "plain"}

	err := mgr.InvokeInit("test", bean, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvokeDestroy_NilFunc_NonLifecycleBean(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	type PlainBean struct{ Name string }
	bean := &PlainBean{Name: "plain"}

	err := mgr.InvokeDestroy("test", bean, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvokeDestroy_BothFuncAndInterface(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	var funcCalled, ifaceCalled bool
	bean := &TestLifecycleBean{}

	err := mgr.InvokeDestroy("test", bean, func(b any) error {
		funcCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !funcCalled {
		t.Error("expected destroy func to be called")
	}
	if !bean.DestroyCalled {
		t.Error("expected LifecycleBean.Destroy to be called")
	}
	_ = ifaceCalled
}

func TestRegisterBean_CreatesRecord(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	bean := &TestLifecycleBean{}
	mgr.RegisterBean("myBean", bean, nil)

	err := mgr.DestroyAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bean.DestroyCalled {
		t.Error("expected bean to be destroyed via DestroyAll")
	}
}

func TestDestroyAll_MultipleErrors(t *testing.T) {
	t.Parallel()
	mgr := NewLifecycleManager()

	bean1 := &TestLifecycleBean{DestroyError: &testError{"err1"}}
	bean2 := &TestLifecycleBean{DestroyError: &testError{"err2"}}

	mgr.RegisterBean("b1", bean1, nil)
	mgr.RegisterBean("b2", bean2, nil)

	err := mgr.DestroyAll()
	if err == nil {
		t.Fatal("expected error from DestroyAll")
	}
	// both beans should be destroyed despite errors
	if !bean1.DestroyCalled || !bean2.DestroyCalled {
		t.Error("expected both beans to be destroyed")
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
