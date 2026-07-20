package metrics

// MetricBuilder 统一的指标构建器
//
// 使用泛型 Builder 模式替代重复的 CounterBuilder/GaugeBuilder/HistogramBuilder。
// 支持链式配置标签，符合 Go 惯用法。
//
// 使用示例：
//
//	// 创建计数器
//	counter := NewMetricBuilder(registry, "http_requests").
//	    Tag("method", "GET").
//	    Tag("status", "200").
//	    BuildCounter()
//
//	// 创建仪表盘
//	gauge := NewMetricBuilder(registry, "memory_usage").
//	    Tag("type", "heap").
//	    BuildGauge()
//
//	// 创建直方图
//	histogram := NewMetricBuilder(registry, "request_duration").
//	    Tag("endpoint", "/api").
//	    BuildHistogram()
type MetricBuilder struct {
	name     string
	tags     map[string]string
	registry MeterRegistry
}

// NewMetricBuilder 创建指标构建器
func NewMetricBuilder(registry MeterRegistry, name string) *MetricBuilder {
	return &MetricBuilder{
		name:     name,
		tags:     make(map[string]string),
		registry: registry,
	}
}

// Tag 添加单个标签
func (b *MetricBuilder) Tag(key, value string) *MetricBuilder {
	b.tags[key] = value
	return b
}

// Tags 批量添加标签
func (b *MetricBuilder) Tags(tags map[string]string) *MetricBuilder {
	for k, v := range tags {
		b.tags[k] = v
	}
	return b
}

// BuildCounter 构建计数器
func (b *MetricBuilder) BuildCounter() Counter {
	return b.registry.Counter(b.name, flattenTags(b.tags)...)
}

// BuildGauge 构建仪表盘
func (b *MetricBuilder) BuildGauge() Gauge {
	return b.registry.Gauge(b.name, flattenTags(b.tags)...)
}

// BuildHistogram 构建直方图
func (b *MetricBuilder) BuildHistogram() Histogram {
	return b.registry.Histogram(b.name, flattenTags(b.tags)...)
}

// CounterBuilder 计数器构建器（向后兼容）
//
// Deprecated: 使用 NewMetricBuilder(registry, name).BuildCounter() 替代
type CounterBuilder struct {
	name     string
	tags     map[string]string
	registry MeterRegistry
}

// NewCounterBuilder 创建计数器构建器
//
// Deprecated: 使用 NewMetricBuilder(registry, name) 替代
func NewCounterBuilder(registry MeterRegistry, name string) *CounterBuilder {
	return &CounterBuilder{
		name:     name,
		tags:     make(map[string]string),
		registry: registry,
	}
}

// Tag 添加单个标签
func (b *CounterBuilder) Tag(key, value string) *CounterBuilder {
	b.tags[key] = value
	return b
}

// Tags 批量添加标签
func (b *CounterBuilder) Tags(tags map[string]string) *CounterBuilder {
	for k, v := range tags {
		b.tags[k] = v
	}
	return b
}

// Build 构建计数器
func (b *CounterBuilder) Build() Counter {
	return b.registry.Counter(b.name, flattenTags(b.tags)...)
}

// GaugeBuilder 仪表盘构建器（向后兼容）
//
// Deprecated: 使用 NewMetricBuilder(registry, name).BuildGauge() 替代
type GaugeBuilder struct {
	name     string
	tags     map[string]string
	registry MeterRegistry
}

// NewGaugeBuilder 创建仪表盘构建器
//
// Deprecated: 使用 NewMetricBuilder(registry, name) 替代
func NewGaugeBuilder(registry MeterRegistry, name string) *GaugeBuilder {
	return &GaugeBuilder{
		name:     name,
		tags:     make(map[string]string),
		registry: registry,
	}
}

// Tag 添加单个标签
func (b *GaugeBuilder) Tag(key, value string) *GaugeBuilder {
	b.tags[key] = value
	return b
}

// Tags 批量添加标签
func (b *GaugeBuilder) Tags(tags map[string]string) *GaugeBuilder {
	for k, v := range tags {
		b.tags[k] = v
	}
	return b
}

// Build 构建仪表盘
func (b *GaugeBuilder) Build() Gauge {
	return b.registry.Gauge(b.name, flattenTags(b.tags)...)
}

// HistogramBuilder 直方图构建器（向后兼容）
//
// Deprecated: 使用 NewMetricBuilder(registry, name).BuildHistogram() 替代
type HistogramBuilder struct {
	name     string
	tags     map[string]string
	registry MeterRegistry
}

// NewHistogramBuilder 创建直方图构建器
//
// Deprecated: 使用 NewMetricBuilder(registry, name) 替代
func NewHistogramBuilder(registry MeterRegistry, name string) *HistogramBuilder {
	return &HistogramBuilder{
		name:     name,
		tags:     make(map[string]string),
		registry: registry,
	}
}

// Tag 添加单个标签
func (b *HistogramBuilder) Tag(key, value string) *HistogramBuilder {
	b.tags[key] = value
	return b
}

// Tags 批量添加标签
func (b *HistogramBuilder) Tags(tags map[string]string) *HistogramBuilder {
	for k, v := range tags {
		b.tags[k] = v
	}
	return b
}

// Build 构建直方图
func (b *HistogramBuilder) Build() Histogram {
	return b.registry.Histogram(b.name, flattenTags(b.tags)...)
}

// flattenTags 将标签 map 转换为切片格式
func flattenTags(tags map[string]string) []string {
	if len(tags) == 0 {
		return nil
	}
	result := make([]string, 0, len(tags)*2)
	for k, v := range tags {
		result = append(result, k, v)
	}
	return result
}
