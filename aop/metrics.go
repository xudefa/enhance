package aop

import (
	"sync"
	"sync/atomic"
)

// AopMetrics AOP指标
//
// 收集AOP相关的性能指标
// 使用 atomic 操作优化高并发场景下的计数性能
type AopMetrics struct {
	mu                 sync.Mutex
	TotalProxies       atomic.Int64
	GeneratedProxies   atomic.Int64
	RuntimeProxies     atomic.Int64
	TotalAspects       atomic.Int64
	TotalInterceptions atomic.Int64
	// AverageLatency 使用 mu 保护，因为涉及浮点数计算
	averageLatency float64
}

// NewAopMetrics 创建AOP指标
func NewAopMetrics() *AopMetrics {
	return &AopMetrics{}
}

// RecordProxyCreated 记录代理创建
func (m *AopMetrics) RecordProxyCreated(isGenerated bool) {
	m.TotalProxies.Add(1)
	if isGenerated {
		m.GeneratedProxies.Add(1)
		return
	}
	m.RuntimeProxies.Add(1)
}

// RecordAspectRegistered 记录切面注册
func (m *AopMetrics) RecordAspectRegistered() {
	m.TotalAspects.Add(1)
}

// RecordInterception 记录拦截
func (m *AopMetrics) RecordInterception(latency float64) {
	count := m.TotalInterceptions.Add(1)

	m.mu.Lock()
	defer m.mu.Unlock()

	// 计算平均延迟
	if count == 1 {
		m.averageLatency = latency
		return
	}
	m.averageLatency = (m.averageLatency*float64(count-1) + latency) / float64(count)
}

// GetMetrics 获取指标
func (m *AopMetrics) GetMetrics() map[string]any {
	m.mu.Lock()
	avgLatency := m.averageLatency
	m.mu.Unlock()

	return map[string]any{
		"total_proxies":       m.TotalProxies.Load(),
		"generated_proxies":   m.GeneratedProxies.Load(),
		"runtime_proxies":     m.RuntimeProxies.Load(),
		"total_aspects":       m.TotalAspects.Load(),
		"total_interceptions": m.TotalInterceptions.Load(),
		"average_latency":     avgLatency,
	}
}

// Reset 重置指标
func (m *AopMetrics) Reset() {
	m.TotalProxies.Store(0)
	m.GeneratedProxies.Store(0)
	m.RuntimeProxies.Store(0)
	m.TotalAspects.Store(0)
	m.TotalInterceptions.Store(0)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.averageLatency = 0
}

// GlobalAopMetrics 全局AOP指标
var GlobalAopMetrics = NewAopMetrics()

// GetGlobalAopMetrics 获取全局AOP指标
func GetGlobalAopMetrics() map[string]any {
	return GlobalAopMetrics.GetMetrics()
}

// ResetGlobalAopMetrics 重置全局AOP指标
func ResetGlobalAopMetrics() {
	GlobalAopMetrics.Reset()
}
