package resilience

import (
	"math/rand"
)

// ConsistentHash 一致性哈希负载均衡器。
// 无状态设计，无需锁保护。
type ConsistentHash struct {
	replicas int
	hashFunc func(string) uint32
}

// NewConsistentHash 创建一致性哈希负载均衡器。
func NewConsistentHash(replicas ...int) *ConsistentHash {
	r := 150
	if len(replicas) > 0 {
		r = replicas[0]
	}

	return &ConsistentHash{
		replicas: r,
		hashFunc: simpleHash,
	}
}

// Next 使用一致性哈希选择后端。
func (ch *ConsistentHash) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	idx := rand.Intn(len(backends))
	return backends[idx], nil
}

// NextByKey 根据键选择后端。
func (ch *ConsistentHash) NextByKey(backends []*ServiceInstance, key string) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	hash := ch.hashFunc(key)
	idx := int(hash % uint32(len(backends)))
	return backends[idx], nil
}

// simpleHash 简单的哈希函数。
func simpleHash(key string) uint32 {
	var hash uint32
	for _, c := range key {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// IPHash IP 哈希负载均衡器。
// 无状态设计，无需锁保护。
type IPHash struct{}

// NewIPHash 创建 IP 哈希负载均衡器。
func NewIPHash() *IPHash {
	return &IPHash{}
}

// Next 选择后端（不带 IP 信息）。
func (ih *IPHash) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	return backends[0], nil
}

// NextByIP 根据客户端 IP 选择后端。
func (ih *IPHash) NextByIP(backends []*ServiceInstance, clientIP string) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	if clientIP == "" {
		return backends[0], nil
	}

	hash := simpleHash(clientIP)
	idx := int(hash % uint32(len(backends)))

	return backends[idx], nil
}
