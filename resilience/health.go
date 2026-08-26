package resilience

import (
	"fmt"
	"sync"
)

// HealthAware 健康感知负载均衡器
type HealthAware struct {
	mu      sync.RWMutex
	inner   Balancer
	failure map[string]int
}

// NewHealthAware 创建健康感知负载均衡器
func NewHealthAware(inner Balancer) (*HealthAware, error) {
	if inner == nil {
		return nil, fmt.Errorf("resilience: inner balancer must not be nil")
	}
	return &HealthAware{
		inner:   inner,
		failure: make(map[string]int),
	}, nil
}

// MustNewHealthAware 创建健康感知负载均衡器，失败则 panic。
func MustNewHealthAware(inner Balancer) *HealthAware {
	h, err := NewHealthAware(inner)
	if err != nil {
		panic(err)
	}
	return h
}

// Next 选择健康的后端
func (ha *HealthAware) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	ha.mu.Lock()
	defer ha.mu.Unlock()

	healthyBackends := make([]*ServiceInstance, 0, len(backends))
	for _, b := range backends {
		if b != nil && b.Health == HealthUp {
			healthyBackends = append(healthyBackends, b)
		}
	}

	if len(healthyBackends) == 0 {
		healthyBackends = nonNilBackends(backends)
	}

	if len(healthyBackends) == 0 {
		return nil, ErrNoBackends
	}

	return ha.inner.Next(healthyBackends)
}

// RecordFailure 记录后端失败
func (ha *HealthAware) RecordFailure(backendURL string) {
	ha.mu.Lock()
	defer ha.mu.Unlock()

	ha.failure[backendURL]++
}

// RecordSuccess 记录后端成功
func (ha *HealthAware) RecordSuccess(backendURL string) {
	ha.mu.Lock()
	defer ha.mu.Unlock()

	delete(ha.failure, backendURL)
}

// GetFailureCount 获取失败计数
func (ha *HealthAware) GetFailureCount(backendURL string) int {
	ha.mu.RLock()
	defer ha.mu.RUnlock()

	return ha.failure[backendURL]
}
