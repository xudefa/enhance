package lifecycle

import (
	"sync"
)

// OnPhaseChange 实现 LifecycleListener 接口。
func (f LifecycleListenerFunc) OnPhaseChange(beanName string, bean any, phase Phase) {
	f(beanName, bean, phase)
}

// defaultLifecycleManager 默认生命周期管理器实现。
type defaultLifecycleManager struct {
	mu          sync.RWMutex
	listeners   []LifecycleListener
	beanRecords []beanRecord
}

// beanRecord 记录 Bean 的生命周期信息。
type beanRecord struct {
	name        string
	instance    any
	destroyFunc func(any) error
}

// RegisterListener 注册生命周期监听器。
func (m *defaultLifecycleManager) RegisterListener(listener LifecycleListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// NotifyPhaseChange 通知生命周期阶段变化。
func (m *defaultLifecycleManager) NotifyPhaseChange(beanName string, bean any, phase Phase) {
	m.mu.RLock()
	listeners := make([]LifecycleListener, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, listener := range listeners {
		listener.OnPhaseChange(beanName, bean, phase)
	}
}

// InvokeInit 调用 Bean 的初始化回调。
func (m *defaultLifecycleManager) InvokeInit(beanName string, bean any, initFunc func(any) error) error {
	// 调用函数式回调
	if initFunc != nil {
		if err := initFunc(bean); err != nil {
			return err
		}
	}

	// 调用 LifecycleBean 接口
	if lifecycleBean, ok := bean.(LifecycleBean); ok {
		if err := lifecycleBean.Init(); err != nil {
			return err
		}
	}

	m.NotifyPhaseChange(beanName, bean, PhaseInitialized)
	return nil
}

// InvokeDestroy 调用 Bean 的销毁回调。
func (m *defaultLifecycleManager) InvokeDestroy(beanName string, bean any, destroyFunc func(any) error) error {
	// 调用函数式回调
	if destroyFunc != nil {
		if err := destroyFunc(bean); err != nil {
			return err
		}
	}

	// 调用 LifecycleBean 接口
	if lifecycleBean, ok := bean.(LifecycleBean); ok {
		if err := lifecycleBean.Destroy(); err != nil {
			return err
		}
	}

	m.NotifyPhaseChange(beanName, bean, PhaseDestroyed)
	return nil
}

// DestroyAll 销毁所有已注册的 Bean。
func (m *defaultLifecycleManager) DestroyAll() error {
	m.mu.Lock()
	records := make([]beanRecord, len(m.beanRecords))
	copy(records, m.beanRecords)
	m.beanRecords = nil
	m.mu.Unlock()

	var firstErr error
	// 逆序销毁
	for i := len(records) - 1; i >= 0; i-- {
		record := records[i]
		if err := m.InvokeDestroy(record.name, record.instance, record.destroyFunc); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// RegisterBean 注册 Bean 记录（内部使用）。
func (m *defaultLifecycleManager) RegisterBean(name string, instance any, destroyFunc func(any) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.beanRecords = append(m.beanRecords, beanRecord{
		name:        name,
		instance:    instance,
		destroyFunc: destroyFunc,
	})
}

// NewLifecycleManager 创建生命周期管理器实例。
func NewLifecycleManager() LifecycleManager {
	return &defaultLifecycleManager{
		listeners:   make([]LifecycleListener, 0),
		beanRecords: make([]beanRecord, 0),
	}
}
