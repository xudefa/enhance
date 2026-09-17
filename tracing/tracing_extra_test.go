package tracing

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestTracer_MaxSpans_Limit(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(
		WithServiceName("test-service"),
		WithMaxSpans(5),
	)

	// 创建 10 个 Span
	for i := 0; i < 10; i++ {
		span := tracer.StartSpan("span")
		span.End()
	}

	spans := tracer.GetSpans()
	if len(spans) != 5 {
		t.Errorf("expected 5 spans, got %d", len(spans))
	}
}

func TestTracer_SpanCount_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(
		WithServiceName("test-service"),
	)

	// 创建 10 个 Span
	for i := 0; i < 10; i++ {
		span := tracer.StartSpan("span")
		span.End()
	}

	count := tracer.GetSpanCount()
	if count != 10 {
		t.Errorf("expected span count 10, got %d", count)
	}

	// 所有 span 已结束，活跃 span 应为 0
	activeCount := tracer.GetActiveSpanCount()
	if activeCount != 0 {
		t.Errorf("expected active span count 0 (all ended), got %d", activeCount)
	}

	// 创建 3 个不结束的 span，验证活跃计数
	for i := 0; i < 3; i++ {
		tracer.StartSpan("active-span")
	}
	activeCount = tracer.GetActiveSpanCount()
	if activeCount != 3 {
		t.Errorf("expected active span count 3, got %d", activeCount)
	}
}

func TestTracer_ConcurrentAccess_Safety(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(
		WithServiceName("test-service"),
		WithMaxSpans(100),
	)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			span := tracer.StartSpan("concurrent-span")
			time.Sleep(time.Millisecond)
			span.SetTag("index", string(rune('0'+n%10)))
			span.End()
		}(i)
	}

	wg.Wait()

	count := tracer.GetSpanCount()
	if count != 100 {
		t.Errorf("expected 100 spans, got %d", count)
	}
}

func TestConstants_Basic(t *testing.T) {
	t.Parallel()
	// 使用 Go http.CanonicalHeaderKey 格式（X-Trace-Id 而非 X-Trace-ID）
	if HeaderTraceID != "X-Trace-Id" {
		t.Errorf("expected HeaderTraceID 'X-Trace-Id', got '%s'", HeaderTraceID)
	}
	if HeaderSpanID != "X-Span-Id" {
		t.Errorf("expected HeaderSpanID 'X-Span-Id', got '%s'", HeaderSpanID)
	}
	if HeaderParentSpanID != "X-Parent-Span-Id" {
		t.Errorf("expected HeaderParentSpanID 'X-Parent-Span-Id', got '%s'", HeaderParentSpanID)
	}
	if HeaderSampled != "X-Sampled" {
		t.Errorf("expected HeaderSampled 'X-Sampled', got '%s'", HeaderSampled)
	}
	if DefaultServiceName != "enhance-app" {
		t.Errorf("expected DefaultServiceName 'enhance-app', got '%s'", DefaultServiceName)
	}
	if DefaultMaxSpans != 10000 {
		t.Errorf("expected DefaultMaxSpans 10000, got %d", DefaultMaxSpans)
	}
}

func TestSpan_MarshalJSON_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(
		WithServiceName("test-service"),
	)

	span := tracer.StartSpan("test-operation")
	time.Sleep(10 * time.Millisecond)
	span.SetTag("http.method", "GET")
	span.SetTag("http.url", "/api/test")
	span.SetStatus(StatusOK)
	span.End()

	// 测试 JSON 序列化
	jsonBytes, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// 验证包含必要字段
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if parsed["trace_id"] == "" {
		t.Error("missing trace_id field")
	}
	if parsed["span_id"] == "" {
		t.Error("missing span_id field")
	}
	if parsed["name"] != "test-operation" {
		t.Errorf("expected name='test-operation', got '%s'", parsed["name"])
	}
	if parsed["status"] != string(StatusOK) {
		t.Errorf("expected status='OK', got '%s'", parsed["status"])
	}
	if parsed["duration_ms"] == nil {
		t.Error("missing duration_ms field")
	}
	if parsed["tags"] == nil {
		t.Error("missing tags field")
	}
}

func TestTracer_Export_NoExporter(t *testing.T) {
	t.Parallel()
	// 创建没有 exporter 的 tracer
	tracer := &Tracer{
		sampler:  &AlwaysOnSampler{},
		maxSpans: DefaultMaxSpans,
	}
	tracer.StartSpan("test-operation")

	err := tracer.Export()
	if !errors.Is(err, ErrExporterNotSet) {
		t.Errorf("expected ErrExporterNotSet, got %v", err)
	}
}

func TestContextWithSpan_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("test-operation")

	ctx := ContextWithSpan(context.Background(), span)
	spanCtx, ok := TraceFromContext(ctx)

	if !ok {
		t.Error("expected to get span context from context")
	}

	if spanCtx.TraceID != span.TraceID {
		t.Errorf("expected trace ID %s, got %s", span.TraceID, spanCtx.TraceID)
	}
}

func TestTraceFromContext_Empty(t *testing.T) {
	t.Parallel()
	ctx, ok := TraceFromContext(context.TODO())

	if ok {
		t.Error("expected false for empty context")
	}

	if ctx.TraceID != "" {
		t.Error("expected empty trace ID")
	}
}

func TestTracer_WithSampler(t *testing.T) {
	t.Parallel()

	sampler := &AlwaysOnSampler{}
	tracer := NewTracer(WithSampler(sampler))

	if tracer.sampler != sampler {
		t.Error("expected sampler to be set")
	}
}

func TestExporter_ExportSpans_Error(t *testing.T) {
	t.Parallel()

	exporter := &errorExporter{}
	tracer := NewTracer(WithExporter(exporter))

	span := tracer.StartSpan("test")
	span.End()

	err := tracer.Export()
	if err == nil {
		t.Error("expected error from exporter")
	}
}

// errorExporter 模拟导出失败的导出器
type errorExporter struct{}

func (e *errorExporter) ExportSpans(spans []*Span) error {
	return errors.New("export failed")
}

func TestSpan_Duration_NotEnded(t *testing.T) {
	t.Parallel()

	tracer := NewTracer()
	span := tracer.StartSpan("test")

	// 未结束的span应该返回从开始到现在的时长（应该>0）
	duration := span.Duration()
	if duration < 0 {
		t.Errorf("expected positive duration for non-ended span, got %d", duration)
	}

	// 结束span后再检查
	span.End()
	endedDuration := span.Duration()
	if endedDuration < 0 {
		t.Errorf("expected positive duration for ended span, got %d", endedDuration)
	}
}

func TestTracer_Extract_EmptyHeaders(t *testing.T) {
	t.Parallel()

	tracer := NewTracer()
	headers := map[string]string{}

	ctx := tracer.Extract(headers)

	if ctx.TraceID != "" {
		t.Errorf("expected empty trace ID, got %s", ctx.TraceID)
	}
}

func TestSpan_SetTag_NilTags(t *testing.T) {
	t.Parallel()

	tracer := NewTracer()
	span := tracer.StartSpan("test")

	// 确保Tags已初始化
	if span.Tags == nil {
		t.Error("expected Tags to be initialized")
	}

	span.SetTag("key", "value")
	if span.Tags["key"] != "value" {
		t.Errorf("expected tag value 'value', got %v", span.Tags["key"])
	}
}
