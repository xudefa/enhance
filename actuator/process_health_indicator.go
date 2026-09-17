package actuator

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/xudefa/enhance/actuator/health"
)

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

	healthResult := health.Health{
		Details:   make(map[string]any),
		Timestamp: time.Now(),
	}

	healthResult.Details["goroutines"] = numGoroutines
	healthResult.Details["cpu_num"] = runtime.NumCPU()

	if numGoroutines > p.goroutineThreshold {
		healthResult.Status = health.StatusDegraded
		healthResult.Details["message"] = fmt.Sprintf("too many goroutines: %d, exceeds threshold: %d", numGoroutines, p.goroutineThreshold)
		return healthResult
	}
	healthResult.Status = health.StatusUp

	return healthResult
}
