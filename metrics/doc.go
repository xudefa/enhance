// Package metrics 提供指标收集和导出功能，用于 enhance 框架。
//
// 该模块提供统一的指标抽象接口，支持多种指标后端集成。
// 包含指标注册、收集、Prometheus 导出等监控支持。
//
// # 架构设计
//
//   - MeterRegistry: 指标注册表，管理所有指标
//   - Counter: 计数器指标，只增不减
//   - Gauge: 仪表盘指标，可增可减
//   - Timer: 计时器指标，记录耗时
//   - Histogram: 直方图指标，记录分布
//
// # 核心功能
//
//   - 指标注册: 支持注册多种类型的指标
//   - 指标收集: 自动收集和聚合指标数据
//   - 指标导出: 支持导出到 Prometheus 等监控系统
//   - 标签支持: 支持带标签的指标
//
// # 使用方式
//
// 创建指标注册表：
//
//	registry := metrics.NewSimpleRegistry()
//
// 注册计数器：
//
//	counter := registry.Counter("http.requests.total", "method", "GET")
//	counter.Inc()
//
// 注册计时器：
//
//	timer := registry.Timer("http.request.duration")
//	timer.Record(time.Duration)
//
// # 集成后端
//
// 具体实现位于 starter 子包：
//
//   - starter/prometheus: Prometheus 集成
//   - starter/otel: OpenTelemetry 集成
package metrics

// Counter 计数器接口
//
// 用于记录单调递增的数值，如请求次数、错误计数。
type Counter interface {
	// Inc 计数器加 1
	Inc()
	// Add 计数器增加指定值
	Add(v float64)
	// Value 返回当前计数值
	Value() float64
	// Reset 重置计数器为 0
	Reset()
}

// Gauge 仪表盘接口
//
// 用于记录可增可减的数值，如当前连接数、CPU 使用率。
type Gauge interface {
	// Set 设置当前值
	Set(v float64)
	// Add 增加指定值（可以为负数）
	Add(v float64)
	// Value 返回当前值
	Value() float64
}

// Histogram 直方图接口
//
// 用于记录分布情况，如请求延迟、响应大小等。
type Histogram interface {
	// Record 记录一个值
	Record(v float64)
	// RecordWithLabels 记录带标签的值
	RecordWithLabels(v float64, labels map[string]string)
	// Count 返回记录的样本数
	Count() int64
	// Sum 返回所有样本的总和
	Sum() float64
	// Reset 重置直方图
	Reset()
}

// Exporter 指标导出器接口
type Exporter interface {
	Export(metrics []Metric) error
}

// MeterRegistry 指标注册表接口
//
// 管理 Counter、Gauge 和 Histogram 的创建与收集，支持按名称获取或创建。
// 提供指标导出功能，可将指标数据导出到不同的监控系统。
type MeterRegistry interface {
	// Counter 获取或创建指定名称的计数器
	// name: 指标名称
	// tags: 标签对，格式为 key1, value1, key2, value2...
	Counter(name string, tags ...string) Counter

	// Gauge 获取或创建指定名称的仪表盘
	// name: 指标名称
	// tags: 标签对，格式为 key1, value1, key2, value2...
	Gauge(name string, tags ...string) Gauge

	// Histogram 获取或创建指定名称的直方图
	// name: 指标名称
	// tags: 标签对，格式为 key1, value1, key2, value2...
	Histogram(name string, tags ...string) Histogram

	// Collect 收集所有已注册的指标快照
	Collect() []Metric

	// RegisterExporter 注册指标导出器
	RegisterExporter(exporter Exporter)

	// Export 导出所有指标到已注册的导出器
	Export() error

	// Reset 重置所有指标为初始状态
	Reset()
}

// Metric 指标快照
//
// 包含指标名称、当前值和标签信息，用于采集和上报。
type Metric struct {
	Name      string            `json:"name"`      // 指标名称
	Value     float64           `json:"value"`     // 指标当前值
	Tags      map[string]string `json:"tags"`      // 指标标签
	Type      string            `json:"type"`      // 指标类型: counter/gauge/histogram
	Timestamp int64             `json:"timestamp"` // 时间戳
	Count     int64             `json:"count"`     // 样本数量（仅直方图）
	Sum       float64           `json:"sum"`       // 样本总和（仅直方图）
}

// ==================== 配置键常量 ====================

const (
	// Metrics 配置
	MetricsEnabled = "metrics.enabled"
)

// ==================== 条件值常量 ====================

const (
	ConditionTrue = "true"
)
