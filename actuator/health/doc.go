// Package health 提供健康检查功能，用于 enhance 框架。
//
// 该模块提供组件健康状态检查和聚合功能。
package health

import (
	"context"
	"sync"
	"time"
)

// Status 健康状态枚举。
type Status int

const (
	// StatusUp 表示组件运行正常，服务完全可用。
	StatusUp Status = iota
	// StatusDown 表示组件不可用，服务中断。
	StatusDown
	// StatusDegraded 表示组件部分功能可用，服务质量下降。
	StatusDegraded
	// StatusOutage 表示组件完全不可用，服务停服。
	StatusOutage
	// StatusUnknown 表示无法确定组件状态。
	StatusUnknown
)

// Health 健康信息。
type Health struct {
	Status    Status         // 组件的健康状态
	Details   map[string]any // 组件健康检查的详细信息
	Error     error          // 检查过程中发生的错误信息
	Timestamp time.Time      // 健康检查完成的时间戳
}

// Indicator 健康指标接口。
type Indicator interface {
	Name() string
	Health(ctx context.Context) Health
}

// Aggregator 健康指标聚合器。
type Aggregator struct {
	mu         sync.RWMutex
	indicators []Indicator
}
