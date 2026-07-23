package refresh

import (
	"sync"
	"time"

	"github.com/xudefa/enhance/config/environment"
)

// refreshManager 刷新管理器实现。
type refreshManager struct {
	mu        sync.RWMutex
	env       *environment.Environment
	listeners []RefreshListener
	running   bool
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

func (m *refreshManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

func (m *refreshManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

func (m *refreshManager) Refresh() error {
	m.mu.RLock()
	listeners := make([]RefreshListener, len(m.listeners))
	copy(listeners, m.listeners)
	env := m.env
	m.mu.RUnlock()

	event := RefreshEvent{
		Environment: env,
		Timestamp:   time.Now().UnixMilli(),
	}

	for _, listener := range listeners {
		if err := listener.OnRefresh(event); err != nil {
			return err
		}
	}

	return nil
}

func (m *refreshManager) AddRefreshListener(listener RefreshListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}
