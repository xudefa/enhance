package resilience

import "sync"

// StickySession 会话保持负载均衡器
type StickySession struct {
	mu sync.RWMutex
	// sessionCookieName 会话 Cookie 名称
	sessionCookieName string
	// sessionBackendMap 会话到后端的映射
	sessionBackendMap map[string]*ServiceInstance
}

// NewStickySession 创建会话保持负载均衡器
func NewStickySession(sessionCookieName string) *StickySession {
	if sessionCookieName == "" {
		sessionCookieName = "JSESSIONID"
	}

	return &StickySession{
		sessionCookieName: sessionCookieName,
		sessionBackendMap: make(map[string]*ServiceInstance),
	}
}

// Next 选择后端（不带会话信息）
func (ss *StickySession) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	return backends[0], nil
}

// NextWithSession 根据会话 ID 选择后端
func (ss *StickySession) NextWithSession(backends []*ServiceInstance, sessionID string) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	if sessionID == "" {
		return backends[0], nil
	}

	ss.mu.RLock()
	backend, exists := ss.sessionBackendMap[sessionID]
	ss.mu.RUnlock()

	if exists {
		for _, b := range backends {
			if b.URL == backend.URL {
				return b, nil
			}
		}
		ss.mu.Lock()
		delete(ss.sessionBackendMap, sessionID)
		ss.mu.Unlock()
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	idx := int(hashKey(sessionID) % uint32(len(backends)))
	backend = backends[idx]
	ss.sessionBackendMap[sessionID] = backend

	return backend, nil
}

// GetSessionBackend 获取会话绑定的后端
func (ss *StickySession) GetSessionBackend(sessionID string) (*ServiceInstance, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	backend, exists := ss.sessionBackendMap[sessionID]
	return backend, exists
}

// RemoveSession 移除会话绑定
func (ss *StickySession) RemoveSession(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.sessionBackendMap, sessionID)
}

// GetSessionCount 获取会话数量
func (ss *StickySession) GetSessionCount() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return len(ss.sessionBackendMap)
}
