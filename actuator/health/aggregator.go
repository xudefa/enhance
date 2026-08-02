// Package health 提供健康检查功能，用于 enhance 框架。
package health

import (
	"context"
	"fmt"
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
	return &Aggregator{
		workers: make(chan struct{}, defaultMaxConcurrentChecks),
	}
}

// defaultMaxConcurrentChecks 同时运行的指标健康检查数量上限。
const defaultMaxConcurrentChecks = 8

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

// DefaultIndicatorTimeout 每个指标的默认超时时间。
const DefaultIndicatorTimeout = 5 * time.Second

// Aggregate 聚合所有指标的健康状态。
// 每个指标调用都有独立的超时保护，防止慢指标阻塞整个健康检查。
func (a *Aggregator) Aggregate(ctx context.Context) Health {
	indicators := a.Indicators()

	overall := StatusUp
	details := make(map[string]any)

	for _, ind := range indicators {
		h := a.aggregateWithTimeout(ctx, ind, DefaultIndicatorTimeout)
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

// aggregateWithTimeout 在独立 goroutine 中执行指标健康检查，带超时保护。
//
// 通过信号量限制同时运行的健康检查数量，即使某个指标忽略 context 卡死不返回，
// 泄漏的 goroutine 数量也被限制在 defaultMaxConcurrentChecks 以内。
func (a *Aggregator) aggregateWithTimeout(ctx context.Context, ind Indicator, timeout time.Duration) Health {
	type result struct {
		health Health
		done   chan struct{}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case a.workers <- struct{}{}:
	case <-ctx.Done():
		return Health{
			Status:  StatusDown,
			Error:   fmt.Errorf("health check for %s timed out after %v", ind.Name(), timeout),
			Details: map[string]any{"error": "timeout"},
		}
	}

	r := result{done: make(chan struct{}, 1)}
	go func() {
		defer func() { <-a.workers }()
		defer func() {
			if p := recover(); p != nil {
				r.health = Health{Status: StatusDown, Error: fmt.Errorf("panic: %v", p)}
			}
			close(r.done)
		}()
		r.health = ind.Health(ctx)
	}()

	select {
	case <-r.done:
		return r.health
	case <-ctx.Done():
		return Health{
			Status:  StatusDown,
			Error:   fmt.Errorf("health check for %s timed out after %v", ind.Name(), timeout),
			Details: map[string]any{"error": "timeout"},
		}
	}
}
