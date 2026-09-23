// Package health 提供健康检查功能，用于 enhance 框架。
package health

import (
	"context"
	"fmt"
	"time"
)

// defaultMaxConcurrentChecks 同时运行的指标健康检查数量上限。
const defaultMaxConcurrentChecks = 8

// DefaultIndicatorTimeout 每个指标的默认超时时间。
const DefaultIndicatorTimeout = 5 * time.Second

type healthCheckResult struct {
	health Health
	done   chan struct{}
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

// String 返回状态的字符串表示
func (s Status) String() string {
	if name, ok := statusNames[s]; ok {
		return name
	}
	return "UNKNOWN"
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
	indicatorList := make([]Indicator, len(a.indicators))
	copy(indicatorList, a.indicators)
	return indicatorList
}

// Aggregate 聚合所有指标的健康状态。
// 每个指标调用都有独立的超时保护，防止慢指标阻塞整个健康检查。
func (a *Aggregator) Aggregate(ctx context.Context) Health {
	indicators := a.Indicators()

	overall := StatusUp
	details := make(map[string]any)

	for _, ind := range indicators {
		indicatorHealth := a.aggregateWithTimeout(ctx, ind, DefaultIndicatorTimeout)
		detail := map[string]any{
			"status": indicatorHealth.Status.String(),
			"detail": indicatorHealth.Details,
		}
		if indicatorHealth.Error != nil {
			detail["error"] = indicatorHealth.Error.Error()
		}
		details[ind.Name()] = detail
		switch indicatorHealth.Status {
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

	checkResult := healthCheckResult{done: make(chan struct{}, 1)}
	go a.runHealthCheck(ind, ctx, &checkResult)

	select {
	case <-checkResult.done:
		return checkResult.health
	case <-ctx.Done():
		return Health{
			Status:  StatusDown,
			Error:   fmt.Errorf("health check for %s timed out after %v", ind.Name(), timeout),
			Details: map[string]any{"error": "timeout"},
		}
	}
}

// runHealthCheck 在独立 goroutine 中执行健康检查（信号量保护）。
func (a *Aggregator) runHealthCheck(ind Indicator, ctx context.Context, r *healthCheckResult) {
	defer func() { <-a.workers }()
	defer func() {
		if panicValue := recover(); panicValue != nil {
			r.health = Health{Status: StatusDown, Error: fmt.Errorf("panic: %v", panicValue)}
		}
		close(r.done)
	}()
	r.health = ind.Health(ctx)
}
