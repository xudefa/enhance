package tracing

import (
	"context"
	"testing"
)

func TestTracerHelper_NewTracer_Defaults(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	if tracer.serviceName != DefaultServiceName {
		t.Errorf("serviceName = %s, want %s", tracer.serviceName, DefaultServiceName)
	}
	if tracer.maxSpans != DefaultMaxSpans {
		t.Errorf("maxSpans = %d, want %d", tracer.maxSpans, DefaultMaxSpans)
	}
}

func TestTracerHelper_NewTracer_WithOptions(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(
		WithServiceName("my-service"),
		WithSampler(&AlwaysOffSampler{}),
		WithExporter(exporter),
		WithMaxSpans(100),
	)

	if tracer.serviceName != "my-service" {
		t.Errorf("serviceName = %s, want my-service", tracer.serviceName)
	}
	if tracer.maxSpans != 100 {
		t.Errorf("maxSpans = %d, want 100", tracer.maxSpans)
	}
}

func TestTracerHelper_StartSpan_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))

	span := tracer.StartSpan("op1")
	if span == nil {
		t.Fatal("StartSpan returned nil")
	}
	if span.Name != "op1" {
		t.Errorf("Name = %s, want op1", span.Name)
	}
	if span.TraceID == "" {
		t.Error("TraceID should not be empty")
	}
	if span.SpanID == "" {
		t.Error("SpanID should not be empty")
	}
	if span.Tags["service.name"] != "test" {
		t.Errorf("service.name = %s, want test", span.Tags["service.name"])
	}
}

func TestTracerHelper_StartSpan_WithParent(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	parent := tracer.StartSpan("parent")
	child := tracer.StartSpan("child", WithParent(parent))

	if child.ParentSpanID != parent.SpanID {
		t.Errorf("ParentSpanID = %s, want %s", child.ParentSpanID, parent.SpanID)
	}
	if child.TraceID != parent.TraceID {
		t.Errorf("TraceID mismatch: parent=%s, child=%s", parent.TraceID, child.TraceID)
	}
}

func TestTracerHelper_StartSpan_WithTags(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	span := tracer.StartSpan("op1", WithTags(map[string]string{
		"http.method": "GET",
		"http.url":    "/api",
	}))

	if span.Tags["http.method"] != "GET" {
		t.Errorf("http.method = %s, want GET", span.Tags["http.method"])
	}
	if span.Tags["http.url"] != "/api" {
		t.Errorf("http.url = %s, want /api", span.Tags["http.url"])
	}
}

func TestTracerHelper_StartSpan_WithContext(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	ctx := SpanContext{
		TraceID: "trace-from-header",
		SpanID:  "span-from-header",
		Sampled: true,
	}

	span := tracer.StartSpan("op1", WithContext(ctx))

	if span.TraceID != "trace-from-header" {
		t.Errorf("TraceID = %s, want trace-from-header", span.TraceID)
	}
}

func TestTracerHelper_StartSpan_OffSampler(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithSampler(&AlwaysOffSampler{}))

	span := tracer.StartSpan("op1")
	if !span.Ended {
		t.Error("span should be ended when not sampled")
	}
	if span.TraceID != "" {
		t.Error("TraceID should be empty for non-sampled span")
	}
}

func TestTracerHelper_InjectExtract(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	ctx := SpanContext{
		TraceID:      "trace-1",
		SpanID:       "span-1",
		ParentSpanID: "parent-1",
		Sampled:      true,
	}

	headers := tracer.Inject(ctx)

	extracted := tracer.Extract(headers)

	if extracted.TraceID != "trace-1" {
		t.Errorf("TraceID = %s, want trace-1", extracted.TraceID)
	}
	if extracted.SpanID != "span-1" {
		t.Errorf("SpanID = %s, want span-1", extracted.SpanID)
	}
	if extracted.ParentSpanID != "parent-1" {
		t.Errorf("ParentSpanID = %s, want parent-1", extracted.ParentSpanID)
	}
	if !extracted.Sampled {
		t.Error("Sampled should be true")
	}
}

func TestTracerHelper_GetSpans(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	tracer.StartSpan("op1")
	tracer.StartSpan("op2")

	spans := tracer.GetSpans()
	if len(spans) != 2 {
		t.Errorf("expected 2 spans, got %d", len(spans))
	}
}

func TestTracerHelper_GetSpanCount(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	tracer.StartSpan("op1")
	tracer.StartSpan("op2")

	if tracer.GetSpanCount() != 2 {
		t.Errorf("SpanCount = %d, want 2", tracer.GetSpanCount())
	}
}

func TestTracerHelper_GetActiveSpanCount(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	span1 := tracer.StartSpan("op1")
	tracer.StartSpan("op2")
	span1.End()

	count := tracer.GetActiveSpanCount()
	if count != 1 {
		t.Errorf("ActiveSpanCount = %d, want 1", count)
	}
}

func TestTracerHelper_Clear(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	tracer.StartSpan("op1")
	tracer.Clear()

	spans := tracer.GetSpans()
	if len(spans) != 0 {
		t.Errorf("expected 0 spans after clear, got %d", len(spans))
	}
}

func TestTracerHelper_MaxSpans(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithMaxSpans(3))

	for i := 0; i < 5; i++ {
		tracer.StartSpan("op")
	}

	spans := tracer.GetSpans()
	if len(spans) != 3 {
		t.Errorf("expected 3 spans (maxSpans), got %d", len(spans))
	}
}

func TestTracerHelper_Export_NilExporter(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithExporter(nil))

	err := tracer.Export()
	if err != ErrExporterNotSet {
		t.Errorf("expected ErrExporterNotSet, got %v", err)
	}
}

func TestAlwaysOnSampler_Basic(t *testing.T) {
	t.Parallel()
	s := &AlwaysOnSampler{}
	if !s.ShouldSample() {
		t.Error("AlwaysOnSampler should always return true")
	}
}

func TestAlwaysOffSampler_Basic(t *testing.T) {
	t.Parallel()
	s := &AlwaysOffSampler{}
	if s.ShouldSample() {
		t.Error("AlwaysOffSampler should always return false")
	}
}

func TestProbabilitySampler_Basic(t *testing.T) {
	t.Parallel()
	s := NewProbabilitySampler(1.0)
	if !s.ShouldSample() {
		t.Error("rate 1.0 should always sample")
	}

	s2 := NewProbabilitySampler(0.0)
	if s2.ShouldSample() {
		t.Error("rate 0.0 should never sample")
	}
}

func TestTraceFromContext_NilCtx(t *testing.T) {
	t.Parallel()
	_, ok := TraceFromContext(nil)
	if ok {
		t.Error("nil context should return false")
	}
}

func TestTraceFromContext_EmptyCtx(t *testing.T) {
	t.Parallel()
	_, ok := TraceFromContext(context.Background())
	if ok {
		t.Error("background context should return false")
	}
}

func TestContextWithSpan_Helper(t *testing.T) {
	t.Parallel()
	span := &Span{
		spanContext: SpanContext{
			TraceID: "trace-1",
			SpanID:  "span-1",
			Sampled: true,
		},
	}

	ctx := ContextWithSpan(context.Background(), span)
	got, ok := TraceFromContext(ctx)
	if !ok {
		t.Fatal("should find SpanContext in context")
	}
	if got.TraceID != "trace-1" {
		t.Errorf("TraceID = %s, want trace-1", got.TraceID)
	}
}

func TestGenerateID_Basic(t *testing.T) {
	t.Parallel()
	id := generateID()
	if id == "" {
		t.Error("generateID should not return empty string")
	}
	if len(id) != 16 {
		t.Errorf("id length = %d, want 16", len(id))
	}

	id2 := generateID()
	if id == id2 {
		t.Error("two consecutive IDs should differ")
	}
}
