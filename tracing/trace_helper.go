package tracing

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

// TraceHTTP 追踪 HTTP 请求。
//
// 自动创建 Span、记录 HTTP 方法和 URL、处理错误状态。
func (h *TraceHelper) TraceHTTP(method, url string, fn func() error) error {
	span := h.tracer.StartSpan("HTTP "+method, WithTags(map[string]string{
		"http.method": method,
		"http.url":    url,
	}))
	defer span.End()

	err := fn()
	if err != nil {
		span.SetStatus(StatusError)
		span.SetTag("error", err.Error())
	} else {
		span.SetStatus(StatusOK)
	}

	return err
}

// TraceDB 追踪数据库操作。
//
// 自动创建 Span、记录数据库操作类型和 SQL 语句、处理错误状态。
func (h *TraceHelper) TraceDB(operation, query string, fn func() error) error {
	span := h.tracer.StartSpan("DB "+operation, WithTags(map[string]string{
		"db.operation": operation,
		"db.statement": query,
	}))
	defer span.End()

	err := fn()
	if err != nil {
		span.SetStatus(StatusError)
		span.SetTag("error", err.Error())
	} else {
		span.SetStatus(StatusOK)
	}

	return err
}

// TraceRPC 追踪 RPC 调用。
//
// 自动创建 Span、记录 RPC 服务和方法、处理错误状态。
func (h *TraceHelper) TraceRPC(service, method string, fn func() error) error {
	span := h.tracer.StartSpan("RPC "+service+"."+method, WithTags(map[string]string{
		"rpc.service": service,
		"rpc.method":  method,
	}))
	defer span.End()

	err := fn()
	if err != nil {
		span.SetStatus(StatusError)
		span.SetTag("error", err.Error())
	} else {
		span.SetStatus(StatusOK)
	}

	return err
}
