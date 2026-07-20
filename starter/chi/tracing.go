package chi

import (
	"fmt"
	"net/http"

	"github.com/xudefa/enhance/tracing"
)

// TracingMiddleware Chi 框架的分布式链路追踪中间件。
//
// 自动从请求头提取追踪上下文，创建 Span 并记录请求信息。
func TracingMiddleware(tracer *tracing.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tracer == nil {
				next.ServeHTTP(w, r)
				return
			}

			headers := make(map[string]string)
			for k, v := range r.Header {
				if len(v) > 0 {
					// 使用 CanonicalHeaderKey 确保与 HeaderTraceID 等常量匹配
					headers[http.CanonicalHeaderKey(k)] = v[0]
				}
			}

			spanCtx := tracer.Extract(headers)

			routePattern := r.URL.Path
			if routeContext := r.Context().Value("RouteContext"); routeContext != nil {
				if rc, ok := routeContext.(*http.Request); ok {
					routePattern = rc.URL.Path
				}
			}

			spanName := r.Method + " " + routePattern

			opts := []tracing.SpanOption{
				tracing.WithTags(map[string]string{
					"http.method": r.Method,
					"http.url":    r.URL.Path,
					"http.host":   r.Host,
				}),
			}

			if spanCtx.TraceID != "" {
				opts = append(opts, tracing.WithContext(spanCtx))
			}

			span := tracer.StartSpan(spanName, opts...)
			defer span.End()

			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			span.SetTag("http.status_code", fmt.Sprintf("%d", rw.statusCode))

			if rw.statusCode >= 400 {
				span.SetStatus(tracing.StatusError)
			} else {
				span.SetStatus(tracing.StatusOK)
			}

			respHeaders := tracer.Inject(span.Context())
			for k, v := range respHeaders {
				w.Header().Set(k, v)
			}
		})
	}
}

// responseWriter 包装 http.ResponseWriter 以捕获状态码。
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}
