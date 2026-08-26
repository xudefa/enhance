package metrics

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewPrometheusExporter_NilWriter(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil writer")
		}
	}()

	NewPrometheusExporter(nil)
}

func TestNewPrometheusExporter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	exporter := NewPrometheusExporter(&buf)
	if exporter == nil {
		t.Fatal("expected non-nil exporter")
	}
}

func TestPrometheusExporter_Export_Counter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	exporter := NewPrometheusExporter(&buf)

	metrics := []Metric{
		{
			Name:  "http_requests",
			Type:  "counter",
			Value: 42,
			Tags:  map[string]string{"method": "GET"},
		},
	}

	err := exporter.Export(metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# TYPE http_requests_total counter") {
		t.Errorf("expected TYPE line, got %q", output)
	}
	if !strings.Contains(output, "http_requests_total") {
		t.Errorf("expected counter name with _total suffix, got %q", output)
	}
	if !strings.Contains(output, `method="GET"`) {
		t.Errorf("expected method label, got %q", output)
	}
}

func TestPrometheusExporter_Export_Gauge(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	exporter := NewPrometheusExporter(&buf)

	metrics := []Metric{
		{
			Name:  "memory_usage",
			Type:  "gauge",
			Value: 1024.5,
			Tags:  map[string]string{"type": "heap"},
		},
	}

	err := exporter.Export(metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# TYPE memory_usage gauge") {
		t.Errorf("expected TYPE gauge line, got %q", output)
	}
	if !strings.Contains(output, "memory_usage") {
		t.Errorf("expected gauge name, got %q", output)
	}
}

func TestPrometheusExporter_Export_Histogram(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	exporter := NewPrometheusExporter(&buf)

	metrics := []Metric{
		{
			Name:  "request_duration",
			Type:  "histogram",
			Value: 0,
			Tags:  map[string]string{"endpoint": "/api"},
			Count: 10,
			Sum:   1500.5,
		},
	}

	err := exporter.Export(metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# TYPE request_duration histogram") {
		t.Errorf("expected TYPE histogram line, got %q", output)
	}
	if !strings.Contains(output, `le="+Inf"`) {
		t.Errorf("expected +Inf bucket, got %q", output)
	}
	if !strings.Contains(output, "request_duration_sum") {
		t.Errorf("expected _sum metric, got %q", output)
	}
	if !strings.Contains(output, "request_duration_count") {
		t.Errorf("expected _count metric, got %q", output)
	}
}

func TestPrometheusExporter_Export_Empty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	exporter := NewPrometheusExporter(&buf)

	err := exporter.Export([]Metric{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestFormatLabels_Empty(t *testing.T) {
	t.Parallel()

	got := FormatLabels(nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	got = FormatLabels(map[string]string{})
	if got != "" {
		t.Errorf("expected empty for empty map, got %q", got)
	}
}

func TestFormatLabels_Single(t *testing.T) {
	t.Parallel()

	got := FormatLabels(map[string]string{"method": "GET"})
	if got != `{method="GET"}` {
		t.Errorf("expected {method=\"GET\"}, got %q", got)
	}
}

func TestFormatLabels_Multiple_Sorted(t *testing.T) {
	t.Parallel()

	got := FormatLabels(map[string]string{"z": "1", "a": "2", "m": "3"})
	expected := `{a="2",m="3",z="1"}`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestEscapeLabelValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "hello", "hello"},
		{"backslash", `a\b`, `a\\b`},
		{"quote", `a"b`, `a\"b`},
		{"newline", "a\nb", `a\nb`},
		{"combined", `a"b\nc`, `a\"b\\nc`},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EscapeLabelValue(tt.input)
			if got != tt.want {
				t.Errorf("EscapeLabelValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
