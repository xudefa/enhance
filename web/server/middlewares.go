package server

import (
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/xudefa/enhance/web/mvc"
)

// RequestIDConfig 是 RequestID 中间件的配置。
type RequestIDConfig struct {
	// HeaderName 是用于存储请求 ID 的请求头名称。
	HeaderName string

	// Generator 是请求 ID 生成函数。
	Generator func() string
}

// Validate 验证 RequestID 配置是否有效。
func (c RequestIDConfig) Validate() error {
	if c.HeaderName == "" {
		return fmt.Errorf("请求头名称不能为空")
	}
	if c.Generator == nil {
		return fmt.Errorf("请求 ID 生成器不能为空")
	}
	return nil
}

// DefaultRequestIDConfig 返回默认的 RequestID 配置。
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		HeaderName: "X-Request-ID",
		Generator: func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		},
	}
}

// RequestIDMiddleware 创建请求 ID 中间件。
func RequestIDMiddleware(config RequestIDConfig) mvc.MiddlewareFunc {
	if err := config.Validate(); err != nil {
		config = DefaultRequestIDConfig()
	}

	return func(ctx mvc.Context) {
		requestID := ctx.Header(config.HeaderName)
		if requestID == "" {
			requestID = config.Generator()
		}

		ctx.SetHeader(config.HeaderName, requestID)
		ctx.Next()
	}
}

// AccessLogConfig 是访问日志中间件的配置。
type AccessLogConfig struct {
	// SlowThreshold 是慢请求阈值。
	SlowThreshold time.Duration

	// Logger 是日志记录器。
	Logger *slog.Logger
}

// Validate 验证 AccessLog 配置是否有效。
func (c AccessLogConfig) Validate() error {
	if c.Logger == nil {
		return fmt.Errorf("日志记录器不能为空")
	}
	return nil
}

// DefaultAccessLogConfig 返回默认的 AccessLog 配置。
func DefaultAccessLogConfig() AccessLogConfig {
	return AccessLogConfig{
		SlowThreshold: 500 * time.Millisecond,
		Logger:        slog.Default(),
	}
}

// AccessLogMiddleware 创建访问日志中间件。
func AccessLogMiddleware(config AccessLogConfig) mvc.MiddlewareFunc {
	if err := config.Validate(); err != nil {
		config = DefaultAccessLogConfig()
	}

	return func(ctx mvc.Context) {
		start := time.Now()
		method := ctx.RequestMethod()
		uri := ctx.RequestURI()

		ctx.Next()

		duration := time.Since(start)

		if duration > config.SlowThreshold {
			config.Logger.Warn("slow request",
				"method", method,
				"uri", uri,
				"duration", duration,
			)
			return
		}
		config.Logger.Info("request completed",
			"method", method,
			"uri", uri,
			"duration", duration,
		)
	}
}

// ErrorConfig 是错误处理中间件的配置。
type ErrorConfig struct {
	// Logger 是日志记录器。
	Logger *slog.Logger
}

// Validate 验证 Error 配置是否有效。
func (c ErrorConfig) Validate() error {
	if c.Logger == nil {
		return fmt.Errorf("日志记录器不能为空")
	}
	return nil
}

// DefaultErrorConfig 返回默认的错误配置。
func DefaultErrorConfig() ErrorConfig {
	return ErrorConfig{
		Logger: slog.Default(),
	}
}

// ErrorMiddleware 创建错误处理中间件。
func ErrorMiddleware(config ErrorConfig) mvc.MiddlewareFunc {
	if err := config.Validate(); err != nil {
		config = DefaultErrorConfig()
	}

	return func(ctx mvc.Context) {
		defer func() {
			if err := recover(); err != nil {
				config.Logger.Error("panic recovered",
					"error", err,
					"method", ctx.RequestMethod(),
					"uri", ctx.RequestURI(),
				)

				ctx.AbortWithStatusJSON(500, map[string]any{
					"error": "Internal Server Error",
				})
			}
		}()

		ctx.Next()
	}
}

// CORSConfig 是 CORS 中间件的配置。
type CORSConfig struct {
	// AllowOrigins 是允许的源列表。
	AllowOrigins []string

	// AllowMethods 是允许的 HTTP 方法列表。
	AllowMethods []string

	// AllowHeaders 是允许的请求头列表。
	AllowHeaders []string

	// AllowCredentials 是否允许携带凭证。
	AllowCredentials bool

	// MaxAge 是预检请求缓存时间。
	MaxAge time.Duration
}

// Validate 验证 CORS 配置是否有效。
func (c CORSConfig) Validate() error {
	if len(c.AllowOrigins) == 0 {
		return fmt.Errorf("允许的源列表不能为空")
	}
	if len(c.AllowMethods) == 0 {
		return fmt.Errorf("允许的 HTTP 方法列表不能为空")
	}
	return nil
}

// DefaultCORSConfig 返回默认的 CORS 配置。
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORSMiddleware 创建 CORS 中间件。
func CORSMiddleware(config CORSConfig) mvc.MiddlewareFunc {
	if err := config.Validate(); err != nil {
		config = DefaultCORSConfig()
	}

	// 预计算配置，避免每次请求重复处理
	hasWildcardOrigin := false
	originSet := make(map[string]bool, len(config.AllowOrigins))
	for _, o := range config.AllowOrigins {
		if o == "*" {
			hasWildcardOrigin = true
		}
		originSet[o] = true
	}

	allowMethodsStr := join(config.AllowMethods)
	allowHeadersStr := join(config.AllowHeaders)
	maxAgeStr := fmt.Sprintf("%.0f", config.MaxAge.Seconds())

	return func(ctx mvc.Context) {
		origin := ctx.Header("Origin")

		// O(1) 查找，替代原来的 O(n) 遍历
		allowed := hasWildcardOrigin || originSet[origin]

		if allowed {
			ctx.SetHeader("Access-Control-Allow-Origin", origin)
			if config.AllowCredentials {
				ctx.SetHeader("Access-Control-Allow-Credentials", "true")
			}
			ctx.SetHeader("Access-Control-Allow-Methods", allowMethodsStr)
			ctx.SetHeader("Access-Control-Allow-Headers", allowHeadersStr)
			ctx.SetHeader("Access-Control-Max-Age", maxAgeStr)
		}

		if ctx.RequestMethod() == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	}
}

// RecoveryMiddleware 创建 panic 恢复中间件。
func RecoveryMiddleware() mvc.MiddlewareFunc {
	return func(ctx mvc.Context) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"method", ctx.RequestMethod(),
					"uri", ctx.RequestURI(),
				)

				ctx.AbortWithStatusJSON(500, map[string]string{
					"error": "Internal Server Error",
				})
			}
		}()

		ctx.Next()
	}
}

// LoggingMiddleware 创建日志中间件。
func LoggingMiddleware() mvc.MiddlewareFunc {
	return func(ctx mvc.Context) {
		start := time.Now()
		slog.Info("request started",
			"method", ctx.RequestMethod(),
			"uri", ctx.RequestURI(),
		)

		ctx.Next()

		slog.Info("request completed",
			"method", ctx.RequestMethod(),
			"uri", ctx.RequestURI(),
			"duration", time.Since(start),
		)
	}
}

// GzipMiddleware 创建 Gzip 压缩中间件。
func GzipMiddleware() mvc.MiddlewareFunc {
	return func(ctx mvc.Context) {
		acceptEncoding := ctx.Header("Accept-Encoding")

		if contains(acceptEncoding, "gzip") {
			ctx.SetHeader("Content-Encoding", "gzip")
		}

		ctx.Next()
	}
}

// RealIPMiddleware 创建真实 IP 中间件。
func RealIPMiddleware() mvc.MiddlewareFunc {
	return func(ctx mvc.Context) {
		realIP := ctx.Header("X-Real-IP")
		if realIP == "" {
			realIP = ctx.Header("X-Forwarded-For")
		}
		if realIP == "" {
			host, _, _ := net.SplitHostPort(ctx.RequestURI())
			realIP = host
		}

		ctx.SetHeader("X-Real-IP", realIP)
		ctx.Next()
	}
}

// 辅助函数

func join(parts []string) string {
	return strings.Join(parts, ", ")
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
