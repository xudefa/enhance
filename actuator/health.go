package actuator

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"time"

	"github.com/xudefa/enhance/actuator/health"
)

// DiskSpaceHealthIndicator 磁盘空间健康指标
//
// 检查磁盘使用率是否超过阈值，当使用率过高时返回降级状态。
type DiskSpaceHealthIndicator struct {
	path      string
	threshold float64 // 阈值，0.0-1.0，超过此比例则为降级状态
}

// NewDiskSpaceHealthIndicator 创建磁盘空间健康指标
//
// 参数：
//   - path: 检查的磁盘路径
//   - threshold: 使用率阈值（0.0-1.0），超过此比例返回降级状态
func NewDiskSpaceHealthIndicator(path string, threshold float64) *DiskSpaceHealthIndicator {
	return &DiskSpaceHealthIndicator{
		path:      path,
		threshold: threshold,
	}
}

// Name 返回健康指标名称
func (d *DiskSpaceHealthIndicator) Name() string {
	return fmt.Sprintf("disk_space_%s", d.path)
}

// Health 执行磁盘空间健康检查（使用真实系统调用）。
func (d *DiskSpaceHealthIndicator) Health(ctx context.Context) health.Health {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(d.path, &stat); err != nil {
		return health.Health{
			Status:    health.StatusUnknown,
			Timestamp: time.Now(),
			Details: map[string]any{
				"path":    d.path,
				"error":   err.Error(),
				"message": "无法获取磁盘空间信息",
			},
		}
	}

	// stat.Bsize 在 macOS 上是 uint32，在 Linux 上是 int64
	var bsize uint64
	switch v := any(stat.Bsize).(type) {
	case uint32:
		bsize = uint64(v)
	case int64:
		bsize = uint64(v)
	default:
		bsize = uint64(stat.Bsize)
	}

	total := stat.Blocks * bsize
	free := stat.Bavail * bsize
	used := total - free

	usagePercent := float64(used) / float64(total)

	h := health.Health{
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	h.Details["path"] = d.path
	h.Details["total_bytes"] = total
	h.Details["used_bytes"] = used
	h.Details["free_bytes"] = total - used
	h.Details["usage_percent"] = fmt.Sprintf("%.2f%%", usagePercent*100)

	if usagePercent > d.threshold {
		h.Status = health.StatusDegraded
		h.Details["message"] = fmt.Sprintf("disk usage %.2f%% exceeds threshold %.2f%%", usagePercent*100, d.threshold*100)
		return h
	}
	h.Status = health.StatusUp

	return h
}

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

	h := health.Health{
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	h.Details["alloc_bytes"] = memStats.Alloc
	h.Details["sys_bytes"] = memStats.Sys
	h.Details["heap_alloc"] = memStats.HeapAlloc
	h.Details["heap_sys"] = memStats.HeapSys
	h.Details["heap_objects"] = memStats.HeapObjects

	// 防止除零错误
	if memStats.Sys == 0 {
		h.Details["heap_percent"] = "0.00%"
		h.Status = health.StatusUp
		return h
	}

	heapPercent := float64(memStats.Alloc) / float64(memStats.Sys)
	h.Details["heap_percent"] = fmt.Sprintf("%.2f%%", heapPercent*100)

	if heapPercent > m.threshold {
		h.Status = health.StatusDegraded
		h.Details["message"] = fmt.Sprintf("heap usage %.2f%% exceeds threshold %.2f%%", heapPercent*100, m.threshold*100)
		return h
	}
	h.Status = health.StatusUp

	return h
}

// ProcessHealthIndicator 进程健康指标
//
// 检查进程状态信息，如goroutine数量等。
type ProcessHealthIndicator struct {
	goroutineThreshold int // goroutine 数量阈值
}

// NewProcessHealthIndicator 创建进程健康指标
//
// 参数：
//   - goroutineThreshold: goroutine 数量阈值，超过此值返回降级状态
func NewProcessHealthIndicator(goroutineThreshold int) *ProcessHealthIndicator {
	return &ProcessHealthIndicator{
		goroutineThreshold: goroutineThreshold,
	}
}

// Name 返回健康指标名称
func (p *ProcessHealthIndicator) Name() string {
	return "process_status"
}

// Health 执行进程状态健康检查
func (p *ProcessHealthIndicator) Health(ctx context.Context) health.Health {
	numGoroutines := runtime.NumGoroutine()

	h := health.Health{
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	h.Details["goroutines"] = numGoroutines
	h.Details["cpu_num"] = runtime.NumCPU()

	if numGoroutines > p.goroutineThreshold {
		h.Status = health.StatusDegraded
		h.Details["message"] = fmt.Sprintf("too many goroutines: %d, exceeds threshold: %d", numGoroutines, p.goroutineThreshold)
		return h
	}
	h.Status = health.StatusUp

	return h
}
