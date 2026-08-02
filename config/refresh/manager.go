package refresh

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/xudefa/enhance/config/environment"
)

// refreshManager 刷新管理器实现。
type refreshManager struct {
	mu           sync.RWMutex
	env          *environment.Environment
	listeners    []RefreshListener
	running      bool
	lastSnapshot map[string]any
}

// NewRefreshManager 创建刷新管理器。
//
// 参数：
//   - env: 环境配置，刷新时会关联到 RefreshEvent 中
func NewRefreshManager(env *environment.Environment) RefreshManager {
	return &refreshManager{
		env:       env,
		listeners: make([]RefreshListener, 0),
	}
}

// Start 启动刷新管理器，允许执行刷新操作。
func (m *refreshManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

// Stop 停止刷新管理器，禁止执行刷新操作。
func (m *refreshManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

// Refresh 触发配置刷新，通知所有已注册的监听器。
//
// 管理器未启动（Stop 后）时直接返回，不执行刷新。
func (m *refreshManager) Refresh() error {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return nil
	}
	listeners := make([]RefreshListener, len(m.listeners))
	copy(listeners, m.listeners)
	env := m.env
	m.mu.RUnlock()

	event := RefreshEvent{
		Environment: env,
		ChangedKeys: m.collectChangedKeys(),
		Timestamp:   time.Now().UnixMilli(),
	}

	var errs []error
	for _, listener := range listeners {
		if err := listener.OnRefresh(event); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// collectChangedKeys 收集自上次刷新以来发生变化的配置键。
//
// 通过对比当前环境属性值与上一次快照，检测哪些键发生了变化。
// 对于支持键枚举的配置源（MapPropertySource、JSONPropertySource），
// 首次调用时会建立初始快照。
func (m *refreshManager) collectChangedKeys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot := m.lastSnapshot
	if snapshot == nil {
		m.initSnapshotLocked()
		return nil
	}

	var changed []string

	// 检测已有键的变化或删除
	for key, oldVal := range snapshot {
		if newVal, ok := m.env.GetProperty(key); ok {
			if !reflect.DeepEqual(oldVal, newVal) {
				changed = append(changed, key)
				snapshot[key] = newVal
			}
		} else {
			changed = append(changed, key)
			delete(snapshot, key)
		}
	}

	// 检测新增的键
	type keyEnumerator interface {
		Keys() []string
	}
	for _, source := range m.env.GetPropertySources() {
		if ks, ok := source.(keyEnumerator); ok {
			for _, key := range ks.Keys() {
				if _, exists := snapshot[key]; !exists {
					if val, ok := m.env.GetProperty(key); ok {
						snapshot[key] = val
						changed = append(changed, key)
					}
				}
			}
		}
	}

	return changed
}

func (m *refreshManager) initSnapshotLocked() {
	if m.lastSnapshot != nil {
		return
	}

	m.lastSnapshot = make(map[string]any)
	type keyEnumerator interface {
		Keys() []string
	}
	for _, source := range m.env.GetPropertySources() {
		if ks, ok := source.(keyEnumerator); ok {
			for _, key := range ks.Keys() {
				if val, ok := m.env.GetProperty(key); ok {
					m.lastSnapshot[key] = val
				}
			}
		}
	}
}

// AddRefreshListener 注册配置刷新监听器，刷新时将收到通知。
func (m *refreshManager) AddRefreshListener(listener RefreshListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}
