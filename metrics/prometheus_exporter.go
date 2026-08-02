package metrics

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// PrometheusExporter Prometheus 格式指标导出器
//
// 将指标数据转换为 Prometheus 兼容的文本格式。
// 支持 Counter、Gauge、Histogram 等指标类型的导出。
type PrometheusExporter struct {
	writer io.Writer // 输出流
}

// NewPrometheusExporter 创建新的 Prometheus 导出器
func NewPrometheusExporter(writer io.Writer) Exporter {
	if writer == nil {
		panic("metrics: prometheus exporter writer must not be nil")
	}
	return &PrometheusExporter{
		writer: writer,
	}
}

// Export 将指标导出为 Prometheus 格式
func (e *PrometheusExporter) Export(metrics []Metric) error {
	for _, metric := range metrics {
		if err := e.writeMetric(metric); err != nil {
			return err
		}
	}
	return nil
}

// writeMetric 将单个指标写入输出流
func (e *PrometheusExporter) writeMetric(metric Metric) error {
	labels := FormatLabels(metric.Tags)

	switch metric.Type {
	case "counter":
		return e.writeCounter(metric, labels)
	case "histogram":
		return e.writeHistogram(metric, labels)
	default:
		return e.writeGauge(metric, labels)
	}
}

// writeCounter 写入 Counter 指标（Prometheus 要求 counter 名称以 _total 结尾）
func (e *PrometheusExporter) writeCounter(metric Metric, labels string) error {
	name := metric.Name + "_total"
	if err := e.writeHeader(name, "counter"); err != nil {
		return err
	}
	return e.writeLine(name, labels, metric.Value)
}

// writeGauge 写入 Gauge 指标
func (e *PrometheusExporter) writeGauge(metric Metric, labels string) error {
	if err := e.writeHeader(metric.Name, "gauge"); err != nil {
		return err
	}
	return e.writeLine(metric.Name, labels, metric.Value)
}

// writeHistogram 写入 Histogram 指标（_bucket / _sum / _count）
//
// 快照未包含桶边界数据，因此只导出累计桶（le="+Inf"）、总和与计数。
func (e *PrometheusExporter) writeHistogram(metric Metric, labels string) error {
	if err := e.writeHeader(metric.Name, "histogram"); err != nil {
		return err
	}
	bucketLabels := `{le="+Inf"}`
	if labels != "" {
		bucketLabels = `{le="+Inf",` + strings.TrimPrefix(labels, "{")
	}
	if err := e.writeLine(metric.Name+"_bucket", bucketLabels, float64(metric.Count)); err != nil {
		return err
	}
	if err := e.writeLine(metric.Name+"_sum", labels, metric.Sum); err != nil {
		return err
	}
	return e.writeLine(metric.Name+"_count", labels, float64(metric.Count))
}

// writeHeader 写入 TYPE 和 HELP 注释行
func (e *PrometheusExporter) writeHeader(name, mtype string) error {
	if _, err := fmt.Fprintf(e.writer, "# HELP %s %s metric\n", name, mtype); err != nil {
		return err
	}
	_, err := fmt.Fprintf(e.writer, "# TYPE %s %s\n", name, mtype)
	return err
}

// writeLine 写入指标样本行
func (e *PrometheusExporter) writeLine(name, labels string, value float64) error {
	_, err := fmt.Fprintf(e.writer, "%s%s %g\n", name, labels, value)
	return err
}

// FormatLabels 格式化标签，并对标签值做 Prometheus 转义。
//
// 标签按键排序以保证输出确定性，空标签返回空字符串。
func FormatLabels(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`, k, EscapeLabelValue(tags[k])))
	}
	return "{" + strings.Join(pairs, ",") + "}"
}

// EscapeLabelValue 转义 Prometheus 标签值中的特殊字符（\\、"、\n）。
func EscapeLabelValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, "\n", `\n`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	return v
}
