package actuator

import (
	"context"
	"fmt"
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
	switch bsizeVal := any(stat.Bsize).(type) {
	case uint32:
		bsize = uint64(bsizeVal)
	case int64:
		bsize = uint64(bsizeVal)
	default:
		bsize = uint64(stat.Bsize)
	}

	total := stat.Blocks * bsize
	free := stat.Bavail * bsize
	used := total - free

	usagePercent := float64(used) / float64(total)

	return d.buildHealthResult(used, total, free, usagePercent)
}

// buildHealthResult 构建磁盘健康检查结果。
func (d *DiskSpaceHealthIndicator) buildHealthResult(used, total, free uint64, usagePercent float64) health.Health {
	healthResult := health.Health{
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	healthResult.Details["path"] = d.path
	healthResult.Details["total_bytes"] = total
	healthResult.Details["used_bytes"] = used
	healthResult.Details["free_bytes"] = total - used
	healthResult.Details["usage_percent"] = fmt.Sprintf("%.2f%%", usagePercent*100)

	if usagePercent > d.threshold {
		healthResult.Status = health.StatusDegraded
		healthResult.Details["message"] = fmt.Sprintf("disk usage %.2f%% exceeds threshold %.2f%%", usagePercent*100, d.threshold*100)
		return healthResult
	}
	healthResult.Status = health.StatusUp

	return healthResult
}
