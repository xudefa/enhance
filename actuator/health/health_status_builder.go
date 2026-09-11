package health

import (
	"maps"
	"time"
)

// HealthBuilder 健康信息构建器
//
// 提供流式 API 来构建 Health 对象。
type HealthBuilder struct {
	health Health
}

// Up 创建健康状态为 UP 的构建器
func Up() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusUp,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// Down 创建健康状态为 DOWN 的构建器
func Down() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusDown,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// Degraded 创建健康状态为 DEGRADED 的构建器
func Degraded() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusDegraded,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// Outage 创建健康状态为 OUTAGE 的构建器
func Outage() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusOutage,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// WithDetail 添加详细信息
func (b *HealthBuilder) WithDetail(key string, value any) *HealthBuilder {
	b.health.Details[key] = value
	return b
}

// WithDetails 批量添加详细信息
func (b *HealthBuilder) WithDetails(details map[string]any) *HealthBuilder {
	maps.Copy(b.health.Details, details)
	return b
}

// WithError 设置错误信息
func (b *HealthBuilder) WithError(err error) *HealthBuilder {
	b.health.Error = err
	return b
}

// Build 构建 Health 对象
func (b *HealthBuilder) Build() Health {
	b.health.Timestamp = time.Now()
	return b.health
}
