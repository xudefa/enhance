package lifecycle

import (
	"fmt"
)

// String 返回阶段的字符串表示。
func (p ApplicationPhase) String() string {
	if name, ok := phaseNames[p]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", p)
}

// NewLifecycleManager 创建生命周期管理器。
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		phase: PhaseInit,
	}
}

// GetPhase 返回当前生命周期阶段
func (m *LifecycleManager) GetPhase() ApplicationPhase {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.phase
}

// SetPhase 设置新的生命周期阶段并通知所有监听器
func (m *LifecycleManager) SetPhase(newPhase ApplicationPhase) error {
	m.mu.Lock()
	oldPhase := m.phase
	if !isForwardTransition(oldPhase, newPhase) {
		m.mu.Unlock()
		return fmt.Errorf("无效的阶段转换，从 %s 到 %s", oldPhase, newPhase)
	}
	m.phase = newPhase
	listeners := make([]PhaseListener, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.Unlock()

	var err error
	for _, listener := range listeners {
		if e := listener.OnPhaseChange(oldPhase, newPhase); e != nil && err == nil {
			err = e
		}
	}

	if err != nil && m.onError != nil {
		m.mu.RLock()
		handler := m.onError
		m.mu.RUnlock()
		if handler != nil {
			handler(oldPhase, newPhase, err)
		}
	}

	return err
}

// AddListener 添加生命周期阶段监听器
func (m *LifecycleManager) AddListener(listener PhaseListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// SetErrorHandler 设置错误处理回调
func (m *LifecycleManager) SetErrorHandler(handler func(oldPhase, newPhase ApplicationPhase, err error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onError = handler
}

// NewLifecycleBuilder 创建生命周期构建器。
func NewLifecycleBuilder() *LifecycleBuilder {
	return &LifecycleBuilder{
		initialPhase: PhaseInit,
		listeners:    make([]PhaseListener, 0),
	}
}

// InitialPhase 设置初始阶段
func (b *LifecycleBuilder) InitialPhase(phase ApplicationPhase) *LifecycleBuilder {
	b.initialPhase = phase
	return b
}

// Listener 添加监听器
func (b *LifecycleBuilder) Listener(listener PhaseListener) *LifecycleBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners = append(b.listeners, listener)
	return b
}

// OnError 设置错误处理回调
func (b *LifecycleBuilder) OnError(handler func(oldPhase, newPhase ApplicationPhase, err error)) *LifecycleBuilder {
	b.onError = handler
	return b
}

// Build 构建生命周期管理器
func (b *LifecycleBuilder) Build() *LifecycleManager {
	mgr := NewLifecycleManager()
	mgr.phase = b.initialPhase

	for _, listener := range b.listeners {
		mgr.AddListener(listener)
	}

	if b.onError != nil {
		mgr.SetErrorHandler(b.onError)
	}

	return mgr
}

// String 返回转换描述。
func (t PhaseTransition) String() string {
	return fmt.Sprintf("%s -> %s", t.OldPhase, t.NewPhase)
}
