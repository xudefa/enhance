package tracing

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

type mockExporter struct {
	mu    sync.Mutex
	spans []*Span
}

func (e *mockExporter) ExportSpans(spans []*Span) error {
	e.mu.Lock()
	e.spans = spans
	e.mu.Unlock()
	return nil
}

func (e *mockExporter) getSpans() []*Span {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.spans
}

func TestTracer_StartSpan_Basic(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(
		WithServiceName("test-service"),
		WithExporter(exporter),
	)

	span := tracer.StartSpan("testOperation")
	if span == nil || span.Name != "testOperation" || span.Tags["service.name"] != "test-service" || !span.spanContext.Sampled {
		t.Fatalf("expected valid span, got %+v", span)
	}
}

func TestTracer_StartSpan_WithParent(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test-service"))

	parent := tracer.StartSpan("parentOperation")
	child := tracer.StartSpan("childOperation", WithParent(parent))

	if child.ParentSpanID != parent.SpanID {
		t.Errorf("expected parent span ID %s, got %s", parent.SpanID, child.ParentSpanID)
	}

	if child.TraceID != parent.TraceID {
		t.Errorf("expected same trace ID, parent=%s, child=%s", parent.TraceID, child.TraceID)
	}
}

func TestTracer_StartSpan_WithContext(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test-service"))

	upstream := SpanContext{
		TraceID: TraceID("abc123"),
		SpanID:  SpanID("parent-456"),
		Sampled: false,
	}

	span := tracer.StartSpan("childOperation", WithContext(upstream))

	if span.TraceID != upstream.TraceID {
		t.Errorf("expected trace ID %s, got %s", upstream.TraceID, span.TraceID)
	}

	if span.ParentSpanID != upstream.SpanID {
		t.Errorf("expected parent span ID %s, got %s", upstream.SpanID, span.ParentSpanID)
	}

	if span.spanContext.ParentSpanID != upstream.SpanID {
		t.Errorf("expected span context parent %s, got %s", upstream.SpanID, span.spanContext.ParentSpanID)
	}

	if span.spanContext.Sampled {
		t.Error("expected Sampled=false to be propagated from context")
	}
}

func TestTracer_StartSpan_WithContextSampled(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test-service"))

	upstream := SpanContext{
		TraceID: TraceID("abc123"),
		SpanID:  SpanID("parent-456"),
		Sampled: true,
	}

	span := tracer.StartSpan("childOperation", WithContext(upstream))

	if !span.spanContext.Sampled {
		t.Error("expected Sampled=true to be propagated from context")
	}
}

func TestSpan_SetTag_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("testOperation")

	span.SetTag("http.method", "GET")
	span.SetTag("http.url", "/api/users")

	if span.Tags["http.method"] != "GET" {
		t.Errorf("expected http.method GET, got %s", span.Tags["http.method"])
	}

	if span.Tags["http.url"] != "/api/users" {
		t.Errorf("expected http.url /api/users, got %s", span.Tags["http.url"])
	}
}

func TestSpan_AddEvent_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("testOperation")

	span.AddEvent("query executed")
	span.AddEvent("result processed", map[string]string{
		"row_count": "10",
	})

	if len(span.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(span.Events))
	}

	if span.Events[0].Name != "query executed" {
		t.Errorf("expected first event name 'query executed', got %s", span.Events[0].Name)
	}

	if span.Events[1].Attributes["row_count"] != "10" {
		t.Errorf("expected row_count 10, got %s", span.Events[1].Attributes["row_count"])
	}
}

func TestSpan_End_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("testOperation")

	time.Sleep(10 * time.Millisecond)
	span.End()

	if !span.Ended {
		t.Error("expected span to be ended")
	}

	if span.EndTime.IsZero() {
		t.Error("expected end time to be set")
	}

	if span.Duration() == 0 {
		t.Error("expected duration to be > 0")
	}
}

func TestSpan_End_Idempotent(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("testOperation")

	span.End()
	span.End() // 第二次调用应该无影响

	if !span.Ended {
		t.Error("expected span to be ended")
	}
}

func TestSpan_SetStatus_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("testOperation")

	span.SetStatus(StatusOK)
	if span.Status != StatusOK {
		t.Errorf("expected status OK, got %s", span.Status)
	}

	span.SetStatus(StatusError)
	if span.Status != StatusError {
		t.Errorf("expected status ERROR, got %s", span.Status)
	}
}

func TestTracer_InjectExtract_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("testOperation")

	// 注入上下文
	headers := tracer.Inject(span.Context())

	// 使用 Go http.CanonicalHeaderKey 格式
	if headers["X-Trace-Id"] != string(span.TraceID) {
		t.Errorf("expected X-Trace-Id %s, got %s", span.TraceID, headers["X-Trace-Id"])
	}

	if headers["X-Span-Id"] != string(span.SpanID) {
		t.Errorf("expected X-Span-Id %s, got %s", span.SpanID, headers["X-Span-Id"])
	}

	// 提取上下文
	ctx := tracer.Extract(headers)

	if ctx.TraceID != span.TraceID {
		t.Errorf("expected trace ID %s, got %s", span.TraceID, ctx.TraceID)
	}

	if ctx.SpanID != span.SpanID {
		t.Errorf("expected span ID %s, got %s", span.SpanID, ctx.SpanID)
	}
}

func TestTracer_Export_Basic(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(WithExporter(exporter))

	tracer.StartSpan("operation1")
	tracer.StartSpan("operation2")

	err := tracer.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(exporter.getSpans()) != 2 {
		t.Errorf("expected 2 spans exported, got %d", len(exporter.getSpans()))
	}
}

func TestTracer_Clear_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	tracer.StartSpan("operation1")
	tracer.StartSpan("operation2")

	if len(tracer.GetSpans()) != 2 {
		t.Errorf("expected 2 spans, got %d", len(tracer.GetSpans()))
	}

	tracer.Clear()

	if len(tracer.GetSpans()) != 0 {
		t.Errorf("expected 0 spans after clear, got %d", len(tracer.GetSpans()))
	}
}

func TestSampler_AlwaysOn_Basic(t *testing.T) {
	t.Parallel()
	sampler := &AlwaysOnSampler{}

	for i := 0; i < 10; i++ {
		if !sampler.ShouldSample() {
			t.Error("expected AlwaysOnSampler to always return true")
		}
	}
}

func TestSampler_AlwaysOff_Basic(t *testing.T) {
	t.Parallel()
	sampler := &AlwaysOffSampler{}

	for i := 0; i < 10; i++ {
		if sampler.ShouldSample() {
			t.Error("expected AlwaysOffSampler to always return false")
		}
	}
}

func TestSampler_Probability_Basic(t *testing.T) {
	t.Parallel()
	sampler := NewProbabilitySampler(0.5)

	sampled := 0
	for i := 0; i < 100; i++ {
		if sampler.ShouldSample() {
			sampled++
		}
	}

	// 大约 50% 的采样率（允许一定误差）
	if sampled < 30 || sampled > 70 {
		t.Errorf("expected approximately 50 samples, got %d", sampled)
	}
}

func TestTraceHelper_TraceHTTP_Success(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(WithExporter(exporter))
	helper := NewTraceHelper(tracer)

	err := helper.TraceHTTP("GET", "/api/users", func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("TraceHTTP failed: %v", err)
	}

	// 导出 span
	err = tracer.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(exporter.getSpans()) != 1 {
		t.Fatalf("expected 1 span, got %d", len(exporter.getSpans()))
	}

	span := exporter.getSpans()[0]
	if span.Tags["http.method"] != "GET" {
		t.Errorf("expected http.method GET, got %s", span.Tags["http.method"])
	}

	if span.Tags["http.url"] != "/api/users" {
		t.Errorf("expected http.url /api/users, got %s", span.Tags["http.url"])
	}

	if span.Status != StatusOK {
		t.Errorf("expected status OK, got %s", span.Status)
	}
}

func TestTraceHelper_TraceHTTP_Error(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(WithExporter(exporter))
	helper := NewTraceHelper(tracer)

	expectedErr := &testError{"connection failed"}
	err := helper.TraceHTTP("POST", "/api/users", func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	// 导出 span
	err = tracer.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(exporter.getSpans()) == 0 {
		t.Fatal("expected at least 1 span")
	}

	span := exporter.getSpans()[0]
	if span.Status != StatusError {
		t.Errorf("expected status ERROR, got %s", span.Status)
	}

	if span.Tags["error"] != "connection failed" {
		t.Errorf("expected error tag 'connection failed', got %s", span.Tags["error"])
	}
}

func TestTraceHelper_TraceDB_Success(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(WithExporter(exporter))
	helper := NewTraceHelper(tracer)

	err := helper.TraceDB("SELECT", "SELECT * FROM users", func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("TraceDB failed: %v", err)
	}

	err = tracer.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(exporter.getSpans()) == 0 {
		t.Fatal("expected at least 1 span")
	}

	span := exporter.getSpans()[0]
	if span.Tags["db.operation"] != "SELECT" {
		t.Errorf("expected db.operation SELECT, got %s", span.Tags["db.operation"])
	}

	if span.Tags["db.statement"] != "SELECT * FROM users" {
		t.Errorf("expected db.statement, got %s", span.Tags["db.statement"])
	}
}

func TestTraceHelper_TraceRPC_Success(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(WithExporter(exporter))
	helper := NewTraceHelper(tracer)

	err := helper.TraceRPC("UserService", "GetUser", func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("TraceRPC failed: %v", err)
	}

	err = tracer.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(exporter.getSpans()) == 0 {
		t.Fatal("expected at least 1 span")
	}

	span := exporter.getSpans()[0]
	if span.Tags["rpc.service"] != "UserService" {
		t.Errorf("expected rpc.service UserService, got %s", span.Tags["rpc.service"])
	}

	if span.Tags["rpc.method"] != "GetUser" {
		t.Errorf("expected rpc.method GetUser, got %s", span.Tags["rpc.method"])
	}
}

func TestSpan_Context_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	span := tracer.StartSpan("testOperation")

	ctx := span.Context()

	if ctx.TraceID != span.TraceID {
		t.Errorf("expected trace ID %s, got %s", span.TraceID, ctx.TraceID)
	}

	if ctx.SpanID != span.SpanID {
		t.Errorf("expected span ID %s, got %s", span.SpanID, ctx.SpanID)
	}
}

func TestTracer_GetSpans_ReturnsCopy(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	span1 := tracer.StartSpan("operation1")
	span2 := tracer.StartSpan("operation2")
	_ = span2

	spans := tracer.GetSpans()

	if len(spans) != 2 {
		t.Errorf("expected 2 spans, got %d", len(spans))
	}

	// 验证返回的 span 与创建的匹配
	if spans[0].Name != "operation1" {
		t.Errorf("expected first span name operation1, got %s", spans[0].Name)
	}

	if spans[1].Name != "operation2" {
		t.Errorf("expected second span name operation2, got %s", spans[1].Name)
	}

	// 修改切片不应该影响 tracer 的内部状态
	originalLen := len(tracer.GetSpans())
	spans[0] = nil
	if len(tracer.GetSpans()) != originalLen {
		t.Error("expected GetSpans to return a copy of the slice")
	}

	_ = span1
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

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
	data, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// 验证包含必要字段
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if result["trace_id"] == "" {
		t.Error("missing trace_id field")
	}
	if result["span_id"] == "" {
		t.Error("missing span_id field")
	}
	if result["name"] != "test-operation" {
		t.Errorf("expected name='test-operation', got '%s'", result["name"])
	}
	if result["status"] != string(StatusOK) {
		t.Errorf("expected status='OK', got '%s'", result["status"])
	}
	if result["duration_ms"] == nil {
		t.Error("missing duration_ms field")
	}
	if result["tags"] == nil {
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
