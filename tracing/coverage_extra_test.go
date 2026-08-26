package tracing

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

// mockApplicationContext implements boot.ApplicationContext for testing.
type mockApplicationContext struct {
	container core.Container
	env       *environment.Environment
	ctx       context.Context
}

func (m *mockApplicationContext) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func (m *mockApplicationContext) Container() core.Container {
	return m.container
}

func (m *mockApplicationContext) Environment() *environment.Environment {
	return m.env
}

func (m *mockApplicationContext) Register(t reflect.Type, opts ...core.BeanOption) error {
	return nil
}

func (m *mockApplicationContext) GetByType(t reflect.Type) (any, error) {
	return nil, errors.New("not found")
}

func (m *mockApplicationContext) EventBus() boot.EventBusResult {
	return nil
}

func TestTracingAutoConfiguration_Configure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		props      map[string]any
		wantErr    bool
		wantRate   float64
		wantSpans  int
		wantLoaded bool
	}{
		{
			name: "default config",
			props: map[string]any{
				"tracing.enabled": "true",
			},
			wantRate:   DefaultSamplingRate,
			wantSpans:  DefaultMaxSpans,
			wantLoaded: true,
		},
		{
			name: "custom config with sampling",
			props: map[string]any{
				"tracing.enabled":       "true",
				"tracing.service_name":  "my-svc",
				"tracing.sampling_rate": "0.5",
				"tracing.max_spans":     "500",
			},
			wantRate:   0.5,
			wantSpans:  500,
			wantLoaded: true,
		},
		{
			name: "custom config without sampling",
			props: map[string]any{
				"tracing.enabled":      "true",
				"tracing.service_name": "full-svc",
				"tracing.max_spans":    "2000",
			},
			wantRate:   1.0,
			wantSpans:  2000,
			wantLoaded: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := environment.NewEnvironment()
			env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, tt.props))

			container := core.NewContainer()
			cfg := &TracingAutoConfiguration{}
			appCtx := &mockApplicationContext{
				container: container,
				env:       env,
			}

			err := cfg.Configure(appCtx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Configure() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantLoaded {
				tracer := cfg.GetTracer()
				if tracer == nil {
					t.Fatal("expected tracer to be set")
				}
				if tracer.serviceName != "my-svc" && tt.name == "custom config with sampling" {
					t.Errorf("expected service name 'my-svc', got %s", tracer.serviceName)
				}
				if tracer.serviceName != "full-svc" && tt.name == "custom config without sampling" {
					t.Errorf("expected service name 'full-svc', got %s", tracer.serviceName)
				}
			}
		})
	}
}

func TestTracingAutoConfiguration_GetTracer(t *testing.T) {
	t.Parallel()
	cfg := &TracingAutoConfiguration{}
	if cfg.GetTracer() != nil {
		t.Error("expected nil tracer before Configure")
	}

	cfg.tracer = NewTracer(WithServiceName("test"))
	if cfg.GetTracer() == nil {
		t.Error("expected non-nil tracer after manual set")
	}
}

func TestTracingAutoConfiguration_LoadConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		props      map[string]any
		wantName   string
		wantRate   float64
		wantMax    int
	}{
		{
			name:     "defaults",
			props:    map[string]any{},
			wantName: DefaultServiceName,
			wantRate: DefaultSamplingRate,
			wantMax:  DefaultMaxSpans,
		},
		{
			name: "partial config",
			props: map[string]any{
				"tracing.service_name": "custom",
			},
			wantName: "custom",
			wantRate: DefaultSamplingRate,
			wantMax:  DefaultMaxSpans,
		},
		{
			name: "zero sampling rate",
			props: map[string]any{
				"tracing.sampling_rate": "0.0",
			},
			wantName: DefaultServiceName,
			wantRate: 0.0,
			wantMax:  DefaultMaxSpans,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := environment.NewEnvironment()
			env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, tt.props))

			cfg := &TracingAutoConfiguration{}
			got, err := cfg.loadConfig(env)
			if err != nil {
				t.Fatalf("loadConfig() error = %v", err)
			}

			if got.ServiceName != tt.wantName {
				t.Errorf("ServiceName = %s, want %s", got.ServiceName, tt.wantName)
			}
			if got.SamplingRate != tt.wantRate {
				t.Errorf("SamplingRate = %f, want %f", got.SamplingRate, tt.wantRate)
			}
			if got.MaxSpans != tt.wantMax {
				t.Errorf("MaxSpans = %d, want %d", got.MaxSpans, tt.wantMax)
			}
		})
	}
}

func TestSpan_SetTag_NilMap(t *testing.T) {
	t.Parallel()
	span := &Span{
		Name:  "test",
		Tags:  nil,
		Events: make([]SpanEvent, 0),
	}

	span.SetTag("key", "value")

	if span.Tags == nil {
		t.Fatal("expected Tags to be initialized")
	}
	if span.Tags["key"] != "value" {
		t.Errorf("expected tag value 'value', got %s", span.Tags["key"])
	}
}

func TestSpan_MarshalJSON_NotEnded(t *testing.T) {
	t.Parallel()
	span := &Span{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Name:      "test-op",
		StartTime: time.Now().Add(-time.Second),
		Ended:     false,
		Tags:      make(map[string]string),
		Events:    make([]SpanEvent, 0),
	}

	data, err := span.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if result["ended"] != false {
		t.Errorf("expected ended=false, got %v", result["ended"])
	}
	if result["duration_ms"] == nil {
		t.Error("expected duration_ms to be set")
	}
}

func TestTraceHelper_TraceDB_Error(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	helper := NewTraceHelper(tracer)
	testErr := errors.New("db timeout")

	err := helper.TraceDB("INSERT", "INSERT INTO t VALUES(1)", func() error {
		return testErr
	})

	if !errors.Is(err, testErr) {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
}

func TestTraceHelper_TraceRPC_Error(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	helper := NewTraceHelper(tracer)
	testErr := errors.New("rpc deadline exceeded")

	err := helper.TraceRPC("UserService", "GetUser", func() error {
		return testErr
	})

	if !errors.Is(err, testErr) {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
}

func TestTracer_Inject_WithoutParentSpanID(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	ctx := SpanContext{
		TraceID: "trace-1",
		SpanID:  "span-1",
		Sampled: true,
	}

	headers := tracer.Inject(ctx)

	if _, ok := headers[HeaderParentSpanID]; ok {
		t.Error("expected no ParentSpanID header when empty")
	}
	if headers[HeaderTraceID] != "trace-1" {
		t.Errorf("expected trace ID 'trace-1', got %s", headers[HeaderTraceID])
	}
	if headers[HeaderSampled] != "true" {
		t.Errorf("expected sampled 'true', got %s", headers[HeaderSampled])
	}
}

func TestTracer_Extract_WithParentSpanID(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	headers := map[string]string{
		HeaderTraceID:      "trace-1",
		HeaderSpanID:       "span-1",
		HeaderParentSpanID: "parent-1",
		HeaderSampled:      "true",
	}

	ctx := tracer.Extract(headers)

	if ctx.ParentSpanID != "parent-1" {
		t.Errorf("expected parent span ID 'parent-1', got %s", ctx.ParentSpanID)
	}
	if ctx.TraceID != "trace-1" {
		t.Errorf("expected trace ID 'trace-1', got %s", ctx.TraceID)
	}
}

func TestTraceFromContext_NilContext(t *testing.T) {
	t.Parallel()
	_, ok := TraceFromContext(nil)
	if ok {
		t.Error("expected false for nil context")
	}
}

func TestTracer_StartSpan_MaxSpansEviction(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithMaxSpans(3))

	for i := 0; i < 5; i++ {
		tracer.StartSpan("span")
	}

	spans := tracer.GetSpans()
	if len(spans) != 3 {
		t.Errorf("expected 3 spans, got %d", len(spans))
	}
	if spans[0].Name != "span" {
		t.Errorf("expected first span name 'span', got %s", spans[0].Name)
	}
}

func TestTracer_StartSpan_ContextNotSampled(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))

	ctx := SpanContext{
		TraceID: "trace-x",
		SpanID:  "span-x",
		Sampled: false,
	}

	span := tracer.StartSpan("child", WithContext(ctx))

	if span.spanContext.Sampled {
		t.Error("expected Sampled=false from context")
	}
	if span.TraceID != "trace-x" {
		t.Errorf("expected TraceID 'trace-x', got %s", span.TraceID)
	}
}

func TestConsoleExporter_ExportSpans(t *testing.T) {
	t.Parallel()
	exporter := &ConsoleExporter{}

	span := &Span{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Name:      "test-op",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(10 * time.Millisecond),
		Ended:     true,
		Tags:      map[string]string{"service.name": "test"},
	}

	err := exporter.ExportSpans([]*Span{span})
	if err != nil {
		t.Fatalf("ExportSpans() error = %v", err)
	}
}

func TestTracer_Export_WithExporter(t *testing.T) {
	t.Parallel()
	exporter := &mockExporter{}
	tracer := NewTracer(WithExporter(exporter))

	tracer.StartSpan("op1")

	err := tracer.Export()
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	if len(exporter.getSpans()) != 1 {
		t.Errorf("expected 1 span exported, got %d", len(exporter.getSpans()))
	}
}

func TestSpan_Duration_Ended(t *testing.T) {
	t.Parallel()
	span := &Span{
		StartTime: time.Now().Add(-time.Second),
		EndTime:   time.Now(),
		Ended:     true,
	}

	d := span.Duration()
	if d < time.Second-time.Millisecond || d > time.Second+time.Millisecond {
		t.Errorf("expected ~1s duration, got %v", d)
	}
}

func TestTracer_Inject_WithParentSpanID(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	ctx := SpanContext{
		TraceID:      "trace-1",
		SpanID:       "span-1",
		ParentSpanID: "parent-1",
		Sampled:      true,
	}

	headers := tracer.Inject(ctx)

	if headers[HeaderParentSpanID] != "parent-1" {
		t.Errorf("expected ParentSpanID 'parent-1', got %s", headers[HeaderParentSpanID])
	}
}

func TestTracer_StartSpan_NotSampled(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithSampler(&AlwaysOffSampler{}))

	span := tracer.StartSpan("not-sampled")

	if !span.Ended {
		t.Error("expected span to be ended for non-sampled")
	}
	if span.TraceID != "" {
		t.Error("expected empty TraceID for non-sampled span")
	}
}

func TestTracer_StartSpan_NoOpts(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))

	span := tracer.StartSpan("no-opts")

	if span.TraceID == "" {
		t.Error("expected non-empty TraceID")
	}
	if span.Tags["service.name"] != "test" {
		t.Errorf("expected service.name 'test', got %s", span.Tags["service.name"])
	}
}
