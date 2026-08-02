// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"reflect"
	"sync"
)

// ==================== AopManager 结构体 ====================

// AopManager AOP 管理器。
type AopManager struct {
	mu      sync.RWMutex
	config  *AopConfig
	aspects []*AspectMeta
}

// DefaultAopConfig 创建默认AOP配置。
func DefaultAopConfig() *AopConfig {
	return &AopConfig{
		Mode:        AopModeMixed,
		Weaver:      NewWeaver(),
		EnableCache: true,
	}
}

// GlobalAopManager 全局AOP管理器。
var GlobalAopManager = &AopManager{
	config:  DefaultAopConfig(),
	aspects: make([]*AspectMeta, 0),
}

// RegisterAspect 注册切面
func (m *AopManager) RegisterAspect(aspect *AspectMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aspects = append(m.aspects, aspect)
}

// RegisterAspects 批量注册切面
func (m *AopManager) RegisterAspects(aspects ...*AspectMeta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.aspects = append(m.aspects, aspects...)
}

// GetAspects 获取所有切面
func (m *AopManager) GetAspects() []*AspectMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// 返回副本，避免调用者修改内部数据
	dst := make([]*AspectMeta, len(m.aspects))
	copy(dst, m.aspects)
	return dst
}

// GetConfig 返回当前配置的快照副本（线程安全）。
func (m *AopManager) GetConfig() *AopConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return nil
	}
	cfg := *m.config
	return &cfg
}

// MatchAspectsForType 匹配指定类型的切面
func (m *AopManager) MatchAspectsForType(beanType reflect.Type) []*AspectMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*AspectMeta
	if beanType == nil {
		return matched
	}
	for _, aspect := range m.aspects {
		if aspect != nil && aspect.PointCut != nil && aspect.PointCut.MatchClass(beanType) {
			matched = append(matched, aspect)
		}
	}
	return matched
}
