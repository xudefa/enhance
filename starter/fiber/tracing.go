package fiber

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/xudefa/enhance/tracing"
)

// TracingMiddleware Fiber 框架的分布式链路追踪中间件。
//
// 自动从请求头提取追踪上下文，创建 Span 并记录请求信息。
func TracingMiddleware(tracer *tracing.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if tracer == nil {
			return c.Next()
		}

		headers := make(map[string]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})

		spanCtx := tracer.Extract(headers)

		// 优先使用实际请求路径，c.Route().Path 在中间件中可能为空
		spanPath := c.Path()
		if spanPath == "" {
			spanPath = "/"
		}
		spanName := c.Method() + " " + spanPath

		opts := []tracing.SpanOption{
			tracing.WithTags(map[string]string{
				"http.method": c.Method(),
				"http.url":    c.Path(),
				"http.host":   c.Hostname(),
			}),
		}

		if spanCtx.TraceID != "" {
			opts = append(opts, tracing.WithContext(spanCtx))
		}

		span := tracer.StartSpan(spanName, opts...)
		defer span.End()

		err := c.Next()

		span.SetTag("http.status_code", fmt.Sprintf("%d", c.Response().StatusCode()))

		if c.Response().StatusCode() >= 400 {
			span.SetStatus(tracing.StatusError)
		} else {
			span.SetStatus(tracing.StatusOK)
		}

		respHeaders := tracer.Inject(span.Context())
		for k, v := range respHeaders {
			c.Set(k, v)
		}

		return err
	}
}
