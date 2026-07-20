package resilience

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// BackendStats 后端统计信息
// 使用 atomic 操作优化计数器，float64 使用 mu 保护
type BackendStats struct {
	mu                sync.Mutex
	TotalRequests     atomic.Int64
	FailedRequests    atomic.Int64
	ActiveConnections atomic.Int64
	LastUpdated       atomic.Int64 // Unix 时间戳

	// avgResponseTime 使用 mu 保护
	avgResponseTime float64
}

// AdaptiveWeight 自适应权重负载均衡器
// 使用 sync.Map 优化读多写少场景的并发性能
type AdaptiveWeight struct {
	stats sync.Map // map[string]*BackendStats
}

// NewAdaptiveWeight 创建自适应权重负载均衡器
func NewAdaptiveWeight() *AdaptiveWeight {
	return &AdaptiveWeight{}
}

// Next 选择权重最高的后端
func (aw *AdaptiveWeight) Next(backends []*ServiceInstance) (*ServiceInstance, error) {
	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	type backendScore struct {
		backend *ServiceInstance
		score   float64
	}

	scores := make([]backendScore, 0, len(backends))
	totalScore := 0.0

	for _, b := range backends {
		if value, ok := aw.stats.Load(b.URL); ok {
			stats := value.(*BackendStats)
			score := aw.calculateScore(stats)
			scores = append(scores, backendScore{backend: b, score: score})
			totalScore += score
			continue
		}
		scores = append(scores, backendScore{backend: b, score: 1.0})
		totalScore += 1.0
	}

	if totalScore <= 0 {
		return backends[0], nil
	}

	r := rand.Float64() * totalScore
	accumulated := 0.0

	for _, bs := range scores {
		accumulated += bs.score
		if r <= accumulated {
			return bs.backend, nil
		}
	}

	return backends[len(backends)-1], nil
}

// calculateScore 计算后端综合得分
func (aw *AdaptiveWeight) calculateScore(stats *BackendStats) float64 {
	totalReqs := stats.TotalRequests.Load()
	failedReqs := stats.FailedRequests.Load()
	avgRespTime := stats.GetAvgResponseTime()
	activeConns := stats.ActiveConnections.Load()

	errorRate := 0.0
	if totalReqs > 0 {
		errorRate = float64(failedReqs) / float64(totalReqs)
	}

	responseTimeScore := 1.0
	if avgRespTime > 0 {
		responseTimeScore = 1000.0 / (avgRespTime + 1)
	}

	connectionScore := 1.0
	if activeConns > 0 {
		connectionScore = 100.0 / (float64(activeConns) + 1)
	}

	score := (1 - errorRate) * responseTimeScore * connectionScore

	return math.Max(score, 0.01)
}

// RecordRequest 记录请求统计
func (aw *AdaptiveWeight) RecordRequest(backendURL string, responseTimeMs float64, failed bool) {
	// 使用 LoadOrStore 获取或创建 stats
	value, _ := aw.stats.LoadOrStore(backendURL, &BackendStats{})
	stats := value.(*BackendStats)

	stats.TotalRequests.Add(1)
	if failed {
		stats.FailedRequests.Add(1)
	}

	// 使用 CAS 更新 AvgResponseTime
	stats.UpdateAvgResponseTime(responseTimeMs)
	stats.LastUpdated.Store(time.Now().UnixNano())
}

// RecordConnection 记录连接数变化
func (aw *AdaptiveWeight) RecordConnection(backendURL string, delta int64) {
	value, _ := aw.stats.LoadOrStore(backendURL, &BackendStats{})
	stats := value.(*BackendStats)

	newConns := stats.ActiveConnections.Add(delta)
	if newConns < 0 {
		stats.ActiveConnections.Store(0)
	}
}

// GetStats 获取后端统计信息
func (aw *AdaptiveWeight) GetStats(backendURL string) (*BackendStats, bool) {
	if value, ok := aw.stats.Load(backendURL); ok {
		return value.(*BackendStats), true
	}
	return nil, false
}

// GetAllStats 获取所有后端统计信息
func (aw *AdaptiveWeight) GetAllStats() map[string]*BackendStats {
	result := make(map[string]*BackendStats)
	aw.stats.Range(func(key, value any) bool {
		result[key.(string)] = value.(*BackendStats)
		return true
	})
	return result
}

// GetAvgResponseTime 获取平均响应时间
func (s *BackendStats) GetAvgResponseTime() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.avgResponseTime
}

// UpdateAvgResponseTime 更新平均响应时间
func (s *BackendStats) UpdateAvgResponseTime(responseTimeMs float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.avgResponseTime == 0 {
		s.avgResponseTime = responseTimeMs
		return
	}
	s.avgResponseTime = 0.9*s.avgResponseTime + 0.1*responseTimeMs
}

// SortByResponseTime 按响应时间排序后端
func SortByResponseTime(backends []*ServiceInstance, stats map[string]*BackendStats) []*ServiceInstance {
	sorted := make([]*ServiceInstance, len(backends))
	copy(sorted, backends)

	sort.Slice(sorted, func(i, j int) bool {
		statsI := stats[sorted[i].URL]
		statsJ := stats[sorted[j].URL]

		if statsI == nil {
			return true
		}
		if statsJ == nil {
			return false
		}

		return statsI.GetAvgResponseTime() < statsJ.GetAvgResponseTime()
	})

	return sorted
}
