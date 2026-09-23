package health

import (
	"context"
	"os"
	"runtime"
	"time"
)

// SystemHealthIndicator 系统健康指标
//
// 提供系统级别的指标，包括:
// - CPU 使用率
// - 内存使用率
// - 磁盘使用率
type SystemHealthIndicator struct{}

// NewSystemHealthIndicator 创建系统健康指标
func NewSystemHealthIndicator() *SystemHealthIndicator {
	return &SystemHealthIndicator{}
}

// Name 返回指标名称
func (s *SystemHealthIndicator) Name() string {
	return "system"
}

// Health 执行健康检查
func (s *SystemHealthIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusUp,
		Details: map[string]any{
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
			"num_cpu":  runtime.NumCPU(),
			"hostname": getHostname(),
		},
		Timestamp: time.Now(),
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
