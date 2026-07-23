// Package resilience 提供弹性容错功能，用于 enhance 框架。
package resilience

import (
	"context"
	"sync"
)

// InMemoryRegistry 内存中的服务注册中心实现。
// 适用于测试和开发环境，不支持持久化。
type InMemoryRegistry struct {
	mu        sync.RWMutex
	instances map[string][]InstanceInfo
}

// NewInMemoryRegistry 创建内存注册中心。
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		instances: make(map[string][]InstanceInfo),
	}
}

// Register 注册服务实例。
func (r *InMemoryRegistry) Register(ctx context.Context, info InstanceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	serviceName := info.ServiceName
	r.instances[serviceName] = append(r.instances[serviceName], info)
	return nil
}

// Deregister 注销服务实例。
func (r *InMemoryRegistry) Deregister(ctx context.Context, info InstanceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	serviceName := info.ServiceName
	instances, ok := r.instances[serviceName]
	if !ok {
		return nil
	}

	// 移除匹配的实例
	for i, inst := range instances {
		if inst.ID == info.ID {
			r.instances[serviceName] = append(instances[:i], instances[i+1:]...)
			break
		}
	}

	return nil
}

// Discover 发现服务实例。
func (r *InMemoryRegistry) Discover(ctx context.Context, serviceName string) ([]InstanceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances, ok := r.instances[serviceName]
	if !ok {
		return []InstanceInfo{}, nil
	}

	// 返回副本，避免外部修改
	result := make([]InstanceInfo, len(instances))
	copy(result, instances)
	return result, nil
}

// Watch 监听服务实例变更。
func (r *InMemoryRegistry) Watch(ctx context.Context, serviceName string) (<-chan []InstanceInfo, error) {
	ch := make(chan []InstanceInfo, 1)
	// 发送当前实例列表
	instances, _ := r.Discover(ctx, serviceName)
	ch <- instances
	return ch, nil
}
