package metrics

import (
	"strings"
	"testing"
)

func TestPrometheusExporter(t *testing.T) {
	t.Parallel()
	// 创建一个简单的内存写入器
	var buffer strings.Builder

	exporter := NewPrometheusExporter(&buffer)

	// 创建测试指标
	metrics := testPrometheusExporterMetrics()

	err := exporter.Export(metrics)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, `# TYPE test_counter_total counter`) {
		t.Fatalf("expected counter TYPE line in output, got: %s", output)
	}
	if !strings.Contains(output, `test_counter_total{method="GET",path="/api"} 42`) {
		t.Fatalf("expected counter with labels in output, got: %s", output)
	}
	if !strings.Contains(output, `test_gauge 123.45`) {
		t.Fatalf("expected gauge in output, got: %s", output)
	}
	if !strings.Contains(output, `# TYPE test_histogram histogram`) {
		t.Fatalf("expected histogram TYPE line in output, got: %s", output)
	}
	if !strings.Contains(output, `test_histogram_bucket{le="+Inf",path="/api"} 3`) ||
		!strings.Contains(output, `test_histogram_sum{path="/api"} 1500`) ||
		!strings.Contains(output, `test_histogram_count{path="/api"} 3`) {
		t.Fatalf("expected histogram lines in output, got: %s", output)
	}
	if !strings.Contains(output, `test_escape{label="a\\b\"c\\nd"} 1`) {
		t.Fatalf("expected escaped label value in output, got: %s", output)
	}
}

func testPrometheusExporterMetrics() []Metric {
	return []Metric{
		{
			Name:  "test_counter",
			Value: 42.0,
			Type:  "counter",
			Tags:  map[string]string{"method": "GET", "path": "/api"},
		},
		{
			Name:  "test_gauge",
			Value: 123.45,
			Type:  "gauge",
			Tags:  nil,
		},
		{
			Name:  "test_histogram",
			Value: 500.0,
			Type:  "histogram",
			Count: 3,
			Sum:   1500.0,
			Tags:  map[string]string{"path": "/api"},
		},
		{
			Name:  "test_escape",
			Value: 1.0,
			Type:  "gauge",
			Tags:  map[string]string{"label": `a\b"c\nd`},
		},
	}
}

func TestRegistry_MultipleExporters(t *testing.T) {
	t.Parallel()
	registry := NewSimpleRegistry()
	registry.Counter("requests").Inc()

	// 创建多个导出器
	var buffer1, buffer2 strings.Builder
	exporter1 := NewPrometheusExporter(&buffer1)
	exporter2 := NewPrometheusExporter(&buffer2)

	registry.RegisterExporter(exporter1)
	registry.RegisterExporter(exporter2)

	err := registry.Export()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 检查两个导出器都收到了数据
	output1 := buffer1.String()
	output2 := buffer2.String()

	if !strings.Contains(output1, "requests") {
		t.Fatalf("expected requests in exporter1 output, got: %s", output1)
	}
	if !strings.Contains(output2, "requests") {
		t.Fatalf("expected requests in exporter2 output, got: %s", output2)
	}
}
