package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/web/core"
)

// DefaultRequestIDConfig 返回默认的 RequestID 配置。
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		HeaderName: "X-Request-ID",
		Generator: func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		},
	}
}

// RequestID 创建请求 ID 中间件。
func RequestID(config ...RequestIDConfig) core.MiddlewareFunc {
	cfg := DefaultRequestIDConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx core.Context) {
		requestID := ctx.Header(cfg.HeaderName)
		if requestID == "" {
			requestID = cfg.Generator()
		}

		ctx.SetHeader(cfg.HeaderName, requestID)
		ctx.Next()
	}
}

// DefaultAccessLogConfig 返回默认的访问日志配置。
func DefaultAccessLogConfig() AccessLogConfig {
	return AccessLogConfig{
		SlowThreshold: 500 * time.Millisecond,
		Logger:        log.Build(),
	}
}

// AccessLog 创建访问日志中间件。
func AccessLog(config ...AccessLogConfig) core.MiddlewareFunc {
	cfg := DefaultAccessLogConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx core.Context) {
		start := time.Now()
		method := ctx.RequestMethod()
		uri := ctx.RequestURI()

		ctx.Next()

		duration := time.Since(start)

		if duration > cfg.SlowThreshold {
			cfg.Logger.Info(ctx.Context(), "慢请求",
				log.KeyValue{Key: "method", Value: method},
				log.KeyValue{Key: "uri", Value: uri},
				log.KeyValue{Key: "duration", Value: duration.String()},
			)
			return
		}

		cfg.Logger.Debug(ctx.Context(), "请求处理完成",
			log.KeyValue{Key: "method", Value: method},
			log.KeyValue{Key: "uri", Value: uri},
			log.KeyValue{Key: "duration", Value: duration.String()},
		)
	}
}

// DefaultErrorConfig 返回默认的错误配置。
func DefaultErrorConfig() ErrorConfig {
	return ErrorConfig{
		Logger: log.Build(),
	}
}

// Error 创建错误处理中间件。
func Error(config ...ErrorConfig) core.MiddlewareFunc {
	cfg := DefaultErrorConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx core.Context) {
		defer func() {
			if err := recover(); err != nil {
				cfg.Logger.Error(ctx.Context(), "捕获到 panic",
					log.KeyValue{Key: "error", Value: fmt.Sprintf("%v", err)},
					log.KeyValue{Key: "method", Value: ctx.RequestMethod()},
					log.KeyValue{Key: "uri", Value: ctx.RequestURI()},
				)

				ctx.AbortWithStatusJSON(500, map[string]any{
					"error": "Internal Server Error",
				})
			}
		}()

		ctx.Next()
	}
}

// DefaultCORSConfig 返回默认的 CORS 配置。
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}

// CORS 创建 CORS 中间件。
func CORS(config ...CORSConfig) core.MiddlewareFunc {
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	// 确保 AllowOrigins 不为空
	if len(cfg.AllowOrigins) == 0 {
		cfg.AllowOrigins = []string{"*"}
	}

	// CORS 规范要求：当 AllowCredentials 为 true 时，Access-Control-Allow-Origin 不能为 "*"
	if cfg.AllowCredentials {
		allowed := false
		for _, o := range cfg.AllowOrigins {
			if o == "*" {
				allowed = true
				break
			}
		}
		if allowed {
			cfg.AllowCredentials = false
		}
	}

	// 预计算 Origin 允许集合
	hasWildcardOrigin := false
	originSet := make(map[string]bool, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		if o == "*" {
			hasWildcardOrigin = true
		}
		originSet[o] = true
	}

	return func(ctx core.Context) {
		origin := ctx.Header("Origin")

		allowed := hasWildcardOrigin || originSet[origin]

		if ctx.RequestMethod() == "OPTIONS" {
			if allowed {
				if hasWildcardOrigin {
					ctx.SetHeader("Access-Control-Allow-Origin", "*")
				} else {
					ctx.SetHeader("Access-Control-Allow-Origin", origin)
				}
				ctx.SetHeader("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods))
				ctx.SetHeader("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders))
				if cfg.AllowCredentials {
					ctx.SetHeader("Access-Control-Allow-Credentials", "true")
				}
				ctx.SetHeader("Access-Control-Max-Age", fmt.Sprintf("%d", int(cfg.MaxAge.Seconds())))
			}
			ctx.AbortWithStatus(204)
			return
		}

		if allowed {
			if hasWildcardOrigin {
				ctx.SetHeader("Access-Control-Allow-Origin", "*")
			} else {
				ctx.SetHeader("Access-Control-Allow-Origin", origin)
			}
			if cfg.AllowCredentials {
				ctx.SetHeader("Access-Control-Allow-Credentials", "true")
			}
		}
		ctx.Next()
	}
}

func joinStrings(strs []string) string {
	var buf strings.Builder
	for i, s := range strs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(s)
	}
	return buf.String()
}
