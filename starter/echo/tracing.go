package echo

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xudefa/enhance/tracing"
)

// TracingMiddleware Echo 框架的分布式链路追踪中间件。
//
// 自动从请求头提取追踪上下文，创建 Span 并记录请求信息。
func TracingMiddleware(tracer *tracing.Tracer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if tracer == nil {
				return next(c)
			}

			headers := make(map[string]string)
			for k, v := range c.Request().Header {
				if len(v) > 0 {
					// 使用 CanonicalHeaderKey 确保与 HeaderTraceID 等常量匹配
					headers[http.CanonicalHeaderKey(k)] = v[0]
				}
			}

			spanCtx := tracer.Extract(headers)

			path := c.Path()
			spanName := c.Request().Method + " " + path

			opts := []tracing.SpanOption{
				tracing.WithTags(map[string]string{
					"http.method": c.Request().Method,
					"http.url":    c.Request().URL.Path,
					"http.host":   c.Request().Host,
				}),
			}

			if spanCtx.TraceID != "" {
				opts = append(opts, tracing.WithContext(spanCtx))
			}

			span := tracer.StartSpan(spanName, opts...)
			defer span.End()

			err := next(c)

			span.SetTag("http.status_code", fmt.Sprintf("%d", c.Response().Status))

			if c.Response().Status >= 400 {
				span.SetStatus(tracing.StatusError)
			} else {
				span.SetStatus(tracing.StatusOK)
			}

			respHeaders := tracer.Inject(span.Context())
			for k, v := range respHeaders {
				c.Response().Header().Set(k, v)
			}

			return err
		}
	}
}
