// Package tracing 提供分布式追踪功能，用于 enhance 框架。
//
// 该模块提供 Span 创建、采样、导出和上下文传播等功能。
// 参考 OpenTelemetry 和 Spring Cloud Sleuth 的设计。
//
// # 架构设计
//
//   - Tracer: 追踪器，管理 Span 的创建、采样和导出
//   - Span: 追踪单元，表示一个操作的完整生命周期
//   - Sampler: 采样器接口，决定是否对请求进行采样
//   - Exporter: 导出器接口，将 Span 数据导出到外部系统
//   - SpanContext: 追踪上下文，用于跨服务传播
//   - TraceHelper: 追踪助手，简化常见场景的追踪代码
//
// # 核心功能
//
//   - Span 创建: 支持父子关系和标签
//   - 采样策略: 支持始终采样、从不采样和概率采样
//   - 导出器: 支持控制台导出和自定义导出
//   - 上下文传播: 支持 HTTP 头部注入和提取
//   - Span 数量限制: 防止内存溢出
//   - 并发安全: 使用读写锁和原子操作优化性能
//   - 安全 ID 生成: 使用 crypto/rand 生成随机 ID
//
// # 使用方式
//
// 创建追踪器：
//
//	tracer := tracing.NewTracer(
//	    tracing.WithServiceName("my-service"),
//	    tracing.WithSampler(tracing.NewProbabilitySampler(0.1)),
//	    tracing.WithExporter(&tracing.ConsoleExporter{}),
//	    tracing.WithMaxSpans(10000),
//	)
//
// 创建 Span：
//
//	span := tracer.StartSpan("HTTP GET", tracing.WithTags(map[string]string{
//	    "http.method": "GET",
//	    "http.url": "/api/users",
//	}))
//	defer span.End()
//
// 注入上下文：
//
//	headers := tracer.Inject(span.Context())
//
// # 集成后端
//
// 具体实现位于 starter 子包：
//
//   - starter/gin: Gin 框架集成
//   - starter/fiber: Fiber 框架集成
//   - starter/echo: Echo 框架集成
//   - starter/chi: Chi 框架集成
package tracing

import (
	"time"
)

// TraceID 追踪 ID 类型。
//
// 用于唯一标识一个完整的请求链路，跨多个服务保持不变。
type TraceID string

// SpanID Span ID 类型。
//
// 用于唯一标识一个 Span，同一链路中的每个 Span 有不同的 SpanID。
type SpanID string

// SpanStatus Span 状态枚举。
//
// 表示 Span 的执行结果状态。
type SpanStatus string

// Span 状态常量。
const (
	StatusOK        SpanStatus = "OK"        // 成功
	StatusError     SpanStatus = "ERROR"     // 错误
	StatusCancelled SpanStatus = "CANCELLED" // 取消
)

// HTTP 头部常量。
//
// 用于在 HTTP 请求间传播追踪上下文。
// 注意：使用 Go 的 http.CanonicalHeaderKey 格式（如 X-Trace-Id 而非 X-Trace-ID）
const (
	HeaderTraceID      = "X-Trace-Id"
	HeaderSpanID       = "X-Span-Id"
	HeaderParentSpanID = "X-Parent-Span-Id"
	HeaderSampled      = "X-Sampled"
)

// 默认配置常量。
const (
	DefaultServiceName  = "enhance-app"
	DefaultMaxSpans     = 10000
	DefaultSamplingRate = 1.0
)

// SpanOption Span 选项函数类型。
type SpanOption func(*Span)

// TracerOption 追踪器选项函数类型。
type TracerOption func(*Tracer)

// SpanEvent Span 事件。
//
// 记录 Span 生命周期中的重要时间点，可附加属性。
type SpanEvent struct {
	Timestamp  time.Time         `json:"timestamp"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// SpanContext Span 追踪上下文。
//
// 包含追踪的核心信息，用于在 HTTP 请求间传播。
type SpanContext struct {
	TraceID      TraceID `json:"trace_id"`
	SpanID       SpanID  `json:"span_id"`
	ParentSpanID SpanID  `json:"parent_span_id,omitempty"`
	Sampled      bool    `json:"sampled"`
}

// Sampler 采样器接口。
//
// 决定是否对请求进行采样。实现此接口可自定义采样策略。
//
// # 内置实现
//
//   - AlwaysOnSampler: 始终采样
//   - AlwaysOffSampler: 从不采样
//   - ProbabilitySampler: 概率采样
type Sampler interface {
	// ShouldSample 返回是否应该采样。
	ShouldSample() bool
}

// Exporter 导出器接口。
//
// 将 Span 数据导出到外部系统，如控制台、Jaeger、Zipkin 等。
// 实现此接口可自定义导出目标。
type Exporter interface {
	// ExportSpans 导出 Span 列表。
	ExportSpans(spans []*Span) error
}
