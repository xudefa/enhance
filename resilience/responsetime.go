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
			avgTime = value.(*backendResponseTime).avgTime
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
	// 尝试更新已存在的记录
	for {
		if value, ok := rtw.avgResponseTimes.Load(backendURL); ok {
			old := value.(*backendResponseTime)
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

		// 键不存在，尝试创建
		newRecord := &backendResponseTime{
			avgTime:     responseTimeMs,
			lastUpdated: time.Now(),
		}
		// 使用 LoadOrStore 避免竞态
		_, swapped := rtw.avgResponseTimes.LoadOrStore(backendURL, newRecord)
		if swapped {
			return // 成功插入
		}
		// 另一个 goroutine 已经插入，现在尝试更新它
		// 继续循环进入上面的更新分支
	}
}

// GetAvgResponseTime 获取后端平均响应时间
func (rtw *ResponseTimeWeighted) GetAvgResponseTime(backendURL string) (float64, bool) {
	if value, ok := rtw.avgResponseTimes.Load(backendURL); ok {
		return value.(*backendResponseTime).avgTime, true
	}
	return 0, false
}
