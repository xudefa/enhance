// Package resilience 提供弹性容错功能，用于 enhance 框架。
package resilience

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
)

// roundRobinSelectorImpl Selector 接口的轮询实现。
type roundRobinSelectorImpl struct {
	counter atomic.Uint64
}

// randomSelectorImpl Selector 接口的随机实现。
type randomSelectorImpl struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// weightedRandomSelectorImpl Selector 接口的加权随机实现。
type weightedRandomSelectorImpl struct {
	mu  sync.Mutex
	rng *rand.Rand
}

var (
	// RandomSelect 随机选择器（已弃用，请使用 NewRandomSelector）。
	// 保留用于向后兼容。
	RandomSelect Selector = NewRandomSelector()
	// RoundRobinSelect 轮询选择器（已弃用，请使用 NewRoundRobinSelector）。
	// 保留用于向后兼容。
	RoundRobinSelect Selector = NewRoundRobinSelector()
)

// NewRoundRobinSelector 创建轮询选择器。
func NewRoundRobinSelector() Selector {
	return &roundRobinSelectorImpl{}
}

// NewRandomSelector 创建随机选择器。
func NewRandomSelector() Selector {
	return &randomSelectorImpl{rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
}

// NewWeightedRandomSelector 创建加权随机选择器。
func NewWeightedRandomSelector() Selector {
	return &weightedRandomSelectorImpl{rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
}

func (s *roundRobinSelectorImpl) Select(instances []InstanceInfo) (InstanceInfo, error) {
	if len(instances) == 0 {
		return InstanceInfo{}, ErrNoInstances
	}
	idx := s.counter.Add(1) - 1
	return instances[idx%uint64(len(instances))], nil
}

func (s *randomSelectorImpl) Select(instances []InstanceInfo) (InstanceInfo, error) {
	if len(instances) == 0 {
		return InstanceInfo{}, ErrNoInstances
	}
	s.mu.Lock()
	idx := s.rng.IntN(len(instances))
	s.mu.Unlock()
	return instances[idx], nil
}

func (s *weightedRandomSelectorImpl) Select(instances []InstanceInfo) (InstanceInfo, error) {
	n := len(instances)
	if n == 0 {
		return InstanceInfo{}, ErrNoInstances
	}
	// 第一步：计算总权重，权重 <= 0 的按 1 计算
	total := 0
	for _, inst := range instances {
		if inst.Weight <= 0 {
			total++
			continue
		}
		total += inst.Weight
	}
	// 第二步：生成随机目标值
	s.mu.Lock()
	target := s.rng.IntN(total)
	s.mu.Unlock()
	// 第三步：遍历实例，累减权重直到找到目标实例
	for _, inst := range instances {
		w := inst.Weight
		if w <= 0 {
			w = 1
		}
		target -= w
		if target < 0 {
			return inst, nil
		}
	}
	// 兜底返回最后一个实例（正常情况下不会执行到此）
	return instances[n-1], nil
}
