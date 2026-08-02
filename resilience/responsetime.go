package resilience

import (
	"math/rand"
	"sync"
	"time"
)

// backendResponseTime 后端响应时间记录
type backendResponseTime struct {
	avgTime     float64
	lastUpdated time.Time
}

// ResponseTimeWeighted 响应时间加权负载均衡器
// 使用 sync.Map 优化读多写少场景的并发性能
type ResponseTimeWeighted struct {
	avgResponseTimes sync.Map // map[string]*backendResponseTime
	decay            float64
}

// NewResponseTimeWeighted 创建响应时间加权负载均衡器
func NewResponseTimeWeighted(decay ...float64) *ResponseTimeWeighted {
	d := 0.9
	if len(decay) > 0 {
		d = decay[0]
	}

	return &ResponseTimeWeighted{
		decay: d,
	}
}

// Next 选择响应时间最短的后端
func (rtw *ResponseTimeWeighted) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	backends = nonNilBackends(backends)
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	type backendWeight struct {
		backend *ServiceInstance
		weight  float64
	}

	weighted := make([]backendWeight, 0, len(backends))
	totalWeight := 0.0

	for _, b := range backends {
		avgTime := 1.0
		if value, ok := rtw.avgResponseTimes.Load(b.URL); ok {
			v, _ := value.(*backendResponseTime)
			if v != nil {
				avgTime = v.avgTime
			}
			if avgTime <= 0 {
				avgTime = 1
			}
		}

		weight := 1.0 / avgTime
		weighted = append(weighted, backendWeight{backend: b, weight: weight})
		totalWeight += weight
	}

	if totalWeight <= 0 {
		return backends[0], nil
	}

	r := rand.Float64() * totalWeight
	accumulated := 0.0

	for _, bw := range weighted {
		accumulated += bw.weight
		if r <= accumulated {
			return bw.backend, nil
		}
	}

	return backends[len(backends)-1], nil
}

// RecordResponseTime 记录后端响应时间
func (rtw *ResponseTimeWeighted) RecordResponseTime(backendURL string, responseTimeMs float64) {
	const maxRetries = 100

	// 尝试更新已存在的记录
	for attempt := 0; attempt < maxRetries; attempt++ {
		if value, ok := rtw.avgResponseTimes.Load(backendURL); ok {
			old, _ := value.(*backendResponseTime)
			if old != nil {
				newAvg := rtw.decay*old.avgTime + (1-rtw.decay)*responseTimeMs
				newRecord := &backendResponseTime{
					avgTime:     newAvg,
					lastUpdated: time.Now(),
				}
				if rtw.avgResponseTimes.CompareAndSwap(backendURL, old, newRecord) {
					return
				}
				// CAS 失败，重试
				continue
			}
		}

		// 键不存在或值为 nil，尝试创建
		newRecord := &backendResponseTime{
			avgTime:     responseTimeMs,
			lastUpdated: time.Now(),
		}
		// 使用 LoadOrStore 避免竞态
		existing, loaded := rtw.avgResponseTimes.LoadOrStore(backendURL, newRecord)
		if !loaded {
			return
		}
		// 另一个 goroutine 已存储，CAS 更新
		_ = existing
	}

	// 重试次数耗尽，直接保存当前记录以保证收敛
	rtw.avgResponseTimes.Store(backendURL, &backendResponseTime{
		avgTime:     responseTimeMs,
		lastUpdated: time.Now(),
	})
}

// GetAvgResponseTime 获取后端平均响应时间
func (rtw *ResponseTimeWeighted) GetAvgResponseTime(backendURL string) (float64, bool) {
	if value, ok := rtw.avgResponseTimes.Load(backendURL); ok {
		v, _ := value.(*backendResponseTime)
		if v != nil {
			return v.avgTime, true
		}
	}
	return 0, false
}
