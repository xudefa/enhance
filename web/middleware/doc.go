// Package middleware 提供 HTTP 中间件实现。
//
// 该模块提供常用的 HTTP 中间件，包括请求 ID、访问日志、错误处理和 CORS 等。
// 所有中间件都实现了 core.MiddlewareFunc 接口。
//
// # 中间件列表
//
//   - RequestID: 为每个请求生成唯一的请求 ID
//   - AccessLog: 记录请求访问日志，支持慢请求检测
//   - Error: 全局错误处理，统一错误响应格式
//   - CORS: 跨域资源共享支持
//
// # 使用方式
//
// 使用默认配置：
//
//	router.Use(middleware.RequestID())
//	router.Use(middleware.AccessLog())
//	router.Use(middleware.Error())
//	router.Use(middleware.CORS())
//
// 使用自定义配置：
//
//	router.Use(middleware.RequestID(middleware.RequestIDConfig{
//	    HeaderName: "X-Trace-ID",
//	}))
//
//	router.Use(middleware.AccessLog(middleware.AccessLogConfig{
//	    SlowThreshold: time.Second,
//	}))
package middleware

import (
	"time"

	"github.com/xudefa/enhance/log"
)

// ==================== RequestID 中间件 ====================

// RequestIDConfig RequestID 中间件配置。
type RequestIDConfig struct {
	// HeaderName 请求 ID 的 HTTP 头名称。
	HeaderName string
	// Generator 请求 ID 生成函数。
	Generator func() string
}

// ==================== AccessLog 中间件 ====================

// AccessLogConfig 访问日志中间件配置。
type AccessLogConfig struct {
	// SlowThreshold 慢请求阈值。
	SlowThreshold time.Duration
	// Logger enhance 日志记录器（优先使用）。
	Logger log.Logger
}

// ==================== Error 中间件 ====================

// ErrorConfig 错误处理中间件配置。
type ErrorConfig struct {
	// Logger enhance 日志记录器（优先使用）。
	Logger log.Logger
}

// ==================== CORS 中间件 ====================

// CORSConfig CORS 中间件配置。
type CORSConfig struct {
	// AllowOrigins 允许的源列表。
	AllowOrigins []string
	// AllowMethods 允许的 HTTP 方法。
	AllowMethods []string
	// AllowHeaders 允许的 HTTP 头。
	AllowHeaders []string
	// AllowCredentials 是否允许携带凭证。
	AllowCredentials bool
	// MaxAge 预检请求缓存时间。
	MaxAge time.Duration
}
