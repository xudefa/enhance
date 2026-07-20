package resilience

import (
	"sync"
)

// RoundRobin 轮询负载均衡器
type RoundRobin struct {
	mu      sync.Mutex
	current int64
}

// NewRoundRobin 创建轮询负载均衡器
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next 选择下一个后端
func (rr *RoundRobin) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	rr.mu.Lock()
	defer rr.mu.Unlock()

	idx := rr.current % int64(len(backends))
	rr.current++
	return backends[idx], nil
}

// WeightedRoundRobin 加权轮询负载均衡器
type WeightedRoundRobin struct {
	mu      sync.Mutex
	current int64
}

// NewWeightedRoundRobin 创建加权轮询负载均衡器
func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{}
}

// Next 根据权重选择后端
func (wrr *WeightedRoundRobin) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	totalWeight := 0
	for _, b := range backends {
		totalWeight += b.Weight
	}

	if totalWeight == 0 {
		return backends[0], nil
	}

	current := wrr.current % int64(totalWeight)
	wrr.current++

	accumulated := int64(0)
	for _, b := range backends {
		accumulated += int64(b.Weight)
		if current < accumulated {
			return b, nil
		}
	}

	return backends[0], nil
}
