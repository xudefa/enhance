// Package health 提供健康检查功能，用于 enhance 框架。
package health

import (
	"context"
	"time"
)

func (s Status) String() string {
	if name, ok := statusNames[s]; ok {
		return name
	}
	return "UNKNOWN"
}

var statusNames = map[Status]string{
	StatusUp:       "UP",
	StatusDown:     "DOWN",
	StatusDegraded: "DEGRADED",
	StatusOutage:   "OUTAGE",
	StatusUnknown:  "UNKNOWN",
}

// NewAggregator 创建聚合器
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// AddIndicator 添加健康指标
func (a *Aggregator) AddIndicator(indicator Indicator) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.indicators = append(a.indicators, indicator)
}

// Indicators 返回所有指标的副本，防止外部修改
func (a *Aggregator) Indicators() []Indicator {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]Indicator, len(a.indicators))
	copy(result, a.indicators)
	return result
}

// Aggregate 聚合所有指标的健康状态
func (a *Aggregator) Aggregate(ctx context.Context) Health {
	indicators := a.Indicators()

	overall := StatusUp
	details := make(map[string]any)

	for _, ind := range indicators {
		h := ind.Health(ctx)
		d := map[string]any{
			"status": h.Status.String(),
			"detail": h.Details,
		}
		if h.Error != nil {
			d["error"] = h.Error.Error()
		}
		details[ind.Name()] = d
		switch h.Status {
		case StatusOutage:
			overall = StatusOutage
		case StatusDown:
			if overall != StatusOutage {
				overall = StatusDown
			}
		case StatusDegraded:
			if overall != StatusOutage && overall != StatusDown {
				overall = StatusDegraded
			}
		}
	}

	return Health{
		Status:    overall,
		Details:   details,
		Timestamp: time.Now(),
	}
}
