package resilience

import "sync/atomic"

// LeastConnections 最少连接负载均衡器
type LeastConnections struct{}

// NewLeastConnections 创建最少连接负载均衡器
func NewLeastConnections() *LeastConnections {
	return &LeastConnections{}
}

// Next 选择连接数最少的后端
func (lc *LeastConnections) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	var selected *ServiceInstance
	minActive := int64(^uint64(0) >> 1)

	for _, b := range backends {
		active := atomic.LoadInt64(&b.Active)
		if active < minActive {
			minActive = active
			selected = b
		}
	}

	return selected, nil
}
