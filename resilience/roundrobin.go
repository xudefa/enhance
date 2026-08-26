package resilience

import (
	"sync/atomic"
)

// nonNilBackends 过滤掉 nil 的后端，避免空指针解引用。
func nonNilBackends(backends []*ServiceInstance) []*ServiceInstance {
	nonNil := make([]*ServiceInstance, 0, len(backends))
	for _, b := range backends {
		if b != nil {
			nonNil = append(nonNil, b)
		}
	}
	return nonNil
}

// RoundRobin 轮询负载均衡器
type RoundRobin struct {
	current atomic.Int64
}

// NewRoundRobin 创建轮询负载均衡器
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next 选择下一个后端
func (rr *RoundRobin) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	idx := (rr.current.Add(1) - 1) % int64(len(backends))
	return backends[idx], nil
}

// WeightedRoundRobin 加权轮询负载均衡器
type WeightedRoundRobin struct {
	current atomic.Int64
}

// NewWeightedRoundRobin 创建加权轮询负载均衡器
func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{}
}

// Next 根据权重选择后端
func (wrr *WeightedRoundRobin) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	totalWeight := 0
	for _, b := range backends {
		w := b.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}

	if totalWeight <= 0 {
		return backends[0], nil
	}

	current := (wrr.current.Add(1) - 1) % int64(totalWeight)

	accumulated := int64(0)
	for _, b := range backends {
		w := int64(b.Weight)
		if w <= 0 {
			w = 1
		}
		accumulated += w
		if current < accumulated {
			return b, nil
		}
	}

	return backends[0], nil
}
