package resilience

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// hashRing 一致性哈希环快照（不可变，用于原子指针替换）。
type hashRing struct {
	ring    []uint32          // 有序哈希环
	nodes   map[uint32]string // hash -> 物理节点ID
	weights map[string]int    // 物理节点ID -> 权重
}

// ConsistentHash 一致性哈希负载均衡器（真 hash ring 实现）。
//
// 使用虚拟节点（replicas）将物理节点映射到 2^32 环上，
// 支持增删节点时的最小数据迁移。
// 使用原子指针换 ring 实现无锁读取。
type ConsistentHash struct {
	replicas int
	ringPtr  atomic.Pointer[hashRing]
	mu       sync.Mutex // 仅用于写操作
}

// NewConsistentHash 创建一致性哈希负载均衡器。
func NewConsistentHash(replicas ...int) *ConsistentHash {
	r := 150
	if len(replicas) > 0 {
		r = replicas[0]
	}

	ch := &ConsistentHash{
		replicas: r,
	}
	ch.ringPtr.Store(&hashRing{
		ring:    make([]uint32, 0),
		nodes:   make(map[uint32]string),
		weights: make(map[string]int),
	})
	return ch
}

// AddNode 向 hash ring 添加节点（带权重）。
func (ch *ConsistentHash) AddNode(nodeID string, weight int) {
	if weight <= 0 {
		weight = 1
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	oldRing := ch.ringPtr.Load()
	newRing := &hashRing{
		ring:    make([]uint32, 0, len(oldRing.ring)+ch.replicas*weight),
		nodes:   make(map[uint32]string, len(oldRing.nodes)+ch.replicas*weight),
		weights: make(map[string]int, len(oldRing.weights)+1),
	}

	// 复制旧数据
	copy(newRing.ring, oldRing.ring)
	for k, v := range oldRing.nodes {
		newRing.nodes[k] = v
	}
	for k, v := range oldRing.weights {
		newRing.weights[k] = v
	}

	// 添加新节点
	newRing.weights[nodeID] = weight
	vnCount := ch.replicas * weight
	for i := 0; i < vnCount; i++ {
		hash := hashKey(fmt.Sprintf("%s:%d", nodeID, i))
		newRing.ring = append(newRing.ring, hash)
		newRing.nodes[hash] = nodeID
	}

	sort.Slice(newRing.ring, func(i, j int) bool { return newRing.ring[i] < newRing.ring[j] })

	// 原子替换 ring
	ch.ringPtr.Store(newRing)
}

// RemoveNode 从 hash ring 移除节点。
func (ch *ConsistentHash) RemoveNode(nodeID string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	oldRing := ch.ringPtr.Load()
	newRing := &hashRing{
		ring:    make([]uint32, 0, len(oldRing.ring)),
		nodes:   make(map[uint32]string, len(oldRing.nodes)),
		weights: make(map[string]int, len(oldRing.weights)),
	}

	// 复制旧数据，排除被移除的节点
	for _, hash := range oldRing.ring {
		if oldRing.nodes[hash] != nodeID {
			newRing.ring = append(newRing.ring, hash)
			newRing.nodes[hash] = oldRing.nodes[hash]
		}
	}
	for k, v := range oldRing.weights {
		if k != nodeID {
			newRing.weights[k] = v
		}
	}

	// 原子替换 ring
	ch.ringPtr.Store(newRing)
}

// GetNode 根据 key 获取对应的节点 ID。
func (ch *ConsistentHash) GetNode(key string) string {
	ring := ch.ringPtr.Load()

	if len(ring.ring) == 0 {
		return ""
	}

	hash := hashKey(key)
	idx := sort.Search(len(ring.ring), func(i int) bool { return ring.ring[i] >= hash })
	if idx >= len(ring.ring) {
		idx = 0
	}
	return ring.nodes[ring.ring[idx]]
}

// Next 使用一致性哈希选择后端（无 key 时随机选择）。
func (ch *ConsistentHash) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	// 无 key 时回退到第一个可用后端
	return backends[0], nil
}

// NextByKey 根据键选择后端（真 hash ring）。
func (ch *ConsistentHash) NextByKey(backends []*ServiceInstance, key string) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	// 如果 ring 为空，根据 backends 动态构建
	ring := ch.ringPtr.Load()
	if len(ring.ring) == 0 {
		ch.mu.Lock()
		// 双重检查
		ring = ch.ringPtr.Load()
		if len(ring.ring) == 0 {
			newRing := &hashRing{
				ring:    make([]uint32, 0, len(backends)*ch.replicas),
				nodes:   make(map[uint32]string, len(backends)*ch.replicas),
				weights: make(map[string]int, len(backends)),
			}
			for _, b := range backends {
				ch.addNodeTo(newRing, b.ID, 1)
			}
			ch.ringPtr.Store(newRing)
		}
		ch.mu.Unlock()
	}

	nodeID := ch.GetNode(key)
	for _, b := range backends {
		if b.ID == nodeID {
			return b, nil
		}
	}
	// fallback
	return backends[0], nil
}

func (ch *ConsistentHash) addNodeTo(ring *hashRing, nodeID string, weight int) {
	if weight <= 0 {
		weight = 1
	}
	ring.weights[nodeID] = weight
	vnCount := ch.replicas * weight
	for i := 0; i < vnCount; i++ {
		hash := hashKey(fmt.Sprintf("%s:%d", nodeID, i))
		ring.ring = append(ring.ring, hash)
		ring.nodes[hash] = nodeID
	}
	sort.Slice(ring.ring, func(i, j int) bool { return ring.ring[i] < ring.ring[j] })
}

// hashKey 使用 MD5 生成 32-bit 哈希值。
func hashKey(key string) uint32 {
	sum := md5.Sum([]byte(key))
	return binary.BigEndian.Uint32(sum[:4])
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

	hash := hashKey(clientIP)
	idx := int(hash % uint32(len(backends)))

	return backends[idx], nil
}
