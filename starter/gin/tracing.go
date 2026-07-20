package gin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/enhance/tracing"
)

// TracingMiddleware Gin 框架的分布式链路追踪中间件。
//
// 自动从请求头提取追踪上下文，创建 Span 并记录请求信息。
func TracingMiddleware(tracer *tracing.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tracer == nil {
			c.Next()
			return
		}

		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				// 使用 CanonicalHeaderKey 确保与 HeaderTraceID 等常量匹配
				headers[http.CanonicalHeaderKey(k)] = v[0]
			}
		}

		spanCtx := tracer.Extract(headers)

		spanName := c.Request.Method + " " + c.FullPath()
		if spanName == " " {
			spanName = c.Request.URL.Path
		}

		opts := []tracing.SpanOption{
			tracing.WithTags(map[string]string{
				"http.method": c.Request.Method,
				"http.url":    c.Request.URL.Path,
				"http.host":   c.Request.Host,
			}),
		}

		if spanCtx.TraceID != "" {
			opts = append(opts, tracing.WithContext(spanCtx))
		}

		span := tracer.StartSpan(spanName, opts...)
		defer span.End()

		// 将 span 存储到 gin.Context 中，方便路由处理函数获取
		c.Set("trace.span", span)
		c.Set("trace.traceId", span.Context().TraceID)
		c.Set("trace.spanId", span.Context().SpanID)

		// 在 c.Next() 之前设置响应头，确保响应头在响应发送前被写入
		respHeaders := tracer.Inject(span.Context())
		for k, v := range respHeaders {
			c.Header(k, v)
		}

		c.Next()

		span.SetTag("http.status_code", fmt.Sprintf("%d", c.Writer.Status()))

		if c.Writer.Status() >= 400 {
			span.SetStatus(tracing.StatusError)
		} else {
			span.SetStatus(tracing.StatusOK)
		}
	}
}
