package metrics

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// simpleCounter 简单计数器实现
//
// 使用 atomic.Uint64 存储 float64 位模式，避免锁竞争。
// 适用于记录单调递增的数值，如请求次数、错误计数等。
type simpleCounter struct {
	value atomic.Uint64
}

// NewSimpleCounter 创建新的简单计数器
func NewSimpleCounter() Counter {
	return &simpleCounter{}
}

// Inc 计数器加 1
func (c *simpleCounter) Inc() {
	c.Add(1)
}

// Add 计数器增加指定值
func (c *simpleCounter) Add(v float64) {
	for {
		oldBits := c.value.Load()
		old := math.Float64frombits(oldBits)
		newBits := math.Float64bits(old + v)
		if c.value.CompareAndSwap(oldBits, newBits) {
			break
		}
	}
}

// Value 返回当前计数值
func (c *simpleCounter) Value() float64 {
	return math.Float64frombits(c.value.Load())
}

// Reset 重置计数器为 0
func (c *simpleCounter) Reset() {
	c.value.Store(math.Float64bits(0))
}

// simpleGauge 简单仪表盘实现
//
// 使用 atomic.Uint64 存储 float64 位模式，支持并发读取。
type simpleGauge struct {
	value atomic.Uint64
}

// NewSimpleGauge 创建新的简单仪表盘
func NewSimpleGauge() Gauge {
	return &simpleGauge{}
}

// Set 设置当前值
func (g *simpleGauge) Set(v float64) {
	g.value.Store(math.Float64bits(v))
}

// Add 增加指定值（可以为负数）
func (g *simpleGauge) Add(v float64) {
	for {
		oldBits := g.value.Load()
		old := math.Float64frombits(oldBits)
		newBits := math.Float64bits(old + v)
		if g.value.CompareAndSwap(oldBits, newBits) {
			break
		}
	}
}

// Value 返回当前值
func (g *simpleGauge) Value() float64 {
	return math.Float64frombits(g.value.Load())
}

// simpleHistogram 简单直方图实现
//
// 使用 sync.Mutex 保证并发安全，支持基本的统计功能。
type simpleHistogram struct {
	mu    sync.Mutex
	name  string
	tags  map[string]string
	count int64
	sum   float64
	min   float64
	max   float64
}

// NewSimpleHistogram 创建新的简单直方图
func NewSimpleHistogram(name string, tags map[string]string) Histogram {
	return &simpleHistogram{
		name: name,
		tags: copyTags(tags),
		min:  math.MaxFloat64,
		max:  math.Inf(-1),
	}
}

// Record 记录一个值
func (h *simpleHistogram) Record(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count++
	h.sum += v
	if v < h.min {
		h.min = v
	}
	if v > h.max {
		h.max = v
	}
}

// RecordWithLabels 记录带标签的值
func (h *simpleHistogram) RecordWithLabels(v float64, labels map[string]string) {
	if len(labels) > 0 {
		h.mu.Lock()
		if h.tags == nil {
			h.tags = make(map[string]string)
		}
		// 合并标签，避免覆盖已有标签
		for k, val := range labels {
			h.tags[k] = val
		}
		h.mu.Unlock()
	}
	h.Record(v)
}

// tagsSnapshot 返回标签的快照副本，调用方持锁，避免与 RecordWithLabels 竞争。
func (h *simpleHistogram) tagsSnapshot() map[string]string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return copyTags(h.tags)
}

// Count 返回记录的样本数
func (h *simpleHistogram) Count() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

// Sum 返回所有样本的总和
func (h *simpleHistogram) Sum() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sum
}

// Reset 重置直方图
func (h *simpleHistogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count = 0
	h.sum = 0
	h.min = math.MaxFloat64
	h.max = math.Inf(-1)
}

// simpleRegistry 简单指标注册表实现
//
// 使用 sync.Map 优化并发访问，避免全局锁竞争。
// 使用 name+tags 组合作为唯一键，支持同名不同标签的指标实例。
type simpleRegistry struct {
	counters   sync.Map // map[string]*simpleCounter
	gauges     sync.Map // map[string]*simpleGauge
	histograms sync.Map // map[string]*simpleHistogram
	tags       sync.Map // map[string]map[string]string
	exporters  []Exporter
	exportMu   sync.RWMutex // 保护 exporters 切片
}

// NewSimpleRegistry 创建新的简单指标注册表
func NewSimpleRegistry() MeterRegistry {
	return &simpleRegistry{
		exporters: make([]Exporter, 0),
	}
}

func parseTags(tags []string) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	if len(tags)%2 != 0 {
		panic("metrics: tags must be provided as key/value pairs, got odd count " + strconv.Itoa(len(tags)))
	}
	result := make(map[string]string, len(tags)/2)
	for i := 0; i < len(tags); i += 2 {
		result[tags[i]] = tags[i+1]
	}
	return result
}

// metricKey 生成指标唯一键，由名称和标签组成
//
// 通过对标签键进行排序，确保相同标签组合总是生成相同的键，
// 避免 map 迭代顺序不确定导致的键冲突。
// 名称、标签键和标签值中的分隔符 | 会被转义，避免名称包含 | 时与其他键冲突。
func metricKey(name string, tags map[string]string) string {
	if len(tags) == 0 {
		return escapeMetricPart(name)
	}

	// 排序标签键以确保确定性
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(escapeMetricPart(name))
	for _, k := range keys {
		sb.WriteString("|")
		sb.WriteString(escapeMetricPart(k))
		sb.WriteString("=")
		sb.WriteString(escapeMetricPart(tags[k]))
	}
	return sb.String()
}

// escapeMetricPart 转义指标键组成部分中的分隔符和转义符本身。
func escapeMetricPart(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `|`, `\|`)
	return s
}

// unescapeMetricName 还原被 escapeMetricPart 转义过的指标名称。
func unescapeMetricName(s string) string {
	s = strings.ReplaceAll(s, `\|`, `|`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}

// Counter 获取或创建指定名称的计数器
//
// 使用 name+tags 组合作为唯一键，同名不同标签会创建不同的计数器实例。
func (r *simpleRegistry) Counter(name string, tags ...string) Counter {
	parsedTags := parseTags(tags)
	key := metricKey(name, parsedTags)
	if c, ok := r.counters.Load(key); ok {
		return c.(*simpleCounter)
	}
	c := &simpleCounter{}
	if actual, loaded := r.counters.LoadOrStore(key, c); loaded {
		return actual.(*simpleCounter)
	}
	if len(tags) > 0 {
		r.tags.Store(key, parsedTags)
	}
	return c
}

// Gauge 获取或创建指定名称的仪表盘
//
// 使用 name+tags 组合作为唯一键，同名不同标签会创建不同的仪表盘实例。
func (r *simpleRegistry) Gauge(name string, tags ...string) Gauge {
	parsedTags := parseTags(tags)
	key := metricKey(name, parsedTags)
	if g, ok := r.gauges.Load(key); ok {
		return g.(*simpleGauge)
	}
	g := &simpleGauge{}
	if actual, loaded := r.gauges.LoadOrStore(key, g); loaded {
		return actual.(*simpleGauge)
	}
	if len(tags) > 0 {
		r.tags.Store(key, parsedTags)
	}
	return g
}

// Histogram 获取或创建指定名称的直方图
//
// 使用 name+tags 组合作为唯一键，同名不同标签会创建不同的直方图实例。
func (r *simpleRegistry) Histogram(name string, tags ...string) Histogram {
	parsedTags := parseTags(tags)
	key := metricKey(name, parsedTags)
	if h, ok := r.histograms.Load(key); ok {
		return h.(*simpleHistogram)
	}
	h := NewSimpleHistogram(name, parsedTags)
	if actual, loaded := r.histograms.LoadOrStore(key, h); loaded {
		return actual.(*simpleHistogram)
	}
	if len(tags) > 0 {
		r.tags.Store(key, parsedTags)
	}
	return h
}

// metricNameFromKey 从 metricKey 中提取纯指标名称
func metricNameFromKey(key string) string {
	name := key
	if idx := findUnescaped(key, "|"); idx >= 0 {
		name = key[:idx]
	}
	return unescapeMetricName(name)
}

// findUnescaped 返回 s 中第一个未被转义符 \ 转义的 sep 的下标，未找到返回 -1。
func findUnescaped(s, sep string) int {
	escaped := false
	for i := 0; i < len(s); i++ {
		if escaped {
			escaped = false
			continue
		}
		if s[i] == '\\' {
			escaped = true
			continue
		}
		if strings.HasPrefix(s[i:], sep) {
			return i
		}
	}
	return -1
}

// Collect 收集所有已注册的指标快照
//
// 返回当前所有计数器、仪表盘和直方图的快照列表。
func (r *simpleRegistry) Collect() []Metric {
	metrics := make([]Metric, 0)
	now := time.Now().UnixMilli()

	r.counters.Range(func(key, value any) bool {
		k := key.(string)
		c := value.(*simpleCounter)
		m := Metric{
			Name:      metricNameFromKey(k),
			Value:     c.Value(),
			Type:      "counter",
			Timestamp: now,
		}
		if tags, ok := r.tags.Load(k); ok {
			m.Tags = copyTags(tags.(map[string]string))
		}
		metrics = append(metrics, m)
		return true
	})

	r.gauges.Range(func(key, value any) bool {
		k := key.(string)
		g := value.(*simpleGauge)
		m := Metric{
			Name:      metricNameFromKey(k),
			Value:     g.Value(),
			Type:      "gauge",
			Timestamp: now,
		}
		if tags, ok := r.tags.Load(k); ok {
			m.Tags = copyTags(tags.(map[string]string))
		}
		metrics = append(metrics, m)
		return true
	})

	r.histograms.Range(func(key, value any) bool {
		k := key.(string)
		h := value.(*simpleHistogram)
		count := h.Count()
		var avg float64
		if count > 0 {
			avg = h.Sum() / float64(count)
		}
		m := Metric{
			Name:      metricNameFromKey(k),
			Value:     avg,
			Type:      "histogram",
			Timestamp: now,
			Count:     count,
			Sum:       h.Sum(),
		}
		if tags := h.tagsSnapshot(); len(tags) > 0 {
			m.Tags = tags
		}
		metrics = append(metrics, m)
		return true
	})

	return metrics
}

// copyTags 复制标签 map，避免调用方与内部 map 共享引用导致数据竞争。
func copyTags(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// RegisterExporter 注册指标导出器
func (r *simpleRegistry) RegisterExporter(exporter Exporter) {
	r.exportMu.Lock()
	defer r.exportMu.Unlock()
	r.exporters = append(r.exporters, exporter)
}

// Export 导出所有指标到已注册的导出器
func (r *simpleRegistry) Export() error {
	metrics := r.Collect()
	r.exportMu.RLock()
	exporters := r.exporters
	r.exportMu.RUnlock()
	for _, exporter := range exporters {
		if err := exporter.Export(metrics); err != nil {
			return err
		}
	}
	return nil
}

// Reset 重置所有指标
func (r *simpleRegistry) Reset() {
	r.counters.Range(func(key, value any) bool {
		value.(*simpleCounter).Reset()
		return true
	})
	r.histograms.Range(func(key, value any) bool {
		value.(*simpleHistogram).Reset()
		return true
	})
	r.gauges.Range(func(key, value any) bool {
		value.(*simpleGauge).Set(0)
		return true
	})
}

// ConsoleExporter 控制台导出器
type ConsoleExporter struct{}

func NewConsoleExporter() Exporter {
	return &ConsoleExporter{}
}

func (e *ConsoleExporter) Export(metrics []Metric) error {
	return nil
}
