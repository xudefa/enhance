package exception

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// DefaultMetricsRecorder 默认指标记录器实现
//
// DefaultMetricsRecorder 是 MetricsRecorder 接口的默认实现。
// 它使用原子操作来记录异常计数，适合开发和测试环境。
// 使用 sync.Map + atomic.Int64 实现无锁高并发计数。
type DefaultMetricsRecorder struct {
	counters sync.Map // map[string]*atomic.Int64
}

// NewDefaultMetricsRecorder 创建默认指标记录器
//
// 返回一个 DefaultMetricsRecorder 实例。
func NewDefaultMetricsRecorder() MetricsRecorder {
	return &DefaultMetricsRecorder{}
}

// RecordException 记录异常指标
//
// 根据异常类型和状态码生成键，并增加对应的计数。
// 键的格式为 "exceptionType:statusCode"。
func (m *DefaultMetricsRecorder) RecordException(exceptionType string, statusCode int) {
	key := fmt.Sprintf("%s:%d", exceptionType, statusCode)
	v, _ := m.counters.LoadOrStore(key, &atomic.Int64{})
	ctr, _ := v.(*atomic.Int64)
	ctr.Add(1)
}

// GetCount 获取计数（用于测试）
//
// 返回指定异常类型和状态码的计数。
// 注意：这个方法主要用于测试，生产环境应该使用更专业的指标系统。
func (m *DefaultMetricsRecorder) GetCount(exceptionType string, statusCode int) int {
	key := fmt.Sprintf("%s:%d", exceptionType, statusCode)
	v, ok := m.counters.Load(key)
	if !ok {
		return 0
	}
	ctr, _ := v.(*atomic.Int64)
	return int(ctr.Load())
}
