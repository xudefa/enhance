// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"reflect"
)

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
	m.aspects = append(m.aspects, aspect)
}

// RegisterAspects 批量注册切面
func (m *AopManager) RegisterAspects(aspects ...*AspectMeta) {
	m.aspects = append(m.aspects, aspects...)
}

// GetAspects 获取所有切面
func (m *AopManager) GetAspects() []*AspectMeta {
	return m.aspects
}

// MatchAspectsForType 匹配指定类型的切面
func (m *AopManager) MatchAspectsForType(beanType any) []*AspectMeta {
	var matched []*AspectMeta
	t := reflect.TypeOf(beanType)
	if t == nil {
		return matched
	}
	for _, aspect := range m.aspects {
		if aspect.PointCut.MatchClass(t) {
			matched = append(matched, aspect)
		}
	}
	return matched
}
