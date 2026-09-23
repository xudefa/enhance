package health

import (
	"context"
	"time"
)

// IndicatorFunc 健康指标函数类型
//
// 简化版健康指标定义，只需提供一个函数即可注册健康检查。
type IndicatorFunc func(ctx context.Context) Health

// SimpleIndicator 简单健康指标实现
//
// 基于 IndicatorFunc 的健康指标实现。
type SimpleIndicator struct {
	name string
	fn   IndicatorFunc
}

// NewSimpleIndicator 创建简单健康指标
func NewSimpleIndicator(name string, fn IndicatorFunc) *SimpleIndicator {
	return &SimpleIndicator{
		name: name,
		fn:   fn,
	}
}

// Name 返回指标名称
func (s *SimpleIndicator) Name() string {
	return s.name
}

// Health 执行健康检查
func (s *SimpleIndicator) Health(ctx context.Context) Health {
	if s.fn == nil {
		return Health{
			Status:    StatusUnknown,
			Timestamp: time.Now(),
		}
	}
	return s.fn(ctx)
}
