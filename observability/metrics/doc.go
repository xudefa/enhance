// Package metrics 提供可观测性指标功能，用于 enhance 框架。
//
// 该模块集成指标收集到可观测性框架，支持自定义指标和指标导出。
// 参考 OpenTelemetry Metrics 的设计理念。
//
// # 架构设计
//
//   - MeterProvider: 指标提供者接口，管理指标器实例
//   - Meter: 指标器接口，创建和管理指标
//   - Counter: 计数器指标接口，只增不减
//   - Gauge: 仪表盘指标接口，可增可减
//   - Histogram: 直方图指标接口，记录分布
//   - MetricsExporter: 指标导出器接口，将指标发送到后端
//
// # 核心功能
//
//   - 指标收集: 支持多种类型的指标收集
//   - 指标导出: 支持导出到 Prometheus、OTLP 等后端
//   - 指标标签: 支持带标签的指标，实现多维度分析
//   - 指标聚合: 支持自动聚合和计算指标
//   - 指标采样: 支持采样策略减少指标量
//
// # 使用方式
//
// 创建指标器：
//
//	provider := metrics.NewMeterProvider()
//	meter := provider.Meter("my-service")
//
// 创建计数器：
//
//	counter := meter.Int64Counter("http.requests.total",
//	    metrics.WithDescription("Total HTTP requests"),
//	    metrics.WithUnit("{request}"),
//	)
//
// 记录指标：
//
//	counter.Add(ctx, 1,
//	    metrics.WithAttributes(attribute.String("http.method", "GET")),
//	    metrics.WithAttributes(attribute.String("http.status", "200")),
//	)
//
// # 指标类型
//
//   - Counter: 计数器，只增不减（如请求总数）
//   - Gauge: 仪表盘，可增可减（如当前连接数）
//   - Histogram: 直方图，记录分布（如请求延迟）
//   - Summary: 摘要，记录分位数（如 P99 延迟）
//
// # 集成后端
//
// 具体实现位于 starter/ 子包：
//
//   - starter/prometheus: Prometheus 集成
//   - starter/otel: OpenTelemetry 集成
package metrics

// Metric 指标接口。
type Metric interface {
	// Name 返回指标名称。
	Name() string

	// Value 返回指标当前值。
	Value() float64
}

// Counter 计数器指标接口。
//
// 用于记录单调递增的数值，如请求次数、错误计数。
type Counter interface {
	Metric

	// Inc 计数器加 1。
	Inc()

	// Add 计数器增加指定值。
	Add(delta float64)

	// Reset 重置计数器为 0。
	Reset()
}

// Gauge 仪表盘指标接口。
//
// 用于记录可增可减的数值，如当前连接数、CPU 使用率。
type Gauge interface {
	Metric

	// Set 设置当前值。
	Set(value float64)

	// Add 增加指定值（可以为负数）。
	Add(delta float64)
}

// Histogram 直方图指标接口。
//
// 用于记录分布情况，如请求延迟、响应大小等。
type Histogram interface {
	Metric

	// Observe 记录观测值。
	Observe(value float64)

	// Count 获取观测次数。
	Count() int64

	// Sum 获取观测值总和。
	Sum() float64

	// Min 获取最小值。
	Min() float64

	// Max 获取最大值。
	Max() float64
}

// MetricsExporter 指标导出器接口。
//
// 将指标数据导出到外部监控系统。
type MetricsExporter interface {
	// Export 导出指标数据。
	Export(metrics []Metric) error

	// Close 关闭导出器。
	Close() error
}

// Meter 指标器接口。
//
// 用于创建和管理各种类型的指标。
type Meter interface {
	// Counter 创建或获取计数器。
	Counter(name string) Counter

	// Gauge 创建或获取仪表盘。
	Gauge(name string) Gauge

	// Histogram 创建或获取直方图。
	Histogram(name string) Histogram
}

// MeterProvider 指标提供者接口。
//
// 管理 Meter 实例的生命周期。
type MeterProvider interface {
	// Meter 获取或创建指定名称的指标器。
	Meter(name string) Meter

	// Shutdown 关闭提供者，释放资源。
	Shutdown() error
}

// MetricsRegistry 指标注册表接口。
//
// 管理所有已注册的指标。
type MetricsRegistry interface {
	// Register 注册指标。
	Register(metric Metric)

	// Get 获取指定名称的指标。
	Get(name string) (Metric, bool)

	// List 列出所有指标。
	List() []Metric
}
