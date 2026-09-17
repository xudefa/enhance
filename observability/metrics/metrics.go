package metrics

import (
	"sync"
	"time"
)

// ==================== Counter 实现 ====================

// counterImpl Counter 接口的默认实现。
type counterImpl struct {
	name  string
	value float64
	mu    sync.Mutex
}

// NewCounter 创建计数器。
func NewCounter(name string) Counter {
	return &counterImpl{name: name}
}

// Name 返回计数器的名称。
func (c *counterImpl) Name() string {
	return c.name
}

// Value 返回计数器的当前值。
func (c *counterImpl) Value() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Inc 将计数器加一。
func (c *counterImpl) Inc() {
	c.Add(1)
}

// Add 将计数器增加指定增量。
func (c *counterImpl) Add(delta float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += delta
}

// Reset 将计数器重置为零。
func (c *counterImpl) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
}

// ==================== Gauge 实现 ====================

// gaugeImpl Gauge 接口的默认实现。
type gaugeImpl struct {
	name  string
	value float64
	mu    sync.Mutex
}

// NewGauge 创建仪表盘。
func NewGauge(name string) Gauge {
	return &gaugeImpl{name: name}
}

// Name 返回仪表盘的名称。
func (g *gaugeImpl) Name() string {
	return g.name
}

// Value 返回仪表盘的当前值。
func (g *gaugeImpl) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Set 设置仪表盘的值。
func (g *gaugeImpl) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Add 将仪表盘增加指定增量。
func (g *gaugeImpl) Add(delta float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += delta
}

// ==================== Histogram 实现 ====================

// histogramImpl Histogram 接口的默认实现。
type histogramImpl struct {
	name  string
	count int64
	sum   float64
	min   float64
	max   float64
	mu    sync.Mutex
}

// NewHistogram 创建直方图。
func NewHistogram(name string) Histogram {
	return &histogramImpl{
		name: name,
		min:  1e99,
		max:  -1e99,
	}
}

// Name 返回直方图的名称。
func (h *histogramImpl) Name() string {
	return h.name
}

// Value 返回直方图的平均值，无观测时返回 0。
func (h *histogramImpl) Value() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.count == 0 {
		return 0
	}
	return h.sum / float64(h.count)
}

// Observe 记录一次观测值并更新求和与最值。
func (h *histogramImpl) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count++
	h.sum += value
	if value < h.min {
		h.min = value
	}
	if value > h.max {
		h.max = value
	}
}

// Count 返回观测次数。
func (h *histogramImpl) Count() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

// Sum 返回观测值之和。
func (h *histogramImpl) Sum() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sum
}

// Min 返回观测最小值。
func (h *histogramImpl) Min() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.min
}

// Max 返回观测最大值。
func (h *histogramImpl) Max() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.max
}

// ==================== Timer 实现 ====================

// Timer 计时器实现 Metric 接口。
type Timer struct {
	name      string
	startTime time.Time
	duration  time.Duration
	mu        sync.Mutex
}

// NewTimer 创建计时器。
func NewTimer(name string) *Timer {
	return &Timer{name: name}
}

// Name 返回计时器的名称。
func (t *Timer) Name() string {
	return t.name
}

// Value 返回计时器的毫秒数。
func (t *Timer) Value() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return float64(t.duration.Milliseconds())
}

// Start 开始计时。
func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.startTime = time.Now()
}

// Stop 停止计时。
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.startTime.IsZero() {
		t.duration = 0
		return
	}
	t.duration = time.Since(t.startTime)
}

// Duration 获取持续时间。
func (t *Timer) Duration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.duration
}

// ==================== MetricsRegistry 实现 ====================

// metricsRegistryImpl MetricsRegistry 接口的默认实现。
type metricsRegistryImpl struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewMetricsRegistry 创建指标注册表。
func NewMetricsRegistry() MetricsRegistry {
	return &metricsRegistryImpl{
		metrics: make(map[string]Metric),
	}
}

// Register 按名称注册指标到注册表。
func (r *metricsRegistryImpl) Register(metric Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[metric.Name()] = metric
}

// Get 按名称查找指标并返回其是否存在。
func (r *metricsRegistryImpl) Get(name string) (Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.metrics[name]
	return metric, exists
}

// List 返回注册表中的全部指标。
func (r *metricsRegistryImpl) List() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metrics := make([]Metric, 0, len(r.metrics))
	for _, m := range r.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}
