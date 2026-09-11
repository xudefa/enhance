package health

import (
	"context"
	"runtime"
	"time"
)

// RuntimeHealthIndicator 运行时健康指标
//
// 提供 Go 运行时相关的健康指标，包括:
// - Goroutine 数量
// - GC 统计信息
// - 内存使用情况
type RuntimeHealthIndicator struct{}

// NewRuntimeHealthIndicator 创建运行时健康指标
func NewRuntimeHealthIndicator() *RuntimeHealthIndicator {
	return &RuntimeHealthIndicator{}
}

// Name 返回指标名称
func (r *RuntimeHealthIndicator) Name() string {
	return "runtime"
}

// Health 执行健康检查
func (r *RuntimeHealthIndicator) Health(ctx context.Context) Health {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return Health{
		Status: StatusUp,
		Details: map[string]any{
			"goroutines":    runtime.NumGoroutine(),
			"heap_alloc":    stats.HeapAlloc,
			"heap_sys":      stats.HeapSys,
			"heap_idle":     stats.HeapIdle,
			"heap_inuse":    stats.HeapInuse,
			"heap_released": stats.HeapReleased,
			"gc_pause_ns":   stats.PauseTotalNs,
			"gc_count":      stats.NumGC,
			"num_cpu":       runtime.NumCPU(),
			"go_version":    runtime.Version(),
		},
		Timestamp: time.Now(),
	}
}
