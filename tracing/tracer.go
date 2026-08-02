package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Tracer 追踪器结构。
//
// 管理 Span 的创建、采样和导出。
// Tracer 是并发安全的，支持高并发场景下的 Span 创建。
type Tracer struct {
	serviceName string
	sampler     Sampler
	exporter    Exporter
	mu          sync.RWMutex
	spans       []*Span
	maxSpans    int
	spanCount   atomic.Int64
	exportMu    sync.Mutex
}

// WithServiceName 设置服务名称。
func WithServiceName(name string) TracerOption {
	return func(t *Tracer) {
		t.serviceName = name
	}
}

// WithSampler 设置采样器。
func WithSampler(sampler Sampler) TracerOption {
	return func(t *Tracer) {
		t.sampler = sampler
	}
}

// WithExporter 设置导出器。
func WithExporter(exporter Exporter) TracerOption {
	return func(t *Tracer) {
		t.exporter = exporter
	}
}

// WithMaxSpans 设置最大 Span 数量限制（防止内存溢出）。
func WithMaxSpans(max int) TracerOption {
	return func(t *Tracer) {
		t.maxSpans = max
	}
}

// WithParent 设置父 Span。
//
// 用于创建父子 Span 关系，子 Span 会继承父 Span 的 TraceID。
func WithParent(parent *Span) SpanOption {
	return func(s *Span) {
		s.ParentSpanID = parent.SpanID
		s.spanContext.ParentSpanID = parent.SpanID
		s.TraceID = parent.TraceID
		s.spanContext.TraceID = parent.TraceID
	}
}

// WithContext 设置追踪上下文。
//
// 用于从外部恢复的 SpanContext 创建 Span。
func WithContext(ctx SpanContext) SpanOption {
	return func(s *Span) {
		s.spanContext = ctx
		s.TraceID = ctx.TraceID
		s.ParentSpanID = ctx.SpanID
		s.contextSet = true
	}
}

// WithTags 设置标签。
//
// 批量设置 Span 标签，用于记录元数据。
func WithTags(tags map[string]string) SpanOption {
	return func(s *Span) {
		for k, v := range tags {
			s.Tags[k] = v
		}
	}
}

// NewTracer 创建追踪器。
//
// 使用函数式选项模式配置 Tracer，支持服务名称、采样器、导出器等配置。
// 默认使用 AlwaysOnSampler 和 ConsoleExporter。
func NewTracer(opts ...TracerOption) *Tracer {
	tracer := &Tracer{
		serviceName: DefaultServiceName,
		sampler:     &AlwaysOnSampler{},
		exporter:    &ConsoleExporter{},
		maxSpans:    DefaultMaxSpans,
	}

	for _, opt := range opts {
		opt(tracer)
	}

	return tracer
}

// StartSpan 创建新的 Span。
//
// 根据采样器决定是否创建真实 Span，未采样时返回空 Span。
// 支持通过 SpanOption 设置父 Span、标签等。
func (t *Tracer) StartSpan(name string, opts ...SpanOption) *Span {
	t.mu.RLock()
	sampler := t.sampler
	serviceName := t.serviceName
	t.mu.RUnlock()

	if !sampler.ShouldSample() {
		return &Span{
			Name:  name,
			Ended: true,
		}
	}

	span := &Span{
		TraceID:   TraceID(generateID()),
		SpanID:    SpanID(generateID()),
		Name:      name,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Events:    make([]SpanEvent, 0),
	}

	span.Tags["service.name"] = serviceName

	for _, opt := range opts {
		opt(span)
	}

	// 如果 WithContext 已设置了有效的 TraceID，使用该 TraceID 而不是新生成的
	extractedTraceID := span.spanContext.TraceID
	if extractedTraceID != "" {
		span.TraceID = extractedTraceID
	}

	sampled := true
	if span.contextSet {
		sampled = span.spanContext.Sampled
	}

	span.spanContext = SpanContext{
		TraceID:      span.TraceID,
		SpanID:       span.SpanID,
		ParentSpanID: span.ParentSpanID,
		Sampled:      sampled,
	}

	t.mu.Lock()
	if t.maxSpans > 0 && len(t.spans) >= t.maxSpans {
		t.spans = t.spans[1:]
	}
	t.spans = append(t.spans, span)
	t.mu.Unlock()

	t.spanCount.Add(1)

	return span
}

// Inject 注入追踪上下文到 HTTP 头部。
//
// 将 SpanContext 转换为 HTTP 头部键值对，用于跨服务传播追踪上下文。
func (t *Tracer) Inject(ctx SpanContext) map[string]string {
	headers := make(map[string]string)
	headers[HeaderTraceID] = string(ctx.TraceID)
	headers[HeaderSpanID] = string(ctx.SpanID)
	if ctx.ParentSpanID != "" {
		headers[HeaderParentSpanID] = string(ctx.ParentSpanID)
	}
	headers[HeaderSampled] = fmt.Sprintf("%v", ctx.Sampled)
	return headers
}

// Extract 从 HTTP 头部提取追踪上下文。
//
// 从 HTTP 请求头中解析 TraceID、SpanID 等信息，恢复追踪上下文。
func (t *Tracer) Extract(headers map[string]string) SpanContext {
	ctx := SpanContext{
		TraceID: TraceID(headers[HeaderTraceID]),
		SpanID:  SpanID(headers[HeaderSpanID]),
		Sampled: headers[HeaderSampled] == "true",
	}

	if parentSpanID, ok := headers[HeaderParentSpanID]; ok {
		ctx.ParentSpanID = SpanID(parentSpanID)
	}

	return ctx
}

// GetSpans 获取所有 Span 的副本。
//
// 返回的切片是内部切片的副本，修改不会影响 Tracer 内部状态。
func (t *Tracer) GetSpans() []*Span {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spans := make([]*Span, len(t.spans))
	copy(spans, t.spans)
	return spans
}

// Export 导出所有 Span。
//
// 使用导出器将当前所有 Span 导出到外部系统（如控制台、Jaeger 等）。
// 导出操作使用独立的锁，不阻塞 Span 创建。
func (t *Tracer) Export() error {
	t.exportMu.Lock()
	defer t.exportMu.Unlock()

	t.mu.RLock()
	spans := make([]*Span, len(t.spans))
	copy(spans, t.spans)
	t.mu.RUnlock()

	if t.exporter == nil {
		return ErrExporterNotSet
	}

	return t.exporter.ExportSpans(spans)
}

// GetSpanCount 获取已创建的 Span 总数。
//
// 使用原子操作，无锁高效统计。
func (t *Tracer) GetSpanCount() int64 {
	return t.spanCount.Load()
}

// GetActiveSpanCount 获取当前活跃的 Span 数量。
func (t *Tracer) GetActiveSpanCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	count := 0
	for _, s := range t.spans {
		s.mu.Lock()
		ended := s.Ended
		s.mu.Unlock()
		if !ended {
			count++
		}
	}
	return count
}

// Clear 清除所有 Span。
func (t *Tracer) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.spans = nil
}

// generateID 生成 16 字节的随机 ID。
//
// 使用 crypto/rand 生成密码学安全的随机 ID。
// 如果 crypto/rand 失败，回退到时间戳方案。
func generateID() string {
	b := make([]byte, 8)
	n, err := rand.Read(b)
	if err != nil || n != len(b) {
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// AlwaysOnSampler 始终采样器。
type AlwaysOnSampler struct{}

// ShouldSample 实现 Sampler 接口，始终返回 true。
func (s *AlwaysOnSampler) ShouldSample() bool {
	return true
}

// AlwaysOffSampler 从不采样器。
type AlwaysOffSampler struct{}

// ShouldSample 实现 Sampler 接口，始终返回 false。
func (s *AlwaysOffSampler) ShouldSample() bool {
	return false
}

// ProbabilitySampler 概率采样器。
type ProbabilitySampler struct {
	rate float64
	mu   sync.Mutex
	rand *mathrand.Rand
}

// NewProbabilitySampler 创建概率采样器。
//
// rate 参数范围 [0.0, 1.0]，0.0 表示不采样，1.0 表示全量采样。
func NewProbabilitySampler(rate float64) *ProbabilitySampler {
	return &ProbabilitySampler{
		rate: rate,
		rand: mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample 实现 Sampler 接口，按概率返回是否采样。
func (s *ProbabilitySampler) ShouldSample() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rand.Float64() < s.rate
}

// ConsoleExporter 控制台导出器。
type ConsoleExporter struct{}

// ExportSpans 实现 Exporter 接口，将 Span 打印到控制台。
func (e *ConsoleExporter) ExportSpans(spans []*Span) error {
	for _, span := range spans {
		span.mu.Lock()
		tags := make(map[string]string, len(span.Tags))
		for k, v := range span.Tags {
			tags[k] = v
		}
		span.mu.Unlock()
		fmt.Printf("[TRACE] %s | %s | %s | %v | tags=%v\n",
			span.TraceID, span.SpanID, span.Name, span.Duration(), tags)
	}
	return nil
}

// tracingContextKey context 中存储 SpanContext 的键类型。
type tracingContextKey struct{}

// TraceFromContext 从 context 获取追踪上下文。
//
// 如果 context 中包含 SpanContext，返回 true；否则返回空的 SpanContext 和 false。
func TraceFromContext(ctx context.Context) (SpanContext, bool) {
	if ctx == nil {
		return SpanContext{}, false
	}
	if sc, ok := ctx.Value(tracingContextKey{}).(SpanContext); ok {
		return sc, true
	}
	return SpanContext{}, false
}

// ContextWithSpan 将 Span 上下文添加到 context。
//
// 返回新的 context，包含 Span 的追踪信息，可用于后续请求传播。
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, tracingContextKey{}, span.Context())
}
