package actuator

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/xudefa/enhance/actuator/health"
)

// MemoryHealthIndicator 内存使用健康指标
//
// 检查内存使用情况，当使用率过高时返回降级状态。
type MemoryHealthIndicator struct {
	threshold float64 // 阈值，0.0-1.0，超过此比例则为降级状态
}

// NewMemoryHealthIndicator 创建内存健康指标
//
// 参数：
//   - threshold: 堆内存使用率阈值（0.0-1.0），超过此比例返回降级状态
func NewMemoryHealthIndicator(threshold float64) *MemoryHealthIndicator {
	return &MemoryHealthIndicator{
		threshold: threshold,
	}
}

// Name 返回健康指标名称
func (m *MemoryHealthIndicator) Name() string {
	return "memory_usage"
}

// Health 执行内存使用健康检查
func (m *MemoryHealthIndicator) Health(ctx context.Context) health.Health {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	healthResult := health.Health{
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	healthResult.Details["alloc_bytes"] = memStats.Alloc
	healthResult.Details["sys_bytes"] = memStats.Sys
	healthResult.Details["heap_alloc"] = memStats.HeapAlloc
	healthResult.Details["heap_sys"] = memStats.HeapSys
	healthResult.Details["heap_objects"] = memStats.HeapObjects

	// 防止除零错误
	if memStats.Sys == 0 {
		healthResult.Details["heap_percent"] = "0.00%"
		healthResult.Status = health.StatusUp
		return healthResult
	}

	heapPercent := float64(memStats.Alloc) / float64(memStats.Sys)
	healthResult.Details["heap_percent"] = fmt.Sprintf("%.2f%%", heapPercent*100)

	if heapPercent > m.threshold {
		healthResult.Status = health.StatusDegraded
		healthResult.Details["message"] = fmt.Sprintf("heap usage %.2f%% exceeds threshold %.2f%%", heapPercent*100, m.threshold*100)
		return healthResult
	}
	healthResult.Status = health.StatusUp

	return healthResult
}
