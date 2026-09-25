package tracing

import "fmt"

// TraceHelper 追踪助手。
//
// 提供便捷的追踪方法，简化常见场景的追踪代码。
type TraceHelper struct {
	tracer *Tracer
}

// NewTraceHelper 创建追踪助手。
func NewTraceHelper(tracer *Tracer) *TraceHelper {
	return &TraceHelper{
		tracer: tracer,
	}
}

// traceOperation 追踪操作的公共逻辑，复用 Span 创建/状态设置/错误处理
func (h *TraceHelper) traceOperation(spanName string, tags map[string]string, errPrefix string, fn func() error) error {
	span := h.tracer.StartSpan(spanName, WithTags(tags))
	defer span.End()

	err := fn()
	if err != nil {
		span.SetStatus(StatusError)
		span.SetTag("error", err.Error())
		return fmt.Errorf("%s: %w", errPrefix, err)
	}
	span.SetStatus(StatusOK)
	return nil
}

// TraceHTTP 追踪 HTTP 请求。
//
// 自动创建 Span、记录 HTTP 方法和 URL、处理错误状态。
func (h *TraceHelper) TraceHTTP(method, url string, fn func() error) error {
	return h.traceOperation("HTTP "+method, map[string]string{
		"http.method": method,
		"http.url":    url,
	}, "HTTP 请求执行失败", fn)
}

// TraceDB 追踪数据库操作。
//
// 自动创建 Span、记录数据库操作类型和 SQL 语句、处理错误状态。
func (h *TraceHelper) TraceDB(operation, query string, fn func() error) error {
	return h.traceOperation("DB "+operation, map[string]string{
		"db.operation": operation,
		"db.statement": query,
	}, "DB 操作执行失败", fn)
}

// TraceRPC 追踪 RPC 调用。
//
// 自动创建 Span、记录 RPC 服务和方法、处理错误状态。
func (h *TraceHelper) TraceRPC(service, method string, fn func() error) error {
	return h.traceOperation("RPC "+service+"."+method, map[string]string{
		"rpc.service": service,
		"rpc.method":  method,
	}, "RPC 调用执行失败", fn)
}
