package resilience

import "math/rand"

// Random 随机负载均衡器
type Random struct{}

// NewRandom 创建随机负载均衡器
func NewRandom() *Random {
	return &Random{}
}

// Next 随机选择一个后端
func (r *Random) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	idx := rand.Intn(len(backends))
	return backends[idx], nil
}
